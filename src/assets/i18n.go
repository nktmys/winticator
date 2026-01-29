package assets

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/lang"
)

//go:embed locales/*.json
var embeddedLocales embed.FS

// Locale はロケール情報を表す
type Locale struct {
	Code string // ロケールコード（例: "en", "ja"）
	Name string // 言語名（例: "English", "日本語"）
}

// AvailableLocales は利用可能なロケール一覧を返す
// assetsに組み込まれたJSONファイルからロケールコードと言語名を抽出する
func AvailableLocales() []Locale {
	entries, err := embeddedLocales.ReadDir("locales")
	if err != nil {
		return nil
	}

	locales := make([]Locale, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		code, ok := strings.CutSuffix(fileName, ".json")
		if !ok {
			continue
		}

		// 言語ファイルから言語名を取得
		name := code
		localeFile := filepath.Join("locales", fileName)
		if data, err := embeddedLocales.ReadFile(localeFile); err == nil {
			var translations map[string]string
			if err := json.Unmarshal(data, &translations); err == nil {
				if n, ok := translations["language.name"]; ok {
					name = n
				}
			}
		}

		locales = append(locales, Locale{Code: code, Name: name})
	}
	return locales
}

// InitI18n は翻訳を初期化する
// アプリケーション起動時に呼び出す
func InitI18n() error {
	return lang.AddTranslationsFS(embeddedLocales, "locales")
}

// InitI18nWithLocale は指定されたロケールで翻訳を初期化する
// localeが空の場合はシステムロケールを使用する
func InitI18nWithLocale(locale string) error {
	// まず全ての翻訳を読み込む
	if err := lang.AddTranslationsFS(embeddedLocales, "locales"); err != nil {
		return err
	}

	// ロケールが指定されていない場合はシステムデフォルトを使用
	if locale == "" {
		return nil
	}

	// 指定されたロケールの翻訳を読み込む
	localeFile := filepath.Join("locales", locale+".json")
	data, err := embeddedLocales.ReadFile(localeFile)
	if err != nil {
		return err
	}

	// Fyneはシステムロケールに基づいて翻訳を選択するため、
	// 指定されたロケールの翻訳をシステムロケールにも適用する
	systemLocale := lang.SystemLocale()
	return lang.AddTranslationsForLocale(data, systemLocale)
}
