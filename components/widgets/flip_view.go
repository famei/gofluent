package widgets

import (
	"math"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// FlipScrollButton is the previous/next scroll button of a FlipView. The Python
// `opacity` pyqtProperty cannot be registered through miqt, so the 150ms fade is
// driven by a frame-based ProgressAnimation (see ANIMATION_GUIDE §5.2).
type FlipScrollButton struct {
	*qt.QToolButton
	_icon     common.FluentIconBase
	_opacity  float64
	isHover   bool
	isPressed bool
	fadeAni   *common.ProgressAnimation
}

// NewFlipScrollButton builds a flip scroll button.
func NewFlipScrollButton(icon common.FluentIconBase, parent *qt.QWidget) *FlipScrollButton {
	b := &FlipScrollButton{QToolButton: qt.NewQToolButton(parent), _icon: icon}
	b.installEvents()
	return b
}

// IsTransparent reports whether the button is fully transparent.
func (b *FlipScrollButton) IsTransparent() bool { return b._opacity == 0 }

// FadeIn animates the button to fully opaque over 150ms.
func (b *FlipScrollButton) FadeIn() {
	b.animateOpacity(1)
}

// FadeOut animates the button to fully transparent over 150ms.
func (b *FlipScrollButton) FadeOut() {
	b.animateOpacity(0)
}

// animateOpacity lerps the button opacity with a 150ms linear fade, matching
// Python's QPropertyAnimation(b'opacity', 150).
func (b *FlipScrollButton) animateOpacity(to float64) {
	if b.fadeAni != nil {
		b.fadeAni.Stop()
		b.fadeAni.Delete()
	}
	b.fadeAni = common.AnimateFloat(b._opacity, to, 150, nil, func(v float64) {
		b._opacity = v
		b.Update()
	})
}

func (b *FlipScrollButton) installEvents() {
	b.OnDestroyed(func() {
		if b.fadeAni != nil {
			b.fadeAni.Delete()
			b.fadeAni = nil
		}
	})
	b.OnMousePressEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		b.isPressed = true
		b.Update()
		super(ev)
	})
	b.OnMouseReleaseEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		b.isPressed = false
		b.Update()
		super(ev)
	})
	b.OnEnterEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		b.isHover = true
		b.Update()
	})
	b.OnLeaveEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		b.isHover = false
		b.Update()
	})
	b.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		// Do NOT call super(ev): the native QToolButton chrome (border/bevel/
		// background) would be painted under the Fluent rounded background and
		// stay visible even when opacity reaches 0, so only the icon would
		// disappear on fade out. Python's ScrollButton.paintEvent draws only
		// the rounded background + icon.
		_ = super
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)
		painter.SetOpacity(b._opacity)

		var bg *qt.QColor
		if common.IsDarkTheme() {
			bg = qt.NewQColor11(44, 44, 44, 245)
		} else {
			bg = qt.NewQColor11(252, 252, 252, 217)
		}
		defer bg.Delete()
		brush := qt.NewQBrush3(bg)
		defer brush.Delete()
		painter.SetBrush(brush)
		painter.DrawRoundedRect2(0, 0, b.Width(), b.Height(), 4, 4)

		var color *qt.QColor
		var iconOpacity float64
		if common.IsDarkTheme() {
			color = qt.NewQColor3(255, 255, 255)
			if b.isHover || b.isPressed {
				iconOpacity = 0.773
			} else {
				iconOpacity = 0.541
			}
		} else {
			color = qt.NewQColor3(0, 0, 0)
			if b.isHover || b.isPressed {
				iconOpacity = 0.616
			} else {
				iconOpacity = 0.45
			}
		}
		defer color.Delete()
		painter.SetOpacity(b._opacity * iconOpacity)

		s := 8
		if b.isPressed {
			s = 6
		}
		rect := qt.NewQRectF4(float64((b.Width()-s))/2, float64((b.Height()-s))/2, float64(s), float64(s))
		defer rect.Delete()
		renderFluentIconWithFill(b._icon, painter, rect, color.Name())
		painter.End()
	})
}

// FlipImageDelegate is the item delegate of a FlipView. Because miqt does not
// expose the delegate paint() virtual as an overridable callback, the image is
// drawn natively from the item's icon (set in setItemImage); this delegate only
// carries the border-radius configuration.
type FlipImageDelegate struct {
	*qt.QStyledItemDelegate
	borderRadius int
}

// NewFlipImageDelegate builds a flip image delegate.
func NewFlipImageDelegate(parent *qt.QWidget) *FlipImageDelegate {
	d := &FlipImageDelegate{QStyledItemDelegate: qt.NewQStyledItemDelegate()}
	return d
}

// SetBorderRadius sets the item border radius.
func (d *FlipImageDelegate) SetBorderRadius(radius int) {
	d.borderRadius = radius
}

// BorderRadius returns the item border radius.
func (d *FlipImageDelegate) BorderRadius() int { return d.borderRadius }

// FlipView is a list widget that flips between images. The Python smooth-scroll
// bar is replaced by a frame-based scrollbar animation driven by
// common.ProgressAnimation (see ANIMATION_GUIDE §5.7), since the custom
// SmoothScrollBar `val` property cannot be registered through miqt.
//
// Constructors
//   - NewFlipView(parent *qt.QWidget)
//   - NewFlipViewOrientation(orientation qt.Orientation, parent *qt.QWidget)
type FlipView struct {
	*qt.QListWidget
	orientation     qt.Orientation
	delegate        *FlipImageDelegate
	preButton       *FlipScrollButton
	nextButton      *FlipScrollButton
	itemSize        *qt.QSize
	aspectRatioMode qt.AspectRatioMode

	isHover               bool
	_currentIndex         int
	images                []*qt.QImage
	onCurrentIndexChanged func(index int)
	scrollAni             *common.ProgressAnimation
}

// NewFlipView builds a horizontal flip view.
func NewFlipView(parent *qt.QWidget) *FlipView {
	return newFlipView(qt.Horizontal, parent)
}

// NewFlipViewOrientation builds a flip view with the given orientation.
func NewFlipViewOrientation(orientation qt.Orientation, parent *qt.QWidget) *FlipView {
	return newFlipView(orientation, parent)
}

func newFlipView(orientation qt.Orientation, parent *qt.QWidget) *FlipView {
	f := &FlipView{QListWidget: qt.NewQListWidget(parent)}
	f.orientation = orientation
	f._currentIndex = -1
	f.aspectRatioMode = qt.IgnoreAspectRatio
	f.itemSize = qt.NewQSize2(480, 270)
	f.isHover = false
	f.OnDestroyed(func() {
		if f.scrollAni != nil {
			f.scrollAni.Delete()
			f.scrollAni = nil
		}
	})

	f.delegate = NewFlipImageDelegate(f.QWidget)
	f.SetItemDelegate(f.delegate.QAbstractItemDelegate)
	f.SetMovement(qt.QListView__Static)
	f.SetVerticalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
	f.SetHorizontalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
	f.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	f.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	f.SetMinimumSize(f.itemSize)
	f.SetIconSize(f.itemSize)
	common.FluentStyleSheet(common.FluentFlipView).Apply(f.QWidget, common.ThemeAuto)

	if f.IsHorizontal() {
		f.SetFlow(qt.QListView__LeftToRight)
		f.preButton = NewFlipScrollButton(common.CaretSolidLeft, f.QWidget)
		f.nextButton = NewFlipScrollButton(common.CaretSolidRight, f.QWidget)
		f.preButton.SetFixedSize2(16, 38)
		f.nextButton.SetFixedSize2(16, 38)
	} else {
		f.SetFlow(qt.QListView__TopToBottom)
		f.preButton = NewFlipScrollButton(common.CaretSolidUp, f.QWidget)
		f.nextButton = NewFlipScrollButton(common.CaretSolidDown, f.QWidget)
		f.preButton.SetFixedSize2(38, 16)
		f.nextButton.SetFixedSize2(38, 16)
	}

	f.preButton.OnClicked(func() { f.ScrollPrevious() })
	f.nextButton.OnClicked(func() { f.ScrollNext() })

	f.OnResizeEvent(func(super func(ev *qt.QResizeEvent), ev *qt.QResizeEvent) {
		super(ev)
		w, h := f.Width(), f.Height()
		bw, bh := f.preButton.Width(), f.preButton.Height()
		if f.IsHorizontal() {
			f.preButton.Move(2, h/2-bh/2)
			f.nextButton.Move(w-bw-2, h/2-bh/2)
		} else {
			f.preButton.Move(w/2-bw/2, 2)
			f.nextButton.Move(w/2-bw/2, h-bh-2)
		}
	})
	f.OnEnterEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		f.isHover = true
		if f._currentIndex > 0 {
			f.preButton.FadeIn()
		}
		if f._currentIndex < f.Count()-1 {
			f.nextButton.FadeIn()
		}
	})
	f.OnLeaveEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		f.isHover = false
		f.preButton.FadeOut()
		f.nextButton.FadeOut()
	})
	f.OnShowEvent(func(super func(ev *qt.QShowEvent), ev *qt.QShowEvent) {
		super(ev)
		if f._currentIndex >= 0 && f._currentIndex < f.Count() {
			if bar := f.scrollBar(); bar != nil {
				bar.SetValue(f.scrollTargetValue(f._currentIndex))
			}
		}
	})
	f.OnWheelEvent(func(super func(ev *qt.QWheelEvent), ev *qt.QWheelEvent) {
		// Do NOT call super(ev): the native QListWidget wheel handling would
		// scroll instantly on top of the animated FlipView transition, so the
		// flip would jump instead of smoothly scrolling. Accept the event and
		// drive ScrollNext/ScrollPrevious exactly like Python's
		// FlipView.wheelEvent.
		_ = super
		ev.SetAccepted(true)
		if f.scrollAni != nil && f.scrollAni.Running() {
			return
		}
		d := ev.AngleDelta()
		y := d.Y()
		if y < 0 {
			f.ScrollNext()
		} else {
			f.ScrollPrevious()
		}
	})
	return f
}

// IsHorizontal reports whether the view scrolls horizontally.
func (f *FlipView) IsHorizontal() bool { return f.orientation == qt.Horizontal }

// SetItemSize sets the base size of items.
func (f *FlipView) SetItemSize(size *qt.QSize) {
	if size.Width() == f.itemSize.Width() && size.Height() == f.itemSize.Height() {
		return
	}
	f.itemSize = qt.NewQSize2(size.Width(), size.Height())
	f.SetIconSize(f.itemSize)
	for i := 0; i < f.Count(); i++ {
		f.adjustItemSize(f.Item(i))
	}
	f.Viewport().Update()
}

// GetItemSize returns the base size of items.
func (f *FlipView) GetItemSize() *qt.QSize { return f.itemSize }

// SetBorderRadius sets the item border radius.
func (f *FlipView) SetBorderRadius(radius int) { f.delegate.SetBorderRadius(radius) }

// GetBorderRadius returns the item border radius.
func (f *FlipView) GetBorderRadius() int { return f.delegate.BorderRadius() }

// GetAspectRatioMode returns the aspect ratio mode.
func (f *FlipView) GetAspectRatioMode() qt.AspectRatioMode { return f.aspectRatioMode }

// SetAspectRatioMode sets the aspect ratio mode.
func (f *FlipView) SetAspectRatioMode(mode qt.AspectRatioMode) {
	if mode == f.aspectRatioMode {
		return
	}
	f.aspectRatioMode = mode
	for i := 0; i < f.Count(); i++ {
		f.adjustItemSize(f.Item(i))
	}
	f.Viewport().Update()
}

// CurrentIndex returns the current index.
func (f *FlipView) CurrentIndex() int { return f._currentIndex }

// SetCurrentIndex scrolls to the given index and fires the change callback.
func (f *FlipView) SetCurrentIndex(index int) {
	if index < 0 || index >= f.Count() || index == f._currentIndex {
		return
	}
	f.scrollToIndex(index)
	f.updateButtonVisibility()
	if f.onCurrentIndexChanged != nil {
		f.onCurrentIndexChanged(index)
	}
}

// OnCurrentIndexChanged registers a callback fired after the index changes.
func (f *FlipView) OnCurrentIndexChanged(cb func(index int)) {
	f.onCurrentIndexChanged = cb
}

// ScrollPrevious scrolls to the previous item.
func (f *FlipView) ScrollPrevious() { f.SetCurrentIndex(f._currentIndex - 1) }

// ScrollNext scrolls to the next item.
func (f *FlipView) ScrollNext() { f.SetCurrentIndex(f._currentIndex + 1) }

// AddImage adds a single image (*qt.QImage, *qt.QPixmap or string path).
func (f *FlipView) AddImage(image interface{}) {
	f.AddImages([]interface{}{image})
}

// AddImages adds multiple images.
func (f *FlipView) AddImages(images []interface{}) {
	if len(images) == 0 {
		return
	}
	n := f.Count()
	labels := make([]string, len(images))
	f.AddItems(labels)
	for i := n; i < f.Count(); i++ {
		f.SetItemImage(i, images[i-n])
	}
	if f._currentIndex < 0 {
		f._currentIndex = 0
	}
}

// SetItemImage sets the image of the item at index.
func (f *FlipView) SetItemImage(index int, image interface{}) {
	if index < 0 || index >= f.Count() {
		return
	}
	f.ensureImages()

	var img *qt.QImage
	switch v := image.(type) {
	case *qt.QImage:
		img = v
	case *qt.QPixmap:
		img = v.ToImage()
	case string:
		img = qt.NewQImage8(v)
	default:
		img = qt.NewQImage()
	}
	f.images[index] = img

	if !img.IsNull() {
		pm := qt.QPixmap_FromImage(img)
		icon := qt.NewQIcon2(pm)
		// pm is GoGC-armed (QPixmap_FromImage) — do NOT Delete
		f.Item(index).SetIcon(icon)
		icon.Delete() // icon is owned (NewQIcon2); SetIcon copies it into the item
	}
	f.adjustItemSize(f.Item(index))
}

// Image returns the image at index.
func (f *FlipView) Image(index int) *qt.QImage {
	if index < 0 || index >= f.Count() {
		return qt.NewQImage()
	}
	img := f.imageAt(index)
	if img == nil {
		return qt.NewQImage()
	}
	return img
}

func (f *FlipView) ensureImages() {
	for len(f.images) < f.Count() {
		f.images = append(f.images, qt.NewQImage())
	}
}

func (f *FlipView) imageAt(index int) *qt.QImage {
	if index < 0 || index >= len(f.images) {
		return nil
	}
	return f.images[index]
}

func (f *FlipView) adjustItemSize(item *qt.QListWidgetItem) {
	idx := f.Row(item)
	img := f.imageAt(idx)
	w, h := f.itemSize.Width(), f.itemSize.Height()
	if f.aspectRatioMode == qt.KeepAspectRatio && img != nil && !img.IsNull() {
		iw, ih := img.Width(), img.Height()
		if f.IsHorizontal() {
			if ih > 0 {
				w = int(float64(iw) * float64(h) / float64(ih))
			}
		} else {
			if iw > 0 {
				h = int(float64(ih) * float64(w) / float64(iw))
			}
		}
	}
	size := qt.NewQSize2(w, h)
	item.SetSizeHint(size)
	size.Delete() // size is owned (NewQSize2); SetSizeHint copies it
}

func (f *FlipView) scrollToIndex(index int) {
	if index < 0 || index >= f.Count() {
		return
	}
	f._currentIndex = index
	if bar := f.scrollBar(); bar != nil {
		f.animateScroll(bar, f.scrollTargetValue(index))
	}
}

// scrollBar returns the scrollbar that controls the flip direction.
func (f *FlipView) scrollBar() *qt.QScrollBar {
	if f.IsHorizontal() {
		return f.HorizontalScrollBar()
	}
	return f.VerticalScrollBar()
}

// scrollTargetValue computes the scrollbar value that centers the item at
// index, matching Python's FlipView.scrollToIndex (sum of previous item sizes
// plus inter-item spacing).
func (f *FlipView) scrollTargetValue(index int) int {
	value := 0
	for i := 0; i < index; i++ {
		it := f.Item(i)
		if it == nil {
			continue
		}
		if f.IsHorizontal() {
			value += it.SizeHint().Width()
		} else {
			value += it.SizeHint().Height()
		}
	}
	return value + (2*index+1)*f.Spacing()
}

// animateScroll lerps the scrollbar value to the target with an OutCubic easing
// (the Go port of SmoothScrollBar.scrollTo). The duration is shortened for very
// short distances like the Python implementation.
func (f *FlipView) animateScroll(bar *qt.QScrollBar, to int) {
	if f.scrollAni != nil {
		f.scrollAni.Stop()
		f.scrollAni.Delete()
	}
	from := bar.Value()
	dv := to - from
	if dv < 0 {
		dv = -dv
	}
	duration := 500
	if dv < 50 {
		duration = 500 * dv / 70
	}
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutCubic)
	f.scrollAni = common.AnimateFloat(float64(from), float64(to), duration, curve, func(v float64) {
		bar.SetValue(int(math.Round(v)))
	})
	curve.Delete() // SetEasingCurve copies; the temporary is freed here.
}

func (f *FlipView) updateButtonVisibility() {
	idx := f._currentIndex
	if idx == 0 {
		f.preButton.FadeOut()
	} else if f.preButton.IsTransparent() && f.isHover {
		f.preButton.FadeIn()
	}

	if idx == f.Count()-1 {
		f.nextButton.FadeOut()
	} else if f.nextButton.IsTransparent() && f.isHover {
		f.nextButton.FadeIn()
	}
}

// HorizontalFlipView is a horizontally scrolling flip view.
type HorizontalFlipView struct {
	*FlipView
}

// NewHorizontalFlipView builds a horizontal flip view.
func NewHorizontalFlipView(parent *qt.QWidget) *HorizontalFlipView {
	return &HorizontalFlipView{FlipView: NewFlipViewOrientation(qt.Horizontal, parent)}
}

// VerticalFlipView is a vertically scrolling flip view.
type VerticalFlipView struct {
	*FlipView
}

// NewVerticalFlipView builds a vertical flip view.
func NewVerticalFlipView(parent *qt.QWidget) *VerticalFlipView {
	return &VerticalFlipView{FlipView: NewFlipViewOrientation(qt.Vertical, parent)}
}
