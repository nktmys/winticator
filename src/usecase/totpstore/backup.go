package totpstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"sort"
	"strings"

	"github.com/rs/xid"
)

// backupVersion は現在のバックアップ形式のバージョン
const backupVersion = 2

// Backup はバックアップファイルの内容を表す構造体
// 旧形式（エントリ配列のみ）は Version 1 として扱う
type Backup struct {
	Version int      `json:"version"`
	Groups  []*Group `json:"groups"`
	Entries []*Entry `json:"entries"`
}

// MarshalBackup は現在のエントリとグループをバックアップ形式のJSONにする
func (s *Store) MarshalBackup() ([]byte, error) {
	b := &Backup{
		Version: backupVersion,
		Groups:  s.GetGroups(),
		Entries: s.GetAll(),
	}
	return json.Marshal(b)
}

// ParseBackup はバックアップのJSONをパースする
// 先頭が '[' なら旧形式（グループなし）として扱い、GroupID は空にする
func ParseBackup(data []byte) (*Backup, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var entries []*Entry
		if err := json.Unmarshal(trimmed, &entries); err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e != nil {
				e.GroupID = ""
			}
		}
		return &Backup{Version: 1, Groups: []*Group{}, Entries: entries}, nil
	}

	var b Backup
	if err := json.Unmarshal(trimmed, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// normalizeBackupGroups はバックアップのグループを検証・正規化したコピーをOrder順で返す
// nil、IDが空、IDが重複（先に現れたものを優先）、名前（トリム後）が空のグループは除外する
func normalizeBackupGroups(groups []*Group) []*Group {
	result := make([]*Group, 0, len(groups))
	seen := make(map[string]bool, len(groups))
	for _, g := range groups {
		if g == nil || g.ID == "" || seen[g.ID] {
			continue
		}
		name := strings.TrimSpace(g.Name)
		if name == "" {
			continue
		}
		seen[g.ID] = true
		result = append(result, &Group{ID: g.ID, Name: name, Order: g.Order})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

// ImportBackup はバックアップを取り込む
// replace が true の場合は既存のエントリとグループをすべて置き換える
// （名前が重複するグループは先頭（Order順）のものを残し、後続の所属エントリは先頭のグループへ入れる）
// false の場合はグループを名前（トリム後・大文字小文字を区別）で突き合わせて統合し、エントリを末尾に追加する
// バックアップに存在しないグループを指すエントリは未分類になる
func (s *Store) ImportBackup(b *Backup, replace bool) error {
	groups := normalizeBackupGroups(b.Groups)

	s.mu.Lock()
	defer s.mu.Unlock()

	// バックアップのグループID → 取り込み後のグループID（空IDは決して登録しない）
	idMap := make(map[string]string)

	// 失敗時にストアを変更しないよう、ローカルの作業用スライスに構築して最後に反映する
	// （validateGroupName は s.groups を参照するため、作業中だけ差し替える）
	origEntries, origGroups := s.entries, s.groups
	newEntries, newGroups := origEntries, origGroups
	if replace {
		newEntries = make([]*Entry, 0, len(b.Entries))
		newGroups = make([]*Group, 0, len(groups))
	}
	newEntries = slices.Clone(newEntries)
	newGroups = slices.Clone(newGroups)
	s.groups = newGroups
	ok := false
	defer func() {
		if !ok {
			s.entries, s.groups = origEntries, origGroups
		}
	}()
	for _, g := range groups {
		_, err := s.validateGroupName(g.Name, "")
		if errors.Is(err, ErrGroupNameDuplicate) {
			if idx := slices.IndexFunc(s.groups, func(x *Group) bool { return x.Name == g.Name }); idx >= 0 {
				idMap[g.ID] = s.groups[idx].ID
			}
			continue
		}
		if err != nil {
			return err
		}
		// IDは常に新しく採番し、既存データとの重複を防ぐ
		id := xid.New().String()
		s.groups = append(s.groups, &Group{ID: id, Name: g.Name, Order: s.nextGroupOrder()})
		idMap[g.ID] = id
	}

	maxOrder := -1
	for _, e := range newEntries {
		if e.Order > maxOrder {
			maxOrder = e.Order
		}
	}
	for _, e := range b.Entries {
		if e == nil {
			continue
		}
		ne := *e
		ne.ID = xid.New().String()    // 既存エントリとのID重複を防ぐため常に新しく採番する
		ne.GroupID = idMap[e.GroupID] // 対応がなければ空（未分類）
		maxOrder++
		ne.Order = maxOrder
		newEntries = append(newEntries, &ne)
	}
	s.entries = newEntries
	ok = true
	return nil
}

// nextGroupOrder は次のグループOrder番号を返す
// 呼び出し側でロックを取得していること
func (s *Store) nextGroupOrder() int {
	maxOrder := -1
	for _, g := range s.groups {
		if g.Order > maxOrder {
			maxOrder = g.Order
		}
	}
	return maxOrder + 1
}
