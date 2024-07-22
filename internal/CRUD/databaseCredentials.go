package crud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	mySettings "github.com/j4m1n-t/goAudit/internal/functions"
)

func InitDBCredentials() error {
	SQLSettings := mySettings.LoadSQLSettings()

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		SQLSettings.User, SQLSettings.Password, SQLSettings.Server, SQLSettings.Port, SQLSettings.Database)

	var err error
	dbPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	err = dbPool.Ping(context.Background())
	if err != nil {
		dbPool.Close()
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database for credentials.")

	if err := EnsureCredentialsTableExists(); err != nil {
		return fmt.Errorf("failed to ensure credentials table exists: %v", err)
	}

	return nil
}

func EnsureCredentialsTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS credentials (
        id SERIAL PRIMARY KEY,
        login_name TEXT NOT NULL,
        password TEXT NOT NULL,
        remember_me BOOLEAN,
        site TEXT,
        program TEXT,
        user_id INTEGER NOT NULL,
        username TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create credentials table: %v", err)
	}

	return nil
}

func CreateCredential(loginName, password, site, program, username string, rememberMe bool) (Credentials, error) {
	User, err := GetOrCreateUser(username)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return Credentials{}, err
	}

	credential := Credentials{
		LoginName:  loginName,
		Password:   password, // Consider encrypting this before storage
		RememberMe: rememberMe,
		Site:       site,
		Program:    program,
		UserID:     User.UserID,
		Username:   User.Username,
	}

	query := `INSERT INTO credentials (login_name, password, remember_me, site, program, user_id, username)
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`

	err = dbPool.QueryRow(context.Background(), query,
		credential.LoginName, credential.Password, credential.RememberMe,
		credential.Site, credential.Program, credential.UserID, credential.Username).
		Scan(&credential.ID, &credential.CreatedAt, &credential.UpdatedAt)

	if err != nil {
		return Credentials{}, err
	}

	return credential, nil
}

func GetCredentials(username string) ([]Credentials, string, error) {
	query := `
    SELECT id, login_name, password, remember_me, site, program, user_id, username, created_at, updated_at
    FROM credentials
    WHERE username = $1
    ORDER BY created_at DESC
    `

	rows, err := dbPool.Query(context.Background(), query, username)
	if err != nil {
		log.Printf("Error querying credentials: %v", err)
		return nil, fmt.Sprintf("Error querying credentials: %v", err), err
	}
	defer rows.Close()

	var credentials []Credentials
	for rows.Next() {
		var cred Credentials
		err := rows.Scan(&cred.ID, &cred.LoginName, &cred.Password, &cred.RememberMe,
			&cred.Site, &cred.Program, &cred.UserID, &cred.Username, &cred.CreatedAt, &cred.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning credential: %v", err)
			continue
		}
		credentials = append(credentials, cred)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(credentials) == 0 {
		return credentials, "No credentials found", nil
	}

	return credentials, "Credentials fetched successfully", nil
}

func UpdateCredential(credential Credentials) (Credentials, error) {
	query := `UPDATE credentials 
              SET login_name=$1, password=$2, remember_me=$3, site=$4, program=$5, updated_at=$6 
              WHERE id=$7 RETURNING id, created_at, updated_at`

	err := dbPool.QueryRow(context.Background(), query,
		credential.LoginName, credential.Password, credential.RememberMe,
		credential.Site, credential.Program, time.Now(), credential.ID).
		Scan(&credential.ID, &credential.CreatedAt, &credential.UpdatedAt)

	if err != nil {
		log.Printf("Error updating credential: %v", err)
		return Credentials{}, err
	}

	return credential, nil
}

func DeleteCredential(id int, username string) error {
	query := `DELETE FROM credentials WHERE id=$1 AND username=$2`
	_, err := dbPool.Exec(context.Background(), query, id, username)
	if err != nil {
		log.Printf("Error deleting credential: %v", err)
		return err
	}
	log.Printf("Credential with ID %d deleted by user %s", id, username)
	return nil
}

func SearchCredentials(searchTerm string, username string) ([]Credentials, string, error) {
	query := `
    SELECT id, login_name, password, remember_me, site, program, user_id, username, created_at, updated_at
    FROM credentials
    WHERE username = $1 AND (login_name ILIKE $2 OR site ILIKE $2 OR program ILIKE $2)
    ORDER BY created_at DESC
    `

	rows, err := dbPool.Query(context.Background(), query, username, "%"+searchTerm+"%")
	if err != nil {
		log.Printf("Error searching credentials: %v", err)
		return nil, fmt.Sprintf("Error searching credentials: %v", err), err
	}
	defer rows.Close()

	var credentials []Credentials
	for rows.Next() {
		var cred Credentials
		err := rows.Scan(&cred.ID, &cred.LoginName, &cred.Password, &cred.RememberMe,
			&cred.Site, &cred.Program, &cred.UserID, &cred.Username, &cred.CreatedAt, &cred.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning credential: %v", err)
			continue
		}
		credentials = append(credentials, cred)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(credentials) == 0 {
		return credentials, "No matching credentials found", nil
	}

	return credentials, "Search completed successfully", nil
}
