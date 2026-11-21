package totpstore

import (
	"testing"
	"time"

	"github.com/nktmys/winticator/src/pkg/machinekey"
	"github.com/nktmys/winticator/src/usecase/preferences"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPreferences はテスト用のモックPreferences
type mockPreferences struct {
	data map[string]any
}

func newMockPreferences() *mockPreferences {
	return &mockPreferences{data: make(map[string]any)}
}

func (m *mockPreferences) Bool(key string) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return false
}

func (m *mockPreferences) BoolWithFallback(key string, fallback bool) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetBool(key string, value bool) {
	m.data[key] = value
}

func (m *mockPreferences) Float(key string) float64 {
	if v, ok := m.data[key].(float64); ok {
		return v
	}
	return 0
}

func (m *mockPreferences) FloatWithFallback(key string, fallback float64) float64 {
	if v, ok := m.data[key].(float64); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetFloat(key string, value float64) {
	m.data[key] = value
}

func (m *mockPreferences) Int(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockPreferences) IntWithFallback(key string, fallback int) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetInt(key string, value int) {
	m.data[key] = value
}

func (m *mockPreferences) String(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockPreferences) StringWithFallback(key string, fallback string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetString(key string, value string) {
	m.data[key] = value
}

func (m *mockPreferences) StringList(key string) []string {
	if v, ok := m.data[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mockPreferences) StringListWithFallback(key string, fallback []string) []string {
	if v, ok := m.data[key].([]string); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetStringList(key string, value []string) {
	m.data[key] = value
}

func (m *mockPreferences) RemoveValue(key string) {
	delete(m.data, key)
}

func (m *mockPreferences) AddChangeListener(func()) {}

func (m *mockPreferences) ChangeListeners() []func() { return nil }

func (m *mockPreferences) BoolList(key string) []bool {
	if v, ok := m.data[key].([]bool); ok {
		return v
	}
	return nil
}

func (m *mockPreferences) BoolListWithFallback(key string, fallback []bool) []bool {
	if v, ok := m.data[key].([]bool); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetBoolList(key string, value []bool) {
	m.data[key] = value
}

func (m *mockPreferences) IntList(key string) []int {
	if v, ok := m.data[key].([]int); ok {
		return v
	}
	return nil
}

func (m *mockPreferences) IntListWithFallback(key string, fallback []int) []int {
	if v, ok := m.data[key].([]int); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetIntList(key string, value []int) {
	m.data[key] = value
}

func (m *mockPreferences) FloatList(key string) []float64 {
	if v, ok := m.data[key].([]float64); ok {
		return v
	}
	return nil
}

func (m *mockPreferences) FloatListWithFallback(key string, fallback []float64) []float64 {
	if v, ok := m.data[key].([]float64); ok {
		return v
	}
	return fallback
}

func (m *mockPreferences) SetFloatList(key string, value []float64) {
	m.data[key] = value
}

func TestStore_AddAndGet(t *testing.T) {
	machinekey.ResetCache()

	prefs := preferences.New(newMockPreferences())
	store := New(prefs)

	// 初期ロード
	err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, 0, store.Count())

	// エントリ追加
	entry := &Entry{
		ID:        "test-id-1",
		Issuer:    "Google",
		Account:   "user@gmail.com",
		Secret:    "JBSWY3DPEHPK3PXP",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		CreatedAt: time.Now(),
	}

	err = store.Add(entry)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Count())
	assert.Equal(t, 0, entry.Order) // 最初のエントリはOrder=0

	// 取得
	got, err := store.Get("test-id-1")
	require.NoError(t, err)
	assert.Equal(t, "Google", got.Issuer)
	assert.Equal(t, "user@gmail.com", got.Account)
}

func TestStore_SaveAndLoad(t *testing.T) {
	machinekey.ResetCache()

	mock := newMockPreferences()
	prefs := preferences.New(mock)
	store := New(prefs)

	err := store.Load()
	require.NoError(t, err)

	// エントリ追加
	entry := &Entry{
		ID:        "test-id-1",
		Issuer:    "GitHub",
		Account:   "myaccount",
		Secret:    "ABCDEFGH",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		CreatedAt: time.Now(),
	}
	err = store.Add(entry)
	require.NoError(t, err)

	// 保存
	err = store.Save()
	require.NoError(t, err)

	// 新しいストアで読み込み
	prefs2 := preferences.New(mock)
	store2 := New(prefs2)
	err = store2.Load()
	require.NoError(t, err)

	assert.Equal(t, 1, store2.Count())

	got, err := store2.Get("test-id-1")
	require.NoError(t, err)
	assert.Equal(t, "GitHub", got.Issuer)
	assert.Equal(t, "myaccount", got.Account)
	assert.Equal(t, "ABCDEFGH", got.Secret)
}

func TestStore_Update(t *testing.T) {
	machinekey.ResetCache()

	prefs := preferences.New(newMockPreferences())
	store := New(prefs)
	err := store.Load()
	require.NoError(t, err)

	entry := &Entry{
		ID:      "test-id-1",
		Issuer:  "OldName",
		Account: "user",
		Secret:  "SECRET",
	}
	err = store.Add(entry)
	require.NoError(t, err)

	// 更新
	entry.Issuer = "NewName"
	err = store.Update(entry)
	require.NoError(t, err)

	got, err := store.Get("test-id-1")
	require.NoError(t, err)
	assert.Equal(t, "NewName", got.Issuer)
}

func TestStore_Delete(t *testing.T) {
	machinekey.ResetCache()

	prefs := preferences.New(newMockPreferences())
	store := New(prefs)
	err := store.Load()
	require.NoError(t, err)

	entry := &Entry{
		ID:      "test-id-1",
		Issuer:  "Test",
		Account: "user",
		Secret:  "SECRET",
	}
	err = store.Add(entry)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Count())

	// 削除
	err = store.Delete("test-id-1")
	require.NoError(t, err)
	assert.Equal(t, 0, store.Count())

	// 存在しないID
	err = store.Delete("nonexistent")
	assert.ErrorIs(t, err, ErrEntryNotFound)
}

func TestStore_Reorder(t *testing.T) {
	machinekey.ResetCache()

	prefs := preferences.New(newMockPreferences())
	store := New(prefs)
	err := store.Load()
	require.NoError(t, err)

	// 3つのエントリを追加
	for i := 1; i <= 3; i++ {
		err := store.Add(&Entry{
			ID:      "id-" + string(rune('0'+i)),
			Issuer:  "Test",
			Account: "user",
			Secret:  "SECRET",
		})
		require.NoError(t, err)
	}

	// 順序を変更: 3, 1, 2
	err = store.Reorder([]string{"id-3", "id-1", "id-2"})
	require.NoError(t, err)

	entries := store.GetAll()
	assert.Equal(t, "id-3", entries[0].ID)
	assert.Equal(t, "id-1", entries[1].ID)
	assert.Equal(t, "id-2", entries[2].ID)
}

func TestStore_GetAll_Sorted(t *testing.T) {
	machinekey.ResetCache()

	prefs := preferences.New(newMockPreferences())
	store := New(prefs)
	err := store.Load()
	require.NoError(t, err)

	// 順序を明示的に設定して追加
	entries := []*Entry{
		{ID: "id-a", Issuer: "A", Order: 2},
		{ID: "id-b", Issuer: "B", Order: 0},
		{ID: "id-c", Issuer: "C", Order: 1},
	}

	store.entries = append(store.entries, entries...)

	// GetAllはOrder順でソートされるべき
	result := store.GetAll()
	assert.Equal(t, "id-b", result[0].ID) // Order=0
	assert.Equal(t, "id-c", result[1].ID) // Order=1
	assert.Equal(t, "id-a", result[2].ID) // Order=2
}

func TestStore_EmptyLoad(t *testing.T) {
	machinekey.ResetCache()

	prefs := preferences.New(newMockPreferences())
	store := New(prefs)

	err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, 0, store.Count())
}

func TestStore_GroupLoadFailureDoesNotOverwriteGroups(t *testing.T) {
	machinekey.ResetCache()

	mock := newMockPreferences()
	store := New(preferences.New(mock))
	require.NoError(t, store.Load())
	_, err := store.AddGroup("仕事")
	require.NoError(t, err)
	require.NoError(t, store.Add(&Entry{ID: "e1", Issuer: "GitHub", Secret: "ABCDEFGH", Digits: 6, Period: 30}))
	require.NoError(t, store.Save())

	// グループデータを壊す
	mock.SetString("totpGroupsData", "corrupted")

	store2 := New(preferences.New(mock))
	require.Error(t, store2.Load())
	assert.Equal(t, 1, store2.Count(), "エントリは保持される")

	// 保存してもグループデータは上書きされない
	require.NoError(t, store2.Save())
	assert.Equal(t, "corrupted", mock.String("totpGroupsData"))
}

func TestStore_Add_AssignsUniqueID(t *testing.T) {
	store := New(preferences.New(newMockPreferences()))

	first := &Entry{ID: "dup", Issuer: "A", Secret: "JBSWY3DPEHPK3PXP"}
	require.NoError(t, store.Add(first))
	assert.Equal(t, "dup", first.ID)

	// 既存と重複するIDは新しく採番される
	second := &Entry{ID: "dup", Issuer: "B", Secret: "JBSWY3DPEHPK3PXP"}
	require.NoError(t, store.Add(second))
	assert.NotEqual(t, "dup", second.ID)
	assert.NotEmpty(t, second.ID)

	// 空のIDも採番される
	third := &Entry{Issuer: "C", Secret: "JBSWY3DPEHPK3PXP"}
	require.NoError(t, store.Add(third))
	assert.NotEmpty(t, third.ID)
	assert.NotEqual(t, second.ID, third.ID)
}

func TestStore_Load_RepairsDuplicateIDs(t *testing.T) {
	machinekey.ResetCache()

	mock := newMockPreferences()
	prefs := preferences.New(mock)

	// 過去のインポートで重複IDが保存されてしまったデータを用意する
	entriesData, err := encryptJSON([]*Entry{
		{ID: "e1", Issuer: "A", Secret: "JBSWY3DPEHPK3PXP", Order: 0},
		{ID: "e1", Issuer: "B", Secret: "JBSWY3DPEHPK3PXP", Order: 1, GroupID: "g1"},
		{ID: "", Issuer: "C", Secret: "JBSWY3DPEHPK3PXP", Order: 2},
	})
	require.NoError(t, err)
	prefs.SetTOTPData(entriesData)
	groupsData, err := encryptJSON([]*Group{
		{ID: "g1", Name: "仕事", Order: 0},
		{ID: "g1", Name: "個人", Order: 1},
		{ID: SelectionUngrouped, Name: "予約", Order: 2},
	})
	require.NoError(t, err)
	prefs.SetTOTPGroupsData(groupsData)

	store := New(prefs)
	require.NoError(t, store.Load())

	entries := store.GetAll()
	require.Len(t, entries, 3)
	assert.Equal(t, "e1", entries[0].ID)
	assert.NotEqual(t, "e1", entries[1].ID)
	assert.NotEmpty(t, entries[2].ID)
	assert.NotEqual(t, entries[1].ID, entries[2].ID)
	assert.Equal(t, "g1", entries[1].GroupID)

	groups := store.GetGroups()
	require.Len(t, groups, 3)
	assert.Equal(t, "g1", groups[0].ID)
	assert.NotEqual(t, "g1", groups[1].ID)
	assert.NotEmpty(t, groups[1].ID)
	assert.NotEqual(t, SelectionUngrouped, groups[2].ID)

	// 修復後のIDで削除すると、その1件だけが消える
	require.NoError(t, store.Delete(entries[1].ID))
	assert.Equal(t, 2, store.Count())
	_, err = store.Get("e1")
	require.NoError(t, err)
}
