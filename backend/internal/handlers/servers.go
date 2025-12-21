package handlers

import (
    "bufio"
    "crypto/rand"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/gorilla/mux"
    "github.com/iamexx/scope2-dayz-api/internal/db"
    "github.com/iamexx/scope2-dayz-api/internal/middleware"
    "github.com/iamexx/scope2-dayz-api/internal/models"
    "github.com/iamexx/scope2-dayz-api/internal/services"
)

// Standard response format for all endpoints
type APIResponse struct {
    Status  string      `json:"status"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// Server-related request/response types
type CreateServerRequest struct {
    Name string `json:"name"`
    Port int    `json:"port"`
}

type CreateServerResponse struct {
    ID      int64  `json:"id"`
    Name    string `json:"name"`
    Port    int    `json:"port"`
    Status  string `json:"status"`
    Message string `json:"message"`
}

type ServerDetailsResponse struct {
    ID         int64     `json:"id"`
    Name       string    `json:"name"`
    Port       int       `json:"port"`
    Status     string    `json:"status"`
    ConfigPath string    `json:"config_path"`
    CreatedAt  string    `json:"created_at"`
    FTPUser    *FTPUserResponse `json:"ftp_user,omitempty"`
}

type FTPUserResponse struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    HomeDir  string `json:"home_dir"`
}

type ServerStatusResponse struct {
    Status  string `json:"status"`
    PID     int    `json:"pid,omitempty"`
    Uptime  int64  `json:"uptime,omitempty"`
    Message string `json:"message,omitempty"`
}

type CreateFTPUserRequest struct {
    Username string `json:"username"`
}

type CreateFTPUserResponse struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Password string `json:"password"`
    HomeDir  string `json:"home_dir"`
    Message  string `json:"message"`
}

type FTPUserCredentialsResponse struct {
    Username string `json:"username"`
    Host     string `json:"host"`
    Port     int    `json:"port"`
    HomeDir  string `json:"home_dir"`
}

type RegeneratePasswordResponse struct {
    Password string `json:"password"`
    Message  string `json:"message"`
}

type LogsResponse struct {
    Logs    string `json:"logs"`
    Message string `json:"message"`
}

type ServerConfigResponse struct {
    Config string `json:"config"`
}

type UpdateConfigRequest struct {
    Config string `json:"config"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

// ServerHandler handles all server-related operations
type ServerHandler struct {
    processManager *services.ProcessManager
    authService    *services.AuthService
    ftpService     *services.FTPService
    logger         *log.Logger
}

// NewServerHandler creates a new ServerHandler instance
func NewServerHandler(processManager *services.ProcessManager, authService *services.AuthService, ftpService *services.FTPService) *ServerHandler {
    return &ServerHandler{
        processManager: processManager,
        authService:    authService,
        ftpService:     ftpService,
        logger:         log.New(os.Stdout, "[SERVER-HANDLER] ", log.LstdFlags|log.Lshortfile),
    }
}

// writeJSONResponse writes a standardized JSON response
func (sh *ServerHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        sh.logger.Printf("Failed to encode JSON response: %v", err)
        http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
    }
}

// logRequest logs API requests for audit trail
func (sh *ServerHandler) logRequest(r *http.Request, statusCode int, userID int64) {
    sh.logger.Printf("API Request - Method: %s, Path: %s, UserID: %d, Status: %d", 
        r.Method, r.URL.Path, userID, statusCode)
}

// handleError handles different types of errors and returns appropriate responses
func (sh *ServerHandler) handleError(w http.ResponseWriter, r *http.Request, err error, userID int64) {
    sh.logger.Printf("Error: %v", err)
    
    var statusCode int
    var errorMsg string
    
    switch {
    case strings.Contains(err.Error(), "not found"):
        statusCode = http.StatusNotFound
        errorMsg = "Resource not found"
    case strings.Contains(err.Error(), "already exists"):
        statusCode = http.StatusConflict
        errorMsg = "Resource already exists"
    case strings.Contains(err.Error(), "invalid"):
        statusCode = http.StatusBadRequest
        errorMsg = "Invalid request"
    case strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "unauthorized"):
        statusCode = http.StatusUnauthorized
        errorMsg = "Unauthorized access"
    case strings.Contains(err.Error(), "binary not found") || strings.Contains(err.Error(), "config file not found"):
        statusCode = http.StatusBadRequest
        errorMsg = "Server setup incomplete"
    default:
        statusCode = http.StatusInternalServerError
        errorMsg = "Internal server error"
    }
    
    sh.writeJSONResponse(w, statusCode, APIResponse{
        Status: "error",
        Error:  fmt.Sprintf("%s: %v", errorMsg, err),
    })
}

// CreateServer provisions a new server
func (sh *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    var req CreateServerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid request body: %v", err), userID)
        return
    }

    if req.Name == "" {
        sh.handleError(w, r, fmt.Errorf("server name is required"), userID)
        return
    }

    if req.Port <= 0 || req.Port > 65535 {
        sh.handleError(w, r, fmt.Errorf("invalid port number"), userID)
        return
    }

    // Check if server with same name or port already exists
    var exists int
    err := db.DB.QueryRow(
        "SELECT COUNT(*) FROM servers WHERE name = ? OR port = ?", 
        req.Name, req.Port,
    ).Scan(&exists)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("database error: %v", err), userID)
        return
    }
    if exists > 0 {
        sh.handleError(w, r, fmt.Errorf("server with same name or port already exists"), userID)
        return
    }

    // Create server in database
    result, err := db.DB.Exec(
        "INSERT INTO servers (name, port, status) VALUES (?, ?, ?)",
        req.Name, req.Port, "stopped",
    )
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to create server: %v", err), userID)
        return
    }

    serverID, err := result.LastInsertId()
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to get server ID: %v", err), userID)
        return
    }

    // Create FTP user for the server
    ftpUsername := fmt.Sprintf("dayz_%s", req.Name)
    ftpPassword, err := sh.generateRandomPassword()
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to generate FTP password: %v", err), userID)
        return
    }

    ftpPasswordHash, err := sh.authService.HashPassword(ftpPassword)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to hash FTP password: %v", err), userID)
        return
    }

    serverFolder := fmt.Sprintf("/srv/dayz-servers/%s", req.Name)
    homeDir := filepath.Join(serverFolder, "files")
    
    _, err = db.DB.Exec(
        "INSERT INTO ftp_users (server_id, username, password_hash, home_dir) VALUES (?, ?, ?, ?)",
        serverID, ftpUsername, ftpPasswordHash, homeDir,
    )
    if err != nil {
        sh.logger.Printf("Warning: Failed to create FTP user: %v", err)
        // Don't fail the entire request, just log the warning
    }

    sh.writeJSONResponse(w, http.StatusCreated, APIResponse{
        Status: "success",
        Data: CreateServerResponse{
            ID:      serverID,
            Name:    req.Name,
            Port:    req.Port,
            Status:  "stopped",
            Message: "Server created successfully",
        },
    })

    sh.logRequest(r, http.StatusCreated, userID)
}

// ListServers returns all servers
func (sh *ServerHandler) ListServers(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    rows, err := db.DB.Query(
        "SELECT id, name, port, status, config_path, created_at FROM servers ORDER BY created_at DESC",
    )
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("database error: %v", err), userID)
        return
    }
    defer rows.Close()

    var servers []models.Server
    for rows.Next() {
        var server models.Server
        err := rows.Scan(
            &server.ID, &server.Name, &server.Port, &server.Status, 
            &server.ConfigPath, &server.CreatedAt,
        )
        if err != nil {
            sh.handleError(w, r, fmt.Errorf("failed to scan server row: %v", err), userID)
            return
        }
        servers = append(servers, server)
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data:   servers,
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// GetServerDetails returns detailed information about a specific server
func (sh *ServerHandler) GetServerDetails(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Get server details
    var server models.Server
    err = db.DB.QueryRow(
        "SELECT id, name, port, status, config_path, created_at FROM servers WHERE id = ?",
        serverID,
    ).Scan(&server.ID, &server.Name, &server.Port, &server.Status, 
        &server.ConfigPath, &server.CreatedAt)
    
    if err == sql.ErrNoRows {
        sh.handleError(w, r, fmt.Errorf("server not found"), userID)
        return
    }
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("database error: %v", err), userID)
        return
    }

    // Get FTP user info if exists
    var ftpUser *FTPUserResponse
    var ftpUserRow models.FTPUser
    err = db.DB.QueryRow(
        "SELECT id, username, home_dir FROM ftp_users WHERE server_id = ? LIMIT 1",
        serverID,
    ).Scan(&ftpUserRow.ID, &ftpUserRow.Username, &ftpUserRow.HomeDir)
    
    if err == nil {
        ftpUser = &FTPUserResponse{
            ID:       ftpUserRow.ID,
            Username: ftpUserRow.Username,
            HomeDir:  ftpUserRow.HomeDir,
        }
    }

    response := ServerDetailsResponse{
        ID:         server.ID,
        Name:       server.Name,
        Port:       server.Port,
        Status:     server.Status,
        ConfigPath: server.ConfigPath,
        CreatedAt:  server.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        FTPUser:    ftpUser,
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data:   response,
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// DeleteServer deletes a server and its associated FTP user
func (sh *ServerHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Check if server exists
    var serverName string
    err = db.DB.QueryRow("SELECT name FROM servers WHERE id = ?", serverID).Scan(&serverName)
    if err == sql.ErrNoRows {
        sh.handleError(w, r, fmt.Errorf("server not found"), userID)
        return
    }
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("database error: %v", err), userID)
        return
    }

    // Check if server is running
    status, _, _, _, err := sh.processManager.GetStatus(serverID)
    if err == nil && status == "running" {
        sh.handleError(w, r, fmt.Errorf("cannot delete running server"), userID)
        return
    }

    // Delete FTP user
    _, err = db.DB.Exec("DELETE FROM ftp_users WHERE server_id = ?", serverID)
    if err != nil {
        sh.logger.Printf("Warning: Failed to delete FTP user: %v", err)
    }

    // Delete server
    _, err = db.DB.Exec("DELETE FROM servers WHERE id = ?", serverID)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to delete server: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status:  "success",
        Message: "Server deleted successfully",
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// StartServer starts a DayZ server
func (sh *ServerHandler) StartServer(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Get server details
    server, err := sh.getServerByID(serverID)
    if err != nil {
        sh.handleError(w, r, err, userID)
        return
    }

    serverFolder := fmt.Sprintf("/srv/dayz-servers/%s", server.Name)

    pid, err := sh.processManager.StartServer(serverID, server, serverFolder)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to start server: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data: ServerStatusResponse{
            Status:  "running",
            PID:     pid,
            Message: "Server started successfully",
        },
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// StopServer stops a DayZ server
func (sh *ServerHandler) StopServer(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    err = sh.processManager.StopServer(serverID)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to stop server: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status:  "success",
        Message: "Server stopped successfully",
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// RestartServer restarts a DayZ server
func (sh *ServerHandler) RestartServer(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Get server details
    server, err := sh.getServerByID(serverID)
    if err != nil {
        sh.handleError(w, r, err, userID)
        return
    }

    serverFolder := fmt.Sprintf("/srv/dayz-servers/%s", server.Name)

    err = sh.processManager.RestartServer(serverID, server, serverFolder)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to restart server: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status:  "success",
        Message: "Server restarted successfully",
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// GetServerStatus returns the current status of a server
func (sh *ServerHandler) GetServerStatus(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    status, pid, uptime, message, err := sh.processManager.GetStatus(serverID)
    
    response := ServerStatusResponse{
        Status:  status,
        PID:     pid,
        Uptime:  int64(uptime.Seconds()),
        Message: message,
    }
    if err != nil && message == "" {
        response.Message = err.Error()
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data:   response,
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// GetServerLogs retrieves server logs
func (sh *ServerHandler) GetServerLogs(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Get server details
    server, err := sh.getServerByID(serverID)
    if err != nil {
        sh.handleError(w, r, err, userID)
        return
    }

    queryParams := r.URL.Query()
    lines := queryParams.Get("lines")
    offset := queryParams.Get("offset")

    lineCount := 100 // default
    if lines != "" {
        if parsedLines, err := strconv.Atoi(lines); err != nil {
            sh.handleError(w, r, fmt.Errorf("invalid lines parameter"), userID)
            return
        } else if parsedLines > 0 && parsedLines <= 1000 {
            lineCount = parsedLines
        }
    }

    offsetCount := 0 // default
    if offset != "" {
        if parsedOffset, err := strconv.Atoi(offset); err != nil {
            sh.handleError(w, r, fmt.Errorf("invalid offset parameter"), userID)
            return
        } else if parsedOffset >= 0 {
            offsetCount = parsedOffset
        }
    }

    serverFolder := fmt.Sprintf("/srv/dayz-servers/%s", server.Name)
    logFile := filepath.Join(serverFolder, "files", "server_log.txt")

    // Check if log file exists
    if _, err := os.Stat(logFile); os.IsNotExist(err) {
        sh.writeJSONResponse(w, http.StatusOK, APIResponse{
            Status:  "success",
            Data:    LogsResponse{Logs: "No logs available", Message: "Log file not found"},
        })
        sh.logRequest(r, http.StatusOK, userID)
        return
    }

    logs, err := sh.readLastLines(logFile, lineCount, offsetCount)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to read logs: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data:   LogsResponse{Logs: logs, Message: fmt.Sprintf("Retrieved %d lines", lineCount)},
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// GetAdminLogs retrieves admin logs
func (sh *ServerHandler) GetAdminLogs(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Get server details
    server, err := sh.getServerByID(serverID)
    if err != nil {
        sh.handleError(w, r, err, userID)
        return
    }

    serverFolder := fmt.Sprintf("/srv/dayz-servers/%s", server.Name)
    adminLogFile := filepath.Join(serverFolder, "files", "admin_log.txt")

    // Check if admin log file exists
    if _, err := os.Stat(adminLogFile); os.IsNotExist(err) {
        sh.writeJSONResponse(w, http.StatusOK, APIResponse{
            Status:  "success",
            Data:    LogsResponse{Logs: "No admin logs available", Message: "Admin log file not found"},
        })
        sh.logRequest(r, http.StatusOK, userID)
        return
    }

    logs, err := sh.readLastLines(adminLogFile, 100, 0)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to read admin logs: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data:   LogsResponse{Logs: logs, Message: "Admin logs retrieved"},
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// GetServerConfig retrieves server configuration
func (sh *ServerHandler) GetServerConfig(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Get server details
    server, err := sh.getServerByID(serverID)
    if err != nil {
        sh.handleError(w, r, err, userID)
        return
    }

    serverFolder := fmt.Sprintf("/srv/dayz-servers/%s", server.Name)
    configFile := filepath.Join(serverFolder, "serverDZ.cfg")

    // Check if config file exists
    if _, err := os.Stat(configFile); os.IsNotExist(err) {
        sh.handleError(w, r, fmt.Errorf("config file not found"), userID)
        return
    }

    configContent, err := os.ReadFile(configFile)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to read config file: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data:   ServerConfigResponse{Config: string(configContent)},
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// UpdateServerConfig updates server configuration
func (sh *ServerHandler) UpdateServerConfig(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Get server details
    server, err := sh.getServerByID(serverID)
    if err != nil {
        sh.handleError(w, r, err, userID)
        return
    }

    var req UpdateConfigRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid request body: %v", err), userID)
        return
    }

    if req.Config == "" {
        sh.handleError(w, r, fmt.Errorf("config content is required"), userID)
        return
    }

    serverFolder := fmt.Sprintf("/srv/dayz-servers/%s", server.Name)
    configFile := filepath.Join(serverFolder, "serverDZ.cfg")

    // Update config file
    err = os.WriteFile(configFile, []byte(req.Config), 0644)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to write config file: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status:  "success",
        Message: "Configuration updated successfully",
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// CreateFTPUser creates FTP user for a server
func (sh *ServerHandler) CreateFTPUser(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    var req CreateFTPUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid request body: %v", err), userID)
        return
    }

    if req.Username == "" {
        sh.handleError(w, r, fmt.Errorf("username is required"), userID)
        return
    }

    // Use FTPService to create the FTP user
    ftpUser, password, err := sh.ftpService.CreateFTPUser(serverID)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to create FTP user: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusCreated, APIResponse{
        Status: "success",
        Data: CreateFTPUserResponse{
            ID:       ftpUser.ID,
            Username: ftpUser.Username,
            Password: password,
            HomeDir:  ftpUser.HomeDir,
            Message:  "FTP user created successfully",
        },
    })

    sh.logRequest(r, http.StatusCreated, userID)
}

// GetFTPCredentials returns FTP user credentials
func (sh *ServerHandler) GetFTPCredentials(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Use FTPService to get credentials
    ftpUser, err := sh.ftpService.GetCredentials(serverID)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to get FTP credentials: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data: FTPUserCredentialsResponse{
            Username: ftpUser.Username,
            Host:     "localhost", // This should be configurable
            Port:     21,          // This should be configurable
            HomeDir:  ftpUser.HomeDir,
        },
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// RegenerateFTPPassword regenerates FTP user password
func (sh *ServerHandler) RegenerateFTPPassword(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Use FTPService to regenerate password
    newPassword, err := sh.ftpService.RegeneratePassword(serverID)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to regenerate password: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status: "success",
        Data: RegeneratePasswordResponse{
            Password: newPassword,
            Message:  "Password regenerated successfully",
        },
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// DeleteFTPUser deletes FTP user for a server
func (sh *ServerHandler) DeleteFTPUser(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserID(r)
    if !ok {
        sh.writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
            Status: "error",
            Error:  "User not authenticated",
        })
        return
    }

    vars := mux.Vars(r)
    serverIDStr := vars["id"]
    
    serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("invalid server ID"), userID)
        return
    }

    // Use FTPService to delete FTP user
    err = sh.ftpService.DeleteFTPUser(serverID)
    if err != nil {
        sh.handleError(w, r, fmt.Errorf("failed to delete FTP user: %v", err), userID)
        return
    }

    sh.writeJSONResponse(w, http.StatusOK, APIResponse{
        Status:  "success",
        Message: "FTP user deleted successfully",
    })

    sh.logRequest(r, http.StatusOK, userID)
}

// Helper functions

func (sh *ServerHandler) getServerByID(serverID int64) (*models.Server, error) {
    var server models.Server
    query := "SELECT id, name, port, status, config_path, created_at FROM servers WHERE id = ?"
    err := db.DB.QueryRow(query, serverID).Scan(
        &server.ID, &server.Name, &server.Port, &server.Status, &server.ConfigPath, &server.CreatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("server not found")
    }
    if err != nil {
        return nil, fmt.Errorf("database error: %v", err)
    }
    return &server, nil
}

func (sh *ServerHandler) generateRandomPassword() (string, error) {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    const length = 16
    
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    
    for i := range bytes {
        bytes[i] = charset[int(bytes[i])%len(charset)]
    }
    
    return string(bytes), nil
}

func (sh *ServerHandler) readLastLines(filename string, lines int, offset int) (string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer file.Close()

    // Get file size
    stat, err := file.Stat()
    if err != nil {
        return "", err
    }

    // If file is empty
    if stat.Size() == 0 {
        return "", nil
    }

    // Read file from the end
    file.Seek(0, io.SeekEnd)
    
    var logLines []string
    var lineCount int
    var totalLines int
    
    // First pass: count total lines and find starting position
    for {
        pos, err := file.Seek(-1, io.SeekCurrent)
        if err != nil {
            break
        }
        
        b := make([]byte, 1)
        _, err = file.Read(b)
        if err != nil {
            break
        }
        
        if b[0] == '\n' {
            totalLines++
        }
        
        if pos <= 0 {
            break
        }
        
        // Seek back to continue reading
        file.Seek(-2, io.SeekCurrent)
    }
    
    // Calculate how many lines to skip
    skipLines := totalLines - lines - offset
    if skipLines < 0 {
        skipLines = 0
    }
    
    // Reset to beginning
    file.Seek(0, io.SeekStart)
    
    // Second pass: read the lines we need
    scanner := NewLineScanner(file)
    for scanner.Scan() {
        if skipLines > 0 {
            skipLines--
            continue
        }
        if lineCount >= lines {
            break
        }
        logLines = append(logLines, scanner.Text())
        lineCount++
    }
    
    return strings.Join(logLines, "\n"), nil
}

// Simple line scanner for large files
type LineScanner struct {
    scanner *bufio.Scanner
}

func NewLineScanner(reader io.Reader) *LineScanner {
    scanner := bufio.NewScanner(reader)
    scanner.Split(bufio.ScanLines)
    return &LineScanner{scanner: scanner}
}

func (ls *LineScanner) Scan() bool {
    return ls.scanner.Scan()
}

func (ls *LineScanner) Text() string {
    return ls.scanner.Text()
}