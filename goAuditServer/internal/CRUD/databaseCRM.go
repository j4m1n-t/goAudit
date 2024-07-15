package internal

import (
	// Standard Library
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
	// External Imports
)

// Set structures to be used for the program

// Note Structure
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
	dbNotes       *sql.DB
	dbCredentials *sql.DB
	dbAudits      *sql.DB
	dbTasks       *sql.DB
	dbCRM         *sql.DB
)

func InitDBCRM() error {
	var err error
	dbCRM, err = sql.Open("postgres", "user=your_user dbname=your_db sslmode=disable password=your_password")
	if err != nil {
		return err
	}
	if err = dbCRM.Ping(); err != nil {
		return err
	}
	log.Println("Database connection established")
	return nil
}

// Convert to int
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// CRM Section
// Create a new CRM in the database

func CreateCRM(db *sql.DB) error {
	query := `INSERT INTO crm (name, created_at, user_id) VALUES ($1, $2, $3) RETURNING id`
	result, err := dbCRM.Exec(query, "New CRM", time.Now(), 1)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	fmt.Println("New CRM created with ID:", id)
	return nil
}

// Get all CRMs from the database for a given user
func GetCRMs(db *sql.DB, userID int) ([]CRM, error) {
	rows, err := dbCRM.Query("SELECT id, name, created_at, user_id FROM crm WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var crm []CRM
	for rows.Next() {
		var c CRM
		err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UserID)
		if err != nil {
			return nil, err
		}
		crm = append(crm, c)
	}
	return crm, nil
}

// Get a specific CRM for the given user
func GetCRM(db *sql.DB, userID, crmID int) (*CRM, error) {
	row := dbCRM.QueryRow("SELECT id, name, created_at, user_id FROM crm WHERE user_id = $1 AND id = $2", userID, crmID)

	var c CRM
	err := row.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UserID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("CRM not found")
	} else if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update a CRM in the database for the given user
func UpdateCRM(db *sql.DB, userID, crmID int, name string) error {
	_, err := dbCRM.Exec("UPDATE crm SET name=$1 WHERE user_id=$2 AND id=$3", name, userID, crmID)
	if err != nil {
		return err
	}
	fmt.Println("CRM updated")
	return nil
}

// Delete a CRM from the database for the given user
func DeleteCRM(db *sql.DB, userID, crmID int) error {
	_, err := dbCRM.Exec("DELETE FROM crm WHERE user_id=$1 AND id=$2", userID, crmID)
	if err != nil {
		return err
	}
	fmt.Println("CRM deleted")
	return nil
}
