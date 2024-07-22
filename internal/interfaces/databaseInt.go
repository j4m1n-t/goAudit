package interfaces

import "time"

type DatabaseOperations interface {
	GetUsers(username string) ([]Users, string, error)
	GetNotes(username string) ([]Note, string, error)
	GetTasks(username string) ([]Tasks, string, error)
	GetAudits(username string) ([]Audits, string, error)
	GetCRMEntries(username string) ([]CRM, string, error)
	GetCredentials(username string) ([]Credentials, string, error)
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
	ID             int       `json:"id"`
	Site           string    `json:"site"`
	Program        string    `json:"program"`
	UserID         int       `json:"-"`
	Username       string    `json:"username"`
	MasterPassword string    `json:"master_password"`
	LoginName      string    `json:"login_name"`
	LoginPass      string    `json:"login_pass"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Owner          string    `json:"owner"`
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
