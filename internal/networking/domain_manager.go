package networking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"DepScout/internal/config"
	"DepScout/internal/utils"

	"golang.org/x/time/rate"
)

type DomainManager struct {
	mu               sync.Mutex
	domains          map[string]*DomainBucket
	config           *config.Config
	logger           *utils.Logger
	domainBuckets    map[string]*rate.Limiter
	discardedDomains sync.Map
}

type DomainBucket struct {
	limiter      *rate.Limiter
	mode         string
	inBackoff    bool
	nextAttempt  time.Time
	retryCounter int
}

func NewDomainManager(cfg *config.Config, logger *utils.Logger) *DomainManager {
	return &DomainManager{
		domains:       make(map[string]*DomainBucket),
		config:        cfg,
		logger:        logger,
		domainBuckets: make(map[string]*rate.Limiter),
	}
}

func (dm *DomainManager) WaitForPermit(ctx context.Context, domain string) error {
	if _, isDiscarded := dm.discardedDomains.Load(domain); isDiscarded {
		return fmt.Errorf("domain %s is discarded", domain)
	}

	dm.mu.Lock()
	bucket, exists := dm.domains[domain]
	if !exists {
		// Inicia com uma taxa baixa e segura. O sistema irá ajustá-la para cima em caso de sucesso.
		initialRate := rate.Limit(2.0)
		limiter := rate.NewLimiter(initialRate, 2)
		bucket = &DomainBucket{limiter: limiter, mode: "AUTO"}
		dm.domains[domain] = bucket
		dm.logger.Debugf("[DomainManager] Initialized bucket for domain '%s': Mode: %s, RPS=%.2f, Burst=%d", domain, bucket.mode, bucket.limiter.Limit(), bucket.limiter.Burst())
	}

	// Check if domain is in backoff period
	if bucket.inBackoff {
		if waitTime := time.Until(bucket.nextAttempt); waitTime > 0 {
			dm.logger.Warnf("[DomainManager] Domain '%s' is in backoff. Waiting for %s", domain, waitTime.Round(time.Second))
			dm.mu.Unlock() // Unlock while waiting

			timer := time.NewTimer(waitTime)
			select {
			case <-ctx.Done(): // Context was canceled (e.g., job timed out)
				timer.Stop()
				dm.logger.Warnf("[DomainManager] Context canceled for '%s' during backoff wait.", domain)
				// Do not re-lock, just return. The Wait() below will handle the ctx error.
				return ctx.Err()
			case <-timer.C:
				// Wait time finished.
			}

			dm.mu.Lock() // Re-lock to safely modify bucket state
		}
		// Backoff period is over, reset it.
		bucket.inBackoff = false
		dm.logger.Infof("[DomainManager] Backoff period for '%s' has ended. Resuming requests.", domain)
	}
	dm.mu.Unlock()

	if err := bucket.limiter.Wait(ctx); err != nil {
		dm.logger.Warnf("[DomainManager] Context canceled while waiting for permit to domain '%s'", domain)
		return err
	}
	return nil
}

func (dm *DomainManager) RecordRequestSent(domain string) {
	// Pode ser expandido no futuro
}

func (dm *DomainManager) RecordRequestResult(domain string, statusCode int, err error) (justDiscarded bool) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	bucket, exists := dm.domains[domain]
	if !exists {
		return // Should not happen if WaitForPermit was called
	}

	switch {
	case statusCode == 429:
		// Reduz a taxa multiplicativamente em caso de 'Too Many Requests'
		newLimit := bucket.limiter.Limit() * 0.7
		if newLimit < 0.1 {
			newLimit = 0.1 // Define um piso mínimo para a taxa
		}
		bucket.limiter.SetLimit(newLimit)

		// Ativa o backoff exponencial
		bucket.retryCounter++
		backoffDuration := time.Second * time.Duration(2<<bucket.retryCounter)
		if backoffDuration > time.Minute {
			backoffDuration = time.Minute
		}
		bucket.inBackoff = true
		bucket.nextAttempt = time.Now().Add(backoffDuration)

		dm.logger.Warnf("[DomainManager] 429 for '%s'. Rate limit reduced to %.2f req/s. Backoff for %s.", domain, newLimit, backoffDuration)

		if bucket.retryCounter > 5 { // Limite de 5 retries de 429 para um domínio
			_, loaded := dm.discardedDomains.LoadOrStore(domain, true)
			if !loaded {
				return true // Sinaliza que o domínio acabou de ser descartado
			}
		}

	default:
		// Aumenta a taxa para qualquer resposta que NÃO seja 429, apenas se não estiver em backoff
		if !bucket.inBackoff && statusCode != 429 && err == nil {
			newLimit := bucket.limiter.Limit() + 0.2
			if newLimit > rate.Limit(dm.config.MaxRateLimit) { // Usa o teto máximo da config
				newLimit = rate.Limit(dm.config.MaxRateLimit)
			}
			bucket.limiter.SetLimit(newLimit)
			dm.logger.Debugf("[DomainManager] Non-429 response for '%s' (status: %d). Rate limit increased to %.2f req/s.", domain, statusCode, newLimit)
		}
		// Note: Não fazemos nada para erros de rede ou outros casos
		// Apenas o 429 reduz a taxa, qualquer resposta válida (não-429) pode aumentar
	}
	return false
}