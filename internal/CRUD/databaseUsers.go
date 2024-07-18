package crud

import (
	"context"
	"fmt"
	"log"
	"time"

	// External Imports
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"

	// Internal Imports
	mySettings "github.com/j4m1n-t/goAudit/internal/functions"
)

func GetUserByAnyID(identifier interface{}) (Users, error) {
	var user Users
	var query string
	var args []interface{}

	switch v := identifier.(type) {
	case int:
		query = `SELECT id, username, user_id, email, status, created_at, updated_at, last_login
                 FROM users WHERE id = $1 OR user_id = $1`
		args = []interface{}{v}
	case string:
		query = `SELECT id, username, user_id, email, status, created_at, updated_at, last_login
                 FROM users WHERE username = $1`
		args = []interface{}{v}
	default:
		return Users{}, fmt.Errorf("invalid identifier type")
	}

	err := dbPool.QueryRow(context.Background(), query, args...).Scan(
		&user.ID, &user.Username, &user.UserID, &user.Email, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Users{}, fmt.Errorf("user not found")
		}
		return Users{}, err
	}
	return user, nil
}

func InitDB() error {
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

	log.Println("Successfully connected to database.")

	if err := EnsureUserTableExists(); err != nil {
		return fmt.Errorf("failed to ensure table exists: %v", err)
	}

	return nil
}

func EnsureUserTableExists() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		user_id INTEGER UNIQUE NOT NULL,
		email TEXT,
		status TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		last_login TIMESTAMP WITH TIME ZONE
	);`

	_, err := dbPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	return nil
}

func Create(username, email, status string, user_id int, id int, created_at, updated_at, last_login time.Time) (Users, error) {
	userItem := Users{Username: username, Email: email, Status: status, UserID: user_id, CreatedAt: created_at, UpdatedAt: updated_at, LastLogin: last_login}
	query := `INSERT INTO users (username, user_id, email, status, created_at, updated_at, last_login) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) 
              RETURNING id, username, user_id, email, status, created_at, updated_at, last_login`

	err := dbPool.QueryRow(context.Background(), query,
		userItem.Username, userItem.UserID, userItem.Email, userItem.Status,
		userItem.CreatedAt, userItem.UpdatedAt, userItem.LastLogin).
		Scan(&userItem.ID, &userItem.Username, &userItem.UserID, &userItem.Email,
			&userItem.Status, &userItem.CreatedAt, &userItem.UpdatedAt, &userItem.LastLogin)

	if err != nil {
		return Users{}, err
	}

	return userItem, nil
}
func GetOrCreateUser(username string) (Users, error) {
	user, err := GetUserByAnyID(username)
	if err != nil {
		// User not found, create a new one
		newUser := Users{
			Username:  username,
			UserID:    generateUserID(),
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		createdUser, err := Create(newUser.Username, "", newUser.Status, newUser.UserID, 0, newUser.CreatedAt, newUser.UpdatedAt, time.Time{})
		if err != nil {
			return Users{}, fmt.Errorf("failed to create user: %v", err)
		}
		return createdUser, nil
	}
	return user, nil
}

func generateUserID() int {
	return int(time.Now().UnixNano())
}

// Do we want this? It seems like an error
func GetAll() ([]Users, error) {
	if dbPool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	query := `SELECT id, user, user_id, email, status, created_at, updated_at, last_login FROM users`

	rows, err := dbPool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []Users
	for rows.Next() {
		var user Users
		err := rows.Scan(&user.ID, &user.Username, &user.UserID, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func Get(id int) (Users, error) {
	var user Users
	query := `SELECT id, user, user_id, email, status, created_at, updated_at, last_login
              FROM users WHERE id = $1`
	err := dbPool.QueryRow(context.Background(), query, id).Scan(
		&user.ID, &user.Username, &user.UserID, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		return Users{}, err
	}
	return user, nil
}

func Update(user Users) (Users, error) {
	query := `UPDATE users SET user=$1, user_id=$2, email=$3, status=$4, updated_at=$5, last_login=$6 
              WHERE id=$7 RETURNING id, created_at, updated_at`
	err := dbPool.QueryRow(context.Background(), query,
		user.Username, user.UserID, user.Email, user.Status, time.Now(), user.LastLogin, user.ID).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return Users{}, err
	}
	return user, nil
}

// Reserve for administrative operations only
func Delete(user Users) error {
	query := `DELETE FROM users WHERE id=$1 AND user_id=$2`
	_, err := dbPool.Exec(context.Background(), query, user.ID, user.UserID)
	return err
}

// Get by ID
//user, err := GetUserByAnyID(1)

// Get by UserID
//user, err := GetUserByAnyID(1001)

// Get by Username
//user, err := GetUserByAnyID("johndoe")
