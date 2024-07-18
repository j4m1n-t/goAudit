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
	UserID    int       `json:"-"`
	Username  string    `json:"username"`
	Open      bool      `json:"open"`
	Author    string    `json:"author"`
	UpdatedBy string    `json:"updated_by"`
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
	ID         int    `json:"id"`
	LoginName  string `json:"login_name"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
	Site       string `json:"site"`
	Program    string `json:"program"`
	UserID     int    `json:"-"`
	Username   string `json:"username"`
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
	Username        string    `json:"username"`
	AdditionalUsers []string  `json:"additional_users"`
	Firm            string    `json:"firm"`
}

// Tasks Structure
type Tasks struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    int       `json:"priority"`
	Notes       string    `json:"notes"`
	DueDate     time.Time `json:"due_date"`
	Completed   bool      `json:"completed"`
	UserID      int       `json:"-"`
	Username    string    `json:"username"`
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
	Username   string    `json:"username"`
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

func InitDBs() error {
	if err := InitDBNotes(); err != nil {
		return err
	}
	// if err := InitDBCredentials(); err!= nil {
	//     return err
	// }
	// if err := InitDBAudits(); err!= nil {
	//     return err
	// }
	if err := InitDBTasks(); err != nil {
		return err
	}
	// if err := InitDBCRM(); err!= nil {
	//     return err
	// }
	if err := InitDBUsers(); err != nil {
		return err
	}
	return nil
}

// Initialize database connection
func InitDBNotes() error {
	SQLSettings := mySettings.LoadSQLSettings()

	connString1 := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		SQLSettings.User, SQLSettings.Password, SQLSettings.Server, SQLSettings.Port, SQLSettings.Database)

	var err error
	dbPool, err = pgxpool.New(context.Background(), connString1)
	if err == nil {
		err = dbPool.Ping(context.Background())
		if err == nil {
			log.Println("Successfully connected to database for notes.")
			if err := EnsureNotesTableExists(); err != nil {
				return fmt.Errorf("failed to ensure notes table exists: %v", err)
			}
			return nil
		}
		dbPool.Close()
	}
	log.Printf("Failed to connect using the notes database: %v", err)
	return fmt.Errorf("failed to connect to notes database: %v", err)
}

func EnsureNotesTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS notes (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        content TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        user_id INTEGER NOT NULL,
        username TEXT,
		updated_by TEXT,
		author TEXT,
        open BOOLEAN DEFAULT FALSE
    );`

	_, err := dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %v", err)
	}

	alterTableSQL := `
    DO $$ 
    BEGIN 
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                       WHERE table_name='notes' AND column_name='open') THEN
            ALTER TABLE notes ADD COLUMN open BOOLEAN DEFAULT FALSE;
        END IF;
    END $$;`

	_, err = dbPool.Exec(context.Background(), alterTableSQL)
	if err != nil {
		return fmt.Errorf("failed to add 'open' column to notes table: %v", err)
	}

	alterTableAuthorSQL := `
    DO $$ 
    BEGIN 
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                       WHERE table_name='notes' AND column_name='author') THEN
            ALTER TABLE notes ADD COLUMN author TEXT;
        END IF;
    END $$;`

	_, err = dbPool.Exec(context.Background(), alterTableAuthorSQL)
	if err != nil {
		return fmt.Errorf("failed to add 'author' column to notes table: %v", err)
	}

	return nil
}

// Put note in database and associate to the user
func CreateNote(title, content string, username string, open bool) (Notes, error) {
	User, err := GetOrCreateUser(username)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return Notes{}, err
	}
	note := Notes{
		Title:    title,
		Content:  content,
		UserID:   User.UserID,
		Username: User.Username,
		Open:     open,
		Author:   username,
	}
	query := `INSERT INTO notes (title, content, user_id, open) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err = dbPool.QueryRow(context.Background(), query, note.Title, note.Content, User.UserID, note.Open).
		Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return Notes{}, err
	}
	return note, nil
}

func GetNotes(username string) ([]Notes, string, error) {
	if dbPool == nil {
		return nil, "", fmt.Errorf("database connection not initialized")
	}

	query := `
    SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, notes.user_id, users.username, notes.open, notes.author
    FROM notes
    JOIN users ON notes.user_id = users.user_id
    WHERE notes.user_id = (SELECT users.user_id FROM users WHERE users.username ILIKE $1) OR notes.open = true
    ORDER BY notes.created_at DESC
    `

	rows, err := dbPool.Query(context.Background(), query, "%"+username+"%")
	if err != nil {
		log.Printf("Error querying notes: %v", err)
		return nil, fmt.Sprintf("Error querying notes: %v", err), err
	}
	defer rows.Close()

	var notes []Notes
	for rows.Next() {
		var note Notes
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
			&note.UserID, &note.Username, &note.Open, &note.Author)
		if err != nil {
			log.Printf("Error scanning note: %v", err)
			continue
		}
		log.Printf("Fetched note: %+v", note)
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(notes) == 0 {
		return notes, "No notes found", nil
	}

	return notes, "Notes fetched successfully", nil
}

func GetNote(id int) (Notes, error) {
	var note Notes
	var user Users
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, 
              notes.user_id, users.id, users.username, users.email
              FROM notes JOIN users ON notes.user_id = user.id WHERE notes.id = $1`
	err := dbPool.QueryRow(context.Background(), query, id).Scan(
		&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
		&user.UserID, &user.ID, &user.Username, &user.Email)
	if err != nil {
		log.Printf("Error getting note. %s", err)
		return Notes{}, err
	}
	return note, nil
}

func UpdateNote(note Notes) (Notes, error) {
	query := `UPDATE notes SET title=$1, content=$2, updated_at=$3, open=$4 WHERE id=$5 RETURNING id, created_at, updated_at`
	err := dbPool.QueryRow(context.Background(), query, note.Title, note.Content, time.Now(), note.Open, note.ID).
		Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		log.Printf("Error updating note. %s", err)
		return Notes{}, err
	}
	return note, nil
}

func DeleteNote(id int, Username string) error {
	query := `DELETE FROM notes WHERE id=$1`
	_, err := dbPool.Exec(context.Background(), query, id)
	if err != nil {
		log.Printf("Error deleting note. %s", err)
		return err
	}
	log.Printf("Note with ID %d marked as deleted by user %s", id, Username)
	return nil
}

func SearchNotes(searchTerm string, username string) ([]Notes, string, error) {
	var notes []Notes
	query := `
    SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, notes.user_id, users.username, notes.open, notes.author
    FROM notes
    JOIN users ON notes.user_id = users.user_id
    WHERE (title ILIKE $1 OR content ILIKE $1)
    AND (notes.user_id = (SELECT users.user_id FROM users WHERE users.username ILIKE $2) OR notes.open = true)
    ORDER BY notes.created_at DESC
    `
	log.Printf("Executing SQL search query: %s\nWith parameters: searchTerm='%s', username='%s'",
		query, "%"+searchTerm+"%", "%"+username+"%")

	rows, err := dbPool.Query(context.Background(), query, "%"+searchTerm+"%", "%"+username+"%")
	if err != nil {
		log.Printf("Error searching notes: %v", err)
		return nil, fmt.Sprintf("Error searching notes: %v", err), err
	}
	defer rows.Close()

	for rows.Next() {
		var note Notes
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
			&note.UserID, &note.Username, &note.Open, &note.Author)
		if err != nil {
			log.Printf("Error scanning note: %v", err)
			continue
		}
		log.Printf("Fetched note: %+v", note)
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(notes) == 0 {
		return notes, "No matching notes found", nil
	}

	return notes, "Search completed successfully", nil
}

func CloseDBConnection() {
	if dbPool != nil {
		dbPool.Close()
	}
}

// TODO //
// Add delete to the view of a note
//
//
//
