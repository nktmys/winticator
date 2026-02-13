package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/assets"
	"github.com/nktmys/winticator/src/ui/custom"
	"github.com/nktmys/winticator/src/usecase/clipboard"
	"github.com/nktmys/winticator/src/usecase/preferences"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// M はlang用のジェネリックマップ型
type M map[string]any

// ページ識別子
type pageID int

const (
	pageTOTP pageID = iota
	pageSettings
	pageAppInfo
)

// App はアプリケーションのUI状態を保持する
type App struct {
	fyneApp     fyne.App
	preferences *preferences.Manager
	clipboard   *clipboard.Manager
	totpStore   *totpstore.Store
	mainWindow  fyne.Window

	// ページコンテナ
	pageContainer *fyne.Container
	pages         map[pageID]fyne.CanvasObject
	currentPage   pageID

	// TOTPリストビュー
	totpListView *totpListTab

	// ツールバーボタン
	totpButton     *widget.Button
	settingsButton *widget.Button
	infoButton     *widget.Button
	addButton      *widget.Button
}

// NewApp は新しいアプリケーションインスタンスを作成する
func NewApp() *App {
	fyneApp := app.New()
	preferences := preferences.New(fyneApp.Preferences())
	clipboard := clipboard.New(fyneApp.Clipboard())

	// 保存されたテーマ設定を読み込み、なければLightをデフォルトに
	variant := preferences.GetThemeVariant()
	fyneApp.Settings().SetTheme(custom.NewTheme(variant))

	// TOTPストアを作成
	store := totpstore.New(preferences)

	return &App{
		fyneApp:     fyneApp,
		preferences: preferences,
		clipboard:   clipboard,
		totpStore:   store,
		pages:       make(map[pageID]fyne.CanvasObject),
	}
}

// Run はアプリケーションを起動する
func (a *App) Run() {
	// 保存された言語設定を読み込み、翻訳を初期化
	savedLanguage := a.preferences.GetLanguage()
	_ = assets.InitI18nWithLocale(savedLanguage)

	// TOTPデータを読み込み（エラーが発生しても空のストアとして続行）
	_ = a.totpStore.Load()

	a.mainWindow = a.fyneApp.NewWindow(lang.L("app.title"))
	a.mainWindow.Resize(fyne.NewSize(650, 450))

	content := a.createUI()
	a.mainWindow.SetContent(content)

	// アプリ終了時にクリップボードをクリア
	a.mainWindow.SetCloseIntercept(func() {
		a.clipboard.Clear()
		a.mainWindow.Close()
	})

	a.mainWindow.ShowAndRun()
}

// createUI はUIコンポーネントを構築する
func (a *App) createUI() fyne.CanvasObject {
	// ツールバーを作成
	toolbar := a.createToolbar()

	// ページコンテナを作成
	a.pageContainer = container.NewStack()

	// 各ページを作成
	a.pages[pageTOTP] = a.createTOTPListTab()
	a.pages[pageSettings] = a.createSettingsTab()
	a.pages[pageAppInfo] = a.createAppInfoTab()

	// 初期ページを表示
	a.showPage(pageTOTP)

	// メインレイアウト
	return container.NewBorder(toolbar, nil, nil, nil, a.pageContainer)
}

// createToolbar はツールバーを作成する
func (a *App) createToolbar() fyne.CanvasObject {
	// ナビゲーションボタン
	a.totpButton = widget.NewButtonWithIcon(lang.L("toolbar.totp"), theme.AccountIcon(), func() {
		a.showPage(pageTOTP)
	})

	a.settingsButton = widget.NewButtonWithIcon(lang.L("toolbar.settings"), theme.SettingsIcon(), func() {
		a.showPage(pageSettings)
	})

	a.infoButton = widget.NewButtonWithIcon(lang.L("toolbar.info"), theme.InfoIcon(), func() {
		a.showPage(pageAppInfo)
	})

	// 追加ボタン
	a.addButton = widget.NewButtonWithIcon(lang.L("toolbar.add"), theme.ContentAddIcon(), func() {
		a.handleAddButton()
	})
	a.addButton.Importance = widget.HighImportance

	// スペーサー
	spacer := widget.NewLabel("")

	// ツールバーコンテナ
	leftButtons := container.NewHBox(
		a.totpButton,
		a.settingsButton,
		a.infoButton,
	)

	toolbar := container.NewBorder(nil, nil, leftButtons, a.addButton, spacer)

	return container.NewVBox(toolbar, widget.NewSeparator())
}

// showPage は指定されたページを表示する
func (a *App) showPage(id pageID) {
	page, ok := a.pages[id]
	if !ok {
		return
	}

	// 現在のページをクリア
	a.pageContainer.Objects = nil

	// 新しいページを追加
	a.pageContainer.Add(page)
	a.pageContainer.Refresh()

	a.currentPage = id

	// ボタンの状態を更新
	a.updateToolbarState()
}

// updateToolbarState はツールバーボタンの状態を更新する
func (a *App) updateToolbarState() {
	// 現在のページに応じてボタンの状態を更新
	a.totpButton.Importance = widget.MediumImportance
	a.settingsButton.Importance = widget.MediumImportance
	a.infoButton.Importance = widget.MediumImportance

	switch a.currentPage {
	case pageTOTP:
		a.totpButton.Importance = widget.HighImportance
		a.addButton.Show()
	case pageSettings:
		a.settingsButton.Importance = widget.HighImportance
		a.addButton.Hide()
	case pageAppInfo:
		a.infoButton.Importance = widget.HighImportance
		a.addButton.Hide()
	}

	a.totpButton.Refresh()
	a.settingsButton.Refresh()
	a.infoButton.Refresh()
}

// handleAddButton は追加ボタンの処理を行う
func (a *App) handleAddButton() {
	if a.totpListView != nil {
		a.totpListView.scanQRCode()
	}
}
