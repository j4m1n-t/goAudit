package layouts

import (
	// Standard Library
	"fmt"

	// Fyne Imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// Internal Imports
	crud "github.com/j4m1n-t/goAudit/internal/databases"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

func CreatePlaceholderAdminTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view administrative functions."),
	)
}

func CreateAdminTabContent(window fyne.Window) fyne.CanvasObject {
	// LDAP Setup
	ldapSetupButton := widget.NewButton("LDAP Setup", func() {
		showLDAPSetupDialog(window)
	})

	// Postgres Setup
	postgresSetupButton := widget.NewButton("Postgres Setup", func() {
		showPostgresSetupDialog(window)
	})

	// Delete functions
	deleteNoteButton := widget.NewButton("Delete Note", func() {
		showDeleteDialog(window, "Note", deleteNote)
	})

	deleteTaskButton := widget.NewButton("Delete Task", func() {
		showDeleteDialog(window, "Task", deleteTask)
	})

	deleteUserButton := widget.NewButton("Delete User", func() {
		showDeleteDialog(window, "User", deleteUser)
	})

	deleteCRMButton := widget.NewButton("Delete CRM Entry", func() {
		showDeleteDialog(window, "CRM Entry", deleteCRM)
	})

	deleteAuditButton := widget.NewButton("Delete Audit", func() {
		showDeleteDialog(window, "Audit", deleteAudit)
	})

	return container.NewVBox(
		widget.NewLabel("Administrative Functions"),
		ldapSetupButton,
		postgresSetupButton,
		deleteNoteButton,
		deleteTaskButton,
		deleteUserButton,
		deleteCRMButton,
		deleteAuditButton,
	)
}

func showLDAPSetupDialog(window fyne.Window) {
	// Implement LDAP setup dialog
	// This should include fields for LDAP server, port, base DN, etc.
}

func showPostgresSetupDialog(window fyne.Window) {
	// Implement Postgres setup dialog
	// This should include fields for host, port, database name, user, password, etc.
}

func showDeleteDialog(window fyne.Window, itemType string, deleteFunc func(int) error) {
	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Enter ID to delete")

	content := container.NewVBox(
		widget.NewLabel("Delete "+itemType),
		idEntry,
	)

	dialog.ShowCustomConfirm("Confirm Delete", "Delete", "Cancel", content, func(confirm bool) {
		if confirm {
			id := parseInt(idEntry.Text)
			if id == 0 {
				dialog.ShowError(fmt.Errorf("Invalid ID"), window)
				return
			}
			err := deleteFunc(id)
			if err != nil {
				dialog.ShowError(err, window)
			} else {
				dialog.ShowInformation("Success", itemType+" deleted successfully", window)
			}
		}
	}, window)
}

var dw *crud.DatabaseWrapper

func deleteNote(id int) error {
	return dw.DeleteNote(id)
}

func deleteTask(id int) error {
	return dw.DeleteTask(id, state.GlobalState.Username)
}

func deleteUser(id int) error {
	// Implement user deletion logic
	return fmt.Errorf("user deletion not implemented")
}

func deleteCRM(id int) error {
	return dw.DeleteCRMEntry(id, state.GlobalState.Username)
}

func deleteAudit(id int) error {
	return dw.DeleteAudit(id, state.GlobalState.Username)
}
