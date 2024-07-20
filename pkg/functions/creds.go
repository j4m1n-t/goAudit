package internal

import (
	// Standard library
	"fmt"

	// Fyne Imports

	// External Imports
	"golang.org/x/crypto/bcrypt"

	// Internal Imports
	myAuth "github.com/j4m1n-t/goAudit/pkg/authentication"
)

type User struct {
	ID              string
	Username        string
	IsAuthenticated bool
}

// credentials storing similar to bitwarden

func MasterPasswordSetup() {
	print("What is your username?  ")
	var userID string
	fmt.Scanln(&userID)
	print("What is your password?  ")
	var pw string
	fmt.Scanln(&pw)
	myAuth.SetMasterPassword(userID, pw)
}

func MasterPasswordLogin(user User, masterPassword string) bool {
	storedHash := GetStoredMasterPasswordHash(user)
	return bcrypt.CompareHashAndPassword(storedHash, []byte(masterPassword)) == nil
}

func GetStoredMasterPasswordHash(user User) []byte {
	// Retrieve hashed master password from secure storage
	// This is a simplified example; use proper secure storage in practice
	return []byte("hashed_master_password")
}

func PromptMasterPassword() string {
	// Prompt the user for their master password
	// and return it as a string
	return "your_master_password_here"
}

func DisplayCredentials(user User, encryptionKey []byte) {
	//CRUD.GetCredentials()

}

func AccessCredentialsTab(user User) {
	if !user.IsAuthenticated {
		// Redirect to LDAP login
		return
	}

	// Prompt for master password
	masterPassword := PromptMasterPassword()

	if !myAuth.VerifyMasterPassword(user.Username, masterPassword) {
		// Display error and deny access
		return
	}

	// Derive encryption key from master password
	encryptionKey := DeriveKey(masterPassword)

	// Use encryptionKey to decrypt and display credentials
	DisplayCredentials(user, encryptionKey)
}

func DeriveKey(masterPassword string) []byte {
	argon2 := make([]byte, len(masterPassword))
	copy(argon2, masterPassword)
	return argon2
}
