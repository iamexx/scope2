package models

import "time"

type AdminUser struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type Server struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Port       int       `json:"port"`
	Status     string    `json:"status"`
	ConfigPath string    `json:"config_path"`
	CreatedAt  time.Time `json:"created_at"`
}

type FTPUser struct {
	ID           int64     `json:"id"`
	ServerID     int64     `json:"server_id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	HomeDir      string    `json:"home_dir"`
	CreatedAt    time.Time `json:"created_at"`
}

type SteamCMDInfo struct {
	ID               int64     `json:"id"`
	LastSyncDate     time.Time `json:"last_sync_date"`
	CurrentVersion   string    `json:"current_version"`
	CentralFilesPath string    `json:"central_files_path"`
}
