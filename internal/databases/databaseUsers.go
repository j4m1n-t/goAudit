package databases

import (
	// Standard Library
	"context"
	"fmt"
	"math/rand"
	"time"

	//External Imports
	"github.com/jackc/pgx"

	// Internal Imports
	interfaces "github.com/j4m1n-t/goAudit/internal/interfaces"
)

func GetUserByAnyID(identifier interface{}) (interfaces.Users, error) {
	var user interfaces.Users
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
		return interfaces.Users{}, fmt.Errorf("invalid identifier type")
	}

	err := DBPool.QueryRow(context.Background(), query, args...).Scan(
		&user.ID, &user.Username, &user.UserID, &user.Email, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		if err == pgx.ErrNoRows {
			return interfaces.Users{}, fmt.Errorf("user not found")
		}
		return interfaces.Users{}, err
	}
	return user, nil
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

	_, err := DBPool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	return nil
}

func Create(username, email, status string, userID int, createdAt, updatedAt, lastLogin time.Time) (interfaces.Users, error) {
	userItem := interfaces.Users{
		Username:  username,
		Email:     email,
		Status:    status,
		UserID:    userID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		LastLogin: lastLogin,
	}
	query := `INSERT INTO users (username, user_id, email, status, created_at, updated_at, last_login) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) 
              ON CONFLICT (user_id) DO UPDATE SET
              username = EXCLUDED.username,
              email = EXCLUDED.email,
              status = EXCLUDED.status,
              updated_at = EXCLUDED.updated_at,
              last_login = EXCLUDED.last_login
              RETURNING id, username, user_id, email, status, created_at, updated_at, last_login`

	err := DBPool.QueryRow(context.Background(), query,
		userItem.Username, userItem.UserID, userItem.Email, userItem.Status,
		userItem.CreatedAt, userItem.UpdatedAt, userItem.LastLogin).
		Scan(&userItem.ID, &userItem.Username, &userItem.UserID, &userItem.Email,
			&userItem.Status, &userItem.CreatedAt, &userItem.UpdatedAt, &userItem.LastLogin)

	if err != nil {
		return interfaces.Users{}, err
	}

	return userItem, nil
}

func (dw *DatabaseWrapper) Create(username string) (interfaces.Users, error) {
	userItem := interfaces.Users{
		Username: username,
		Status:   "Active",
		UserID:   generateUserID(),
	}
	query := `INSERT INTO users (username, user_id, status) 
              VALUES ($1, $2, $3) 
              ON CONFLICT (user_id) DO UPDATE SET
              username = EXCLUDED.username,
              status = EXCLUDED.status,
              RETURNING id, username, user_id, email, status, created_at, updated_at, last_login`

	err := DBPool.QueryRow(context.Background(), query,
		userItem.Username, userItem.UserID, userItem.Email, userItem.Status,
		userItem.CreatedAt, userItem.UpdatedAt, userItem.LastLogin).
		Scan(&userItem.ID, &userItem.Username, &userItem.UserID, &userItem.Email,
			&userItem.Status, &userItem.CreatedAt, &userItem.UpdatedAt, &userItem.LastLogin)

	if err != nil {
		return interfaces.Users{}, err
	}

	return userItem, nil
}

func (dw *DatabaseWrapper) GetOrCreateUser(username string) (interfaces.Users, error) {
	user, err := GetUserByAnyID(username)
	if err != nil {
		// User not found, create a new one
		newUser := interfaces.Users{
			Username:  username,
			UserID:    generateUserID(),
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		createdUser, err := Create(newUser.Username, "", newUser.Status, newUser.UserID, newUser.CreatedAt, newUser.UpdatedAt, time.Time{})
		if err != nil {
			return interfaces.Users{}, fmt.Errorf("failed to create user: %v", err)
		}
		return createdUser, nil
	}
	return user, nil
}

func (dw *DatabaseWrapper) GetUsers(username string) ([]interfaces.Users, string, error) {
	var users []interfaces.Users
	user, err := GetUserByAnyID(username)
	if err != nil {
		if err != nil {
			return []interfaces.Users{}, fmt.Sprintf("user not found: %s", username), err
		}
		users = append(users, user)
		return users, fmt.Sprintf("User %s", username), nil
	}
	users = append(users, user)
	return users, fmt.Sprintf("User %s", username), nil
}

func generateUserID() int {
	timestamp := time.Now().Unix()
	randomPart := rand.Intn(10000)
	userID := (int(timestamp)%100000)*10000 + randomPart
	return userID
}

func (dw *DatabaseWrapper) GetAll() ([]interfaces.Users, error) {
	if DBPool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	query := `SELECT id, username, user_id, email, status, created_at, updated_at, last_login FROM users`

	rows, err := DBPool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []interfaces.Users
	for rows.Next() {
		var user interfaces.Users
		err := rows.Scan(&user.ID, &user.Username, &user.UserID, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func Get(id int) (interfaces.Users, error) {
	var user interfaces.Users
	query := `SELECT id, username, user_id, email, status, created_at, updated_at, last_login
              FROM users WHERE id = $1`
	err := DBPool.QueryRow(context.Background(), query, id).Scan(
		&user.ID, &user.Username, &user.UserID, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		return interfaces.Users{}, err
	}
	return user, nil
}

func (dw *DatabaseWrapper) Update(user interfaces.Users) (interfaces.Users, error) {
	query := `UPDATE users SET username=$1, user_id=$2, email=$3, status=$4, updated_at=$5, last_login=$6 
              WHERE id=$7 RETURNING id, created_at, updated_at`
	err := DBPool.QueryRow(context.Background(), query,
		user.Username, user.UserID, user.Email, user.Status, time.Now(), user.LastLogin, user.ID).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return interfaces.Users{}, err
	}
	return user, nil
}

func (dw *DatabaseWrapper) Delete(user interfaces.Users) error {
	query := `DELETE FROM users WHERE id=$1 AND user_id=$2`
	_, err := DBPool.Exec(context.Background(), query, user.ID, user.UserID)
	return err
}
