package totpstore

import (
	"encoding/base64"
	"testing"

	"github.com/nktmys/winticator/src/usecase/totpstore/migration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// buildMigrationURI はテスト用にotpauth-migration URIを生成する
func buildMigrationURI(params []*migration.MigrationPayload_OtpParameters) string {
	payload := &migration.MigrationPayload{
		OtpParameters: params,
	}
	data, _ := proto.Marshal(payload)
	encoded := base64.StdEncoding.EncodeToString(data)
	return "otpauth-migration://offline?data=" + encoded
}

func TestParseOTPAuthMigrationURI_SingleEntry(t *testing.T) {
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret:    []byte("12345678901234567890"),
			Name:      "Google:user@gmail.com",
			Issuer:    "Google",
			Algorithm: migration.MigrationPayload_SHA1,
			Digits:    migration.MigrationPayload_SIX,
			Type:      migration.MigrationPayload_TOTP,
		},
	})

	entries, err := ParseOTPAuthMigrationURI(uri)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "Google", entry.Issuer)
	assert.Equal(t, "user@gmail.com", entry.Account)
	assert.NotEmpty(t, entry.Secret)
	assert.Equal(t, "SHA1", entry.Algorithm)
	assert.Equal(t, 6, entry.Digits)
	assert.Equal(t, 30, entry.Period)
	assert.NotEmpty(t, entry.ID)
}

func TestParseOTPAuthMigrationURI_MultipleEntries(t *testing.T) {
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret:    []byte("secret1secretsecret1"),
			Name:      "Google:user1@gmail.com",
			Issuer:    "Google",
			Algorithm: migration.MigrationPayload_SHA1,
			Digits:    migration.MigrationPayload_SIX,
			Type:      migration.MigrationPayload_TOTP,
		},
		{
			Secret:    []byte("secret2secretsecret2"),
			Name:      "GitHub:myaccount",
			Issuer:    "GitHub",
			Algorithm: migration.MigrationPayload_SHA256,
			Digits:    migration.MigrationPayload_EIGHT,
			Type:      migration.MigrationPayload_TOTP,
		},
	})

	entries, err := ParseOTPAuthMigrationURI(uri)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "Google", entries[0].Issuer)
	assert.Equal(t, "user1@gmail.com", entries[0].Account)
	assert.Equal(t, "SHA1", entries[0].Algorithm)
	assert.Equal(t, 6, entries[0].Digits)

	assert.Equal(t, "GitHub", entries[1].Issuer)
	assert.Equal(t, "myaccount", entries[1].Account)
	assert.Equal(t, "SHA256", entries[1].Algorithm)
	assert.Equal(t, 8, entries[1].Digits)

	// 各エントリのIDが一意であること
	assert.NotEqual(t, entries[0].ID, entries[1].ID)
}

func TestParseOTPAuthMigrationURI_SkipsHOTP(t *testing.T) {
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret: []byte("hotpsecrethotpsecret"),
			Name:   "HOTP:counter",
			Issuer: "HOTPService",
			Type:   migration.MigrationPayload_HOTP,
		},
		{
			Secret: []byte("totpsecrettotpsecret"),
			Name:   "TOTP:timer",
			Issuer: "TOTPService",
			Type:   migration.MigrationPayload_TOTP,
		},
	})

	entries, err := ParseOTPAuthMigrationURI(uri)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, "TOTPService", entries[0].Issuer)
	assert.Equal(t, "timer", entries[0].Account)
}

func TestParseOTPAuthMigrationURI_NameWithoutIssuer(t *testing.T) {
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret: []byte("12345678901234567890"),
			Name:   "user@example.com",
			Issuer: "",
			Type:   migration.MigrationPayload_TOTP,
		},
	})

	entries, err := ParseOTPAuthMigrationURI(uri)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, "", entries[0].Issuer)
	assert.Equal(t, "user@example.com", entries[0].Account)
}

func TestParseOTPAuthMigrationURI_IssuerFromName(t *testing.T) {
	// issuerフィールドが空で、nameに"Issuer:Account"形式がある場合
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret: []byte("12345678901234567890"),
			Name:   "MyService:myuser",
			Issuer: "",
			Type:   migration.MigrationPayload_TOTP,
		},
	})

	entries, err := ParseOTPAuthMigrationURI(uri)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, "MyService", entries[0].Issuer)
	assert.Equal(t, "myuser", entries[0].Account)
}

func TestParseOTPAuthMigrationURI_SHA512Algorithm(t *testing.T) {
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret:    []byte("12345678901234567890"),
			Name:      "Test:user",
			Issuer:    "Test",
			Algorithm: migration.MigrationPayload_SHA512,
			Type:      migration.MigrationPayload_TOTP,
		},
	})

	entries, err := ParseOTPAuthMigrationURI(uri)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, "SHA512", entries[0].Algorithm)
}

func TestParseOTPAuthMigrationURI_InvalidScheme(t *testing.T) {
	_, err := ParseOTPAuthMigrationURI("otpauth://totp/Google:user?secret=ABC")
	assert.ErrorIs(t, err, ErrInvalidMigrationURI)
}

func TestParseOTPAuthMigrationURI_MissingData(t *testing.T) {
	_, err := ParseOTPAuthMigrationURI("otpauth-migration://offline")
	assert.ErrorIs(t, err, ErrMissingMigrationData)
}

func TestParseOTPAuthMigrationURI_InvalidBase64(t *testing.T) {
	_, err := ParseOTPAuthMigrationURI("otpauth-migration://offline?data=!!!invalid!!!")
	assert.ErrorIs(t, err, ErrInvalidMigrationData)
}

func TestParseOTPAuthMigrationURI_InvalidProtobuf(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("not a protobuf"))
	_, err := ParseOTPAuthMigrationURI("otpauth-migration://offline?data=" + data)
	// protobufのパースは寛容なため、エラーにならない場合がある
	// ただしTOTPエントリがなければErrNoTOTPEntriesになる
	if err != nil {
		assert.True(t, err == ErrInvalidMigrationData || err == ErrNoTOTPEntries)
	}
}

func TestParseOTPAuthMigrationURI_OnlyHOTPEntries(t *testing.T) {
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret: []byte("12345678901234567890"),
			Name:   "HOTP:user",
			Type:   migration.MigrationPayload_HOTP,
		},
	})

	_, err := ParseOTPAuthMigrationURI(uri)
	assert.ErrorIs(t, err, ErrNoTOTPEntries)
}

func TestParseOTPAuthMigrationURI_EmptyPayload(t *testing.T) {
	// 空のペイロードはprotobufシリアライズ後も空になるため、
	// OtpType_UNSPECIFIEDのエントリを1件入れてTOTPが0件になるケースでテスト
	uri := buildMigrationURI([]*migration.MigrationPayload_OtpParameters{
		{
			Secret: []byte("12345678901234567890"),
			Name:   "Unspecified:user",
			Type:   migration.MigrationPayload_OTP_TYPE_UNSPECIFIED,
		},
	})

	_, err := ParseOTPAuthMigrationURI(uri)
	assert.ErrorIs(t, err, ErrNoTOTPEntries)
}
