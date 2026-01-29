package machinekey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKey(t *testing.T) {
	// キャッシュをリセット
	ResetCache()

	key, err := DeriveKey()
	require.NoError(t, err)
	assert.Len(t, key, KeySize, "Key should be 32 bytes")

	// 2回目の呼び出しで同じキーが返されることを確認
	key2, err := DeriveKey()
	require.NoError(t, err)
	assert.Equal(t, key, key2, "Cached key should be returned")
}

func TestDeriveKeyConsistency(t *testing.T) {
	// キャッシュをリセット
	ResetCache()

	key1, err := DeriveKey()
	require.NoError(t, err)

	// キャッシュをリセットして再導出
	ResetCache()

	key2, err := DeriveKey()
	require.NoError(t, err)

	// 同じマシンでは同じキーが導出されるべき
	assert.Equal(t, key1, key2, "Same machine should derive same key")
}

func TestGetCPUID(t *testing.T) {
	cpuID, err := getCPUID()
	require.NoError(t, err)
	assert.NotEmpty(t, cpuID, "CPU ID should not be empty")

	// SHA-256ハッシュは64文字の16進数文字列
	assert.Len(t, cpuID, 64, "CPU ID should be 64 hex characters")
}

func TestKeySize(t *testing.T) {
	ResetCache()

	key, err := DeriveKey()
	require.NoError(t, err)

	// AES-256に必要な32バイトであることを確認
	assert.Equal(t, 32, len(key), "Key must be 32 bytes for AES-256")
}
