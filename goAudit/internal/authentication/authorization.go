package auth

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
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	server := os.Getenv("LDAP_SERVER")
	domain := os.Getenv("LDAP_DOMAIN")
	ou := os.Getenv("LDAP_OU")
	readonlyPassword := os.Getenv("LDAP_READONLY_PASSWORD")

	if server == "" || domain == "" || ou == "" || readonlyPassword == "" {
		return nil, fmt.Errorf("one or more required environment variables are not set")
	}

	ldapServer := fmt.Sprintf("ldaps://%s.%s:636", server, domain)
	log.Println("Connecting to:", ldapServer)

	l, err := ldap.DialURL(ldapServer)
	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAP server: %w", err)
	}

	// Construct the bindDN
	domainParts := strings.Split(domain, ".")
	dcString := strings.Join(domainParts, ",DC=")
	bindDN := fmt.Sprintf("CN=readonly,CN=Users,DC=%s", strings.Replace(domain, ".", ",DC=", -1))

	log.Printf("Binding as read-only user: %s", bindDN)
	err = l.Bind(bindDN, readonlyPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to LDAP server as read-only: %w", err)
	}

	// Search for user
	searchFilter := fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))", ldap.EscapeFilter(username))
	searchBase := fmt.Sprintf("OU=%s,DC=%s", ou, dcString)
	searchRequest := ldap.NewSearchRequest(
		searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"distinguishedName"},
		nil,
	)
	log.Printf("Searching for user with filter: %s in base: %s", searchFilter, searchBase)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user: %w", err)
	}
	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("user not found or too many entries returned")
	}

	userDN := sr.Entries[0].DN
	log.Printf("Found user: %s", userDN)

	// Now bind as the user
	log.Printf("Attempting to bind as user: %s", userDN)
	err = l.Bind(userDN, password)
	if err != nil {
		return nil, fmt.Errorf("failed to bind as user: %w", err)
	}

	log.Println("Successfully authenticated user")
	return &LDAPConnection{
		Conn:     l,
		Username: username,
		Password: password,
		Server:   server,
		Domain:   domain,
	}, nil
}

// Logout as current user
func LogoutUser(c *LDAPConnection) error {
	log.Println("Logging out as:", c.Username)
	return c.Conn.Close()
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
	return bdn
}

