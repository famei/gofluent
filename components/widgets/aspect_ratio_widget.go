package widgets

import (
	"math"

	qt "github.com/mappu/miqt/qt"
)

// AspectRatioWidget keeps a single child widget at a fixed width/height ratio,
// centered, with letterbox or pillarbox margins around it.
//
// It is aimed at video surfaces, in particular at external players embedded through
// `mplayer.exe -wid <hwnd>`: the widget whose WinId() is handed over is kept exactly
// inside the ratio box, so the embedded window can never overflow the visible area,
// and ContentRect() tells the caller the exact size to give that window.
//
// Two independent behaviours:
//
//   - the child is always fitted to the ratio inside the container (the default);
//   - SetSelfAspectLocked(true) additionally makes the *container* take
//     height = width / ratio, which is what a widget in a layout needs when the
//     widget itself has to keep the ratio (Qt's heightForWidth virtual is not
//     reachable from Go, so the height is re-applied on every resize instead).
type AspectRatioWidget struct {
	*qt.QWidget
	ratio       float64
	child       *qt.QWidget
	selfLocked  bool
	lastLockedH int
}

// NewAspectRatioWidget builds an aspect-ratio container. ratio is width/height
// (e.g. 16.0/9.0); a value <= 0 leaves the child filling the whole widget.
func NewAspectRatioWidget(ratio float64, parent *qt.QWidget) *AspectRatioWidget {
	w := &AspectRatioWidget{QWidget: qt.NewQWidget(parent), ratio: ratio}
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		w.applyLayout()
	})
	return w
}

// SetAspectRatio sets width/height and re-applies the layout.
func (a *AspectRatioWidget) SetAspectRatio(ratio float64) {
	a.ratio = ratio
	a.applyLayout()
}

// NewAspectRatioWidgetFromSize builds an aspect-ratio container from a pixel size,
// which is the form the ratio usually arrives in (a video's width and height). It
// takes the ratio as width/height, so passing a size cannot accidentally invert it,
// and it avoids the integer division of writing float64(height/width) with two ints
// (which is 0 for any landscape size — the child then just fills the widget and the
// ratio looks unmaintained).
func NewAspectRatioWidgetFromSize(width, height int, parent *qt.QWidget) *AspectRatioWidget {
	return NewAspectRatioWidget(RatioOfSize(width, height), parent)
}

// SetAspectRatioFromSize sets the ratio from a pixel size (see
// NewAspectRatioWidgetFromSize) and re-applies the layout.
func (a *AspectRatioWidget) SetAspectRatioFromSize(width, height int) {
	a.SetAspectRatio(RatioOfSize(width, height))
}

// RatioOfSize returns width/height as a float64, or 0 when either side is not
// positive (the container then lets the child fill it).
func RatioOfSize(width, height int) float64 {
	if width <= 0 || height <= 0 {
		return 0
	}
	return float64(width) / float64(height)
}

// AspectRatio returns the current width/height ratio (0 when the child fills).
func (a *AspectRatioWidget) AspectRatio() float64 { return a.ratio }

// SetWidget puts child inside the container. The container does not take ownership.
func (a *AspectRatioWidget) SetWidget(child *qt.QWidget) {
	if child == nil {
		return
	}
	child.SetParent(a.QWidget)
	a.child = child
	a.applyLayout()
}

// Widget returns the child widget.
func (a *AspectRatioWidget) Widget() *qt.QWidget { return a.child }

// SetSelfAspectLocked makes the container keep its own height at width/ratio, so a
// container placed in a layout keeps the aspect ratio itself.
func (a *AspectRatioWidget) SetSelfAspectLocked(locked bool) {
	a.selfLocked = locked
	if !locked {
		a.SetMinimumHeight(0)
		a.SetMaximumHeight(16777215)
	}
	a.applyLayout()
}

// IsSelfAspectLocked reports whether the container keeps its own aspect ratio.
func (a *AspectRatioWidget) IsSelfAspectLocked() bool { return a.selfLocked }

// ContentRect returns the rect the child occupies inside the container, in the
// container's coordinates. This is the rect to give to an embedded external window
// (mplayer -wid) so it matches the video surface exactly.
func (a *AspectRatioWidget) ContentRect() *qt.QRect {
	return a.fitRect()
}

// Refresh re-applies the layout immediately (the resize event does this
// automatically; call it after resizing the container from code if the child has to
// be up to date before the event loop runs again).
func (a *AspectRatioWidget) Refresh() { a.applyLayout() }

// applyLayout fits the child and, when locked, the container's own height.
func (a *AspectRatioWidget) applyLayout() {
	if a.selfLocked && a.ratio > 0 {
		h := int(math.Round(float64(a.Width()) / a.ratio))
		if h != a.lastLockedH && h > 0 {
			a.lastLockedH = h
			a.SetFixedHeight(h)
		}
	}
	if a.child == nil {
		return
	}
	r := a.fitRect()
	defer r.Delete()
	a.child.SetGeometry(r.X(), r.Y(), r.Width(), r.Height())
}

// fitRect returns the largest rect with the configured ratio that fits inside the
// container, centered (letterbox / pillarbox).
func (a *AspectRatioWidget) fitRect() *qt.QRect {
	w, h := a.Width(), a.Height()
	if w <= 0 || h <= 0 {
		return qt.NewQRect4(0, 0, w, h)
	}
	if a.ratio <= 0 {
		return qt.NewQRect4(0, 0, w, h)
	}

	cw, ch := w, int(math.Round(float64(w)/a.ratio))
	if ch > h {
		ch = h
		cw = int(math.Round(float64(h) * a.ratio))
	}
	if cw < 1 {
		cw = 1
	}
	if ch < 1 {
		ch = 1
	}
	return qt.NewQRect4((w-cw)/2, (h-ch)/2, cw, ch)
}
