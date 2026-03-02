// Package totpstore はTOTPエントリの暗号化保存・読み込みを提供する
package totpstore

import (
	"encoding/base64"
	"encoding/json"
	"slices"
	"sort"
	"sync"

	"github.com/nktmys/winticator/src/pkg/machinekey"
	"github.com/nktmys/winticator/src/usecase/crypto"
	"github.com/nktmys/winticator/src/usecase/preferences"
)

// Store はTOTPエントリの保存・読み込みを管理する
type Store struct {
	prefs   *preferences.Manager
	entries []*Entry
	mu      sync.RWMutex
	loaded  bool
}

// New は新しいStoreインスタンスを作成する
func New(prefs *preferences.Manager) *Store {
	return &Store{
		prefs:   prefs,
		entries: make([]*Entry, 0),
	}
}

// Load は保存されたTOTPエントリを読み込む
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 暗号化されたデータを取得
	encrypted := s.prefs.GetTOTPData()
	if encrypted == "" {
		s.entries = make([]*Entry, 0)
		s.loaded = true
		return nil
	}

	// Base64デコード
	data, err := base64.StdEncoding.DecodeString(encrypted)
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
	var entries []*Entry
	if err := json.Unmarshal(decrypted, &entries); err != nil {
		return err
	}

	s.entries = entries
	s.loaded = true
	return nil
}

// Save は現在のTOTPエントリを保存する
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// JSONエンコード
	data, err := json.Marshal(s.entries)
	if err != nil {
		return err
	}

	// マシンキー取得
	key, err := machinekey.DeriveKey()
	if err != nil {
		return err
	}

	// 暗号化
	encrypted, err := crypto.Encrypt(string(key), data)
	if err != nil {
		return err
	}

	// Base64エンコードして保存
	s.prefs.SetTOTPData(base64.StdEncoding.EncodeToString(encrypted))
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

	s.entries = append(s.entries, entry)
	return nil
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
