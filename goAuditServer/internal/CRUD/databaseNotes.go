package internal

import (
	// Standard Library
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	// External Imports
	"github.com/gorilla/mux"
)
// Initialize database connection
func InitDBNotes() error {
	var err error

	dbNotes, err = sql.Open("postgres", "user=your_user dbname=your_db sslmode=disable password=your_password")
	if err != nil {
		log.Fatal(err)
	}
	if err = dbNotes.Ping(); err != nil {
		return err
	}
	log.Println("Database connection established")
	return nil

}
// Notes section
// Put note in database and associate to the user
func CreateNote(db *sql.DB) error {
	note := Notes{Title: "New Note", Content: "This is a new note.", UserID: 1}
	query := `INSERT INTO notes (title, content, user_id) VALUES ($1, $2, $3) RETURNING id`
	err := dbNotes.QueryRow(query, note.Title, note.Content, note.UserID).Scan(&note.ID)
	if err != nil {
		log.Println("Error inserting note.", err)
	}
	return nil
}

// Get all notes associated with the user
func GetNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var notes []Notes

	result, err := dbNotes.Query("SELECT id, title, content, created_at, updated_at, user_id, users.name FROM notes JOIN users ON notes.user_id = users.id")
	if err != nil {
		log.Printf("Error getting notes. %s", err)
	}

	defer result.Close()

	for result.Next() {
		var note Notes
		var userName string

		err := result.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt, &note.UserID, &userName)
		if err != nil {
			log.Printf("Error getting notes. %s", err)
		}

		note.User = userName
		notes = append(notes, note)
	}

	json.NewEncoder(w).Encode(notes)
}

// Get a specific note from the database for the given user
func GetNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	result, err := dbNotes.Query("SELECT id, title, content, created_at, updated_at, user_id, users.name FROM notes JOIN users ON notes.user_id = users.id WHERE id = $1", params["id"])
	if err != nil {
		log.Printf("Error getting note. %s", err)
	}

	defer result.Close()

	var note Notes
	var userName string

	for result.Next() {
		err := result.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt, &note.UserID, &userName)
		if err != nil {
			log.Printf("Error getting note. %s", err)
		}

		note.User = userName
	}

	json.NewEncoder(w).Encode(note)
}

// Update a note in the database for the given user
func UpdateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	var note Notes
	_ = json.NewDecoder(r.Body).Decode(&note)
	note.ID = parseInt(params["id"])

	query := `UPDATE notes SET title=$1, content=$2, updated_at=$3 WHERE id=$4 RETURNING id`
	result, err := dbNotes.Query(query, note.Title, note.Content, time.Now(), note.ID)
	if err != nil {
		log.Printf("Error updating note. %s", err)
	}

	defer result.Close()

	for result.Next() {
		err := result.Scan(&note.ID)
		if err != nil {
			log.Printf("Error updating note. %s", err)
		}
	}

	json.NewEncoder(w).Encode(note)
}

// Delete a note from the database for the given user
func DeleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	query := `DELETE FROM notes WHERE id=$1`
	_, err := dbNotes.Exec(query, params["id"])
	if err != nil {
		log.Printf("Error deleting note. %s", err)
	}

	fmt.Fprintf(w, "Note with ID = %s deleted", params["id"])
}
