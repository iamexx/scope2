package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/iamexx/scope2-dayz-api/internal/models"
)

type ServerHandlers struct {
	logService    *LogService
	configEditor  *ConfigEditor
	serverService *ServerService
}

func NewServerHandlers() *ServerHandlers {
	return &ServerHandlers{
		logService:    NewLogService(),
		configEditor:  NewConfigEditor(),
		serverService: &ServerService{},
	}
}

func (h *ServerHandlers) GetLogs(w http.ResponseWriter, r *http.Request, serverID int64) {
	server, err := h.serverService.GetServer(serverID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Server not found")
		return
	}

	logType := r.URL.Query().Get("type")
	if logType == "" {
		logType = "console"
	}

	linesStr := r.URL.Query().Get("lines")
	if linesStr == "" {
		linesStr = "100"
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}

	lines := 100
	offset := 0
	fmt.Sscanf(linesStr, "%d", &lines)
	fmt.Sscanf(offsetStr, "%d", &offset)

	if lines <= 0 {
		lines = 100
	}
	if lines > 1000 {
		lines = 1000
	}

	logs, totalLines, err := h.logService.ReadLogs(server.ConfigPath, logType, lines, offset)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, http.StatusNotFound, "Log file not found")
		} else {
			writeJSONError(w, http.StatusInternalServerError, "Failed to read logs")
		}
		return
	}

	response := struct {
		Logs  []LogEntry `json:"logs"`
		Total int        `json:"total"`
	}{
		Logs:  logs,
		Total: totalLines,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *ServerHandlers) GetAdminLogs(w http.ResponseWriter, r *http.Request, serverID int64) {
	server, err := h.serverService.GetServer(serverID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Server not found")
		return
	}

	linesStr := r.URL.Query().Get("lines")
	if linesStr == "" {
		linesStr = "100"
	}

	lines := 100
	fmt.Sscanf(linesStr, "%d", &lines)

	if lines <= 0 {
		lines = 100
	}
	if lines > 1000 {
		lines = 1000
	}

	events, err := h.logService.ParseAdminLog(server.ConfigPath, lines)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to parse admin logs")
		return
	}

	writeJSONResponse(w, http.StatusOK, events)
}

func (h *ServerHandlers) GetConfig(w http.ResponseWriter, r *http.Request, serverID int64) {
	server, err := h.serverService.GetServer(serverID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Server not found")
		return
	}

	config, err := h.configEditor.ReadConfig(server.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, http.StatusNotFound, "Config file not found")
		} else {
			writeJSONError(w, http.StatusInternalServerError, "Failed to read config")
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, config)
}

func (h *ServerHandlers) UpdateConfig(w http.ResponseWriter, r *http.Request, serverID int64, userID string) {
	server, err := h.serverService.GetServer(serverID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Server not found")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if len(updates) == 0 {
		writeJSONError(w, http.StatusBadRequest, "No updates provided")
		return
	}

	if err := h.configEditor.BackupConfig(server.ConfigPath); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to create backup")
		return
	}

	config, validationErrors, err := h.configEditor.UpdateConfig(server.ConfigPath, updates)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to update config")
		return
	}

	if len(validationErrors) > 0 {
		response := struct {
			Status  string                  `json:"status"`
			Message string                  `json:"message"`
			Errors  []ConfigValidationError `json:"errors"`
		}{
			Status:  "error",
			Message: "Validation failed",
			Errors:  validationErrors,
		}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	if err := h.configEditor.WriteConfig(server.ConfigPath, config); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to write config")
		return
	}

	if err := h.configEditor.LogConfigChange(userID, updates); err != nil {
		fmt.Printf("Warning: Failed to log config change: %v\n", err)
	}

	response := struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Config  *ServerConfig `json:"config"`
	}{
		Status:  "success",
		Message: "Configuration updated successfully",
		Config:  config,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *ServerHandlers) GetConfigBackup(w http.ResponseWriter, r *http.Request, serverID int64) {
	server, err := h.serverService.GetServer(serverID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Server not found")
		return
	}

	backupPath := h.configEditor.GetBackupPath(server.ConfigPath)
	if backupPath == "" {
		writeJSONError(w, http.StatusNotFound, "No backup found")
		return
	}

	config, err := h.configEditor.ReadConfig(backupPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to read backup config")
		return
	}

	timestamp := getBackupTimestamp(backupPath)

	response := struct {
		Timestamp int64         `json:"timestamp"`
		Config    *ServerConfig `json:"config"`
	}{
		Timestamp: timestamp,
		Config:    config,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *ServerHandlers) RestoreConfig(w http.ResponseWriter, r *http.Request, serverID int64, userID string) {
	server, err := h.serverService.GetServer(serverID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Server not found")
		return
	}

	if err := h.configEditor.RestoreFromBackup(server.ConfigPath); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to restore config")
		return
	}

	changeLog := map[string]interface{}{
		"action": "restore_from_backup",
		"path":   server.ConfigPath,
	}

	if err := h.configEditor.LogConfigChange(userID, changeLog); err != nil {
		fmt.Printf("Warning: Failed to log config restoration: %v\n", err)
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "Configuration restored from backup",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

type ServerService struct {
	db interface{}
}

func (s *ServerService) GetServer(id int64) (*models.Server, error) {
	return &models.Server{
		ID:         id,
		Name:       fmt.Sprintf("Server %d", id),
		Port:       2302 + int(id-1),
		Status:     "running",
		ConfigPath: filepath.Join("/opt/dayz", fmt.Sprintf("server_%d", id)),
	}, nil
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	response := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}
	writeJSONResponse(w, status, response)
}

func getBackupTimestamp(backupPath string) int64 {
	pattern := `\.bak\.(\d+)$`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(backupPath)
	if len(matches) > 1 {
		var timestamp int64
		fmt.Sscanf(matches[1], "%d", &timestamp)
		return timestamp
	}
	return 0
}
