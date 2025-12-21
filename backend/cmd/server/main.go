package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/iamexx/scope2-dayz-api/internal/db"
	"github.com/iamexx/scope2-dayz-api/internal/handlers"
	"github.com/iamexx/scope2-dayz-api/internal/middleware"
	"github.com/iamexx/scope2-dayz-api/internal/services"
)

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

func steamcmdStatusHandler(steamCmdService *services.SteamCMDService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		status, err := steamCmdService.GetStatus()
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"version":     status["version"],
			"lastSync":    status["lastSync"],
			"centralPath": status["centralPath"],
			"status":      status["status"],
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func steamcmdSyncHandler(jobManager *services.JobManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		jobID, err := jobManager.StartSyncJob()
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"jobId":  jobID,
			"status": "running",
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}
}

func steamcmdProgressHandler(jobManager *services.JobManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		jobID := r.URL.Query().Get("jobId")
		if jobID == "" {
			http.Error(w, `{"error": "jobId parameter is required"}`, http.StatusBadRequest)
			return
		}

		job, err := jobManager.GetJobStatus(jobID)
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}

		response := map[string]interface{}{
			"status":   job.Status,
			"progress": job.Progress,
			"message":  job.Message,
			"jobId":    job.ID,
		}

		w.WriteHeader(http.StatusOK)
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

	// Initialize services
	authService := services.NewAuthService()
	processManager := services.NewProcessManager()
	steamCmdService := services.NewSteamCMDService(db.DB)
	jobManager := services.NewJobManager(steamCmdService)
	ftpService := services.NewFTPService(authService)

	// Start job cleanup routine
	jobManager.StartCleanupRoutine()

	// Check for first run
	if err := steamCmdService.CheckFirstRun(); err != nil {
		log.Printf("Warning: Failed first run check: %v", err)
	}

	// Initialize server handler
	serverHandler := handlers.NewServerHandler(processManager, authService, ftpService)

	// Create main router
	mainRouter := mux.NewRouter()

	// Apply global middleware
	mainRouter.Use(middleware.LoggingMiddleware)
	mainRouter.Use(middleware.ErrorMiddleware)

	// Public routes (no authentication required)
	mainRouter.HandleFunc("/api/health", healthHandler).Methods(http.MethodGet)
	mainRouter.HandleFunc("/api/auth/setup", setupHandler(authService)).Methods(http.MethodPost)
	mainRouter.HandleFunc("/api/auth/login", loginHandler(authService)).Methods(http.MethodPost)

	// Create JWT middleware
	jwtMiddleware := middleware.JWTMiddleware(authService)

	// User management (protected by JWT)
	meHandlerFunc := meHandler(authService)
	mainRouter.Handle("/api/auth/me", jwtMiddleware(http.HandlerFunc(meHandlerFunc))).Methods(http.MethodGet)

	// SteamCMD API endpoints (protected by JWT)
	mainRouter.Handle("/api/steamcmd/status", jwtMiddleware(http.HandlerFunc(steamcmdStatusHandler(steamCmdService)))).Methods(http.MethodGet)
	mainRouter.Handle("/api/steamcmd/sync", jwtMiddleware(http.HandlerFunc(steamcmdSyncHandler(jobManager)))).Methods(http.MethodPost)
	mainRouter.Handle("/api/steamcmd/sync/progress", jwtMiddleware(http.HandlerFunc(steamcmdProgressHandler(jobManager)))).Methods(http.MethodGet)

	// Server management routes (protected by JWT)
	serverRouter := mainRouter.PathPrefix("/api/servers").Subrouter()
	serverRouter.Use(jwtMiddleware)
	serverHandler.RegisterRoutes(serverRouter)

	// Start server on port 8080
	serverAddr := ":8080"
	log.Printf("Starting server on %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, mainRouter); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
