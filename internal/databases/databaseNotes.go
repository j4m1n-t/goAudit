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

func EnsureNotesTableExists() error {
	// Ensure notes table exists
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

	_, err := DBPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %v", err)
	}

	// Add additional columns if they do not exist
	alterTableSQL := `
    DO $$ 
    BEGIN 
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                       WHERE table_name='notes' AND column_name='open') THEN
            ALTER TABLE notes ADD COLUMN open BOOLEAN DEFAULT FALSE;
        END IF;
    END $$;`

	_, err = DBPool.Exec(context.Background(), alterTableSQL)
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

	_, err = DBPool.Exec(context.Background(), alterTableAuthorSQL)
	if err != nil {
		return fmt.Errorf("failed to add 'author' column to notes table: %v", err)
	}

	return nil
}

func (dw *DatabaseWrapper) CreateNote(title, content string, username string, open bool) (interfaces.Note, error) {
	User, err := dw.GetOrCreateUser(username)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return interfaces.Note{}, err
	}
	note := interfaces.Note{
		Title:    title,
		Content:  content,
		UserID:   User.UserID,
		Username: User.Username,
		Open:     open,
		Author:   User.Username,
	}
	query := `INSERT INTO notes (title, content, user_id, open) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err = DBPool.QueryRow(context.Background(), query, note.Title, note.Content, User.UserID, note.Open).
		Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return interfaces.Note{}, err
	}
	return note, nil
}

func (dw *DatabaseWrapper) GetNotes(username string) ([]interfaces.Note, string, error) {
	if DBPool == nil {
		return nil, "", fmt.Errorf("database connection not initialized")
	}

	query := `
    SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, notes.user_id, users.username, notes.open, notes.author
    FROM notes
    JOIN users ON notes.user_id = users.user_id
    WHERE notes.user_id = (SELECT users.user_id FROM users WHERE users.username ILIKE $1) OR notes.open = true
    ORDER BY notes.created_at DESC
    `

	rows, err := DBPool.Query(context.Background(), query, "%"+username+"%")
	if err != nil {
		log.Printf("Error querying notes: %v", err)
		return nil, fmt.Sprintf("Error querying notes: %v", err), err
	}
	defer rows.Close()

	var notes []interfaces.Note
	for rows.Next() {
		var note interfaces.Note
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

func (dw *DatabaseWrapper) GetNote(id int) (interfaces.Note, error) {
	var note interfaces.Note
	query := `SELECT notes.id, notes.title, notes.content, notes.created_at, notes.updated_at, 
              notes.user_id, users.username, users.email
              FROM notes JOIN users ON notes.user_id = users.user_id WHERE notes.id = $1`
	err := DBPool.QueryRow(context.Background(), query, id).Scan(
		&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt,
		&note.UserID, &note.Username, &note.Username)
	if err != nil {
		log.Printf("Error getting note. %s", err)
		return interfaces.Note{}, err
	}
	return note, nil
}

func (dw *DatabaseWrapper) UpdateNote(note interfaces.Note) (interfaces.Note, error) {
	query := `UPDATE notes SET title=$1, content=$2, updated_at=$3, open=$4 WHERE id=$5 RETURNING id, created_at, updated_at`
	err := DBPool.QueryRow(context.Background(), query, note.Title, note.Content, time.Now(), note.Open, note.ID).
		Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		log.Printf("Error updating note. %s", err)
		return interfaces.Note{}, err
	}
	return note, nil
}

func (dw *DatabaseWrapper) DeleteNote(id int) error {
	query := `DELETE FROM notes WHERE id=$1`
	_, err := DBPool.Exec(context.Background(), query, id)
	if err != nil {
		log.Printf("Error deleting note. %s", err)
		return err
	}
	log.Printf("Note with ID %d marked as deleted", id)
	return nil
}

func (dw *DatabaseWrapper) SearchNotes(searchTerm string, username string) ([]interfaces.Note, string, error) {
	var notes []interfaces.Note
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

	rows, err := DBPool.Query(context.Background(), query, "%"+searchTerm+"%", "%"+username+"%")
	if err != nil {
		log.Printf("Error searching notes: %v", err)
		return nil, fmt.Sprintf("Error searching notes: %v", err), err
	}
	defer rows.Close()

	for rows.Next() {
		var note interfaces.Note
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
