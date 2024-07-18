package state

import (
	"log"

	crud "github.com/j4m1n-t/goAudit/internal/CRUD"
	myAuth "github.com/j4m1n-t/goAudit/internal/authentication"
)

type AppState struct {
	LDAPConn *myAuth.LDAPConnection
	Username string
	Notes    []crud.Notes
	Message  string
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
