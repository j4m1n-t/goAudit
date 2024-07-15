package internal

import (
	// Standard Library
	"database/sql"
	"fmt"
	"log"
	"time"
	// External Imports
)

// Initialize database connection
func InitDBTasks() error {
	var err error

	dbTasks, err = sql.Open("postgres", "user=your_user dbname=your_db sslmode=disable password=your_password")
	if err != nil {
		log.Fatal(err)
	}
	if err = dbTasks.Ping(); err != nil {
		return err
	}
	log.Println("Database connection established")
	return nil

}

// Tasks Section
// Create a task in the database for the given user
func CreateTask(db *sql.DB) error {
	query := `INSERT INTO tasks (title, description, due_date, created_at, completed, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	result, err := dbTasks.Exec(query, "Task Title", "Task Description", time.Now(), false, 1)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	fmt.Println("New task created with ID:", id)
	return nil
}

// Get all tasks from the database for a given user
func GetTasks(db *sql.DB, userID int) ([]Tasks, error) {
	rows, err := dbTasks.Query("SELECT id, title, description, due_date, created_at, completed, user_id FROM tasks WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Tasks
	for rows.Next() {
		var task Tasks
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.CreatedAt, &task.Completed, &task.UserID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// Get a specific task for the given user
func GetTask(db *sql.DB, userID, taskID int) (*Tasks, error) {
	row := dbTasks.QueryRow("SELECT id, title, description, due_date, created_at, completed, user_id FROM tasks WHERE user_id = $1 AND id = $2", userID, taskID)

	var task Tasks
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.CreatedAt, &task.Completed, &task.UserID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	} else if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update a task in the database for the given user
func UpdateTask(db *sql.DB, userID, taskID int, title, description string, dueDate time.Time, completed bool) error {
	_, err := dbTasks.Exec("UPDATE tasks SET title=$1, description=$2, due_date=$3, completed=$4 WHERE user_id=$5 AND id=$6", title, description, dueDate, completed, userID, taskID)
	if err != nil {
		return err
	}
	fmt.Println("Task updated")
	return nil
}

// Delete a task from the database for the given user
func DeleteTask(db *sql.DB, userID, taskID int) error {
	_, err := dbTasks.Exec("DELETE FROM tasks WHERE user_id=$1 AND id=$2", userID, taskID)
	if err != nil {
		return err
	}
	fmt.Println("Task deleted")
	return nil
}
