package totpstore

import (
	"testing"

	"github.com/nktmys/winticator/src/pkg/machinekey"
	"github.com/nktmys/winticator/src/usecase/preferences"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newGroupTestStore はグループテスト用のStoreとPreferencesを作成する
func newGroupTestStore(t *testing.T) (*Store, *preferences.Manager) {
	t.Helper()
	machinekey.ResetCache()

	prefs := preferences.New(newMockPreferences())
	store := New(prefs)
	require.NoError(t, store.Load())
	return store, prefs
}

func groupNames(groups []*Group) []string {
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name)
	}
	return names
}

func TestStore_AddGroup(t *testing.T) {
	store, _ := newGroupTestStore(t)

	g1, err := store.AddGroup("  仕事 ")
	require.NoError(t, err)
	assert.Equal(t, "仕事", g1.Name)
	assert.NotEmpty(t, g1.ID)

	g2, err := store.AddGroup("個人")
	require.NoError(t, err)
	assert.Greater(t, g2.Order, g1.Order)
}

func TestStore_AddGroup_Validation(t *testing.T) {
	store, _ := newGroupTestStore(t)

	_, err := store.AddGroup("")
	require.ErrorIs(t, err, ErrGroupNameEmpty)
	_, err = store.AddGroup("   ")
	require.ErrorIs(t, err, ErrGroupNameEmpty)

	_, err = store.AddGroup("仕事")
	require.NoError(t, err)
	_, err = store.AddGroup(" 仕事")
	require.ErrorIs(t, err, ErrGroupNameDuplicate)

	_, err = store.AddGroup("仕事2")
	require.NoError(t, err)
	_, err = store.AddGroup("Work")
	require.NoError(t, err)
	_, err = store.AddGroup("work")
	require.NoError(t, err)
}

func TestStore_RenameGroup(t *testing.T) {
	store, _ := newGroupTestStore(t)

	a, err := store.AddGroup("A")
	require.NoError(t, err)
	_, err = store.AddGroup("B")
	require.NoError(t, err)

	require.NoError(t, store.RenameGroup(a.ID, "A"))
	require.ErrorIs(t, store.RenameGroup(a.ID, "B"), ErrGroupNameDuplicate)
	require.ErrorIs(t, store.RenameGroup(a.ID, "  "), ErrGroupNameEmpty)
	require.ErrorIs(t, store.RenameGroup("missing", "C"), ErrGroupNotFound)

	require.NoError(t, store.RenameGroup(a.ID, " C "))
	got, err := store.GetGroup(a.ID)
	require.NoError(t, err)
	assert.Equal(t, "C", got.Name)
}

func TestStore_ReorderGroups(t *testing.T) {
	store, _ := newGroupTestStore(t)

	a, _ := store.AddGroup("a")
	b, _ := store.AddGroup("b")
	c, _ := store.AddGroup("c")

	require.NoError(t, store.ReorderGroups([]string{c.ID, a.ID, b.ID}))
	assert.Equal(t, []string{"c", "a", "b"}, groupNames(store.GetGroups()))
}

func TestStore_DeleteGroup_UngroupsEntries(t *testing.T) {
	store, _ := newGroupTestStore(t)

	ga, err := store.AddGroup("A")
	require.NoError(t, err)
	gb, err := store.AddGroup("B")
	require.NoError(t, err)

	for _, spec := range []struct{ id, group string }{
		{"e1", ga.ID}, {"e2", ga.ID}, {"e3", gb.ID},
	} {
		e := NewEntry("Issuer", spec.id, "JBSWY3DPEHPK3PXP")
		e.ID = spec.id
		e.GroupID = spec.group
		require.NoError(t, store.Add(e))
	}

	require.NoError(t, store.DeleteGroup(ga.ID))
	assert.Equal(t, 3, store.Count())
	for _, id := range []string{"e1", "e2"} {
		e, err := store.Get(id)
		require.NoError(t, err)
		assert.Empty(t, e.GroupID)
	}
	e3, err := store.Get("e3")
	require.NoError(t, err)
	assert.Equal(t, gb.ID, e3.GroupID)

	_, err = store.GetGroup(ga.ID)
	require.ErrorIs(t, err, ErrGroupNotFound)
	require.ErrorIs(t, store.DeleteGroup("missing"), ErrGroupNotFound)
}

func TestStore_SaveAndLoad_Groups(t *testing.T) {
	store, prefs := newGroupTestStore(t)

	ga, _ := store.AddGroup("A")
	gb, _ := store.AddGroup("B")
	e := NewEntry("Issuer", "acc", "JBSWY3DPEHPK3PXP")
	e.GroupID = gb.ID
	require.NoError(t, store.Add(e))
	require.NoError(t, store.Save())

	loaded := New(prefs)
	require.NoError(t, loaded.Load())

	assert.Equal(t, []*Group{ga, gb}, loaded.GetGroups())
	got, err := loaded.Get(e.ID)
	require.NoError(t, err)
	assert.Equal(t, gb.ID, got.GroupID)
}

func TestStore_LoadLegacyWithoutGroups(t *testing.T) {
	store, prefs := newGroupTestStore(t)

	require.NoError(t, store.Add(NewEntry("I1", "a", "JBSWY3DPEHPK3PXP")))
	require.NoError(t, store.Add(NewEntry("I2", "b", "JBSWY3DPEHPK3PXP")))
	require.NoError(t, store.Save())
	prefs.SetTOTPGroupsData("")

	loaded := New(prefs)
	require.NoError(t, loaded.Load())
	assert.Empty(t, loaded.GetGroups())
	require.Equal(t, 2, loaded.Count())
	for _, e := range loaded.GetAll() {
		assert.Empty(t, e.GroupID)
	}
}

func entryIDs(entries []*Entry) []string {
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.ID)
	}
	return ids
}

func TestFilterByGroup(t *testing.T) {
	groups := []*Group{{ID: "A"}, {ID: "B"}}
	entries := []*Entry{
		{ID: "1", GroupID: "A"},
		{ID: "2"},
		{ID: "3", GroupID: "B"},
		{ID: "4", GroupID: "A"},
	}

	all := FilterByGroup(entries, groups, SelectionAll)
	assert.Equal(t, []string{"1", "2", "3", "4"}, entryIDs(all))

	a := FilterByGroup(entries, groups, "A")
	assert.Equal(t, []string{"1", "4"}, entryIDs(a))

	ungrouped := FilterByGroup(entries, groups, SelectionUngrouped)
	assert.Equal(t, []string{"2"}, entryIDs(ungrouped))
}

func TestFilterByGroup_OrphanIsUngrouped(t *testing.T) {
	groups := []*Group{{ID: "A"}}
	entries := []*Entry{
		{ID: "1", GroupID: "deleted"},
		{ID: "2", GroupID: "A"},
	}

	ungrouped := FilterByGroup(entries, groups, SelectionUngrouped)
	assert.Equal(t, []string{"1"}, entryIDs(ungrouped))

	deleted := FilterByGroup(entries, groups, "deleted")
	assert.Empty(t, deleted)
}

func TestResolveSelection(t *testing.T) {
	groups := []*Group{{ID: "A"}}

	assert.Equal(t, SelectionAll, ResolveSelection(groups, SelectionAll))
	assert.Equal(t, SelectionUngrouped, ResolveSelection(groups, SelectionUngrouped))
	assert.Equal(t, "A", ResolveSelection(groups, "A"))
	assert.Equal(t, SelectionAll, ResolveSelection(groups, "missing"))
}

func TestGroupIDForSelection(t *testing.T) {
	groups := []*Group{{ID: "A"}}

	assert.Empty(t, GroupIDForSelection(groups, SelectionAll))
	assert.Empty(t, GroupIDForSelection(groups, SelectionUngrouped))
	assert.Equal(t, "A", GroupIDForSelection(groups, "A"))
	assert.Empty(t, GroupIDForSelection(groups, "missing"))
}

func TestStore_MoveEntryToGroup(t *testing.T) {
	store, _ := newGroupTestStore(t)

	g, err := store.AddGroup("A")
	require.NoError(t, err)

	e1 := NewEntry("I1", "a", "JBSWY3DPEHPK3PXP")
	e2 := NewEntry("I2", "b", "JBSWY3DPEHPK3PXP")
	e3 := NewEntry("I3", "c", "JBSWY3DPEHPK3PXP")
	e3.GroupID = g.ID
	require.NoError(t, store.Add(e1))
	require.NoError(t, store.Add(e2))
	require.NoError(t, store.Add(e3))

	require.NoError(t, store.MoveEntryToGroup(e1.ID, g.ID))
	got, err := store.Get(e1.ID)
	require.NoError(t, err)
	assert.Equal(t, g.ID, got.GroupID)
	inA := FilterByGroup(store.GetAll(), store.GetGroups(), g.ID)
	assert.Equal(t, []string{e3.ID, e1.ID}, entryIDs(inA))

	require.NoError(t, store.MoveEntryToGroup(e1.ID, ""))
	ung := FilterByGroup(store.GetAll(), store.GetGroups(), SelectionUngrouped)
	assert.Equal(t, []string{e2.ID, e1.ID}, entryIDs(ung))

	require.ErrorIs(t, store.MoveEntryToGroup(e1.ID, "missing"), ErrGroupNotFound)
	require.ErrorIs(t, store.MoveEntryToGroup("missing", g.ID), ErrEntryNotFound)
}
