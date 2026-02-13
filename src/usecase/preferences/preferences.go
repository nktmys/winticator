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

// Manager はアプリケーション設定を管理する
type Manager struct {
	preferences fyne.Preferences
}

// New は新しいPreferencesインスタンスを作成する
func New(preferences fyne.Preferences) *Manager {
	return &Manager{preferences: preferences}
}

// GetThemeVariant はテーマ設定を取得する
func (m *Manager) GetThemeVariant() fyne.ThemeVariant {
	variant := m.preferences.IntWithFallback(keyThemeVariant, defaultThemeVariant)
	return fyne.ThemeVariant(variant)
}

// SetThemeVariant はテーマ設定を保存する
func (m *Manager) SetThemeVariant(variant fyne.ThemeVariant) {
	m.preferences.SetInt(keyThemeVariant, int(variant))
}

// GetLanguage は言語設定を取得する
// 空文字列の場合はシステムロケールを使用する
func (m *Manager) GetLanguage() string {
	return m.preferences.StringWithFallback(keyLanguage, defaultLanguage)
}

// SetLanguage は言語設定を保存する
func (m *Manager) SetLanguage(language string) {
	m.preferences.SetString(keyLanguage, language)
}

// GetTOTPData はTOTPデータを取得する
func (m *Manager) GetTOTPData() string {
	return m.preferences.StringWithFallback(keyTOTPData, "")
}

// SetTOTPData はTOTPデータを保存する
func (m *Manager) SetTOTPData(data string) {
	m.preferences.SetString(keyTOTPData, data)
}
