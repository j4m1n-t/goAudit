package layouts

import (
	// Standard Library
	"errors"

	// Fyne Imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	// Internal Imports
	auth "github.com/j4m1n-t/goAudit/internal/authentication"
	interfaces "github.com/j4m1n-t/goAudit/internal/interfaces"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

var credentialsList *widget.List

func CreatePlaceholderCredentialsTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view credentials."),
	)
}

func CreateCredentialsTabContent(window fyne.Window) fyne.CanvasObject {
	var appState *state.AppState
	if state.GlobalState.UserID == 0 {
		// Check if the user has a master password set
		err := auth.CheckIfMPPresent(appState)
		if err != nil {
			dialog.ShowError(err, window)
			return nil
		}

		// Show the appropriate dialog based on master password presence
		if state.GlobalState.MPPresent {
			ShowLoginDialog(window, appState)
		} else {
			showSignUpDialog(window)
		}
		return nil
	}

	newCredentialButton := widget.NewButton("New Credential", func() {
		showCredentialDialog(window, nil)
	})

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search credentials...")

	searchButton := widget.NewButton("Search", func() {
		searchCredentials(window, searchEntry.Text)
	})

	credentialsList = widget.NewList(
		func() int { return len(state.GlobalState.Credentials) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.HomeIcon()),
				widget.NewLabel("Site"),
				widget.NewLabel("Username"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(state.GlobalState.Credentials) {
				cred := state.GlobalState.Credentials[id]
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(cred.Site)
				item.(*fyne.Container).Objects[2].(*widget.Label).SetText(cred.Username)
			}
		},
	)

	credentialsList.OnSelected = func(id widget.ListItemID) {
		if id < len(state.GlobalState.Credentials) {
			showCredentialDialog(window, &state.GlobalState.Credentials[id])
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Credentials"),
			container.NewHBox(searchEntry, searchButton),
			newCredentialButton,
		),
		nil, nil, nil,
		credentialsList,
	)
}

func showCredentialDialog(window fyne.Window, credential *interfaces.Credentials) {
	var siteEntry, programEntry, usernameEntry, emailEntry, loginNameEntry, loginPassEntry *widget.Entry

	siteEntry = widget.NewEntry()
	siteEntry.SetPlaceHolder("Enter site")

	programEntry = widget.NewEntry()
	programEntry.SetPlaceHolder("Enter program")

	usernameEntry = widget.NewEntry()
	usernameEntry.SetPlaceHolder("Enter username")

	emailEntry = widget.NewEntry()
	emailEntry.SetPlaceHolder("Enter email")

	loginNameEntry = widget.NewEntry()
	loginNameEntry.SetPlaceHolder("Enter login name")

	loginPassEntry = widget.NewPasswordEntry()
	loginPassEntry.SetPlaceHolder("Enter login password")

	if credential != nil {
		siteEntry.SetText(credential.Site)
		programEntry.SetText(credential.Program)
		usernameEntry.SetText(credential.Username)
		emailEntry.SetText(credential.Email)
		loginNameEntry.SetText(credential.LoginName)
		loginPassEntry.SetText(credential.LoginPass)
	}

	saveButton := widget.NewButton("Save", func() {
		if credential == nil {
			newCredential := interfaces.Credentials{
				Site:      siteEntry.Text,
				Program:   programEntry.Text,
				Username:  usernameEntry.Text,
				UserID:    state.GlobalState.UserID,
				Email:     emailEntry.Text,
				LoginName: loginNameEntry.Text,
				LoginPass: loginPassEntry.Text,
				Owner:     state.GlobalState.Username,
			}
			_, err := state.GlobalState.DB.CreateCredential(newCredential)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		} else {
			credential.Site = siteEntry.Text
			credential.Program = programEntry.Text
			credential.Username = usernameEntry.Text
			credential.Email = emailEntry.Text
			credential.LoginName = loginNameEntry.Text
			credential.LoginPass = loginPassEntry.Text
			_, err := state.GlobalState.DB.UpdateCredential(*credential)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		}

		refreshCredentials(window)
		dialog.ShowInformation("Success", "Credential saved successfully", window)
	})

	var buttons fyne.CanvasObject
	if credential != nil {
		deleteButton := widget.NewButton("Delete", func() {
			dialog.ShowConfirm("Confirm Delete", "Are you sure you want to delete this credential?", func(confirm bool) {
				if confirm {
					err := state.GlobalState.DB.DeleteCredential(credential.ID, state.GlobalState.Username)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					refreshCredentials(window)
					dialog.ShowInformation("Success", "Credential deleted successfully", window)
				}
			}, window)
		})
		buttons = container.NewHBox(saveButton, deleteButton)
	} else {
		buttons = container.NewHBox(saveButton)
	}

	content := container.NewVBox(
		widget.NewLabel("Site"),
		siteEntry,
		widget.NewLabel("Program"),
		programEntry,
		widget.NewLabel("Username"),
		usernameEntry,
		widget.NewLabel("Email"),
		emailEntry,
		widget.NewLabel("Login Name"),
		loginNameEntry,
		widget.NewLabel("Login Password"),
		loginPassEntry,
		buttons,
	)

	dialog.ShowCustom("Credential Details", "Close", content, window)
}

func refreshCredentials(window fyne.Window) {
	credentials, _, err := state.GlobalState.DB.GetCredentials(state.GlobalState.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	state.GlobalState.Credentials = credentials
	credentialsList.Refresh()
}

func searchCredentials(window fyne.Window, searchTerm string) {
	credentials, message, err := state.GlobalState.DB.SearchCredentials(searchTerm, state.GlobalState.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	state.GlobalState.Credentials = credentials
	credentialsList.Refresh()

	if message != "" {
		dialog.ShowInformation("Search Results", message, window)
	}
}

func ShowLoginDialog(window fyne.Window, appState *state.AppState) {
	usernameEntry := widget.NewEntry()
	passwordEntry := widget.NewPasswordEntry()

	dialog.ShowForm("Login", "Login", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Username", usernameEntry),
		widget.NewFormItem("Password", passwordEntry),
	}, func(res bool) {
		if res {
			user, err := auth.AuthenticateUser(state.GlobalState.DB, usernameEntry.Text, passwordEntry.Text)
			if err != nil {
				showSignUpDialog(window)
				dialog.ShowError(err, window)
				return
			}
			state.GlobalState.UserID = user.UserID
			state.GlobalState.Username = user.Username
			window.SetContent(CreateCredentialsTabContent(window))
		} else {
			showSignUpDialog(window)
		}
	}, window)
}

func showSignUpDialog(window fyne.Window) {
	// Create Entry widgets
	usernameEntry := widget.NewEntry()
	passwordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry := widget.NewPasswordEntry()
	emailEntry := widget.NewEntry()

	// Set the size of the entries
	usernameEntry.Resize(fyne.NewSize(450, 50))
	passwordEntry.Resize(fyne.NewSize(450, 50))
	confirmPasswordEntry.Resize(fyne.NewSize(450, 50))
	emailEntry.Resize(fyne.NewSize(450, 50))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
			{Text: "Confirm Password", Widget: confirmPasswordEntry},
			{Text: "Email", Widget: emailEntry},
		},
		OnSubmit: func() {
			if passwordEntry.Text != confirmPasswordEntry.Text {
				dialog.ShowError(errors.New("passwords do not match"), window)
				return
			}

			hashedPassword, err := auth.HashPassword(passwordEntry.Text)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			newUser, err := state.GlobalState.DB.CreateCredUser(usernameEntry.Text, hashedPassword, emailEntry.Text)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			state.GlobalState.UserID = newUser.ID

			state.GlobalState.Username = usernameEntry.Text
			state.GlobalState.SetMasterPassword(state.GlobalState.Username, passwordEntry.Text)

			window.SetContent(CreateCredentialsTabContent(window))
		},
		OnCancel: func() {
			// Handle cancel action
		},
	}

	// Show the dialog with the larger form
	d := dialog.NewCustom("Sign Up", "Cancel", form, window)
	d.Resize(fyne.NewSize(500, 300)) // Set a custom size for the dialog if needed
	d.Show()
}

func showSetMasterPasswordDialog(window fyne.Window) {
	passwordEntry := widget.NewPasswordEntry()
	confirmEntry := widget.NewPasswordEntry()

	dialog.ShowForm("Set Master Password", "Set", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Password", passwordEntry),
		widget.NewFormItem("Confirm Password", confirmEntry),
	}, func(set bool) {
		if set {
			if passwordEntry.Text == confirmEntry.Text {
				state.GlobalState.SetMasterPassword(state.GlobalState.Username, passwordEntry.Text)
				window.SetContent(CreateCredentialsTabContent(window))
			} else {
				dialog.ShowError(errors.New("passwords do not match"), window)
			}
		}
	}, window)
}
