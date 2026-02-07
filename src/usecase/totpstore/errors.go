// Package totpstore はTOTPエントリの暗号化保存・読み込みを提供する
package totpstore

import (
	"errors"
)

var (
	// ErrEntryNotFound はエントリが見つからない場合のエラー
	ErrEntryNotFound = errors.New("entry not found")

	// ErrInvalidURIScheme はURIスキームがotpauthでない場合のエラー
	ErrInvalidURIScheme = errors.New("invalid URI scheme: expected otpauth")

	// ErrNotTOTP はホストがtotpでない場合のエラー
	ErrNotTOTP = errors.New("not a TOTP URI: expected totp type")

	// ErrMissingSecret はシークレットが指定されていない場合のエラー
	ErrMissingSecret = errors.New("missing secret in URI")

	// ErrInvalidSecret はシークレットが無効な場合のエラー
	ErrInvalidSecret = errors.New("invalid Base32 secret")

	// ErrInvalidMigrationURI はotpauth-migration URIが無効な場合のエラー
	ErrInvalidMigrationURI = errors.New("invalid otpauth-migration URI")

	// ErrMissingMigrationData はmigration URIにdataパラメータがない場合のエラー
	ErrMissingMigrationData = errors.New("missing data parameter in migration URI")

	// ErrInvalidMigrationData はmigrationデータが無効な場合のエラー
	ErrInvalidMigrationData = errors.New("invalid migration data")

	// ErrNoTOTPEntries はmigrationデータにTOTPエントリがない場合のエラー
	ErrNoTOTPEntries = errors.New("no TOTP entries found in migration data")
)
