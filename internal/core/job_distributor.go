package core

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"
)

// JobDistributor manages intelligent distribution of jobs across domains
type JobDistributor struct {
	mu                 sync.RWMutex
	domainQueues       map[string]chan Job      // Separate queue for each domain
	globalQueue        chan Job                 // Fallback queue for non-domain jobs
	availableDomains   []string                 // List of domains with pending jobs
	domainJobCounts    map[string]int           // Track jobs per domain
	maxQueueSize       int                      // Maximum size per domain queue
	closed             bool
	distributorCtx     context.Context
	distributorCancel  context.CancelFunc
}

// NewJobDistributor creates a new intelligent job distributor
func NewJobDistributor(maxConcurrency int) *JobDistributor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &JobDistributor{
		domainQueues:     make(map[string]chan Job),
		globalQueue:      make(chan Job, maxConcurrency*2),
		availableDomains: make([]string, 0),
		domainJobCounts:  make(map[string]int),
		maxQueueSize:     maxConcurrency * 2, // Buffer size per domain
		distributorCtx:   ctx,
		distributorCancel: cancel,
	}
}

// AddJob adds a job to the appropriate domain queue or global queue
func (jd *JobDistributor) AddJob(job Job) error {
	jd.mu.Lock()
	defer jd.mu.Unlock()

	if jd.closed {
		return fmt.Errorf("job distributor is closed")
	}

	domain := jd.extractDomain(job)
	
	if domain == "" {
		// Non-domain jobs (local files, ProcessJS, VerifyPackage) go to global queue
		select {
		case jd.globalQueue <- job:
			return nil
		default:
			return fmt.Errorf("global queue is full")
		}
	}

	// Ensure domain queue exists
	if _, exists := jd.domainQueues[domain]; !exists {
		jd.domainQueues[domain] = make(chan Job, jd.maxQueueSize)
		jd.domainJobCounts[domain] = 0
	}

	// Add to domain-specific queue
	select {
	case jd.domainQueues[domain] <- job:
		jd.domainJobCounts[domain]++
		jd.updateAvailableDomains()
		return nil
	default:
		// Domain queue is full, add to global queue as fallback
		select {
		case jd.globalQueue <- job:
			return nil
		default:
			return fmt.Errorf("both domain and global queues are full")
		}
	}
}

// GetNextJob implements intelligent job selection for workers
func (jd *JobDistributor) GetNextJob(workerID int) (Job, bool) {
	for {
		jd.mu.Lock()
		if jd.closed {
			jd.mu.Unlock()
			// When closed, check global queue one last time but be careful about closed channels
			select {
			case job, ok := <-jd.globalQueue:
				if !ok {
					// Channel is closed, no more jobs
					return Job{}, false
				}
				return job, true
			default:
				// No jobs in global queue, distributor is done
				return Job{}, false
			}
		}

		// Strategy 1: Try to get a job from a domain that's not currently being processed
		// by other workers (intelligent distribution)
		selectedDomain := jd.selectOptimalDomain(workerID)
		if selectedDomain != "" {
			queue := jd.domainQueues[selectedDomain]
			jd.mu.Unlock()
			
			select {
			case job, ok := <-queue:
				if !ok {
					// Domain queue was closed, continue to next strategy
					continue
				}
				jd.mu.Lock()
				jd.domainJobCounts[selectedDomain]--
				if jd.domainJobCounts[selectedDomain] <= 0 {
					jd.updateAvailableDomains()
				}
				jd.mu.Unlock()
				return job, true
			case <-time.After(50 * time.Millisecond):
				// Quick timeout, try another strategy
			}
		} else {
			jd.mu.Unlock()
		}

		// Strategy 2: Try global queue (ProcessJS, VerifyPackage, local files)
		select {
		case job, ok := <-jd.globalQueue:
			if !ok {
				// Global queue is closed, no more jobs
				return Job{}, false
			}
			return job, true
		case <-time.After(50 * time.Millisecond):
			// Quick timeout, try strategy 3
		}

		// Strategy 3: If no optimal domain, try any available domain
		jd.mu.RLock()
		if len(jd.availableDomains) > 0 {
			// Pick a random domain to avoid thundering herd
			domain := jd.availableDomains[rand.Intn(len(jd.availableDomains))]
			queue := jd.domainQueues[domain]
			jd.mu.RUnlock()
			
			select {
			case job, ok := <-queue:
				if !ok {
					// Domain queue was closed, continue
					continue
				}
				jd.mu.Lock()
				jd.domainJobCounts[domain]--
				if jd.domainJobCounts[domain] <= 0 {
					jd.updateAvailableDomains()
				}
				jd.mu.Unlock()
				return job, true
			case <-time.After(100 * time.Millisecond):
				// Timeout, continue loop
			}
		} else {
			jd.mu.RUnlock()
		}

		// Strategy 4: Block briefly if nothing is available
		select {
		case <-jd.distributorCtx.Done():
			return Job{}, false
		case <-time.After(200 * time.Millisecond):
			// Continue the loop to try again
		}
	}
}

// selectOptimalDomain chooses the best domain for a worker
func (jd *JobDistributor) selectOptimalDomain(workerID int) string {
	if len(jd.availableDomains) == 0 {
		return ""
	}

	// Use worker ID to create some affinity but still allow distribution
	// This helps reduce contention while maintaining good distribution
	startIndex := workerID % len(jd.availableDomains)
	
	// Try domains starting from worker's preferred index
	for i := 0; i < len(jd.availableDomains); i++ {
		index := (startIndex + i) % len(jd.availableDomains)
		domain := jd.availableDomains[index]
		
		// Check if this domain has jobs
		if jd.domainJobCounts[domain] > 0 {
			return domain
		}
	}
	
	return ""
}

// extractDomain extracts domain from job for FetchJS jobs
func (jd *JobDistributor) extractDomain(job Job) string {
	if job.Type != FetchJS {
		return "" // Non-FetchJS jobs go to global queue
	}

	if !strings.HasPrefix(job.Input, "http://") && !strings.HasPrefix(job.Input, "https://") {
		return "" // Local files go to global queue
	}

	u, err := url.Parse(job.Input)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

// updateAvailableDomains updates the list of domains with available jobs
func (jd *JobDistributor) updateAvailableDomains() {
	jd.availableDomains = jd.availableDomains[:0] // Clear slice efficiently
	
	for domain, count := range jd.domainJobCounts {
		if count > 0 {
			jd.availableDomains = append(jd.availableDomains, domain)
		}
	}
}

// Close shuts down the job distributor
func (jd *JobDistributor) Close() {
	jd.mu.Lock()
	defer jd.mu.Unlock()
	
	if !jd.closed {
		jd.closed = true
		jd.distributorCancel()
		
		// Close all domain queues
		for _, queue := range jd.domainQueues {
			close(queue)
		}
		close(jd.globalQueue)
	}
}

// GetStats returns statistics about the distributor
func (jd *JobDistributor) GetStats() map[string]interface{} {
	jd.mu.RLock()
	defer jd.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_domains"] = len(jd.domainQueues)
	stats["available_domains"] = len(jd.availableDomains)
	stats["domain_job_counts"] = make(map[string]int)
	
	for domain, count := range jd.domainJobCounts {
		if count > 0 {
			stats["domain_job_counts"].(map[string]int)[domain] = count
		}
	}
	
	return stats
} 