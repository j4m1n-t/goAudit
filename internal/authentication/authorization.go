package auth

import (
	// Standard Library
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	// Fyne Imports

	// External Imports
	"github.com/go-ldap/ldap/v3"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	// Internal Imports
	"github.com/j4m1n-t/goAudit/internal/interfaces"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

type LDAPWrapper struct{}

type Auth struct {
	DB   interfaces.DatabaseOperations
	LDAP interfaces.LDAPOperations
}

func NewAuth(db interfaces.DatabaseOperations, ldap interfaces.LDAPOperations) *Auth {
	return &Auth{
		DB:   db,
		LDAP: ldap,
	}
}

// Connect to the LDAP server using the provided username and password
func (lw *LDAPWrapper) ConnectToAdServer(username, password string) (*interfaces.LDAPConnection, error) {
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
	return &interfaces.LDAPConnection{
		Conn:     l,
		Username: username,
		Password: password,
		Server:   server,
		Domain:   domain,
	}, nil
}

// Logout as current user
func (lw *LDAPWrapper) LogoutUser(c *interfaces.LDAPConnection) error {
	log.Println("Logging out as:", c.Username)
	return c.Conn.Close()
}

// Formats Domain to OU style for ldap
func (lw LDAPWrapper) DomaintoOU(domain string) string {

	parts := strings.Split(domain, ".")
	var ous []string
	for _, part := range parts {
		ous = append(ous, "DC="+strings.ToLower(part))
	}
	return strings.Join(ous, ",")
}

// Checks if the domain is in the proper format to convert to OU
func (lw *LDAPWrapper) IsProperDomain(domain string) bool {

	parts := strings.Split(domain, ".")
	return len(parts) >= 2 && len(parts) <= 3
}

// Combine OU and Domain to create a Base DN for search
func (lw *LDAPWrapper) OUwithDomain(ou string, domain string) string {

	bdn := fmt.Sprintf("OU=%s,%s", ou, lw.DomaintoOU(domain))
	return bdn
}

// Authenticate with Master Password for use of credentials module

// VerifyCredentialAccess verifies the credential-specific login
func (lw *LDAPWrapper) VerifyCredentialAccess(loginName, loginPass string) error {
	// Fetch the credential from the database
	creds, err := state.GlobalState.DB.GetCredentialByLoginName(loginName)
	if err != nil {
		return fmt.Errorf("credential not found: %w", err)
	}
	// Check if there are any credentials
	if len(creds) == 0 {
		return fmt.Errorf("no credentials found")
	}
	// User the first credential
	cred := creds[0]
	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(cred.LoginPass), []byte(loginPass))
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	// Set the credential authentication status in the global state
	state.GlobalState.SetCredentialAuthenticated(true, cred.Username)
	return nil
}

// CreateCredential creates a new credential entry
func CreateMPCredential(cred *interfaces.Credentials) error {
	// Hash the password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(cred.LoginPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	cred.LoginPass = string(hashedPass)

	// Save the credential to the database
	newCred, err := state.GlobalState.DB.CreateCredential(*cred)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	*cred = newCred

	return nil
}

func AuthenticateUser(db interfaces.DatabaseOperations, username, password string) (*interfaces.Users, error) {
	users, _, err := db.GetUsers(username)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	user := users[0]

	// You'll need to implement this method to get the hashed password for the user
	hashedPassword, err := db.GetUserPassword(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}

func CreateUser(db interfaces.DatabaseOperations, username, password string) (*interfaces.Credentials, error) {
	// Check if the username already exists
	existingUser, _, err := db.GetUsers(username)
	if err != nil {
		return nil, err
	}
	if len(existingUser) > 0 {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create the new user
	newUser := &interfaces.Credentials{
		Username: username,
		// Add other fields as necessary
	}

	// Save the new user to the database
	// You'll need to implement this method in your DatabaseOperations interface
	newUser, err = db.CreateCredUser(newUser.Username, string(hashedPassword), newUser.Email)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}
