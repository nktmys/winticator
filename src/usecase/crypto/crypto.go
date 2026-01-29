// Package crypto はAES-256-GCM + Argon2idによる暗号化/復号を提供する
package crypto

import (
	"errors"

	"github.com/google/uuid"
	crypto "github.com/nktmys/winticator/src/pkg/crypto/aes256"
)

var (
	// ErrInvalidData はデータが無効な場合に返されるエラー
	ErrInvalidData = errors.New("invalid data")
	// ErrUnknownFormat は不明な形式の場合に返されるエラー
	ErrUnknownFormat = errors.New("unknown format")

	// GUID はV1形式の識別子: AES-256-GCM + Argon2id
	GUID = uuid.UUID{
		0x01, 0x00, 0x00, 0x00, // バージョン1
		0xAE, 0x5C, 0xBC, 0x00, // AES-GCM識別
		0xA2, 0x9D, 0x1D, 0x00, // Argon2id識別
		0x00, 0x00, 0x00, 0x01, // リビジョン1
	}

	// defaultKDFParams はデフォルトのKDFパラメータ
	defaultKDFParams = &crypto.KDFParams{
		Time:    1,         // 時間コスト（反復回数） 1回
		Memory:  64 * 1024, // メモリコスト 64MB
		Threads: 4,         // 並列度 4
	}

	// SizeParams は各種サイズパラメータ
	defaultSizeParams = &crypto.SizeParams{
		SaltSize:  32, // Argon2id用のソルトサイズ 32バイト
		NonceSize: 12, // AES-GCM用のNonceサイズ 12バイト
		KeySize:   32, // AES256のキーサイズ 32バイト
	}

	// guidSize はGUIDのサイズ
	guidSize = len(GUID)

	// minDataSize は暗号化データの最小サイズ（GUID + Salt + Nonce）
	minDataSize = guidSize + int(defaultSizeParams.SaltSize+defaultSizeParams.NonceSize)
)

// Decrypt はパスワードでデータを復号する
func Decrypt(password string, encryptedData []byte) ([]byte, error) {
	// 暗号化データの形式を検証
	err := ValidateEncryptedData(encryptedData)
	if err != nil {
		return nil, err
	}

	// 暗号化データからソルトとNonceを抽出
	salt := salt(encryptedData)
	nonce := nonce(encryptedData)

	// create AES256 instance
	crypto, err := crypto.NewAES256(password, salt, nonce, defaultSizeParams, defaultKDFParams)
	if err != nil {
		return nil, err
	}

	// 暗号化データを復号
	return crypto.Decrypt(encryptedData[minDataSize:])
}

// Encrypt はパスワードでデータを暗号化する
func Encrypt(password string, plainData []byte) ([]byte, error) {
	// create new AES256 instance with default parameters
	crypto, err := crypto.NewAES256(password, nil, nil, defaultSizeParams, defaultKDFParams)
	if err != nil {
		return nil, err
	}

	// 平文データを暗号化
	encryptedData, err := crypto.Encrypt(plainData)
	if err != nil {
		return nil, err
	}

	// 結果データを構築
	size := guidSize + len(crypto.Salt) + len(crypto.Nonce) + len(encryptedData)
	result := make([]byte, 0, size)

	// GUIDを先頭に追加
	result = append(result, GUID[:]...)
	// Saltを追加
	result = append(result, crypto.Salt...)
	// Nonceを追加
	result = append(result, crypto.Nonce...)
	// 暗号化データを追加
	result = append(result, encryptedData...)

	return result, nil
}

// ValidateEncryptedData は暗号化データの形式を検証する
func ValidateEncryptedData(data []byte) error {
	if len(data) < minDataSize {
		return ErrInvalidData
	}
	for i, v := range GUID {
		if data[i] != v {
			return ErrUnknownFormat
		}
	}
	return nil
}

// salt は暗号化データからソルトを抽出する
func salt(data []byte) []byte {
	saltStart := guidSize
	saltEnd := guidSize + int(defaultSizeParams.SaltSize)
	return data[saltStart:saltEnd]
}

// nonce は暗号化データからNonceを抽出する
func nonce(data []byte) []byte {
	nonceStart := guidSize + int(defaultSizeParams.SaltSize)
	nonceEnd := nonceStart + int(defaultSizeParams.NonceSize)
	return data[nonceStart:nonceEnd]
}
