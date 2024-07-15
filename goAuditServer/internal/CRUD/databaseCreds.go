package internal

import (
	// Standard Library
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// External Imports
	"github.com/gorilla/mux"
)

// Initialize database connection
func InitDBCred() error {
	var err error

	dbCredentials, err = sql.Open("postgres", "user=your_user dbname=your_db sslmode=disable password=your_password")
	if err != nil {
		log.Fatal(err)
	}
	if err = dbCredentials.Ping(); err != nil {
		return err
	}
	log.Println("Database connection established")
	return nil

}

// Credentials Section
// Create a credential and place it in the database
func CreateCredential(db *sql.DB) error {
	credential := Credentials{Username: "admin", Password: "password"}
	query := `INSERT INTO credentials (username, password, id) VALUES ($1, $2, $3)`
	_, err := dbCredentials.Exec(query, credential.Username, credential.Password, credential.ID)
	if err != nil {
		log.Println("Error inserting credential. ", err)
	}
	return nil
}

// Get all credentials from the database for a given user
func GetCredentials(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var credentials []Credentials

	result, err := dbCredentials.Query("SELECT id, username, password, role FROM credentials")
	if err != nil {
		log.Println("Error retrieving credentials.", err)
	}

	defer result.Close()

	for result.Next() {
		var credential Credentials

		err := result.Scan(&credential.ID, &credential.Username, &credential.Password)
		if err != nil {
			log.Println("Error retrieving credentials.", err)
		}

		credentials = append(credentials, credential)
	}
	json.NewEncoder(w).Encode(credentials)
}

// Get a specific credential from the database for the given user
func GetCredential(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	result, err := dbCredentials.Query("SELECT id, username, password, role FROM credentials WHERE id = $1", params["id"])
	if err != nil {
		log.Printf("Error retrieving credential. %s", err)
	}

	defer result.Close()

	var credential Credentials

	for result.Next() {
		err := result.Scan(&credential.ID, &credential.Username, &credential.Password)
		if err != nil {
			log.Printf("Error retrieving credential. %s", err)
		}
	}
	json.NewEncoder(w).Encode(credential)
}

// Update a credential in the database for the given user
func UpdateCredential(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	var credential Credentials
	_ = json.NewDecoder(r.Body).Decode(&credential)
	credential.ID = parseInt(params["id"])

	query := `UPDATE credentials SET username=$1, password=$2 WHERE id=$3 RETURNING id`
	result, err := dbCredentials.Query(query, credential.Username, credential.Password, credential.ID)
	if err != nil {
		log.Printf("Error updating credential. %s", err)
	}

	defer result.Close()

	for result.Next() {
		err := result.Scan(&credential.ID)
		if err != nil {
			log.Printf("Error updating credential. %s", err)
		}
	}
	json.NewEncoder(w).Encode(credential)
}

// Delete a credential from the database for the given user
func DeleteCredential(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	query := `DELETE FROM credentials WHERE id=$1`
	_, err := dbCredentials.Exec(query, params["id"])
	if err != nil {
		log.Printf("Error deleting credential. %s", err)
	}

	fmt.Fprintf(w, "Credential with ID = %s deleted", params["id"])
}
