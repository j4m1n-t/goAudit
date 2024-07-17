package internal

import (
	// Standard Library
	"fmt"
	"os"
	// Fyne Imports
	// External Imports
	// Internal Imports
)

type LDAPSettings struct {
	Server           string
	Domain           string
	OU               string
	ReadOnlyPassword string
}

type SQLSettings struct {
	Server   string
	User     string
	Password string
	Database string
	Port     string
}

func SaveLDAPSettings(server, domain, ou, readOnlyPassword string) error {
	envContent := fmt.Sprintf("LDAP_SERVER=%s\nLDAP_DOMAIN=%s\nLDAP_OU=%s\nLDAP_READONLY_PASSWORD=%s\n", server, domain, ou, readOnlyPassword)
	return os.WriteFile(".env", []byte(envContent), 0644)
}

func LoadLDAPSettings() LDAPSettings {
	return LDAPSettings{
		Server:           os.Getenv("LDAP_SERVER"),
		Domain:           os.Getenv("LDAP_DOMAIN"),
		OU:               os.Getenv("LDAP_OU"),
		ReadOnlyPassword: os.Getenv("LDAP_READONLY_PASSWORD"),
	}
}

func SaveSQLSettings(server, user, password, database string, port string) error {
	envContent := fmt.Sprintf("SQL_SERVER=%s\nSQL_USER=%s\nSQL_PASSWORD=%s\nSQL_DATABASE=%s\nSQL_PORT=%s", server, user, password, database, port)
	return os.WriteFile(".env", []byte(envContent), 0644)
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
