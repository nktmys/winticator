package assets

import (
	"embed"
	"path"

	"fyne.io/fyne/v2"
)

//go:embed icons/*
var embeddedFiles embed.FS

// fyne.Resource インターフェースを実装していることを確認
var _ fyne.Resource = (*resource)(nil)

// resource はgo:embedで埋め込まれたリソースを表す
type resource struct {
	path string
}

// Resource は指定されたファイル名のリソースを作成する
func Resource(path string) fyne.Resource {
	return &resource{path: path}
}

// Name はリソースのファイル名を返す
func (r *resource) Name() string {
	return path.Base(r.path)
}

// Content はリソースの内容を返す
func (r *resource) Content() []byte {
	data, err := embeddedFiles.ReadFile(r.path)
	if err != nil {
		return nil
	}
	return data
}
