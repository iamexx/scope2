package services

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/iamexx/scope2-dayz-api/internal/db"
	"github.com/iamexx/scope2-dayz-api/internal/models"
)

const (
	ServerStatusRunning = "running"
	ServerStatusStopped = "stopped"
	ServerStatusError   = "error"

	ShutdownTimeout     = 30 * time.Second
	StatusCheckInterval = 5 * time.Second
)

type ProcessInfo struct {
	PID       int
	StartTime time.Time
	Stdout    *bytes.Buffer
	Stderr    *bytes.Buffer
}

type ProcessManager struct {
	mu        sync.RWMutex
	db        *sql.DB
	processes map[int64]*ProcessInfo
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		db:        db.DB,
		processes: make(map[int64]*ProcessInfo),
	}
}

func (pm *ProcessManager) StartServer(serverID int64, server *models.Server, serverFolder string) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if server is already running
	if _, running := pm.processes[serverID]; running {
		if pm.isProcessRunning(pm.processes[serverID].PID) {
			return 0, fmt.Errorf("server is already running")
		}
		delete(pm.processes, serverID)
	}

	// Verify binary exists
	binaryPath := filepath.Join(serverFolder, "DayZServer")
	if _, err := os.Stat(binaryPath); err != nil {
		return 0, fmt.Errorf("DayZ server binary not found at %s: %v", binaryPath, err)
	}

	// Verify config file exists
	configPath := filepath.Join(serverFolder, "serverDZ.cfg")
	if _, err := os.Stat(configPath); err != nil {
		return 0, fmt.Errorf("config file not found at %s: %v", configPath, err)
	}

	// Build command
	cmd := exec.Command(
		"./DayZServer",
		fmt.Sprintf("-config=serverDZ.cfg"),
		fmt.Sprintf("-port=%d", server.Port),
		"-BEpath=battleye",
		"-profiles=profiles",
		"-dologs",
		"-adminlog",
		"-netlog",
		"-freezecheck",
	)

	// Set working directory
	cmd.Dir = serverFolder

	// Create output buffers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the process
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start server: %v", err)
	}

	pid := cmd.Process.Pid
	processInfo := &ProcessInfo{
		PID:       pid,
		StartTime: time.Now(),
		Stdout:    &stdout,
		Stderr:    &stderr,
	}

	pm.processes[serverID] = processInfo

	// Update database status
	if err := pm.updateServerStatus(serverID, ServerStatusRunning); err != nil {
		log.Printf("warning: failed to update server status in database: %v", err)
	}

	// Log the operation
	log.Printf("Server %d started with PID %d", serverID, pid)

	// Monitor process in goroutine
	go pm.monitorProcess(serverID, cmd, server.Name)

	return pid, nil
}

func (pm *ProcessManager) StopServer(serverID int64) error {
	pm.mu.Lock()
	processInfo, exists := pm.processes[serverID]
	pm.mu.Unlock()

	if !exists {
		return fmt.Errorf("server is not running")
	}

	pid := processInfo.PID

	// Send SIGTERM for graceful shutdown
	process, err := os.FindProcess(pid)
	if err != nil {
		pm.mu.Lock()
		delete(pm.processes, serverID)
		pm.mu.Unlock()
		return fmt.Errorf("server process not found: %v", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %v", err)
	}

	// Wait for graceful shutdown with timeout
	deadline := time.Now().Add(ShutdownTimeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !pm.isProcessRunning(pid) {
				// Process terminated gracefully
				pm.mu.Lock()
				delete(pm.processes, serverID)
				pm.mu.Unlock()

				if err := pm.updateServerStatus(serverID, ServerStatusStopped); err != nil {
					log.Printf("warning: failed to update server status in database: %v", err)
				}

				log.Printf("Server %d stopped gracefully", serverID)
				return nil
			}
		case <-time.After(time.Until(deadline)):
			// Force kill with SIGKILL
			if err := process.Signal(syscall.SIGKILL); err != nil {
				log.Printf("warning: failed to send SIGKILL: %v", err)
			}

			pm.mu.Lock()
			delete(pm.processes, serverID)
			pm.mu.Unlock()

			if err := pm.updateServerStatus(serverID, ServerStatusStopped); err != nil {
				log.Printf("warning: failed to update server status in database: %v", err)
			}

			log.Printf("Server %d force killed after timeout", serverID)
			return nil
		}
	}
}

func (pm *ProcessManager) RestartServer(serverID int64, server *models.Server, serverFolder string) error {
	// Stop the server
	if err := pm.StopServer(serverID); err != nil {
		// If server wasn't running, continue with start
		if fmt.Sprintf("%v", err) != "server is not running" {
			return fmt.Errorf("failed to stop server: %v", err)
		}
	}

	// Wait a bit for cleanup
	time.Sleep(2 * time.Second)

	// Start the server
	_, err := pm.StartServer(serverID, server, serverFolder)
	return err
}

func (pm *ProcessManager) GetStatus(serverID int64) (string, int, time.Duration, string, error) {
	pm.mu.RLock()
	processInfo, exists := pm.processes[serverID]
	pm.mu.RUnlock()

	if !exists {
		return ServerStatusStopped, 0, 0, "", nil
	}

	pid := processInfo.PID
	if !pm.isProcessRunning(pid) {
		pm.mu.Lock()
		delete(pm.processes, serverID)
		pm.mu.Unlock()

		if err := pm.updateServerStatus(serverID, ServerStatusStopped); err != nil {
			log.Printf("warning: failed to update server status in database: %v", err)
		}

		return ServerStatusStopped, 0, 0, "", nil
	}

	uptime := time.Since(processInfo.StartTime)
	return ServerStatusRunning, pid, uptime, "", nil
}

func (pm *ProcessManager) monitorProcess(serverID int64, cmd *exec.Cmd, serverName string) {
	// Wait for the process to exit
	err := cmd.Wait()

	pm.mu.Lock()
	delete(pm.processes, serverID)
	pm.mu.Unlock()

	if err := pm.updateServerStatus(serverID, ServerStatusStopped); err != nil {
		log.Printf("warning: failed to update server status in database: %v", err)
	}

	if err != nil {
		log.Printf("Server %s (ID: %d) crashed: %v", serverName, serverID, err)
	} else {
		log.Printf("Server %s (ID: %d) exited cleanly", serverName, serverID)
	}
}

func (pm *ProcessManager) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (pm *ProcessManager) updateServerStatus(serverID int64, status string) error {
	if pm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := "UPDATE servers SET status = ? WHERE id = ?"
	_, err := pm.db.Exec(query, status, serverID)
	return err
}
