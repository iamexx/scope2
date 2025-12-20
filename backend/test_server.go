// Test file to verify SteamCMD service interfaces
type SteamCMDService interface {
	EnsureSteamCMD() error
	DownloadDayZFiles() error
	UpdateDayZFiles() error
	ValidateFiles() error
	GetStatus() (map[string]interface{}, error)
	CheckFirstRun() error
}

type JobManager interface {
	StartSyncJob() (string, error)
	GetJobStatus(jobID string) (*services.Job, error)
	StartCleanupRoutine()
}

// This test file verifies the interfaces are properly implemented
var _ SteamCMDService = (*services.SteamCMDService)(nil)
var _ JobManager = (*services.JobManager)(nil)