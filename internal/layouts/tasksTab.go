package layouts

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	crud "github.com/j4m1n-t/goAudit/internal/CRUD"
	state "github.com/j4m1n-t/goAudit/internal/status"
)

var tasksList *widget.List

func CreatePlaceholderTaskTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view your tasks."),
	)
}

func CreateTasksTabContent(window fyne.Window) fyne.CanvasObject {
	newTaskButton := widget.NewButton("New Task", func() {
		showTaskDialog(window, nil)
	})

	tasksList = widget.NewList(
		func() int {
			return len(state.GlobalState.Tasks)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Task Title"),
				widget.NewLabel("Due Date"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(state.GlobalState.Tasks) {
				task := state.GlobalState.Tasks[id]
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(task.Title)
				item.(*fyne.Container).Objects[2].(*widget.Label).SetText(task.DueDate.Format("2006-01-02"))
			}
		},
	)

	tasksList.OnSelected = func(id widget.ListItemID) {
		if id < len(state.GlobalState.Tasks) {
			showTaskDialog(window, &state.GlobalState.Tasks[id])
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Tasks"),
			newTaskButton,
		),
		nil, nil, nil,
		tasksList,
	)
}

func showTaskDialog(window fyne.Window, task *crud.Tasks) {
	var titleEntry *widget.Entry
	var descriptionEntry *widget.Entry
	var statusEntry *widget.Entry
	var priorityEntry *widget.Entry
	var dueDateEntry *widget.Entry
	var completedCheck *widget.Check

	titleEntry = widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter title")

	descriptionEntry = widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Enter description")
	descriptionEntry.Wrapping = fyne.TextWrapWord

	statusEntry = widget.NewEntry()
	statusEntry.SetPlaceHolder("Enter status")

	priorityEntry = widget.NewEntry()
	priorityEntry.SetPlaceHolder("Enter priority (1-5)")

	dueDateEntry = widget.NewEntry()
	dueDateEntry.SetPlaceHolder("Enter due date (YYYY-MM-DD)")

	completedCheck = widget.NewCheck("Completed", nil)

	if task != nil {
		titleEntry.SetText(task.Title)
		descriptionEntry.SetText(task.Description)
		statusEntry.SetText(task.Status)
		priorityEntry.SetText(fmt.Sprintf("%d", task.Priority))
		dueDateEntry.SetText(task.DueDate.Format("2006-01-02"))
		completedCheck.SetChecked(task.Completed)
	}

	saveButton := widget.NewButton("Save", func() {
		priority := parseInt(priorityEntry.Text)
		dueDate, _ := time.Parse("2006-01-02", dueDateEntry.Text)

		if task == nil {
			newTask := crud.Tasks{
				Title:       titleEntry.Text,
				Description: descriptionEntry.Text,
				Status:      statusEntry.Text,
				Priority:    priority,
				DueDate:     dueDate,
				Completed:   completedCheck.Checked,
				Username:    state.GlobalState.Username,
			}
			_, err := crud.CreateTask(newTask)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		} else {
			task.Title = titleEntry.Text
			task.Description = descriptionEntry.Text
			task.Status = statusEntry.Text
			task.Priority = priority
			task.DueDate = dueDate
			task.Completed = completedCheck.Checked
			_, err := crud.UpdateTask(*task)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
		}

		refreshTasks(window)
		dialog.ShowInformation("Success", "Task saved successfully", window)
	})

	var buttons fyne.CanvasObject
	if task != nil {
		deleteButton := widget.NewButton("Delete", func() {
			dialog.ShowConfirm("Confirm Delete", "Are you sure you want to delete this task?", func(confirm bool) {
				if confirm {
					err := crud.DeleteTask(task.ID, state.GlobalState.Username)
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					refreshTasks(window)
					dialog.ShowInformation("Success", "Task deleted successfully", window)
				}
			}, window)
		})
		buttons = container.NewHBox(saveButton, deleteButton)
	} else {
		buttons = container.NewHBox(saveButton)
	}

	content := container.NewVBox(
		widget.NewLabel("Title"),
		titleEntry,
		widget.NewLabel("Description"),
		descriptionEntry,
		widget.NewLabel("Status"),
		statusEntry,
		widget.NewLabel("Priority"),
		priorityEntry,
		widget.NewLabel("Due Date"),
		dueDateEntry,
		completedCheck,
		buttons,
	)

	dialog.ShowCustom("Task Details", "Close", content, window)
}

func refreshTasks(window fyne.Window) {
	tasks, _, err := crud.GetTasks(state.GlobalState.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	state.GlobalState.Tasks = tasks
	tasksList.Refresh()
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
