package totpstore

import (
	"slices"
	"sort"
	"strings"

	"github.com/rs/xid"
)

// Group はTOTPエントリのグループを表す構造体
type Group struct {
	ID    string `json:"id"`    // xid
	Name  string `json:"name"`  // グループ名
	Order int    `json:"order"` // 表示順序
}

// validateGroupName はグループ名をトリムして検証する
// excludeID と同じIDのグループは重複判定から除外する（大文字小文字は区別する）
// 呼び出し側でロックを取得していること
func (s *Store) validateGroupName(name, excludeID string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrGroupNameEmpty
	}
	for _, g := range s.groups {
		if g.ID != excludeID && g.Name == name {
			return "", ErrGroupNameDuplicate
		}
	}
	return name, nil
}

// GetGroups は全てのグループを取得する（Order順でソート済みのコピー）
func (s *Store) GetGroups() []*Group {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := slices.Clone(s.groups)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})
	return result
}

// GetGroup は指定したIDのグループを取得する
func (s *Store) GetGroup(id string) (*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, g := range s.groups {
		if g.ID == id {
			return g, nil
		}
	}
	return nil, ErrGroupNotFound
}

// AddGroup は新しいグループを追加する
func (s *Store) AddGroup(name string) (*Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name, err := s.validateGroupName(name, "")
	if err != nil {
		return nil, err
	}

	// 次のOrder番号を設定
	maxOrder := -1
	for _, g := range s.groups {
		if g.Order > maxOrder {
			maxOrder = g.Order
		}
	}

	group := &Group{ID: xid.New().String(), Name: name, Order: maxOrder + 1}
	s.groups = append(s.groups, group)
	return group, nil
}

// RenameGroup はグループ名を変更する
func (s *Store) RenameGroup(id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := slices.IndexFunc(s.groups, func(g *Group) bool { return g.ID == id })
	if idx < 0 {
		return ErrGroupNotFound
	}

	name, err := s.validateGroupName(name, id)
	if err != nil {
		return err
	}
	s.groups[idx].Name = name
	return nil
}

// DeleteGroup はグループを削除する
// 所属していたエントリは未分類（GroupID == ""）になる
func (s *Store) DeleteGroup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := slices.IndexFunc(s.groups, func(g *Group) bool { return g.ID == id })
	if idx < 0 {
		return ErrGroupNotFound
	}

	for _, e := range s.entries {
		if e.GroupID == id {
			e.GroupID = ""
		}
	}
	s.groups = slices.Delete(s.groups, idx, idx+1)
	return nil
}

// ReorderGroups はグループの順序を更新する
func (s *Store) ReorderGroups(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for order, id := range ids {
		for _, g := range s.groups {
			if g.ID == id {
				g.Order = order
				break
			}
		}
	}
	return nil
}

const (
	// SelectionAll は「すべて」の選択値
	SelectionAll = ""
	// SelectionUngrouped は「未分類」の選択値
	SelectionUngrouped = "__ungrouped__"
)

// groupExists は指定したIDのグループが存在するか返す
func groupExists(groups []*Group, id string) bool {
	return slices.ContainsFunc(groups, func(g *Group) bool { return g.ID == id })
}

// FilterByGroup は選択値に該当するエントリを入力の順序を保って返す（新しいスライス）
// GroupIDが空、または存在しないグループを指すエントリは未分類として扱う
func FilterByGroup(entries []*Entry, groups []*Group, selection string) []*Entry {
	result := make([]*Entry, 0, len(entries))
	for _, e := range entries {
		switch selection {
		case SelectionAll:
			result = append(result, e)
		case SelectionUngrouped:
			if e.GroupID == "" || !groupExists(groups, e.GroupID) {
				result = append(result, e)
			}
		default:
			if e.GroupID == selection && groupExists(groups, selection) {
				result = append(result, e)
			}
		}
	}
	return result
}

// ResolveSelection は選択値を検証し、存在しないグループIDは「すべて」に解決する
func ResolveSelection(groups []*Group, selection string) string {
	if selection == SelectionAll || selection == SelectionUngrouped || groupExists(groups, selection) {
		return selection
	}
	return SelectionAll
}

// GroupIDForSelection は選択値から新規エントリの初期グループIDを返す
// 存在するグループIDならそのID、それ以外は未分類（""）
func GroupIDForSelection(groups []*Group, selection string) string {
	if selection != SelectionAll && selection != SelectionUngrouped && groupExists(groups, selection) {
		return selection
	}
	return ""
}

// MoveEntryToGroup はエントリを指定グループの末尾へ移動する
// groupID が空の場合は未分類へ移動する
func (s *Store) MoveEntryToGroup(entryID, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if groupID != "" && !groupExists(s.groups, groupID) {
		return ErrGroupNotFound
	}

	var target *Entry
	maxOrder := -1
	for _, e := range s.entries {
		if e.ID == entryID {
			target = e
		}
		if e.Order > maxOrder {
			maxOrder = e.Order
		}
	}
	if target == nil {
		return ErrEntryNotFound
	}
	if target.GroupID == groupID {
		return nil
	}

	target.GroupID = groupID
	target.Order = maxOrder + 1
	return nil
}
