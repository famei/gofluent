package window

import (
	"runtime"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/internal/platform"
	qt "github.com/mappu/miqt/qt"
)

// ---------------------------------------------------------------------------
// Title bar buttons
//
// The Python title bar buttons come from the external qframelesswindow package
// (MinimizeButton / MaximizeButton / CloseButton). Each is a flat 46x32 surface
// with a theme-colored hover/pressed background and a hand-drawn system glyph:
// a horizontal line, a maximize square (or the restore pair when maximized) and
// an X. The glyphs are drawn directly here instead of reusing the Fluent SVG
// icon set, which has a different visual weight from the reference.

// titleBarButtonKind identifies which system glyph a title bar button draws.
type titleBarButtonKind int

const (
	titleBarButtonMinimize titleBarButtonKind = iota
	titleBarButtonMaximize
	titleBarButtonClose
)

// TitleBarButton is a custom-painted title bar button (minimize/maximize/close).
type TitleBarButton struct {
	*qt.QWidget
	kind      titleBarButtonKind
	isHover   bool
	isPressed bool
	onClicked func()
}

// NewTitleBarButton builds an empty title bar button (minimize glyph by default).
func NewTitleBarButton(parent *qt.QWidget) *TitleBarButton {
	b := &TitleBarButton{QWidget: qt.NewQWidget(parent)}
	// qframelesswindow's TitleBarButton is 46x32; without an explicit size an
	// empty custom-painted widget collapses to zero width inside the layout.
	b.SetFixedWidth(46)
	b.SetFixedHeight(32)
	// Mark the button so FramelessWindow can exclude it from the drag area and
	// let the click reach the widget (see widgets.FramelessWindow.nativeHitTest).
	v := qt.NewQVariant11(true)
	b.SetProperty("isTitleBarButton", v)
	v.Delete()
	b.installEvents()
	return b
}

// OnClicked registers the clicked callback (replaces the Python clicked signal).
func (b *TitleBarButton) OnClicked(f func()) { b.onClicked = f }

func (b *TitleBarButton) installEvents() {
	b.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		b.isHover = true
		b.Update()
		super(e)
	})
	b.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		b.isHover = false
		b.isPressed = false
		b.Update()
		super(e)
	})
	b.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		b.isPressed = true
		b.Update()
		super(e)
	})
	b.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		b.isPressed = false
		b.Update()
		super(e)
		if b.onClicked != nil {
			b.onClicked()
		}
	})
	b.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		b.paint()
	})
}

// backgroundColor returns the button background color for the current state.
// The colors mirror qframelesswindow's title bar button QSS (fluent_window.qss):
// hover rgba(0,0,0,26) / rgba(255,255,255,26), pressed rgba(0,0,0,51) /
// rgba(255,255,255,51), and the close button hover rgb(232,17,35) / pressed
// rgb(241,112,122).
func (b *TitleBarButton) backgroundColor() *qt.QColor {
	if b.kind == titleBarButtonClose && (b.isHover || b.isPressed) {
		if b.isPressed {
			return qt.NewQColor3(241, 112, 122)
		}
		return qt.NewQColor3(232, 17, 35)
	}
	if b.isPressed {
		if common.IsDarkTheme() {
			return qt.NewQColor11(255, 255, 255, 51)
		}
		return qt.NewQColor11(0, 0, 0, 51)
	}
	if b.isHover {
		if common.IsDarkTheme() {
			return qt.NewQColor11(255, 255, 255, 26)
		}
		return qt.NewQColor11(0, 0, 0, 26)
	}
	return qt.NewQColor11(0, 0, 0, 0)
}

// iconColor returns the glyph color for the current state (mirrors the
// MinimizeButton/MaximizeButton/CloseButton QSS: black in light theme, white in
// dark theme, and always white for a hovered/pressed close button).
func (b *TitleBarButton) iconColor() *qt.QColor {
	if b.kind == titleBarButtonClose && (b.isHover || b.isPressed) {
		return qt.NewQColor3(255, 255, 255)
	}
	if common.IsDarkTheme() {
		return qt.NewQColor3(255, 255, 255)
	}
	return qt.NewQColor3(0, 0, 0)
}

func (b *TitleBarButton) paint() {
	painter := qt.NewQPainter2(b.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	bg := b.backgroundColor()
	brush := qt.NewQBrush3(bg)
	painter.SetBrush(brush)
	brush.Delete()
	bg.Delete()
	painter.SetPenWithStyle(qt.NoPen)

	rect := b.Rect()
	painter.DrawRect2(rect.X(), rect.Y(), rect.Width(), rect.Height())

	// Draw the system glyph with a 1px cosmetic pen (matches qframelesswindow).
	color := b.iconColor()
	pen := qt.NewQPen3(color)
	color.Delete()
	defer pen.Delete()
	pen.SetCosmetic(true)
	painter.SetPenWithPen(pen)
	painter.SetBrushWithStyle(qt.NoBrush)

	switch b.kind {
	case titleBarButtonMinimize:
		painter.DrawLine2(18, 16, 28, 16)
	case titleBarButtonMaximize:
		b.drawMaximizeGlyph(painter)
	case titleBarButtonClose:
		painter.DrawLine2(18, 11, 28, 21)
		painter.DrawLine2(28, 11, 18, 21)
	}
	painter.End()
}

// drawMaximizeGlyph draws the maximize square, or the two-overlapping-squares
// restore glyph when the window is maximized (port of qframelesswindow's
// MaximizeButton.paintEvent, including the device-pixel-ratio scaling).
func (b *TitleBarButton) drawMaximizeGlyph(painter *qt.QPainter) {
	r := b.DevicePixelRatioF()
	painter.Scale(1/r, 1/r)
	if !b.isMaximized() {
		painter.DrawRect2(int(18*r), int(11*r), int(10*r), int(10*r))
		return
	}

	painter.DrawRect2(int(18*r), int(13*r), int(8*r), int(8*r))
	x0 := int(18*r) + int(2*r)
	y0 := 13 * r
	dw := int(2 * r)
	path := qt.NewQPainterPath()
	defer path.Delete()
	path.MoveTo2(float64(x0), y0)
	path.LineTo2(float64(x0), y0-float64(dw))
	path.LineTo2(float64(x0)+8*r, y0-float64(dw))
	path.LineTo2(float64(x0)+8*r, y0-float64(dw)+8*r)
	path.LineTo2(float64(x0)+8*r-float64(dw), y0-float64(dw)+8*r)
	painter.DrawPath(path)
}

// isMaximized reports whether the owning window is currently maximized.
func (b *TitleBarButton) isMaximized() bool {
	win := b.Window()
	return win != nil && win.IsMaximized()
}

// FluentTitleBarButton is an alias kept for API parity with the Python export.
type FluentTitleBarButton = TitleBarButton

// ---------------------------------------------------------------------------
// Title bar
//
// TitleBar re-implements the subset of qframelesswindow.TitleBar used by the
// Fluent window classes: an hBoxLayout carrying the three system buttons.

// TitleBar is the base custom title bar.
type TitleBar struct {
	*qt.QWidget
	hBoxLayout *qt.QHBoxLayout
	minBtn     *TitleBarButton
	maxBtn     *TitleBarButton
	closeBtn   *TitleBarButton
}

// NewTitleBar builds a title bar with minimize/maximize/close buttons wired to
// the containing window.
func NewTitleBar(parent *qt.QWidget) *TitleBar {
	t := &TitleBar{QWidget: qt.NewQWidget(parent)}
	// qframelesswindow's TitleBarBase does `self.resize(200, 32)` before fixing
	// the height so the bar has a usable initial width; without it an empty
	// custom-painted bar can collapse to zero width and hide its system buttons.
	t.Resize(200, 32)
	t.SetFixedHeight(32)
	t.hBoxLayout = qt.NewQHBoxLayout(t.QWidget)
	t.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	t.hBoxLayout.SetSpacing(0)

	t.minBtn = NewTitleBarButton(t.QWidget)
	t.maxBtn = NewTitleBarButton(t.QWidget)
	t.closeBtn = NewTitleBarButton(t.QWidget)
	t.minBtn.kind = titleBarButtonMinimize
	t.maxBtn.kind = titleBarButtonMaximize
	t.closeBtn.kind = titleBarButtonClose

	t.minBtn.OnClicked(func() { t.Window().ShowMinimized() })
	t.maxBtn.OnClicked(func() {
		win := t.Window()
		if win.IsMaximized() {
			win.ShowNormal()
		} else {
			win.ShowMaximized()
		}
	})
	t.closeBtn.OnClicked(func() { t.Window().Close() })

	// The stretch keeps the system buttons pinned to the right edge; the icon
	// and title labels are inserted before it by FluentTitleBar/SplitTitleBar.
	//
	// The buttons are anchored to the top: they are a fixed 46x32, and inside a
	// taller custom title bar an unaligned layout item would centre them
	// vertically (the buttons ended up halfway down a 70px bar instead of in the
	// window's top-right corner).
	t.hBoxLayout.AddStretchWithStretch(1)
	t.hBoxLayout.AddWidget3(t.minBtn.QWidget, 0, qt.AlignTop)
	t.hBoxLayout.AddWidget3(t.maxBtn.QWidget, 0, qt.AlignTop)
	t.hBoxLayout.AddWidget3(t.closeBtn.QWidget, 0, qt.AlignTop)
	return t
}

// HBoxLayout returns the title bar's horizontal layout so windows/examples can
// insert custom widgets (e.g. a TabBar) between the title and the system buttons.
func (t *TitleBar) HBoxLayout() *qt.QHBoxLayout { return t.hBoxLayout }

// FluentTitleBar lays out the icon, title and system buttons in Fluent style.
type FluentTitleBar struct {
	*TitleBar
	iconLabel    *qt.QLabel
	titleLabel   *widgets.CaptionLabel
	vBoxLayout   *qt.QVBoxLayout
	buttonLayout *qt.QHBoxLayout

	// Title icon source and the state of its theme hook (see title_icon.go).
	iconSource common.FluentIconBase
	iconHooked bool
}

// NewFluentTitleBar builds a Fluent title bar.
func NewFluentTitleBar(parent *qt.QWidget) *FluentTitleBar {
	t := &FluentTitleBar{TitleBar: NewTitleBar(parent)}
	t.SetFixedHeight(48)
	// Object name matched by fluent_window.qss `FluentTitleBar>QLabel#titleLabel`
	// (translated to `QWidget#fluentTitleBar>QLabel#titleLabel` in style_sheet.go)
	// so the title label gets the reference's 13px font and 4px padding.
	t.SetObjectName("fluentTitleBar")
	t.hBoxLayout.RemoveWidget(t.minBtn.QWidget)
	t.hBoxLayout.RemoveWidget(t.maxBtn.QWidget)
	t.hBoxLayout.RemoveWidget(t.closeBtn.QWidget)

	t.iconLabel = qt.NewQLabel(t.QWidget)
	t.iconLabel.SetObjectName("iconLabel")
	t.iconLabel.SetFixedSize2(18, 18)
	t.hBoxLayout.InsertWidget3(0, t.iconLabel.QWidget, 0, qt.AlignLeft|qt.AlignVCenter)
	t.Window().OnWindowIconChanged(func(icon *qt.QIcon) { t.SetIcon(icon) })

	t.titleLabel = widgets.NewCaptionLabel(t.QWidget)
	t.hBoxLayout.InsertWidget3(1, t.titleLabel.QWidget, 0, qt.AlignLeft|qt.AlignVCenter)
	t.titleLabel.SetObjectName("titleLabel")
	t.Window().OnWindowTitleChanged(func(title string) { t.SetTitle(title) })

	t.vBoxLayout = qt.NewQVBoxLayout2()
	t.buttonLayout = qt.NewQHBoxLayout2()
	t.buttonLayout.SetSpacing(0)
	t.buttonLayout.SetContentsMargins(0, 0, 0, 0)
	// Keep the system buttons pinned to the top edge of the 48px title bar
	// (mirrors qframelesswindow's `buttonLayout.setAlignment(Qt.AlignTop)`).
	// This must be QLayoutItem::setAlignment on the buttonLayout itself; the
	// QLayout::setAlignment(QLayout*, ...) overload is a no-op while the child
	// layout has not been added yet.
	t.buttonLayout.QLayoutItem.SetAlignment(qt.AlignTop)
	t.buttonLayout.AddWidget(t.minBtn.QWidget)
	t.buttonLayout.AddWidget(t.maxBtn.QWidget)
	t.buttonLayout.AddWidget(t.closeBtn.QWidget)
	t.vBoxLayout.AddLayout(t.buttonLayout.QLayout)
	t.vBoxLayout.AddStretchWithStretch(1)
	t.hBoxLayout.AddLayout2(t.vBoxLayout.QLayout, 0)

	common.FluentStyleSheet(common.FluentWindow).Apply(t.QWidget, common.ThemeAuto)
	return t
}

// SetTitle updates the title label.
func (t *FluentTitleBar) SetTitle(title string) {
	t.titleLabel.SetText(title)
	t.titleLabel.AdjustSize()
}

// SetIcon updates the icon label.
func (t *FluentTitleBar) SetIcon(icon *qt.QIcon) {
	pm := icon.Pixmap2(18, 18) // GoGC-armed — do NOT Delete
	t.iconLabel.SetPixmap(pm)
}

// MSFluentTitleBar adds Microsoft Store style spacing to the title bar.
type MSFluentTitleBar struct{ *FluentTitleBar }

// NewMSFluentTitleBar builds a Microsoft Store style title bar.
func NewMSFluentTitleBar(parent *qt.QWidget) *MSFluentTitleBar {
	t := &MSFluentTitleBar{FluentTitleBar: NewFluentTitleBar(parent)}
	// Mark the bar so fluent_window.qss `MSFluentTitleBar>QLabel#titleLabel`
	// (translated to `QWidget#fluentTitleBar[isMsTitleBar=true]>QLabel#titleLabel`)
	// applies the Microsoft Store 10px title padding. Re-polish so the attribute
	// selector is picked up after the base FluentTitleBar already applied the QSS.
	v := qt.NewQVariant11(true)
	t.SetProperty("isMsTitleBar", v)
	v.Delete()
	t.SetStyle(qt.QApplication_Style())
	t.hBoxLayout.InsertSpacing(0, 20)
	t.hBoxLayout.InsertSpacing(2, 2)
	return t
}

// FluentWidgetTitleBar is the title bar used by the plain FluentWidget window.
type FluentWidgetTitleBar struct{ *FluentTitleBar }

// NewFluentWidgetTitleBar builds a widget title bar.
func NewFluentWidgetTitleBar(parent *qt.QWidget) *FluentWidgetTitleBar {
	t := &FluentWidgetTitleBar{FluentTitleBar: NewFluentTitleBar(parent)}
	if runtime.GOOS == "darwin" {
		t.iconLabel.Hide()
		t.titleLabel.Hide()
		t.SetFixedHeight(28)
	} else {
		t.hBoxLayout.SetContentsMargins(16, 0, 0, 0)
		sh := t.buttonLayout.SizeHint()
		t.SetFixedHeight(sh.Height())
	}
	common.FluentStyleSheet(common.FluentWindow).Apply(t.minBtn.QWidget, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentWindow).Apply(t.maxBtn.QWidget, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentWindow).Apply(t.closeBtn.QWidget, common.ThemeAuto)
	return t
}

// SplitTitleBar is a title bar with icon and title aligned to the bottom.
type SplitTitleBar struct {
	*TitleBar
	iconLabel  *qt.QLabel
	titleLabel *qt.QLabel

	// Title icon source and the state of its theme hook (see title_icon.go).
	iconSource common.FluentIconBase
	iconHooked bool
}

// NewSplitTitleBar builds a split title bar.
func NewSplitTitleBar(parent *qt.QWidget) *SplitTitleBar {
	t := &SplitTitleBar{TitleBar: NewTitleBar(parent)}
	// The reference SplitTitleBar is 48px high (the base TitleBar is 32px): the split
	// style keeps its icon and title bottom-aligned on purpose, and in a 32px bar that
	// leaves them squeezed into the bottom-left corner.
	t.SetFixedHeight(48)
	// Object name matched by fluent_window.qss `SplitTitleBar>QLabel#titleLabel`
	// (translated to `QWidget#splitTitleBar>QLabel#titleLabel` in style_sheet.go)
	// so the title label gets the reference's 13px font and 5px padding.
	t.SetObjectName("splitTitleBar")
	t.iconLabel = qt.NewQLabel(t.QWidget)
	t.iconLabel.SetObjectName("iconLabel")
	t.iconLabel.SetFixedSize2(18, 18)
	t.hBoxLayout.InsertSpacing(0, 12)
	t.hBoxLayout.InsertWidget3(1, t.iconLabel.QWidget, 0, qt.AlignLeft|qt.AlignBottom)
	t.Window().OnWindowIconChanged(func(icon *qt.QIcon) { t.SetIcon(icon) })

	t.titleLabel = qt.NewQLabel(t.QWidget)
	t.hBoxLayout.InsertWidget3(2, t.titleLabel.QWidget, 0, qt.AlignLeft|qt.AlignBottom)
	t.titleLabel.SetObjectName("titleLabel")
	t.Window().OnWindowTitleChanged(func(title string) { t.SetTitle(title) })

	common.FluentStyleSheet(common.FluentWindow).Apply(t.QWidget, common.ThemeAuto)
	return t
}

// SetTitle updates the title label.
func (t *SplitTitleBar) SetTitle(title string) {
	t.titleLabel.SetText(title)
	t.titleLabel.AdjustSize()
}

// SetIcon updates the icon label.
func (t *SplitTitleBar) SetIcon(icon *qt.QIcon) {
	pm := icon.Pixmap2(18, 18) // GoGC-armed — do NOT Delete
	t.iconLabel.SetPixmap(pm)
}

// ---------------------------------------------------------------------------
// FluentWidget / FluentWindowBase

// FluentWidget is a frameless window with a Fluent title bar and a solid
// light/dark background. On Windows 11 the Mica backdrop (and its Windows 10
// acrylic fallback) is applied via DWM; the background becomes transparent so
// the backdrop shows through.
type FluentWidget struct {
	*widgets.FramelessWindow
	isMicaEnabled        bool
	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor
	backgroundColor      *qt.QColor
	bgAni                *common.ProgressAnimation
}

// NewFluentWidget builds a Fluent window.
func NewFluentWidget(parent *qt.QWidget) *FluentWidget {
	w := &FluentWidget{FramelessWindow: widgets.NewFramelessWindow(parent)}
	w.isMicaEnabled = false
	w.lightBackgroundColor = qt.NewQColor3(240, 244, 249)
	w.darkBackgroundColor = qt.NewQColor3(32, 32, 32)
	w.backgroundColor = qt.NewQColor3(240, 244, 249)

	w.SetMicaEffectEnabled(true)
	w.SetTitleBar(NewFluentWidgetTitleBar(w.QWidget).QWidget)

	// The theme listener is tied to the window: the registry drops it when the
	// window is destroyed, so closing it, deleting it with DeleteLater() and
	// switching the theme afterwards cannot call updateBackgroundColor ->
	// Update() on the freed widget (a Go closure is not a Qt slot and would not
	// be disconnected).
	common.QConfigInstance.OnThemeChangedFinishedFor(w.QObject, w.onThemeChangedFinished)
	w.OnDestroyed(func() {
		if w.bgAni != nil {
			w.bgAni.Delete()
			w.bgAni = nil
		}
	})
	w.installEvents()
	return w
}

func (w *FluentWidget) installEvents() {
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetPenWithStyle(qt.NoPen)
		brush := qt.NewQBrush3(w.backgroundColor)
		painter.SetBrush(brush)
		brush.Delete()
		rect := w.Rect()
		painter.DrawRect2(rect.X(), rect.Y(), rect.Width(), rect.Height())

		painter.End()
	})

	// Re-apply the backdrop once the native handle is fully initialized on show.
	w.SetShownHandler(func() {
		if w.isMicaEnabled {
			w.applyBackdrop()
		}
	})
}

// SetCustomBackgroundColor sets the light/dark window background colors.
func (w *FluentWidget) SetCustomBackgroundColor(light, dark *qt.QColor) {
	w.lightBackgroundColor = cloneColor(light)
	w.darkBackgroundColor = cloneColor(dark)
	w.updateBackgroundColor()
}

// SetMicaEffectEnabled toggles the Mica effect. It is a Windows 11 feature only: on
// Windows 10 (and every other platform) the window keeps its solid theme background,
// exactly like the upstream FluentWidget, whose setMicaEffectEnabled returns early
// there. Asking for the backdrop anyway would send the window down the Windows 10
// acrylic fallback, which re-composites an opaque frameless window and makes it lose its
// shadow (and leaves repaint artefacts) instead of adding a backdrop.
func (w *FluentWidget) SetMicaEffectEnabled(isEnabled bool) {
	if runtime.GOOS != "windows" || !platform.IsWindows11() {
		w.isMicaEnabled = false
		w.updateBackgroundColor()
		return
	}

	w.isMicaEnabled = isEnabled
	if isEnabled {
		w.applyBackdrop()
	} else {
		platform.RemoveBackdrop(platform.HWND(w.WinId()))
	}
	w.updateBackgroundColor()
}

// IsMicaEffectEnabled reports whether the Mica backdrop is enabled.
func (w *FluentWidget) IsMicaEffectEnabled() bool { return w.isMicaEnabled }

// WindowsBuild returns the Windows build number (0 on other platforms).
func WindowsBuild() uint32 { return platform.WindowsBuild() }

// IsWindows11 reports whether the Windows 11 DWM features are in use (the Mica backdrop
// and the DWM rounded-corner preference). It is false on every older Windows — and on a
// newer one once SetForceWindows10(true) asked for the Windows 10 behaviour.
func IsWindows11() bool { return platform.IsWindows11() }

// SetForceWindows10 forces the Windows 10 code paths — no Mica backdrop, and rounded
// corners cut with a window region instead of the DWM corner preference — on a newer
// Windows. It exists so the Windows 10 behaviour can be tested on a Windows 11 machine
// (the test example examples/window/rounded_corner switches it at runtime); the
// GOFLUENT_FORCE_WIN10 environment variable sets the same flag before startup.
//
// The windows that are already open keep the corners they applied: call
// FluentWidget.RefreshRoundedCorners (or SetMicaEffectEnabled) afterwards.
func SetForceWindows10(force bool) { platform.SetForceWindows10(force) }

// BackgroundColor returns the current background color.
func (w *FluentWidget) BackgroundColor() *qt.QColor { return w.backgroundColor }

// SetBackgroundColor sets the background color directly.
func (w *FluentWidget) SetBackgroundColor(color *qt.QColor) {
	w.backgroundColor = color
	w.Update()
}

func (w *FluentWidget) normalBackgroundColor() *qt.QColor {
	// The DWM Mica backdrop cannot be shown through the opaque Qt window surface
	// without WA_TranslucentBackground, which breaks native hit-testing (drag)
	// and crashes on theme switch. Degrade to a solid light/dark background
	// (documented in MIGRATION_GUIDE §10) so the window never turns white when
	// Mica is enabled.
	if common.IsDarkTheme() {
		return cloneColor(w.darkBackgroundColor)
	}
	return cloneColor(w.lightBackgroundColor)
}

func (w *FluentWidget) updateBackgroundColor() {
	target := w.normalBackgroundColor()
	from := cloneColor(w.backgroundColor)

	// Mirror the Python BackgroundAnimationWidget: fade the window background
	// to the new theme color over ~120ms instead of snapping. The custom
	// "backgroundColor" meta-property cannot be registered from Go, so the
	// color is lerped frame-by-frame via common.AnimateColor and repainted.
	if w.bgAni != nil {
		w.bgAni.Stop()
		w.bgAni.Delete()
		w.bgAni = nil
	}
	w.bgAni = common.AnimateColor(from, target, 120, nil, func(c *qt.QColor) {
		newC := cloneColor(c)
		old := w.backgroundColor
		w.backgroundColor = newC
		if old != nil {
			old.Delete()
		}
		w.Update()
	})
	from.Delete()
	target.Delete()
}

// applyBackdrop re-applies the DWM backdrop for the current theme. Only Windows 11
// has the Mica backdrop; on Windows 10 platform.EnableMica reports that it applied
// nothing and the window keeps its solid background.
func (w *FluentWidget) applyBackdrop() {
	if runtime.GOOS != "windows" || !platform.IsWindows11() {
		return
	}
	if !platform.EnableMica(platform.HWND(w.WinId()), common.IsDarkTheme()) {
		// No backdrop available after all: fall back to the solid background.
		w.isMicaEnabled = false
		w.updateBackgroundColor()
	}
}

func (w *FluentWidget) onThemeChangedFinished() {
	// Re-apply the Mica backdrop with the new theme and refresh the background
	// color (solid, or transparent while Mica is enabled).
	w.updateBackgroundColor()
	if w.isMicaEnabled {
		w.applyBackdrop()
	}
}

// SystemTitleBarRect returns the macOS system title bar rect.
func (w *FluentWidget) SystemTitleBarRect(size *qt.QSize) *qt.QRect {
	y := 0
	if !w.IsFullScreen() {
		y = 2
	}
	return qt.NewQRect4(0, y, 75, size.Height())
}

// navigationBarInterface is the subset of the navigation interfaces used by the
// window base class.
type navigationBarInterface interface {
	SetCurrentItem(routeKey string)
	RemoveWidget(routeKey string)
}

// FluentWindowBase is the shared base of FluentWindow and MSFluentWindow.
type FluentWindowBase struct {
	*FluentWidget
	hBoxLayout          *qt.QHBoxLayout
	stackedWidget       *StackedWidget
	navigationInterface navigationBarInterface
}

func newFluentWindowBase(parent *qt.QWidget) *FluentWindowBase {
	w := &FluentWindowBase{FluentWidget: NewFluentWidget(parent)}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.stackedWidget = NewStackedWidget(w.QWidget)

	w.hBoxLayout.SetSpacing(0)
	w.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	common.FluentStyleSheet(common.FluentWindow).Apply(w.stackedWidget.QWidget, common.ThemeAuto)
	return w
}

// AddSubInterface adds a sub interface (implemented by the concrete windows).
func (w *FluentWindowBase) AddSubInterface(interfaceWidget *qt.QWidget, icon interface{}, text string, position navigation.NavigationItemPosition) interface{} {
	panic("window: FluentWindowBase.AddSubInterface is abstract")
}

// RemoveInterface removes a sub interface.
func (w *FluentWindowBase) RemoveInterface(interfaceWidget *qt.QWidget, isDelete bool) {
	w.navigationInterface.RemoveWidget(interfaceWidget.ObjectName())
	w.stackedWidget.RemoveWidget(interfaceWidget)
	interfaceWidget.Hide()
	if isDelete {
		interfaceWidget.DeleteLater()
	}
}

// SwitchTo switches to the given sub interface without the pop-out animation.
func (w *FluentWindowBase) SwitchTo(interfaceWidget *qt.QWidget) {
	w.stackedWidget.SetCurrentWidget(interfaceWidget, false)
}

func (w *FluentWindowBase) onCurrentInterfaceChanged(index int) {
	interfaceWidget := w.stackedWidget.Widget(index)
	if interfaceWidget == nil {
		return
	}
	name := interfaceWidget.ObjectName()
	w.navigationInterface.SetCurrentItem(name)
	common.RouterInstance.Push(w.stackedWidget.QStackedWidget(), name)
	w.updateStackedBackground()
}

func (w *FluentWindowBase) updateStackedBackground() {
	current := w.stackedWidget.CurrentWidget()
	if current == nil {
		return
	}
	isTransparent := false
	if v := current.Property("isStackedTransparent"); v != nil && !v.IsNull() {
		isTransparent = v.ToBool()
	}
	currentTransparent := false
	if v := w.stackedWidget.Property("isTransparent"); v != nil && !v.IsNull() {
		currentTransparent = v.ToBool()
	}
	if currentTransparent == isTransparent {
		return
	}
	w.stackedWidget.SetProperty("isTransparent", qt.NewQVariant11(isTransparent))
	w.stackedWidget.SetStyle(qt.QApplication_Style())
}

// SystemTitleBarRect returns the macOS system title bar rect.
func (w *FluentWindowBase) SystemTitleBarRect(size *qt.QSize) *qt.QRect {
	y := 0
	if !w.IsFullScreen() {
		y = 8
	}
	return qt.NewQRect4(size.Width()-75, y, 75, size.Height())
}

// ---------------------------------------------------------------------------
// FluentWindow

// FluentWindow is the standard Fluent window with a left navigation interface.
type FluentWindow struct {
	*FluentWindowBase
	nav          *navigation.NavigationInterface
	widgetLayout *qt.QHBoxLayout
}

// NewFluentWindow builds a Fluent window.
func NewFluentWindow(parent *qt.QWidget) *FluentWindow {
	w := &FluentWindow{FluentWindowBase: newFluentWindowBase(parent)}
	w.SetTitleBar(NewFluentTitleBar(w.QWidget).QWidget)

	w.nav = navigation.NewNavigationInterface(w.QWidget, true, true, true)
	w.navigationInterface = w.nav
	w.widgetLayout = qt.NewQHBoxLayout2()

	w.hBoxLayout.AddWidget(w.nav.QWidget)
	w.hBoxLayout.AddLayout(w.widgetLayout.QLayout)
	w.hBoxLayout.SetStretchFactor2(w.widgetLayout.QLayout, 1)

	w.widgetLayout.AddWidget(w.stackedWidget.QWidget)
	// Reserve the top 48px for the title bar on the content only (mirrors
	// fluent_window.py:271 `widgetLayout.setContentsMargins(0, 48, 0, 0)`): the
	// navigation interface stays at y=0 and the raised title bar covers its top
	// strip, while the stacked widget starts below the title bar.
	w.widgetLayout.SetContentsMargins(0, 48, 0, 0)

	w.nav.OnDisplayModeChanged(func(navigation.NavigationDisplayMode) { w.TitleBar().Raise() })
	w.TitleBar().Raise()

	w.positionTitleBar()
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		w.positionTitleBar()
	})
	return w
}

// NavigationInterface returns the underlying navigation interface used by the
// window's left navigation panel.
func (w *FluentWindow) NavigationInterface() *navigation.NavigationInterface { return w.nav }

func (w *FluentWindow) positionTitleBar() {
	tb := w.TitleBar()
	if tb == nil {
		return
	}
	tb.Move(46, 0)
	tb.Resize(w.Width()-46, tb.Height())
}

// AddSubInterface adds a sub interface; the object name of interfaceWidget must
// already be set.
func (w *FluentWindow) AddSubInterface(interfaceWidget *qt.QWidget, icon interface{}, text string, position navigation.NavigationItemPosition, parent interface{}, isTransparent bool) interface{} {
	if interfaceWidget.ObjectName() == "" {
		panic("window: the object name of `interface` can't be empty string.")
	}

	parentRouteKey := ""
	switch p := parent.(type) {
	case *qt.QWidget:
		parentRouteKey = p.ObjectName()
		if parentRouteKey == "" {
			panic("window: the object name of `parent` can't be empty string.")
		}
	case string:
		parentRouteKey = p
	}

	interfaceWidget.SetProperty("isStackedTransparent", qt.NewQVariant11(isTransparent))
	w.stackedWidget.AddWidget(interfaceWidget)

	routeKey := interfaceWidget.ObjectName()
	item := w.nav.AddItem(routeKey, icon, text, func(bool) { w.SwitchTo(interfaceWidget) }, true, position, text, parentRouteKey)

	if w.stackedWidget.Count() == 1 {
		w.stackedWidget.OnCurrentChanged(w.onCurrentInterfaceChanged)
		w.nav.SetCurrentItem(routeKey)
		common.RouterInstance.SetDefaultRouteKey(w.stackedWidget.QStackedWidget(), routeKey)
	}

	w.updateStackedBackground()
	return item
}

// ---------------------------------------------------------------------------
// MSFluentWindow

// MSFluentWindow is the Microsoft Store style window with a vertical
// navigation bar.
type MSFluentWindow struct {
	*FluentWindowBase
	nav        *navigation.NavigationBar
	msTitleBar *MSFluentTitleBar
}

// NewMSFluentWindow builds a Microsoft Store style window.
func NewMSFluentWindow(parent *qt.QWidget) *MSFluentWindow {
	w := &MSFluentWindow{FluentWindowBase: newFluentWindowBase(parent)}
	w.msTitleBar = NewMSFluentTitleBar(w.QWidget)
	w.SetTitleBar(w.msTitleBar.QWidget)

	w.nav = navigation.NewNavigationBar(w.QWidget)
	w.navigationInterface = w.nav

	w.hBoxLayout.SetContentsMargins(0, 48, 0, 0)
	w.hBoxLayout.AddWidget(w.nav.QWidget)
	w.hBoxLayout.AddWidget2(w.stackedWidget.QWidget, 1)

	w.TitleBar().Raise()
	w.TitleBar().SetAttribute(qt.WA_StyledBackground)

	w.positionTitleBar()
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		w.positionTitleBar()
	})
	return w
}

// NavigationBar returns the underlying navigation bar used by the window's
// bottom navigation panel.
func (w *MSFluentWindow) NavigationBar() *navigation.NavigationBar { return w.nav }

// MSFTitleBar returns the Microsoft Store title bar, exposing its layout so
// examples can insert a TabBar or tool buttons alongside the title.
func (w *MSFluentWindow) MSFTitleBar() *MSFluentTitleBar { return w.msTitleBar }

func (w *MSFluentWindow) positionTitleBar() {
	tb := w.TitleBar()
	if tb == nil {
		return
	}
	tb.Move(0, 0)
	tb.Resize(w.Width(), tb.Height())
}

// AddSubInterface adds a sub interface; the object name of interfaceWidget must
// already be set.
func (w *MSFluentWindow) AddSubInterface(interfaceWidget *qt.QWidget, icon interface{}, text string, selectedIcon interface{}, position navigation.NavigationItemPosition, isTransparent bool) interface{} {
	if interfaceWidget.ObjectName() == "" {
		panic("window: the object name of `interface` can't be empty string.")
	}

	interfaceWidget.SetProperty("isStackedTransparent", qt.NewQVariant11(isTransparent))
	w.stackedWidget.AddWidget(interfaceWidget)

	routeKey := interfaceWidget.ObjectName()
	item := w.nav.AddItem(routeKey, icon, text, func(bool) { w.SwitchTo(interfaceWidget) }, true, selectedIcon, position)

	if w.stackedWidget.Count() == 1 {
		w.stackedWidget.OnCurrentChanged(w.onCurrentInterfaceChanged)
		w.nav.SetCurrentItem(routeKey)
		common.RouterInstance.SetDefaultRouteKey(w.stackedWidget.QStackedWidget(), routeKey)
	}

	w.updateStackedBackground()
	return item
}

// ---------------------------------------------------------------------------
// SplitFluentWindow

// SplitFluentWindow is a Fluent window with a split style title bar.
type SplitFluentWindow struct{ *FluentWindow }

// NewSplitFluentWindow builds a split style Fluent window.
func NewSplitFluentWindow(parent *qt.QWidget) *SplitFluentWindow {
	w := &SplitFluentWindow{FluentWindow: NewFluentWindow(parent)}
	w.SetTitleBar(NewSplitTitleBar(w.QWidget).QWidget)
	// The split title bar overlays the content instead of reserving a strip,
	// so drop the 48px top margin on the content (mirrors fluent_window.py).
	w.widgetLayout.SetContentsMargins(0, 0, 0, 0)
	w.TitleBar().Raise()
	return w
}

// ---------------------------------------------------------------------------
// FluentBackgroundTheme

// FluentBackgroundTheme provides the default window background color pairs.
type FluentBackgroundTheme struct{}

// Default returns the default light/dark background colors.
func (FluentBackgroundTheme) Default() (light, dark *qt.QColor) {
	return qt.NewQColor3(243, 243, 243), qt.NewQColor3(32, 32, 32)
}

// DefaultBlue returns the default blue light/dark background colors.
func (FluentBackgroundTheme) DefaultBlue() (light, dark *qt.QColor) {
	return qt.NewQColor3(240, 244, 249), qt.NewQColor3(25, 33, 42)
}
