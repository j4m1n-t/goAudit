package layouts

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	crud "github.com/j4m1n-t/goAudit/internal/CRUD"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

var auditsList *widget.List

func CreatePlaceholderAuditsTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view audits."),
	)
}

func CreateAuditsTabContent(window fyne.Window) fyne.CanvasObject {
	newAuditButton := widget.NewButton("New Audit", func() {
		showAuditDialog(window, nil)
	})

	auditsList = widget.NewList(
		func() int {
			return len(state.GlobalState.Audits)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Audit Action"),
				widget.NewLabel("Audit Type"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(state.GlobalState.Audits) {
				audit := state.GlobalState.Audits[id]
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(audit.Action)
				item.(*fyne.Container).Objects[2].(*widget.Label).SetText(audit.AuditType)
			}
		},
	)

	auditsList.OnSelected = func(id widget.ListItemID) {
		if id < len(state.GlobalState.Audits) {
			showAuditDialog(window, &state.GlobalState.Audits[id])
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Audits"),
			newAuditButton,
		),
		nil, nil, nil,
		auditsList,
	)
}

func showAuditDialog(window fyne.Window, audit *crud.Audits) {
	var actionEntry, auditTypeEntry, auditAreaEntry, notesEntry, assignedUserEntry, firmEntry *widget.Entry
	var completedCheck *widget.Check

	actionEntry = widget.NewEntry()
	actionEntry.SetPlaceHolder("Enter action")

	auditTypeEntry = widget.NewEntry()
	auditTypeEntry.SetPlaceHolder("Enter audit type")

	auditAreaEntry = widget.NewEntry()
	auditAreaEntry.SetPlaceHolder("Enter audit area")

	notesEntry = widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Enter notes")
	notesEntry.Wrapping = fyne.TextWrapWord

	assignedUserEntry = widget.NewEntry()
	assignedUserEntry.SetPlaceHolder("Enter assigned user")

	firmEntry = widget.NewEntry()
	firmEntry.SetPlaceHolder("Enter firm")

	completedCheck = widget.NewCheck("Completed", nil)

	if audit != nil {
		actionEntry.SetText(audit.Action)
		auditTypeEntry.SetText(audit.AuditType)
		auditAreaEntry.SetText(audit.AuditArea)
		notesEntry.SetText(audit.Notes)
		assignedUserEntry.SetText(audit.AssignedUser)
		firmEntry.SetText(audit.Firm)
		completedCheck.SetChecked(audit.Completed)
	}

	saveButton := widget.NewButton("Save", func() {
		if audit == nil {
			newAudit := crud.Audits{
				Action:       actionEntry.Text,
				AuditType:    auditTypeEntry.Text,
				AuditArea:    auditAreaEntry.Text,
				Notes:        notesEntry.Text,
				AssignedUser: assignedUserEntry.Text,
				Firm:         firmEntry.Text,
				Completed:    completedCheck.Checked,
				Username:     state.GlobalState.Username,
			}
			_, err := crud.CreateAudit(newAudit)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		} else {
			audit.Action = actionEntry.Text
			audit.AuditType = auditTypeEntry.Text
			audit.AuditArea = auditAreaEntry.Text
			audit.Notes = notesEntry.Text
			audit.AssignedUser = assignedUserEntry.Text
			audit.Firm = firmEntry.Text
			audit.Completed = completedCheck.Checked
			if audit.Completed {
				audit.CompletedAt = time.Now()
			}
			_, err := crud.UpdateAudit(*audit)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		}

		refreshAudits(window)
		dialog.ShowInformation("Success", "Audit saved successfully", window)
	})

	var buttons fyne.CanvasObject
	if audit != nil {
		deleteButton := widget.NewButton("Delete", func() {
			dialog.ShowConfirm("Confirm Delete", "Are you sure you want to delete this audit?", func(confirm bool) {
				if confirm {
					err := crud.DeleteAudit(audit.ID, state.GlobalState.Username)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					refreshAudits(window)
					dialog.ShowInformation("Success", "Audit deleted successfully", window)
				}
			}, window)
		})
		buttons = container.NewHBox(saveButton, deleteButton)
	} else {
		buttons = container.NewHBox(saveButton)
	}

	content := container.NewVBox(
		widget.NewLabel("Action"),
		actionEntry,
		widget.NewLabel("Audit Type"),
		auditTypeEntry,
		widget.NewLabel("Audit Area"),
		auditAreaEntry,
		widget.NewLabel("Notes"),
		notesEntry,
		widget.NewLabel("Assigned User"),
		assignedUserEntry,
		widget.NewLabel("Firm"),
		firmEntry,
		completedCheck,
		buttons,
	)

	dialog.ShowCustom("Audit Details", "Close", content, window)
}

func refreshAudits(window fyne.Window) {
	audits, _, err := crud.GetAudits(state.GlobalState.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	state.GlobalState.Audits = audits
	auditsList.Refresh()
}
