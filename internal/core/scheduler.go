package core

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"DepScout/internal/config"
	"DepScout/internal/networking"
	"DepScout/internal/output"
	"DepScout/internal/report"
	"DepScout/internal/utils"
)

// JobType define o tipo de trabalho que um Job representa.
type JobType int

const (
	// FetchJS é um trabalho para baixar o conteúdo de um arquivo JavaScript.
	FetchJS JobType = iota
	// ProcessJS é um trabalho para processar o conteúdo de um arquivo JavaScript.
	ProcessJS
	// VerifyPackage é um novo tipo de job para verificar um pacote no registro npm.
	VerifyPackage
)

func (jt JobType) String() string {
	return [...]string{"FetchJS", "ProcessJS", "VerifyPackage"}[jt]
}

// Job representa uma unidade de trabalho para os workers.
type Job struct {
	Input      string
	Type       JobType
	SourceURL  string
	Body       []byte
	BaseDomain string
	Retries    int
}

// Scheduler orchestrates the scanning process.
type Scheduler struct {
	config         *config.Config
	client         *networking.Client
	processor      *Processor
	domainManager  *networking.DomainManager
	logger         *utils.Logger
	reporter       *report.Reporter
	progBar        *output.ProgressBar
	jobQueue       chan Job
	jobsWg         sync.WaitGroup
	producersWg    sync.WaitGroup
	workersWg      sync.WaitGroup
	initialAddWg   sync.WaitGroup
	requestCount   atomic.Int64
	stopRpsCounter chan bool
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(
	cfg *config.Config,
	client *networking.Client,
	processor *Processor,
	domainManager *networking.DomainManager,
	logger *utils.Logger,
	reporter *report.Reporter,
	progBar *output.ProgressBar,
) *Scheduler {
	return &Scheduler{
		config:        cfg,
		client:        client,
		processor:     processor,
		domainManager: domainManager,
		logger:        logger,
		reporter:      reporter,
		progBar:       progBar,
		jobQueue:      make(chan Job, cfg.Concurrency*2),
		stopRpsCounter: make(chan bool),
	}
}

// AddJob é o método síncrono para adicionar um novo trabalho ao pipeline.
// Ele garante que o WaitGroup seja incrementado corretamente.
func (s *Scheduler) AddJob(job Job) {
	s.jobsWg.Add(1)
	s.jobQueue <- job
}

// AddJobAsync adiciona um job de forma assíncrona para evitar deadlocks.
func (s *Scheduler) AddJobAsync(job Job) {
	s.producersWg.Add(1)
	go func() {
		defer s.producersWg.Done()
		s.AddJob(job)
	}()
}

// requeueJob adiciona um trabalho de volta à fila sem incrementar o WaitGroup.
// Usado para retries onde o trabalho original já foi contabilizado.
func (s *Scheduler) requeueJob(job Job) {
	s.jobQueue <- job
}

// AddInitialTargets adiciona uma lista de alvos iniciais ao scheduler.
// Esta função é síncrona e irá bloquear se o canal de jobs estiver cheio,
// o que é o comportamento esperado, pois os workers já estarão processando.
func (s *Scheduler) AddInitialTargets(targets []string) {
	s.initialAddWg.Add(1)
	go func() {
		defer s.initialAddWg.Done()
		for _, target := range targets {
			if target != "" {
				s.AddJob(NewJob(target, FetchJS))
			}
		}
	}()
}

// StartScan begins the scanning process.
func (s *Scheduler) StartScan() {
	s.workersWg.Add(s.config.Concurrency)
	for i := 0; i < s.config.Concurrency; i++ {
		go s.worker()
	}
	s.startRpsCounter()
}

func (s *Scheduler) Wait() {
	s.initialAddWg.Wait() // Espera a adição inicial de jobs terminar.
	s.jobsWg.Wait()       // Espera todos os jobs (iniciais e subsequentes) serem processados.
	s.producersWg.Wait()  // Espera todas as goroutines produtoras de jobs terminarem.
	close(s.jobQueue)
	s.workersWg.Wait()
	s.stopRpsCounter <- true // Para o contador de RPS
}

func (s *Scheduler) worker() {
	defer s.workersWg.Done()
	for job := range s.jobQueue {
		// Use a timeout for each job's context
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Timeout)*time.Second)
		switch job.Type {
		case FetchJS:
			s.processFetchJob(ctx, job)
		case ProcessJS:
			s.processor.ProcessJSFileContent(job.SourceURL, job.Body)
			s.handleJobSuccess(job.Input, "ProcessJS")
		case VerifyPackage:
			s.processVerifyPackageJob(ctx, job)
		}
		cancel()
		s.jobsWg.Done()
	}
}

func (s *Scheduler) processFetchJob(ctx context.Context, job Job) {
	isLocalFile := !strings.HasPrefix(job.Input, "http://") && !strings.HasPrefix(job.Input, "https://")
	var body []byte
	var err error
	maxBytes := int64(s.config.MaxFileSize * 1024)

	if isLocalFile {
		s.logger.Debugf("Reading local file: %s", job.Input)

		if !s.config.NoLimit {
			fileInfo, statErr := os.Stat(job.Input)
			if statErr != nil {
				s.handleJobFailure(job, fmt.Errorf("failed to stat local file: %w", statErr))
				return
			}
			if fileInfo.Size() > maxBytes {
				s.logger.Warnf("Skipping local file %s, size (%d KB) exceeds limit (%d KB)", job.Input, fileInfo.Size()/1024, s.config.MaxFileSize)
				s.handleJobFailure(job, fmt.Errorf("file size exceeds limit"))
				return
			}
		}

		body, err = os.ReadFile(job.Input)
		if err != nil {
			s.handleJobFailure(job, fmt.Errorf("failed to read local file: %w", err))
			return
		}
	} else {
		u, parseErr := url.Parse(job.Input)
		if parseErr != nil {
			s.handleJobFailure(job, parseErr)
			return
		}
		job.BaseDomain = u.Hostname()

		if err := s.domainManager.WaitForPermit(ctx, job.BaseDomain); err != nil {
			s.handleJobFailure(job, err)
			return
		}

		s.requestCount.Add(1)
		reqData := networking.RequestData{URL: job.Input, Method: "GET", Ctx: ctx}
		respData := s.client.Do(reqData)
		justDiscarded := s.domainManager.RecordRequestResult(job.BaseDomain, respData.StatusCode, respData.Error)

		if justDiscarded {
			s.logger.PublicWarnf("Domain '%s' has been discarded due to excessive 429 responses. All subsequent requests to this domain will be ignored.", job.BaseDomain)
		}

		// Se recebermos um 429, reenfileiramos o job para uma nova tentativa, se o limite não foi atingido.
		if respData.StatusCode == 429 {
			if job.Retries < 3 { // Limite de 3 retries por job
				job.Retries++
				s.logger.Warnf("Re-queueing job for %s due to 429. Attempt %d.", job.Input, job.Retries)
				s.requeueJob(job) // Usa requeueJob para não incrementar o WaitGroup novamente
			} else {
				s.logger.Errorf("Job for %s failed after %d retries due to 429.", job.Input, job.Retries)
				s.handleJobFailure(job, fmt.Errorf("HTTP status 429 after max retries"))
			}
			return
		}

		if respData.Error != nil || respData.StatusCode >= 400 {
			err = respData.Error
			if err == nil {
				err = fmt.Errorf("HTTP status %d", respData.StatusCode)
			}
			s.handleJobFailure(job, err)
			return
		}
		defer respData.Response.Body.Close()

		var limitedReader io.Reader = respData.Response.Body
		if !s.config.NoLimit {
			limitedReader = io.LimitReader(respData.Response.Body, maxBytes)
		}
		body, err = io.ReadAll(limitedReader)

		if err != nil {
			s.handleJobFailure(job, err)
			return
		}
		if !s.config.NoLimit && int64(len(body)) == maxBytes {
			s.logger.Warnf("File at %s may have been truncated as it reached the size limit of %d KB", job.Input, s.config.MaxFileSize)
		}
	}

	processJob := NewJob(job.SourceURL, ProcessJS)
	processJob.Body = body
	s.AddJobAsync(processJob)

	s.handleJobSuccess(job.Input, "FetchJS")
}

func (s *Scheduler) processVerifyPackageJob(ctx context.Context, job Job) {
	packageName := job.Input
	checkURL := "https://registry.npmjs.org/" + url.PathEscape(packageName)
	reqData := networking.RequestData{URL: checkURL, Method: "HEAD", Ctx: ctx}
	s.logger.Debugf("Verifying package '%s' at %s", packageName, checkURL)
	s.requestCount.Add(1)
	respData := s.client.Do(reqData)
	if respData.Error != nil {
		s.handleJobFailure(job, respData.Error)
		return
	}
	defer respData.Response.Body.Close()

	if respData.StatusCode == 404 {
		s.logger.Successf("Unclaimed package found: '%s'", packageName)
		s.reporter.AddFinding(report.Finding{
			UnclaimedPackage: packageName,
			FoundInSourceURL: job.SourceURL,
		})
	} else if respData.StatusCode == 200 {
		s.logger.Debugf("Package '%s' is claimed.", packageName)
	} else {
		s.logger.Warnf("Unexpected status code %d for package '%s'", respData.StatusCode, packageName)
	}
	// Nota: VerifyPackage jobs não incrementam a barra de progresso
	s.logger.Debugf("Job succeeded: %s (VerifyPackage)", packageName)
}

// handleJobSuccess incrementa a barra de progresso apenas para trabalhos FetchJS
func (s *Scheduler) handleJobSuccess(input, jobType string) {
	if jobType == "FetchJS" {
		s.progBar.Increment()
	}
	s.logger.Debugf("Job succeeded: %s (%s)", input, jobType)
}

// handleJobFailure incrementa a barra de progresso apenas para trabalhos FetchJS
func (s *Scheduler) handleJobFailure(job Job, err error) {
	if job.Type == FetchJS {
		s.progBar.Increment()
	}
	s.logger.Errorf("Job failed: %s (%s). Error: %v", job.Input, job.Type.String(), err)
}

func NewJob(input string, jobType JobType) Job {
	return Job{Input: input, SourceURL: input, Type: jobType, Retries: 0}
}

func (s *Scheduler) startRpsCounter() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		lastCount := int64(0)
		lastTime := time.Now()

		for {
			select {
			case <-s.stopRpsCounter:
				return
			case <-ticker.C:
				currentCount := s.requestCount.Load()
				currentTime := time.Now()
				duration := currentTime.Sub(lastTime).Seconds()
				if duration > 0 {
					rps := float64(currentCount-lastCount) / duration
					s.progBar.SetRPS(rps)
				}
				lastCount = currentCount
				lastTime = currentTime
			}
		}
	}()
}
