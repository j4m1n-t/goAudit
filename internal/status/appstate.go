package state

import (
	"log"

	myAuth "github.com/j4m1n-t/goAudit/internal/authentication"
	crud "github.com/j4m1n-t/goAudit/internal/databases"
)

type AppState struct {
	LDAPConn    *myAuth.LDAPConnection
	Username    string
	Notes       []crud.Notes
	Tasks       []crud.Tasks
	Audits      []crud.Audits
	CRMEntries  []crud.CRM
	Credentials []crud.Credentials
	Message     string
}

var GlobalState AppState

func (appState *AppState) FetchNotes() error {
	if appState.Username == "" {
		log.Println("FetchNotes: No username provided")
		return nil
	}
	notes, message, err := crud.GetNotes(appState.Username)
	if err != nil {
		log.Printf("Error getting notes: %v", err)
		appState.Notes = []crud.Notes{}
	} else {
		appState.Notes = notes
		return nil
	}
	log.Printf("FetchNotes message: %s", message)
	appState.Message = message
	return err
}

func (appState *AppState) FetchTasks() error {
	if appState.Username == "" {
		log.Println("FetchTasks: No username provided")
		return nil
	}
	tasks, _, err := crud.GetTasks(appState.Username)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		appState.Tasks = []crud.Tasks{}
	} else {
		appState.Tasks = tasks
	}
	return err
}

func (appState *AppState) FetchAudits() error {
	if appState.Username == "" {
		log.Println("FetchAudits: No username provided")
		return nil
	}
	audits, _, err := crud.GetAudits(appState.Username)
	if err != nil {
		log.Printf("Error getting audits: %v", err)
		appState.Audits = []crud.Audits{}
	} else {
		appState.Audits = audits
	}
	return err
}

func (appState *AppState) FetchCRMEntries() error {
	if appState.Username == "" {
		log.Println("FetchCRMEntries: No username provided")
		return nil
	}
	crmEntries, _, err := crud.GetCRMEntries(appState.Username)
	if err != nil {
		log.Printf("Error getting CRM entries: %v", err)
		appState.CRMEntries = []crud.CRM{}
	} else {
		appState.CRMEntries = crmEntries
	}
	return err
}

func (appState *AppState) FetchCredentials() error {
	if appState.Username == "" {
		log.Println("FetchCredentials: No username provided")
		return nil
	}
	credentials, message, err := crud.GetCredentials(appState.Username)
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
		appState.Credentials = []crud.Credentials{}
	} else {
		appState.Credentials = credentials
	}
	log.Printf("FetchCredentials message: %s", message)
	appState.Message = message
	return err
}

func (appState *AppState) FetchAll() {
	appState.FetchNotes()
	appState.FetchTasks()
	appState.FetchCredentials()
	appState.FetchCRMEntries()
	appState.FetchAudits()
}
