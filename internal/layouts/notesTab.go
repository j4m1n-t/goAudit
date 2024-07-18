package layouts

import (
	// Standard Library
	"log"

	// Fyne Imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// External Imports

	// Internal Imports
	crud "github.com/j4m1n-t/goAudit/internal/CRUD"
	myAuth "github.com/j4m1n-t/goAudit/internal/authentication"
)

var notesList *widget.List
var ldapConn myAuth.LDAPConnection

func CreateNotesTabContent(window fyne.Window) fyne.CanvasObject {
	// Search functionality
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search notes...")
	searchButton := widget.NewButton("Search", func() {
		performSearch(searchEntry.Text, window)
	})

	user, err := crud.GetOrCreateUser(ldapConn.Username)
	if err != nil {
		log.Printf("Error getting or creating user: %v", err)
		dialog.ShowError(err, window)
		return nil
	}
	log.Printf("User details: ID=%d, Username=%s, UserID=%d", user.ID, user.Username, user.UserID)

	// List of notes
	notesList = widget.NewList(
		func() int {
			notes, err := crud.GetNotes()
			if err != nil {
				//log.Printf("Error getting notes: %v", err)
				return 0
			}
			return len(notes)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Note Title")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			notes, err := crud.GetNotes()
			if err != nil {
				log.Printf("Error getting notes: %v", err)
				return
			}
			if id < len(notes) {
				item.(*widget.Label).SetText(notes[id].Title)
			}
		},
	)

	// Create new note button
	newNoteButton := widget.NewButton("New Note", func() {
		showNoteDialog(window, nil)
	})

	// Edit note when list item is selected
	notesList.OnSelected = func(id widget.ListItemID) {
		notes, err := crud.GetNotes()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if id < len(notes) {
			showNoteDialog(window, &notes[id])
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Notes"),
			container.NewBorder(nil, nil, nil, searchButton, searchEntry),
			newNoteButton,
		),
		nil, nil, nil,
		notesList,
	)
}

func showNoteDialog(window fyne.Window, note *crud.Notes) {
	var titleEntry *widget.Entry
	var contentEntry *widget.Entry
	var customDialog dialog.Dialog

	user, err := crud.GetOrCreateUser(ldapConn.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	titleEntry = widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter title")

	contentEntry = widget.NewMultiLineEntry()
	contentEntry.SetPlaceHolder("Enter content")
	contentEntry.Wrapping = fyne.TextWrapWord

	if note != nil {
		titleEntry.SetText(note.Title)
		contentEntry.SetText(note.Content)
	}

	saveButton := widget.NewButton("Save", func() {
		if note == nil {
			// Create new note
			newNote, err := crud.CreateNote(titleEntry.Text, contentEntry.Text, user) // Pass the entire user object
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Created new note with ID: %d", newNote.ID)
		} else {
			// Update existing note
			note.Title = titleEntry.Text
			note.Content = contentEntry.Text
			updatedNote, err := crud.UpdateNote(*note)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Updated note with ID: %d", updatedNote.ID)
		}
		notesList.Refresh()
		customDialog.Hide()
	})
	buttons := container.NewHBox(saveButton)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Title"),
			titleEntry,
		),
		nil,
		nil,
		nil,
		container.NewVBox(
			widget.NewLabel("Content"),
			contentEntry,
		),
	)

	// Wrap content in a padded container
	paddedContent := container.NewPadded(content)

	// Create a container with buttons at the bottom
	mainContainer := container.NewBorder(nil, buttons, nil, nil, paddedContent)

	customDialog = dialog.NewCustom("Note", "Cancel", mainContainer, window)
	customDialog.Resize(fyne.NewSize(600, 500))

	customDialog.SetOnClosed(func() {
		notesList.Refresh()
		notesList.UnselectAll()
		log.Println("Note dialog closed")
	})

	customDialog.Show()
}

func performSearch(searchTerm string, window fyne.Window) {
	searchResults, err := crud.SearchNotes(searchTerm)
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
		showNoteDialog(window, &searchResults[id])
	}

	dialog.ShowCustom("Search Results", "Close", resultList, window)
}
