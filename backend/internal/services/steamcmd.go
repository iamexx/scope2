package services

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// CentralFilesPath is the path where DayZ server files are stored
	CentralFilesPath = "~/dayz/serverfiles"

	// SteamCMDPath is the path where SteamCMD is installed
	SteamCMDPath = "~/dayz/steamcmd"

	// DayZAppID is the Steam App ID for DayZ Server
	DayZAppID = "223350"

	// DayZExperimentalAppID is the Steam App ID for DayZ Experimental Server
	DayZExperimentalAppID = "1042420"

	// SteamCMDURL is the download URL for SteamCMD Linux x64
	SteamCMDURL = "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz"

	// RequiredDiskSpace is the minimum disk space required in bytes (30GB)
	RequiredDiskSpace = 30 * 1024 * 1024 * 1024
)

// SteamCMDService manages SteamCMD operations for DayZ server files
type SteamCMDService struct {
	db    *sql.DB
	mu    sync.RWMutex
	jobs  map[string]*Job
	jobMu sync.RWMutex
}

// Job represents a background task
type Job struct {
	ID       string
	Status   string
	Progress int
	Message  string
	Created  time.Time
}

// NewSteamCMDService creates a new SteamCMD service
func NewSteamCMDService(db *sql.DB) *SteamCMDService {
	service := &SteamCMDService{
		db:   db,
		jobs: make(map[string]*Job),
	}

	// Create directories if they don't exist
	service.createDirectories()

	return service
}

// createDirectories creates the required directory structure
func (s *SteamCMDService) createDirectories() error {
	directories := []string{
		expandPath(CentralFilesPath),
		expandPath(SteamCMDPath),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logError(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// EnsureSteamCMD downloads and extracts SteamCMD if it's not present
func (s *SteamCMDService) EnsureSteamCMD() error {
	steamCmdPath := expandPath(filepath.Join(SteamCMDPath, "steamcmd.sh"))

	// Check if steamcmd already exists
	if _, err := os.Stat(steamCmdPath); err == nil {
		logInfo("SteamCMD already exists, skipping download")
		return nil
	}

	logInfo("SteamCMD not found, downloading...")

	// Download SteamCMD
	if err := s.downloadSteamCMD(); err != nil {
		return fmt.Errorf("failed to download SteamCMD: %w", err)
	}

	// Ensure steamcmd.sh is executable
	if err := os.Chmod(steamCmdPath, 0755); err != nil {
		return fmt.Errorf("failed to make steamcmd.sh executable: %w", err)
	}

	logInfo("SteamCMD downloaded and extracted successfully")
	return nil
}

// downloadSteamCMD downloads and extracts SteamCMD
func (s *SteamCMDService) downloadSteamCMD() error {
	// Download the file
	resp, err := http.Get(SteamCMDURL)
	if err != nil {
		return fmt.Errorf("failed to download SteamCMD: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download SteamCMD: status code %d", resp.StatusCode)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "steamcmd_*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy the download to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save SteamCMD: %w", err)
	}

	// Extract the tar.gz
	if err := s.extractTarGz(tmpFile.Name(), expandPath(SteamCMDPath)); err != nil {
		return fmt.Errorf("failed to extract SteamCMD: %w", err)
	}

	return nil
}

// extractTarGz extracts a tar.gz file to the specified destination
func (s *SteamCMDService) extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

// DownloadDayZFiles downloads the DayZ server files
func (s *SteamCMDService) DownloadDayZFiles() error {
	return s.runSteamCMDCommand(DayZAppID, "Installing DayZ Server files...", true)
}

// UpdateDayZFiles updates the DayZ server files incrementally
func (s *SteamCMDService) UpdateDayZFiles() error {
	return s.runSteamCMDCommand(DayZAppID, "Updating DayZ Server files...", false)
}

// ValidateFiles validates the DayZ server files
func (s *SteamCMDService) ValidateFiles() error {
	return s.runSteamCMDCommand(DayZAppID, "Validating DayZ Server files...", true)
}

// runSteamCMDCommand executes a SteamCMD command
func (s *SteamCMDService) runSteamCMDCommand(appID, message string, validate bool) error {
	// Ensure we have enough disk space
	if err := s.checkDiskSpace(); err != nil {
		return err
	}

	// Ensure SteamCMD is installed
	if err := s.EnsureSteamCMD(); err != nil {
		return err
	}

	validateFlag := ""
	if validate {
		validateFlag = " validate"
	}

	cmd := exec.Command(
		expandPath(filepath.Join(SteamCMDPath, "steamcmd.sh")),
		"+force_install_dir", expandPath(CentralFilesPath),
		"+login", "anonymous",
		"+app_update", appID+validateFlag,
		"+quit",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logError(fmt.Sprintf("SteamCMD error: %v, output: %s", err, output))
		return fmt.Errorf("steamcmd failed: %w", err)
	}

	// Parse output for errors or success
	outputStr := string(output)
	if strings.Contains(outputStr, "ERROR") || strings.Contains(outputStr, "Failed") {
		logError(fmt.Sprintf("SteamCMD reported error: %s", outputStr))
		return fmt.Errorf("steamcmd reported error: %s", outputStr)
	}

	// Check if successful
	if strings.Contains(outputStr, "Success") || strings.Contains(outputStr, "fully installed") || strings.Contains(outputStr, "up to date") {
		// Update version info
		version := s.extractVersionFromOutput(outputStr)
		if err := s.updateVersionInfo(version); err != nil {
			logError(fmt.Sprintf("Failed to update version info: %v", err))
		}

		logInfo(fmt.Sprintf("DayZ server files %s successfully", message))
		return nil
	}

	logInfo(fmt.Sprintf("SteamCMD output: %s", outputStr))
	return nil
}

// extractVersionFromOutput extracts version info from SteamCMD output
func (s *SteamCMDService) extractVersionFromOutput(output string) string {
	// Look for version information in the output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Success") && strings.Contains(line, "app") {
			return strings.TrimSpace(line)
		}
	}
	// If no version found, return current timestamp
	return fmt.Sprintf("Version from %s", time.Now().Format("2006-01-02 15:04:05"))
}

// updateVersionInfo updates the SteamCMD version info in the database
func (s *SteamCMDService) updateVersionInfo(version string) error {
	query := `
        INSERT OR REPLACE INTO steamcmd_info (id, last_sync_date, current_version, central_files_path)
        VALUES (1, ?, ?, ?)
    `

	_, err := s.db.Exec(query, time.Now(), version, expandPath(CentralFilesPath))
	if err != nil {
		return fmt.Errorf("failed to update version info: %w", err)
	}

	return nil
}

// GetStatus returns the current status of SteamCMD
func (s *SteamCMDService) GetStatus() (map[string]interface{}, error) {
	var lastSync sql.NullTime
	var currentVersion sql.NullString
	var centralFilesPath sql.NullString

	query := `SELECT last_sync_date, current_version, central_files_path FROM steamcmd_info WHERE id = 1`
	err := s.db.QueryRow(query).Scan(&lastSync, &currentVersion, &centralFilesPath)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Check if files exist
	centralPath := expandPath(CentralFilesPath)
	filesExist := false
	shExists := false

	if _, err := os.Stat(centralPath); err == nil {
		// Check for key files
		if _, err := os.Stat(filepath.Join(centralPath, "DayZServer")); err == nil {
			filesExist = true
			shExists = true
		}
		if _, err := os.Stat(filepath.Join(centralPath, "DayZServer_x64")); err == nil {
			filesExist = true
			shExists = true
		}
	}

	status := "unknown"
	if filesExist && shExists && lastSync.Valid {
		status = "ready"
	} else if filesExist {
		status = "incomplete"
	} else {
		status = "missing"
	}

	return map[string]interface{}{
		"version":            currentVersion.String,
		"lastSync":           lastSync.Time.Format(time.RFC3339),
		"centralPath":        centralPath,
		"status":             status,
		"filesExist":         filesExist,
		"serverBinaryExists": shExists,
	}, nil
}

// CheckFirstRun checks if this is the first run and initiates download if needed
func (s *SteamCMDService) CheckFirstRun() error {
	status, err := s.GetStatus()
	if err != nil {
		return err
	}

	if status["status"] == "missing" {
		logInfo("First run detected: DayZ server files not found, starting download...")

		// Run download synchronously on startup
		if err := s.DownloadDayZFiles(); err != nil {
			logError(fmt.Sprintf("Failed to download DayZ files on first run: %v", err))
			return err
		}

		logInfo("First run completed: DayZ server files downloaded")
	}

	return nil
}

// checkDiskSpace checks if there's enough disk space
func (s *SteamCMDService) checkDiskSpace() error {
	var stat syscall.Statfs_t

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	if err := syscall.Statfs(homeDir, &stat); err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	availableSpace := stat.Bavail * uint64(stat.Bsize)

	if availableSpace < RequiredDiskSpace {
		freeGB := float64(availableSpace) / (1024 * 1024 * 1024)
		return fmt.Errorf("insufficient disk space: %.2f GB available, %.2f GB required", freeGB, float64(RequiredDiskSpace)/(1024*1024*1024))
	}

	return nil
}

// expandPath expands ~ to the user's home directory and cleans the path
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = filepath.Join(homeDir, path[2:])
	}
	return filepath.Clean(path)
}

// logInfo logs info messages
func logInfo(message string) {
	fmt.Printf("[INFO] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// logError logs error messages
func logError(message string) {
	fmt.Printf("[ERROR] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}
