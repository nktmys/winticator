package ui

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// groupNameError はグループ名に関するエラーを表示用のエラーに変換する
func groupNameError(err error) error {
	switch {
	case errors.Is(err, totpstore.ErrGroupNameEmpty):
		return errors.New(lang.L("totp.group.error.empty"))
	case errors.Is(err, totpstore.ErrGroupNameDuplicate):
		return errors.New(lang.L("totp.group.error.duplicate"))
	}
	return err
}

// checkReservedGroupName は予約された名前（すべて/未分類）かを検査する
func checkReservedGroupName(name string) error {
	trimmed := strings.TrimSpace(name)
	// 現在の表示言語に加え、他言語の表記（日本語・英語）も予約名として扱う
	reserved := []string{
		lang.L("totp.group.all"), lang.L("totp.group.ungrouped"),
		"すべて", "未分類", "All", "Ungrouped",
	}
	if slices.Contains(reserved, trimmed) {
		return errors.New(lang.L("totp.group.error.reserved"))
	}
	return nil
}

// showGroupManageDialog はグループ管理ダイアログを表示する
func (t *totpListTab) showGroupManageDialog() {
	win := t.app.mainWindow
	rows := container.NewVBox()

	// 保存して行を作り直す。失敗時はエラーを表示する
	var rebuild func()
	commit := func(err error) {
		if err == nil {
			err = t.store.Save()
		}
		if err != nil {
			dialog.ShowError(groupNameError(err), win)
		}
		rebuild()
	}

	move := func(groups []*totpstore.Group, i, delta int) {
		ids := make([]string, len(groups))
		for k, g := range groups {
			ids[k] = g.ID
		}
		ids[i], ids[i+delta] = ids[i+delta], ids[i]
		commit(t.store.ReorderGroups(ids))
	}

	rename := func(g *totpstore.Group) {
		nameEntry := widget.NewEntry()
		nameEntry.SetText(g.Name)
		form := dialog.NewForm(
			lang.L("totp.group.rename.title"),
			lang.L("dialog.save"),
			lang.L("dialog.cancel"),
			[]*widget.FormItem{widget.NewFormItem(lang.L("totp.group.name"), nameEntry)},
			func(confirmed bool) {
				if !confirmed {
					return
				}
				if err := checkReservedGroupName(nameEntry.Text); err != nil {
					dialog.ShowError(err, win)
					return
				}
				commit(t.store.RenameGroup(g.ID, nameEntry.Text))
			},
			win,
		)
		form.Resize(fyne.NewSize(400, 150))
		form.Show()
	}

	remove := func(g *totpstore.Group, count int) {
		dialog.ShowConfirm(
			lang.L("totp.group.delete.title"),
			lang.L("totp.group.delete.message", M{"Name": g.Name, "Count": strconv.Itoa(count)}),
			func(confirmed bool) {
				if confirmed {
					commit(t.store.DeleteGroup(g.ID))
				}
			},
			win,
		)
	}

	rebuild = func() {
		groups := t.store.GetGroups()
		entries := t.store.GetAll()
		rows.Objects = nil
		for i, g := range groups {
			count := len(totpstore.FilterByGroup(entries, groups, g.ID))
			label := widget.NewLabel(fmt.Sprintf("%s (%d)", g.Name, count))
			label.Truncation = fyne.TextTruncateEllipsis

			up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { move(groups, i, -1) })
			down := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { move(groups, i, 1) })
			if i == 0 {
				up.Disable()
			}
			if i == len(groups)-1 {
				down.Disable()
			}
			edit := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() { rename(g) })
			del := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() { remove(g, count) })

			rows.Add(container.NewBorder(nil, nil, nil, container.NewHBox(up, down, edit, del), label))
		}
		rows.Refresh()
	}

	newEntry := widget.NewEntry()
	newEntry.SetPlaceHolder(lang.L("totp.group.new.placeholder"))
	create := func() {
		if err := checkReservedGroupName(newEntry.Text); err != nil {
			dialog.ShowError(err, win)
			return
		}
		_, err := t.store.AddGroup(newEntry.Text)
		if err == nil {
			newEntry.SetText("")
		}
		commit(err)
	}
	newEntry.OnSubmitted = func(string) { create() }
	createButton := widget.NewButton(lang.L("totp.group.create"), create)

	rebuild()

	content := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, createButton, newEntry),
		nil, nil,
		container.NewVScroll(rows),
	)

	d := dialog.NewCustom(lang.L("totp.group.manage.title"), lang.L("dialog.close"), content, win)
	d.SetOnClosed(t.refreshEntries)
	d.Resize(fyne.NewSize(400, 350))
	d.Show()
}
