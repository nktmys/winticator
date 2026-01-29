// Package machinekey はマシン固有の識別子から暗号化キーを導出する
package machinekey

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"golang.org/x/crypto/argon2"
)

const (
	// KeySize は導出される鍵のサイズ（バイト）
	KeySize = 32

	// Argon2idパラメータ
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64MB
	argonThreads = 4
)

// 固定ソルト（アプリケーション固有）
var fixedSalt = []byte{
	0x57, 0x69, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74,
	0x6f, 0x72, 0x2d, 0x54, 0x4f, 0x54, 0x50, 0x21,
	0x32, 0x30, 0x32, 0x35, 0x2d, 0x4b, 0x65, 0x79,
	0x44, 0x65, 0x72, 0x69, 0x76, 0x65, 0x21, 0x21,
}

var (
	cachedKey  []byte
	cacheOnce  sync.Once
	cacheError error
)

// DeriveKey はマシン固有のCPU IDから暗号化キーを導出する
// 導出されたキーはキャッシュされ、以降の呼び出しでは同じキーが返される
func DeriveKey() ([]byte, error) {
	cacheOnce.Do(func() {
		cachedKey, cacheError = deriveKeyInternal()
	})
	return cachedKey, cacheError
}

// deriveKeyInternal は実際のキー導出処理を行う
func deriveKeyInternal() ([]byte, error) {
	cpuID, err := getCPUID()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU ID: %w", err)
	}

	// CPU IDをArgon2idで鍵導出
	key := argon2.IDKey([]byte(cpuID), fixedSalt, argonTime, argonMemory, argonThreads, KeySize)
	return key, nil
}

// getCPUID はCPU識別子を取得する
func getCPUID() (string, error) {
	infos, err := cpu.Info()
	if err != nil {
		return "", err
	}

	if len(infos) == 0 {
		return "", fmt.Errorf("no CPU info available")
	}

	// CPU情報から一意の識別子を生成
	// VendorID + Family + Model + PhysicalID を組み合わせる
	info := infos[0]
	combined := fmt.Sprintf("%s-%s-%s-%s-%s",
		info.VendorID,
		info.ModelName,
		info.PhysicalID,
		info.Family,
		info.Model,
	)

	// SHA-256でハッシュ化して固定長の識別子を生成
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash), nil
}

// ResetCache はキャッシュをリセットする（テスト用）
func ResetCache() {
	cacheOnce = sync.Once{}
	cachedKey = nil
	cacheError = nil
}
