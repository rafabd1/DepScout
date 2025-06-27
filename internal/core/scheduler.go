package core

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
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
}

// Scheduler orchestrates the scanning process.
type Scheduler struct {
	config        *config.Config
	client        *networking.Client
	processor     *Processor
	domainManager *networking.DomainManager
	logger        *utils.Logger
	reporter      *report.Reporter
	progBar       *output.ProgressBar
	jobQueue      chan Job
	jobsWg        sync.WaitGroup
	workersWg     sync.WaitGroup
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
	}
}

// AddInitialTargets adiciona uma lista de alvos iniciais ao scheduler.
// Isso deve ser chamado antes de StartScan.
func (s *Scheduler) AddInitialTargets(targets []string) {
	for _, target := range targets {
		if target != "" {
			s.addJob(NewJob(target, FetchJS))
		}
	}
}

// StartScan begins the scanning process.
func (s *Scheduler) StartScan() {
	s.workersWg.Add(s.config.Concurrency)
	for i := 0; i < s.config.Concurrency; i++ {
		go s.worker()
	}
}

func (s *Scheduler) Wait() {
	s.jobsWg.Wait()
	close(s.jobQueue)
	s.workersWg.Wait()
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
	}
}

func (s *Scheduler) addJob(job Job) {
	s.jobsWg.Add(1)
	s.jobQueue <- job
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
			err = respData.Error
			if err == nil {
				err = fmt.Errorf("HTTP status %d", respData.StatusCode)
			}
			s.handleJobFailure(job, err)
			return
		}
		defer respData.Response.Body.Close()
		body, err = io.ReadAll(respData.Response.Body)
		if err != nil {
			s.handleJobFailure(job, err)
			return
		}
	}

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
	s.progBar.Increment()
	s.logger.Debugf("Job succeeded: %s (%s)", input, jobType)
	s.jobsWg.Done()
}

func (s *Scheduler) handleJobFailure(job Job, err error) {
	s.progBar.Increment()
	s.logger.Errorf("Job failed: %s (%s). Error: %v", job.Input, job.Type.String(), err)
	s.jobsWg.Done()
}

func NewJob(input string, jobType JobType) Job {
	return Job{Input: input, SourceURL: input, Type: jobType}
}
