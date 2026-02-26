package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SearchEntry は検索用のエントリウィジェット。
// テキストが入力されるとクリアボタンを表示し、OnChanged コールバックを呼び出す。
type SearchEntry struct {
	widget.Entry

	OnChanged func(query string)
	clearIcon *searchClearIcon
}

func NewSearchEntry(placeholder string) *SearchEntry {
	s := &SearchEntry{}
	s.PlaceHolder = placeholder
	s.clearIcon = newSearchClearIcon(s)
	s.ActionItem = s.clearIcon
	s.ExtendBaseWidget(s)
	s.Entry.OnChanged = func(text string) {
		s.clearIcon.setVisible(text != "")
		if s.OnChanged != nil {
			s.OnChanged(text)
		}
	}
	return s
}

// Clear はテキストをクリアする
func (s *SearchEntry) Clear() {
	s.SetText("")
}

// searchClearIcon is a custom tappable icon widget for the search clear button.
// It controls its own visibility by returning MinSize(0, 0) when hidden.
var (
	_ desktop.Cursorable = (*searchClearIcon)(nil)
	_ fyne.Tappable      = (*searchClearIcon)(nil)
	_ fyne.Widget        = (*searchClearIcon)(nil)
)

type searchClearIcon struct {
	widget.BaseWidget

	icon    *canvas.Image
	entry   *SearchEntry
	visible bool
}

func newSearchClearIcon(entry *SearchEntry) *searchClearIcon {
	icon := canvas.NewImageFromResource(theme.CancelIcon())
	c := &searchClearIcon{
		icon:  icon,
		entry: entry,
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *searchClearIcon) CreateRenderer() fyne.WidgetRenderer {
	return &searchClearIconRenderer{
		WidgetRenderer: widget.NewSimpleRenderer(c.icon),
		icon:           c.icon,
		clearIcon:      c,
	}
}

func (c *searchClearIcon) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

func (c *searchClearIcon) Tapped(*fyne.PointEvent) {
	c.entry.Clear()
}

func (c *searchClearIcon) setVisible(v bool) {
	c.visible = v
	c.Refresh()
}

var _ fyne.WidgetRenderer = (*searchClearIconRenderer)(nil)

type searchClearIconRenderer struct {
	fyne.WidgetRenderer
	icon      *canvas.Image
	clearIcon *searchClearIcon
}

func (r *searchClearIconRenderer) MinSize() fyne.Size {
	if !r.clearIcon.visible {
		return fyne.NewSize(0, 0)
	}
	iconSize := theme.IconInlineSize()
	return fyne.NewSquareSize(iconSize + theme.InnerPadding()*2)
}

func (r *searchClearIconRenderer) Layout(size fyne.Size) {
	if !r.clearIcon.visible {
		r.icon.Hide()
		return
	}
	r.icon.Show()
	iconSize := theme.IconInlineSize()
	r.icon.Resize(fyne.NewSquareSize(iconSize))
	r.icon.Move(fyne.NewPos((size.Width-iconSize)/2, (size.Height-iconSize)/2))
}

func (r *searchClearIconRenderer) Refresh() {
	if !r.clearIcon.visible {
		r.icon.Hide()
	} else {
		r.icon.Show()
		r.icon.Resource = theme.CancelIcon()
	}
	r.icon.Refresh()
}
