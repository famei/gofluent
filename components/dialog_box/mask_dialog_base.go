package dialog_box

import (
	"fmt"
	"math"
	"runtime"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// MaskDialogBase is a frameless, translucent dialog with a full-window mask.
// All content widgets take the center "widget" frame as their parent.
type MaskDialogBase struct {
	*qt.QDialog
	closableOnMaskClicked bool
	draggable             bool
	dragging              bool
	dragX                 int
	dragY                 int
	hBoxLayout            *qt.QHBoxLayout
	windowMask            *qt.QWidget
	widget                *qt.QFrame
	eventFilterHook       func(watched *qt.QObject, event *qt.QEvent)
	showEventHook         func()

	// Drop shadow of the center widget, remembered so the drag can keep the
	// shadow inside the window (see shadowInset).
	shadowBlurRadius float64
	shadowOffsetX    float64
	shadowOffsetY    float64

	// maskColor is the colour the full-window mask is tinted with. The alpha is the
	// fully faded-in opacity; the fade scales it (see animateMaskAlpha).
	maskR, maskG, maskB, maskA int
}

// NewMaskDialogBase builds a mask dialog base. parent is expected to be the
// window the mask should cover.
func NewMaskDialogBase(parent *qt.QWidget) *MaskDialogBase {
	d := &MaskDialogBase{QDialog: qt.NewQDialog(parent)}
	d.hBoxLayout = qt.NewQHBoxLayout(d.QWidget)
	d.windowMask = qt.NewQWidget(d.QWidget)
	d.widget = qt.NewQFrame(d.QWidget)
	d.widget.SetObjectName("centerWidget")

	// Keep Qt::Dialog so the mask dialog stays a top-level window; setting only
	// FramelessWindowHint clears the Window/Dialog type bit and prevents the
	// dialog from being shown as a modal window by QDialog::exec().
	d.SetWindowFlags(qt.FramelessWindowHint | qt.Dialog)
	d.SetAttribute(qt.WA_TranslucentBackground)
	// Position the mask over the parent window at its real screen position rather than at
	// (0,0): a translucent mask placed at the top-left corner of the screen put the centred
	// card in the wrong place (bug #9).
	d.fitToParent()

	c := 0
	if !common.IsDarkTheme() {
		c = 255
	}
	d.windowMask.Resize(d.Width(), d.Height())
	// 153 is the Python original's "rgba(c, c, c, 0.6)" alpha, written as a byte so the
	// fade can scale it (a float alpha is not readable back from a stylesheet).
	d.maskR, d.maskG, d.maskB, d.maskA = c, c, c, 153
	d.applyMaskColor(1)
	d.hBoxLayout.AddWidget(d.widget.QWidget)
	d.SetShadowEffect(60, 0, 10, qt.NewQColor11(0, 0, 0, 100))

	if runtime.GOOS != "windows" {
		// The fade-in animates the window opacity from 0 (startFadeIn), and a window is
		// mapped with the opacity it was last given: with the default 1 the dialog showed
		// one fully opaque frame, then jumped to 0 and faded in, which is the flash the
		// message box made when it opened. Setting it before the window exists removes that
		// frame on every platform that honours window opacity.
		d.SetWindowOpacity(0)
	}

	d.Window().InstallEventFilter(d.QObject)
	d.windowMask.InstallEventFilter(d.QObject)
	d.widget.InstallEventFilter(d.QObject)
	d.installEvents()
	return d
}

// fitToParent covers the parent window with the mask: the mask has to sit exactly on top of
// it, whatever the parent's position and size are. A nil parent (the standalone message box)
// leaves the geometry alone.
func (d *MaskDialogBase) fitToParent() {
	parent := d.ParentWidget()
	if parent == nil {
		return
	}
	origin := qt.NewQPoint2(0, 0)
	topLeft := parent.MapToGlobal(origin) // GoGC-armed — do NOT Delete
	origin.Delete()
	d.SetGeometry(topLeft.X(), topLeft.Y(), parent.Width(), parent.Height())
	// SetGeometry, like move(), puts the window *frame* where it was asked to; the content
	// (the mask and the card) has to land on the parent instead (see MoveWindowContentTo).
	common.MoveWindowContentTo(d.QWidget, topLeft.X(), topLeft.Y())
	d.windowMask.Resize(d.Width(), d.Height())
}

// SetShadowEffect adds a drop shadow to the center widget.
func (d *MaskDialogBase) SetShadowEffect(blurRadius, offsetX, offsetY float64, color *qt.QColor) {
	effect := qt.NewQGraphicsDropShadowEffect2(d.widget.QObject)
	effect.SetBlurRadius(blurRadius)
	effect.SetOffset2(offsetX, offsetY)
	effect.SetColor(color)
	d.widget.SetGraphicsEffect(nil)
	d.widget.SetGraphicsEffect(effect.QGraphicsEffect)
	d.shadowBlurRadius, d.shadowOffsetX, d.shadowOffsetY = blurRadius, offsetX, offsetY
}

// shadowInset reports how far the center widget's drop shadow reaches beyond the
// widget on each side.
func (d *MaskDialogBase) shadowInset() (left, top, right, bottom int) {
	blur := int(math.Ceil(d.shadowBlurRadius))
	dx := int(math.Round(d.shadowOffsetX))
	dy := int(math.Round(d.shadowOffsetY))
	return blur - dx, blur - dy, blur + dx, blur + dy
}

// SetMaskColor sets the window mask background color.
func (d *MaskDialogBase) SetMaskColor(color *qt.QColor) {
	if color == nil {
		return
	}
	d.maskR, d.maskG, d.maskB, d.maskA = color.Red(), color.Green(), color.Blue(), color.Alpha()
	d.applyMaskColor(1)
}

// applyMaskColor writes the mask colour with its alpha scaled by scale (1 is the fully
// faded-in mask, 0 a completely transparent one). The mask is a plain widget, so the
// colour lives in its stylesheet.
func (d *MaskDialogBase) applyMaskColor(scale float64) {
	alpha := d.maskA
	if scale < 1 {
		alpha = int(math.Round(float64(d.maskA) * scale))
	}
	d.windowMask.SetStyleSheet(fmt.Sprintf("background: rgba(%d, %d, %d, %d)",
		d.maskR, d.maskG, d.maskB, alpha))
}

// animateMaskAlpha fades the mask by scaling the alpha of its colour, one stylesheet
// update per animation step. Setting the stylesheet repaints the mask, which is what
// makes the fade visible where fading the window itself does not reach the screen (see
// startFadeIn).
func (d *MaskDialogBase) animateMaskAlpha(from, to float64, duration int) *qt.QVariantAnimation {
	ani := qt.NewQVariantAnimation2(d.QObject)
	ani.SetStartValue(qt.NewQVariant12(from))
	ani.SetEndValue(qt.NewQVariant12(to))
	ani.SetDuration(duration)
	ani.SetEasingCurve(qt.NewQEasingCurve3(qt.QEasingCurve__InSine))
	ani.OnValueChanged(func(value *qt.QVariant) { d.applyMaskColor(value.ToDouble()) })
	ani.Start()
	return ani
}

// IsClosableOnMaskClicked reports whether clicking the mask closes the dialog.
func (d *MaskDialogBase) IsClosableOnMaskClicked() bool { return d.closableOnMaskClicked }

// SetClosableOnMaskClicked toggles closing when the mask is clicked.
func (d *MaskDialogBase) SetClosableOnMaskClicked(isClosable bool) {
	d.closableOnMaskClicked = isClosable
}

// SetDraggable toggles dragging of the center widget.
func (d *MaskDialogBase) SetDraggable(draggable bool) { d.draggable = draggable }

// IsDraggable reports whether the center widget is draggable.
func (d *MaskDialogBase) IsDraggable() bool { return d.draggable }

// Widget exposes the center widget (all content parents to it).
func (d *MaskDialogBase) Widget() *qt.QFrame { return d.widget }

// WindowMask exposes the full-window mask widget.
func (d *MaskDialogBase) WindowMask() *qt.QWidget { return d.windowMask }

// HBoxLayout exposes the outer layout.
func (d *MaskDialogBase) HBoxLayout() *qt.QHBoxLayout { return d.hBoxLayout }

func (d *MaskDialogBase) installEvents() {
	// Fade in on show.
	d.OnShowEvent(func(super func(e *qt.QShowEvent), e *qt.QShowEvent) {
		if d.showEventHook != nil {
			d.showEventHook()
		}
		d.startFadeIn()
		super(e)
		// A window manager can override a geometry that was set before the window was
		// mapped (see common.ApplyAfterMap), which would leave the mask next to the parent
		// instead of on it.
		d.fitToParent()
		common.ApplyAfterMap(d.QWidget, d.fitToParent)
	})

	// Fade out on done.
	d.OnDone(func(super func(code int), code int) {
		d.startFadeOut(func() { super(code) })
	})

	d.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		d.windowMask.Resize(d.Width(), d.Height())
	})

	d.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		switch {
		case watched.UnsafePointer() == d.Window().QObject.UnsafePointer():
			if event.Type() == qt.QEvent__Resize {
				d.Resize(d.Window().Width(), d.Window().Height())
			}
		case watched.UnsafePointer() == d.windowMask.QObject.UnsafePointer():
			if event.Type() == qt.QEvent__MouseButtonRelease && d.closableOnMaskClicked {
				d.Reject()
			}
		case watched.UnsafePointer() == d.widget.QObject.UnsafePointer() && d.draggable:
			me := mouseEventOf(event)
			if me != nil {
				d.handleDrag(me)
			}
		}
		if d.eventFilterHook != nil {
			d.eventFilterHook(watched, event)
		}
		return super(watched, event)
	})
}

// startFadeIn fades the dialog in over 200 ms.
//
// The Python original animates a QGraphicsOpacityEffect on the dialog itself. That works
// on Windows, where the effect is composited into the window, but on X11 the effect
// animates its property while the native toplevel is never repainted, so the fade never
// reaches the screen (the dialog simply appears). There the mask is faded instead - it is
// a child widget, so its repaints always show - together with the window opacity, which
// the compositor applies to the whole window where it supports it.
func (d *MaskDialogBase) startFadeIn() {
	if runtime.GOOS == "windows" {
		effect := qt.NewQGraphicsOpacityEffect2(d.QObject)
		d.SetGraphicsEffect(effect.QGraphicsEffect)
		ani := qt.NewQPropertyAnimation4(effect.QObject, []byte("opacity"), d.QObject)
		ani.SetStartValue(qt.NewQVariant12(0))
		ani.SetEndValue(qt.NewQVariant12(1))
		ani.SetDuration(200)
		ani.SetEasingCurve(qt.NewQEasingCurve3(qt.QEasingCurve__InSine))
		ani.OnFinished(func() {
			d.SetGraphicsEffect(nil)
		})
		ani.Start()
		return
	}

	d.applyMaskColor(0)
	d.animateMaskAlpha(0, 1, 200)
	common.FadeWindowIn(d.QWidget, 200, qt.NewQEasingCurve3(qt.QEasingCurve__InSine))
}

// startFadeOut fades the dialog out over 100 ms and then runs done (QDialog.done), which
// is what actually closes it.
func (d *MaskDialogBase) startFadeOut(done func()) {
	if runtime.GOOS == "windows" {
		// The card carries the drop shadow as a graphics effect and Qt gives a widget
		// only one, so the shadow is dropped for the fade (the dialog is closing).
		d.widget.SetGraphicsEffect(nil)
		effect := qt.NewQGraphicsOpacityEffect2(d.QObject)
		d.SetGraphicsEffect(effect.QGraphicsEffect)
		ani := qt.NewQPropertyAnimation4(effect.QObject, []byte("opacity"), d.QObject)
		ani.SetStartValue(qt.NewQVariant12(1))
		ani.SetEndValue(qt.NewQVariant12(0))
		ani.SetDuration(100)
		ani.OnFinished(func() {
			d.SetGraphicsEffect(nil)
			done()
		})
		ani.Start()
		return
	}

	// The card keeps its shadow here (nothing else animates it).
	ani := d.animateMaskAlpha(1, 0, 100)
	common.FadeWindowOut(d.QWidget, 100, qt.NewQEasingCurve3(qt.QEasingCurve__InSine))
	ani.OnFinished(done)
}

func (d *MaskDialogBase) handleDrag(me *qt.QMouseEvent) {
	switch me.Type() {
	case qt.QEvent__MouseButtonPress:
		if me.Button() == qt.LeftButton {
			pos := me.Pos()
			d.dragX = pos.X()
			d.dragY = pos.Y()

			d.dragging = true
		}
	case qt.QEvent__MouseMove:
		if d.dragging {
			pos := me.Pos()
			newX := d.widget.X() + pos.X() - d.dragX
			newY := d.widget.Y() + pos.Y() - d.dragY

			// Clamp by the shadow inset rather than by 0: the mask is a
			// translucent layered window, and Qt fails its layered-window update
			// ("UpdateLayeredWindowIndirect failed ... dirty=(...) 参数错误") as
			// soon as the effect-expanded card region leaves the window — which is
			// exactly what dragging the card to the mask edge used to do.
			insetL, insetT, insetR, insetB := d.shadowInset()
			newX = clampInt(newX, insetL, d.Width()-d.widget.Width()-insetR)
			newY = clampInt(newY, insetT, d.Height()-d.widget.Height()-insetB)

			d.widget.Move(newX, newY)
		}
	case qt.QEvent__MouseButtonRelease:
		d.dragging = false
	}
}

// clampInt clamps v into [min,max]; when the window is too small for the card
// plus its shadow (max < min) the card is pinned to min.
func clampInt(v, min, max int) int {
	if max < min {
		max = min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// mouseEventOf downcasts a QEvent to a QMouseEvent for mouse event types. It
// borrows the underlying C pointer (no ownership is transferred).
func mouseEventOf(e *qt.QEvent) *qt.QMouseEvent {
	if e == nil {
		return nil
	}
	switch e.Type() {
	case qt.QEvent__MouseButtonPress, qt.QEvent__MouseButtonRelease, qt.QEvent__MouseMove:
		return qt.UnsafeNewQMouseEvent(e.UnsafePointer())
	}
	return nil
}
