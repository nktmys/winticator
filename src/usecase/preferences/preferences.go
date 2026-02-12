package preferences

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// 設定キー（非公開）
const (
	keyThemeVariant = "themeVariant"
	keyLanguage     = "language"
	keyTOTPData     = "totpData"
)

// デフォルト値（非公開）
var (
	defaultThemeVariant = int(theme.VariantLight) // Lightをデフォルトに
	defaultLanguage     = ""                      // 空文字列はシステムロケールを使用
)

// Preferences はアプリケーション設定を管理する
type Preferences struct {
	fyne fyne.Preferences
}

// New は新しいPreferencesインスタンスを作成する
func New(p fyne.Preferences) *Preferences {
	return &Preferences{fyne: p}
}

// GetThemeVariant はテーマ設定を取得する
func (p *Preferences) GetThemeVariant() fyne.ThemeVariant {
	variant := p.fyne.IntWithFallback(keyThemeVariant, defaultThemeVariant)
	return fyne.ThemeVariant(variant)
}

// SetThemeVariant はテーマ設定を保存する
func (p *Preferences) SetThemeVariant(variant fyne.ThemeVariant) {
	p.fyne.SetInt(keyThemeVariant, int(variant))
}

// GetLanguage は言語設定を取得する
// 空文字列の場合はシステムロケールを使用する
func (p *Preferences) GetLanguage() string {
	return p.fyne.StringWithFallback(keyLanguage, defaultLanguage)
}

// SetLanguage は言語設定を保存する
func (p *Preferences) SetLanguage(language string) {
	p.fyne.SetString(keyLanguage, language)
}

// GetTOTPData はTOTPデータを取得する
func (p *Preferences) GetTOTPData() string {
	return p.fyne.StringWithFallback(keyTOTPData, "")
}

// SetTOTPData はTOTPデータを保存する
func (p *Preferences) SetTOTPData(data string) {
	p.fyne.SetString(keyTOTPData, data)
}
