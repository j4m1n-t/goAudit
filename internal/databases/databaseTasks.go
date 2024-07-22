package databases

import (
	"context"
	"fmt"
	"log"
	"time"

	interfaces "github.com/j4m1n-t/goAudit/internal/interfaces"
)

func EnsureTaskTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS tasks (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT,
        status TEXT,
        priority INTEGER,
        notes TEXT,
        due_date TIMESTAMP WITH TIME ZONE,
        completed BOOLEAN DEFAULT FALSE,
        user_id INTEGER NOT NULL,
        username TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := DBPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %v", err)
	}

	return nil
}

func CreateTask(task interfaces.Tasks) (interfaces.Tasks, error) {
	query := `INSERT INTO tasks (title, description, status, priority, notes, due_date, completed, user_id, username)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
              RETURNING id, created_at, updated_at`

	err := DBPool.QueryRow(context.Background(), query,
		task.Title, task.Description, task.Status, task.Priority, task.Notes, task.DueDate, task.Completed, task.UserID, task.Username).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return interfaces.Tasks{}, err
	}

	return task, nil
}

func (dw *DatabaseWrapper) GetTasks(username string) ([]interfaces.Tasks, string, error) {
	query := `SELECT id, title, description, status, priority, notes, due_date, completed, user_id, username, created_at, updated_at
              FROM tasks
              WHERE username = $1
              ORDER BY due_date ASC`

	rows, err := DBPool.Query(context.Background(), query, username)
	if err != nil {
		return nil, fmt.Sprintf("Error querying tasks: %v", err), err
	}
	defer rows.Close()

	var tasks []interfaces.Tasks
	for rows.Next() {
		var task interfaces.Tasks
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority, &task.Notes,
			&task.DueDate, &task.Completed, &task.UserID, &task.Username, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Sprintf("Error after scanning all rows: %v", err), err
	}

	if len(tasks) == 0 {
		return tasks, "No tasks found", nil
	}

	return tasks, "Tasks fetched successfully", nil
}

func UpdateTask(task interfaces.Tasks) (interfaces.Tasks, error) {
	query := `UPDATE tasks SET title=$1, description=$2, status=$3, priority=$4, notes=$5, due_date=$6, completed=$7, updated_at=$8
              WHERE id=$9 RETURNING id, created_at, updated_at`

	err := DBPool.QueryRow(context.Background(), query,
		task.Title, task.Description, task.Status, task.Priority, task.Notes, task.DueDate, task.Completed, time.Now(), task.ID).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return interfaces.Tasks{}, err
	}

	return task, nil
}

func DeleteTask(id int, username string) error {
	query := `DELETE FROM tasks WHERE id=$1 AND username=$2`
	_, err := DBPool.Exec(context.Background(), query, id, username)
	return err
}
