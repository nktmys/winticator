package custom

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Theme はテーマバリアントを制御するカスタムテーマ
type Theme struct {
	variant fyne.ThemeVariant
}

func NewTheme(variant fyne.ThemeVariant) *Theme {
	return &Theme{variant: variant}
}

func (t *Theme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, t.variant)
}

func (t *Theme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
