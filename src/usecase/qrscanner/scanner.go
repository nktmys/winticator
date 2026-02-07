// Package qrscanner は画面キャプチャからQRコードを読み取る機能を提供する
package qrscanner

import (
	"errors"
	"image"
	"image/draw"
	"strings"

	"github.com/go-vgo/robotgo"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/nktmys/winticator/src/usecase/totpstore"
)

var (
	// ErrNoQRCodeFound はQRコードが見つからない場合のエラー
	ErrNoQRCodeFound = errors.New("no QR code found in the captured image")

	// ErrNoTOTPQRFound はTOTP用のQRコードが見つからない場合のエラー
	ErrNoTOTPQRFound = errors.New("no TOTP QR code found (otpauth://totp/...)")

	// ErrScreenCaptureFailed は画面キャプチャに失敗した場合のエラー
	ErrScreenCaptureFailed = errors.New("failed to capture screen")
)

// ScanResult はQRスキャン結果を表す構造体
type ScanResult struct {
	Entry *totpstore.Entry // パース済みのTOTPエントリ
	URI   string           // 元のotpauth:// URI
}

// CaptureAndScan は画面全体をキャプチャしてQRコードをスキャンする
func CaptureAndScan() ([]ScanResult, error) {
	// 画面をキャプチャ
	img, err := captureScreen()
	if err != nil {
		return nil, err
	}

	// QRコードをデコード
	return scanQRCodes(img)
}

// captureScreen は画面全体をキャプチャしてimage.Imageを返す
func captureScreen() (image.Image, error) {
	// robotgoで画面サイズを取得
	width, height := robotgo.GetScreenSize()

	// 画面全体をキャプチャ
	bitmap := robotgo.CaptureScreen(0, 0, width, height)
	if bitmap == nil {
		return nil, ErrScreenCaptureFailed
	}
	defer robotgo.FreeBitmap(bitmap)

	// bitmapをimage.Imageに変換
	img := robotgo.ToImage(bitmap)
	if img == nil {
		return nil, ErrScreenCaptureFailed
	}

	// robotgo.ToImageが返す画像は境界とピクセルデータの容量が一致しない場合があるため、
	// 新しい画像にコピーして正しいメモリレイアウトを保証する
	return copyImage(img), nil
}

// copyImage は画像を新しいRGBA画像にコピーする
// robotgoの画像は内部的にstrideやPixスライスの容量が不正な場合があるため、
// 有効な範囲のみをコピーする
func copyImage(src image.Image) *image.RGBA {
	// robotgoの画像はPixスライスの容量が不足している場合があるため、
	// 実際にアクセス可能な範囲を計算する
	validBounds := calculateValidBounds(src)

	dst := image.NewRGBA(validBounds)
	draw.Draw(dst, validBounds, src, validBounds.Min, draw.Src)
	return dst
}

// calculateValidBounds はPixスライスの容量に基づいて実際にアクセス可能な境界を計算する
func calculateValidBounds(img image.Image) image.Rectangle {
	bounds := img.Bounds()

	// *image.RGBAの場合、Pixスライスの容量をチェック
	if rgba, ok := img.(*image.RGBA); ok {
		pixLen := len(rgba.Pix)
		if pixLen == 0 || rgba.Stride == 0 {
			return image.Rectangle{}
		}

		// 有効な行数を計算
		// 最後の行は完全なStrideを必要としないため、少し複雑な計算が必要
		width := bounds.Dx()
		bytesPerPixel := 4
		lastRowBytes := width * bytesPerPixel

		// 完全な行数 + 最後の不完全な行があるかどうか
		validHeight := min((pixLen+rgba.Stride-lastRowBytes)/rgba.Stride, bounds.Dy())

		return image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+validHeight)
	}

	return bounds
}

// scanQRCodes は画像からQRコードを検出してTOTPエントリを返す
func scanQRCodes(img image.Image) ([]ScanResult, error) {
	// gozxingのBinaryBitmapを作成
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}

	// QRコードリーダーを作成
	reader := qrcode.NewQRCodeReader()

	// QRコードをデコード
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		// 見つからない場合
		if strings.Contains(err.Error(), "NotFoundException") {
			return nil, ErrNoQRCodeFound
		}
		return nil, err
	}

	// 結果をパース
	uri := result.GetText()

	switch {
	// otpauth-migration:// URI（Google Authenticatorエクスポート形式）
	case strings.HasPrefix(uri, "otpauth-migration://"):
		entries, err := totpstore.ParseOTPAuthMigrationURI(uri)
		if err != nil {
			return nil, err
		}
		results := make([]ScanResult, len(entries))
		for i, entry := range entries {
			results[i] = ScanResult{Entry: entry, URI: uri}
		}
		return results, nil

	// otpauth://totp/ URI（標準TOTP形式）
	case strings.HasPrefix(uri, "otpauth://totp/"):
		entry, err := totpstore.ParseOTPAuthURI(uri)
		if err != nil {
			return nil, err
		}
		return []ScanResult{
			{
				Entry: entry,
				URI:   uri,
			},
		}, nil

	default:
		return nil, ErrNoTOTPQRFound
	}
}

// ScanImage は指定した画像からQRコードをスキャンする（テスト用）
func ScanImage(img image.Image) ([]ScanResult, error) {
	return scanQRCodes(img)
}
