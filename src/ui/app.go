package ui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/assets"
	"github.com/nktmys/winticator/src/ui/custom"
	"github.com/nktmys/winticator/src/usecase/preferences"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

// M はlang用のジェネリックマップ型
type M map[string]any

// ページ識別子
type pageID int

const (
	pageTOTP pageID = iota
	pageSetting
	pageAppInfo
)

// App はアプリケーションのUI状態を保持する
type App struct {
	fyneApp      fyne.App
	mainWindow   fyne.Window
	preferences  *preferences.Preferences
	totpStore    *totpstore.Store
	totpListView *totpListView

	// ページコンテナ
	pageContainer *fyne.Container
	pages         map[pageID]fyne.CanvasObject
	currentPage   pageID

	// ツールバーボタン
	totpButton    *widget.Button
	settingButton *widget.Button
	infoButton    *widget.Button
	addButton     *widget.Button

	// クリップボード管理
	clipboardTimer *time.Timer
	copiedCode     string
	clipboardMu    sync.Mutex
}

// NewApp は新しいアプリケーションインスタンスを作成する
func NewApp() *App {
	fyneApp := app.New()
	prefs := preferences.New(fyneApp.Preferences())

	// 保存されたテーマ設定を読み込み、なければLightをデフォルトに
	variant := prefs.GetThemeVariant()
	fyneApp.Settings().SetTheme(custom.NewTheme(fyne.ThemeVariant(variant)))

	// TOTPストアを作成
	store := totpstore.New(prefs)

	return &App{
		fyneApp:     fyneApp,
		preferences: prefs,
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
		a.clearClipboard()
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
	a.pages[pageTOTP] = a.createTOTPListView()
	a.pages[pageSetting] = a.createSettingView()
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

	a.settingButton = widget.NewButtonWithIcon(lang.L("toolbar.setting"), theme.SettingsIcon(), func() {
		a.showPage(pageSetting)
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
		a.settingButton,
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
	a.settingButton.Importance = widget.MediumImportance
	a.infoButton.Importance = widget.MediumImportance

	switch a.currentPage {
	case pageTOTP:
		a.totpButton.Importance = widget.HighImportance
		a.addButton.Show()
	case pageSetting:
		a.settingButton.Importance = widget.HighImportance
		a.addButton.Hide()
	case pageAppInfo:
		a.infoButton.Importance = widget.HighImportance
		a.addButton.Hide()
	}

	a.totpButton.Refresh()
	a.settingButton.Refresh()
	a.infoButton.Refresh()
}

// scheduleClipboardClear はクリップボードの内容を指定時間後にクリアする
func (a *App) scheduleClipboardClear(copiedCode string, delay time.Duration) {
	a.clipboardMu.Lock()
	defer a.clipboardMu.Unlock()

	if a.clipboardTimer != nil {
		a.clipboardTimer.Stop()
	}

	a.copiedCode = copiedCode
	a.clipboardTimer = time.AfterFunc(delay, func() {
		fyne.Do(func() {
			a.clipboardMu.Lock()
			code := a.copiedCode
			a.copiedCode = ""
			a.clipboardMu.Unlock()

			if code != "" && a.fyneApp.Clipboard().Content() == code {
				a.fyneApp.Clipboard().SetContent("")
			}
		})
	})
}

// clearClipboard はコピーしたTOTPコードをクリップボードからクリアする
func (a *App) clearClipboard() {
	a.clipboardMu.Lock()
	if a.clipboardTimer != nil {
		a.clipboardTimer.Stop()
		a.clipboardTimer = nil
	}
	code := a.copiedCode
	a.copiedCode = ""
	a.clipboardMu.Unlock()

	if code != "" && a.fyneApp.Clipboard().Content() == code {
		a.fyneApp.Clipboard().SetContent("")
	}
}

// handleAddButton は追加ボタンの処理を行う
func (a *App) handleAddButton() {
	if a.totpListView != nil {
		a.totpListView.scanQRCode()
	}
}

// createSettingView は設定画面を作成する
func (a *App) createSettingView() fyne.CanvasObject {
	return a.createSettingTab()
}
