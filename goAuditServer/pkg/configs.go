package serverfunc

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

func SaveSQLSettings(server, user, password, database string) error {
	envContent := fmt.Sprintf("SQL_SERVER=%s\nSQL_USER=%s\nSQL_PASSWORD=%s\nSQL_DATABASE=%s\n", server, user, password, database)
	return os.WriteFile(".env", []byte(envContent), 0644)
}

func LoadSQLSettings() SQLSettings {
	return SQLSettings{
		Server:   os.Getenv("SQL_SERVER"),
		User:     os.Getenv("SQL_USER"),
		Password: os.Getenv("SQL_PASSWORD"),
		Database: os.Getenv("SQL_DATABASE"),
	}
}

func SavegoAuditServerSettings(inUse bool, server string) error {
	inUseStr := "false"
	if inUse {
		inUseStr = "true"
	}
	envContent := fmt.Sprintf("goAudit_Server_INUSE=%s\ngoAudit_SERVER=%s\n", inUseStr, server)
	return os.WriteFile(".env", []byte(envContent), 0644)
}
