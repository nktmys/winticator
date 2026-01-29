// Package totp はRFC 6238準拠のTOTP生成機能を提供する
package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"strings"
	"time"
)

// Generate はTOTPコードを生成する
func Generate(secret string, timestamp time.Time, digits int, period int, algorithm string) (string, error) {
	// Base32デコード
	secret = strings.ToUpper(strings.TrimSpace(secret))
	// パディングを追加（必要な場合）
	if m := len(secret) % 8; m != 0 {
		secret += strings.Repeat("=", 8-m)
	}

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	// 時間カウンターを計算
	counter := uint64(timestamp.Unix()) / uint64(period)

	// カウンターをビッグエンディアンでバイト列に変換
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	// HMACを計算
	var h func() hash.Hash
	switch strings.ToUpper(algorithm) {
	case "SHA256":
		h = sha256.New
	case "SHA512":
		h = sha512.New
	default:
		h = sha1.New
	}

	mac := hmac.New(h, key)
	mac.Write(counterBytes)
	sum := mac.Sum(nil)

	// Dynamic Truncation
	offset := sum[len(sum)-1] & 0x0f
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	// 桁数に合わせてmod
	mod := uint32(1)
	for range digits {
		mod *= 10
	}
	code = code % mod

	// 桁数に合わせてゼロパディング
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, code), nil
}

// RemainingSeconds は次のコード更新までの残り秒数を返す
func RemainingSeconds(period int) int {
	return period - int(time.Now().Unix()%int64(period))
}
