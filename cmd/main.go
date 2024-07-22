package main

import (
	// Standard Library

	"image/color"
	"log"
	"os"
	"path/filepath"

	// Fyne Imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	// External Imports
	"github.com/joho/godotenv"

	// Internal Imports
	myAuth "github.com/j4m1n-t/goAudit/internal/authentication"
	crud "github.com/j4m1n-t/goAudit/internal/databases"
	myFunctions "github.com/j4m1n-t/goAudit/internal/functions"
	state "github.com/j4m1n-t/goAudit/internal/status"
	myLayout "github.com/j4m1n-t/goAudit/pkg/tablayout"
)

// Theme structure
type myTheme struct{}

// Set the theme colors
func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xf0, G: 0xf0, B: 0xff, A: 0xff} // Light blue
		}
		return color.NRGBA{R: 0x20, G: 0x20, B: 0x40, A: 0xff} // Dark blue
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Set the theme icon
func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Set the theme font
func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Set the theme size
func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Set the toggle function for the theme
var isCustomTheme bool = false

func toggleTheme(a fyne.App) {
	if isCustomTheme {
		a.Settings().SetTheme(theme.DefaultTheme())
		isCustomTheme = false
	} else {
		a.Settings().SetTheme(&myTheme{})
		isCustomTheme = true
	}
}

var (
	tabs           *container.AppTabs
	configPath     string
	adminTab       fyne.CanvasObject
	auditTab       fyne.CanvasObject
	credentialsTab fyne.CanvasObject
	crmTab         fyne.CanvasObject
	notesTab       fyne.CanvasObject
	tasksTab       fyne.CanvasObject
)

func main() {
	//Load environment variables
	env := godotenv.Load()
	if env != nil {
		log.Fatalf("Error loading.env file.")
	}
	// Set logging and configuration
	myFunctions.SetLog()
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Failed to get user config dir: %v", err)
		fyne.LogError("Failed to get user config dir.", err)
	}
	log.Printf("Config directory: %s", configDir)
	configPath = filepath.Join(configDir, "goAudit", "config.json")
	// Initialize connection to db server(s)
	crud.InitDBs()
	// Set the default app layout
	myApp := app.New()
	myWindow := myApp.NewWindow("goAudit")
	myWindow.SetTitle("goAudit")
	myWindow.Resize(fyne.NewSize(800, 700))
	myWindow.SetPadded(true)
	// Icons and mutible items

	// Menu Items
	// Check if ldap is configured
	if !myFunctions.LDAPConfigured() {
		myFunctions.ShowLDAPDialog(myWindow)
		if !myFunctions.LDAPConfigured() {
			errorMessage := widget.NewLabel("The LDAP must be configured to use the application.")
			errorMessage.Wrapping = fyne.TextWrapWord

			content := container.NewVBox(
				errorMessage,
			)

			customDialog := dialog.NewCustom("Configuration Error", "OK", content, myWindow)
			customDialog.Resize(fyne.NewSize(300, 150))
			customDialog.Show()
		}
	}
	Menu := fyne.NewMainMenu()
	FileMenu := fyne.NewMenu("File")
	QuitItem := fyne.NewMenuItem("Quit", func() {
		// Graceful shutdown handling
		if state.GlobalState.LDAPConn != nil && state.GlobalState.LDAPConn.Conn != nil {
			state.GlobalState.LDAPConn.Conn.Close()
		}
		os.Exit(0)
	})
	LogoutItem := fyne.NewMenuItem("Logout", func() {
		myAuth.LogoutUser(state.GlobalState.LDAPConn)
	})
	SettingsMenu := fyne.NewMenu("Settings")
	ThemeItem := fyne.NewMenuItem("Toggle Theme", func() { toggleTheme(myApp) })
	SettingsMenu.Items = append(SettingsMenu.Items, ThemeItem)
	FileMenu.Items = append(FileMenu.Items, LogoutItem, QuitItem)
	Menu.Items = append(Menu.Items, FileMenu)
	Menu.Items = append(Menu.Items, SettingsMenu)
	myWindow.SetMainMenu(Menu)

	// Tabs

	username := widget.NewEntry()
	password := widget.NewPasswordEntry()
	loginForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Username: ", Widget: username},
			{Text: "Password: ", Widget: password},
		},
		OnSubmit: func() {
			var err error
			state.GlobalState.LDAPConn, err = myAuth.ConnectToAdServer(username.Text, password.Text)
			if err != nil {
				dialog.ShowError(err, myWindow)
				fyne.LogError("Error connecting to LDAP server.", err)
				return
			}
			state.GlobalState.Username = username.Text
			state.GlobalState.FetchAll()
			log.Printf("Notes fetched for user: %s: %+v", state.GlobalState.Username, state.GlobalState.Notes)
			myFunctions.UpdateTabsForUser(myWindow)
			isAdmin := myFunctions.CheckIfAdmin(state.GlobalState.LDAPConn, username.Text)
			myFunctions.UpdateMenuForUser(isAdmin, myWindow)
			tabs.SelectIndex(1)
		},
	}

	adminTab = myLayout.CreatePlaceholderAdminTab()
	auditTab = myLayout.CreatePlaceholderAuditsTab()
	credentialsTab = myLayout.CreatePlaceholderCredentialsTab()
	crmTab = myLayout.CreatePlaceholderCRMTab()
	notesTab = myLayout.CreatePlaceholderNotesTab()
	tasksTab = myLayout.CreatePlaceholderTaskTab()
	tabs = container.NewAppTabs(
		container.NewTabItem("Admin", adminTab),
		container.NewTabItem("Audits", auditTab),
		container.NewTabItem("CRM", crmTab),
		container.NewTabItem("Credentials", credentialsTab),
		container.NewTabItem("Login", loginForm),
		container.NewTabItem("Notes", notesTab),
		container.NewTabItem("Tasks", tasksTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	tabs.SelectIndex(4)
	myWindow.SetContent(tabs)

	// Show and run the application
	myWindow.ShowAndRun()
	// Graceful shutdown
	myWindow.SetOnClosed(func() { os.Exit(0) })
	myApp.Quit()
	os.Exit(0)
}
