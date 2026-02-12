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
func (v *totpListView) createListItem() fyne.CanvasObject {
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
func (v *totpListView) updateListItem(id widget.ListItemID, item fyne.CanvasObject) {
	if id >= len(v.entries) {
		return
	}

	entry := v.entries[id]
	blocker := item.(*components.HoverBlocker)
	border := blocker.Content.(*fyne.Container)

	// 左側のコンテンツを取得
	leftContent := border.Objects[0].(*fyne.Container)
	displayNameLabel := leftContent.Objects[0].(*widget.Label)
	paddedCode := leftContent.Objects[1].(*fyne.Container)
	codeText := paddedCode.Objects[1].(*components.StyledText)

	// 右側のコンテンツを取得
	rightBox := border.Objects[1].(*fyne.Container)
	circularProgress := rightBox.Objects[0].(*components.CircularProgress)
	menuButton := rightBox.Objects[1].(*widget.Button)

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
	total := len(v.entries)
	menuButton.OnTapped = func() {
		v.showEntryMenu(entryCopy, menuButton, index, total)
	}
}

// copyCode はTOTPコードをクリップボードにコピーする
func (v *totpListView) copyCode(entry *totpstore.Entry) {
	code, err := entry.TOTP()
	if err != nil {
		return
	}

	// クリップボードにコピーし、有効時間+3秒後にクリアをスケジュール
	remaining := time.Duration(entry.RemainingSeconds()+3) * time.Second
	v.app.clipboard.Copy(code, remaining)

	// トースト通知を表示
	components.ShowToast(
		v.app.mainWindow,
		lang.L("totp.copied.message"),
	)
}

// showEntryMenu はエントリのメニューを表示する
func (v *totpListView) showEntryMenu(entry *totpstore.Entry, anchor fyne.CanvasObject, index int, total int) {
	var items []*fyne.MenuItem

	// 先頭でなければ「上へ移動」を表示
	if index > 0 {
		items = append(items, fyne.NewMenuItem(lang.L("totp.menu.moveup"), func() {
			v.moveEntry(entry.ID, -1)
		}))
	}

	// 末尾でなければ「下へ移動」を表示
	if index < total-1 {
		items = append(items, fyne.NewMenuItem(lang.L("totp.menu.movedown"), func() {
			v.moveEntry(entry.ID, +1)
		}))
	}

	// 移動メニューがある場合はセパレータを追加
	if len(items) > 0 {
		items = append(items, fyne.NewMenuItemSeparator())
	}

	items = append(items,
		fyne.NewMenuItem(lang.L("totp.menu.edit"), func() {
			v.showEditDialog(entry)
		}),
		fyne.NewMenuItem(lang.L("totp.menu.showqr"), func() {
			v.showQRCode(entry)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(lang.L("totp.menu.delete"), func() {
			v.confirmDelete(entry)
		}),
	)

	menu := fyne.NewMenu("", items...)
	popup := widget.NewPopUpMenu(menu, v.app.mainWindow.Canvas())
	rel := fyne.NewPos(anchor.Size().Width/2-popup.Size().Width, anchor.Size().Height/2)
	popup.ShowAtRelativePosition(rel, anchor)
}

// moveEntry はエントリを指定方向に移動する（direction: -1=上, +1=下）
func (v *totpListView) moveEntry(id string, direction int) {
	// 現在のエントリからIDリストを生成
	ids := make([]string, len(v.entries))
	targetIdx := -1
	for i, e := range v.entries {
		ids[i] = e.ID
		if e.ID == id {
			targetIdx = i
		}
	}

	if targetIdx < 0 {
		return
	}

	// 隣接エントリと入れ替え
	swapIdx := targetIdx + direction
	if swapIdx < 0 || swapIdx >= len(ids) {
		return
	}
	ids[targetIdx], ids[swapIdx] = ids[swapIdx], ids[targetIdx]

	// 永続化してリスト更新
	if err := v.store.Reorder(ids); err != nil {
		return
	}
	if err := v.store.Save(); err != nil {
		return
	}
	v.refreshEntries()
}

// showEditDialog は編集ダイアログを表示する
func (v *totpListView) showEditDialog(entry *totpstore.Entry) {
	v.showEntryFormDialog(
		entry,
		lang.L("totp.edit.title"),
		lang.L("dialog.save"),
		func(e *totpstore.Entry) error {
			return v.store.Update(e)
		},
	)
}

// showQRCode はQRコードを表示する
func (v *totpListView) showQRCode(entry *totpstore.Entry) {
	uri := entry.ToOTPAuthURI()

	// QRコード生成
	qr, err := qrscanner.GenerateQRCodeImage(uri)
	if err != nil {
		dialog.ShowError(err, v.app.mainWindow)
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
		v.app.mainWindow,
	)
}

// confirmDelete は削除確認ダイアログを表示する
func (v *totpListView) confirmDelete(entry *totpstore.Entry) {
	dialog.ShowConfirm(
		lang.L("totp.delete.title"),
		lang.L("totp.delete.message", M{"displayName": entry.DisplayName()}),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := v.store.Delete(entry.ID); err != nil {
				dialog.ShowError(err, v.app.mainWindow)
				return
			}
			if err := v.store.Save(); err != nil {
				dialog.ShowError(err, v.app.mainWindow)
				return
			}
			v.refreshEntries()
		},
		v.app.mainWindow,
	)
}
