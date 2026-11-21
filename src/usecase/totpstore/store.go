package totpstore

import (
	"encoding/base64"
	"encoding/json"
	"slices"
	"sort"
	"sync"

	"github.com/rs/xid"

	"github.com/nktmys/winticator/src/pkg/machinekey"
	"github.com/nktmys/winticator/src/usecase/crypto"
	"github.com/nktmys/winticator/src/usecase/preferences"
)

// Store はTOTPエントリの保存・読み込みを管理する
type Store struct {
	prefs   *preferences.Manager
	entries []*Entry
	groups  []*Group
	mu      sync.RWMutex
	loaded  bool
	// groupsLoadFailed はグループの読み込みに失敗した状態を表す
	// true の間は保存済みのグループデータを上書きしない
	groupsLoadFailed bool
}

// New は新しいStoreインスタンスを作成する
func New(prefs *preferences.Manager) *Store {
	return &Store{
		prefs:   prefs,
		entries: make([]*Entry, 0),
		groups:  make([]*Group, 0),
	}
}

// encryptJSON は値をJSON化し、マシンキーで暗号化してBase64文字列にする
func encryptJSON(v any) (string, error) {
	// JSONエンコード
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	// マシンキー取得
	key, err := machinekey.DeriveKey()
	if err != nil {
		return "", err
	}

	// 暗号化
	encrypted, err := crypto.Encrypt(string(key), data)
	if err != nil {
		return "", err
	}

	// Base64エンコード
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// decryptJSON はBase64文字列を復号してJSONとしてvにデコードする
func decryptJSON(encoded string, v any) error {
	// Base64デコード
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	// マシンキー取得
	key, err := machinekey.DeriveKey()
	if err != nil {
		return err
	}

	// 復号
	decrypted, err := crypto.Decrypt(string(key), data)
	if err != nil {
		return err
	}

	// JSONデコード
	return json.Unmarshal(decrypted, v)
}

// Load は保存されたTOTPエントリとグループを読み込む
// グループの読み込みに失敗した場合も、エントリは保持したままエラーを返す
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// エントリを先に読み込む
	s.entries = make([]*Entry, 0)
	s.groups = make([]*Group, 0)
	s.groupsLoadFailed = false
	if encrypted := s.prefs.GetTOTPData(); encrypted != "" {
		var entries []*Entry
		if err := decryptJSON(encrypted, &entries); err != nil {
			return err
		}
		s.entries = entries
	}
	s.loaded = true
	s.ensureUniqueIDs()

	// グループを読み込む（キーが空ならグループなし）
	if encrypted := s.prefs.GetTOTPGroupsData(); encrypted != "" {
		var groups []*Group
		if err := decryptJSON(encrypted, &groups); err != nil {
			s.groupsLoadFailed = true
			return err
		}
		s.groups = groups
	}
	s.ensureUniqueIDs()
	return nil
}

// Save は現在のTOTPエントリとグループを保存する
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := encryptJSON(s.entries)
	if err != nil {
		return err
	}
	var groups string
	if !s.groupsLoadFailed {
		groups, err = encryptJSON(s.groups)
		if err != nil {
			return err
		}
	}

	s.prefs.SetTOTPData(entries)
	// グループの読み込みに失敗している間は、保存済みのグループデータを壊さないよう書き込まない
	if !s.groupsLoadFailed {
		s.prefs.SetTOTPGroupsData(groups)
	}
	return nil
}

// GetAll は全てのTOTPエントリを取得する（Order順でソート済み）
func (s *Store) GetAll() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// コピーを返す
	result := slices.Clone(s.entries)

	// Order順でソート
	sort.Slice(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})

	return result
}

// Get は指定したIDのエントリを取得する
func (s *Store) Get(id string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, entry := range s.entries {
		if entry.ID == id {
			return entry, nil
		}
	}
	return nil, ErrEntryNotFound
}

// Add は新しいエントリを追加する
func (s *Store) Add(entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 次のOrder番号を設定
	maxOrder := -1
	for _, e := range s.entries {
		if e.Order > maxOrder {
			maxOrder = e.Order
		}
	}
	entry.Order = maxOrder + 1

	// IDが空または既存と重複する場合は新しく採番する
	if entry.ID == "" || slices.ContainsFunc(s.entries, func(e *Entry) bool { return e.ID == entry.ID }) {
		entry.ID = xid.New().String()
	}

	s.entries = append(s.entries, entry)
	return nil
}

// ensureUniqueIDs は空または重複したエントリ・グループのIDを新しく採番する
// 過去のインポートで重複IDが保存されたデータを修復するために使う
// 呼び出し側でロックを取得していること
func (s *Store) ensureUniqueIDs() {
	s.entries = slices.DeleteFunc(s.entries, func(e *Entry) bool { return e == nil })
	seen := make(map[string]bool, len(s.entries))
	for _, e := range s.entries {
		if e.ID == "" || seen[e.ID] {
			e.ID = xid.New().String()
		}
		seen[e.ID] = true
	}

	s.groups = slices.DeleteFunc(s.groups, func(g *Group) bool { return g == nil })
	seenGroups := make(map[string]bool, len(s.groups))
	for _, g := range s.groups {
		// 「未分類」の選択値と同じIDは選択と衝突するため振り直す
		if g.ID == "" || g.ID == SelectionUngrouped || seenGroups[g.ID] {
			g.ID = xid.New().String()
		}
		seenGroups[g.ID] = true
	}
}

// Update は既存のエントリを更新する
func (s *Store) Update(entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.entries {
		if e.ID == entry.ID {
			s.entries[i] = entry
			return nil
		}
	}
	return ErrEntryNotFound
}

// Delete は指定したIDのエントリを削除する
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, entry := range s.entries {
		if entry.ID == id {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			return nil
		}
	}
	return ErrEntryNotFound
}

// Reorder はエントリの順序を更新する
func (s *Store) Reorder(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// IDからエントリを検索してOrder更新
	for order, id := range ids {
		for _, entry := range s.entries {
			if entry.ID == id {
				entry.Order = order
				break
			}
		}
	}
	return nil
}

// Count はエントリ数を返す
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
