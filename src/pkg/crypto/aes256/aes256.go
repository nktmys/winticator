// Package aes256 はAES-256-GCM + Argon2idによる暗号化/復号を提供する
package aes256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"golang.org/x/crypto/argon2"
)

// AES256 はAES-256-GCM + Argon2idによる暗号化/復号を提供する構造体
type AES256 struct {
	Salt  []byte      // Argon2id用ソルト
	Nonce []byte      // メタデータ暗号化用Nonce
	key   []byte      // AES256キー
	gcm   cipher.AEAD // AES-GCMモード
}

// NewAES256 はAES256構造体の新しいインスタンスを生成する
func NewAES256(password string, salt, nonce []byte, sizeParams *SizeParams, kdfParams *KDFParams) (*AES256, error) {
	// ソルトを生成
	if len(salt) == 0 {
		salt = make([]byte, sizeParams.SaltSize)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, err
		}
	}

	// メタデータ用Nonceを生成
	if len(nonce) == 0 {
		nonce = make([]byte, sizeParams.NonceSize)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}
	}

	// パスワードからキーを導出（Argon2id）
	key := argon2.IDKey([]byte(password), salt, kdfParams.Time, kdfParams.Memory, kdfParams.Threads, sizeParams.KeySize)

	// AESブロック暗号を作成
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// GCMモードを使用
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AES256{
		Salt:  salt,
		Nonce: nonce,
		gcm:   gcm,
		key:   key,
	}, nil
}

// Encrypt はデータをAES256-GCMで暗号化する
func (c *AES256) Encrypt(plain []byte) ([]byte, error) {
	// データを暗号化
	encrypted := c.gcm.Seal(nil, c.Nonce, plain, nil)
	return encrypted, nil
}

// Decrypt はデータをAES256-GCMで復号する
func (c *AES256) Decrypt(encrypted []byte) ([]byte, error) {
	// データを復号
	plain, err := c.gcm.Open(nil, c.Nonce, encrypted, nil)
	return plain, err
}
