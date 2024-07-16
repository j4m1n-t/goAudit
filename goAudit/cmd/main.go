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

	// Internal Imports
	myFunctions "github.com/j4m1n-t/goAudit/goAudit/internal/functions"
	myLayout "github.com/j4m1n-t/goAudit/goAudit/internal/layouts"
	serverSide "github.com/j4m1n-t/goAudit/goAuditServer/pkg"
	crud "github.com/j4m1n-t/goAudit/goAuditServer/pkg/CRUD"
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
	ldapConn   *serverSide.LDAPConnection
	tabs       *container.AppTabs
	configPath string
)

func main() {
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
	crud.InitDBNotes()

	// Set the default app layout
	myApp := app.New()
	myWindow := myApp.NewWindow("goAudit")
	myWindow.SetTitle("goAudit")
	myWindow.Resize(fyne.NewSize(800, 600))
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
		if ldapConn != nil && ldapConn.Conn != nil {
			ldapConn.Conn.Close()
		}
		os.Exit(0)
	})
	LogoutItem := fyne.NewMenuItem("Logout", func() {
		serverSide.LogoutUser(ldapConn)
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
			ldapConn, err = serverSide.ConnectToAdServer(username.Text, password.Text)
			if err != nil {
				dialog.ShowError(err, myWindow)
				fyne.LogError("Error connecting to LDAP server.", err)
				return
			}
			isAdmin := myFunctions.CheckIfAdmin(ldapConn, username.Text)
			myFunctions.UpdateMenuForUser(isAdmin, myWindow)
			// Create a scroll container with a minimum size
			// Switch to the 'Search' tab after successful login
			tabs.SelectIndex(1)
		},
	}

	NotesTab := myLayout.CreateNotesTabContent(myWindow)
	TasksTab := container.NewVBox(
		widget.NewLabel("Tasks"),
		widget.NewLabel("Users tasks will show in this area."),
	)
	tabs = container.NewAppTabs(
		// container.NewTabItem("Audits", auditTab),
		// container.NewTabItem("CRM", crmTab),
		// container.NewTabItem("Credentials", credentialsTab),
		container.NewTabItem("Login", loginForm),
		container.NewTabItem("Notes", NotesTab),
		container.NewTabItem("Tasks", TasksTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	myWindow.SetContent(tabs)

	// Show and run the application
	myWindow.ShowAndRun()
	// Graceful shutdown
	myWindow.SetOnClosed(func() { os.Exit(0) })
	myApp.Quit()
	os.Exit(0)
}
