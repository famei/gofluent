package widgets

import (
	"sync"
	"unsafe"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// InfoBarIcon enumerates the four info bar icons. The glyphs live under
// images/info_bar/ and are rendered through readEmbeddedImage.
type InfoBarIcon string

const (
	// InfoBarIconInformation is the information ("i") icon.
	InfoBarIconInformation InfoBarIcon = "Info"
	// InfoBarIconSuccess is the success icon.
	InfoBarIconSuccess InfoBarIcon = "Success"
	// InfoBarIconWarning is the warning icon.
	InfoBarIconWarning InfoBarIcon = "Warning"
	// InfoBarIconError is the error icon.
	InfoBarIconError InfoBarIcon = "Error"
)

func infoBarIconColorSuffix() string {
	if common.IsDarkTheme() {
		return "dark"
	}
	return "light"
}

// render draws the info bar icon SVG into painter at rect.
func (i InfoBarIcon) render(painter *qt.QPainter, rect *qt.QRectF) {
	b := readEmbeddedImage("info_bar/" + string(i) + "_" + infoBarIconColorSuffix() + ".svg")
	drawSvgBytes(b, painter, rect)
}

// drawInfoBarIconOrGeneric renders an icon source that may be an InfoBarIcon
// (the information glyph is recolored with the theme color, matching the
// InfoIconWidget) or any other FluentIcon / QIcon / string source. It is shared
// by the flyout / teaching-tip icon widget and InfoIconWidget so that
// InfoBarIcon values render correctly when used as a Flyout/TeachingTip icon.
func drawInfoBarIconOrGeneric(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	if ib, ok := icon.(InfoBarIcon); ok {
		if ib == InfoBarIconInformation {
			color := common.ThemeColorPrimary.Color()
			defer color.Delete()
			b := readEmbeddedImage("info_bar/Info_" + infoBarIconColorSuffix() + ".svg")
			drawSvgBytes([]byte(common.RecolorSvg(string(b), color.Name())), painter, rect)
		} else {
			ib.render(painter, rect)
		}
		return
	}
	renderFluentIcon(icon, painter, rect, common.ThemeAuto)
}

// InfoBarPosition enumerates the display position of an info bar.
type InfoBarPosition int

const (
	InfoBarPositionTop InfoBarPosition = iota
	InfoBarPositionBottom
	InfoBarPositionTopLeft
	InfoBarPositionTopRight
	InfoBarPositionBottomLeft
	InfoBarPositionBottomRight
	InfoBarPositionNone
)

// InfoIconWidget paints the info bar icon centered in a 36x36 area.
type InfoIconWidget struct {
	*qt.QWidget
	icon interface{}
}

// NewInfoIconWidget builds an info icon widget.
func NewInfoIconWidget(icon interface{}, parent *qt.QWidget) *InfoIconWidget {
	w := &InfoIconWidget{QWidget: qt.NewQWidget(parent), icon: icon}
	w.SetFixedSize2(36, 36)
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

		rect := qt.NewQRectF4(10, 10, 15, 15)
		defer rect.Delete()

		drawInfoBarIconOrGeneric(w.icon, painter, rect)
		painter.End()
	})
	return w
}

// InfoBar is a transient information bar with a title, a content label and an
// optional close button. On show the bar slides into its target position (200ms
// OutQuad) and fades out on close.
type InfoBar struct {
	*qt.QFrame
	title      string
	content    string
	orient     qt.Orientation
	icon       interface{}
	duration   int
	isClosable bool
	position   InfoBarPosition

	titleLabel   *qt.QLabel
	contentLabel *qt.QLabel
	closeButton  *qt.QToolButton
	iconWidget   *InfoIconWidget

	hBoxLayout   *qt.QHBoxLayout
	textLayout   *qt.QBoxLayout
	widgetLayout *qt.QBoxLayout

	opacityEffect *qt.QGraphicsOpacityEffect
	opacityAni    *qt.QPropertyAnimation
	timer         *qt.QTimer

	// dropAni is the reusable reposition animation (port of InfoBarManager's
	// per-bar "drop" QPropertyAnimation on `pos`). When an earlier info bar is
	// closed, the bars stacked below it slide up into the vacated slot instead
	// of jumping instantly.
	dropAni *qt.QPropertyAnimation

	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor

	OnClosed func()
}

// DesktopInfoBarView is the full-screen container that hosts desktop
// (parentless) info bars, mirroring the Python DesktopInfoBarView. Giving a
// desktop notification a parent lets the info bar manager stack it and lets its
// QSS background render normally.
type DesktopInfoBarView struct {
	*qt.QWidget
}

var desktopInfoBarViewOnce sync.Once
var desktopInfoBarView *DesktopInfoBarView

// getDesktopInfoBarView returns (creating on first use) the desktop container.
func getDesktopInfoBarView() *DesktopInfoBarView {
	desktopInfoBarViewOnce.Do(func() {
		v := &DesktopInfoBarView{QWidget: qt.NewQWidget2()}
		v.SetWindowFlags(qt.FramelessWindowHint | qt.WindowStaysOnTopHint | qt.Tool)
		v.SetAttribute(qt.WA_TransparentForMouseEvents)
		v.SetAttribute(qt.WA_TranslucentBackground)

		g := common.GetCurrentScreenGeometry(true) // GoGC-armed — do NOT Delete
		v.SetGeometryWithGeometry(g)

		v.Show()
		desktopInfoBarView = v
	})
	return desktopInfoBarView
}

// NewInfoBar builds an info bar.
func NewInfoBar(icon interface{}, title, content string, orient qt.Orientation, isClosable bool, duration int, position InfoBarPosition, parent *qt.QWidget) *InfoBar {
	// Desktop (parentless) info bars are hosted by the full-screen desktop view
	// container so they get a parent (port of InfoBar.desktopView()).
	if parent == nil {
		parent = getDesktopInfoBarView().QWidget
	}

	b := &InfoBar{
		QFrame:     qt.NewQFrame(parent),
		title:      title,
		content:    content,
		orient:     orient,
		icon:       icon,
		duration:   duration,
		isClosable: isClosable,
		position:   position,
	}

	b.titleLabel = qt.NewQLabel(b.QWidget)
	b.contentLabel = qt.NewQLabel(b.QWidget)
	b.closeButton = NewTransparentToolButtonIcon(common.Cancel, b.QWidget).QToolButton
	b.iconWidget = NewInfoIconWidget(icon, nil)

	b.opacityEffect = qt.NewQGraphicsOpacityEffect2(b.QObject)
	b.opacityAni = qt.NewQPropertyAnimation2(b.opacityEffect.QObject, []byte("opacity"))
	b.timer = qt.NewQTimer2(b.QObject)
	b.timer.SetSingleShot(true)
	b.timer.OnTimeout(b.fadeOut)

	b.initWidget()
	return b
}

func (b *InfoBar) initWidget() {
	b.opacityEffect.SetOpacity(1)
	b.SetGraphicsEffect(b.opacityEffect.QGraphicsEffect)

	b.closeButton.SetFixedSize2(36, 36)
	size := qt.NewQSize2(12, 12)
	b.closeButton.SetIconSize(size)
	size.Delete()
	b.closeButton.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))
	b.closeButton.SetVisible(b.isClosable)

	b.setQss()
	b.initLayout()
	b.closeButton.OnClicked(func() { b.Close() })

	b.OnShowEvent(func(super func(event *qt.QShowEvent), event *qt.QShowEvent) {
		b.adjustText()
		super(event)

		if b.duration >= 0 {
			b.timer.Start(b.duration)
		}
	})
	b.OnCloseEvent(func(super func(event *qt.QCloseEvent), event *qt.QCloseEvent) {
		event.Ignore()
		if b.OnClosed != nil {
			b.OnClosed()
		}
		defaultInfoBarManager.remove(b)
		b.DeleteLater()
	})
	b.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		if b.lightBackgroundColor == nil {
			return
		}
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		var color *qt.QColor
		if common.IsDarkTheme() {
			color = cloneColor(b.darkBackgroundColor)
		} else {
			color = cloneColor(b.lightBackgroundColor)
		}
		defer color.Delete()
		brush := qt.NewQBrush3(color)
		defer brush.Delete()
		painter.SetBrush(brush)

		rect := qt.NewQRectF4(1, 1, float64(b.Width()-2), float64(b.Height()-2))
		defer rect.Delete()
		painter.DrawRoundedRect(rect, 6, 6)
		painter.End()
	})
}

// Show displays the info bar. For managed positions the bar is first moved to
// its off-screen slide start point and only then shown, after which it slides
// into place. Positioning inside showEvent is too late: the widget is already
// visible when the event is delivered, so its first frame can still be painted
// at the pre-show (0,0) geometry — the one-frame "window flash". Pre-positioning
// before Show() mirrors RoundMenu.Exec, which also Moves to the slide start
// before calling Show().
func (b *InfoBar) Show() {
	b.adjustText()

	if b.position == InfoBarPositionNone {
		b.QFrame.Show()
		return
	}

	defaultInfoBarManager.add(b)
	x, y := defaultInfoBarManager.pos(b)
	sx, sy := b.slideStartPosition(x, y)

	// Anchor the widget off-screen before it becomes visible so the first
	// painted frame is already at the animation start.
	b.Move(sx, sy)

	b.QFrame.Show()

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	slideStart := qt.NewQPoint2(sx, sy)
	slideEnd := qt.NewQPoint2(x, y)
	common.SlidePos(b.QWidget, slideStart, slideEnd, 200, curve)
	curve.Delete()
	slideStart.Delete()
	slideEnd.Delete()
}

func (b *InfoBar) setQss() {
	b.SetObjectName("InfoBar")
	b.titleLabel.SetObjectName("titleLabel")
	b.contentLabel.SetObjectName("contentLabel")
	if ib, ok := b.icon.(InfoBarIcon); ok {
		b.SetProperty("type", qt.NewQVariant14(string(ib)))
	}
	common.FluentStyleSheet(common.FluentInfoBar).Apply(b.QWidget, common.ThemeAuto)
}

func (b *InfoBar) initLayout() {
	b.hBoxLayout = qt.NewQHBoxLayout(b.QWidget)
	if b.orient == qt.Horizontal {
		b.textLayout = qt.NewQHBoxLayout2().QBoxLayout
		b.widgetLayout = qt.NewQHBoxLayout2().QBoxLayout
	} else {
		b.textLayout = qt.NewQVBoxLayout2().QBoxLayout
		b.widgetLayout = qt.NewQVBoxLayout2().QBoxLayout
	}

	b.hBoxLayout.SetContentsMargins(6, 6, 6, 6)
	b.hBoxLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)
	b.textLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)
	b.textLayout.SetContentsMargins(1, 8, 0, 8)

	b.hBoxLayout.SetSpacing(0)
	b.textLayout.SetSpacing(5)

	b.hBoxLayout.AddWidget3(b.iconWidget.QWidget, 0, qt.AlignTop|qt.AlignLeft)

	b.textLayout.AddWidget3(b.titleLabel.QWidget, 1, qt.AlignTop)
	b.titleLabel.SetVisible(b.title != "")

	if b.orient == qt.Horizontal {
		b.textLayout.AddSpacing(7)
	}
	b.textLayout.AddWidget3(b.contentLabel.QWidget, 1, qt.AlignTop)
	b.contentLabel.SetVisible(b.content != "")
	b.hBoxLayout.AddLayout(b.textLayout.QLayout)

	if b.orient == qt.Horizontal {
		b.hBoxLayout.AddLayout(b.widgetLayout.QLayout)
		b.widgetLayout.SetSpacing(10)
	} else {
		b.textLayout.AddLayout(b.widgetLayout.QLayout)
	}

	b.hBoxLayout.AddSpacing(12)
	b.hBoxLayout.AddWidget3(b.closeButton.QWidget, 0, qt.AlignTop|qt.AlignLeft)

	b.adjustText()
}

func (b *InfoBar) fadeOut() {
	b.opacityAni.SetDuration(200)
	v := qt.NewQVariant12(1.0)
	b.opacityAni.SetStartValue(v)
	v.Delete()
	v = qt.NewQVariant12(0.0)
	b.opacityAni.SetEndValue(v)
	v.Delete()
	b.opacityAni.OnFinished(func() { b.Close() })
	b.opacityAni.Start()
}

func (b *InfoBar) adjustText() {
	w := 900
	if b.ParentWidget() != nil {
		w = b.ParentWidget().Width() - 50
	}

	titleChars := w / 10
	if titleChars > 120 {
		titleChars = 120
	}
	if titleChars < 30 {
		titleChars = 30
	}
	contentChars := w / 9
	if contentChars > 120 {
		contentChars = 120
	}
	if contentChars < 30 {
		contentChars = 30
	}

	t, _ := common.Wrap(b.title, titleChars, false)
	b.titleLabel.SetText(t)
	c, _ := common.Wrap(b.content, contentChars, false)
	b.contentLabel.SetText(c)
	b.AdjustSize()
}

// AddWidget adds a custom widget to the info bar.
func (b *InfoBar) AddWidget(widget *qt.QWidget, stretch int) {
	b.widgetLayout.AddSpacing(6)
	align := qt.AlignVCenter
	if b.orient == qt.Vertical {
		align = qt.AlignTop
	}
	b.widgetLayout.AddWidget3(widget, stretch, qt.AlignLeft|align)
}

// SetCustomBackgroundColor sets the custom background color for light/dark mode.
func (b *InfoBar) SetCustomBackgroundColor(light, dark *qt.QColor) {
	b.lightBackgroundColor = cloneColor(light)
	b.darkBackgroundColor = cloneColor(dark)
	b.Update()
}

// slideStartPosition returns the off-screen start position of the slide-in
// animation (port of InfoBarManager._slideStartPos for each position).
func (b *InfoBar) slideStartPosition(x, y int) (int, int) {
	// Fall back to the screen geometry for top-level bars (nil parent, e.g. the
	// desktop notification case) so the start position stays off-screen.
	pw := 0
	if p := b.ParentWidget(); p != nil {
		pw = p.Width()
	} else if screen := common.GetCurrentScreenGeometry(false); screen != nil {
		// screen is GoGC-armed (QScreen.Geometry) — do NOT Delete
		pw = screen.Width()
	}

	switch b.position {
	case InfoBarPositionTop:
		return x, y - 16
	case InfoBarPositionTopRight, InfoBarPositionBottomRight:
		return pw, y
	case InfoBarPositionTopLeft, InfoBarPositionBottomLeft:
		return -b.Width(), y
	case InfoBarPositionBottom:
		return x, y + 16
	default:
		return x, y
	}
}

// infoBarManagerKey groups live info bars by their parent widget and display
// position so that each corner (and the top/bottom centers) keeps an
// independent stack and count. This mirrors InfoBarManager.make(position) in
// info_bar.py, where every position has its own manager and its own
// parent-keyed infoBars dictionary.
type infoBarManagerKey struct {
	parent   unsafe.Pointer
	position InfoBarPosition
}

// infoBarManager tracks the live info bars per (parent widget, position) pair
// so multiple bars stack vertically instead of overlapping at the same anchor
// (port of the InfoBarManager in info_bar.py). spacing is the gap between
// stacked bars and margin is the distance to the parent's edge.
type infoBarManager struct {
	spacing int
	margin  int
	bars    map[infoBarManagerKey][]*InfoBar
}

var defaultInfoBarManager = &infoBarManager{
	spacing: 16,
	margin:  24,
	bars:    map[infoBarManagerKey][]*InfoBar{},
}

func (m *infoBarManager) key(b *InfoBar) (infoBarManagerKey, bool) {
	p := b.ParentWidget()
	if p == nil {
		return infoBarManagerKey{}, false
	}
	return infoBarManagerKey{parent: p.UnsafePointer(), position: b.position}, true
}

// add registers an info bar so that bars shown after it stack below (top
// positions) or above (bottom positions) it.
func (m *infoBarManager) add(b *InfoBar) {
	key, ok := m.key(b)
	if !ok {
		return
	}
	for _, bar := range m.bars[key] {
		if bar == b {
			return
		}
	}
	// A bar shown while other bars are already visible will later slide into a
	// new stacked position when one of the earlier bars closes, so give it a
	// reusable drop (reposition) animation now (InfoBarManager.add).
	if len(m.bars[key]) > 0 && b.dropAni == nil {
		b.dropAni = qt.NewQPropertyAnimation4(b.QObject, []byte("pos"), b.QObject)
		b.dropAni.SetDuration(200)
	}
	m.bars[key] = append(m.bars[key], b)
}

// remove unregisters an info bar and repositions the remaining bars to close
// the gap it left behind.
func (m *infoBarManager) remove(b *InfoBar) {
	key, ok := m.key(b)
	if !ok {
		return
	}
	list := m.bars[key]
	removed := false
	for i, bar := range list {
		if bar == b {
			m.bars[key] = append(list[:i], list[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		return
	}
	if len(m.bars[key]) == 0 {
		delete(m.bars, key)
	} else {
		m.reposition(key)
	}
}

func (m *infoBarManager) indexOf(b *InfoBar) int {
	key, ok := m.key(b)
	if !ok {
		return -1
	}
	for i, bar := range m.bars[key] {
		if bar == b {
			return i
		}
	}
	return -1
}

// pos returns the target position of the info bar, stacking it relative to the
// bars already shown for the same parent.
func (m *infoBarManager) pos(b *InfoBar) (int, int) {
	pw, ph := 0, 0
	if p := b.ParentWidget(); p != nil {
		pw, ph = p.Width(), p.Height()
	} else if screen := common.GetCurrentScreenGeometry(false); screen != nil {
		// screen is GoGC-armed (QScreen.Geometry) — do NOT Delete
		pw, ph = screen.Width(), screen.Height()
	}

	w, h := b.Width(), b.Height()
	index := m.indexOf(b)

	var previous []*InfoBar
	if key, ok := m.key(b); ok {
		list := m.bars[key]
		if index >= 0 && index <= len(list) {
			previous = list[:index]
		}
	}

	sumHeight := func() int {
		y := 0
		for _, bar := range previous {
			y += bar.Height() + m.spacing
		}
		return y
	}

	switch b.position {
	case InfoBarPositionTop:
		return (pw - w) / 2, m.margin + sumHeight()
	case InfoBarPositionTopRight:
		return pw - w - m.margin, m.margin + sumHeight()
	case InfoBarPositionTopLeft:
		return m.margin, m.margin + sumHeight()
	case InfoBarPositionBottom:
		return (pw - w) / 2, ph - h - m.margin - sumHeight()
	case InfoBarPositionBottomRight:
		return pw - w - m.margin, ph - h - m.margin - sumHeight()
	case InfoBarPositionBottomLeft:
		return m.margin, ph - h - m.margin - sumHeight()
	default:
		return b.X(), b.Y()
	}
}

// reposition moves every remaining info bar of a parent to its stacked
// position after one of them was removed. Bars that have a drop animation
// (every bar except the first for a given parent) animate from their current
// position to the new one over 200ms, matching the reference InfoBarManager
// drop animation, instead of teleporting.
func (m *infoBarManager) reposition(key infoBarManagerKey) {
	for _, bar := range m.bars[key] {
		x, y := m.pos(bar)
		if bar.dropAni == nil {
			bar.Move(x, y)
			continue
		}

		bar.dropAni.Stop()
		start := qt.NewQPoint2(bar.X(), bar.Y())
		end := qt.NewQPoint2(x, y)

		v := qt.NewQVariant27(start)
		bar.dropAni.SetStartValue(v)
		v.Delete()

		v = qt.NewQVariant27(end)
		bar.dropAni.SetEndValue(v)
		v.Delete()

		start.Delete()
		end.Delete()
		bar.dropAni.Start()
	}
}

func showInfoBar(icon interface{}, title, content string, parent *qt.QWidget) *InfoBar {
	b := NewInfoBar(icon, title, content, qt.Horizontal, true, 1000, InfoBarPositionTopRight, parent)
	b.Show()
	return b
}

// InfoBarInfo creates and shows an information info bar.
func InfoBarInfo(title, content string, parent *qt.QWidget) *InfoBar {
	return showInfoBar(InfoBarIconInformation, title, content, parent)
}

// InfoBarSuccess creates and shows a success info bar.
func InfoBarSuccess(title, content string, parent *qt.QWidget) *InfoBar {
	return showInfoBar(InfoBarIconSuccess, title, content, parent)
}

// InfoBarWarning creates and shows a warning info bar.
func InfoBarWarning(title, content string, parent *qt.QWidget) *InfoBar {
	return showInfoBar(InfoBarIconWarning, title, content, parent)
}

// InfoBarError creates and shows an error info bar.
func InfoBarError(title, content string, parent *qt.QWidget) *InfoBar {
	return showInfoBar(InfoBarIconError, title, content, parent)
}
