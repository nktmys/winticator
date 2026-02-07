package totpstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOTPAuthURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		issuer  string
		account string
		secret  string
		wantErr bool
	}{
		{
			name:    "Standard Google Authenticator URI",
			uri:     "otpauth://totp/Google:user@gmail.com?secret=JBSWY3DPEHPK3PXP&issuer=Google",
			issuer:  "Google",
			account: "user@gmail.com",
			secret:  "JBSWY3DPEHPK3PXP",
			wantErr: false,
		},
		{
			name:    "URI without issuer in path",
			uri:     "otpauth://totp/user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example",
			issuer:  "Example",
			account: "user@example.com",
			secret:  "JBSWY3DPEHPK3PXP",
			wantErr: false,
		},
		{
			name:    "URI with all parameters",
			uri:     "otpauth://totp/GitHub:myaccount?secret=ABCDEFGH&issuer=GitHub&algorithm=SHA256&digits=8&period=60",
			issuer:  "GitHub",
			account: "myaccount",
			secret:  "ABCDEFGH",
			wantErr: false,
		},
		{
			name:    "Invalid scheme",
			uri:     "https://example.com",
			wantErr: true,
		},
		{
			name:    "Not TOTP",
			uri:     "otpauth://hotp/Test:user?secret=ABC",
			wantErr: true,
		},
		{
			name:    "Missing secret",
			uri:     "otpauth://totp/Test:user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseOTPAuthURI(tt.uri)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.issuer, entry.Issuer)
			assert.Equal(t, tt.account, entry.Account)
			assert.Equal(t, tt.secret, entry.Secret)
		})
	}
}

func TestEntryToOTPAuthURI(t *testing.T) {
	entry := &Entry{
		ID:        "test-id",
		Issuer:    "Google",
		Account:   "user@gmail.com",
		Secret:    "JBSWY3DPEHPK3PXP",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
	}

	uri := entry.ToOTPAuthURI()
	assert.Contains(t, uri, "otpauth://totp/")
	assert.Contains(t, uri, "Google")
	assert.Contains(t, uri, "gmail.com")
	assert.Contains(t, uri, "secret=JBSWY3DPEHPK3PXP")

	// URIをパースして往復確認
	parsed, err := ParseOTPAuthURI(uri)
	require.NoError(t, err)
	assert.Equal(t, entry.Issuer, parsed.Issuer)
	assert.Equal(t, entry.Account, parsed.Account)
	assert.Equal(t, entry.Secret, parsed.Secret)
}

func TestNewEntry(t *testing.T) {
	entry := NewEntry("GitHub", "myaccount", "SECRETKEY")

	assert.NotEmpty(t, entry.ID)
	assert.Equal(t, "GitHub", entry.Issuer)
	assert.Equal(t, "myaccount", entry.Account)
	assert.Equal(t, "SECRETKEY", entry.Secret)
	assert.Equal(t, "SHA1", entry.Algorithm)
	assert.Equal(t, 6, entry.Digits)
	assert.Equal(t, 30, entry.Period)
	assert.False(t, entry.CreatedAt.IsZero())
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		issuer   string
		account  string
		expected string
	}{
		{"Google", "user@gmail.com", "Google: user@gmail.com"},
		{"Google", "", "Google"},
		{"", "user@example.com", "user@example.com"},
	}

	for _, tt := range tests {
		entry := &Entry{Issuer: tt.issuer, Account: tt.account}
		assert.Equal(t, tt.expected, entry.DisplayName())
	}
}
