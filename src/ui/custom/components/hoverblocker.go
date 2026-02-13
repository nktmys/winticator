package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// HoverBlocker はホバーイベントをブロックするラッパーウィジェット。
// desktop.Hoverable を空実装することで、親ウィジェットへのホバーイベント伝播を防ぐ。
type HoverBlocker struct {
	widget.BaseWidget

	Content fyne.CanvasObject
}

var (
	_ fyne.Widget       = (*HoverBlocker)(nil)
	_ desktop.Hoverable = (*HoverBlocker)(nil)
)

func NewHoverBlocker(content fyne.CanvasObject) *HoverBlocker {
	h := &HoverBlocker{Content: content}
	h.ExtendBaseWidget(h)
	return h
}

func (h *HoverBlocker) CreateRenderer() fyne.WidgetRenderer {
	return &hoverBlockerRenderer{blocker: h, objects: []fyne.CanvasObject{h.Content}}
}

func (h *HoverBlocker) MouseIn(*desktop.MouseEvent)    {}
func (h *HoverBlocker) MouseMoved(*desktop.MouseEvent) {}
func (h *HoverBlocker) MouseOut()                      {}

type hoverBlockerRenderer struct {
	blocker *HoverBlocker
	objects []fyne.CanvasObject
}

func (r *hoverBlockerRenderer) Destroy()                     {}
func (r *hoverBlockerRenderer) Layout(size fyne.Size)        { r.blocker.Content.Resize(size) }
func (r *hoverBlockerRenderer) MinSize() fyne.Size           { return r.blocker.Content.MinSize() }
func (r *hoverBlockerRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *hoverBlockerRenderer) Refresh()                     { r.blocker.Content.Refresh() }
