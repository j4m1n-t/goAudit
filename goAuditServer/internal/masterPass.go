package serverinternal

import (
	"database/sql"
	"encoding/json"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	masterPasswords = make(map[string][]byte)
	mu              sync.RWMutex
)

func setMasterPassword(userID string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	mu.Lock()
	masterPasswords[userID] = hashedPassword
	mu.Unlock()

	return nil
}

func VerifyMasterPassword(userID string, password string) bool {
	mu.RLock()
	hashedPassword, exists := masterPasswords[userID]
	mu.RUnlock()

	if !exists {
		return false
	}

	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil
}

func saveMasterPasswords() error {
	mu.RLock()
	defer mu.RUnlock()

	// Save to encrypted file or secure storage
	// This is a simplified example; use proper encryption in practice
	file, err := os.Create("master_passwords.enc")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(masterPasswords)
}

func loadMasterPasswords() error {
	// Load from encrypted file or secure storage
	file, err := os.Open("master_passwords.enc")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&masterPasswords)
}

func syncMasterPasswords() error {
	var db *sql.DB
	mu.RLock()
	defer mu.RUnlock()

	for userID, hashedPassword := range masterPasswords {
		// Use a prepared statement for better performance and security
		_, err := db.Exec("UPDATE users SET master_password_hash = $1 WHERE id = $2", hashedPassword, userID)
		if err != nil {
			return err
		}
	}
	return nil
}
