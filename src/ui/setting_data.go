package ui

import (
	"encoding/base64"
	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/usecase/crypto"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// handleExport はエクスポート処理を行う
func (t *settingTab) handleExport() {
	// パスワード入力ダイアログ
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = lang.L("setting.export.password")

	form := dialog.NewForm(
		lang.L("setting.export.title"),
		lang.L("dialog.save"),
		lang.L("dialog.cancel"),
		[]*widget.FormItem{
			widget.NewFormItem("", passwordEntry),
		},
		func(confirmed bool) {
			if !confirmed || passwordEntry.Text == "" {
				return
			}
			t.doExport(passwordEntry.Text)
		},
		t.app.mainWindow,
	)
	form.Resize(fyne.NewSize(400, 160))
	form.Show()
}

// doExport は実際のエクスポート処理を行う
func (t *settingTab) doExport(password string) {
	// エントリを取得
	entries := t.app.totpStore.GetAll()
	if len(entries) == 0 {
		dialog.ShowInformation(
			lang.L("setting.export.title"),
			lang.L("totp.empty"),
			t.app.mainWindow,
		)
		return
	}

	// JSONにシリアライズ
	data, err := json.Marshal(entries)
	if err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}

	// パスワードで暗号化
	encrypted, err := crypto.Encrypt(password, data)
	if err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}

	// Base64エンコード
	encoded := []byte(base64.StdEncoding.EncodeToString(encrypted))

	// ファイル保存ダイアログ
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, t.app.mainWindow)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = writer.Write(encoded)
		if err != nil {
			dialog.ShowError(err, t.app.mainWindow)
			return
		}

		dialog.ShowInformation(
			lang.L("setting.export.title"),
			lang.L("setting.export.success"),
			t.app.mainWindow,
		)
	}, t.app.mainWindow)

	saveDialog.SetFileName("winticator_backup.wtbackup")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".wtbackup"}))
	saveDialog.Show()
}

// handleImport はインポート処理を行う
func (t *settingTab) handleImport() {
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, t.app.mainWindow)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// ファイルを読み込み
		data := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// パスワード入力ダイアログ
		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.PlaceHolder = lang.L("setting.import.password")

		form := dialog.NewForm(
			lang.L("setting.import.title"),
			lang.L("dialog.save"),
			lang.L("dialog.cancel"),
			[]*widget.FormItem{
				widget.NewFormItem("", passwordEntry),
			},
			func(confirmed bool) {
				if !confirmed || passwordEntry.Text == "" {
					return
				}
				t.doImport(data, passwordEntry.Text)
			},
			t.app.mainWindow,
		)
		form.Resize(fyne.NewSize(400, 160))
		form.Show()
	}, t.app.mainWindow)

	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".wtbackup"}))
	openDialog.Show()
}

// doImport は実際のインポート処理を行う
func (t *settingTab) doImport(data []byte, password string) {
	// Base64デコード
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}

	// パスワードで復号
	decrypted, err := crypto.Decrypt(password, decoded)
	if err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}

	// JSONをデシリアライズ
	var entries []*totpstore.Entry
	if err := json.Unmarshal(decrypted, &entries); err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}

	// 既存データとマージするか確認
	if t.app.totpStore.Count() > 0 {
		dialog.ShowConfirm(
			lang.L("setting.import.title"),
			lang.L("setting.import.merge"),
			func(merge bool) {
				if !merge {
					// 既存データを削除
					for _, e := range t.app.totpStore.GetAll() {
						_ = t.app.totpStore.Delete(e.ID)
					}
				}
				t.importEntries(entries)
			},
			t.app.mainWindow,
		)
	} else {
		t.importEntries(entries)
	}
}

// importEntries はエントリをインポートする
func (t *settingTab) importEntries(entries []*totpstore.Entry) {
	for _, entry := range entries {
		if err := t.app.totpStore.Add(entry); err != nil {
			dialog.ShowError(err, t.app.mainWindow)
			return
		}
	}

	if err := t.app.totpStore.Save(); err != nil {
		dialog.ShowError(err, t.app.mainWindow)
		return
	}

	// TOTPリストを更新
	if t.app.totpListView != nil {
		t.app.totpListView.refreshEntries()
	}

	dialog.ShowInformation(
		lang.L("setting.import.title"),
		lang.L("setting.import.success"),
		t.app.mainWindow,
	)
}
