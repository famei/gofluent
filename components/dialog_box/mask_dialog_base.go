package dialog_box

import (
	"fmt"

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
	if parent != nil {
		// Position the mask over the parent window at its real screen position
		// rather than at (0,0). The previous code placed the translucent mask at
		// the top-left corner of the screen, so the centered MessageBox card
		// appeared in the wrong place (bug #9).
		origin := qt.NewQPoint2(0, 0)
		topLeft := parent.MapToGlobal(origin) // GoGC-armed — do NOT Delete
		origin.Delete()
		d.SetGeometry(topLeft.X(), topLeft.Y(), parent.Width(), parent.Height())
	}

	c := 0
	if !common.IsDarkTheme() {
		c = 255
	}
	d.windowMask.Resize(d.Width(), d.Height())
	d.windowMask.SetStyleSheet(fmt.Sprintf("background:rgba(%d, %d, %d, 0.6)", c, c, c))
	d.hBoxLayout.AddWidget(d.widget.QWidget)
	d.SetShadowEffect(60, 0, 10, qt.NewQColor11(0, 0, 0, 100))

	d.Window().InstallEventFilter(d.QObject)
	d.windowMask.InstallEventFilter(d.QObject)
	d.widget.InstallEventFilter(d.QObject)
	d.installEvents()
	return d
}

// SetShadowEffect adds a drop shadow to the center widget.
func (d *MaskDialogBase) SetShadowEffect(blurRadius, offsetX, offsetY float64, color *qt.QColor) {
	effect := qt.NewQGraphicsDropShadowEffect2(d.widget.QObject)
	effect.SetBlurRadius(blurRadius)
	effect.SetOffset2(offsetX, offsetY)
	effect.SetColor(color)
	d.widget.SetGraphicsEffect(nil)
	d.widget.SetGraphicsEffect(effect.QGraphicsEffect)
}

// SetMaskColor sets the window mask background color.
func (d *MaskDialogBase) SetMaskColor(color *qt.QColor) {
	qss := fmt.Sprintf("background: rgba(%d, %d, %d, %d)", color.Red(), color.Green(), color.Blue(), color.Alpha())
	d.windowMask.SetStyleSheet(qss)
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
		super(e)
	})

	// Fade out on done.
	d.OnDone(func(super func(code int), code int) {
		d.widget.SetGraphicsEffect(nil)
		effect := qt.NewQGraphicsOpacityEffect2(d.QObject)
		d.SetGraphicsEffect(effect.QGraphicsEffect)
		ani := qt.NewQPropertyAnimation4(effect.QObject, []byte("opacity"), d.QObject)
		ani.SetStartValue(qt.NewQVariant12(1))
		ani.SetEndValue(qt.NewQVariant12(0))
		ani.SetDuration(100)
		ani.OnFinished(func() {
			d.SetGraphicsEffect(nil)
			super(code)
		})
		ani.Start()
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

			if newX < 0 {
				newX = 0
			}
			if newX > d.Width()-d.widget.Width() {
				newX = d.Width() - d.widget.Width()
			}
			if newY < 0 {
				newY = 0
			}
			if newY > d.Height()-d.widget.Height() {
				newY = d.Height() - d.widget.Height()
			}
			d.widget.Move(newX, newY)
		}
	case qt.QEvent__MouseButtonRelease:
		d.dragging = false
	}
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
