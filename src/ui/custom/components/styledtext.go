package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// StyledText はスタイル付きテキストウィジェット
type StyledText struct {
	widget.BaseWidget
	text *canvas.Text
}

// NewStyledText は新しいStyledTextを作成する
func NewStyledText(text string, col color.Color, size float32) *StyledText {
	t := &StyledText{}
	t.text = canvas.NewText(text, col)
	t.text.TextSize = size
	t.text.TextStyle = fyne.TextStyle{Monospace: true}
	t.ExtendBaseWidget(t)
	return t
}

// SetText はテキストを設定する
func (t *StyledText) SetText(text string) {
	t.text.Text = text
	t.Refresh()
}

// SetColor はテキストの色を設定する
func (t *StyledText) SetColor(col color.Color) {
	t.text.Color = col
	t.Refresh()
}

// CreateRenderer はウィジェットのレンダラーを作成する
func (t *StyledText) CreateRenderer() fyne.WidgetRenderer {
	return &styledTextRenderer{
		text: t,
	}
}

// MinSize はウィジェットの最小サイズを返す
func (t *StyledText) MinSize() fyne.Size {
	return t.text.MinSize()
}

// styledTextRenderer はStyledTextのレンダラー
type styledTextRenderer struct {
	text *StyledText
}

func (r *styledTextRenderer) Layout(size fyne.Size) {
	r.text.text.Resize(size)
}

func (r *styledTextRenderer) MinSize() fyne.Size {
	return r.text.text.MinSize()
}

func (r *styledTextRenderer) Refresh() {
	r.text.text.Refresh()
}

func (r *styledTextRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text.text}
}

func (r *styledTextRenderer) Destroy() {}
