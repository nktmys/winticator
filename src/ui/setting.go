package ui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/assets"
	"github.com/nktmys/winticator/src/ui/custom"
	"github.com/nktmys/winticator/src/usecase/preferences"
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
