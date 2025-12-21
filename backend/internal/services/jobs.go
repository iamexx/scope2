package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// JobManager handles background job execution
type JobManager struct {
	service *SteamCMDService
}

// NewJobManager creates a new job manager
func NewJobManager(service *SteamCMDService) *JobManager {
	return &JobManager{
		service: service,
	}
}

// generateJobID generates a unique job ID
func generateJobID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if crypto random fails
		return fmt.Sprintf("job_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// StartSyncJob starts a sync job in the background
func (jm *JobManager) StartSyncJob() (string, error) {
	jobID := generateJobID()

	// Create job entry
	jm.service.jobMu.Lock()
	jm.service.jobs[jobID] = &Job{
		ID:       jobID,
		Status:   "running",
		Progress: 0,
		Message:  "Starting sync...",
		Created:  time.Now(),
	}
	jm.service.jobMu.Unlock()

	// Run in background
	go jm.runSyncJob(jobID)

	return jobID, nil
}

// runSyncJob executes the sync job
func (jm *JobManager) runSyncJob(jobID string) {
	defer func() {
		if r := recover(); r != nil {
			jm.service.jobMu.Lock()
			if job, exists := jm.service.jobs[jobID]; exists {
				job.Status = "failed"
				job.Message = fmt.Sprintf("Job failed with panic: %v", r)
			}
			jm.service.jobMu.Unlock()
			logError(fmt.Sprintf("Job %s panicked: %v", jobID, r))
		}
	}()

	// Update progress
	jm.service.jobMu.Lock()
	jm.service.jobs[jobID].Progress = 10
	jm.service.jobs[jobID].Message = "Checking SteamCMD..."
	jm.service.jobMu.Unlock()

	// Ensure SteamCMD is installed
	if err := jm.service.EnsureSteamCMD(); err != nil {
		jm.service.jobMu.Lock()
		jm.service.jobs[jobID].Status = "failed"
		jm.service.jobs[jobID].Message = fmt.Sprintf("Failed to ensure SteamCMD: %v", err)
		jm.service.jobMu.Unlock()
		return
	}

	// Check current status
	status, err := jm.service.GetStatus()
	if err != nil {
		jm.service.jobMu.Lock()
		jm.service.jobs[jobID].Status = "failed"
		jm.service.jobs[jobID].Message = fmt.Sprintf("Failed to get status: %v", err)
		jm.service.jobMu.Unlock()
		return
	}

	// Determine if we need to download or update
	jobType := "update"
	if status["status"] == "missing" {
		jobType = "download"
	}

	jm.service.jobMu.Lock()
	jm.service.jobs[jobID].Progress = 20
	jm.service.jobs[jobID].Message = fmt.Sprintf("Starting %s...", jobType)
	jm.service.jobMu.Unlock()

	// Execute the appropriate operation
	var operationErr error
	if jobType == "download" {
		operationErr = jm.service.DownloadDayZFiles()
	} else {
		operationErr = jm.service.UpdateDayZFiles()
	}

	if operationErr != nil {
		jm.service.jobMu.Lock()
		jm.service.jobs[jobID].Status = "failed"
		jm.service.jobs[jobID].Message = fmt.Sprintf("%s failed: %v", jobType, operationErr)
		jm.service.jobMu.Unlock()
		return
	}

	// Update progress to complete
	jm.service.jobMu.Lock()
	jm.service.jobs[jobID].Status = "completed"
	jm.service.jobs[jobID].Progress = 100
	jm.service.jobs[jobID].Message = fmt.Sprintf("%s completed successfully", jobType)
	jm.service.jobMu.Unlock()

	logInfo(fmt.Sprintf("Job %s completed", jobID))
}

// GetJobStatus returns the status of a job
func (jm *JobManager) GetJobStatus(jobID string) (*Job, error) {
	jm.service.jobMu.RLock()
	defer jm.service.jobMu.RUnlock()

	job, exists := jm.service.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}

	return job, nil
}

// CleanupOldJobs removes jobs older than 24 hours
func (jm *JobManager) CleanupOldJobs() {
	jm.service.jobMu.Lock()
	defer jm.service.jobMu.Unlock()

	cutoffTime := time.Now().Add(-24 * time.Hour)

	for jobID, job := range jm.service.jobs {
		if job.Created.Before(cutoffTime) {
			delete(jm.service.jobs, jobID)
			logInfo(fmt.Sprintf("Cleaned up old job %s", jobID))
		}
	}
}

// StartCleanupRoutine starts the cleanup routine for old jobs
func (jm *JobManager) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			jm.CleanupOldJobs()
		}
	}()

	logInfo("Started job cleanup routine")
}

// GetAllJobs returns all current jobs (for debugging/internal use)
func (jm *JobManager) GetAllJobs() map[string]*Job {
	jm.service.jobMu.RLock()
	defer jm.service.jobMu.RUnlock()

	// Return a copy to avoid concurrent modification issues
	jobsCopy := make(map[string]*Job)
	for k, v := range jm.service.jobs {
		jobsCopy[k] = &Job{
			ID:       v.ID,
			Status:   v.Status,
			Progress: v.Progress,
			Message:  v.Message,
			Created:  v.Created,
		}
	}

	return jobsCopy
}
