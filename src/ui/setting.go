package ui

import (
	"encoding/base64"
	"encoding/json"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/assets"
	"github.com/nktmys/winticator/src/ui/custom"
	"github.com/nktmys/winticator/src/usecase/crypto"
	"github.com/nktmys/winticator/src/usecase/preferences"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// createSettingTab は設定タブのUIを構築する
func (a *App) createSettingTab() fyne.CanvasObject {
	tab := &settingTab{
		app:         a,
		preferences: a.preferences,
		setting:     a.fyneApp.Settings(),
	}

	// テーマ設定
	themeLabel := widget.NewLabel(lang.L("setting.theme"))

	// 現在のテーマ設定を取得
	currentVariant := a.preferences.GetThemeVariant()

	// テーマ選択ラジオボタン
	themeOptions := []string{
		lang.L("setting.theme.dark"),
		lang.L("setting.theme.light"),
	}
	tab.themeRadio = widget.NewRadioGroup(themeOptions, tab.handleThemeRadio)
	tab.themeRadio.Horizontal = true

	// 現在の設定を反映
	if fyne.ThemeVariant(currentVariant) == theme.VariantDark {
		tab.themeRadio.SetSelected(lang.L("setting.theme.dark"))
	} else {
		tab.themeRadio.SetSelected(lang.L("setting.theme.light"))
	}

	// 言語設定
	languageLabel := widget.NewLabel(lang.L("setting.language"))

	// 利用可能な言語一覧を取得し、UIオプションを構築
	tab.locales = append(
		[]assets.Locale{{Code: "", Name: lang.L("setting.language.system")}},
		assets.AvailableLocales()...,
	)

	languageOptions := make([]string, len(tab.locales))
	for i, locale := range tab.locales {
		languageOptions[i] = locale.Name
	}

	tab.languageSelect = widget.NewSelect(languageOptions, tab.handleLanguageSelect)

	// 再起動通知ラベル（SetSelectedより前に初期化する必要がある）
	tab.restartLabel = widget.NewLabel(lang.L("setting.language.restart"))
	tab.restartLabel.Hide()

	// 現在の言語設定を反映
	currentLanguage := a.preferences.GetLanguage()
	currentIndex := slices.IndexFunc(tab.locales, func(l assets.Locale) bool {
		return l.Code == currentLanguage
	})
	if currentIndex >= 0 {
		tab.languageSelect.SetSelected(tab.locales[currentIndex].Name)
	} else {
		tab.languageSelect.SetSelected(tab.locales[0].Name) // システム設定をデフォルト
	}

	// データ管理セクション
	dataLabel := widget.NewLabel(lang.L("setting.data"))
	exportButton := widget.NewButton(lang.L("setting.export"), tab.handleExport)
	importButton := widget.NewButton(lang.L("setting.import"), tab.handleImport)
	dataButtons := container.NewHBox(exportButton, importButton)

	content := container.NewVBox(
		widget.NewLabel(lang.L("setting.header")),
		widget.NewSeparator(),
		themeLabel,
		tab.themeRadio,
		widget.NewSeparator(),
		languageLabel,
		tab.languageSelect,
		tab.restartLabel,
		widget.NewSeparator(),
		dataLabel,
		dataButtons,
	)

	return container.NewPadded(content)
}

// settingTab は設定タブの状態を保持する
type settingTab struct {
	app            *App
	preferences    *preferences.Preferences
	setting        fyne.Settings
	themeRadio     *widget.RadioGroup
	languageSelect *widget.Select
	restartLabel   *widget.Label
	locales        []assets.Locale
}

// handleThemeRadio はテーマ変更時の処理を行う
func (t *settingTab) handleThemeRadio(selected string) {
	index := slices.Index(t.themeRadio.Options, selected)
	if index < 0 {
		return
	}

	// 値に変更がなければ何もしない
	if t.preferences.GetThemeVariant() == index {
		return
	}

	variant := fyne.ThemeVariant(index)
	t.preferences.SetThemeVariant(index)
	t.setting.SetTheme(custom.NewTheme(variant))
}

// handleLanguageSelect は言語変更時の処理を行う
func (t *settingTab) handleLanguageSelect(selected string) {
	index := slices.IndexFunc(t.locales, func(l assets.Locale) bool {
		return l.Name == selected
	})
	if index < 0 {
		return
	}

	languageCode := t.locales[index].Code

	// 値に変更がなければ何もしない
	if t.preferences.GetLanguage() == languageCode {
		return
	}

	t.preferences.SetLanguage(languageCode)

	// 再起動が必要な旨を表示
	t.restartLabel.Show()
}

// handleExport はエクスポート処理を行う
func (t *settingTab) handleExport() {
	// パスワード入力ダイアログ
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = lang.L("setting.export.password")

	dialog.ShowForm(
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

		dialog.ShowForm(
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
