package auth

import (
	// Standard Library
	"context"
	"fmt"
	"sync"
	"time"

	// External Imports
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	// Internal Imports
	"github.com/j4m1n-t/goAudit/internal/interfaces"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

var (
	masterPasswords = make(map[string][]byte)
	mu              sync.RWMutex
	dbPool          *pgxpool.Pool
)

func InitDB(connString string) error {
	var err error
	dbPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	return nil
}

func SetMasterPassword(userID string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	mu.Lock()
	masterPasswords[userID] = hashedPassword
	mu.Unlock()

	// Update the database
	_, err = dbPool.Exec(context.Background(),
		"UPDATE credentials SET master_password = $1 WHERE user_id = $2",
		hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update master password in database: %v", err)
	}

	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func VerifyMasterPassword(username string, password string) bool {
	mu.RLock()
	hashedPassword, exists := masterPasswords[username]
	mu.RUnlock()

	if !exists {
		// If not in memory, check the database
		var dbHashedPassword []byte
		err := dbPool.QueryRow(context.Background(),
			"SELECT master_password FROM credentials WHERE user_id = $1", username).Scan(&dbHashedPassword)
		if err != nil {
			return false
		}
		hashedPassword = dbHashedPassword
	}

	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil
}

// func SaveMasterPasswords() error {
// 	mu.RLock()
// 	defer mu.RUnlock()
// 	// Save to encrypted file or secure storage
// 	// This is a simplified example; use proper encryption in practice
// 	file, err := os.Create("master_passwords.enc")
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
// 	encoder := json.NewEncoder(file)
// 	return encoder.Encode(masterPasswords)
// }

// func LoadMasterPasswords() error {
// 	// Load from encrypted file or secure storage
// 	file, err := os.Open("master_passwords.enc")
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
// 	decoder := json.NewDecoder(file)
// 	return decoder.Decode(&masterPasswords)
// }

func SyncMasterPasswords() error {
	mu.RLock()
	defer mu.RUnlock()

	for userID, hashedPassword := range masterPasswords {
		_, err := dbPool.Exec(context.Background(),
			"UPDATE users SET master_password = $1 WHERE user_id = $2",
			hashedPassword, userID)
		if err != nil {
			return fmt.Errorf("failed to sync master password for user %s: %v", userID, err)
		}
	}
	return nil
}

// New function to handle credential-specific login
func VerifyCredentialLogin(loginName, loginPass string) (*interfaces.Credentials, error) {
	var cred interfaces.Credentials
	err := dbPool.QueryRow(context.Background(),
		`SELECT id, site, program, user_id, username, master_password, login_name, login_pass, created_at, updated_at, owner
         FROM credentials WHERE login_name = $1`, loginName).
		Scan(&cred.ID, &cred.Site, &cred.Program, &cred.UserID, &cred.Username, &cred.MasterPassword,
			&cred.LoginName, &cred.LoginPass, &cred.CreatedAt, &cred.UpdatedAt, &cred.Owner)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %v", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(cred.LoginPass), []byte(loginPass)) != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &cred, nil
}

// New function to create a credential
func CreateCredential(cred *interfaces.Credentials) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(cred.LoginPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	_, err = dbPool.Exec(context.Background(),
		`INSERT INTO credentials (site, program, user_id, username, master_password, login_name, login_pass, created_at, updated_at, owner)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		cred.Site, cred.Program, cred.UserID, cred.Username, cred.MasterPassword, cred.LoginName, string(hashedPass),
		time.Now(), time.Now(), cred.Owner)
	if err != nil {
		return fmt.Errorf("failed to create credential: %v", err)
	}

	return nil
}

// This needs fixed
func CheckIfMPPresent(appState *state.AppState) error {
	var count int
	err := appState.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM credentials WHERE username = $1 AND master_password IS NOT NULL", appState.Username).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check master password presence: %v", err)
	}

	appState.MPPresent = count > 0
	return nil
}
