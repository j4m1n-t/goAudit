package internal

import (
    // Standard library
    "fmt"


    // Fyne Imports

    // External Imports
	"golang.org/x/crypto/bcrypt"

    // Internal Imports
	serverSide "github.com/j4m1n-t/goAudit/goAuditServer/pkg"
  //  CRUD "github.com/j4m1n-t/goAudit/goAuditServer/pkg/CRUD"
)

type User struct {
	ID              string
	Username        string
	IsAuthenticated bool
}

// credentials storing similar to bitwarden

func setup() {
    print("What is your username?  ")
    var userID string
    fmt.Scanln(&userID)
    print("What is your password?  ")
    var pw string
    fmt.Scanln(&pw)
    serverSide.SetMasterPassword(userID, pw)
}

func login(user User, masterPassword string) bool {
	storedHash := getStoredMasterPasswordHash(user)
	return bcrypt.CompareHashAndPassword(storedHash, []byte(masterPassword)) == nil
}

func getStoredMasterPasswordHash(user User) []byte {
    // Retrieve hashed master password from secure storage
    // This is a simplified example; use proper secure storage in practice
    return []byte("hashed_master_password")
}

func promptMasterPassword() string {
	// Prompt the user for their master password
	// and return it as a string
    return "your_master_password_here"
}

func displayCredentials(user User, encryptionKey []byte) {
    //CRUD.GetCredentials()

}

func accessCredentialsTab(user User) {
	if !user.IsAuthenticated {
		// Redirect to LDAP login
		return
	}

	// Prompt for master password
	masterPassword := promptMasterPassword()

	if !serverSide.VerifyMasterPassword(user.Username, masterPassword) {
		// Display error and deny access
		return
	}

	// Derive encryption key from master password
	encryptionKey := deriveKey(masterPassword)

	// Use encryptionKey to decrypt and display credentials
	displayCredentials(user, encryptionKey)
}

func deriveKey(masterPassword string) []byte {
    argon2 := make([]byte, len(masterPassword))
    copy(argon2, masterPassword)
    return argon2
}
