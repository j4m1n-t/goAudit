package crud

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	// External Imports
	"github.com/jackc/pgx/v5"

	// Internal Imports
	serverSide "github.com/j4m1n-t/goAudit/goAuditServer/pkg"
)

type Notes struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    int       `json:"-"`
	User      string    `json:"user"`
	Open      bool      `json:"open"`
}

// Credential Structure
type Credentials struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
	Site       string `json:"site"`
	Program    string `json:"program"`
	UserID     int    `json:"-"`
	User       string `json:"user"`
}

// Audit Structure
type Audits struct {
	ID              int       `json:"id"`
	Action          string    `json:"action"`
	AuditID         int       `json:"audit_id"`
	AuditType       string    `json:"audit_type"`
	AuditArea       string    `json:"audit_area"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Notes           string    `json:"notes"`
	AssignedUser    string    `json:"assigned_user"`
	CompletedAt     time.Time `json:"completed_at"`
	Completed       bool      `json:"completed"`
	UserID          int       `json:"-"`
	User            string    `json:"user"`
	AdditionalUsers []string  `json:"additional_users"`
	Firm            string    `json:"firm"`
}

// Tasks Structure
type Tasks struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Notes       string    `json:"notes"`
	DueDate     time.Time `json:"due_date"`
	Completed   bool      `json:"completed"`
	UserID      int       `json:"-"`
	User        string    `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CRM Structure

type CRM struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Company    string    `json:"company"`
	Notes      []string  `json:"notes"`
	UserID     int       `json:"-"`
	User       string    `json:"user"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Tasks      []Tasks   `json:"tasks"`
	Contacts   []CRM     `json:"contacts"`
	Projects   []CRM     `json:"projects"`
	Activities []CRM     `json:"activities"`
	Open       bool      `json:"open"`
}

// Initialize database connection
var (
	dbNotes       *pgx.Conn
	dbCredentials *sql.DB
	dbAudits      *sql.DB
	dbTasks       *sql.DB
	dbCRM         *sql.DB
)

// Convert to int
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// Initialize database connection
func InitDBNotes() error {
	SQLSettings := serverSide.LoadSQLSettings()
	connString := fmt.Sprintf("user=%s dbname=%s password=%s",
		SQLSettings.User, SQLSettings.Database, SQLSettings.Password)

	var err error
	dbNotes, err = pgx.Connect(context.Background(), connString)
	if err != nil {
		return err
	}

	return nil
}

// Put note in database and associate to the user
func CreateNote(title, content string, userID int) (Notes, error) {
	note := Notes{Title: title, Content: content, UserID: userID}
	query := `INSERT INTO notes (title, content, user_id) 
              VALUES ($1, $2, $3) 
              RETURNING id, created_at, updated_at`

	err := dbNotes.QueryRow(context.Background(), query,
		note.Title, note.Content, note.UserID).
		Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)

	if err != nil {
		return Notes{}, err
	}

	return note, nil
}

func GetNotes() ([]Notes, error) {
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at, 
              notes.updated_at, notes.user_id, users.name 
              FROM notes JOIN users ON notes.user_id = users.id`

	rows, err := dbNotes.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Notes
	for rows.Next() {
		var note Notes
		var userName string
		err := rows.Scan(&note.ID, &note.Title, &note.Content,
			&note.CreatedAt, &note.UpdatedAt,
			&note.UserID, &userName)
		if err != nil {
			return nil, err
		}
		note.User = userName
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// Get a specific note from the database for the given user
func GetNote(id int) (Notes, error) {
	var note Notes
	var userName string
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, notes.user_id, users.name 
              FROM notes JOIN users ON notes.user_id = users.id WHERE notes.id = $1`
	err := dbNotes.QueryRow(context.Background(), query, id).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt, &note.UserID, &userName)
	if err != nil {
		log.Printf("Error getting note. %s", err)
		return Notes{}, err
	}
	note.User = userName
	return note, nil
}

// Update a specific note for the given user
func UpdateNote(note Notes) (Notes, error) {
	query := `UPDATE notes SET title=$1, content=$2, updated_at=$3 WHERE id=$4 RETURNING id, created_at, updated_at`
	err := dbNotes.QueryRow(context.Background(), query, note.Title, note.Content, time.Now(), note.ID).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		log.Printf("Error updating note. %s", err)
		return Notes{}, err
	}
	return note, nil
}

// Delete a specific note for the given user
func DeleteNote(id int) error {
	query := `DELETE FROM notes WHERE id=$1`
	_, err := dbNotes.Exec(context.Background(), query, id)
	if err != nil {
		log.Printf("Error deleting note. %s", err)
		return err
	}
	return nil
}

// Search notes by title or content using PostgreSQL's ILIKE operator
func SearchNotes(searchTerm string) ([]Notes, error) {
	var notes []Notes
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, notes.user_id, users.name 
              FROM notes JOIN users ON notes.user_id = users.id 
              WHERE notes.title ILIKE $1 OR notes.content ILIKE $1`
	rows, err := dbNotes.Query(context.Background(), query, "%"+searchTerm+"%")
	if err != nil {
		log.Printf("Error searching notes. %s", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var note Notes
		var userName string
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt, &note.UserID, &userName)
		if err != nil {
			log.Printf("Error scanning note. %s", err)
			continue
		}
		note.User = userName
		notes = append(notes, note)
	}
	return notes, nil
}
