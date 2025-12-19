package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	// Expand ~ if needed
	if strings.HasPrefix(dbPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %v", err)
		}
		dbPath = filepath.Join(homeDir, dbPath[2:])
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Connection pool setup
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)

	return createTables()
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INTEGER PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			port INTEGER UNIQUE NOT NULL,
			status TEXT DEFAULT 'stopped',
			config_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS ftp_users (
			id INTEGER PRIMARY KEY,
			server_id INTEGER NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			home_dir TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(server_id) REFERENCES servers(id)
		);`,
		`CREATE TABLE IF NOT EXISTS steamcmd_info (
			id INTEGER PRIMARY KEY,
			last_sync_date DATETIME,
			current_version TEXT,
			central_files_path TEXT
		);`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("error creating table: %v", err)
		}
	}
	return nil
}
