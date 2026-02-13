package clipboard

import (
	"testing"
	"time"

	fynetest "fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	clip := fynetest.NewClipboard()
	m := New(clip)

	m.Copy("123456", 1*time.Hour)

	assert.Equal(t, "123456", clip.Content())
}

func TestCopy_OverwritesPrevious(t *testing.T) {
	clip := fynetest.NewClipboard()
	m := New(clip)

	m.Copy("111111", 1*time.Hour)
	m.Copy("222222", 1*time.Hour)

	assert.Equal(t, "222222", clip.Content())
}

func TestClear_ClearsMatchingContent(t *testing.T) {
	clip := fynetest.NewClipboard()
	m := New(clip)

	m.Copy("123456", 1*time.Hour)
	require.Equal(t, "123456", clip.Content())

	m.Clear()

	assert.Equal(t, "", clip.Content())
}

func TestClear_SkipsWhenContentChanged(t *testing.T) {
	clip := fynetest.NewClipboard()
	m := New(clip)

	m.Copy("123456", 1*time.Hour)
	clip.SetContent("user pasted something else")

	m.Clear()

	assert.Equal(t, "user pasted something else", clip.Content())
}

func TestClear_NoPanicWhenNothingCopied(t *testing.T) {
	clip := fynetest.NewClipboard()
	m := New(clip)

	assert.NotPanics(t, func() {
		m.Clear()
	})
}

func TestCopy_AutoClearAfterDelay(t *testing.T) {
	clip := fynetest.NewClipboard()
	m := New(clip)

	m.Copy("123456", 50*time.Millisecond)

	require.Equal(t, "123456", clip.Content())

	time.Sleep(150 * time.Millisecond)

	assert.Equal(t, "", clip.Content())
}

func TestCopy_AutoClearSkipsWhenContentChanged(t *testing.T) {
	clip := fynetest.NewClipboard()
	m := New(clip)

	m.Copy("123456", 50*time.Millisecond)
	clip.SetContent("something else")

	time.Sleep(150 * time.Millisecond)

	assert.Equal(t, "something else", clip.Content())
}
