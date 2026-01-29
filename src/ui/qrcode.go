package ui

import (
	"image"

	qrcode "github.com/skip2/go-qrcode"
)

// generateQRCodeImage はURIからQRコード画像を生成する
func generateQRCodeImage(uri string) (image.Image, error) {
	qr, err := qrcode.New(uri, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	return qr.Image(200), nil
}
