package databases

import (
	"context"
	"fmt"
	"log"
	"time"

	interfaces "github.com/j4m1n-t/goAudit/internal/interfaces"
)

func EnsureCredentialsTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS credentials (
        id SERIAL PRIMARY KEY,
        site TEXT NOT NULL,
        program TEXT NOT NULL,
        username TEXT NOT NULL,
        master_password TEXT NOT NULL,
        login_name TEXT,
        login_pass TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        owner TEXT
    );`

	_, err := DBPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create credentials table: %v", err)
	}

	return nil
}

func CreateCredential(credential interfaces.Credentials) (interfaces.Credentials, error) {
	query := `INSERT INTO credentials (site, program, username, master_password, login_name, login_pass, owner)
              VALUES ($1, $2, $3, $4, $5, $6, $7)
              RETURNING id, created_at, updated_at`

	err := DBPool.QueryRow(context.Background(), query,
		credential.Site, credential.Program, credential.Username, credential.MasterPassword,
		credential.LoginName, credential.LoginPass, credential.Owner).
		Scan(&credential.ID, &credential.CreatedAt, &credential.UpdatedAt)

	if err != nil {
		return interfaces.Credentials{}, err
	}

	return credential, nil
}

func (dw *DatabaseWrapper) GetCredentials(owner string) ([]interfaces.Credentials, string, error) {
	query := `SELECT id, site, program, username, master_password, login_name, login_pass, created_at, updated_at, owner
              FROM credentials
              WHERE owner = $1
              ORDER BY created_at DESC`

	rows, err := DBPool.Query(context.Background(), query, owner)
	if err != nil {
		return nil, fmt.Sprintf("Error querying credentials: %v", err), err
	}
	defer rows.Close()

	var credentials []interfaces.Credentials
	for rows.Next() {
		var credential interfaces.Credentials
		err := rows.Scan(&credential.ID, &credential.Site, &credential.Program, &credential.Username,
			&credential.MasterPassword, &credential.LoginName, &credential.LoginPass, &credential.CreatedAt,
			&credential.UpdatedAt, &credential.Owner)
		if err != nil {
			log.Printf("Error scanning credential: %v", err)
			continue
		}
		credentials = append(credentials, credential)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(credentials) == 0 {
		return credentials, "No credentials found", nil
	}

	return credentials, "Credentials fetched successfully", nil
}

func UpdateCredential(credential interfaces.Credentials) (interfaces.Credentials, error) {
	query := `UPDATE credentials SET site=$1, program=$2, username=$3, master_password=$4, login_name=$5,
              login_pass=$6, updated_at=$7, owner=$8
              WHERE id=$9 RETURNING id, created_at, updated_at`

	err := DBPool.QueryRow(context.Background(), query,
		credential.Site, credential.Program, credential.Username, credential.MasterPassword, credential.LoginName,
		credential.LoginPass, time.Now(), credential.Owner, credential.ID).
		Scan(&credential.ID, &credential.CreatedAt, &credential.UpdatedAt)

	if err != nil {
		return interfaces.Credentials{}, err
	}

	return credential, nil
}

func DeleteCredential(id int, owner string) error {
	query := `DELETE FROM credentials WHERE id=$1 AND owner=$2`
	_, err := DBPool.Exec(context.Background(), query, id, owner)
	return err
}

func SearchCredentials(searchTerm, owner string) ([]interfaces.Credentials, string, error) {
	// Using a more lenient search to match any part of the login name or site
	query := `
	SELECT id, site, program, username, master_password, login_name, login_pass, created_at, updated_at, owner
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
			&cred.ID, &cred.Site, &cred.Program, &cred.Username, &cred.MasterPassword, &cred.LoginName,
			&cred.LoginPass, &cred.CreatedAt, &cred.UpdatedAt, &cred.Owner,
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
