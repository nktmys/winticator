package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/ui/custom"
	"github.com/nktmys/winticator/src/ui/custom/components"
	"github.com/nktmys/winticator/src/usecase/qrscanner"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// createListItem はリストアイテムのテンプレートを作成する
func (t *totpListTab) createListItem() fyne.CanvasObject {
	// 表示名（Account または Issuer）
	displayNameLabel := widget.NewLabel("DisplayName")

	// TOTPコード（青色・大きいフォント）
	codeText := components.NewStyledText("000 000", custom.ColorPrimaryBlue, 28)

	// 左パディング用のスペーサーを追加（widget.LabelのInnerPaddingに合わせる）
	codePadding := canvas.NewRectangle(color.Transparent)
	codePadding.SetMinSize(fyne.NewSize(theme.InnerPadding(), 0))
	paddedCode := container.NewHBox(codePadding, codeText)

	// 円形プログレス
	circularProgress := components.NewCircularProgress(32)

	// メニューボタン
	menuButton := widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), nil)

	// 左側: 表示名 + コード（パディング付き）
	leftContent := container.NewVBox(displayNameLabel, paddedCode)

	// 右側: 円形プログレス + メニュー
	rightContent := container.NewHBox(circularProgress, menuButton)

	content := container.NewBorder(nil, nil, leftContent, rightContent)
	return components.NewHoverBlocker(content)
}

// updateListItem はリストアイテムを更新する
func (t *totpListTab) updateListItem(id widget.ListItemID, item fyne.CanvasObject) {
	if id >= len(t.filteredEntries) {
		return
	}

	entry := t.filteredEntries[id]
	blocker, _ := item.(*components.HoverBlocker)
	border, _ := blocker.Content.(*fyne.Container)

	// 左側のコンテンツを取得
	leftContent, _ := border.Objects[0].(*fyne.Container)
	displayNameLabel, _ := leftContent.Objects[0].(*widget.Label)
	paddedCode, _ := leftContent.Objects[1].(*fyne.Container)
	codeText, _ := paddedCode.Objects[1].(*components.StyledText)

	// 右側のコンテンツを取得
	rightBox, _ := border.Objects[1].(*fyne.Container)
	circularProgress, _ := rightBox.Objects[0].(*components.CircularProgress)
	menuButton, _ := rightBox.Objects[1].(*widget.Button)
	// テーマ変更時にアイコンが更新されるようにする
	menuButton.SetIcon(theme.MoreHorizontalIcon())

	// 表示名を設定
	displayNameLabel.SetText(entry.DisplayName())

	// TOTPコードを生成
	code, err := entry.TOTP()
	if err != nil {
		codeText.SetText("------")
	} else {
		// 3桁ごとにスペースを挿入
		if len(code) == 6 {
			code = code[:3] + " " + code[3:]
		}
		codeText.SetText(code)
	}

	// 残り時間を設定
	remaining := entry.RemainingSeconds()
	circularProgress.Max = float64(entry.Period)
	circularProgress.SetValue(float64(remaining))

	// 残り5秒未満で赤色に変更
	if remaining < 5 {
		warningColor := custom.ColorPrimaryRed
		codeText.SetColor(warningColor)
		circularProgress.SetColor(warningColor)
	} else {
		normalColor := custom.ColorPrimaryBlue
		codeText.SetColor(normalColor)
		circularProgress.SetColor(normalColor)
	}

	// メニューボタン
	entryCopy := entry
	index := id
	total := len(t.filteredEntries)
	menuButton.OnTapped = func() {
		t.showEntryMenu(entryCopy, menuButton, index, total)
	}
}

// copyCode はTOTPコードをクリップボードにコピーする
func (t *totpListTab) copyCode(entry *totpstore.Entry) {
	code, err := entry.TOTP()
	if err != nil {
		return
	}

	// クリップボードにコピーし、有効時間+3秒後にクリアをスケジュール
	remaining := time.Duration(entry.RemainingSeconds()+3) * time.Second
	t.clipboard.Copy(code, remaining)

	// トースト通知を表示
	components.ShowToast(
		t.app.mainWindow,
		lang.L("totp.copied.message"),
	)
}

// showEntryMenu はエントリのメニューを表示する
func (t *totpListTab) showEntryMenu(entry *totpstore.Entry, anchor fyne.CanvasObject, index int, total int) {
	var items []*components.MenuItem

	// 検索中でなければ移動メニューを表示
	if !t.isSearching() {
		// 先頭でなければ「上へ移動」を表示
		if index > 0 {
			items = append(items, components.NewMenuItem(lang.L("totp.menu.moveup"), func() {
				t.moveEntry(entry.ID, -1)
			}))
		}

		// 末尾でなければ「下へ移動」を表示
		if index < total-1 {
			items = append(items, components.NewMenuItem(lang.L("totp.menu.movedown"), func() {
				t.moveEntry(entry.ID, +1)
			}))
		}

		// 移動メニューがある場合はセパレータを追加
		if len(items) > 0 {
			items = append(items, components.NewMenuItemSeparator())
		}
	}

	// 移動先がある場合のみ「グループへ移動…」を表示
	if len(t.moveGroupMenuItems(entry)) > 0 {
		items = append(items, components.NewMenuItem(lang.L("totp.menu.movegroup"), func() {
			t.showMoveGroupMenu(entry, anchor)
		}))
	}

	items = append(items,
		components.NewMenuItem(lang.L("totp.menu.edit"), func() {
			t.showEditDialog(entry)
		}),
		components.NewMenuItem(lang.L("totp.menu.showqr"), func() {
			t.showQRCode(entry)
		}),
		components.NewMenuItemSeparator(),
		// 削除は破壊的操作のため赤字で表示
		components.NewColoredMenuItem(lang.L("totp.menu.delete"), custom.ColorPrimaryRed, func() {
			t.confirmDelete(entry)
		}),
	)

	popup := components.NewPopUpMenu(t.app.mainWindow.Canvas(), items...)
	rel := fyne.NewPos(anchor.Size().Width/2-popup.Size().Width, anchor.Size().Height/2)
	popup.ShowAtRelativePosition(rel, anchor)
}

// moveEntry はエントリを表示中の隣のエントリと入れ替える（direction: -1=上, +1=下）
func (t *totpListTab) moveEntry(id string, direction int) {
	// 表示中のリスト上で隣のエントリを決める
	pos := -1
	for i, e := range t.filteredEntries {
		if e.ID == id {
			pos = i
			break
		}
	}
	if pos < 0 {
		return
	}
	neighborPos := pos + direction
	if neighborPos < 0 || neighborPos >= len(t.filteredEntries) {
		return
	}
	neighborID := t.filteredEntries[neighborPos].ID

	// 全エントリのIDリスト上で2件の位置を入れ替える
	ids, idx, neighborIdx := t.entryIDsWithPositions(id, neighborID)
	if idx < 0 || neighborIdx < 0 {
		return
	}
	ids[idx], ids[neighborIdx] = ids[neighborIdx], ids[idx]

	// 永続化してリスト更新
	if err := t.store.Reorder(ids); err != nil {
		return
	}
	if err := t.store.Save(); err != nil {
		return
	}
	t.refreshEntries()
}

// entryIDsWithPositions は全エントリのIDリストと、指定2件のそのリスト上の位置を返す（見つからなければ -1）
func (t *totpListTab) entryIDsWithPositions(a, b string) ([]string, int, int) {
	ids := make([]string, len(t.entries))
	aIdx, bIdx := -1, -1
	for i, e := range t.entries {
		ids[i] = e.ID
		switch e.ID {
		case a:
			aIdx = i
		case b:
			bIdx = i
		}
	}
	return ids, aIdx, bIdx
}

// showEditDialog は編集ダイアログを表示する
func (t *totpListTab) showEditDialog(entry *totpstore.Entry) {
	t.showEntryFormDialog(
		entry,
		lang.L("totp.edit.title"),
		lang.L("dialog.save"),
		func(e *totpstore.Entry) error {
			return t.store.Update(e)
		},
	)
}

// showQRCode はQRコードを表示する
func (t *totpListTab) showQRCode(entry *totpstore.Entry) {
	uri := entry.ToOTPAuthURI()

	// QRコード生成
	qr, err := qrscanner.GenerateQRCodeImage(uri)
	if err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}

	img := canvas.NewImageFromImage(qr)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(200, 200))

	content := container.NewVBox(
		img,
		widget.NewLabel(entry.DisplayName()),
	)

	dialog.ShowCustom(
		lang.L("totp.qr.title"),
		lang.L("dialog.close"),
		content,
		t.app.mainWindow,
	)
}

// confirmDelete は削除確認ダイアログを表示する
func (t *totpListTab) confirmDelete(entry *totpstore.Entry) {
	dialog.ShowConfirm(
		lang.L("totp.delete.title"),
		lang.L("totp.delete.message", M{"displayName": entry.DisplayName()}),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := t.store.Delete(entry.ID); err != nil {
				dialog.ShowError(err, t.app.mainWindow)
				return
			}
			if err := t.store.Save(); err != nil {
				dialog.ShowError(err, t.app.mainWindow)
				return
			}
			t.refreshEntries()
		},
		t.app.mainWindow,
	)
}
