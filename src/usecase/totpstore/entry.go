// Package totpstore はTOTPエントリの管理と暗号化保存を提供する
package totpstore

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nktmys/winticator/src/pkg/totp"
	"github.com/rs/xid"
)

// Entry はTOTPエントリを表す構造体
type Entry struct {
	ID        string    `json:"id"`         // UUID
	Issuer    string    `json:"issuer"`     // サービス名 (例: "Google")
	Account   string    `json:"account"`    // アカウント名 (例: "user@gmail.com")
	Secret    string    `json:"secret"`     // Base32シークレットキー
	Algorithm string    `json:"algorithm"`  // "SHA1", "SHA256", "SHA512"
	Digits    int       `json:"digits"`     // 6 または 8
	Period    int       `json:"period"`     // 秒単位 (通常30)
	Order     int       `json:"order"`      // 表示順序
	CreatedAt time.Time `json:"created_at"` // 登録日時
}

// NewEntry は新しいTOTPエントリを作成する
func NewEntry(issuer string, account string, secret string) *Entry {
	return &Entry{
		ID:        xid.New().String(),
		Issuer:    issuer,
		Account:   account,
		Secret:    secret,
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Order:     0,
		CreatedAt: time.Now(),
	}
}

// ParseOTPAuthURI はotpauth:// URIをパースしてEntryを生成する
// 形式: otpauth://totp/ISSUER:ACCOUNT?secret=SECRET&issuer=ISSUER&algorithm=SHA1&digits=6&period=30
func ParseOTPAuthURI(uri string) (*Entry, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "otpauth" {
		return nil, ErrInvalidURIScheme
	}

	if u.Host != "totp" {
		return nil, ErrNotTOTP
	}

	// パスからissuerとaccountを抽出
	// 形式: /ISSUER:ACCOUNT または /ACCOUNT
	path := strings.TrimPrefix(u.Path, "/")
	var issuer, account string

	if strings.Contains(path, ":") {
		parts := strings.SplitN(path, ":", 2)
		issuer = parts[0]
		account = parts[1]
	} else {
		account = path
	}

	// クエリパラメータを取得
	query := u.Query()

	secret := query.Get("secret")
	if secret == "" {
		return nil, ErrMissingSecret
	}

	// issuerがパスにない場合、クエリから取得
	if issuer == "" {
		issuer = query.Get("issuer")
	}

	// デフォルト値を設定
	algorithm := query.Get("algorithm")
	if algorithm == "" {
		algorithm = "SHA1"
	}

	digits := 6
	if d := query.Get("digits"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			digits = parsed
		}
	}

	period := 30
	if p := query.Get("period"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			period = parsed
		}
	}

	return &Entry{
		ID:        xid.New().String(),
		Issuer:    issuer,
		Account:   account,
		Secret:    strings.ToUpper(secret), // Base32は大文字
		Algorithm: strings.ToUpper(algorithm),
		Digits:    digits,
		Period:    period,
		Order:     0,
		CreatedAt: time.Now(),
	}, nil
}

// ToOTPAuthURI はEntryをotpauth:// URI形式に変換する
func (e *Entry) ToOTPAuthURI() string {
	// ラベルを構築
	var label string
	if e.Issuer != "" {
		label = url.PathEscape(e.Issuer) + ":" + url.PathEscape(e.Account)
	} else {
		label = url.PathEscape(e.Account)
	}

	// クエリパラメータを構築
	params := url.Values{}
	params.Set("secret", e.Secret)
	if e.Issuer != "" {
		params.Set("issuer", e.Issuer)
	}
	if e.Algorithm != "SHA1" {
		params.Set("algorithm", e.Algorithm)
	}
	if e.Digits != 6 {
		params.Set("digits", strconv.Itoa(e.Digits))
	}
	if e.Period != 30 {
		params.Set("period", strconv.Itoa(e.Period))
	}

	return "otpauth://totp/" + label + "?" + params.Encode()
}

// DisplayName は表示用の名前を返す
func (e *Entry) DisplayName() string {
	if e.Account != "" {
		return e.Account
	}
	return e.Issuer
}

// TOTP はEntryから現在のTOTPコードを生成する
func (e *Entry) TOTP() (string, error) {
	code, err := totp.Generate(e.Secret, time.Now(), e.Digits, e.Period, e.Algorithm)
	if err != nil {
		return "", ErrInvalidSecret
	}
	return code, nil
}

// RemainingSeconds は次のコード更新までの残り秒数を返す
func (e *Entry) RemainingSeconds() int {
	return totp.RemainingSeconds(e.Period)
}
