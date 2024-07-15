package serverfunc

import (
	// Standard Library
	"fmt"
	"log"
	"os"
	"strings"

	// Fyne Imports

	// External Imports
	"github.com/go-ldap/ldap/v3"
	"github.com/joho/godotenv"
	// Internal Imports
)

// Set the structure for the ldap connection to
// be used throughout the program
type LDAPConnection struct {
	Conn     *ldap.Conn
	Username string
	Password string
	Server   string
	Domain   string
}

// Connect to the LDAP server using the provided username and password
func ConnectToAdServer(username, password string) (*LDAPConnection, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading.env file")
	}
	server := os.Getenv("LDAP_SERVER")
	domain := os.Getenv("LDAP_DOMAIN")
	ou := os.Getenv("LDAP_OU")
	if ou == "" {
		log.Fatal("LDAP_OU environment variable is not set")
	}
	if server == "" {
		log.Fatal("LDAP_SERVER environment variable is not set")
	}
	if domain == "" {
		log.Fatal("LDAP_DOMAIN environment variable is not set")
	}
	properUsername := strings.ToUpper(username)
	ldapServer := fmt.Sprintf("ldaps://%s.%s:636", server, domain) //Using LDAPS
	log.Println("Connecting to:", ldapServer)
	l, err := ldap.DialURL(ldapServer)
	if err != nil {
		log.Printf("Failed to dial LDAPserver %s: %v", ldapServer, err)
		return nil, fmt.Errorf("failed to dial LDAP server: %w", err)
	}

	bDN := OUwithDomain(ou, domain)
	bindDN := fmt.Sprintf("cn=%s,%s", properUsername, bDN) // bDN should allow for usuage of multiple organizational units
	err = l.Bind(bindDN, password)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	// Does the OU need to be returned and added to the structure?

	return &LDAPConnection{
		Conn:     l,
		Username: username,
		Password: password,
		Server:   server,
		Domain:   domain,
	}, nil
}

// Formats Domain to OU style for ldap
func DomaintoOU(domain string) string {
	parts := strings.Split(domain, ".")
	var ous []string
	for _, part := range parts {
		ous = append(ous, "DC="+strings.ToLower(part))
	}
	return strings.Join(ous, ",")
}

// Checks if the domain is in the proper format to convert to OU
func IsProperDomain(domain string) bool {
	parts := strings.Split(domain, ".")
	return len(parts) >= 2 && len(parts) <= 3
}

// Combine OU and Domain to create a Base DN for search
func OUwithDomain(ou string, domain string) string {
	bdn := fmt.Sprintf("OU=%s,%s", ou, DomaintoOU(domain))
	return strings.ToLower(bdn)
}
