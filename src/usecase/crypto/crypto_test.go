package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		plainData []byte
	}{
		{
			name:      "simple text",
			password:  "password123",
			plainData: []byte("Hello, World!"),
		},
		{
			name:      "empty data",
			password:  "password123",
			plainData: []byte{},
		},
		{
			name:      "binary data",
			password:  "password123",
			plainData: []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd},
		},
		{
			name:      "large data",
			password:  "password123",
			plainData: bytes.Repeat([]byte("A"), 10000),
		},
		{
			name:      "unicode password",
			password:  "パスワード🔐",
			plainData: []byte("Secret data"),
		},
		{
			name:      "unicode data",
			password:  "password123",
			plainData: []byte("日本語テキスト🎉"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.password, tt.plainData)
			require.NoError(t, err)
			require.NotNil(t, encrypted)

			decrypted, err := Decrypt(tt.password, encrypted)
			require.NoError(t, err)
			if len(tt.plainData) == 0 {
				assert.Empty(t, decrypted)
			} else {
				assert.Equal(t, tt.plainData, decrypted)
			}
		})
	}
}

func TestEncryptedDataFormat(t *testing.T) {
	password := "password123"
	plainData := []byte("test data")

	encrypted, err := Encrypt(password, plainData)
	require.NoError(t, err)

	// GUIDが先頭に含まれていることを確認
	assert.True(t, bytes.HasPrefix(encrypted, GUID[:]))

	// 最小サイズを満たしていることを確認
	assert.GreaterOrEqual(t, len(encrypted), minDataSize)
}

func TestDecryptWithWrongPassword(t *testing.T) {
	password := "correctPassword"
	wrongPassword := "wrongPassword"
	plainData := []byte("Secret data")

	encrypted, err := Encrypt(password, plainData)
	require.NoError(t, err)

	_, err = Decrypt(wrongPassword, encrypted)
	assert.Error(t, err)
}

func TestValidateEncryptedData(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedErr error
	}{
		{
			name:        "valid data",
			data:        append(GUID[:], make([]byte, minDataSize-guidSize+10)...),
			expectedErr: nil,
		},
		{
			name:        "data too short",
			data:        make([]byte, minDataSize-1),
			expectedErr: ErrInvalidData,
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "nil data",
			data:        nil,
			expectedErr: ErrInvalidData,
		},
		{
			name:        "wrong GUID",
			data:        append([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, make([]byte, minDataSize-guidSize+10)...),
			expectedErr: ErrUnknownFormat,
		},
		{
			name:        "partial GUID match",
			data:        append([]byte{0x01, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, make([]byte, minDataSize-guidSize+10)...),
			expectedErr: ErrUnknownFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEncryptedData(tt.data)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestDecryptInvalidData(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedErr error
	}{
		{
			name:        "data too short",
			data:        make([]byte, minDataSize-1),
			expectedErr: ErrInvalidData,
		},
		{
			name:        "wrong GUID",
			data:        append([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, make([]byte, minDataSize)...),
			expectedErr: ErrUnknownFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt("password", tt.data)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	password := "password123"
	plainData := []byte("Same data")

	encrypted1, err := Encrypt(password, plainData)
	require.NoError(t, err)

	encrypted2, err := Encrypt(password, plainData)
	require.NoError(t, err)

	// 同じ平文でも異なる暗号文が生成されることを確認（Salt/Nonceがランダム）
	assert.NotEqual(t, encrypted1, encrypted2)

	// 両方とも正しく復号できることを確認
	decrypted1, err := Decrypt(password, encrypted1)
	require.NoError(t, err)
	assert.Equal(t, plainData, decrypted1)

	decrypted2, err := Decrypt(password, encrypted2)
	require.NoError(t, err)
	assert.Equal(t, plainData, decrypted2)
}

func TestGUIDValue(t *testing.T) {
	// GUIDが期待通りの値であることを確認
	expected := []byte{
		0x01, 0x00, 0x00, 0x00, // バージョン1
		0xAE, 0x5C, 0xBC, 0x00, // AES-GCM識別
		0xA2, 0x9D, 0x1D, 0x00, // Argon2id識別
		0x00, 0x00, 0x00, 0x01, // リビジョン1
	}
	assert.Equal(t, expected, GUID[:])
	assert.Equal(t, 16, len(GUID))
}

func TestMinDataSize(t *testing.T) {
	// GUID(16) + Salt(32) + Nonce(12) = 60
	expectedMinSize := 16 + 32 + 12
	assert.Equal(t, expectedMinSize, minDataSize)
}
