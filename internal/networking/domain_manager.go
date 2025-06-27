package networking

import (
	"context"
	"sync"

	"DepScout/internal/config"
	"DepScout/internal/utils"

	"golang.org/x/time/rate"
)

type DomainManager struct {
	mu      sync.Mutex
	domains map[string]*DomainBucket
	config  *config.Config
	logger  utils.Logger
}

type DomainBucket struct {
	limiter *rate.Limiter
	mode    string
}

func NewDomainManager(cfg *config.Config, logger utils.Logger) *DomainManager {
	return &DomainManager{
		domains: make(map[string]*DomainBucket),
		config:  cfg,
		logger:  logger,
	}
}

func (dm *DomainManager) WaitForPermit(ctx context.Context, domain string) {
	dm.mu.Lock()
	bucket, exists := dm.domains[domain]
	if !exists {
		limiter := rate.NewLimiter(1.0, 1) // Default 1 RPS
		bucket = &DomainBucket{limiter: limiter, mode: "AUTO"}
		dm.domains[domain] = bucket
		dm.logger.Debugf("[DomainManager] Initialized bucket for domain '%s': Mode: %s, RPS=%.2f", domain, bucket.mode, bucket.limiter.Limit())
	}
	dm.mu.Unlock()

	if err := bucket.limiter.Wait(ctx); err != nil {
		dm.logger.Warnf("[DomainManager] Context canceled while waiting for permit to domain '%s'", domain)
	}
}

func (dm *DomainManager) RecordRequestSent(domain string) {
	// Pode ser expandido no futuro
}

func (dm *DomainManager) RecordRequestResult(domain string, statusCode int, err error) {
	// Pode ser expandido no futuro para ajustar o rate limit
}