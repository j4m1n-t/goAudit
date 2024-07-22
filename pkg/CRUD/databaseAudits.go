package crud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	mySettings "github.com/j4m1n-t/goAudit/pkg/functions"
)

func InitDBAudits() error {
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

	log.Println("Successfully connected to database for audits.")

	if err := EnsureAuditTableExists(); err != nil {
		return fmt.Errorf("failed to ensure audit table exists: %v", err)
	}

	return nil
}

func EnsureAuditTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS audits (
        id SERIAL PRIMARY KEY,
        action TEXT NOT NULL,
        audit_id INTEGER,
        audit_type TEXT,
        audit_area TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        notes TEXT,
        assigned_user TEXT,
        completed_at TIMESTAMP WITH TIME ZONE,
        completed BOOLEAN DEFAULT FALSE,
        user_id INTEGER NOT NULL,
        username TEXT,
        additional_users TEXT[],
        firm TEXT
    );`

	_, err := dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create audits table: %v", err)
	}

	return nil
}

func CreateAudit(audit Audits) (Audits, error) {
	query := `INSERT INTO audits (action, audit_id, audit_type, audit_area, notes, assigned_user, completed, user_id, username, additional_users, firm)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
              RETURNING id, created_at, updated_at`

	err := dbPool.QueryRow(context.Background(), query,
		audit.Action, audit.AuditID, audit.AuditType, audit.AuditArea, audit.Notes, audit.AssignedUser, audit.Completed,
		audit.UserID, audit.Username, audit.AdditionalUsers, audit.Firm).
		Scan(&audit.ID, &audit.CreatedAt, &audit.UpdatedAt)

	if err != nil {
		return Audits{}, err
	}

	return audit, nil
}

func GetAudits(username string) ([]Audits, string, error) {
	query := `SELECT id, action, audit_id, audit_type, audit_area, created_at, updated_at, notes, assigned_user, completed_at, completed, user_id, username, additional_users, firm
              FROM audits
              WHERE username = $1 OR $1 = ANY(additional_users)
              ORDER BY created_at DESC`

	rows, err := dbPool.Query(context.Background(), query, username)
	if err != nil {
		return nil, fmt.Sprintf("Error querying audits: %v", err), err
	}
	defer rows.Close()

	var audits []Audits
	for rows.Next() {
		var audit Audits
		err := rows.Scan(&audit.ID, &audit.Action, &audit.AuditID, &audit.AuditType, &audit.AuditArea, &audit.CreatedAt,
			&audit.UpdatedAt, &audit.Notes, &audit.AssignedUser, &audit.CompletedAt, &audit.Completed, &audit.UserID,
			&audit.Username, &audit.AdditionalUsers, &audit.Firm)
		if err != nil {
			log.Printf("Error scanning audit: %v", err)
			continue
		}
		audits = append(audits, audit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(audits) == 0 {
		return audits, "No audits found", nil
	}

	return audits, "Audits fetched successfully", nil
}

func UpdateAudit(audit Audits) (Audits, error) {
	query := `UPDATE audits SET action=$1, audit_id=$2, audit_type=$3, audit_area=$4, notes=$5, assigned_user=$6,
              completed_at=$7, completed=$8, additional_users=$9, firm=$10, updated_at=$11
              WHERE id=$12 RETURNING id, created_at, updated_at`

	err := dbPool.QueryRow(context.Background(), query,
		audit.Action, audit.AuditID, audit.AuditType, audit.AuditArea, audit.Notes, audit.AssignedUser,
		audit.CompletedAt, audit.Completed, audit.AdditionalUsers, audit.Firm, time.Now(), audit.ID).
		Scan(&audit.ID, &audit.CreatedAt, &audit.UpdatedAt)

	if err != nil {
		return Audits{}, err
	}

	return audit, nil
}

func DeleteAudit(id int, username string) error {
	query := `DELETE FROM audits WHERE id=$1 AND username=$2`
	_, err := dbPool.Exec(context.Background(), query, id, username)
	return err
}
