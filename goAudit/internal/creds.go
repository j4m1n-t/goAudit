package internal

import (
	"context"

	"golang.org/x/crypto/bcrypt"

    serverSide "github.com/j4m1n-t/goAudit/goAuditServer/Internal"
)

type User struct {
    ID       string
    Username string
    IsAuthenticated bool
}

// credentials storing similar to bitwarden

func setup(ctx context.Context) {

}

func login(user User, masterPassword string) bool {
    storedHash := getStoredMasterPasswordHash(user)
    return bcrypt.CompareHashAndPassword(storedHash, []byte(masterPassword)) == nil
}

func getStoredMasterPasswordHash(user User) []byte {
    // Retrieve the stored master password hash from a secure storage
    // (e.g., a database, a key vault, or a file)
}

func promptMasterPassword() string {
    // Prompt the user for their master password
    // and return it as a string
}

func displayCredentials(user User, encryptionKey []byte) {
    // Decrypt and display credentials using the provided encryption key
    // (e.g., decrypt JSON files, decrypt password fields, etc.)
}



func accessCredentialsTab(user User) {
    if !user.IsAuthenticated {
        // Redirect to LDAP login
        return
    }

    // Prompt for master password
    masterPassword := promptMasterPassword()

    if !serverSide.VerifyMasterPassword(user.ID, masterPassword) {
        // Display error and deny access
        return
    }

    // Derive encryption key from master password
    encryptionKey := deriveKey(masterPassword)

    // Use encryptionKey to decrypt and display credentials
    displayCredentials(user, encryptionKey)
}

func verifyMasterPassword(user User, masterPassword string) bool {
    storedHash := getStoredMasterPasswordHash(user)
    return bcrypt.CompareHashAndPassword(storedHash, []byte(masterPassword)) == nil
}

func deriveKey(masterPassword string) []byte {
    // Use a key derivation function like Argon2 or PBKDF2
    // to derive an encryption key from the master password
}