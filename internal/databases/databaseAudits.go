package databases

import (
	// Standard Library
	"context"
	"fmt"
	"log"
	"time"

	// Internal Imports
	interfaces "github.com/j4m1n-t/goAudit/internal/interfaces"
)

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

	_, err := DBPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create audits table: %v", err)
	}

	return nil
}

func (dw *DatabaseWrapper) CreateAudit(audit interfaces.Audits) (interfaces.Audits, error) {
	query := `INSERT INTO audits (action, audit_id, audit_type, audit_area, notes, assigned_user, completed, user_id, username, additional_users, firm)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
              RETURNING id, created_at, updated_at`

	err := DBPool.QueryRow(context.Background(), query,
		audit.Action, audit.AuditID, audit.AuditType, audit.AuditArea, audit.Notes, audit.AssignedUser, audit.Completed,
		audit.UserID, audit.Username, audit.AdditionalUsers, audit.Firm).
		Scan(&audit.ID, &audit.CreatedAt, &audit.UpdatedAt)

	if err != nil {
		return interfaces.Audits{}, err
	}

	return audit, nil
}

func (dw *DatabaseWrapper) GetAudits(username string) ([]interfaces.Audits, string, error) {
	query := `SELECT id, action, audit_id, audit_type, audit_area, created_at, updated_at, notes, assigned_user, completed_at, completed, user_id, username, additional_users, firm
              FROM audits
              WHERE username = $1 OR $1 = ANY(additional_users)
              ORDER BY created_at DESC`

	rows, err := DBPool.Query(context.Background(), query, username)
	if err != nil {
		return nil, fmt.Sprintf("Error querying audits: %v", err), err
	}
	defer rows.Close()

	var audits []interfaces.Audits
	for rows.Next() {
		var audit interfaces.Audits
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

func (dw *DatabaseWrapper) UpdateAudit(audit interfaces.Audits) (interfaces.Audits, error) {
	query := `UPDATE audits SET action=$1, audit_id=$2, audit_type=$3, audit_area=$4, notes=$5, assigned_user=$6,
              completed_at=$7, completed=$8, additional_users=$9, firm=$10, updated_at=$11
              WHERE id=$12 RETURNING id, created_at, updated_at`

	err := DBPool.QueryRow(context.Background(), query,
		audit.Action, audit.AuditID, audit.AuditType, audit.AuditArea, audit.Notes, audit.AssignedUser,
		audit.CompletedAt, audit.Completed, audit.AdditionalUsers, audit.Firm, time.Now(), audit.ID).
		Scan(&audit.ID, &audit.CreatedAt, &audit.UpdatedAt)

	if err != nil {
		return interfaces.Audits{}, err
	}

	return audit, nil
}

func (dw *DatabaseWrapper) DeleteAudit(id int, username string) error {
	query := `DELETE FROM audits WHERE id=$1 AND username=$2`
	_, err := DBPool.Exec(context.Background(), query, id, username)
	return err
}
