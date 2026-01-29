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
)
