package layouts

import (
	// Standard Library
	"errors"
	"log"

	// Fyne Imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	// Internal Imports

	interfaces "github.com/j4m1n-t/goAudit/internal/interfaces"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

var notesList *widget.List
var ldapConn interfaces.LDAPConnection

func CreatePlaceholderNotesTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view your notes."),
	)
}

func CreateNotesTabContent(window fyne.Window) fyne.CanvasObject {
	var appState *state.AppState
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search notes...")
	searchButton := widget.NewButton("Search", func() {
		performSearch(searchEntry.Text, window, appState)
	})

	// Fetch notes
	appState.FetchNotes()

	messageLabel := widget.NewLabel(appState.Message)

	notesList = widget.NewList(
		func() int {
			return len(appState.Notes)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Title"),
				widget.NewLabel("(Open)"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(appState.Notes) {
				note := appState.Notes[id]
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(note.Title)
				openLabel := item.(*fyne.Container).Objects[2].(*widget.Label)
				if note.Open {
					openLabel.Show()
				} else {
					openLabel.Hide()
				}
			}
		},
	)

	newNoteButton := widget.NewButton("New Note", func() {
		showNoteDialog(window, nil, appState)
	})

	notesList.OnSelected = func(id widget.ListItemID) {
		if id < len(appState.Notes) {
			showNoteDialog(window, &appState.Notes[id], appState)
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Notes"),
			container.NewBorder(nil, nil, nil, searchButton, searchEntry),
			messageLabel,
			newNoteButton,
		),
		nil, nil, nil,
		notesList,
	)
}

func showNoteDialog(window fyne.Window, note *interfaces.Note, appState *state.AppState) {
	var titleEntry *widget.Entry
	var contentEntry *widget.Entry
	var customDialog dialog.Dialog
	var openCheck *widget.Check

	titleEntry = widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter title")

	contentEntry = widget.NewMultiLineEntry()
	contentEntry.SetPlaceHolder("Enter content")
	contentEntry.Wrapping = fyne.TextWrapWord

	openCheck = widget.NewCheck("Open Note to All", func(checked bool) {
		if note != nil {
			note.Open = checked
		}
		log.Printf("Open status changed to: %v", checked)
	})

	if note != nil {
		titleEntry.SetText(note.Title)
		contentEntry.SetText(note.Content)
		openCheck.SetChecked(note.Open)
	}

	saveButton := widget.NewButton("Save", func() {
		if appState.Username == "" {
			dialog.ShowError(errors.New("user is not logged in"), window)
			return
		}
		log.Printf("username changed to: %v", appState.Username)
		log.Printf("Note Title changed to: %v", titleEntry.Text)
		log.Printf("Note Content changed to: %v", contentEntry.Text)
		if note == nil {
			newNote, err := dw.CreateNote(titleEntry.Text, contentEntry.Text, appState.Username, openCheck.Checked)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Created new note with ID: %d for user: %s", newNote.ID, appState.Username)
		} else {
			note.Title = titleEntry.Text
			note.Content = contentEntry.Text
			note.Open = openCheck.Checked
			updatedNote, err := dw.UpdateNote(*note)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Updated note with ID: %d", updatedNote.ID)
		}
		appState.FetchNotes()
		notesList.Refresh()
		customDialog.Hide()
	})

	deleteButton := widget.NewButton("Delete", func() {
		if note != nil {
			confirmDialog := dialog.NewConfirm("Confirm Delete", "Are you sure you want to delete this note?", func(confirm bool) {
				if confirm {
					err := dw.DeleteNote(note.ID)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					log.Printf("Deleted note with ID: %d by username: %s", note.ID, appState.Username)
					appState.FetchNotes()
					notesList.Refresh()
					customDialog.Hide()
				}
			}, window)
			confirmDialog.Show()
		}
	})

	var buttons fyne.CanvasObject
	if note != nil {
		buttons = container.NewHBox(saveButton, deleteButton)
	} else {
		buttons = container.NewHBox(saveButton)
	}

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Title"),
			titleEntry,
			openCheck,
		),
		nil,
		nil,
		nil,
		container.NewVBox(
			widget.NewLabel("Content"),
			contentEntry,
		),
	)

	paddedContent := container.NewPadded(content)
	mainContainer := container.NewBorder(nil, buttons, nil, nil, paddedContent)

	customDialog = dialog.NewCustom("Note", "Cancel", mainContainer, window)
	customDialog.Resize(fyne.NewSize(600, 500))

	customDialog.SetOnClosed(func() {
		appState.FetchNotes()
		notesList.Refresh()
		notesList.UnselectAll()
		log.Println("Note dialog closed")
	})

	customDialog.Show()
}

func performSearch(searchTerm string, window fyne.Window, appState *state.AppState) {
	if appState == nil || appState.Username == "" {
		dialog.ShowError(errors.New("user is not logged in"), window)
		return
	}

	searchResults, message, err := dw.SearchNotes(searchTerm, appState.Username)
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
			item.(*widget.Label).SetText(searchResults[id].Title)
		},
	)

	resultList.OnSelected = func(id widget.ListItemID) {
		showNoteDialog(window, &searchResults[id], appState)
	}

	dialog.ShowCustom(message, "Close", resultList, window)
}
