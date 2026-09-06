package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// StateCloseButton is the close button of a StateToolTip.
type StateCloseButton struct {
	*qt.QToolButton
	isPressed bool
	isEnter   bool
}

// NewStateCloseButton builds a state-tooltip close button.
func NewStateCloseButton(parent *qt.QWidget) *StateCloseButton {
	w := &StateCloseButton{QToolButton: qt.NewQToolButton(parent)}
	w.SetFixedSize2(12, 12)
	w.installEvents()
	return w
}

func (w *StateCloseButton) installEvents() {
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isEnter = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isEnter = false
		w.isPressed = false
		w.Update()
	})
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = true
		w.Update()
		super(e)
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = false
		w.Update()
		super(e)
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		if w.isPressed {
			painter.SetOpacity(0.6)
		} else if w.isEnter {
			painter.SetOpacity(0.8)
		}
		rect := qt.NewQRectF5(w.Rect())
		defer rect.Delete()
		common.Close.Render(painter, rect, reversedTheme())
		painter.End()
	})
}

// StateToolTip is a spinning/complete state tooltip shown in the corner of a
// window.
type StateToolTip struct {
	*qt.QWidget
	title         string
	content       string
	titleLabel    *qt.QLabel
	contentLabel  *qt.QLabel
	rotateTimer   *qt.QTimer
	fadeOutTimer  *qt.QTimer
	opacityEffect *qt.QGraphicsOpacityEffect
	animation     *qt.QPropertyAnimation
	closeButton   *StateCloseButton
	isDone        bool
	rotateAngle   int
	deltaAngle    int
	onClosed      func()
}

// NewStateToolTip builds a state tooltip.
func NewStateToolTip(title, content string, parent *qt.QWidget) *StateToolTip {
	w := &StateToolTip{QWidget: qt.NewQWidget(parent)}
	w.title = title
	w.content = content
	w.titleLabel = qt.NewQLabel5(title, w.QWidget)
	w.contentLabel = qt.NewQLabel5(content, w.QWidget)
	w.rotateTimer = qt.NewQTimer2(w.QObject)
	w.fadeOutTimer = qt.NewQTimer2(w.QObject)
	w.opacityEffect = qt.NewQGraphicsOpacityEffect2(w.QObject)
	w.animation = qt.NewQPropertyAnimation2(w.opacityEffect.QObject, []byte("opacity"))
	w.closeButton = NewStateCloseButton(w.QWidget)
	w.isDone = false
	w.rotateAngle = 0
	w.deltaAngle = 20
	w.initWidget()
	return w
}

// OnClosed registers the callback emitted when the close button is clicked.
func (w *StateToolTip) OnClosed(f func()) { w.onClosed = f }

func (w *StateToolTip) initWidget() {
	w.SetAttribute(qt.WA_StyledBackground)
	w.SetGraphicsEffect(w.opacityEffect.QGraphicsEffect)
	w.opacityEffect.SetOpacity(1)
	w.rotateTimer.SetInterval(50)
	w.contentLabel.SetMinimumWidth(200)

	w.closeButton.OnClicked(func() {
		w.onCloseButtonClicked()
	})
	w.rotateTimer.OnTimeout(func() {
		w.rotateTimerFlowSlot()
	})
	w.fadeOutTimer.SetSingleShot(true)
	w.fadeOutTimer.OnTimeout(func() {
		w.fadeOut()
	})

	w.setQss()
	w.initLayout()

	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		w.paint()
	})

	w.rotateTimer.Start(50)
}

func (w *StateToolTip) initLayout() {
	width := w.titleLabel.Width()
	if w.contentLabel.Width() > width {
		width = w.contentLabel.Width()
	}
	w.SetFixedSize2(width+56, 51)
	w.titleLabel.Move(32, 9)
	w.contentLabel.Move(12, 27)
	w.closeButton.Move(w.Width()-24, 19)
}

func (w *StateToolTip) setQss() {
	w.SetObjectName("StateToolTip")
	w.titleLabel.SetObjectName("titleLabel")
	w.contentLabel.SetObjectName("contentLabel")
	common.FluentStyleSheet(common.FluentStateToolTip).Apply(w.QWidget, common.ThemeAuto)
	w.titleLabel.AdjustSize()
	w.contentLabel.AdjustSize()
}

// SetTitle sets the tooltip title.
func (w *StateToolTip) SetTitle(title string) {
	w.title = title
	w.titleLabel.SetText(title)
	w.titleLabel.AdjustSize()
}

// SetContent sets the tooltip content.
func (w *StateToolTip) SetContent(content string) {
	w.content = content
	w.contentLabel.SetText(content)
	w.contentLabel.AdjustSize()
}

// SetState sets the completion state; when done the tooltip fades out.
func (w *StateToolTip) SetState(isDone bool) {
	w.isDone = isDone
	w.Update()
	if isDone {
		w.fadeOutTimer.Start(1000)
	}
}

func (w *StateToolTip) onCloseButtonClicked() {
	if w.onClosed != nil {
		w.onClosed()
	}
	w.Hide()
}

func (w *StateToolTip) fadeOut() {
	w.rotateTimer.Stop()
	w.animation.SetDuration(200)
	w.animation.SetStartValue(qt.NewQVariant12(1))
	w.animation.SetEndValue(qt.NewQVariant12(0))
	w.animation.OnFinished(func() {
		w.DeleteLater()
	})
	w.animation.Start()
}

func (w *StateToolTip) rotateTimerFlowSlot() {
	w.rotateAngle = (w.rotateAngle + w.deltaAngle) % 360
	w.Update()
}

// GetSuitablePos returns a suitable position inside the parent window.
func (w *StateToolTip) GetSuitablePos() *qt.QPoint {
	parent := w.ParentWidget()
	if parent == nil {
		return qt.NewQPoint2(0, 0)
	}
	return qt.NewQPoint2(parent.Width()-w.Width()-24, 50)
}

func (w *StateToolTip) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)
	theme := reversedTheme()

	if !w.isDone {
		offset := qt.NewQPointF3(19, 18)
		painter.Translate(offset)
		offset.Delete()
		painter.Rotate(float64(w.rotateAngle))
		rect := qt.NewQRectF4(-8, -8, 16, 16)
		defer rect.Delete()
		common.Sync.Render(painter, rect, theme)
	} else {
		rect := qt.NewQRectF4(11, 10, 16, 16)
		defer rect.Delete()
		common.Completed.Render(painter, rect, theme)
	}
	painter.End()
}
