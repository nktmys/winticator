package components

import (
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SearchEntry は検索用のエントリウィジェット。
// テキストが入力されるとクリアボタンを表示し、OnChanged コールバックを呼び出す。
type SearchEntry struct {
	widget.Entry

	OnChanged func(query string)
	clearBtn  *widget.Button
}

func NewSearchEntry(placeholder string) *SearchEntry {
	s := &SearchEntry{}
	s.PlaceHolder = placeholder
	s.clearBtn = widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		s.Clear()
	})
	s.ExtendBaseWidget(s)
	s.Entry.OnChanged = func(text string) {
		s.updateActionItem()
		if s.OnChanged != nil {
			s.OnChanged(text)
		}
	}
	return s
}

// Clear はテキストをクリアする
func (s *SearchEntry) Clear() {
	s.SetText("")
}

// updateActionItem はテキストの有無に応じてクリアボタンの表示を切り替える
func (s *SearchEntry) updateActionItem() {
	if s.Text != "" {
		s.ActionItem = s.clearBtn
	} else {
		s.ActionItem = nil
	}
	s.Refresh()
}
