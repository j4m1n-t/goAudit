package layouts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	crud "github.com/j4m1n-t/goAudit/internal/databases"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

var crmList *widget.List

func CreatePlaceholderCRMTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view CRM entries."),
	)
}

func CreateCRMTabContent(window fyne.Window) fyne.CanvasObject {
	newCRMButton := widget.NewButton("New CRM Entry", func() {
		showCRMDialog(window, nil)
	})

	crmList = widget.NewList(
		func() int {
			return len(state.GlobalState.CRMEntries)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("Name"),
				widget.NewLabel("Company"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(state.GlobalState.CRMEntries) {
				crm := state.GlobalState.CRMEntries[id]
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(crm.Name)
				item.(*fyne.Container).Objects[2].(*widget.Label).SetText(crm.Company)
			}
		},
	)

	crmList.OnSelected = func(id widget.ListItemID) {
		if id < len(state.GlobalState.CRMEntries) {
			showCRMDialog(window, &state.GlobalState.CRMEntries[id])
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("CRM Entries"),
			newCRMButton,
		),
		nil, nil, nil,
		crmList,
	)
}

func showCRMDialog(window fyne.Window, crm *crud.CRM) {
	var nameEntry, emailEntry, phoneEntry, companyEntry *widget.Entry
	var notesEntry *widget.Entry
	var openCheck *widget.Check

	nameEntry = widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter name")

	emailEntry = widget.NewEntry()
	emailEntry.SetPlaceHolder("Enter email")

	phoneEntry = widget.NewEntry()
	phoneEntry.SetPlaceHolder("Enter phone")

	companyEntry = widget.NewEntry()
	companyEntry.SetPlaceHolder("Enter company")

	notesEntry = widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Enter notes")
	notesEntry.Wrapping = fyne.TextWrapWord

	openCheck = widget.NewCheck("Open to All", nil)

	if crm != nil {
		nameEntry.SetText(crm.Name)
		emailEntry.SetText(crm.Email)
		phoneEntry.SetText(crm.Phone)
		companyEntry.SetText(crm.Company)
		notesEntry.SetText(crm.Notes[0]) // Assuming we're using only the first note for simplicity
		openCheck.SetChecked(crm.Open)
	}

	saveButton := widget.NewButton("Save", func() {
		if crm == nil {
			newCRM := crud.CRM{
				Name:     nameEntry.Text,
				Email:    emailEntry.Text,
				Phone:    phoneEntry.Text,
				Company:  companyEntry.Text,
				Notes:    []string{notesEntry.Text},
				Open:     openCheck.Checked,
				Username: state.GlobalState.Username,
			}
			_, err := crud.CreateCRMEntry(newCRM)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		} else {
			crm.Name = nameEntry.Text
			crm.Email = emailEntry.Text
			crm.Phone = phoneEntry.Text
			crm.Company = companyEntry.Text
			crm.Notes = []string{notesEntry.Text}
			crm.Open = openCheck.Checked
			_, err := crud.UpdateCRMEntry(*crm)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		}

		refreshCRM(window)
		dialog.ShowInformation("Success", "CRM entry saved successfully", window)
	})

	var buttons fyne.CanvasObject
	if crm != nil {
		deleteButton := widget.NewButton("Delete", func() {
			dialog.ShowConfirm("Confirm Delete", "Are you sure you want to delete this CRM entry?", func(confirm bool) {
				if confirm {
					err := crud.DeleteCRMEntry(crm.ID, state.GlobalState.Username)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					refreshCRM(window)
					dialog.ShowInformation("Success", "CRM entry deleted successfully", window)
				}
			}, window)
		})
		buttons = container.NewHBox(saveButton, deleteButton)
	} else {
		buttons = container.NewHBox(saveButton)
	}

	content := container.NewVBox(
		widget.NewLabel("Name"),
		nameEntry,
		widget.NewLabel("Email"),
		emailEntry,
		widget.NewLabel("Phone"),
		phoneEntry,
		widget.NewLabel("Company"),
		companyEntry,
		widget.NewLabel("Notes"),
		notesEntry,
		openCheck,
		buttons,
	)

	dialog.ShowCustom("CRM Entry Details", "Close", content, window)
}

func refreshCRM(window fyne.Window) {
	crmEntries, _, err := crud.GetCRMEntries(state.GlobalState.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	state.GlobalState.CRMEntries = crmEntries
	crmList.Refresh()
}
