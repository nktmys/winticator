package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"github.com/nktmys/winticator/src/ui/custom/components"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// groupLabel は選択値に対応する表示名を返す
func (t *totpListTab) groupLabel(selection string) string {
	switch selection {
	case totpstore.SelectionAll:
		return lang.L("totp.group.all")
	case totpstore.SelectionUngrouped:
		return lang.L("totp.group.ungrouped")
	}
	for _, g := range t.store.GetGroups() {
		if g.ID == selection {
			return g.Name
		}
	}
	return lang.L("totp.group.ungrouped")
}

// groupButtonLabelMaxRunes はグループボタンに表示する最大文字数
const groupButtonLabelMaxRunes = 20

// groupButtonLabel はグループボタン用の表示名を返す（長い名前は省略する。メニューは全文表示）
func (t *totpListTab) groupButtonLabel(selection string) string {
	runes := []rune(t.groupLabel(selection))
	if len(runes) <= groupButtonLabelMaxRunes {
		return string(runes)
	}
	return string(runes[:groupButtonLabelMaxRunes]) + "…"
}

// updateGroupButton はグループ選択ボタンの表示を現在の選択に合わせる
func (t *totpListTab) updateGroupButton() {
	t.groupButton.SetText(t.groupButtonLabel(t.selectedGroup))
}

// selectGroup はグループを選択して保存し、リストを絞り込む
func (t *totpListTab) selectGroup(selection string) {
	t.selectedGroup = totpstore.ResolveSelection(t.store.GetGroups(), selection)
	t.app.preferences.SetSelectedGroup(t.selectedGroup)
	t.updateGroupButton()
	t.filterEntries(t.searchEntry.Text)
}

// defaultGroupID は新規追加エントリの初期グループIDを返す
func (t *totpListTab) defaultGroupID() string {
	return totpstore.GroupIDForSelection(t.store.GetGroups(), t.selectedGroup)
}

// showGroupMenu はグループ選択メニューをボタンの直下に表示する
func (t *totpListTab) showGroupMenu() {
	groups := t.store.GetGroups()

	// 選択中の項目には先頭にチェックマークを付ける
	newItem := func(sel, name string) *components.MenuItem {
		n := len(totpstore.FilterByGroup(t.entries, groups, sel))
		label := fmt.Sprintf("%s (%d)", name, n)
		if sel == t.selectedGroup {
			label = "✓ " + label
		}
		return components.NewMenuItem(label, func() { t.selectGroup(sel) })
	}

	items := []*components.MenuItem{
		newItem(totpstore.SelectionAll, lang.L("totp.group.all")),
		components.NewMenuItemSeparator(),
	}
	for _, g := range groups {
		items = append(items, newItem(g.ID, g.Name))
	}
	if len(totpstore.FilterByGroup(t.entries, groups, totpstore.SelectionUngrouped)) > 0 {
		items = append(items, newItem(totpstore.SelectionUngrouped, lang.L("totp.group.ungrouped")))
	}

	items = append(items,
		components.NewMenuItemSeparator(),
		components.NewMenuItem(lang.L("totp.group.manage"), t.showGroupManageDialog),
	)

	popup := components.NewPopUpMenu(t.app.mainWindow.Canvas(), items...)
	popup.ShowAtRelativePosition(fyne.NewPos(0, t.groupButton.Size().Height), t.groupButton)
}

// realGroupID は実在するグループのIDならそのまま、空・孤立なら空文字（未分類）を返す
func (t *totpListTab) realGroupID(groupID string) string {
	for _, g := range t.store.GetGroups() {
		if g.ID == groupID {
			return groupID
		}
	}
	return ""
}

// showMoveGroupMenu はエントリの移動先グループメニューをアンカー位置に表示する
func (t *totpListTab) showMoveGroupMenu(entry *totpstore.Entry, anchor fyne.CanvasObject) {
	items := t.moveGroupMenuItems(entry)
	if len(items) == 0 {
		return
	}
	popup := components.NewPopUpMenu(t.app.mainWindow.Canvas(), items...)
	rel := fyne.NewPos(anchor.Size().Width/2-popup.Size().Width, anchor.Size().Height/2)
	popup.ShowAtRelativePosition(rel, anchor)
}

// moveGroupMenuItems はエントリの現在のグループを除いた移動先メニュー項目を返す
func (t *totpListTab) moveGroupMenuItems(entry *totpstore.Entry) []*components.MenuItem {
	current := t.realGroupID(entry.GroupID)
	var items []*components.MenuItem
	for _, g := range t.store.GetGroups() {
		if g.ID == current {
			continue
		}
		items = append(items, components.NewMenuItem(g.Name, func() { t.moveEntryToGroup(entry.ID, g.ID) }))
	}
	if current != "" {
		items = append(items, components.NewMenuItem(lang.L("totp.group.ungrouped"), func() { t.moveEntryToGroup(entry.ID, "") }))
	}
	return items
}

// moveEntryToGroup はエントリを指定グループの末尾へ移動して保存する
func (t *totpListTab) moveEntryToGroup(entryID, groupID string) {
	if err := t.store.MoveEntryToGroup(entryID, groupID); err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}
	if err := t.store.Save(); err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}
	t.refreshEntries()
}
