package services

import (
    "crypto/rand"
    "crypto/subtle"
    "database/sql"
    "encoding/base64"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/argon2"
    "github.com/iamexx/scope2-dayz-api/internal/db"
    "github.com/iamexx/scope2-dayz-api/internal/models"
)

var (
    // JWT secret key - in production, use environment variable
    jwtSecret = []byte("scope2-dayz-secret-key-change-in-production")
    
    // JWT expiration time
    jwtExpiration = 24 * time.Hour
)

type AuthService struct {
    db *sql.DB
}

func NewAuthService() *AuthService {
    return &AuthService{
        db: db.DB,
    }
}

// Password hashing parameters for Argon2
const (
    timeCost   = 1
    memoryCost = 64 * 1024 // 64MB
    threads    = 4
    keyLen     = 32
    saltLen    = 16
)

// HashPassword hashes a password using Argon2id
func (a *AuthService) HashPassword(password string) (string, error) {
    salt := make([]byte, saltLen)
    if _, err := rand.Read(salt); err != nil {
        return "", fmt.Errorf("failed to generate salt: %v", err)
    }

    hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, keyLen)

    // Combine salt and hash in a single string: $argon2id$v=19$m=65536,t=1,p=4$base64(salt)$base64(hash)
    saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
    hashBase64 := base64.RawStdEncoding.EncodeToString(hash)
    
    return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", 
        memoryCost, timeCost, threads, saltBase64, hashBase64), nil
}

// VerifyPassword verifies a password against a hash
func (a *AuthService) VerifyPassword(password, hash string) bool {
    // Parse the hash format
    parts := strings.Split(hash, "$")
    if len(parts) != 6 || parts[1] != "argon2id" {
        return false
    }

    // Extract parameters
    var memCost, tCost uint32
    var par uint8
    _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memCost, &tCost, &par)
    if err != nil {
        return false
    }

    salt, err := base64.RawStdEncoding.DecodeString(parts[4])
    if err != nil {
        return false
    }

    expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
    if err != nil {
        return false
    }

    // Hash the password with the same parameters
    hashToVerify := argon2.IDKey([]byte(password), salt, tCost, memCost, par, uint32(len(expectedHash)))

    // Use constant time comparison
    return subtle.ConstantTimeCompare(hashToVerify, expectedHash) == 1
}

// GenerateJWT generates a JWT token for a user
func (a *AuthService) GenerateJWT(userID int64, username string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "exp":      time.Now().Add(jwtExpiration).Unix(),
        "iat":      time.Now().Unix(),
    })

    return token.SignedString(jwtSecret)
}

// ValidateJWT validates a JWT token and returns the claims
func (a *AuthService) ValidateJWT(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %v", err)
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // Check expiration
        if exp, ok := claims["exp"].(float64); ok {
            if time.Unix(int64(exp), 0).Before(time.Now()) {
                return nil, errors.New("token has expired")
            }
        }
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

// IsFirstRun checks if this is the first run (no admin user exists)
func (a *AuthService) IsFirstRun() (bool, error) {
    var count int
    err := a.db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&count)
    if err != nil {
        return false, fmt.Errorf("failed to check admin users: %v", err)
    }
    return count == 0, nil
}

// CreateAdminUser creates a new admin user
func (a *AuthService) CreateAdminUser(username, password string) (*models.AdminUser, error) {
    // Check if this is the first run
    isFirstRun, err := a.IsFirstRun()
    if err != nil {
        return nil, err
    }
    if !isFirstRun {
        return nil, errors.New("admin user already exists")
    }

    // Hash the password
    passwordHash, err := a.HashPassword(password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %v", err)
    }

    // Create the admin user
    result, err := a.db.Exec(
        "INSERT INTO admin_users (username, password_hash) VALUES (?, ?)",
        username,
        passwordHash,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create admin user: %v", err)
    }

    userID, err := result.LastInsertId()
    if err != nil {
        return nil, fmt.Errorf("failed to get user ID: %v", err)
    }

    return &models.AdminUser{
        ID:           userID,
        Username:     username,
        PasswordHash: passwordHash,
        CreatedAt:    time.Now(),
    }, nil
}

// AuthenticateUser authenticates a user with username and password
func (a *AuthService) AuthenticateUser(username, password string) (*models.AdminUser, error) {
    var user models.AdminUser
    
    err := a.db.QueryRow(
        "SELECT id, username, password_hash, created_at FROM admin_users WHERE username = ?",
        username,
    ).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
    
    if err == sql.ErrNoRows {
        return nil, errors.New("invalid credentials")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query user: %v", err)
    }

    // Verify password
    if !a.VerifyPassword(password, user.PasswordHash) {
        return nil, errors.New("invalid credentials")
    }

    return &user, nil
}

// GetUserByID retrieves a user by ID
func (a *AuthService) GetUserByID(userID int64) (*models.AdminUser, error) {
    var user models.AdminUser
    
    err := a.db.QueryRow(
        "SELECT id, username, password_hash, created_at FROM admin_users WHERE id = ?",
        userID,
    ).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
    
    if err == sql.ErrNoRows {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query user: %v", err)
    }

    return &user, nil
}