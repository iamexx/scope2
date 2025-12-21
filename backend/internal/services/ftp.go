package services

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os/exec"
	"strings"
	"time"

	"github.com/iamexx/scope2-dayz-api/internal/db"
	"github.com/iamexx/scope2-dayz-api/internal/models"
)

type FTPService struct {
	db          *sql.DB
	authService *AuthService
}

func NewFTPService(authService *AuthService) *FTPService {
	return &FTPService{
		db:          db.DB,
		authService: authService,
	}
}

// GenerateRandomPassword generates a random password of given length
func (s *FTPService) GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

// CreateFTPUser creates an OS user and a DB record
func (s *FTPService) CreateFTPUser(serverID int64) (*models.FTPUser, string, error) {
	// Get server info
	var serverName string
	err := s.db.QueryRow("SELECT name FROM servers WHERE id = ?", serverID).Scan(&serverName)
	if err != nil {
		return nil, "", fmt.Errorf("server not found: %v", err)
	}

	username := fmt.Sprintf("dayz_%s", strings.ToLower(serverName))
	// Clean username of spaces just in case
	username = strings.ReplaceAll(username, " ", "_")
	
	password, err := s.GenerateRandomPassword(16)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate password: %v", err)
	}

	homeDir := fmt.Sprintf("/srv/dayz-servers/%s", serverName)

	// Check if user already exists in DB
	var exists int
	s.db.QueryRow("SELECT COUNT(*) FROM ftp_users WHERE server_id = ?", serverID).Scan(&exists)
	if exists > 0 {
		return nil, "", fmt.Errorf("ftp user for this server already exists")
	}

	// Check if user exists in OS and delete if necessary (cleanup)
	if err := exec.Command("id", "-u", username).Run(); err == nil {
		exec.Command("userdel", "-f", username).Run()
	}

	// Create OS user
	// -m creates the home directory if it doesn't exist
	cmd := exec.Command("useradd", "-m", "-s", "/usr/sbin/nologin", "-d", homeDir, username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("failed to create OS user: %v, output: %s", err, string(output))
	}

	// Set password
	cmd = exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, password))
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("failed to set password: %v, output: %s", err, string(output))
	}

	// Set permissions
	if err := exec.Command("chmod", "755", homeDir).Run(); err != nil {
		log.Printf("Warning: failed to chmod home dir: %v", err)
	}
	
	// Ensure user owns their directory
    // chown -R username:username homeDir
    // Note: We might want the server process user (steam/dayz) to also have access?
    // Assuming dayz server runs as a specific user or root. 
    // If it runs as root, it's fine. If it runs as 'steam', we might have permissions issues.
    // The instructions don't specify the server runner user, but typically it is 'steam' or similar.
    // Given we just created a user 'dayz_{name}', maybe the server should run as this user? 
    // But ProcessManager just runs the binary.
    
    // For now, let's just make sure the new user owns the dir.
	if err := exec.Command("chown", "-R", fmt.Sprintf("%s:%s", username, username), homeDir).Run(); err != nil {
		log.Printf("Warning: failed to chown home dir: %v", err)
	}

	// Hash password for DB
	hash, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %v", err)
	}

	// Store in DB
	res, err := s.db.Exec("INSERT INTO ftp_users (server_id, username, password_hash, home_dir) VALUES (?, ?, ?, ?)",
		serverID, username, hash, homeDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to insert ftp user into db: %v", err)
	}

	id, _ := res.LastInsertId()

	return &models.FTPUser{
		ID:           id,
		ServerID:     serverID,
		Username:     username,
		PasswordHash: hash,
		HomeDir:      homeDir,
		CreatedAt:    time.Now(),
	}, password, nil
}

// DeleteFTPUser removes the OS user and DB record
func (s *FTPService) DeleteFTPUser(serverID int64) error {
	var username string
	err := s.db.QueryRow("SELECT username FROM ftp_users WHERE server_id = ?", serverID).Scan(&username)
	if err == sql.ErrNoRows {
		return nil // Already deleted
	}
	if err != nil {
		return fmt.Errorf("db error: %v", err)
	}

	// Delete from OS
	if output, err := exec.Command("userdel", "-f", username).CombinedOutput(); err != nil {
		log.Printf("Warning: failed to delete OS user %s: %v, output: %s", username, err, string(output))
	}

	// Delete from DB
	_, err = s.db.Exec("DELETE FROM ftp_users WHERE server_id = ?", serverID)
	return err
}

// RegeneratePassword generates a new password, updates OS user and DB
func (s *FTPService) RegeneratePassword(serverID int64) (string, error) {
	var username string
	err := s.db.QueryRow("SELECT username FROM ftp_users WHERE server_id = ?", serverID).Scan(&username)
	if err != nil {
		return "", fmt.Errorf("ftp user not found: %v", err)
	}

	newPassword, err := s.GenerateRandomPassword(16)
	if err != nil {
		return "", err
	}

	// Update OS
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, newPassword))
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to update OS password: %v, output: %s", err, string(output))
	}

	// Update DB
	hash, err := s.authService.HashPassword(newPassword)
	if err != nil {
		return "", err
	}

	_, err = s.db.Exec("UPDATE ftp_users SET password_hash = ? WHERE server_id = ?", hash, serverID)
	if err != nil {
		return "", fmt.Errorf("failed to update db: %v", err)
	}

	return newPassword, nil
}

// GetCredentials returns the FTP user details
func (s *FTPService) GetCredentials(serverID int64) (*models.FTPUser, error) {
	var user models.FTPUser
	err := s.db.QueryRow("SELECT id, server_id, username, home_dir, created_at FROM ftp_users WHERE server_id = ?", serverID).Scan(
		&user.ID, &user.ServerID, &user.Username, &user.HomeDir, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidateUser checks if the user exists in both DB and OS
func (s *FTPService) ValidateUser(serverID int64) (bool, error) {
	var username string
	err := s.db.QueryRow("SELECT username FROM ftp_users WHERE server_id = ?", serverID).Scan(&username)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Check OS
	if err := exec.Command("id", "-u", username).Run(); err != nil {
		return false, nil
	}

	return true, nil
}
