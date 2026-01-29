package totp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	// RFC 6238 テストベクター
	// https://datatracker.ietf.org/doc/html/rfc6238#appendix-B
	// シークレット: "12345678901234567890" (ASCII) = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ" (Base32)
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

	tests := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "Test at Unix time 59",
			timestamp: time.Unix(59, 0),
			expected:  "287082",
		},
		{
			name:      "Test at Unix time 1111111109",
			timestamp: time.Unix(1111111109, 0),
			expected:  "081804",
		},
		{
			name:      "Test at Unix time 1111111111",
			timestamp: time.Unix(1111111111, 0),
			expected:  "050471",
		},
		{
			name:      "Test at Unix time 1234567890",
			timestamp: time.Unix(1234567890, 0),
			expected:  "005924",
		},
		{
			name:      "Test at Unix time 2000000000",
			timestamp: time.Unix(2000000000, 0),
			expected:  "279037",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := Generate(secret, tt.timestamp, 6, 30, "SHA1")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, code)
		})
	}
}

func TestGenerateWithPadding(t *testing.T) {
	// パディングなしのシークレットでもテスト
	secret := "JBSWY3DPEHPK3PXP" // 一般的なテストシークレット

	code, err := Generate(secret, time.Now(), 6, 30, "SHA1")
	require.NoError(t, err)
	assert.Len(t, code, 6)
}

func TestGenerateInvalidSecret(t *testing.T) {
	_, err := Generate("invalid!@#$", time.Now(), 6, 30, "SHA1")
	assert.Error(t, err)
}

func TestGenerateDifferentAlgorithms(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	timestamp := time.Unix(59, 0)

	algorithms := []string{"SHA1", "SHA256", "SHA512"}
	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			code, err := Generate(secret, timestamp, 6, 30, algo)
			require.NoError(t, err)
			assert.Len(t, code, 6)
		})
	}
}

func TestGenerateDifferentDigits(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	timestamp := time.Now()

	tests := []struct {
		digits   int
		expected int
	}{
		{6, 6},
		{8, 8},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			code, err := Generate(secret, timestamp, tt.digits, 30, "SHA1")
			require.NoError(t, err)
			assert.Len(t, code, tt.expected)
		})
	}
}

func TestRemainingSeconds(t *testing.T) {
	remaining := RemainingSeconds(30)
	assert.GreaterOrEqual(t, remaining, 1)
	assert.LessOrEqual(t, remaining, 30)
}
