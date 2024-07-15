package crud

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
func InitDBAudit() error {
	var err error

	dbAudits, err = sql.Open("postgres", "user=your_user dbname=your_db sslmode=disable password=your_password")
	if err != nil {
		log.Fatal(err)
	}
	if err = dbAudits.Ping(); err != nil {
		return err
	}
	log.Println("Database connection established")
	return nil

}

// Audit Section
// Create an audit in the database

func CreateAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var audit Audits
	_ = json.NewDecoder(r.Body).Decode(&audit)

	query := `INSERT INTO audits (action, audit_id, audit_type, audit_area, created_at, notes, assigned_user, user_id) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
              RETURNING id`

	err := dbAudits.QueryRow(query,
		audit.Action,
		audit.AuditID,
		audit.AuditType,
		audit.AuditArea,
		time.Now(),
		audit.Notes,
		audit.AssignedUser,
		audit.UserID).Scan(&audit.ID)

	if err != nil {
		log.Printf("Error creating audit: %v", err)
		http.Error(w, "Error creating audit", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(audit)
}

// Get all audits from the database for a given user
func GetAudits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var audits []Audits

	result, err := dbAudits.Query("SELECT id, action, audit_id, audit_type, audit_area, created_at, notes, assigned_user, user_id FROM audits WHERE user_id = $1", r.Context().Value("user_id").(int))
	if err != nil {
		log.Printf("Error retrieving audits: %v", err)
		http.Error(w, "Error retrieving audits", http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var audit Audits

		err := result.Scan(&audit.ID, &audit.Action, &audit.AuditID, &audit.AuditType, &audit.AuditArea, &audit.CreatedAt, &audit.Notes)
		if err != nil {
			log.Printf("Error retrieving audits: %v", err)
			http.Error(w, "Error retrieving audits", http.StatusInternalServerError)
			return
		}

		audits = append(audits, audit)
	}
	json.NewEncoder(w).Encode(audits)
}

// Get a specific audit for the given user
func GetAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	result, err := dbAudits.Query("SELECT id, action, audit_id, audit_type, audit_area, created_at, notes, assigned_user, user_id FROM audits WHERE user_id = $1 AND id = $2", r.Context().Value("user_id").(int), params["id"])
	if err != nil {
		log.Printf("Error retrieving audit: %v", err)
		http.Error(w, "Error retrieving audit", http.StatusInternalServerError)
		return
	}

	defer result.Close()

	var audit Audits

	for result.Next() {
		err := result.Scan(&audit.ID, &audit.Action, &audit.AuditID, &audit.AuditType, &audit.AuditArea, &audit.CreatedAt, &audit.Notes)
		if err != nil {
			log.Printf("Error retrieving audit: %v", err)
			http.Error(w, "Error retrieving audit", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(audit)
	}
}

// Update an audit in the database for the given user

func UpdateAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	var audit Audits
	_ = json.NewDecoder(r.Body).Decode(&audit)
	audit.ID = parseInt(params["id"])

	query := `UPDATE audits SET action=$1, audit_id=$2, audit_type=$3, audit_area=$4, created_at=$5, notes=$6, assigned_user=$7 WHERE id=$8 RETURNING id`

	err := dbAudits.QueryRow(query,
		audit.Action,
		audit.AuditID,
		audit.AuditType,
		audit.AuditArea,
		audit.CreatedAt,
		audit.Notes,
		audit.AssignedUser,
		audit.ID).Scan(&audit.ID)

	if err != nil {
		log.Println("Error updating audit.", err)
		http.Error(w, "Error updating audit", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(audit)
}

// Delete an audit from the database for the given user

func DeleteAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	query := `DELETE FROM audits WHERE id=$1 AND user_id=$2`
	_, err := dbAudits.Exec(query, params["id"], r.Context().Value("user_id").(int))
	if err != nil {
		log.Println("Error deleting audit.", err)
		http.Error(w, "Error deleting audit", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Audit with ID = %s deleted", params["id"])
}
