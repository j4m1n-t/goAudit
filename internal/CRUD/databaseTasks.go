package crud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	mySettings "github.com/j4m1n-t/goAudit/internal/functions"
)

// Initialize database connection

func InitDBTasks() error {
	SQLSettings := mySettings.LoadSQLSettings()

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		SQLSettings.User, SQLSettings.Password, SQLSettings.Server, SQLSettings.Port, SQLSettings.Database)

	var err error
	dbPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %v", err)
	}

	err = dbPool.Ping(context.Background())
	if err != nil {
		dbPool.Close()
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database for tasks.")

	if err := EnsureTasksTableExists(); err != nil {
		return fmt.Errorf("failed to ensure tasks table exists: %v", err)
	}

	return nil
}

func EnsureTasksTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS tasks (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT,
        status TEXT,
        priority INTEGER,
        notes TEXT,
        due_date TIMESTAMP WITH TIME ZONE,
        completed BOOLEAN,
        user_id INTEGER NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %v", err)
	}
	return nil
}

func CreateTask(task Tasks) (Tasks, error) {
	query := `INSERT INTO tasks (title, description, status, priority, notes, due_date, completed, user_id)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id, created_at, updated_at`

	err := dbPool.QueryRow(context.Background(), query,
		task.Title, task.Description, task.Status, task.Priority,
		task.Notes, task.DueDate, task.Completed, task.UserID).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return Tasks{}, fmt.Errorf("failed to create task: %v", err)
	}

	return task, nil
}

func GetTasks() ([]Tasks, error) {
	query := `SELECT t.id, t.title, t.description, t.status, t.priority, t.notes,
              t.due_date, t.completed, t.created_at, t.updated_at,
              u.id, u.username, u.email
              FROM tasks t JOIN users u ON t.user_id = u.id`

	rows, err := dbPool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %v", err)
	}
	defer rows.Close()

	var tasks []Tasks
	for rows.Next() {
		var task Tasks
		var user Users
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status,
			&task.Priority, &task.Notes, &task.DueDate, &task.Completed,
			&task.CreatedAt, &task.UpdatedAt,
			&user.ID, &user.Username, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func GetTask(id int) (Tasks, error) {
	var task Tasks
	var user Users
	query := `SELECT t.id, t.title, t.description, t.status, t.priority, t.notes,
              t.due_date, t.completed, t.created_at, t.updated_at,
              u.id, u.username, u.email
              FROM tasks t JOIN users u ON t.user_id = u.id WHERE t.id = $1`

	err := dbPool.QueryRow(context.Background(), query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.Priority, &task.Notes, &task.DueDate, &task.Completed,
		&task.CreatedAt, &task.UpdatedAt,
		&user.ID, &user.Username, &user.Email)

	if err != nil {
		return Tasks{}, fmt.Errorf("failed to get task: %v", err)
	}

	return task, nil
}

func UpdateTask(task Tasks) (Tasks, error) {
	query := `UPDATE tasks SET title=$1, description=$2, status=$3, priority=$4,
              notes=$5, due_date=$6, completed=$7, updated_at=$8
              WHERE id=$9 RETURNING created_at, updated_at`

	err := dbPool.QueryRow(context.Background(), query,
		task.Title, task.Description, task.Status, task.Priority,
		task.Notes, task.DueDate, task.Completed, time.Now(), task.ID).
		Scan(&task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return Tasks{}, fmt.Errorf("failed to update task: %v", err)
	}

	return task, nil
}

func DeleteTask(id int) error {
	query := `DELETE FROM tasks WHERE id=$1`
	_, err := dbPool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}
	return nil
}

func SearchTasks(searchTerm string) ([]Tasks, error) {
	query := `SELECT t.id, t.title, t.description, t.status, t.priority, t.notes,
              t.due_date, t.completed, t.created_at, t.updated_at,
              u.id, u.username, u.email
              FROM tasks t JOIN users u ON t.user_id = u.id
              WHERE t.title ILIKE $1 OR t.description ILIKE $1 OR t.notes ILIKE $1`

	rows, err := dbPool.Query(context.Background(), query, "%"+searchTerm+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %v", err)
	}
	defer rows.Close()

	var tasks []Tasks
	for rows.Next() {
		var task Tasks
		var user Users
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status,
			&task.Priority, &task.Notes, &task.DueDate, &task.Completed,
			&task.CreatedAt, &task.UpdatedAt,
			&user.ID, &user.Username, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task during search: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
