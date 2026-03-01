package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/ui/custom/components"
	"github.com/nktmys/winticator/src/usecase/clipboard"
	"github.com/nktmys/winticator/src/usecase/qrscanner"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// createTOTPListTab はTOTPリスト画面を作成する
func (a *App) createTOTPListTab() fyne.CanvasObject {
	view := &totpListTab{
		app:       a,
		store:     a.totpStore,
		clipboard: a.clipboard,
		entries:   make([]*totpstore.Entry, 0),
		stopChan:  make(chan bool),
	}

	// ストアからエントリを読み込み
	view.entries = view.store.GetAll()
	view.filteredEntries = make([]*totpstore.Entry, len(view.entries))
	copy(view.filteredEntries, view.entries)

	// 検索エントリを作成
	view.searchEntry = components.NewSearchEntry(lang.L("totp.search.placeholder"))
	view.searchEntry.OnChanged = func(query string) {
		view.filterEntries(query)
	}

	// リストを作成
	view.list = widget.NewList(
		func() int {
			return len(view.filteredEntries)
		},
		func() fyne.CanvasObject {
			return view.createListItem()
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			view.updateListItem(id, item)
		},
	)
	view.list.OnSelected = func(id widget.ListItemID) {
		// 選択・フォーカスを解除
		view.list.UnselectAll()
		a.mainWindow.Canvas().Unfocus()
		if id >= len(view.filteredEntries) {
			return
		}
		// タップでTOTPコードをコピー
		view.copyCode(view.filteredEntries[id])
	}

	// 空の場合のメッセージ
	view.emptyLabel = widget.NewLabel(lang.L("totp.empty"))
	view.emptyLabel.Alignment = fyne.TextAlignCenter

	// コンテナを作成
	listStack := container.NewStack(view.list, view.emptyLabel)
	view.container = container.NewBorder(view.searchEntry, nil, nil, nil, listStack)
	view.updateEmptyState()

	// 定期更新を開始
	view.startRefresh()

	// Appに参照を保持
	a.totpListView = view

	return container.NewPadded(view.container)
}

// totpListTab はTOTPリスト画面の状態を保持する
type totpListTab struct {
	app             *App
	store           *totpstore.Store
	clipboard       *clipboard.Manager
	list            *widget.List
	entries         []*totpstore.Entry
	filteredEntries []*totpstore.Entry
	searchEntry     *components.SearchEntry
	emptyLabel      *widget.Label
	ticker          *time.Ticker
	stopChan        chan bool
	container       *fyne.Container
}

// updateEmptyState は空の状態表示を更新する
func (t *totpListTab) updateEmptyState() {
	if len(t.filteredEntries) == 0 {
		if t.isSearching() {
			t.emptyLabel.SetText(lang.L("totp.search.empty"))
		} else {
			t.emptyLabel.SetText(lang.L("totp.empty"))
		}
		t.list.Hide()
		t.emptyLabel.Show()
	} else {
		t.emptyLabel.Hide()
		t.list.Show()
	}
}

// startRefresh は定期的な画面更新を開始する
func (t *totpListTab) startRefresh() {
	t.ticker = time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-t.ticker.C:
				fyne.Do(func() {
					t.list.Refresh()
				})
			case <-t.stopChan:
				t.ticker.Stop()
				return
			}
		}
	}()
}

// isSearching は検索中かどうかを返す
func (t *totpListTab) isSearching() bool {
	return t.searchEntry.Text != ""
}

// filterEntries は検索クエリに基づいてエントリをフィルタリングする
func (t *totpListTab) filterEntries(query string) {
	if query == "" {
		t.filteredEntries = make([]*totpstore.Entry, len(t.entries))
		copy(t.filteredEntries, t.entries)
	} else {
		q := strings.ToLower(query)
		t.filteredEntries = make([]*totpstore.Entry, 0)
		for _, entry := range t.entries {
			if strings.Contains(strings.ToLower(entry.Issuer), q) ||
				strings.Contains(strings.ToLower(entry.Account), q) {
				t.filteredEntries = append(t.filteredEntries, entry)
			}
		}
	}
	t.list.Refresh()

	// 検索中はAddボタンを無効化
	if t.isSearching() {
		t.app.addButton.Disable()
	} else {
		t.app.addButton.Enable()
	}

	t.updateEmptyState()
}

// refreshEntries はエントリリストを更新する
func (t *totpListTab) refreshEntries() {
	t.entries = t.store.GetAll()
	t.filterEntries(t.searchEntry.Text)
}

// scanQRCode はQRコードをスキャンしてエントリを追加する
func (t *totpListTab) scanQRCode() {
	go func() {
		// メインスレッドがHideを処理できるよう待機
		time.Sleep(500 * time.Millisecond)

		// スキャン実行
		results, err := qrscanner.CaptureAndScan()

		// UIスレッドでウィンドウ復帰と結果処理
		fyne.Do(func() {
			t.app.mainWindow.Show()
			t.app.mainWindow.RequestFocus()

			if err != nil {
				var errMsg string
				switch err {
				case qrscanner.ErrNoQRCodeFound:
					errMsg = lang.L("totp.scan.notfound")
				case qrscanner.ErrNoTOTPQRFound:
					errMsg = lang.L("totp.scan.nottotp")
				case totpstore.ErrNoTOTPEntries:
					errMsg = lang.L("totp.scan.nomigrationtotp")
				default:
					errMsg = fmt.Sprintf("%s: %v", lang.L("totp.scan.error"), err)
				}
				dialog.ShowError(errors.New(errMsg), t.app.mainWindow)
				return
			}

			if len(results) == 0 {
				dialog.ShowError(errors.New(lang.L("totp.scan.notfound")), t.app.mainWindow)
				return
			}

			if len(results) == 1 {
				t.showAddConfirmDialog(results[0].Entry)
			} else {
				t.showBatchAddConfirmDialog(results)
			}
		})
	}()

	// goroutine起動後にHide → 関数がすぐにreturnしイベントループがHideを処理
	t.app.mainWindow.Hide()
}

// showAddConfirmDialog は追加確認ダイアログを表示する
func (t *totpListTab) showAddConfirmDialog(entry *totpstore.Entry) {
	t.showEntryFormDialog(
		entry,
		lang.L("totp.add.title"),
		lang.L("dialog.add"),
		func(e *totpstore.Entry) error {
			return t.store.Add(e)
		},
	)
}

// showBatchAddConfirmDialog は複数エントリの一括追加確認ダイアログを表示する
func (t *totpListTab) showBatchAddConfirmDialog(results []qrscanner.ScanResult) {
	// エントリ名一覧を作成
	var names []string
	for _, r := range results {
		names = append(names, "- "+r.Entry.DisplayName())
	}
	count := strconv.Itoa(len(results))
	message := lang.L("totp.migration.confirm", M{"Count": count}) + "\n\n" + strings.Join(names, "\n")

	dialog.ShowConfirm(
		lang.L("totp.migration.title"),
		message,
		func(confirmed bool) {
			if !confirmed {
				return
			}
			for _, r := range results {
				if err := t.store.Add(r.Entry); err != nil {
					dialog.ShowError(err, t.app.mainWindow)
					return
				}
			}
			if err := t.store.Save(); err != nil {
				dialog.ShowError(err, t.app.mainWindow)
				return
			}
			t.refreshEntries()
			dialog.ShowInformation(
				lang.L("totp.migration.title"),
				lang.L("totp.migration.success", M{"Count": count}),
				t.app.mainWindow,
			)
		},
		t.app.mainWindow,
	)
}

// showEntryFormDialog はエントリのフォームダイアログを表示する共通ヘルパー
func (t *totpListTab) showEntryFormDialog(
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
	form.Resize(fyne.NewSize(400, 200))
	form.Show()
}
