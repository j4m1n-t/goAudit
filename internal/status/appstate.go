package state

import (
	// Standard Library
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	// Fyne Imports
	"fyne.io/fyne/v2"
	// External Imports
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	// Internal Imports
	"github.com/j4m1n-t/goAudit/internal/interfaces"
)

type AppState struct {
	LDAPConn             *interfaces.LDAPConnection
	Username             string
	UserID               int
	CredentialAuthStatus bool
	CredentialUsername   string
	MPPresent            bool
	Notes                []interfaces.Note
	Tasks                []interfaces.Tasks
	Audits               []interfaces.Audits
	CRMEntries           []interfaces.CRM
	Credentials          []interfaces.Credentials
	Message              string
	DB                   interfaces.DatabaseOperations
	lw                   interfaces.LDAPOperations
	window               fyne.Window
}

var GlobalState = &AppState{}

// Global State
func (appState *AppState) SetDB(db interfaces.DatabaseOperations) {
	if db == nil {
		log.Fatal("SetDB: Database instance cannot be nil")
	}
	appState.DB = db
}

func (appState *AppState) SetLDAP(lw interfaces.LDAPOperations) {
	if lw == nil {
		log.Fatal("SetDB: Database instance cannot be nil")
	}
	appState.lw = lw
}

func (appState *AppState) SetWindow(window fyne.Window) {
	appState.window = window
}

func (appState *AppState) SetMPPresent() {
	appState.MPPresent = false
}
func (appState *AppState) SetMasterPassword(Username string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	mu.Lock()
	masterPasswords[Username] = hashedPassword
	mu.Unlock()

	// Update the database
	_, err = dbPool.Exec(context.Background(),
		"UPDATE credentials SET master_password = $1 WHERE user_id = $2",
		hashedPassword, Username)
	if err != nil {
		return fmt.Errorf("failed to update master password in database: %v", err)
	}

	return nil
}

// Database
func (appState *AppState) checkInitialization() error {
	if appState.DB == nil {
		return errors.New("database is not initialized")
	}
	if appState.Username == "" {
		return errors.New("username is not set")
	}
	return nil
}

// Database fetch
func (appState *AppState) FetchNotes() error {
	if err := appState.checkInitialization(); err != nil {
		log.Println("FetchNotes:", err)
		return err
	}

	notes, message, err := appState.DB.GetNotes(appState.Username)
	if err != nil {
		log.Printf("Error getting notes: %v", err)
		appState.Notes = []interfaces.Note{}
	} else {
		appState.Notes = notes
	}
	log.Printf("FetchNotes message: %s", message)
	appState.Message = message
	return err
}

func (appState *AppState) FetchTasks() error {
	if err := appState.checkInitialization(); err != nil {
		log.Println("FetchTasks:", err)
		return err
	}

	tasks, _, err := appState.DB.GetTasks(appState.Username)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		appState.Tasks = []interfaces.Tasks{}
	} else {
		appState.Tasks = tasks
	}
	return err
}

func (appState *AppState) FetchAudits() error {
	if err := appState.checkInitialization(); err != nil {
		log.Println("FetchAudits:", err)
		return err
	}

	audits, _, err := appState.DB.GetAudits(appState.Username)
	if err != nil {
		log.Printf("Error getting audits: %v", err)
		appState.Audits = []interfaces.Audits{}
	} else {
		appState.Audits = audits
	}
	return err
}

func (appState *AppState) FetchCRMEntries() error {
	if err := appState.checkInitialization(); err != nil {
		log.Println("FetchCRMEntries:", err)
		return err
	}

	crmEntries, _, err := appState.DB.GetCRMEntries(appState.Username)
	if err != nil {
		log.Printf("Error getting CRM entries: %v", err)
		appState.CRMEntries = []interfaces.CRM{}
	} else {
		appState.CRMEntries = crmEntries
	}
	return err
}

func (appState *AppState) FetchCredentials() error {
	if err := appState.checkInitialization(); err != nil {
		log.Println("FetchCredentials:", err)
		return err
	}
	if appState.DB == nil {
		return fmt.Errorf("database is not initialized")
	}

	credentials, message, err := appState.DB.GetCredentials(appState.Username)
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
		appState.Credentials = []interfaces.Credentials{}
	} else {
		appState.Credentials = credentials
	}
	log.Printf("FetchCredentials message: %s", message)
	appState.Message = message
	return err
}

func (appState *AppState) FetchAll() error {
	if err := appState.FetchNotes(); err != nil {
		return err
	}
	if err := appState.FetchTasks(); err != nil {
		return err
	}
	if err := appState.FetchCredentials(); err != nil {
		return err
	}
	if err := appState.FetchCRMEntries(); err != nil {
		return err
	}
	if err := appState.FetchAudits(); err != nil {
		return err
	}
	return nil
}

// Credentials
func (appState *AppState) SetCredentialAuthenticated(status bool, username string) {
	appState.CredentialAuthStatus = status
	appState.CredentialUsername = username
}

func (appState *AppState) IsCredentialAuthenticated() (bool, string) {
	return appState.CredentialAuthStatus, appState.CredentialUsername
}

func (appState *AppState) ClearCredentialAuthentication() {
	appState.CredentialAuthStatus = false
	appState.CredentialUsername = ""
}

func (s *AppState) IsMasterPasswordSet() bool {
	// Logic to check if the master password is set
	// This could be checking a database or a config file
	return s.CredentialAuthStatus // Adjust according to your logic
}

var (
	masterPasswords = make(map[string][]byte)
	mu              sync.RWMutex
	dbPool          *pgxpool.Pool
)

func (s *AppState) VerifyMasterPassword(userID string, password string) bool {
	mu.RLock()
	hashedPassword, exists := masterPasswords[userID]
	mu.RUnlock()

	if !exists {
		// If not in memory, check the database
		var dbHashedPassword []byte
		err := dbPool.QueryRow(context.Background(),
			"SELECT master_password FROM credentials WHERE user_id = $1", userID).Scan(&dbHashedPassword)
		if err != nil {
			return false
		}
		hashedPassword = dbHashedPassword
	}

	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil
}
