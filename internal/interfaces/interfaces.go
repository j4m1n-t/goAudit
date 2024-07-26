package interfaces

import (
	// Standard Library
	"sync"
	"time"

	// External Imports
	"github.com/go-ldap/ldap/v3"
)

// Database Interface

type DatabaseOperations interface {
	// Users
	Create(username string) (Users, error)
	GetUsers(username string) ([]Users, string, error)
	GetOrCreateUser(username string) (Users, error)
	GetAll() ([]Users, error)
	Update(user Users) (Users, error)
	Delete(user Users) error
	// Notes
	GetNote(id int) (Note, error)
	GetNotes(username string) ([]Note, string, error)
	UpdateNote(note Note) (Note, error)
	DeleteNote(id int) error
	CreateNote(title, content string, username string, open bool) (Note, error)
	SearchNotes(searchTerm string, username string) ([]Note, string, error)
	// Tasks
	GetTasks(username string) ([]Tasks, string, error)
	CreateTask(task Tasks) (Tasks, error)
	UpdateTask(task Tasks) (Tasks, error)
	DeleteTask(id int, username string) error
	// Audits
	GetAudits(username string) ([]Audits, string, error)
	DeleteAudit(id int, username string) error
	UpdateAudit(audit Audits) (Audits, error)
	CreateAudit(audit Audits) (Audits, error)
	// CRM
	GetCRMEntries(username string) ([]CRM, string, error)
	DeleteCRMEntry(id int, username string) error
	UpdateCRMEntry(crm CRM) (CRM, error)
	CreateCRMEntry(crm CRM) (CRM, error)
	// Credentials
	GetCredentials(username string) ([]Credentials, string, error)
	GetCredentialByLoginName(loginName string) ([]Credentials, error)
	CreateCredential(credential Credentials) (Credentials, error)
	UpdateCredential(credential Credentials) (Credentials, error)
	DeleteCredential(id int, owner string) error
	SearchCredentials(searchTerm, owner string) ([]Credentials, string, error)
	CreateCredUser(username string, hashedPassword string, email string) (*Credentials, error)
	GetUserPassword(username string) (string, error)
}

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    int       `json:"-"`
	Username  string    `json:"username"`
	Open      bool      `json:"open"`
	Author    string    `json:"author"`
}

type Tasks struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	Priority    int       `json:"priority"`
	Notes       string    `json:"notes"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      int       `json:"-"`
	Username    string    `json:"username"`
}

type Audits struct {
	ID              int       `json:"id"`
	Action          string    `json:"action"`
	AuditID         int       `json:"audit_id"`
	AuditType       string    `json:"audit_type"`
	AuditArea       string    `json:"audit_area"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Notes           string    `json:"notes"`
	AssignedUser    string    `json:"assigned_user"`
	CompletedAt     time.Time `json:"completed_at"`
	Completed       bool      `json:"completed"`
	UserID          int       `json:"-"`
	Username        string    `json:"username"`
	AdditionalUsers []string  `json:"additional_users"`
	Firm            string    `json:"firm"`
}

type CRM struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Company   string    `json:"company"`
	Notes     []string  `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Open      bool      `json:"open"`
	UserID    int       `json:"-"`
	Username  string    `json:"username"`
}

type Credentials struct {
	ID              int       `json:"id"`
	Site            string    `json:"site"`
	Program         string    `json:"program"`
	UserID          int       `json:"-"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	MasterPassword  string    `json:"master_password"`
	LoginName       string    `json:"login_name"`
	LoginPass       string    `json:"login_pass"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Owner           string    `json:"owner"`
	PasswordHistory []string  `json:"password_history"`
}

type Users struct {
	ID        int       `json:"id"`
	UserID    int       `json:"-"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Admin     bool      `json:"admin"`
	LastLogin time.Time `json:"last_login"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
}

// LDAP Interface

type LDAPConnection struct {
	Conn     *ldap.Conn
	Username string
	Password string
	Server   string
	Domain   string
}

type LDAPUser struct {
	DN       string
	Username string
}
type CredentialAuth struct {
	IsAuthenticated bool
	Username        string
}

var (
	credentialAuth     CredentialAuth
	credentialAuthLock sync.RWMutex
)

type LDAPOperations interface {
	// LDAPOperations
	ConnectToAdServer(username, password string) (*LDAPConnection, error)
	LogoutUser(c *LDAPConnection) error
	DomaintoOU(domain string) string
	IsProperDomain(domain string) bool
	OUwithDomain(ou string, domain string) string
	// Master Password Operations
	VerifyCredentialAccess(loginName, loginPass string) error
}
