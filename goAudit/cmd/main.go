package main

import (
	// Standard Library
	"log"
	"os"
	"path/filepath"

	// Fyne Imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// External Imports

	// Internal Imports
	clientSide "github.com/j4m1n-t/goAudit/goAudit/internal"
	serverSide "github.com/j4m1n-t/goAudit/goAuditServer/pkg"
)

var (
	ldapConn   *serverSide.LDAPConnection
	tabs       *container.AppTabs
	configPath string
)

func main() {

	// Set logging
	// config := clientSide.LoadConfig()
	clientSide.SetLog()
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Failed to get user config dir: %v", err)
		fyne.LogError("Failed to get user config dir.", err)
	}
	log.Printf("Config directory: %s", configDir)
	configPath = filepath.Join(configDir, "goAudit", "config.json")
	// Set the default app layout
	myApp := app.New()
	myWindow := myApp.NewWindow("goAudit")
	myWindow.SetTitle("goAudit")
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.SetPadded(true)
	// Icons and mutible items

	// Menu Items
	done := make(chan struct{})
	Menu := fyne.NewMainMenu()
	FileMenu := fyne.NewMenu("File")
	QuitItem := fyne.NewMenuItem("Quit", func() {
		// Graceful shutdown handling
		myWindow.Close()
		myApp.Quit()
		<-done
		if ldapConn != nil && ldapConn.Conn != nil {
			ldapConn.Conn.Close()
		}
		os.Exit(0)
	})
	SettingsMenu := fyne.NewMenu("Settings")
	FileMenu.Items = append(FileMenu.Items, QuitItem)
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
			// Create a scroll container with a minimum size
			// Switch to the 'Search' tab after successful login
			tabs.SelectIndex(1)
		},
	}
	NotesTab := container.NewVBox(
		widget.NewLabel("Notes"),
		widget.NewLabel("Users notes will show in this area."),
	)
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
