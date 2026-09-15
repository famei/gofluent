package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// FlyoutAnimationType enumerates the flyout pop-up animation styles.
type FlyoutAnimationType int

const (
	// FlyoutAnimationPullUp animates the flyout pulling up.
	FlyoutAnimationPullUp FlyoutAnimationType = iota
	// FlyoutAnimationDropDown animates the flyout dropping down.
	FlyoutAnimationDropDown
	// FlyoutAnimationSlideLeft animates the flyout sliding left.
	FlyoutAnimationSlideLeft
	// FlyoutAnimationSlideRight animates the flyout sliding right.
	FlyoutAnimationSlideRight
	// FlyoutAnimationFadeIn animates the flyout fading in.
	FlyoutAnimationFadeIn
	// FlyoutAnimationNone shows the flyout without animation.
	FlyoutAnimationNone
)

// FlyoutIconWidget is a small widget that paints the flyout icon.
type FlyoutIconWidget struct {
	*qt.QWidget
	icon interface{}
}

// NewFlyoutIconWidget builds a flyout icon widget.
func NewFlyoutIconWidget(icon interface{}, parent *qt.QWidget) *FlyoutIconWidget {
	w := &FlyoutIconWidget{QWidget: qt.NewQWidget(parent), icon: icon}
	w.SetFixedSize2(36, 54)
	w.installPaintEvent()
	return w
}

func (w *FlyoutIconWidget) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		if w.icon == nil {
			return
		}
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)
		rect := qt.NewQRectF4(8, float64(w.Height()-20)/2, 20, 20)
		defer rect.Delete()
		drawInfoBarIconOrGeneric(w.icon, painter, rect)
		painter.End()
	})
}

// FlyoutViewBase is the base class of flyout views. It paints the rounded
// background and border.
type FlyoutViewBase struct {
	*qt.QWidget
	drawBackground bool
}

// NewFlyoutViewBase builds a flyout view base.
func NewFlyoutViewBase(parent *qt.QWidget) *FlyoutViewBase {
	w := &FlyoutViewBase{QWidget: qt.NewQWidget(parent), drawBackground: true}
	w.installPaintEvent()
	return w
}

// SetDrawBackground toggles whether the rounded background is painted. Some
// views (e.g. the teaching tip) paint their own background elsewhere.
func (w *FlyoutViewBase) SetDrawBackground(draw bool) {
	w.drawBackground = draw
	w.Update()
}

// BackgroundColor returns the theme-appropriate background color.
func (w *FlyoutViewBase) BackgroundColor() *qt.QColor {
	if common.IsDarkTheme() {
		return qt.NewQColor3(40, 40, 40)
	}
	return qt.NewQColor3(248, 248, 248)
}

// BorderColor returns the theme-appropriate border color.
func (w *FlyoutViewBase) BorderColor() *qt.QColor {
	if common.IsDarkTheme() {
		return qt.NewQColor11(0, 0, 0, 45)
	}
	return qt.NewQColor11(0, 0, 0, 17)
}

func (w *FlyoutViewBase) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		if !w.drawBackground {
			return
		}
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		bg := w.BackgroundColor()
		defer bg.Delete()
		border := w.BorderColor()
		defer border.Delete()
		brush := qt.NewQBrush3(bg)
		defer brush.Delete()
		painter.SetBrush(brush)
		painter.SetPen(border)

		r := w.Rect()
		rect := r.Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete

		painter.DrawRoundedRect3(rect, 8, 8)
		painter.End()
	})
}

// FlyoutView is the default flyout view with title, content, icon and image.
type FlyoutView struct {
	*FlyoutViewBase
	icon         interface{}
	title        string
	image        interface{}
	content      string
	isClosable   bool
	vBoxLayout   *qt.QVBoxLayout
	viewLayout   *qt.QHBoxLayout
	widgetLayout *qt.QVBoxLayout
	titleLabel   *qt.QLabel
	contentLabel *qt.QLabel
	iconWidget   *FlyoutIconWidget
	imageLabel   *ImageLabel
	closeButton  *qt.QToolButton
	onClosed     func()
}

// NewFlyoutView builds a flyout view.
func NewFlyoutView(title, content string, icon, image interface{}, isClosable bool, parent *qt.QWidget) *FlyoutView {
	w := &FlyoutView{
		FlyoutViewBase: NewFlyoutViewBase(parent),
		icon:           icon,
		title:          title,
		image:          image,
		content:        content,
		isClosable:     isClosable,
	}
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.viewLayout = qt.NewQHBoxLayout2()
	w.widgetLayout = qt.NewQVBoxLayout2()
	w.titleLabel = qt.NewQLabel5(title, w.QWidget)
	w.contentLabel = qt.NewQLabel5(content, w.QWidget)
	w.iconWidget = NewFlyoutIconWidget(icon, w.QWidget)
	w.imageLabel = NewImageLabel(w.QWidget)
	w.closeButton = NewTransparentToolButtonIcon(common.Cancel, w.QWidget).QToolButton
	w.initWidgets()
	return w
}

func (w *FlyoutView) initWidgets() {
	w.imageLabel.SetImage(w.image)

	w.closeButton.SetFixedSize2(32, 32)
	w.closeButton.SetIconSize(qt.NewQSize2(12, 12))
	w.closeButton.SetVisible(w.isClosable)
	w.titleLabel.SetVisible(w.title != "")
	w.contentLabel.SetVisible(w.content != "")
	w.iconWidget.SetHidden(w.icon == nil)

	w.closeButton.OnClicked(func() {
		if w.onClosed != nil {
			w.onClosed()
		}
	})

	w.titleLabel.SetObjectName("titleLabel")
	w.contentLabel.SetObjectName("contentLabel")
	common.FluentStyleSheet(common.FluentTeachingTip).Apply(w.QWidget, common.ThemeAuto)

	w.initLayout()
}

func (w *FlyoutView) initLayout() {
	w.vBoxLayout.SetContentsMargins(1, 1, 1, 1)
	w.widgetLayout.SetContentsMargins(0, 8, 0, 8)
	w.viewLayout.SetSpacing(4)
	w.widgetLayout.SetSpacing(0)
	w.vBoxLayout.SetSpacing(0)

	if w.title == "" || w.content == "" {
		w.iconWidget.SetFixedHeight(36)
	}

	w.vBoxLayout.AddLayout(w.viewLayout.QLayout)
	w.viewLayout.AddWidget3(w.iconWidget.QWidget, 0, qt.AlignTop)

	w.adjustText()
	w.widgetLayout.AddWidget(w.titleLabel.QWidget)
	w.widgetLayout.AddWidget(w.contentLabel.QWidget)
	w.viewLayout.AddLayout(w.widgetLayout.QLayout)

	w.closeButton.SetVisible(w.isClosable)
	w.viewLayout.AddWidget3(w.closeButton.QWidget, 0, qt.AlignRight|qt.AlignTop)

	left, right := 6, 6
	if w.icon == nil {
		left = 20
	} else {
		left = 5
	}
	if !w.isClosable {
		right = 20
	} else {
		right = 6
	}
	w.viewLayout.SetContentsMargins(left, 5, right, 5)

	w.adjustImage()
	w.addImageToLayout()
}

// AddWidget adds a widget below the title/content.
func (w *FlyoutView) AddWidget(widget *qt.QWidget, stretch int, align qt.AlignmentFlag) {
	w.widgetLayout.AddSpacing(8)
	w.widgetLayout.AddWidget3(widget, stretch, align)
}

// SetOnClosed registers the callback emitted when the close button is clicked.
func (w *FlyoutView) SetOnClosed(f func()) { w.onClosed = f }

func (w *FlyoutView) addImageToLayout() {
	w.imageLabel.SetBorderRadius(8, 8, 0, 0)
	w.imageLabel.SetHidden(w.imageLabel.IsNull())
	w.vBoxLayout.InsertWidget(0, w.imageLabel.QWidget)
}

func (w *FlyoutView) adjustText() {
	screenW := 900
	if g := common.GetCurrentScreenGeometry(false); g != nil {
		screenW = g.Width() - 200
		if screenW > 900 {
			screenW = 900
		}
	}
	chars := max(min(screenW/10, 120), 30)
	title, _ := common.Wrap(w.title, chars, false)
	w.titleLabel.SetText(title)
	chars = max(min(screenW/9, 120), 30)
	content, _ := common.Wrap(w.content, chars, false)
	w.contentLabel.SetText(content)
}

func (w *FlyoutView) adjustImage() {
	sh := w.vBoxLayout.SizeHint()
	w.imageLabel.ScaledToWidth(sh.Width() - 2)
}

// Flyout is a pop-up container that shows a flyout view.
type Flyout struct {
	*qt.QWidget
	view            *FlyoutViewBase
	hBoxLayout      *qt.QHBoxLayout
	shadowEffect    *qt.QGraphicsDropShadowEffect
	isDeleteOnClose bool
	onClosed        func()
}

// NewFlyout builds a flyout for the given view.
func NewFlyout(view *FlyoutViewBase, parent *qt.QWidget, isDeleteOnClose bool) *Flyout {
	w := &Flyout{
		QWidget:         qt.NewQWidget(parent),
		view:            view,
		isDeleteOnClose: isDeleteOnClose,
	}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.hBoxLayout.SetContentsMargins(15, 8, 15, 20)
	w.hBoxLayout.AddWidget(view.QWidget)
	w.SetShadowEffect(35, 0, 8)
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetWindowFlags(qt.Popup | qt.FramelessWindowHint | qt.NoDropShadowWindowHint)
	w.installEvents()
	return w
}

// SetShadowEffect adds a drop shadow behind the view.
func (w *Flyout) SetShadowEffect(blurRadius float64, dx, dy float64) {
	alpha := 30
	if common.IsDarkTheme() {
		alpha = 80
	}
	color := qt.NewQColor11(0, 0, 0, alpha)
	defer color.Delete()
	w.shadowEffect = qt.NewQGraphicsDropShadowEffect2(w.view.QObject)
	w.shadowEffect.SetBlurRadius(blurRadius)
	w.shadowEffect.SetOffset2(dx, dy)
	w.shadowEffect.SetColor(color)
	w.view.SetGraphicsEffect(nil)
	w.view.SetGraphicsEffect(w.shadowEffect.QGraphicsEffect)
}

// OnClosed registers a callback emitted when the flyout closes.
func (w *Flyout) OnClosed(f func()) { w.onClosed = f }

func (w *Flyout) installEvents() {
	w.OnCloseEvent(func(super func(event *qt.QCloseEvent), event *qt.QCloseEvent) {
		if w.isDeleteOnClose {
			w.DeleteLater()
		}
		super(event)
		if w.onClosed != nil {
			w.onClosed()
		}
	})
	w.OnShowEvent(func(super func(event *qt.QShowEvent), event *qt.QShowEvent) {
		w.ActivateWindow()
		super(event)
	})
}

// Exec moves the flyout to the adjusted position and shows it with the
// matching slide + fade animation (a faithful port of FlyoutAnimationManager:
// 187ms OutQuad, 8px slide offset for pull-up / drop-down / slide-left /
// slide-right, fade-only for FADE_IN and no animation for NONE).
func (w *Flyout) Exec(pos *qt.QPoint, aniType FlyoutAnimationType) {
	adjusted := w.adjustedPos(pos)
	defer adjusted.Delete()

	dx, dy := 0, 0
	switch aniType {
	case FlyoutAnimationDropDown:
		dy = -8
	case FlyoutAnimationSlideLeft:
		dx = 8
	case FlyoutAnimationSlideRight:
		dx = -8
	case FlyoutAnimationPullUp:
		dy = 8
	}

	start := qt.NewQPoint2(adjusted.X()+dx, adjusted.Y()+dy)
	defer start.Delete()

	w.SetWindowOpacity(0)
	w.MoveWithQPoint(start)
	w.Show()

	if aniType == FlyoutAnimationNone {
		w.SetWindowOpacity(1)
		return
	}

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	defer curve.Delete()

	common.FadeWindowIn(w.QWidget, 187, curve)

	if aniType != FlyoutAnimationFadeIn {
		end := qt.NewQPoint2(adjusted.X(), adjusted.Y())
		defer end.Delete()
		common.SlidePos(w.QWidget, start, end, 187, curve)
	}
}

func (w *Flyout) adjustedPos(pos *qt.QPoint) *qt.QPoint {
	g := common.GetCurrentScreenGeometry(false)
	if g == nil {
		return qt.NewQPoint2(pos.X(), pos.Y()-4)
	}
	// g is GoGC-armed (QScreen.Geometry) — do NOT Delete

	sh := w.SizeHint()
	wd := sh.Width() + 5
	h := sh.Height()

	x := pos.X()
	if x < g.Left() {
		x = g.Left()
	}
	if x > g.Right()-wd {
		x = g.Right() - wd
	}
	y := pos.Y() - 4
	if y < g.Top() {
		y = g.Top()
	}
	if y > g.Bottom()-h+5 {
		y = g.Bottom() - h + 5
	}
	return qt.NewQPoint2(x, y)
}

// FlyoutMake creates and shows a flyout from a view. target may be a
// *qt.QWidget or *qt.QPoint.
func FlyoutMake(view *FlyoutViewBase, target interface{}, parent *qt.QWidget, aniType FlyoutAnimationType, isDeleteOnClose bool) *Flyout {
	w := NewFlyout(view, parent, isDeleteOnClose)
	if target == nil {
		return w
	}
	w.Show()
	var pos *qt.QPoint
	switch t := target.(type) {
	case *qt.QWidget:
		pos = flyoutPosition(t, w, aniType)
	case *qt.QPoint:
		pos = qt.NewQPoint2(t.X(), t.Y())
	default:
		pos = qt.NewQPoint2(0, 0)
	}
	w.Exec(pos, aniType)
	pos.Delete()
	return w
}

// FlyoutCreate creates and shows a flyout using the default FlyoutView.
func FlyoutCreate(title, content string, icon, image interface{}, isClosable bool, target interface{}, parent *qt.QWidget, aniType FlyoutAnimationType, isDeleteOnClose bool) *Flyout {
	view := NewFlyoutView(title, content, icon, image, isClosable, nil)
	w := FlyoutMake(view.FlyoutViewBase, target, parent, aniType, isDeleteOnClose)
	view.SetOnClosed(func() { w.Close() })
	return w
}

func flyoutPosition(target *qt.QWidget, flyout *Flyout, aniType FlyoutAnimationType) *qt.QPoint {
	sh := flyout.SizeHint()
	fw := sh.Width()
	fh := sh.Height()
	m := flyout.Layout().ContentsMargins()

	switch aniType {
	case FlyoutAnimationDropDown:
		p := qt.NewQPoint2(0, target.Height())
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		x := pos.X() + target.Width()/2 - fw/2
		y := pos.Y() - m.Top() + 8
		return qt.NewQPoint2(x, y)
	case FlyoutAnimationSlideLeft:
		p := qt.NewQPoint2(0, 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		x := pos.X() - fw + 8
		y := pos.Y() - fh/2 + target.Height()/2 + m.Top()
		return qt.NewQPoint2(x, y)
	case FlyoutAnimationSlideRight:
		p := qt.NewQPoint2(0, 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		x := pos.X() + target.Width() - 8
		y := pos.Y() - fh/2 + target.Height()/2 + m.Top()
		return qt.NewQPoint2(x, y)
	default: // FlyoutAnimationPullUp, FlyoutAnimationFadeIn, FlyoutAnimationNone
		p := qt.NewQPoint2(0, 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		x := pos.X() + target.Width()/2 - fw/2
		y := pos.Y() - fh + m.Bottom()
		return qt.NewQPoint2(x, y)
	}
}
