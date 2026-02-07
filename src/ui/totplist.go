package ui

import (
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
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
	v.showEntryFormDialog(
		entry,
		lang.L("totp.add.title"),
		lang.L("dialog.add"),
		func(e *totpstore.Entry) error {
			return v.store.Add(e)
		},
	)
}

// showEntryFormDialog はエントリのフォームダイアログを表示する共通ヘルパー
func (v *totpListView) showEntryFormDialog(
	entry *totpstore.Entry,
	title, confirmLabel string,
	onSave func(*totpstore.Entry) error,
) {
	issuerEntry := widget.NewEntry()
	issuerEntry.SetText(entry.Issuer)

	accountEntry := widget.NewEntry()
	accountEntry.SetText(entry.Account)

	form := dialog.NewForm(
		title,
		confirmLabel,
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
			if err := onSave(entry); err != nil {
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
