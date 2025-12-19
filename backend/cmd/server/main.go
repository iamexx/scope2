package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/iamexx/scope2-dayz-api/internal/db"
)

type HealthResponse struct {
    Status string `json:"status"`
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
    
    // Register handlers
    http.HandleFunc("/api/health", healthHandler)
    
    // Start server on port 8080
    serverAddr := ":8080"
    log.Printf("Starting server on %s", serverAddr)
    
    if err := http.ListenAndServe(serverAddr, nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}