package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/nktmys/winticator/src/ui/custom"
)

// CircularProgress は円形のプログレスバーウィジェット
type CircularProgress struct {
	widget.BaseWidget
	Min     float64
	Max     float64
	Value   float64
	size    float32
	fgColor color.Color
}

// NewCircularProgress は新しいCircularProgressを作成する
func NewCircularProgress(size float32) *CircularProgress {
	c := &CircularProgress{
		Min:     0,
		Max:     30,
		size:    size,
		fgColor: custom.ColorPrimaryBlue,
	}
	c.ExtendBaseWidget(c)
	return c
}

// SetColor は前景色を設定する
func (c *CircularProgress) SetColor(col color.Color) {
	c.fgColor = col
	c.Refresh()
}

// SetValue は現在の値を設定する
func (c *CircularProgress) SetValue(value float64) {
	c.Value = value
	c.Refresh()
}

// CreateRenderer はウィジェットのレンダラーを作成する
func (c *CircularProgress) CreateRenderer() fyne.WidgetRenderer {
	return newCircularProgressRenderer(c)
}

// MinSize はウィジェットの最小サイズを返す
func (c *CircularProgress) MinSize() fyne.Size {
	return fyne.NewSize(c.size, c.size)
}

// circularProgressRenderer はCircularProgressのレンダラー
type circularProgressRenderer struct {
	progress   *CircularProgress
	background *canvas.Circle
	foreground *canvas.Arc
}

func newCircularProgressRenderer(p *CircularProgress) *circularProgressRenderer {
	// 背景: 塗りつぶしの円（満ち欠けの「空」状態を表現）
	bgColor := custom.ColorProgressBackground
	bg := canvas.NewCircle(bgColor)

	// 前景: 塗りつぶしアーク（満ち欠けスタイル）
	// CutoutRatio=0 でパイチャート形式に
	fg := canvas.NewArc(0, 360, 0, p.fgColor)

	return &circularProgressRenderer{
		progress:   p,
		background: bg,
		foreground: fg,
	}
}

func (r *circularProgressRenderer) Layout(size fyne.Size) {
	// 小さい方の辺を使用して正方形を維持
	minDim := size.Width
	if size.Height < minDim {
		minDim = size.Height
	}
	circleSize := fyne.NewSize(minDim, minDim)

	// 中央に配置
	xOffset := (size.Width - minDim) / 2
	yOffset := (size.Height - minDim) / 2

	r.background.Move(fyne.NewPos(xOffset, yOffset))
	r.background.Resize(circleSize)
	r.foreground.Move(fyne.NewPos(xOffset, yOffset))
	r.foreground.Resize(circleSize)
}

func (r *circularProgressRenderer) MinSize() fyne.Size {
	return r.progress.MinSize()
}

func (r *circularProgressRenderer) Refresh() {
	// 進捗率を計算（残り時間 / 最大時間）
	ratio := float64(0)
	if r.progress.Max > r.progress.Min {
		ratio = (r.progress.Value - r.progress.Min) / (r.progress.Max - r.progress.Min)
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	// 空き部分が時計回りに増える（＝残り部分は上から反時計回りの位置に残る）
	// ratio=1 → StartAngle=0, EndAngle=360 → フル
	// ratio=0.5 → StartAngle=180, EndAngle=360 → 左半分のみ表示
	// ratio=0 → StartAngle=360, EndAngle=360 → 空
	r.foreground.StartAngle = float32(360 * (1 - ratio))
	r.foreground.EndAngle = 360

	// ウィジェットの色設定を適用
	r.foreground.FillColor = r.progress.fgColor

	r.foreground.Refresh()
	r.background.Refresh()
}

func (r *circularProgressRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.foreground}
}

func (r *circularProgressRenderer) Destroy() {}
