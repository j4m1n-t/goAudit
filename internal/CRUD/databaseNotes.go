package crud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	// External Imports
	"github.com/jackc/pgx/v5/pgxpool"

	// Internal Imports
	mySettings "github.com/j4m1n-t/goAudit/internal/functions"
)

type Notes struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    []Users   `json:"-"`
	Username      []Users   `json:"username"`
	Open      bool      `json:"open"`
}
type Users struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    UserID    int       `json:"user_id"`
    Email     string    `json:"email"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    LastLogin time.Time `json:"last_login"`
}

// Credential Structure
type Credentials struct {
	ID         int     `json:"id"`
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	RememberMe bool    `json:"remember_me"`
	Site       string  `json:"site"`
	Program    string  `json:"program"`
	UserID     []Users `json:"-"`
	Username       []Users `json:"username"`
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
	UserID          []Users   `json:"-"`
	Username            []Users   `json:"username"`
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
	UserID      []Users   `json:"-"`
	Username        []Users   `json:"username"`
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
	UserID     []Users   `json:"-"`
	Username       []Users   `json:"username"`
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
	dbPool *pgxpool.Pool
)

// Convert to int
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func EnsureTablesExist() error {
	if err := EnsureNotesTableExists(); err != nil {
		return err
	}
	// if err := EnsureCredentialsTableExists(); err != nil {
	//     return err
	// }
	// if err := EnsureAuditsTableExists(); err != nil {
	//     return err
	// }
	// if err := EnsureTasksTableExists(); err != nil {
	//     return err
	// }
	// if err := EnsureCRMTableExists(); err != nil {
	//     return err
	// }
	return nil
}

// Initialize database connection
func InitDBNotes() error {
	SQLSettings := mySettings.LoadSQLSettings()

	// First connection string (URL style)
	connString1 := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		SQLSettings.User, SQLSettings.Password, SQLSettings.Server, SQLSettings.Port, SQLSettings.Database)

	// Try the first connection string
	var err error
	dbPool, err = pgxpool.New(context.Background(), connString1)
	if err == nil {
		err = dbPool.Ping(context.Background())
		if err == nil {
			log.Println("Successfully connected to database using the connection string.")
			if err := EnsureNotesTableExists(); err != nil {
				return fmt.Errorf("failed to ensure notes table exists: %v", err)
			}

			return nil
		}
		dbPool.Close()
	}
	log.Printf("Failed to connect using the first connection string: %v", err)

	return fmt.Errorf("failed to connect to database using the connection string: %v", err)

}

// Ensure that the database contains the table
func EnsureNotesTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS notes (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        content TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        user_id INTEGER NOT NULL,
        user_name TEXT,
        open BOOLEAN
    );`

	_, err := dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %v", err)
	}
	return nil
}

// Put note in database and associate to the user
func CreateNote(title, content string, userID int) (Notes, error) {
	note := Notes{
		Title:   title,
		Content: content,
		UserID:  []Users{{ID: userID}},
		User:    []Users{{ID: userID}},
	}

	query := `INSERT INTO notes (title, content, user_id) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := dbPool.QueryRow(context.Background(), query, note.Title, note.Content, userID).
		Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return Notes{}, err
	}

	return note, nil
}

func GetNotes() ([]Notes, error) {
	if dbPool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at,
              notes.updated_at, notes.user_id, users.id, users.user, users.email
              FROM notes JOIN users ON notes.user_id = users.id`
	rows, err := dbPool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []Notes
	for rows.Next() {
		var note Notes
		var user Users
		err := rows.Scan(&note.ID, &note.Title, &note.Content,
			&note.CreatedAt, &note.UpdatedAt,
			&user.UserID, &user.ID, &user.User, &user.Email)
		if err != nil {
			return nil, err
		}
		note.UserID = []Users{user}
		note.User = []Users{user}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}

func GetNote(id int) (Notes, error) {
	var note Notes
	var user Users
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, 
              notes.user_id, users.id, users.user, users.email
              FROM notes JOIN users ON notes.user_id = users.id WHERE notes.id = $1`
	err := dbPool.QueryRow(context.Background(), query, id).Scan(
		&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
		&user.UserID, &user.ID, &user.User, &user.Email)
	if err != nil {
		log.Printf("Error getting note. %s", err)
		return Notes{}, err
	}
	note.UserID = []Users{user}
	note.User = []Users{user}
	return note, nil
}

func UpdateNote(note Notes) (Notes, error) {
	query := `UPDATE notes SET title=$1, content=$2, updated_at=$3 WHERE id=$4 RETURNING id, created_at, updated_at`
	err := dbPool.QueryRow(context.Background(), query, note.Title, note.Content, time.Now(), note.ID).
		Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		log.Printf("Error updating note. %s", err)
		return Notes{}, err
	}
	return note, nil
}

func DeleteNote(id int) error {
	query := `DELETE FROM notes WHERE id=$1`
	_, err := dbPool.Exec(context.Background(), query, id)
	if err != nil {
		log.Printf("Error deleting note. %s", err)
		return err
	}
	return nil
}

func SearchNotes(searchTerm string) ([]Notes, error) {
	var notes []Notes
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, 
              notes.user_id, users.id, users.user, users.email
              FROM notes JOIN users ON notes.user_id = users.id
              WHERE notes.title ILIKE $1 OR notes.content ILIKE $1`
	rows, err := dbPool.Query(context.Background(), query, "%"+searchTerm+"%")
	if err != nil {
		log.Printf("Error searching notes. %s", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var note Notes
		var user Users
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
			&user.UserID, &user.ID, &user.User, &user.Email)
		if err != nil {
			log.Printf("Error scanning note. %s", err)
			continue
		}
		note.UserID = []Users{user}
		note.User = []Users{user}
		notes = append(notes, note)
	}
	return notes, nil
}

func CloseDBConnection() {
	if dbPool != nil {
		dbPool.Close()
	}
}
