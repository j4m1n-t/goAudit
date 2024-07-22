package databases

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

type DatabaseWrapper struct {
	Pool *pgxpool.Pool
}

type SQLSettings struct {
	Server   string
	User     string
	Password string
	Database string
	Port     string
}

func LoadSQLSettings() SQLSettings {
	return SQLSettings{
		Server:   os.Getenv("SQL_SERVER"),
		User:     os.Getenv("SQL_USER"),
		Password: os.Getenv("SQL_PASSWORD"),
		Database: os.Getenv("SQL_DATABASE"),
		Port:     os.Getenv("SQL_PORT"),
	}
}

func InitDB() (*DatabaseWrapper, error) {
	SQLSettings := LoadSQLSettings()

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		SQLSettings.User, SQLSettings.Password, SQLSettings.Server, SQLSettings.Port, SQLSettings.Database)

	var err error
	DBPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = DBPool.Ping(context.Background())
	if err != nil {
		DBPool.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	log.Println("Successfully connected to the database.")

	dbWrapper := &DatabaseWrapper{Pool: DBPool}

	return dbWrapper, nil
}
