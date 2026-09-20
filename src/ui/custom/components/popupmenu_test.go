package components

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
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
