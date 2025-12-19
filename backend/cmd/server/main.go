package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "regexp"
    "strconv"
    "strings"
    
    "github.com/iamexx/scope2-dayz-api/internal/db"
    "github.com/iamexx/scope2-dayz-api/internal/middleware"
    "github.com/iamexx/scope2-dayz-api/internal/services"
)

func extractServerID(path string) int64 {
    // Extract server ID from path like /api/servers/{id}/...
    re := regexp.MustCompile(`/api/servers/(\d+)/`)
    matches := re.FindStringSubmatch(path)
    if len(matches) < 2 {
        return 0
    }
    id, err := strconv.ParseInt(matches[1], 10, 64)
    if err != nil {
        return 0
    }
    return id
}

func parseServerIDFromPath(path string) string {
    // Extract the numeric part from /api/servers/{id}/...
    parts := strings.Split(path, "/")
    for i, part := range parts {
        if part == "servers" && i+1 < len(parts) {
            return parts[i+1]
        }
    }
    return ""
}

type HealthResponse struct {
    Status string `json:"status"`
}

type SetupRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type SetupResponse struct {
    Token   string `json:"token"`
    Message string `json:"message"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token   string `json:"token"`
    Message string `json:"message"`
}

type MeResponse struct {
    ID        int64  `json:"id"`
    Username  string `json:"username"`
    CreatedAt string `json:"created_at"`
}

func setupHandler(authService *services.AuthService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        var req SetupRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
            return
        }
        
        if req.Username == "" || req.Password == "" {
            http.Error(w, `{"error": "username and password are required"}`, http.StatusBadRequest)
            return
        }
        
        user, err := authService.CreateAdminUser(req.Username, req.Password)
        if err != nil {
            http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
            return
        }
        
        token, err := authService.GenerateJWT(user.ID, user.Username)
        if err != nil {
            http.Error(w, `{"error": "failed to generate token"}`, http.StatusInternalServerError)
            return
        }
        
        response := SetupResponse{
            Token:   token,
            Message: "Admin user created successfully",
        }
        
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(response)
    }
}

func loginHandler(authService *services.AuthService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        var req LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
            return
        }
        
        if req.Username == "" || req.Password == "" {
            http.Error(w, `{"error": "username and password are required"}`, http.StatusBadRequest)
            return
        }
        
        user, err := authService.AuthenticateUser(req.Username, req.Password)
        if err != nil {
            http.Error(w, `{"error": "invalid credentials"}`, http.StatusUnauthorized)
            return
        }
        
        token, err := authService.GenerateJWT(user.ID, user.Username)
        if err != nil {
            http.Error(w, `{"error": "failed to generate token"}`, http.StatusInternalServerError)
            return
        }
        
        response := LoginResponse{
            Token:   token,
            Message: "Login successful",
        }
        
        json.NewEncoder(w).Encode(response)
    }
}

func meHandler(authService *services.AuthService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        userID, ok := middleware.GetUserID(r)
        if !ok {
            http.Error(w, `{"error": "user not authenticated"}`, http.StatusUnauthorized)
            return
        }
        
        user, err := authService.GetUserByID(userID)
        if err != nil {
            http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
            return
        }
        
        response := MeResponse{
            ID:        user.ID,
            Username:  user.Username,
            CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        }
        
        json.NewEncoder(w).Encode(response)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    
    response := HealthResponse{
        Status: "ok",
    }
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
}

func main() {
    // Set up logging
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    // Initialize database
    if err := db.InitDB("~/.dayz/dayz.db"); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    log.Println("Database initialized successfully")
    
    // Initialize auth service
    authService := services.NewAuthService()
    
    // Create JWT middleware
    jwtMiddleware := middleware.JWTMiddleware(authService)
    
    // Register handlers
    http.HandleFunc("/api/health", healthHandler)
    http.HandleFunc("/api/auth/setup", setupHandler(authService))
    http.HandleFunc("/api/auth/login", loginHandler(authService))
    
    // Protected routes
    meHandler := meHandler(authService)
    http.Handle("/api/auth/me", jwtMiddleware(http.HandlerFunc(meHandler)))
    
    // Initialize server handlers
    serverHandlers := services.NewServerHandlers()
    
    // Server routes with JWT protection
    http.Handle("/api/servers/", jwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        mux := http.NewServeMux()
        
        // Extract server ID from URL
        serverID := extractServerID(r.URL.Path)
        if serverID == 0 {
            http.Error(w, `{"error": "invalid server ID"}`, http.StatusBadRequest)
            return
        }
        
        userID, _ := middleware.GetUserID(r)
        userIDStr := fmt.Sprintf("%d", userID)
        
        // Define route handlers
        mux.HandleFunc("/api/servers/{id}/logs", func(w http.ResponseWriter, r *http.Request) {
            if r.Method != http.MethodGet {
                http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
                return
            }
            serverHandlers.GetLogs(w, r, serverID)
        })
        
        mux.HandleFunc("/api/servers/{id}/logs/admin", func(w http.ResponseWriter, r *http.Request) {
            if r.Method != http.MethodGet {
                http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
                return
            }
            serverHandlers.GetAdminLogs(w, r, serverID)
        })
        
        mux.HandleFunc("/api/servers/{id}/config", func(w http.ResponseWriter, r *http.Request) {
            switch r.Method {
            case http.MethodGet:
                serverHandlers.GetConfig(w, r, serverID)
            case http.MethodPut:
                serverHandlers.UpdateConfig(w, r, serverID, userIDStr)
            default:
                http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
            }
        })
        
        mux.HandleFunc("/api/servers/{id}/config/backup", func(w http.ResponseWriter, r *http.Request) {
            switch r.Method {
            case http.MethodGet:
                serverHandlers.GetConfigBackup(w, r, serverID)
            default:
                http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
            }
        })
        
        mux.ServeHTTP(w, r)
    })))
    
    // Start server on port 8080
    serverAddr := ":8080"
    log.Printf("Starting server on %s", serverAddr)
    
    if err := http.ListenAndServe(serverAddr, nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}