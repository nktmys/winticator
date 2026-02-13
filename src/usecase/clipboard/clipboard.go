package clipboard

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// Manager はクリップボードのコピーと自動クリアを管理する
type Manager struct {
	clipboard     fyne.Clipboard
	timer         *time.Timer
	copiedContent string
	mu            sync.Mutex
}

// New は新しいManagerを作成する
func New(clipboard fyne.Clipboard) *Manager {
	return &Manager{clipboard: clipboard}
}

// Copy はコードをクリップボードにコピーし、delay後に自動クリアをスケジュールする
func (m *Manager) Copy(content string, delay time.Duration) {
	m.Clear()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.clipboard.SetContent(content)
	m.copiedContent = content
	m.timer = time.AfterFunc(delay, func() {
		fyne.Do(func() {
			m.Clear()
		})
	})
}

// Clear はコピーしたコードをクリップボードから即座にクリアする
func (m *Manager) Clear() {
	m.mu.Lock()
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	content := m.copiedContent
	m.copiedContent = ""
	m.mu.Unlock()

	if content != "" && m.clipboard.Content() == content {
		m.clipboard.SetContent("")
	}
}
