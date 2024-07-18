package layouts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreatePlaceholderCredentialsTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Please log in to view your stored credentials."),
	)
}
