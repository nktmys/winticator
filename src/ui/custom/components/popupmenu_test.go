package components

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRed = color.RGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff}

// newTestPopUpMenu は「編集 / 区切り線 / 削除（赤）」のメニューを作成する
func newTestPopUpMenu(t *testing.T, edited, deleted *bool) (*PopUpMenu, fyne.Window) {
	t.Helper()

	test.NewTempApp(t)
	win := test.NewTempWindow(t, widget.NewLabel("dummy"))

	menu := NewPopUpMenu(win.Canvas(),
		NewMenuItem("edit", func() { *edited = true }),
		NewMenuItemSeparator(),
		NewColoredMenuItem("delete", testRed, func() { *deleted = true }),
	)
	return menu, win
}

func TestNewPopUpMenuItemColors(t *testing.T) {
	var edited, deleted bool
	menu, _ := newTestPopUpMenu(t, &edited, &deleted)

	box, ok := menu.Content.(*fyne.Container)
	require.True(t, ok)
	require.Len(t, box.Objects, 3)
	assert.IsType(t, &widget.Separator{}, box.Objects[1])
	require.Len(t, menu.rows, 2)

	// 色未指定の項目はテーマの前景色、指定した項目は指定色で表示される
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	assert.Equal(t, menu.rows[0].Theme().Color(theme.ColorNameForeground, variant), menu.rows[0].textColor())
	assert.Equal(t, color.Color(testRed), menu.rows[1].textColor())
}

func TestPopUpMenuTap(t *testing.T) {
	var edited, deleted bool
	menu, _ := newTestPopUpMenu(t, &edited, &deleted)

	menu.Show()
	require.True(t, menu.Visible())

	test.Tap(menu.rows[1])
	assert.True(t, deleted)
	assert.False(t, edited)
	assert.False(t, menu.Visible())
}

func TestPopUpMenuHoverSelectsItem(t *testing.T) {
	var edited, deleted bool
	menu, _ := newTestPopUpMenu(t, &edited, &deleted)
	menu.Show()

	menu.rows[1].MouseIn(&desktop.MouseEvent{})
	assert.True(t, menu.rows[1].isActive())
	assert.False(t, menu.rows[0].isActive())

	menu.rows[1].MouseOut()
	assert.False(t, menu.rows[1].isActive())
}

func TestPopUpMenuKeyboard(t *testing.T) {
	var edited, deleted bool
	menu, win := newTestPopUpMenu(t, &edited, &deleted)

	menu.Show()
	// キーイベントを受け取れるようフォーカスが設定される
	require.Equal(t, fyne.Focusable(menu.rows[0]), win.Canvas().Focused())
	// 表示直後はどの項目も選択されていない
	assert.Equal(t, -1, menu.active)

	focused := win.Canvas().Focused()
	focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	assert.True(t, menu.rows[0].isActive())

	// 区切り線は飛ばして次の項目へ移動し、末尾からは先頭に戻る
	focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	assert.True(t, menu.rows[1].isActive())
	focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	assert.True(t, menu.rows[0].isActive())

	// 上キーは先頭から末尾へ循環する
	focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	assert.True(t, menu.rows[1].isActive())

	// Enter で選択中の項目を実行し、メニューを閉じる
	focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	assert.True(t, deleted)
	assert.False(t, edited)
	assert.False(t, menu.Visible())
}

func TestPopUpMenuEscapeClosesWithoutAction(t *testing.T) {
	var edited, deleted bool
	menu, win := newTestPopUpMenu(t, &edited, &deleted)

	menu.Show()
	focused := win.Canvas().Focused()
	require.NotNil(t, focused)

	focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})

	assert.False(t, menu.Visible())
	assert.False(t, edited)
	assert.False(t, deleted)
}

func TestPopUpMenuKeepsInsideCanvas(t *testing.T) {
	var edited, deleted bool
	menu, win := newTestPopUpMenu(t, &edited, &deleted)
	win.Resize(fyne.NewSize(240, 200))

	areaPos, areaSize := win.Canvas().InteractiveArea()
	t.Logf("interactive area pos=%v size=%v menu=%v", areaPos, areaSize, menu.Size())

	// 画面右下をはみ出す位置を指定しても、全体が収まる位置に補正される
	menu.ShowAtPosition(fyne.NewPos(areaSize.Width-5, areaSize.Height-5))
	pos := menu.Position()
	assert.LessOrEqual(t, pos.X+menu.Size().Width, areaPos.X+areaSize.Width)
	assert.LessOrEqual(t, pos.Y+menu.Size().Height, areaPos.Y+areaSize.Height)
	assert.GreaterOrEqual(t, pos.X, areaPos.X)
	assert.GreaterOrEqual(t, pos.Y, areaPos.Y)
	menu.Hide()

	// 収まる位置なら補正しない
	menu.ShowAtPosition(fyne.NewPos(10, 10))
	assert.Equal(t, fyne.NewPos(10, 10), menu.Position())
}

func TestPopUpMenuRelativePositionKeepsInsideCanvas(t *testing.T) {
	test.NewTempApp(t)

	// 画面下部にメニューボタンがある状態（一覧の最下段エントリ相当）を再現する
	anchor := widget.NewButton("menu", nil)
	win := test.NewTempWindow(t, container.NewVBox(
		widget.NewLabel("entry1"),
		widget.NewLabel("entry2"),
		widget.NewLabel("entry3"),
		anchor,
	))
	win.Resize(fyne.NewSize(240, 200))

	menu := NewPopUpMenu(win.Canvas(),
		NewMenuItem("move up", nil),
		NewMenuItem("move down", nil),
		NewMenuItemSeparator(),
		NewMenuItem("edit", nil),
		NewMenuItem("show qr", nil),
		NewMenuItemSeparator(),
		NewColoredMenuItem("delete", testRed, nil),
	)

	// totplist_item.go と同じ相対位置でメニューを開く
	rel := fyne.NewPos(anchor.Size().Width/2-menu.Size().Width, anchor.Size().Height/2)
	menu.ShowAtRelativePosition(rel, anchor)

	areaPos, areaSize := win.Canvas().InteractiveArea()
	pos := menu.Position()
	assert.LessOrEqual(t, pos.Y+menu.Size().Height, areaPos.Y+areaSize.Height)
	assert.LessOrEqual(t, pos.X+menu.Size().Width, areaPos.X+areaSize.Width)
	assert.GreaterOrEqual(t, pos.X, areaPos.X)
	assert.GreaterOrEqual(t, pos.Y, areaPos.Y)
}

func TestPopUpMenuLargerThanCanvas(t *testing.T) {
	var edited, deleted bool
	menu, win := newTestPopUpMenu(t, &edited, &deleted)

	// メニューより小さい画面では左上に寄せる（はみ出す分は下側に出す）。
	// オーバーレイによる中央寄せを避けるため、原点とは一致させない
	win.Resize(fyne.NewSize(40, 40))
	areaPos, _ := win.Canvas().InteractiveArea()
	menu.ShowAtPosition(fyne.NewPos(30, 30))
	assert.Equal(t, areaPos.AddXY(minMenuOffset, 0), menu.Position())
}
