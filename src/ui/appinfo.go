package ui

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/assets"
	"github.com/nktmys/winticator/src/pkg/version"
	"golang.org/x/sync/errgroup"
)

const (
	appLicense   = "Copyright (c) 2026 nktmys."
	licensesInfo = "MIT License."

	// GitHub URLs
	githubReleasesAPIURL = "https://api.github.com/repos/nktmys/winticator/releases/latest"
	downloadPageBaseURL  = "https://github.com/nktmys/winticator/releases/tag"
	licensesPageURL      = "https://github.com/nktmys/winticator/tree/main/licenses"
)

// createAppInfoTab はアプリ情報タブのUIを構築する
func (a *App) createAppInfoTab() fyne.CanvasObject {
	tab := &appInfoTab{
		fyneApp:    a.fyneApp,
		mainWindow: a.mainWindow,
	}

	// App metadata
	metadata := a.fyneApp.Metadata()
	tab.currentVersion = metadata.Version

	// アプリアイコン
	appIcon := canvas.NewImageFromResource(assets.IconApp)
	appIcon.FillMode = canvas.ImageFillContain
	appIcon.SetMinSize(fyne.NewSize(128, 128))

	// 名称
	nameLabel := widget.NewLabel(lang.L("appinfo.name"))
	nameValue := widget.NewLabel(metadata.Name)
	nameValue.TextStyle = fyne.TextStyle{Bold: true}

	// バージョン
	versionLabel := widget.NewLabel(lang.L("appinfo.version"))
	versionValue := widget.NewLabel(metadata.Version)

	// プログレスダイアログを作成
	tab.progressDialog = dialog.NewCustomWithoutButtons(
		lang.L("appinfo.checking"),
		container.NewVBox(
			widget.NewLabel(lang.L("appinfo.checkingMessage")),
			widget.NewProgressBarInfinite(),
		),
		tab.mainWindow,
	)

	// 新バージョン確認ボタン
	checkUpdateBtn := widget.NewButton(lang.L("appinfo.checkUpdate"), tab.handleCheckUpdateButton)

	// Copyright（下部固定用）
	copyrightLabel := widget.NewLabel(appLicense)
	copyrightLabel.Alignment = fyne.TextAlignCenter

	// レイアウト（3列: ラベル、値、ボタン）
	infoGrid := container.NewGridWithColumns(3,
		nameLabel, nameValue, layout.NewSpacer(),
		versionLabel, versionValue, container.NewBorder(nil, nil, nil, checkUpdateBtn),
	)

	// ライセンス
	licensesLabel := widget.NewLabel(lang.L("appinfo.licenses"))
	licensesValue := widget.NewLabel(licensesInfo)
	licensesValue.Wrapping = fyne.TextWrapWord

	// ライセンスセクション（ラベルの右端にボタン配置）
	licensesBtn := widget.NewButton(lang.L("appinfo.viewLicenses"), tab.handleLicensesButton)
	licensesHeader := container.NewBorder(nil, nil, licensesLabel, licensesBtn)

	// メインコンテンツ（スクロール可能な部分）
	mainContent := container.NewVBox(
		container.NewCenter(appIcon),
		widget.NewSeparator(),
		infoGrid,
		widget.NewSeparator(),
		licensesHeader,
		licensesValue,
	)

	// 下部にcopyrightを固定配置
	content := container.NewBorder(
		nil,                                 // top
		container.NewCenter(copyrightLabel), // bottom
		nil,                                 // left
		nil,                                 // right
		mainContent,                         // center
	)

	return container.NewPadded(content)
}

// appInfoTab はアプリ情報タブの状態を保持する
type appInfoTab struct {
	fyneApp        fyne.App             // Fyneアプリケーション
	mainWindow     fyne.Window          // メインウィンドウ
	currentVersion string               // 現在のバージョン
	progressDialog *dialog.CustomDialog // プログレスダイアログ
}

// handleCheckUpdateButton は新バージョン確認ボタンの処理を行う
func (t *appInfoTab) handleCheckUpdateButton() {
	// プログレスダイアログを表示
	t.progressDialog.Show()

	// 非同期で最新バージョンを取得（最小1秒のプログレス表示を保証）
	go func() {
		errg := new(errgroup.Group)

		// 最小1秒の表示を保証
		errg.Go(func() error {
			time.Sleep(1 * time.Second)
			return nil
		})

		// 最新バージョン取得処理
		var latestVersion string
		errg.Go(func() error {
			var err error
			latestVersion, err = fetchLatestVersion()
			return err
		})
		err := errg.Wait()

		fyne.Do(func() {
			t.progressDialog.Hide()

			if err != nil {
				dialog.ShowError(err, t.mainWindow)
				return
			}

			if !version.IsNewer(latestVersion, t.currentVersion) {
				// 最新バージョンを使用中
				dialog.ShowInformation(lang.L("appinfo.noUpdate"), lang.L("appinfo.latestVersion"), t.mainWindow)
				return
			}

			// 新しいバージョンがある場合
			downloadURL := downloadPageBaseURL + "/v" + latestVersion
			dialog.ShowConfirm(
				lang.L("appinfo.updateAvailable"),
				lang.X("appinfo.updateMessage", "Version", M{"Version": latestVersion}),
				func(ok bool) {
					if ok {
						t.openURL(downloadURL)
					}
				},
				t.mainWindow,
			)
		})
	}()
}

// handleLicensesButton はライセンス表示ボタンの処理を行う
func (t *appInfoTab) handleLicensesButton() {
	t.openURL(licensesPageURL)
}

// openURL は指定されたURLを開く
func (t *appInfoTab) openURL(u string) {
	parsedURL, _ := url.Parse(u)
	t.fyneApp.OpenURL(parsedURL)
}

// GitHubReleaseResponse はGitHub APIからのリリース情報
type GitHubReleaseResponse struct {
	TagName string `json:"tag_name"`
}

// fetchLatestVersion はGitHubから最新バージョンを取得する
func fetchLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(githubReleasesAPIURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release GitHubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	// "v" プレフィックスを削除
	ver := strings.TrimPrefix(release.TagName, "v")
	return ver, nil
}
