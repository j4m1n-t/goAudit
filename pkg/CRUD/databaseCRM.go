package crud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	mySettings "github.com/j4m1n-t/goAudit/pkg/functions"
)

func InitDBCRM() error {
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

	log.Println("Successfully connected to database for CRM.")

	if err := EnsureCRMTableExists(); err != nil {
		return fmt.Errorf("failed to ensure CRM table exists: %v", err)
	}

	return nil
}

func EnsureCRMTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS crm (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT,
        phone TEXT,
        company TEXT,
        notes TEXT[],
        user_id INTEGER NOT NULL,
        username TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        open BOOLEAN DEFAULT TRUE
    );`

	_, err := dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create CRM table: %v", err)
	}

	return nil
}

func CreateCRMEntry(crm CRM) (CRM, error) {
	query := `INSERT INTO crm (name, email, phone, company, notes, user_id, username, open)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id, created_at, updated_at`

	err := dbPool.QueryRow(context.Background(), query,
		crm.Name, crm.Email, crm.Phone, crm.Company, crm.Notes, crm.UserID, crm.Username, crm.Open).
		Scan(&crm.ID, &crm.CreatedAt, &crm.UpdatedAt)

	if err != nil {
		return CRM{}, err
	}

	return crm, nil
}

func GetCRMEntries(username string) ([]CRM, string, error) {
	query := `SELECT id, name, email, phone, company, notes, user_id, username, created_at, updated_at, open
              FROM crm
              WHERE username = $1 OR open = true
              ORDER BY updated_at DESC`

	rows, err := dbPool.Query(context.Background(), query, username)
	if err != nil {
		return nil, fmt.Sprintf("Error querying CRM entries: %v", err), err
	}
	defer rows.Close()

	var crmEntries []CRM
	for rows.Next() {
		var crm CRM
		err := rows.Scan(&crm.ID, &crm.Name, &crm.Email, &crm.Phone, &crm.Company, &crm.Notes,
			&crm.UserID, &crm.Username, &crm.CreatedAt, &crm.UpdatedAt, &crm.Open)
		if err != nil {
			log.Printf("Error scanning CRM entry: %v", err)
			continue
		}
		crmEntries = append(crmEntries, crm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(crmEntries) == 0 {
		return crmEntries, "No CRM entries found", nil
	}

	return crmEntries, "CRM entries fetched successfully", nil
}

func UpdateCRMEntry(crm CRM) (CRM, error) {
	query := `UPDATE crm SET name=$1, email=$2, phone=$3, company=$4, notes=$5, open=$6, updated_at=$7
              WHERE id=$8 RETURNING id, created_at, updated_at`

	err := dbPool.QueryRow(context.Background(), query,
		crm.Name, crm.Email, crm.Phone, crm.Company, crm.Notes, crm.Open, time.Now(), crm.ID).
		Scan(&crm.ID, &crm.CreatedAt, &crm.UpdatedAt)

	if err != nil {
		return CRM{}, err
	}

	return crm, nil
}

func DeleteCRMEntry(id int, username string) error {
	query := `DELETE FROM crm WHERE id=$1 AND username=$2`
	_, err := dbPool.Exec(context.Background(), query, id, username)
	return err
}
