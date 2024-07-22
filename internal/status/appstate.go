package state

import (
	"errors"
	"fmt"
	"log"

	myAuth "github.com/j4m1n-t/goAudit/internal/authentication"
	"github.com/j4m1n-t/goAudit/internal/interfaces"
)

type AppState struct {
	LDAPConn    *myAuth.LDAPConnection
	Username    string
	Notes       []interfaces.Note
	Tasks       []interfaces.Tasks
	Audits      []interfaces.Audits
	CRMEntries  []interfaces.CRM
	Credentials []interfaces.Credentials
	Message     string
	DB          interfaces.DatabaseOperations
}

var GlobalState = &AppState{}

func (appState *AppState) SetDB(db interfaces.DatabaseOperations) {
	if db == nil {
		log.Fatal("SetDB: Database instance cannot be nil")
	}
	appState.DB = db
}

func (appState *AppState) checkInitialization() error {
	if appState.DB == nil {
		return errors.New("database is not initialized")
	}
	if appState.Username == "" {
		return errors.New("username is not set")
	}
	return nil
}

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
