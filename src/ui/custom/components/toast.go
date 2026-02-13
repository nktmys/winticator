package components

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowToast はウィンドウ下部に一時的なトースト通知を表示する
func ShowToast(win fyne.Window, message string) {
	label := widget.NewLabel(message)

	content := container.NewPadded(label)
	popup := widget.NewPopUp(content, win.Canvas())

	// ウィンドウ下部中央に配置
	canvasSize := win.Canvas().Size()
	popupSize := popup.MinSize()
	popup.ShowAtPosition(fyne.NewPos(
		(canvasSize.Width-popupSize.Width)/2,
		canvasSize.Height-popupSize.Height-20,
	))

	// 2秒後に自動的に非表示
	go func() {
		time.Sleep(2 * time.Second)
		fyne.Do(func() {
			popup.Hide()
		})
	}()
}
