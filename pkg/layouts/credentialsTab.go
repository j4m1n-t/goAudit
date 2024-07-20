package layouts

import (
	"errors"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	crud "github.com/j4m1n-t/goAudit/pkg/CRUD"
	state "github.com/j4m1n-t/goAudit/pkg/status"
)

var credentialsList *widget.List

func CreatePlaceholderCredentialsTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view your credentials."),
	)
}

func CreateCredentialsTabContent(window fyne.Window, appState *state.AppState) fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search credentials...")
	searchButton := widget.NewButton("Search", func() {
		performCredentialSearch(searchEntry.Text, window, appState)
	})

	// Fetch credentials
	appState.FetchCredentials()

	messageLabel := widget.NewLabel(appState.Message)

	credentialsList = widget.NewList(
		func() int {
			return len(appState.Credentials)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Login Name"),
				widget.NewLabel("Site"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(appState.Credentials) {
				credential := appState.Credentials[id]
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(credential.LoginName)
				item.(*fyne.Container).Objects[2].(*widget.Label).SetText(credential.Site)
			}
		},
	)

	newCredentialButton := widget.NewButton("New Credential", func() {
		showCredentialDialog(window, nil, appState)
	})

	credentialsList.OnSelected = func(id widget.ListItemID) {
		if id < len(appState.Credentials) {
			showCredentialDialog(window, &appState.Credentials[id], appState)
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Credentials"),
			container.NewBorder(nil, nil, nil, searchButton, searchEntry),
			messageLabel,
			newCredentialButton,
		),
		nil, nil, nil,
		credentialsList,
	)
}
func showCredentialDialog(window fyne.Window, credential *crud.Credentials, appState *state.AppState) {
	var loginNameEntry *widget.Entry
	var passwordEntry *widget.Entry
	var siteEntry *widget.Entry
	var programEntry *widget.Entry
	var rememberMeCheck *widget.Check
	var customDialog dialog.Dialog

	loginNameEntry = widget.NewEntry()
	loginNameEntry.SetPlaceHolder("Enter login name")

	passwordEntry = widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter password")

	siteEntry = widget.NewEntry()
	siteEntry.SetPlaceHolder("Enter site")

	programEntry = widget.NewEntry()
	programEntry.SetPlaceHolder("Enter program")

	rememberMeCheck = widget.NewCheck("Remember Me", nil)

	if credential != nil {
		loginNameEntry.SetText(credential.LoginName)
		passwordEntry.SetText(credential.Password)
		siteEntry.SetText(credential.Site)
		programEntry.SetText(credential.Program)
		rememberMeCheck.SetChecked(credential.RememberMe)
	}

	saveButton := widget.NewButton("Save", func() {
		if appState.Username == "" {
			dialog.ShowError(errors.New("user is not logged in"), window)
			return
		}

		if credential == nil {
			newCredential, err := crud.CreateCredential(
				loginNameEntry.Text,
				passwordEntry.Text,
				siteEntry.Text,
				programEntry.Text,
				appState.Username,
				rememberMeCheck.Checked,
			)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Created new credential with ID: %d for user: %s", newCredential.ID, appState.Username)
		} else {
			credential.LoginName = loginNameEntry.Text
			credential.Password = passwordEntry.Text
			credential.Site = siteEntry.Text
			credential.Program = programEntry.Text
			credential.RememberMe = rememberMeCheck.Checked
			updatedCredential, err := crud.UpdateCredential(*credential)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Updated credential with ID: %d", updatedCredential.ID)
		}
		appState.FetchCredentials()
		credentialsList.Refresh()
		customDialog.Hide()
	})

	deleteButton := widget.NewButton("Delete", func() {
		if credential != nil {
			confirmDialog := dialog.NewConfirm("Confirm Delete", "Are you sure you want to delete this credential?", func(confirm bool) {
				if confirm {
					err := crud.DeleteCredential(credential.ID, appState.Username)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					log.Printf("Deleted credential with ID: %d by username: %s", credential.ID, appState.Username)
					appState.FetchCredentials()
					credentialsList.Refresh()
					customDialog.Hide()
				}
			}, window)
			confirmDialog.Show()
		}
	})

	var buttons fyne.CanvasObject
	if credential != nil {
		buttons = container.NewHBox(saveButton, deleteButton)
	} else {
		buttons = container.NewHBox(saveButton)
	}

	content := container.NewVBox(
		widget.NewLabel("Login Name"),
		loginNameEntry,
		widget.NewLabel("Password"),
		passwordEntry,
		widget.NewLabel("Site"),
		siteEntry,
		widget.NewLabel("Program"),
		programEntry,
		rememberMeCheck,
	)

	paddedContent := container.NewPadded(content)
	mainContainer := container.NewBorder(nil, buttons, nil, nil, paddedContent)

	customDialog = dialog.NewCustom("Credential", "Cancel", mainContainer, window)
	customDialog.Resize(fyne.NewSize(400, 500))

	customDialog.SetOnClosed(func() {
		appState.FetchCredentials()
		credentialsList.Refresh()
		credentialsList.UnselectAll()
		log.Println("Credential dialog closed")
	})

	customDialog.Show()
}

func performCredentialSearch(searchTerm string, window fyne.Window, appState *state.AppState) {
	if appState == nil || appState.Username == "" {
		dialog.ShowError(errors.New("user is not logged in"), window)
		return
	}

	searchResults, message, err := crud.SearchCredentials(searchTerm, appState.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	resultList := widget.NewList(
		func() int {
			return len(searchResults)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Search Result")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(searchResults[id].LoginName + " - " + searchResults[id].Site)
		},
	)

	resultList.OnSelected = func(id widget.ListItemID) {
		showCredentialDialog(window, &searchResults[id], appState)
	}

	dialog.ShowCustom(message, "Close", resultList, window)
}
