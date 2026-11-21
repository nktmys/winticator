package totpstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBackupTestStore() *Store {
	return &Store{entries: make([]*Entry, 0), groups: make([]*Group, 0)}
}

func TestStore_MarshalBackup_RoundTrip(t *testing.T) {
	s := newBackupTestStore()
	g1, err := s.AddGroup("仕事")
	require.NoError(t, err)
	g2, err := s.AddGroup("個人")
	require.NoError(t, err)

	e1 := NewEntry("A", "a", "JBSWY3DPEHPK3PXP")
	e1.GroupID = g1.ID
	e2 := NewEntry("B", "b", "JBSWY3DPEHPK3PXP")
	e2.GroupID = g2.ID
	e3 := NewEntry("C", "c", "JBSWY3DPEHPK3PXP")
	require.NoError(t, s.Add(e1))
	require.NoError(t, s.Add(e2))
	require.NoError(t, s.Add(e3))

	data, err := s.MarshalBackup()
	require.NoError(t, err)

	b, err := ParseBackup(data)
	require.NoError(t, err)
	assert.Equal(t, 2, b.Version)
	require.Len(t, b.Groups, 2)
	assert.Equal(t, g1.ID, b.Groups[0].ID)
	assert.Equal(t, "個人", b.Groups[1].Name)
	require.Len(t, b.Entries, 3)
	assert.Equal(t, e1.ID, b.Entries[0].ID)
	assert.Equal(t, g1.ID, b.Entries[0].GroupID)
	assert.Equal(t, g2.ID, b.Entries[1].GroupID)
	assert.Empty(t, b.Entries[2].GroupID)
}

func TestParseBackup_Legacy(t *testing.T) {
	data := []byte(` [{"id":"x","issuer":"G","account":"a","secret":"JBSWY3DPEHPK3PXP","group_id":"g1"}]`)
	b, err := ParseBackup(data)
	require.NoError(t, err)
	assert.Equal(t, 1, b.Version)
	assert.Empty(t, b.Groups)
	require.Len(t, b.Entries, 1)
	assert.Empty(t, b.Entries[0].GroupID)
}

func TestParseBackup_Invalid(t *testing.T) {
	_, err := ParseBackup([]byte("not json"))
	assert.Error(t, err)
}

func TestStore_ImportBackup_Replace(t *testing.T) {
	s := newBackupTestStore()
	_, err := s.AddGroup("旧")
	require.NoError(t, err)
	require.NoError(t, s.Add(NewEntry("Old", "o", "JBSWY3DPEHPK3PXP")))

	b := &Backup{
		Version: 2,
		Groups: []*Group{
			{ID: "B1", Name: "仕事", Order: 5},
			{ID: "B2", Name: "個人", Order: 9},
		},
		Entries: []*Entry{
			{ID: "e1", Issuer: "A", Secret: "JBSWY3DPEHPK3PXP", GroupID: "B2"},
			{ID: "e2", Issuer: "B", Secret: "JBSWY3DPEHPK3PXP", GroupID: "missing"},
		},
	}
	require.NoError(t, s.ImportBackup(b, true))

	// IDはバックアップの値を使わず新しく採番される
	groups := s.GetGroups()
	require.Len(t, groups, 2)
	assert.NotEqual(t, "B1", groups[0].ID)
	assert.Equal(t, "仕事", groups[0].Name)
	assert.Equal(t, 0, groups[0].Order)
	assert.NotEqual(t, "B2", groups[1].ID)
	assert.Equal(t, "個人", groups[1].Name)
	assert.Equal(t, 1, groups[1].Order)

	entries := s.GetAll()
	require.Len(t, entries, 2)
	assert.NotEqual(t, "e1", entries[0].ID)
	assert.Equal(t, "A", entries[0].Issuer)
	assert.Equal(t, groups[1].ID, entries[0].GroupID)
	assert.NotEqual(t, "e2", entries[1].ID)
	assert.Empty(t, entries[1].GroupID)
}

func TestStore_ImportBackup_MergeSameBackupTwice_UniqueIDs(t *testing.T) {
	s := newBackupTestStore()
	work, err := s.AddGroup("仕事")
	require.NoError(t, err)
	require.NoError(t, s.Add(NewEntry("GitHub", "me", "JBSWY3DPEHPK3PXP")))

	// 同じバックアップを統合すると同名エントリが2件になるが、IDは別になる
	data, err := s.MarshalBackup()
	require.NoError(t, err)
	b, err := ParseBackup(data)
	require.NoError(t, err)
	require.NoError(t, s.ImportBackup(b, false))

	entries := s.GetAll()
	require.Len(t, entries, 2)
	require.NotEqual(t, entries[0].ID, entries[1].ID)

	// 2件目を別グループへ移動して削除すると、2件目だけが消える
	second := entries[1].ID
	require.NoError(t, s.MoveEntryToGroup(second, work.ID))
	require.NoError(t, s.Delete(second))

	remaining := s.GetAll()
	require.Len(t, remaining, 1)
	assert.Equal(t, entries[0].ID, remaining[0].ID)
	assert.Empty(t, remaining[0].GroupID)
}

func TestStore_ImportBackup_MergeByName(t *testing.T) {
	s := newBackupTestStore()
	local, err := s.AddGroup("仕事")
	require.NoError(t, err)
	require.NoError(t, s.Add(NewEntry("Old", "o", "JBSWY3DPEHPK3PXP")))

	b := &Backup{
		Version: 2,
		Groups: []*Group{
			{ID: "B1", Name: "仕事", Order: 0},
			{ID: "B2", Name: "個人", Order: 1},
		},
		Entries: []*Entry{
			{ID: "e1", Issuer: "A", Secret: "JBSWY3DPEHPK3PXP", GroupID: "B1"},
			{ID: "e2", Issuer: "B", Secret: "JBSWY3DPEHPK3PXP", GroupID: "B2"},
		},
	}
	require.NoError(t, s.ImportBackup(b, false))

	groups := s.GetGroups()
	require.Len(t, groups, 2)
	assert.Equal(t, local.ID, groups[0].ID)
	assert.Equal(t, "個人", groups[1].Name)

	entries := s.GetAll()
	require.Len(t, entries, 3)
	assert.Equal(t, "Old", entries[0].Issuer)
	assert.Equal(t, local.ID, entries[1].GroupID)
	assert.Equal(t, groups[1].ID, entries[2].GroupID)
	assert.Less(t, entries[1].Order, entries[2].Order)
}

func TestStore_ImportBackup_LegacyMergeUngroups(t *testing.T) {
	s := newBackupTestStore()
	s.groups = append(s.groups, &Group{ID: "g1", Name: "仕事"})

	b, err := ParseBackup([]byte(`[{"id":"x","issuer":"G","account":"a","secret":"JBSWY3DPEHPK3PXP","group_id":"g1"}]`))
	require.NoError(t, err)
	require.NoError(t, s.ImportBackup(b, false))

	entries := s.GetAll()
	require.Len(t, entries, 1)
	assert.Empty(t, entries[0].GroupID)
}

func TestStore_ImportBackup_NullGroupAndEntry(t *testing.T) {
	for _, replace := range []bool{true, false} {
		s := newBackupTestStore()
		b, err := ParseBackup([]byte(`{"version":2,"groups":[null,{"id":"B1","name":"仕事","order":0}],"entries":[null,{"id":"e1","secret":"JBSWY3DPEHPK3PXP","group_id":"B1"}]}`))
		require.NoError(t, err)
		require.NoError(t, s.ImportBackup(b, replace))
		groups := s.GetGroups()
		require.Len(t, groups, 1)
		entries := s.GetAll()
		require.Len(t, entries, 1)
		assert.Equal(t, groups[0].ID, entries[0].GroupID)
	}
}

func TestStore_ImportBackup_EmptyIDGroupDoesNotCaptureUngrouped(t *testing.T) {
	for _, replace := range []bool{true, false} {
		s := newBackupTestStore()
		b := &Backup{
			Version: 2,
			Groups:  []*Group{{ID: "", Name: "空ID"}},
			Entries: []*Entry{{ID: "e1", Secret: "JBSWY3DPEHPK3PXP"}},
		}
		require.NoError(t, s.ImportBackup(b, replace))
		assert.Empty(t, s.GetGroups())
		require.Len(t, s.GetAll(), 1)
		assert.Empty(t, s.GetAll()[0].GroupID)
	}
}

func TestStore_ImportBackup_DuplicateBackupGroupIDs(t *testing.T) {
	b := &Backup{
		Version: 2,
		Groups: []*Group{
			{ID: "B1", Name: "仕事", Order: 0},
			{ID: "B1", Name: "個人", Order: 1},
		},
		Entries: []*Entry{{ID: "e1", Secret: "JBSWY3DPEHPK3PXP", GroupID: "B1"}},
	}
	for _, replace := range []bool{true, false} {
		s := newBackupTestStore()
		require.NoError(t, s.ImportBackup(b, replace))
		groups := s.GetGroups()
		require.Len(t, groups, 1)
		assert.Equal(t, "仕事", groups[0].Name)
		assert.Equal(t, groups[0].ID, s.GetAll()[0].GroupID)
	}
}

func TestStore_ImportBackup_ReplaceDuplicateNames(t *testing.T) {
	s := newBackupTestStore()
	b := &Backup{
		Version: 2,
		Groups: []*Group{
			{ID: "B1", Name: "仕事", Order: 0},
			{ID: "B2", Name: " 仕事 ", Order: 1},
		},
		Entries: []*Entry{{ID: "e1", Secret: "JBSWY3DPEHPK3PXP", GroupID: "B2"}},
	}
	require.NoError(t, s.ImportBackup(b, true))
	groups := s.GetGroups()
	require.Len(t, groups, 1)
	assert.Equal(t, groups[0].ID, s.GetAll()[0].GroupID)
}

func TestStore_ImportBackup_MergeNameMatching(t *testing.T) {
	s := newBackupTestStore()
	work, err := s.AddGroup("Work")
	require.NoError(t, err)
	b := &Backup{
		Version: 2,
		Groups: []*Group{
			{ID: "B1", Name: "  Work  ", Order: 0},
			{ID: "B2", Name: "work", Order: 1},
			{ID: "B3", Name: "新規", Order: 2},
		},
		Entries: []*Entry{
			{ID: "e1", Secret: "JBSWY3DPEHPK3PXP", GroupID: "B1"},
			{ID: "e2", Secret: "JBSWY3DPEHPK3PXP", GroupID: "B2"},
		},
	}
	require.NoError(t, s.ImportBackup(b, false))
	groups := s.GetGroups()
	require.Len(t, groups, 3)
	assert.Equal(t, work.ID, groups[0].ID)
	assert.Equal(t, "work", groups[1].Name)
	assert.Equal(t, "新規", groups[2].Name)
	assert.Less(t, groups[0].Order, groups[1].Order)
	assert.Less(t, groups[1].Order, groups[2].Order)
	entries := s.GetAll()
	assert.Equal(t, work.ID, entries[0].GroupID)
	assert.Equal(t, groups[1].ID, entries[1].GroupID)
}
