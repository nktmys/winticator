package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MenuItem はポップアップメニューの項目。
// fyne.MenuItem と異なり、項目ごとに文字色を指定できる。
type MenuItem struct {
	Label string
	// Color はラベルの文字色。nil の場合はテーマの前景色を使用する
	Color       color.Color
	IsSeparator bool
	Action      func()
}

// NewMenuItem はラベルと動作からメニュー項目を作成する
func NewMenuItem(label string, action func()) *MenuItem {
	return &MenuItem{Label: label, Action: action}
}

// NewColoredMenuItem はラベルを指定色で表示するメニュー項目を作成する
func NewColoredMenuItem(label string, col color.Color, action func()) *MenuItem {
	return &MenuItem{Label: label, Color: col, Action: action}
}

// NewMenuItemSeparator は区切り線となるメニュー項目を作成する
func NewMenuItemSeparator() *MenuItem {
	return &MenuItem{IsSeparator: true}
}

// PopUpMenu は項目ごとに文字色を指定できるポップアップメニュー。
// マウス操作に加え、上下キーでの項目移動、Enter/Space での実行、
// Escape でのクローズに対応する。
type PopUpMenu struct {
	*widget.PopUp

	canvas fyne.Canvas
	rows   []*menuRow
	active int // 選択中の項目インデックス（-1 は未選択）
}

// NewPopUpMenu は項目ごとに文字色を指定できるポップアップメニューを作成する。
// widget.NewPopUpMenu と同様に、生成時点でサイズが確定する。
func NewPopUpMenu(c fyne.Canvas, items ...*MenuItem) *PopUpMenu {
	m := &PopUpMenu{canvas: c, active: -1}

	objects := make([]fyne.CanvasObject, 0, len(items))
	for _, item := range items {
		if item.IsSeparator {
			objects = append(objects, widget.NewSeparator())
			continue
		}

		row := newMenuRow(m, len(m.rows), item.Label, item.Color, item.Action)
		m.rows = append(m.rows, row)
		objects = append(objects, row)
	}

	m.PopUp = widget.NewPopUp(container.NewVBox(objects...), c)
	m.Resize(m.MinSize())
	return m
}

// ShowAtPosition は指定位置にメニューを表示する
func (m *PopUpMenu) ShowAtPosition(pos fyne.Position) {
	m.PopUp.ShowAtPosition(pos)
	m.focusForKeys()
}

// ShowAtRelativePosition は指定オブジェクトからの相対位置にメニューを表示する
func (m *PopUpMenu) ShowAtRelativePosition(rel fyne.Position, to fyne.CanvasObject) {
	m.PopUp.ShowAtRelativePosition(rel, to)
	m.focusForKeys()
}

// Show はメニューを表示する
func (m *PopUpMenu) Show() {
	m.PopUp.Show()
	m.focusForKeys()
}

// focusForKeys はキーイベントを受け取るために先頭項目へフォーカスを移す。
// 項目の選択状態とは独立しているため、表示直後はどの項目も選択されない。
func (m *PopUpMenu) focusForKeys() {
	if len(m.rows) > 0 {
		m.canvas.Focus(m.rows[0])
	}
}

// setActive は選択中の項目を変更する（index が範囲外なら未選択にする）
func (m *PopUpMenu) setActive(index int) {
	if index < 0 || index >= len(m.rows) {
		index = -1
	}
	if m.active == index {
		return
	}

	previous := m.active
	m.active = index
	if previous >= 0 && previous < len(m.rows) {
		m.rows[previous].Refresh()
	}
	if index >= 0 {
		m.rows[index].Refresh()
	}
}

// moveActive は選択を指定方向に移動する（末尾と先頭は循環する）
func (m *PopUpMenu) moveActive(direction int) {
	if len(m.rows) == 0 {
		return
	}

	next := m.active + direction
	switch {
	case m.active < 0 && direction < 0:
		next = len(m.rows) - 1
	case next < 0:
		next = len(m.rows) - 1
	case next >= len(m.rows):
		next = 0
	}
	m.setActive(next)
}

// activate は選択中の項目を実行してメニューを閉じる
func (m *PopUpMenu) activate() {
	if m.active < 0 || m.active >= len(m.rows) {
		return
	}
	m.rows[m.active].trigger()
}

// menuRow はポップアップメニューの1項目を描画するウィジェット
type menuRow struct {
	widget.BaseWidget

	menu   *PopUpMenu
	index  int
	label  string
	color  color.Color
	action func()
}

var (
	_ fyne.Widget       = (*menuRow)(nil)
	_ fyne.Tappable     = (*menuRow)(nil)
	_ fyne.Focusable    = (*menuRow)(nil)
	_ desktop.Hoverable = (*menuRow)(nil)
)

func newMenuRow(menu *PopUpMenu, index int, label string, col color.Color, action func()) *menuRow {
	r := &menuRow{menu: menu, index: index, label: label, color: col, action: action}
	r.ExtendBaseWidget(r)
	return r
}

// CreateRenderer はウィジェットのレンダラーを作成する
func (r *menuRow) CreateRenderer() fyne.WidgetRenderer {
	th := r.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	background := canvas.NewRectangle(th.Color(theme.ColorNameHover, v))
	background.CornerRadius = th.Size(theme.SizeNameMenuRadius)
	background.Hide()

	text := canvas.NewText(r.label, r.textColor())
	text.TextSize = th.Size(theme.SizeNameText)

	return &menuRowRenderer{row: r, background: background, text: text}
}

func (r *menuRow) Tapped(*fyne.PointEvent) {
	r.trigger()
}

func (r *menuRow) MouseIn(*desktop.MouseEvent) {
	r.menu.setActive(r.index)
}

func (r *menuRow) MouseMoved(*desktop.MouseEvent) {}

func (r *menuRow) MouseOut() {
	if r.menu.active == r.index {
		r.menu.setActive(-1)
	}
}

// FocusGained はフォーカス取得時に呼ばれる。表示はしない（選択状態は menu が保持する）
func (r *menuRow) FocusGained() {}

// FocusLost はフォーカス喪失時に呼ばれる
func (r *menuRow) FocusLost() {}

func (r *menuRow) TypedRune(rune) {}

// TypedKey はメニューのキーボード操作を処理する
func (r *menuRow) TypedKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyDown:
		r.menu.moveActive(+1)
	case fyne.KeyUp:
		r.menu.moveActive(-1)
	case fyne.KeyReturn, fyne.KeyEnter, fyne.KeySpace:
		r.menu.activate()
	case fyne.KeyEscape:
		r.menu.Hide()
	default:
		// その他のキーはメニューでは扱わない
	}
}

// trigger はメニューを閉じてから項目の動作を実行する
func (r *menuRow) trigger() {
	r.menu.Hide()
	if r.action != nil {
		r.action()
	}
}

// isActive は自身が選択中かどうかを返す
func (r *menuRow) isActive() bool {
	return r.menu.active == r.index
}

// textColor はラベルの文字色を返す
func (r *menuRow) textColor() color.Color {
	if r.color != nil {
		return r.color
	}
	return r.Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant())
}

// menuRowRenderer は menuRow のレンダラー
type menuRowRenderer struct {
	row        *menuRow
	background *canvas.Rectangle
	text       *canvas.Text
}

func (r *menuRowRenderer) Layout(size fyne.Size) {
	th := r.row.Theme()
	pad := th.Size(theme.SizeNamePadding)
	innerPad := th.Size(theme.SizeNameInnerPadding)

	r.background.Resize(size.Subtract(fyne.NewSquareSize(pad)))
	r.background.Move(fyne.NewPos(pad/2, pad/2))

	r.text.Resize(fyne.NewSize(size.Width-innerPad*2, r.text.MinSize().Height))
	r.text.Move(fyne.NewPos(innerPad, innerPad))
}

func (r *menuRowRenderer) MinSize() fyne.Size {
	innerPad := r.row.Theme().Size(theme.SizeNameInnerPadding)
	return r.text.MinSize().AddWidthHeight(innerPad*2, innerPad*2)
}

func (r *menuRowRenderer) Refresh() {
	th := r.row.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.background.FillColor = th.Color(theme.ColorNameHover, v)
	r.background.CornerRadius = th.Size(theme.SizeNameMenuRadius)
	if r.row.isActive() {
		r.background.Show()
	} else {
		r.background.Hide()
	}
	r.background.Refresh()

	r.text.Text = r.row.label
	r.text.Color = r.row.textColor()
	r.text.TextSize = th.Size(theme.SizeNameText)
	r.text.Refresh()
}

func (r *menuRowRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.text}
}

func (r *menuRowRenderer) Destroy() {}
