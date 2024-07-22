package functions

import (
	//Standard Library Imports//
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	// Fyne Imports//
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// External Imports
	"github.com/go-ldap/ldap/v3"
	"github.com/joho/godotenv"

	// Internal Imports
	myAuth "github.com/j4m1n-t/goAudit/internal/authentication"
	layouts "github.com/j4m1n-t/goAudit/internal/layouts"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

type AppConfig struct {
	IconPath   string `json:"iconPath"`
	ConfigPath string `json:"config"`
}

var configPath string

func SetLog() {
	AppDataDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Failed to get user AppData dir: %v\n", err)
		return
	}
	logDir := filepath.Join(AppDataDir, "goAudit")
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}
	logFile, err := os.OpenFile(filepath.Join(logDir, "goAudit.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Logging initialized")
}

func SaveConfig(config AppConfig) error {
	err := os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&config)
	if err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	return nil
}

func LoadConfig() AppConfig {
	config := AppConfig{}
	configDir, err := os.UserConfigDir()
	if err != nil {
		fyne.LogError("Failed to get user config dir.", err)
	}
	configPath = filepath.Join(configDir, "goAudit", "config.json")
	iconPath := filepath.Join(configDir, "goAudit", "goAudit.ico")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the directory if it doesn't exist
			err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
			if err != nil {
				log.Printf("Failed to create config directory: %v", err)
			}
			// Return default config as the file doesn't exist yet
			return config
		}
		log.Printf("Failed to open config file: %v", err)
		return config
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("Failed to decode config file: %v", err)
		// Return default config if decoding fails
		return AppConfig{ConfigPath: configPath, IconPath: iconPath}
	}
	return config
}

func LDAPConfigured() bool {
	_, err := os.Stat(".env")
	if os.IsNotExist(err) {
		return false
	}

	// Check if all necessary LDAP settings are present
	envMap, err := godotenv.Read(".env")
	if err != nil {
		return false
	}

	requiredKeys := []string{"LDAP_SERVER", "LDAP_DOMAIN", "LDAP_OU"}
	for _, key := range requiredKeys {
		if _, exists := envMap[key]; !exists {
			return false
		}
	}

	return true
}

func ShowLDAPDialog(window fyne.Window) {

	serverEntry := widget.NewEntry()
	serverEntry.SetPlaceHolder("LDAP Server")

	domainEntry := widget.NewEntry()
	domainEntry.SetPlaceHolder("LDAP Domain")

	ouEntry := widget.NewEntry()
	ouEntry.SetPlaceHolder("LDAP OU")

	readonlyEntry := widget.NewEntry()
	readonlyEntry.SetPlaceHolder("Read Only Password")

	// Load the current LDAP settings
	ldapSettings := LoadLDAPSettings()

	// Set the current LDAP settings in the dialog fields
	serverEntry.SetText(ldapSettings.Server)
	domainEntry.SetText(ldapSettings.Domain)
	ouEntry.SetText(ldapSettings.OU)
	maskedPassword := strings.Repeat("*", len(ldapSettings.ReadOnlyPassword))
	readonlyEntry.SetText(maskedPassword)

	// Show Password Toggle button
	showPasswordButton := widget.NewButton("Show/Hide Password", func() {
		if readonlyEntry.Password {
			readonlyEntry.Password = false
			readonlyEntry.SetText(ldapSettings.ReadOnlyPassword)
		} else {
			readonlyEntry.Password = true
			readonlyEntry.SetText(maskedPassword)
		}
		readonlyEntry.Refresh()
	})

	content := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("LDAP Server", serverEntry),
			widget.NewFormItem("LDAP Domain", domainEntry),
			widget.NewFormItem("LDAP OU", ouEntry),
			widget.NewFormItem("Read Only Password", readonlyEntry),
		),
	)

	// Create a custom dialog
	customDialog := dialog.NewCustom("LDAP Settings", "Save", content, window)

	// Add a cancel button
	cancelButton := widget.NewButton("Cancel", func() {
		customDialog.Hide()
	})

	// Create a save button
	saveButton := widget.NewButton("Save", func() {
		err := SaveLDAPSettings(serverEntry.Text, domainEntry.Text, ouEntry.Text, readonlyEntry.Text)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		dialog.ShowInformation("Success", "LDAP settings saved successfully", window)
		customDialog.Hide()
	})

	// Create a container for buttons
	buttons := container.NewHBox(cancelButton, saveButton, showPasswordButton)

	// Set the buttons to the dialog
	customDialog.SetButtons([]fyne.CanvasObject{buttons})

	// Set a larger size for the dialog
	customDialog.Resize(fyne.NewSize(400, 300))

	// Show the dialog
	customDialog.Show()
}

func ShowSQLDialog(window fyne.Window) {
	serverEntry := widget.NewEntry()
	serverEntry.SetPlaceHolder("SQL Server")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("SQL User")

	passEntry := widget.NewEntry()
	passEntry.SetPlaceHolder("SQL Password")

	dbEntry := widget.NewEntry()
	dbEntry.SetPlaceHolder("Database Name")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Port Number")
	portEntry.SetPlaceHolder("5432")

	content := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("SQL Server", serverEntry),
			widget.NewFormItem("SQL User", userEntry),
			widget.NewFormItem("SQL Password", passEntry),
			widget.NewFormItem("Database Name", dbEntry),
			widget.NewFormItem("Port Number", portEntry),
		),
	)

	// Create a custom dialog
	customDialog := dialog.NewCustom("SQL Settings", "Save", content, window)

	// Add a cancel button
	cancelButton := widget.NewButton("Cancel", func() {
		customDialog.Hide()
	})

	// Create a save button
	saveButton := widget.NewButton("Save", func() {
		err := SaveSQLSettings(serverEntry.Text, userEntry.Text, passEntry.Text, dbEntry.Text, portEntry.Text)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		dialog.ShowInformation("Success", "SQL settings saved successfully", window)
		customDialog.Hide()
	})

	// Create a container for buttons
	buttons := container.NewHBox(cancelButton, saveButton)

	// Set the buttons to the dialog
	customDialog.SetButtons([]fyne.CanvasObject{buttons})

	// Set a larger size for the dialog
	customDialog.Resize(fyne.NewSize(400, 300))

	// Show the dialog
	customDialog.Show()
}
func UpdateMenuForUser(isAdmin bool, window fyne.Window) {
	mainMenu := window.MainMenu()
	settingsMenu := mainMenu.Items[1] // Assuming Settings is the first menu

	if isAdmin {
		// Add Config item if it's not already there
		if len(settingsMenu.Items) == 1 {
			LDAPItem := fyne.NewMenuItem("LDAP Configuration", func() { ShowLDAPDialog(window) })
			SQLItem := fyne.NewMenuItem("SQL Configuration", func() { ShowSQLDialog(window) })
			settingsMenu.Items = append(settingsMenu.Items, LDAPItem, SQLItem)
		}
	} else {
		// Remove Config item if it's there
		if len(settingsMenu.Items) > 1 {
			settingsMenu.Items = settingsMenu.Items[1:]
		}
	}

	window.SetMainMenu(mainMenu)
}

func CheckIfAdmin(conn *myAuth.LDAPConnection, username string) bool {
	if conn == nil || conn.Conn == nil {
		log.Println("LDAP connection is nil")
		return false
	}

	searchBase := fmt.Sprintf("DC=%s", strings.Replace(conn.Domain, ".", ",DC=", -1))
	searchFilter := fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))", ldap.EscapeFilter(username))

	searchRequest := ldap.NewSearchRequest(
		searchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		searchFilter,
		[]string{"memberOf"},
		nil,
	)

	log.Printf("Searching for user with filter: %s in base: %s", searchFilter, searchBase)

	sr, err := conn.Conn.Search(searchRequest)
	if err != nil {
		log.Printf("Error searching for user: %v", err)
		fyne.LogError("Error searching for user", err)
		return false
	}

	if len(sr.Entries) == 0 {
		log.Printf("No entries found for user: %s", username)
		return false
	}

	for _, entry := range sr.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "memberOf" {
				for _, member := range attr.Values {
					log.Printf("User %s is member of: %s", username, member)
					if strings.Contains(strings.ToLower(member), "admin") {
						log.Printf("User %s is an admin (member of %s)", username, member)
						return true
					}
				}
			}
		}
	}

	log.Printf("User %s is not an admin", username)
	return false
}

func UpdateTabsForUser(window fyne.Window, appState *state.AppState) {
	var tabs *container.AppTabs
	// Set get content for tabs
	auditsTab := layouts.CreateAuditsTabContent(window)
	credentialsTab := layouts.CreateCredentialsTabContent(window, appState)
	crmTab := layouts.CreateCRMTabContent(window)
	notesTab := layouts.CreateNotesTabContent(window, appState)
	tasksTab := layouts.CreateTasksTabContent(window)
	// Set the content for tabs
	tabs.Items[2].Content = auditsTab
	tabs.Items[3].Content = credentialsTab
	tabs.Items[4].Content = crmTab
	tabs.Items[5].Content = notesTab
	tabs.Items[6].Content = tasksTab

}
