package qrscanner

import (
	"encoding/base64"
	"image"
	"image/color"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/nktmys/winticator/src/usecase/totpstore/migration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// generateQRImage はテスト用にQRコード画像を生成する
func generateQRImage(content string) (image.Image, error) {
	writer := qrcode.NewQRCodeWriter()
	hints := make(map[gozxing.EncodeHintType]interface{})

	matrix, err := writer.Encode(content, gozxing.BarcodeFormat_QR_CODE, 200, 200, hints)
	if err != nil {
		return nil, err
	}

	// BitMatrixをimage.Imageに変換
	bounds := image.Rect(0, 0, matrix.GetWidth(), matrix.GetHeight())
	img := image.NewRGBA(bounds)

	for y := 0; y < matrix.GetHeight(); y++ {
		for x := 0; x < matrix.GetWidth(); x++ {
			if matrix.Get(x, y) {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	return img, nil
}

func TestScanImage_ValidTOTPQR(t *testing.T) {
	// TOTP QRコード画像を生成
	uri := "otpauth://totp/Google:user@gmail.com?secret=JBSWY3DPEHPK3PXP&issuer=Google"
	img, err := generateQRImage(uri)
	require.NoError(t, err)

	// スキャン
	results, err := ScanImage(img)
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, uri, result.URI)
	assert.Equal(t, "Google", result.Entry.Issuer)
	assert.Equal(t, "user@gmail.com", result.Entry.Account)
	assert.Equal(t, "JBSWY3DPEHPK3PXP", result.Entry.Secret)
}

func TestScanImage_NonTOTPQR(t *testing.T) {
	// 非TOTPのQRコード画像を生成
	img, err := generateQRImage("https://example.com")
	require.NoError(t, err)

	// スキャン
	_, err = ScanImage(img)
	assert.ErrorIs(t, err, ErrNoTOTPQRFound)
}

func TestScanImage_NoQRCode(t *testing.T) {
	// 空の画像
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// 白で塗りつぶし
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.White)
		}
	}

	// スキャン
	_, err := ScanImage(img)
	assert.ErrorIs(t, err, ErrNoQRCodeFound)
}

func TestScanImage_ComplexURI(t *testing.T) {
	// 複雑なパラメータを持つURI
	uri := "otpauth://totp/GitHub:myaccount?secret=ABCDEFGHIJKLMNOP&issuer=GitHub&algorithm=SHA256&digits=8&period=60"
	img, err := generateQRImage(uri)
	require.NoError(t, err)

	results, err := ScanImage(img)
	require.NoError(t, err)
	require.Len(t, results, 1)

	entry := results[0].Entry
	assert.Equal(t, "GitHub", entry.Issuer)
	assert.Equal(t, "myaccount", entry.Account)
	assert.Equal(t, "ABCDEFGHIJKLMNOP", entry.Secret)
	assert.Equal(t, "SHA256", entry.Algorithm)
	assert.Equal(t, 8, entry.Digits)
	assert.Equal(t, 60, entry.Period)
}

func TestScanImage_MigrationQR(t *testing.T) {
	// otpauth-migration QRコード画像を生成
	payload := &migration.MigrationPayload{
		OtpParameters: []*migration.MigrationPayload_OtpParameters{
			{
				Secret:    []byte("12345678901234567890"),
				Name:      "Google:user@gmail.com",
				Issuer:    "Google",
				Algorithm: migration.MigrationPayload_SHA1,
				Digits:    migration.MigrationPayload_SIX,
				Type:      migration.MigrationPayload_TOTP,
			},
			{
				Secret:    []byte("abcdefghijklmnopqrst"),
				Name:      "GitHub:myaccount",
				Issuer:    "GitHub",
				Algorithm: migration.MigrationPayload_SHA256,
				Digits:    migration.MigrationPayload_EIGHT,
				Type:      migration.MigrationPayload_TOTP,
			},
		},
	}
	data, err := proto.Marshal(payload)
	require.NoError(t, err)

	uri := "otpauth-migration://offline?data=" + base64.StdEncoding.EncodeToString(data)
	img, err := generateQRImage(uri)
	require.NoError(t, err)

	// スキャン
	results, err := ScanImage(img)
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "Google", results[0].Entry.Issuer)
	assert.Equal(t, "user@gmail.com", results[0].Entry.Account)
	assert.Equal(t, "SHA1", results[0].Entry.Algorithm)
	assert.Equal(t, 6, results[0].Entry.Digits)

	assert.Equal(t, "GitHub", results[1].Entry.Issuer)
	assert.Equal(t, "myaccount", results[1].Entry.Account)
	assert.Equal(t, "SHA256", results[1].Entry.Algorithm)
	assert.Equal(t, 8, results[1].Entry.Digits)
}
