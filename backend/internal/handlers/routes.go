package handlers

import (
    "net/http"

    "github.com/gorilla/mux"
)

// RegisterRoutes registers all server-related routes
func (sh *ServerHandler) RegisterRoutes(router *mux.Router) {
    // Server CRUD operations
    router.HandleFunc("", sh.CreateServer).Methods(http.MethodPost)
    router.HandleFunc("", sh.ListServers).Methods(http.MethodGet)
    router.HandleFunc("/{id}", sh.GetServerDetails).Methods(http.MethodGet)
    router.HandleFunc("/{id}", sh.DeleteServer).Methods(http.MethodDelete)

    // Server process management
    router.HandleFunc("/{id}/start", sh.StartServer).Methods(http.MethodPost)
    router.HandleFunc("/{id}/stop", sh.StopServer).Methods(http.MethodPost)
    router.HandleFunc("/{id}/restart", sh.RestartServer).Methods(http.MethodPost)
    router.HandleFunc("/{id}/status", sh.GetServerStatus).Methods(http.MethodGet)

    // Server logging
    router.HandleFunc("/{id}/logs", sh.GetServerLogs).Methods(http.MethodGet)
    router.HandleFunc("/{id}/logs/admin", sh.GetAdminLogs).Methods(http.MethodGet)

    // Server configuration
    router.HandleFunc("/{id}/config", sh.GetServerConfig).Methods(http.MethodGet)
    router.HandleFunc("/{id}/config", sh.UpdateServerConfig).Methods(http.MethodPut)

    // FTP user management
    router.HandleFunc("/{id}/ftp/create", sh.CreateFTPUser).Methods(http.MethodPost)
    router.HandleFunc("/{id}/ftp/credentials", sh.GetFTPCredentials).Methods(http.MethodGet)
    router.HandleFunc("/{id}/ftp/regenerate-password", sh.RegenerateFTPPassword).Methods(http.MethodPost)
    router.HandleFunc("/{id}/ftp/user", sh.DeleteFTPUser).Methods(http.MethodDelete)
    router.HandleFunc("/{id}/ftp/status", sh.GetFTPStatus).Methods(http.MethodGet)
}

// SetupServerRoutes sets up a router with all server routes and middleware
func (sh *ServerHandler) SetupServerRoutes() *mux.Router {
    router := mux.NewRouter()
    
    // Register all server routes
    sh.RegisterRoutes(router)
    
    return router
}