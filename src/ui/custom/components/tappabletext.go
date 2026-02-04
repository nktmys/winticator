package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// TappableText はタップ可能なスタイル付きテキストウィジェット
type TappableText struct {
	widget.BaseWidget
	text     *canvas.Text
	OnTapped func()
}

// NewTappableText は新しいTappableTextを作成する
func NewTappableText(text string, col color.Color, size float32) *TappableText {
	t := &TappableText{}
	t.text = canvas.NewText(text, col)
	t.text.TextSize = size
	t.text.TextStyle = fyne.TextStyle{Monospace: true}
	t.ExtendBaseWidget(t)
	return t
}

// SetText はテキストを設定する
func (t *TappableText) SetText(text string) {
	t.text.Text = text
	t.Refresh()
}

// SetColor はテキストの色を設定する
func (t *TappableText) SetColor(col color.Color) {
	t.text.Color = col
	t.Refresh()
}

// Tapped はタップイベントを処理する
func (t *TappableText) Tapped(_ *fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}

// TappedSecondary は右クリックイベントを処理する
func (t *TappableText) TappedSecondary(_ *fyne.PointEvent) {}

// Cursor はカーソルの形状を返す
func (t *TappableText) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// CreateRenderer はウィジェットのレンダラーを作成する
func (t *TappableText) CreateRenderer() fyne.WidgetRenderer {
	return &tappableTextRenderer{
		text: t,
	}
}

// MinSize はウィジェットの最小サイズを返す
func (t *TappableText) MinSize() fyne.Size {
	return t.text.MinSize()
}

// tappableTextRenderer はTappableTextのレンダラー
type tappableTextRenderer struct {
	text *TappableText
}

func (r *tappableTextRenderer) Layout(size fyne.Size) {
	r.text.text.Resize(size)
}

func (r *tappableTextRenderer) MinSize() fyne.Size {
	return r.text.text.MinSize()
}

func (r *tappableTextRenderer) Refresh() {
	r.text.text.Refresh()
}

func (r *tappableTextRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text.text}
}

func (r *tappableTextRenderer) Destroy() {}
