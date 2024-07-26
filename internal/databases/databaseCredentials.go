package databases

import (
	// Standard Library
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	// Internal Imports
	interfaces "github.com/j4m1n-t/goAudit/internal/interfaces"
	"github.com/jackc/pgx"
	"github.com/lib/pq"
)

func EnsureCredentialsTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS credentials (
        id SERIAL PRIMARY KEY,
        site TEXT,
        program TEXT,
		user_id INTEGER NOT NULL,
        username TEXT NOT NULL,
		email TEXT,
        master_password TEXT NOT NULL,
        login_name TEXT,
        login_pass TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        owner TEXT,
		password_history JSONB DEFAULT '[]'::jsonb
    );`

	_, err := DBPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create credentials table: %v", err)
	}

	return nil
}

func CreateCredential(credential interfaces.Credentials) (interfaces.Credentials, error) {
	query := `INSERT INTO credentials (site, program, username,user_id, email master_password, login_name, login_pass, owner)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
              RETURNING id, created_at, updated_at`

	err := DBPool.QueryRow(context.Background(), query,
		credential.Site, credential.Program, credential.Username, credential.UserID, credential.Email, credential.MasterPassword,
		credential.LoginName, credential.LoginPass, credential.Owner).
		Scan(&credential.ID, &credential.CreatedAt, &credential.UpdatedAt)
	if err != nil {
		return interfaces.Credentials{}, err
	}

	return credential, nil
}

func (dw *DatabaseWrapper) CreateCredential(credential interfaces.Credentials) (interfaces.Credentials, error) {
	passwordHistoryJSON, err := json.Marshal(credential.PasswordHistory)
	if err != nil {
		return interfaces.Credentials{}, fmt.Errorf("failed to marshal password history: %v", err)
	}

	query := `INSERT INTO credentials (site, program, username, user_id, email, master_password, login_name, login_pass, owner, password_history)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
              RETURNING id, created_at, updated_at`

	err = DBPool.QueryRow(context.Background(), query,
		credential.Site, credential.Program, credential.Username, credential.UserID, credential.Email, credential.MasterPassword,
		credential.LoginName, credential.LoginPass, credential.Owner, passwordHistoryJSON).
		Scan(&credential.ID, &credential.CreatedAt, &credential.UpdatedAt)

	if err != nil {
		return interfaces.Credentials{}, err
	}

	return credential, nil
}

func (dw *DatabaseWrapper) GetCredentials(owner string) ([]interfaces.Credentials, string, error) {
	query := `SELECT id, site, program, username, user_id, email, master_password, login_name, login_pass, created_at, updated_at, owner, password_history
              FROM credentials
              WHERE owner = $1
              ORDER BY created_at DESC`

	rows, err := DBPool.Query(context.Background(), query, owner)
	if err != nil {
		return nil, fmt.Sprintf("Error connecting to database: %s", err), err
	}
	defer rows.Close()

	var credentials []interfaces.Credentials
	for rows.Next() {
		var credential interfaces.Credentials
		var passwordHistoryJSON []byte
		err := rows.Scan(&credential.ID, &credential.Site, &credential.Program, &credential.Username, &credential.UserID, &credential.Email,
			&credential.MasterPassword, &credential.LoginName, &credential.LoginPass, &credential.CreatedAt,
			&credential.UpdatedAt, &credential.Owner, &passwordHistoryJSON)
		if err != nil {
			log.Printf("Error scanning credential: %v", err)
			continue
		}

		err = json.Unmarshal(passwordHistoryJSON, &credential.PasswordHistory)
		if err != nil {
			log.Printf("Error unmarshaling password history: %v", err)
			credential.PasswordHistory = []string{}
		}

		credentials = append(credentials, credential)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Sprintf("Search returned an error: %s", err), err
	}

	if len(credentials) == 0 {
		return credentials, "No credentials were found.", nil
	}

	return credentials, "", nil
}

func (dw *DatabaseWrapper) GetCredentialByLoginName(loginName string) ([]interfaces.Credentials, error) {
	var creds []interfaces.Credentials
	query := `SELECT id, site, program, username,user_id, email, master_password, login_name, login_pass, created_at, updated_at, owner, password_history
              FROM credentials
              WHERE login_name = $1
              LIMIT 1`

	var cred interfaces.Credentials
	err := DBPool.QueryRow(context.Background(), query, loginName).Scan(
		&cred.ID, &cred.Site, &cred.Program, &cred.Username, &cred.UserID, &cred.Email, &cred.MasterPassword,
		&cred.LoginName, &cred.LoginPass, &cred.CreatedAt, &cred.UpdatedAt, &cred.Owner, pq.Array(&cred.PasswordHistory))
	if err != nil {
		creds = append(creds, cred)
		return creds, fmt.Errorf("error getting credential: %w", err)
	}

	return creds, nil
}

// Double check this works properly
func (dw *DatabaseWrapper) UpdateCredential(credential interfaces.Credentials) (interfaces.Credentials, error) {
	passwordHistoryJSON, err := json.Marshal(credential.PasswordHistory)
	if err != nil {
		return interfaces.Credentials{}, fmt.Errorf("failed to marshal password history: %v", err)
	}

	query := `UPDATE credentials SET site=$1, program=$2, username=$3, master_password=$4, login_name=$5,
              login_pass=$6, updated_at=$7, owner=$8, password_history=$9
              WHERE id=$10 RETURNING id, created_at, updated_at`

	err = DBPool.QueryRow(context.Background(), query,
		credential.Site, credential.Program, credential.Username, credential.MasterPassword, credential.LoginName,
		credential.LoginPass, time.Now(), credential.Owner, passwordHistoryJSON, credential.ID).
		Scan(&credential.ID, &credential.CreatedAt, &credential.UpdatedAt)

	if err != nil {
		return interfaces.Credentials{}, err
	}

	return credential, nil
}

func (dw *DatabaseWrapper) DeleteCredential(id int, owner string) error {
	query := `DELETE FROM credentials WHERE id=$1 AND owner=$2`
	_, err := DBPool.Exec(context.Background(), query, id, owner)
	return err
}

func (dw *DatabaseWrapper) SearchCredentials(searchTerm, owner string) ([]interfaces.Credentials, string, error) {
	// Using a more lenient search to match any part of the login name or site
	query := `
	SELECT id, site, program, username, user_id, email, master_password, login_name, login_pass, created_at, updated_at, owner, password_history
	FROM credentials
	WHERE (login_name ILIKE $1 OR site ILIKE $1) AND owner = $2
	ORDER BY created_at DESC
	`

	// Prepare the search term with wildcards for partial matches
	searchPattern := "%" + searchTerm + "%"

	rows, err := DBPool.Query(context.Background(), query, searchPattern, owner)
	if err != nil {
		return nil, "", fmt.Errorf("error querying credentials: %v", err)
	}
	defer rows.Close()

	var credentials []interfaces.Credentials
	for rows.Next() {
		var cred interfaces.Credentials
		err := rows.Scan(
			&cred.ID, &cred.Site, &cred.Program, &cred.Username, &cred.UserID, &cred.Email, &cred.MasterPassword, &cred.LoginName,
			&cred.LoginPass, &cred.CreatedAt, &cred.UpdatedAt, &cred.Owner, pq.Array(&cred.PasswordHistory),
		)
		if err != nil {
			return nil, "", fmt.Errorf("error scanning row: %v", err)
		}
		credentials = append(credentials, cred)
	}

	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("error after scanning rows: %v", err)
	}

	if len(credentials) == 0 {
		return credentials, "No credentials found", nil
	}

	return credentials, "Credentials fetched successfully", nil
}
func (dw *DatabaseWrapper) CreateCredUser(username, hashedPassword, email string) (*interfaces.Credentials, error) {
	// Check if the db pool is initialized
	if dw.Pool == nil {
		return nil, errors.New("database pool is not initialized")
	}
	// First, we need to get the user_id from the users table
	println(username, email)
	var userID int
	userQuery := `SELECT users.user_id FROM users WHERE username = $1`
	err := dw.Pool.QueryRow(context.Background(), userQuery, username).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no user found with username: %s", username)
		}
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	// Now we can insert into the credentials table using the user_id
	query := `INSERT INTO credentials (user_id, username, master_password, email, created_at, updated_at)
              VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
              RETURNING id, username, created_at, updated_at`
	var cred interfaces.Credentials
	err = dw.Pool.QueryRow(context.Background(), query, userID, username, hashedPassword, email).
		Scan(&cred.ID, &cred.Username, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error creating credential: %v", err)
	}

	// Set the UserID in the returned credential struct
	cred.UserID = userID

	return &cred, nil
}

func (dw *DatabaseWrapper) GetUserPassword(username string) (string, error) {
	var hashedPassword string
	query := `SELECT password FROM credentials WHERE username = $1`
	err := DBPool.QueryRow(context.Background(), query, username).Scan(&hashedPassword)
	return hashedPassword, err
}
