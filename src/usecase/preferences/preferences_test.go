package preferences

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestManager_TOTPGroupsData(t *testing.T) {
	m := New(test.NewApp().Preferences())

	// 未設定時は空文字列
	assert.Empty(t, m.GetTOTPGroupsData())

	m.SetTOTPGroupsData("encrypted-groups")
	assert.Equal(t, "encrypted-groups", m.GetTOTPGroupsData())

	// TOTPデータとは独立している
	assert.Empty(t, m.GetTOTPData())
}

func TestManager_SelectedGroup(t *testing.T) {
	m := New(test.NewApp().Preferences())

	// 未設定時は空文字列（すべて）
	assert.Empty(t, m.GetSelectedGroup())

	m.SetSelectedGroup("group-1")
	assert.Equal(t, "group-1", m.GetSelectedGroup())

	m.SetSelectedGroup("__ungrouped__")
	assert.Equal(t, "__ungrouped__", m.GetSelectedGroup())

	m.SetSelectedGroup("")
	assert.Empty(t, m.GetSelectedGroup())
}
