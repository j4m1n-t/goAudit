package layouts

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	crud "github.com/j4m1n-t/goAudit/internal/CRUD"
)

var tasksList *widget.List

func CreatePlaceholderTaskTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view your tasks."),
	)
}
func CreateTasksTabContent(window fyne.Window) fyne.CanvasObject {
	// Search functionality
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search tasks...")
	searchButton := widget.NewButton("Search", func() {
		performTaskSearch(searchEntry.Text, window)
	})

	// List of tasks
	tasksList = widget.NewList(
		func() int {
			tasks, err := crud.GetTasks()
			if err != nil {
				return 0
			}
			return len(tasks)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Task Title")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			tasks, err := crud.GetTasks()
			if err != nil {
				log.Printf("Error getting tasks: %v", err)
				return
			}
			if id < len(tasks) {
				item.(*widget.Label).SetText(tasks[id].Title)
			}
		},
	)

	// Create new task button
	newTaskButton := widget.NewButton("New Task", func() {
		showTaskDialog(window, nil)
	})

	// Edit task when list item is selected
	tasksList.OnSelected = func(id widget.ListItemID) {
		tasks, err := crud.GetTasks()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if id < len(tasks) {
			showTaskDialog(window, &tasks[id])
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Tasks"),
			container.NewBorder(nil, nil, nil, searchButton, searchEntry),
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
	var customDialog dialog.Dialog

	user, err := crud.GetOrCreateUser(ldapConn.Username)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

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
		priorityEntry.SetText(strconv.Itoa(task.Priority))
		dueDateEntry.SetText(task.DueDate.Format("2006-01-02"))
		completedCheck.SetChecked(task.Completed)
	}

	saveButton := widget.NewButton("Save", func() {
		dueDate, err := time.Parse("2006-01-02", dueDateEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid date format: %v", err), window)
			return
		}
		priority, err := strconv.Atoi(priorityEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid priority: must be a number"), window)
			return
		}

		newTask := crud.Tasks{
			Title:       titleEntry.Text,
			Description: descriptionEntry.Text,
			Status:      statusEntry.Text,
			Priority:    priority,
			DueDate:     dueDate,
			Completed:   completedCheck.Checked,
			UserID:      user.UserID,
			Username:    user.Username,
		}

		if task == nil {
			// Create new task
			createdTask, err := crud.CreateTask(newTask)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Created new task with ID: %d", createdTask.ID)
		} else {
			// Update existing task
			newTask.ID = task.ID
			updatedTask, err := crud.UpdateTask(newTask)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			log.Printf("Updated task with ID: %d", updatedTask.ID)
		}
		tasksList.Refresh()
		customDialog.Hide()
	})
	buttons := container.NewHBox(saveButton)

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
	)

	// Wrap content in a padded container
	paddedContent := container.NewPadded(content)

	// Create a container with buttons at the bottom
	mainContainer := container.NewBorder(nil, buttons, nil, nil, paddedContent)

	customDialog = dialog.NewCustom("Task", "Cancel", mainContainer, window)
	customDialog.Resize(fyne.NewSize(600, 500))

	customDialog.SetOnClosed(func() {
		tasksList.Refresh()
		tasksList.UnselectAll()
		log.Println("Task dialog closed")
	})

	customDialog.Show()
}

func performTaskSearch(searchTerm string, window fyne.Window) {
	searchResults, err := crud.SearchTasks(searchTerm)
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
		showTaskDialog(window, &searchResults[id])
	}

	dialog.ShowCustom("Search Results", "Close", resultList, window)
}
