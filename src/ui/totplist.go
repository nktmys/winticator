package ui

import (
	"errors"
	"fmt"
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

// createTOTPListView はTOTPリスト画面を作成する
func (a *App) createTOTPListView() fyne.CanvasObject {
	view := &totpListView{
		app:      a,
		store:    a.totpStore,
		entries:  make([]*totpstore.Entry, 0),
		stopChan: make(chan bool),
	}

	// ストアからエントリを読み込み
	view.entries = view.store.GetAll()

	// リストを作成
	view.list = widget.NewList(
		func() int {
			return len(view.entries)
		},
		func() fyne.CanvasObject {
			return view.createListItem()
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			view.updateListItem(id, item)
		},
	)

	// 空の場合のメッセージ
	emptyLabel := widget.NewLabel(lang.L("totp.empty"))
	emptyLabel.Alignment = fyne.TextAlignCenter

	// コンテナを作成
	view.container = container.NewStack(view.list, emptyLabel)
	view.updateEmptyState(emptyLabel)

	// 定期更新を開始
	view.startRefresh()

	// Appに参照を保持
	a.totpListView = view

	return container.NewPadded(view.container)
}

// totpListView はTOTPリスト画面の状態を保持する
type totpListView struct {
	app       *App
	store     *totpstore.Store
	list      *widget.List
	entries   []*totpstore.Entry
	ticker    *time.Ticker
	stopChan  chan bool
	container *fyne.Container
}

// createListItem はリストアイテムのテンプレートを作成する
func (v *totpListView) createListItem() fyne.CanvasObject {
	// 表示名（Account または Issuer）
	displayNameLabel := widget.NewLabel("DisplayName")

	// TOTPコード（青色・大きいフォント・タップ可能）
	codeText := components.NewTappableText("000 000",
		custom.ColorPrimaryBlue, 28)

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

	return container.NewBorder(nil, nil, leftContent, rightContent)
}

// updateListItem はリストアイテムを更新する
func (v *totpListView) updateListItem(id widget.ListItemID, item fyne.CanvasObject) {
	if id >= len(v.entries) {
		return
	}

	entry := v.entries[id]
	border := item.(*fyne.Container)

	// 左側のコンテンツを取得
	leftContent := border.Objects[0].(*fyne.Container)
	displayNameLabel := leftContent.Objects[0].(*widget.Label)
	paddedCode := leftContent.Objects[1].(*fyne.Container)
	codeText := paddedCode.Objects[1].(*components.TappableText)

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

	// コードクリックでコピー（entryをキャプチャ）
	entryCopy := entry
	codeText.OnTapped = func() {
		v.copyCode(entryCopy)
	}

	// メニューボタン
	menuButton.OnTapped = func() {
		v.showEntryMenu(entryCopy, menuButton)
	}
}

// updateEmptyState は空の状態表示を更新する
func (v *totpListView) updateEmptyState(emptyLabel *widget.Label) {
	if len(v.entries) == 0 {
		v.list.Hide()
		emptyLabel.Show()
	} else {
		emptyLabel.Hide()
		v.list.Show()
	}
}

// startRefresh は定期的な画面更新を開始する
func (v *totpListView) startRefresh() {
	v.ticker = time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-v.ticker.C:
				fyne.Do(func() {
					v.list.Refresh()
				})
			case <-v.stopChan:
				v.ticker.Stop()
				return
			}
		}
	}()
}

// copyCode はTOTPコードをクリップボードにコピーする
func (v *totpListView) copyCode(entry *totpstore.Entry) {
	code, err := entry.TOTP()
	if err != nil {
		return
	}

	v.app.fyneApp.Clipboard().SetContent(code)

	// コピー完了通知
	dialog.ShowInformation(
		lang.L("totp.copied.title"),
		lang.L("totp.copied.message", M{"displayName": entry.DisplayName()}),
		v.app.mainWindow,
	)
}

// showEntryMenu はエントリのメニューを表示する
func (v *totpListView) showEntryMenu(entry *totpstore.Entry, anchor fyne.CanvasObject) {
	items := []*fyne.MenuItem{
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
	}

	menu := fyne.NewMenu("", items...)
	popup := widget.NewPopUpMenu(menu, v.app.mainWindow.Canvas())
	rel := fyne.NewPos(anchor.Size().Width/2-popup.Size().Width, anchor.Size().Height/2)
	popup.ShowAtRelativePosition(rel, anchor)
}

// showEditDialog は編集ダイアログを表示する
func (v *totpListView) showEditDialog(entry *totpstore.Entry) {
	issuerEntry := widget.NewEntry()
	issuerEntry.SetText(entry.Issuer)

	accountEntry := widget.NewEntry()
	accountEntry.SetText(entry.Account)

	form := dialog.NewForm(
		lang.L("totp.edit.title"),
		lang.L("dialog.save"),
		lang.L("dialog.cancel"),
		[]*widget.FormItem{
			widget.NewFormItem(lang.L("totp.edit.issuer"), issuerEntry),
			widget.NewFormItem(lang.L("totp.edit.account"), accountEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			entry.Issuer = issuerEntry.Text
			entry.Account = accountEntry.Text
			if err := v.store.Update(entry); err != nil {
				dialog.ShowError(err, v.app.mainWindow)
				return
			}
			if err := v.store.Save(); err != nil {
				dialog.ShowError(err, v.app.mainWindow)
				return
			}
			v.list.Refresh()
		},
		v.app.mainWindow,
	)
	form.Resize(fyne.NewSize(400, 200))
	form.Show()
}

// showQRCode はQRコードを表示する
func (v *totpListView) showQRCode(entry *totpstore.Entry) {
	uri := entry.ToOTPAuthURI()

	// QRコード生成
	qr, err := generateQRCodeImage(uri)
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

// scanQRCode はQRコードをスキャンしてエントリを追加する
func (v *totpListView) scanQRCode() {
	go func() {
		// メインスレッドがHideを処理できるよう待機
		time.Sleep(500 * time.Millisecond)

		// スキャン実行
		results, err := qrscanner.CaptureAndScan()

		// UIスレッドでウィンドウ復帰と結果処理
		fyne.Do(func() {
			v.app.mainWindow.Show()
			v.app.mainWindow.RequestFocus()

			if err != nil {
				var errMsg string
				switch err {
				case qrscanner.ErrNoQRCodeFound:
					errMsg = lang.L("totp.scan.notfound")
				case qrscanner.ErrNoTOTPQRFound:
					errMsg = lang.L("totp.scan.nottotp")
				default:
					errMsg = fmt.Sprintf("%s: %v", lang.L("totp.scan.error"), err)
				}
				dialog.ShowError(errors.New(errMsg), v.app.mainWindow)
				return
			}

			if len(results) == 0 {
				dialog.ShowError(errors.New(lang.L("totp.scan.notfound")), v.app.mainWindow)
				return
			}

			result := results[0]
			v.showAddConfirmDialog(result.Entry)
		})
	}()

	// goroutine起動後にHide → 関数がすぐにreturnしイベントループがHideを処理
	v.app.mainWindow.Hide()
}

// showAddConfirmDialog は追加確認ダイアログを表示する
func (v *totpListView) showAddConfirmDialog(entry *totpstore.Entry) {
	issuerEntry := widget.NewEntry()
	issuerEntry.SetText(entry.Issuer)

	accountEntry := widget.NewEntry()
	accountEntry.SetText(entry.Account)

	form := dialog.NewForm(
		lang.L("totp.add.title"),
		lang.L("dialog.add"),
		lang.L("dialog.cancel"),
		[]*widget.FormItem{
			widget.NewFormItem(lang.L("totp.edit.issuer"), issuerEntry),
			widget.NewFormItem(lang.L("totp.edit.account"), accountEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			entry.Issuer = issuerEntry.Text
			entry.Account = accountEntry.Text
			if err := v.store.Add(entry); err != nil {
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
	form.Resize(fyne.NewSize(400, 200))
	form.Show()
}

// refreshEntries はエントリリストを更新する
func (v *totpListView) refreshEntries() {
	v.entries = v.store.GetAll()
	v.list.Refresh()

	// 空状態の更新
	if len(v.container.Objects) >= 2 {
		if emptyLabel, ok := v.container.Objects[1].(*widget.Label); ok {
			v.updateEmptyState(emptyLabel)
		}
	}
}
