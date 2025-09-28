package core

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rafabd1/DepScout/internal/networking"
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
	domainManager      *networking.DomainManager // Reference to check domain status
}

// NewJobDistributor creates a new intelligent job distributor
func NewJobDistributor(maxConcurrency int, domainManager *networking.DomainManager) *JobDistributor {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Calculate appropriate queue sizes based on expected workload
	// Global queue needs to handle many VerifyPackage jobs from JS processing
	globalQueueSize := maxConcurrency * 50  // Much larger for VerifyPackage jobs
	domainQueueSize := maxConcurrency * 10  // Reasonable size for domain-specific FetchJS jobs
	
	return &JobDistributor{
		domainQueues:     make(map[string]chan Job),
		globalQueue:      make(chan Job, globalQueueSize),
		availableDomains: make([]string, 0),
		domainJobCounts:  make(map[string]int),
		maxQueueSize:     domainQueueSize,
		distributorCtx:   ctx,
		distributorCancel: cancel,
		domainManager:    domainManager,
	}
}

// AddJob adds a job to the appropriate domain queue or global queue
func (jd *JobDistributor) AddJob(job Job) error {
	jd.mu.Lock()
	if jd.closed {
		jd.mu.Unlock()
		return fmt.Errorf("job distributor is closed")
	}

	domain := jd.extractDomain(job)
	
	if domain == "" {
		jd.mu.Unlock()
		// Non-domain jobs (local files, ProcessJS, VerifyPackage) go to global queue
		// Use a timeout to avoid indefinite blocking
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		select {
		case jd.globalQueue <- job:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for global queue space")
		}
	}

	// Ensure domain queue exists
	if _, exists := jd.domainQueues[domain]; !exists {
		jd.domainQueues[domain] = make(chan Job, jd.maxQueueSize)
		jd.domainJobCounts[domain] = 0
	}

	queue := jd.domainQueues[domain]
	jd.domainJobCounts[domain]++
	jd.updateAvailableDomains()
	jd.mu.Unlock()

	// Try to add to domain-specific queue with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	select {
	case queue <- job:
		return nil
	case <-ctx.Done():
		// Domain queue is full, try global queue as fallback
		jd.mu.Lock()
		jd.domainJobCounts[domain]-- // Revert the count since job wasn't actually added
		jd.updateAvailableDomains()
		jd.mu.Unlock()
		
		select {
		case jd.globalQueue <- job:
			return nil
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout: both domain and global queues are full")
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

		// Periodically refresh available domains to reactivate domains that came out of backoff
		jd.refreshAvailableDomains()

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

// selectOptimalDomain chooses the best domain for a worker, excluding blocked domains
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
		
		// Skip domains that are blocked, discarded, or in backoff
		if jd.isDomainBlocked(domain) {
			continue
		}
		
		// Check if this domain has jobs
		if jd.domainJobCounts[domain] > 0 {
			return domain
		}
	}
	
	return ""
}

// isDomainBlocked checks if a domain should be avoided for job distribution
func (jd *JobDistributor) isDomainBlocked(domain string) bool {
	if jd.domainManager == nil {
		return false
	}
	
	// Check if domain is discarded permanently
	if jd.domainManager.IsDiscarded(domain) {
		return true
	}
	
	// Check if domain is in backoff
	if jd.domainManager.IsDomainInBackoff(domain) {
		return true
	}
	
	return false
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

// updateAvailableDomains updates the list of domains with available jobs, excluding blocked domains
func (jd *JobDistributor) updateAvailableDomains() {
	jd.availableDomains = jd.availableDomains[:0] // Clear slice efficiently
	
	for domain, count := range jd.domainJobCounts {
		if count > 0 && !jd.isDomainBlocked(domain) {
			jd.availableDomains = append(jd.availableDomains, domain)
		}
	}
}

// refreshAvailableDomains updates available domains, potentially reactivating domains that came out of backoff
func (jd *JobDistributor) refreshAvailableDomains() {
	// Only refresh if we have few available domains but there are domain queues with jobs
	if len(jd.availableDomains) < len(jd.domainJobCounts)/3 {
		// First, try to redistribute jobs from blocked domains
		jd.redistributeBlockedJobsInternal()
		// Then update available domains list
		jd.updateAvailableDomains()
	}
}

// redistributeBlockedJobsInternal is called while already holding the lock
func (jd *JobDistributor) redistributeBlockedJobsInternal() {
	if jd.domainManager == nil {
		return
	}
	
	var redistributedCount int
	
	for domain, queue := range jd.domainQueues {
		if jd.isDomainBlocked(domain) && jd.domainJobCounts[domain] > 0 {
			// Try to move a few jobs from blocked domain queue to global queue
			moved := 0
			maxToMove := 5 // Don't move too many at once to avoid blocking
			
			for moved < maxToMove && moved < jd.domainJobCounts[domain] {
				select {
				case job := <-queue:
					select {
					case jd.globalQueue <- job:
						moved++
						redistributedCount++
					default:
						// Global queue is full, put job back and stop
						select {
						case queue <- job:
						default:
							// Both queues full, job will be lost but that's better than blocking
						}
						goto nextDomain
					}
				default:
					// No more jobs in this domain queue
					break
				}
			}
			
			nextDomain:
			// Update job count for this domain
			jd.domainJobCounts[domain] -= moved
			if jd.domainJobCounts[domain] < 0 {
				jd.domainJobCounts[domain] = 0
			}
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

// RedistributeBlockedJobs moves jobs from blocked domains to the global queue
func (jd *JobDistributor) RedistributeBlockedJobs() {
	jd.mu.Lock()
	defer jd.mu.Unlock()
	
	if jd.domainManager == nil {
		return
	}
	
	var redistributedCount int
	
	for domain, queue := range jd.domainQueues {
		if jd.isDomainBlocked(domain) && jd.domainJobCounts[domain] > 0 {
			// Move jobs from blocked domain queue to global queue
			moved := 0
			for moved < jd.domainJobCounts[domain] {
				select {
				case job := <-queue:
					select {
					case jd.globalQueue <- job:
						moved++
						redistributedCount++
					default:
						// Global queue is full, put job back and stop
						// This is a non-blocking operation, so we don't get stuck
						select {
						case queue <- job:
						default:
							// Both queues full, job will be lost but that's better than blocking
						}
						goto nextDomain
					}
				default:
					// No more jobs in this domain queue
					break
				}
			}
			
			nextDomain:
			// Update job count for this domain
			jd.domainJobCounts[domain] -= moved
			if jd.domainJobCounts[domain] < 0 {
				jd.domainJobCounts[domain] = 0
			}
		}
	}
	
	if redistributedCount > 0 {
		jd.updateAvailableDomains()
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
	
	// Add info about blocked domains
	var blockedDomains []string
	if jd.domainManager != nil {
		blockedDomains = jd.domainManager.GetBlockedDomains()
	}
	stats["blocked_domains"] = len(blockedDomains)
	stats["blocked_domain_list"] = blockedDomains
	
	for domain, count := range jd.domainJobCounts {
		if count > 0 {
			stats["domain_job_counts"].(map[string]int)[domain] = count
		}
	}
	
	return stats
} 