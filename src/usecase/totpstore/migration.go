package totpstore

import (
	"encoding/base32"
	"encoding/base64"
	"net/url"
	"strings"
	"time"

	"github.com/nktmys/winticator/src/usecase/totpstore/migration"
	"github.com/rs/xid"
	"google.golang.org/protobuf/proto"
)

// ParseOTPAuthMigrationURI はotpauth-migration:// URIをパースして複数のEntryを生成する
// 形式: otpauth-migration://offline?data=BASE64_ENCODED_PROTOBUF
func ParseOTPAuthMigrationURI(uri string) ([]*Entry, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, ErrInvalidMigrationURI
	}

	if u.Scheme != "otpauth-migration" {
		return nil, ErrInvalidMigrationURI
	}

	// dataパラメータを取得
	data := u.Query().Get("data")
	if data == "" {
		return nil, ErrMissingMigrationData
	}

	// Base64デコード
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, ErrInvalidMigrationData
	}

	// Protobufデコード
	payload := &migration.MigrationPayload{}
	if err := proto.Unmarshal(decoded, payload); err != nil {
		return nil, ErrInvalidMigrationData
	}

	// OtpParametersをEntryに変換（TOTPのみ）
	var entries []*Entry
	for _, otp := range payload.GetOtpParameters() {
		if otp.GetType() != migration.MigrationPayload_TOTP {
			continue
		}

		issuer, account := parseMigrationName(otp.GetName(), otp.GetIssuer())
		secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(otp.GetSecret())

		entries = append(entries, &Entry{
			ID:        xid.New().String(),
			Issuer:    issuer,
			Account:   account,
			Secret:    strings.ToUpper(secret),
			Algorithm: migrationAlgorithm(otp.GetAlgorithm()),
			Digits:    migrationDigits(otp.GetDigits()),
			Period:    30,
			Order:     0,
			CreatedAt: time.Now(),
		})
	}

	if len(entries) == 0 {
		return nil, ErrNoTOTPEntries
	}

	return entries, nil
}

// parseMigrationName はmigrationのname/issuerフィールドからissuerとaccountを抽出する
func parseMigrationName(name, issuer string) (string, string) {
	// nameが "Issuer:Account" 形式の場合
	if strings.Contains(name, ":") {
		parts := strings.SplitN(name, ":", 2)
		nameIssuer := strings.TrimSpace(parts[0])
		account := strings.TrimSpace(parts[1])
		// issuerフィールドが空ならnameから取得
		if issuer == "" {
			issuer = nameIssuer
		}
		return issuer, account
	}

	// nameがacountのみの場合
	return issuer, name
}

// migrationAlgorithm はProtobufのAlgorithmをEntry用の文字列に変換する
func migrationAlgorithm(algo migration.MigrationPayload_Algorithm) string {
	switch algo {
	case migration.MigrationPayload_SHA256:
		return "SHA256"
	case migration.MigrationPayload_SHA512:
		return "SHA512"
	default:
		return "SHA1"
	}
}

// migrationDigits はProtobufのDigitCountをEntry用のint値に変換する
func migrationDigits(digits migration.MigrationPayload_DigitCount) int {
	switch digits {
	case migration.MigrationPayload_EIGHT:
		return 8
	default:
		return 6
	}
}
