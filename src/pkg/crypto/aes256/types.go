package aes256

// SizeParams は各種サイズパラメータ
type SizeParams struct {
	// SaltSize はArgon2id用のソルトサイズ（バイト）
	SaltSize uint32 `json:"saltSize"`
	// NonceSize はAES-GCM用のNonceサイズ（バイト）
	NonceSize uint32 `json:"nonceSize"`
	// KeySize はAES256のキーサイズ（バイト）
	KeySize uint32 `json:"keySize"`
}

// KDFParams はArgon2idのパラメータ
type KDFParams struct {
	// Time は時間コスト（反復回数）
	Time uint32 `json:"time"`
	// Memory はメモリコスト（KB単位）
	Memory uint32 `json:"memory"`
	// Threads は並列度
	Threads uint8 `json:"threads"`
}
