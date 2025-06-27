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
	Attempt    int
	SourceURL  string
	Body       []byte
	BaseDomain string
}

// Scheduler orchestrates the scanning process.
type Scheduler struct {
	config        *config.Config
	client        *networking.Client
	processor     *Processor
	domainManager *networking.DomainManager
	logger        utils.Logger
	reporter      *report.Reporter
	jobQueue      chan Job
	wg            sync.WaitGroup
	activeJobs    atomic.Int32
	ctx           context.Context
	cancel        context.CancelFunc
	progressBar   *output.ProgressBar
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(cfg *config.Config, client *networking.Client, processor *Processor, dm *networking.DomainManager, logger utils.Logger, reporter *report.Reporter) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		config:        cfg,
		client:        client,
		processor:     processor,
		domainManager: dm,
		logger:        logger,
		reporter:      reporter,
		jobQueue:      make(chan Job, cfg.Concurrency),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// AddInitialTargets adiciona uma lista de alvos iniciais ao scheduler.
// Isso deve ser chamado antes de StartScan.
func (s *Scheduler) AddInitialTargets(targets []string) {
	for _, target := range targets {
		s.addJob(NewJob(target, FetchJS))
	}
}

// StartScan begins the scanning process.
func (s *Scheduler) StartScan() {
	if s.activeJobs.Load() == 0 {
		return
	}
	if !s.config.Silent {
		s.progressBar = output.NewProgressBar(int(s.activeJobs.Load()), 40)
		s.progressBar.Start()
	}
	for i := 0; i < s.config.Concurrency; i++ {
		s.wg.Add(1)
		go s.worker(s.ctx, i+1)
	}
	s.wg.Wait()
	if s.progressBar != nil {
		s.progressBar.Stop()
	}
}

func (s *Scheduler) worker(ctx context.Context, id int) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-s.jobQueue:
			if !ok { return }
			switch job.Type {
			case FetchJS:
				s.processFetchJob(ctx, job)
			case ProcessJS:
				s.processor.ProcessJSFileContent(job.SourceURL, job.Body)
				s.handleJobSuccess(job.Input, "ProcessJS")
			case VerifyPackage:
				s.processVerifyPackageJob(ctx, job)
			}
		}
	}
}

func (s *Scheduler) addJob(job Job) {
	s.activeJobs.Add(1)
	go func() { s.jobQueue <- job }()
}

func (s *Scheduler) processFetchJob(ctx context.Context, job Job) {
	isLocalFile := !strings.HasPrefix(job.Input, "http://") && !strings.HasPrefix(job.Input, "https://")
	var body []byte
	var err error

	if isLocalFile {
		s.logger.Debugf("Reading local file: %s", job.Input)
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

		s.domainManager.WaitForPermit(ctx, job.BaseDomain)

		reqData := networking.RequestData{URL: job.Input, Method: "GET", Ctx: ctx}
		respData := s.client.Do(reqData)
		s.domainManager.RecordRequestResult(job.BaseDomain, respData.StatusCode, respData.Error)

		if respData.Error != nil || respData.StatusCode >= 400 {
			s.handleJobFailure(job, respData.Error)
			return
		}
		defer respData.Response.Body.Close()

		body, err = io.ReadAll(respData.Response.Body)
		if err != nil {
			s.handleJobFailure(job, err)
			return
		}
	}

	// Cria o job de processamento
	processJob := NewJob(job.SourceURL, ProcessJS)
	processJob.Body = body
	s.addJob(processJob)

	s.handleJobSuccess(job.Input, "FetchJS")
}

func (s *Scheduler) processVerifyPackageJob(ctx context.Context, job Job) {
	packageName := job.Input
	checkURL := "https://registry.npmjs.org/" + url.PathEscape(packageName)

	reqData := networking.RequestData{URL: checkURL, Method: "HEAD", Ctx: ctx}
	
	s.logger.Debugf("Verifying package '%s' at %s", packageName, checkURL)
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

	s.handleJobSuccess(packageName, "VerifyPackage")
}

func (s *Scheduler) handleJobSuccess(input, jobType string) {
	s.logger.Debugf("Job succeeded: %s (%s)", input, jobType)
	s.handleJobCompletion()
}

func (s *Scheduler) handleJobFailure(job Job, err error) {
	if job.Attempt >= s.config.MaxRetries+1 {
		s.logger.Errorf("Job failed after %d attempts: %s. Error: %v", job.Attempt, job.Input, err)
		s.handleJobCompletion()
		return
	}
	
	s.logger.Warnf("Job failed (attempt %d), retrying: %s", job.Attempt, job.Input)
	job.Attempt++
	time.Sleep(time.Duration(job.Attempt) * 2 * time.Second) // Backoff simples
	s.addJob(job)
	s.handleJobCompletion()
}

func (s *Scheduler) handleJobCompletion() {
	if s.progressBar != nil {
		s.progressBar.Increment()
	}
	if s.activeJobs.Add(-1) <= 0 {
		s.shutdown()
	}
}

func (s *Scheduler) shutdown() {
	s.logger.Debugf("Scheduler shutting down...")
	if !s.isChannelClosed(s.jobQueue) {
		close(s.jobQueue)
	}
	s.cancel()
}

func (s *Scheduler) isChannelClosed(ch <-chan Job) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}

func NewJob(input string, jobType JobType) Job {
	return Job{Input: input, Type: jobType, Attempt: 1, SourceURL: input}
}
