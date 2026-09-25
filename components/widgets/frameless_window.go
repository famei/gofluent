package widgets

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"unsafe"

	"github.com/famei/gofluent/internal/platform"
	qt "github.com/mappu/miqt/qt"
)

// FramelessWindow is a window without the native title bar. On Windows it
// installs a native event hook that answers WM_NCHITTEST (border resizing and
// title-bar dragging), adds WS_THICKFRAME on first show so the resize codes
// actually enter the modal size loop, and extends the DWM frame into the client
// area to get a native shadow/rounded corners without a visible border. On
// non-Windows platforms every Win32 call degrades to a no-op and the window
// simply keeps the Qt frameless hint.
type FramelessWindow struct {
	*qt.QWidget
	resizeEnabled      bool
	moveEnabled        bool
	titleBarVisible    bool
	doubleClickEnabled bool
	titleBar           *qt.QWidget
	shownHandler       func()

	// dragByMouseMove switches dragging from the native hit test to the Qt mouse
	// event handover (see SetMouseMoveDragEnabled); mouseMoveDragging holds "the
	// left button went down on the title bar, waiting for the move that starts the
	// drag". resizeByMouseMove/mouseMoveResizeCode are the same for border resizing
	// (SetMouseMoveResizeEnabled).
	dragByMouseMove   bool
	mouseMoveDragging bool
	resizeByMouseMove bool
	manual            manualResize
	resizeCursorShape qt.CursorShape
	resizeCursorSet   bool

	// cornerRadius is the rounded-corner radius in logical pixels (see
	// SetCornerRadius); rounded records what was applied last so a resize only
	// re-applies the region when something actually changed.
	cornerRadius int
	rounded      roundedState
	// roundedFilter re-applies the corners on show/resize (see
	// installRoundedCornerFilter).
	roundedFilter *qt.QObject
}

// CornerRoundingKind says how a frameless window rounds its corners.
type CornerRoundingKind int

const (
	// CornerRoundingNone leaves the window square (radius 0, a maximized window, or a
	// platform whose window server has no rounding this library can ask for).
	CornerRoundingNone CornerRoundingKind = iota
	// CornerRoundingDWM is the Windows 11 DWM corner preference
	// (DWMWA_WINDOW_CORNER_PREFERENCE): smooth corners drawn by the window manager.
	CornerRoundingDWM
	// CornerRoundingRegion is a window region (SetWindowRgn), the Windows 10 fallback:
	// Windows 10 ignores the DWM corner preference, so the corners are cut out of the
	// window itself. A region is a 1 bit mask, so these corners are not anti-aliased.
	CornerRoundingRegion
	// CornerRoundingMask is the Linux/X11 way: QWidget.SetMask cuts the corners through
	// the Shape extension, which clips the window and everything inside it. Wayland has
	// no client side shape — a Wayland session runs X11 clients through XWayland, which
	// applies the shape. Like the Windows 10 region the mask is not anti-aliased.
	CornerRoundingMask
)

// String returns the rounding kind as it is printed in diagnostics.
func (k CornerRoundingKind) String() string {
	switch k {
	case CornerRoundingDWM:
		return "DWM corner preference (Windows 11)"
	case CornerRoundingRegion:
		return "window region (Windows 10 fallback)"
	case CornerRoundingMask:
		return "window mask (Linux/X11 shape)"
	default:
		return "none"
	}
}

// roundedState remembers the window geometry and state the rounded corners were
// applied for, so a resize does not repeat the platform calls for every step that keeps
// the same size.
type roundedState struct {
	valid     bool
	width     int
	height    int
	maximized bool
	radius    int
	kind      CornerRoundingKind
}

// NewFramelessWindow builds a frameless window with resizing, moving, title-bar
// dragging and double-click maximize enabled.
func NewFramelessWindow(parent *qt.QWidget) *FramelessWindow {
	w := &FramelessWindow{QWidget: qt.NewQWidget(parent)}
	w.SetWindowFlags(qt.FramelessWindowHint)
	w.resizeEnabled = true
	w.moveEnabled = true
	w.titleBarVisible = true
	w.doubleClickEnabled = true
	w.cornerRadius = platform.DefaultCornerRadius
	w.installNativeHook()
	w.installShowHook()
	w.installResizeHook()
	w.installMouseMoveDrag()
	w.installRoundedCornerFilter()
	if runtime.GOOS != "windows" {
		// WM_NCHITTEST does not exist outside Windows, so the native route never sees a
		// drag there: enable the Qt routes instead, which hand the move over to
		// QWindow::startSystemMove (X11/Wayland) and resize the window from the mouse
		// handler. See both setters for the caveats of the resize route.
		w.SetMouseMoveDragEnabled(true)
		w.SetMouseMoveResizeEnabled(true)
	}
	return w
}

// installRoundedCornerFilter re-applies the rounded corners whenever the window is
// shown or resized. A QObject event filter is used (rather than the Qt resize hook)
// because the window subclasses register their own OnResizeEvent handler, which replaces
// - and would silently drop - the one of FramelessWindow. The filter never consumes an
// event.
func (w *FramelessWindow) installRoundedCornerFilter() {
	w.roundedFilter = qt.NewQObject2(w.QObject)
	w.InstallEventFilter(w.roundedFilter)
	w.roundedFilter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		switch event.Type() {
		case qt.QEvent__Resize, qt.QEvent__Show:
			w.applyRoundedCorners()
		}
		return super(watched, event)
	})
}

// SetShownHandler registers a callback invoked after the native frame is
// (re)applied each time the window is shown. window.FluentWidget uses it to
// re-apply the Mica backdrop once the native handle is fully initialized.
func (w *FramelessWindow) SetShownHandler(f func()) { w.shownHandler = f }

// SetResizeEnabled toggles window resizing.
func (w *FramelessWindow) SetResizeEnabled(isEnabled bool) {
	w.resizeEnabled = isEnabled
}

// IsResizeEnabled reports whether resizing is enabled.
func (w *FramelessWindow) IsResizeEnabled() bool { return w.resizeEnabled }

// SetMoveEnabled toggles window dragging.
func (w *FramelessWindow) SetMoveEnabled(isEnabled bool) {
	w.moveEnabled = isEnabled
}

// IsMoveEnabled reports whether dragging is enabled.
func (w *FramelessWindow) IsMoveEnabled() bool { return w.moveEnabled }

// SetMouseMoveDragEnabled switches window dragging between the two available
// mechanisms:
//
//   - false (default): the native route. The window answers WM_NCHITTEST with
//     HTCAPTION over the draggable part of the title bar and Windows runs its own
//     move loop.
//   - true: the Qt route. The window watches its own mouse events and, on the first
//     move while the left button is held on the title bar, calls ReleaseCapture()
//     and sends WM_NCLBUTTONDOWN with HTCAPTION (see StartSystemMove), which hands
//     the move over to the same system move loop from inside a Qt mouseMoveEvent.
//
// Use the Qt route when the native one cannot start a drag —e.g. when another
// process's window owns part of the client area, when something changed the window
// style at runtime, or on a platform where WM_NCHITTEST is not delivered. Both
// routes use the same drag area (HitTest: the title bar minus interactive controls)
// and can stay enabled together: if the native route captures the mouse first, Qt
// never sees the move, and if it does not, the Qt route takes over.
func (w *FramelessWindow) SetMouseMoveDragEnabled(isEnabled bool) {
	w.dragByMouseMove = isEnabled
}

// IsMouseMoveDragEnabled reports whether dragging uses the Qt mouse-event handover
// instead of the native WM_NCHITTEST route.
func (w *FramelessWindow) IsMouseMoveDragEnabled() bool { return w.dragByMouseMove }

// StartSystemMove hands a drag over to the system's modal move loop: it releases the
// mouse capture and sends WM_NCLBUTTONDOWN with HTCAPTION, exactly like a click on a
// native caption. Windows then moves the window until the button is released (a
// double click while moving maximizes it).
//
// It blocks until the move ends, so call it from a mouse handler (the
// SetMouseMoveDragEnabled route does exactly that) or from a timer, not from a
// paint/layout path. Safe to call when nothing is pressed: the move ends as soon as
// the system sees no button down.
//
// Outside Windows there is no such message: the move is handed to
// QWindow::startSystemMove, which asks the window system to start its own move
// (X11 and Wayland both implement it). It returns false when the platform cannot
// start a move right then - typically because no button is pressed.
func (w *FramelessWindow) StartSystemMove() {
	if runtime.GOOS != "windows" {
		if handle := w.WindowHandle(); handle != nil {
			handle.StartSystemMove()
		}
		return
	}
	hwnd := platform.HWND(w.WinId())
	platform.ReleaseCapture()
	platform.SendMessage(hwnd, platform.WM_NCLBUTTONDOWN, uintptr(platform.HTCAPTION), 0)
}

// StartSystemResize hands a border resize over to the system's modal size loop: it
// releases the mouse capture and sends WM_NCLBUTTONDOWN with a resize hit code
// (HTLEFT, HTTOP, HTBOTTOMRIGHT, ... as returned by HitTest), exactly like grabbing a
// native sizing border. Windows then resizes the window until the button is released.
//
// Unlike the move handover, the size loop needs the mouse position in the message's
// lParam (the WM_NCLBUTTONDOWN convention: the cursor in physical screen
// coordinates); sending 0 there leaves the loop without an anchor and no edge
// resizes at all, so the position is passed explicitly. Qt's QCursor::pos is in
// logical pixels, hence platform.GetCursorPos.
//
// Like StartSystemMove it blocks until the resize ends, so call it from a mouse
// handler or a timer.
//
// Outside Windows the hit code is translated to the qt.Edge mask the platform wants
// and handed to QWindow::startSystemResize. Linux/X11 window managers differ in what
// they support: when the request is refused, the caller still has the window-driven
// route (SetMouseMoveResizeEnabled), which resizes without the window system.
func (w *FramelessWindow) StartSystemResize(hitCode int32) {
	if !isResizeHitCode(hitCode) {
		return
	}
	if runtime.GOOS != "windows" {
		if handle := w.WindowHandle(); handle != nil {
			handle.StartSystemResize(resizeEdges(hitCode))
		}
		return
	}
	hwnd := platform.HWND(w.WinId())
	x, y := platform.GetCursorPos()
	lparam := uintptr(uint32(x)&0xFFFF | uint32(y)<<16)
	platform.ReleaseCapture()
	platform.SendMessage(hwnd, platform.WM_NCLBUTTONDOWN, uintptr(hitCode), lparam)
}

// resizeEdges converts a WM_NCHITTEST resize code to the qt.Edge mask used by
// QWindow::startSystemResize. The codes are Windows constants but the eight
// directions they name are the same everywhere, and HitTest returns them on every
// platform.
func resizeEdges(hitCode int32) qt.Edge {
	var edges qt.Edge
	switch hitCode {
	case platform.HTLEFT, platform.HTTOPLEFT, platform.HTBOTTOMLEFT:
		edges |= qt.LeftEdge
	case platform.HTRIGHT, platform.HTTOPRIGHT, platform.HTBOTTOMRIGHT:
		edges |= qt.RightEdge
	}
	switch hitCode {
	case platform.HTTOP, platform.HTTOPLEFT, platform.HTTOPRIGHT:
		edges |= qt.TopEdge
	case platform.HTBOTTOM, platform.HTBOTTOMLEFT, platform.HTBOTTOMRIGHT:
		edges |= qt.BottomEdge
	}
	return edges
}

// isResizeHitCode reports whether a WM_NCHITTEST code is one of the eight border
// resize codes.
func isResizeHitCode(code int32) bool {
	switch code {
	case platform.HTLEFT, platform.HTRIGHT, platform.HTTOP, platform.HTBOTTOM,
		platform.HTTOPLEFT, platform.HTTOPRIGHT, platform.HTBOTTOMLEFT, platform.HTBOTTOMRIGHT:
		return true
	}
	return false
}

// manualResize tracks an in-progress border resize performed by the window itself
// (see SetMouseMoveResizeEnabled): the edge being dragged, the mouse position and the
// window geometry where the drag started.
type manualResize struct {
	code                 int32
	startX, startY       int
	winX, winY, winW, wH int
}

// startManualResize begins a window-driven resize for the given hit code, using the
// press' global position as the anchor.
func (w *FramelessWindow) startManualResize(code int32, globalX, globalY int) {
	geo := w.Geometry() // GoGC-armed �?do NOT Delete
	w.manual = manualResize{
		code:   code,
		startX: globalX, startY: globalY,
		winX: geo.X(), winY: geo.Y(), winW: geo.Width(), wH: geo.Height(),
	}
}

// manualResizeTo applies an in-progress resize for the mouse position globalX/globalY
// (Qt logical coordinates, so it is DPI-safe). The window is resized directly instead
// of handing the drag to the system's modal size loop, which needs a real button
// state and does nothing when the resize is started from a Qt mouse event.
func (w *FramelessWindow) manualResizeTo(globalX, globalY int) {
	m := &w.manual
	if m.code == 0 {
		return
	}
	dx := globalX - m.startX
	dy := globalY - m.startY

	minW, minH := w.MinimumWidth(), w.MinimumHeight()
	if minW < 80 {
		minW = 80
	}
	if minH < 40 {
		minH = 40
	}

	x, y, cw, ch := m.winX, m.winY, m.winW, m.wH
	switch m.code {
	case platform.HTLEFT:
		x, cw = m.winX+dx, m.winW-dx
	case platform.HTRIGHT:
		cw = m.winW + dx
	case platform.HTTOP:
		y, ch = m.winY+dy, m.wH-dy
	case platform.HTBOTTOM:
		ch = m.wH + dy
	case platform.HTTOPLEFT:
		x, cw = m.winX+dx, m.winW-dx
		y, ch = m.winY+dy, m.wH-dy
	case platform.HTTOPRIGHT:
		cw = m.winW + dx
		y, ch = m.winY+dy, m.wH-dy
	case platform.HTBOTTOMLEFT:
		x, cw = m.winX+dx, m.winW-dx
		ch = m.wH + dy
	case platform.HTBOTTOMRIGHT:
		cw = m.winW + dx
		ch = m.wH + dy
	}

	// Keep the opposite edge anchored while the dragged edge hits the minimum size.
	if cw < minW {
		if m.code == platform.HTLEFT || m.code == platform.HTTOPLEFT || m.code == platform.HTBOTTOMLEFT {
			x = m.winX + m.winW - minW
		}
		cw = minW
	}
	if ch < minH {
		if m.code == platform.HTTOP || m.code == platform.HTTOPLEFT || m.code == platform.HTTOPRIGHT {
			y = m.winY + m.wH - minH
		}
		ch = minH
	}

	if x != m.winX || y != m.winY || cw != m.winW || ch != m.wH {
		w.SetGeometry(x, y, cw, ch)
		// Repaint synchronously after the geometry change. Qt repositions the window's
		// child widgets (title bar, navigation panel, pages) from the resize/move
		// handlers that SetGeometry runs synchronously, but the native move/resize
		// itself can paint the window before those handlers have finished - and the
		// pending layouts are only posted, so a repaint can happen before them too.
		// Without this, dragging the left or top border sometimes leaves a child
		// painted at its old position until something else repaints the window.
		if l := w.Layout(); l != nil {
			l.Activate()
		}
		// Force the whole tree to repaint: toggling updates discards Qt's backing store
		// for the window and every child, so no stale region can survive. A fluent
		// window is full of translucent children (the navigation panel, the title bar)
		// whose composited pixels are otherwise reused — which showed up as a
		// shadow-like ghost of the child at its previous position while an edge (as
		// opposed to a corner, which happens to invalidate more) was dragged.
		w.SetUpdatesEnabled(false)
		w.SetUpdatesEnabled(true)
		w.Update()
		w.Repaint()
		// A window *move* also leaves the DWM frame and its shadow painted where the
		// window used to be; DWM recomputes them on a size change but not reliably on a
		// move. Invalidating the frame makes DWM redraw it.
		platform.RedrawWindow(platform.HWND(w.WinId()),
			platform.RDW_INVALIDATE|platform.RDW_FRAME|platform.RDW_UPDATENOW|platform.RDW_ALLCHILDREN)
		// A child that has its own native window (an embedded browser, a widget that
		// asked for a native surface) paints itself and is therefore *not* repainted by
		// the window's repaint.
		repaintNativeChildren(w.QWidget)

		frameLogResize(w)
	}
}

// frameLogResize reports, when GOFLUENT_DEBUG_FRAME is set, the window geometry and
// the screen position of its child widgets after a resize step. It distinguishes a
// wrong geometry (the child really is somewhere else) from a stale paint (the child's
// geometry is right but the window shows it at the old place).
func frameLogResize(w *FramelessWindow) {
	if !frameEnabled() {
		return
	}
	geo := w.Geometry() // GoGC-armed — do NOT Delete
	line := fmt.Sprintf("resize: win=(%d,%d) %dx%d", geo.X(), geo.Y(), geo.Width(), geo.Height())
	shown := 0
	for _, child := range w.Children() {
		if child == nil || !child.Inherits("QWidget") || shown >= 5 {
			continue
		}
		cw := qt.UnsafeNewQWidget(child.UnsafePointer())
		if cw == nil || !cw.IsVisible() {
			continue
		}
		p := cw.MapToGlobal(qt.NewQPoint2(0, 0)) // GoGC-armed — do NOT Delete
		name := cw.ObjectName()
		if name == "" {
			name = classNameOf(cw)
		}
		line += fmt.Sprintf("  %s screen=(%d,%d) %dx%d native=%v parent=%d", name, p.X(), p.Y(), cw.Width(), cw.Height(), cw.InternalWinId() != 0, platform.GetParent(platform.HWND(cw.InternalWinId())))
		shown++
	}
	frameLog("%s", line)
}

// repaintNativeChildren repaints every descendant widget that owns a native window, so
// their own surface follows a geometry change of the window (the parent's repaint does
// not reach them).
func repaintNativeChildren(w *qt.QWidget) {
	if w == nil {
		return
	}
	for _, child := range w.Children() {
		if child == nil || !child.Inherits("QWidget") {
			continue
		}
		cw := qt.UnsafeNewQWidget(child.UnsafePointer())
		if cw == nil {
			continue
		}
		if cw.InternalWinId() != 0 {
			cw.Update()
			cw.Repaint()
		}
		repaintNativeChildren(cw)
	}
}

// classNameOf reports a widget's Qt class name.
func classNameOf(w *qt.QWidget) string {
	if mo := w.MetaObject(); mo != nil {
		return mo.ClassName()
	}
	return "?"
}

// SetMouseMoveResizeEnabled switches border resizing to the Qt mouse-event route,
// the counterpart of SetMouseMoveDragEnabled:
//
//   - false (default): the native route. WM_NCHITTEST returns the resize codes and
//     Windows runs its own size loop.
//   - true: the Qt route. The window watches its own mouse events and, on the first
//     move while the left button is held inside the resize border, calls
//     StartSystemResize with the code HitTest reported.
//
// Use it when the native route cannot start a resize. An embedded WebView2 (the
// web_engine example) subclasses the window procedure and answers the hit test
// itself, so every edge stops resizing even though HitTest still returns the resize
// codes and the window style is intact —the same interception that makes the Qt
// drag route necessary. The border that starts the resize is the same for both
// routes (HitTest), so the Qt route needs the border to be free of native child
// windows (keep a margin larger than the border width around embedded views).
//
// Caveats — this route changes how mouse events reach every widget of the window:
//
//   - It turns mouse tracking on for the window and every child widget
//     (enableMouseTrackingDeep), and refreshes it on each resize, because the resize
//     cursor has to keep updating even where a child covers the border. Qt's default is
//     the opposite: a widget without mouse tracking only receives MouseMove while a
//     button is held, and that assumption is what most mouse handlers are written
//     against.
//   - As a result a plain hover (no button) now arrives at child widgets as MouseMove
//     too. Any handler that acts on a move must ignore events with
//     Buttons() == qt.NoButton or gate itself on its own press state, otherwise the
//     widget reacts to hovering alone. A fluent overlay scroll bar did exactly that and
//     scrolled the table under the pointer. The library's own controls are guarded:
//     ScrollBar (isPressed plus a held button), Slider, TabItem and HuePanel all require
//     a button press. A custom or third-party widget added to such a window needs its
//     own guard.
//   - Turning the route off again does not restore the previous tracking state: it is
//     only ever enabled, never disabled, for widgets that were already reached.
//
// Only this resize route enables tracking deep in the tree; SetMouseMoveDragEnabled
// leaves the children alone.
func (w *FramelessWindow) SetMouseMoveResizeEnabled(isEnabled bool) {
	w.resizeByMouseMove = isEnabled
	if isEnabled {
		enableMouseTrackingDeep(w.QWidget)
	}
}

// IsMouseMoveResizeEnabled reports whether resizing uses the Qt mouse-event handover.
func (w *FramelessWindow) IsMouseMoveResizeEnabled() bool { return w.resizeByMouseMove }

// useResizeCursor keeps the mouse cursor in step with the resize border under the
// cursor: the horizontal/vertical/diagonal double arrow for the eight border codes,
// and the widget's own cursor everywhere else.
//
// A frameless window normally gets those cursors from Windows itself, which sets them
// from the WM_NCHITTEST result. Here that message is answered by whatever native child
// is under the cursor — an embedded browser, an external player — so the border only
// exists for the Qt route and the cursor would otherwise stay an arrow, making the
// window look unresizable.
func (w *FramelessWindow) useResizeCursor(code int32) {
	if !w.resizeByMouseMove {
		return
	}
	var shape qt.CursorShape
	switch code {
	case platform.HTLEFT, platform.HTRIGHT:
		shape = qt.SizeHorCursor
	case platform.HTTOP, platform.HTBOTTOM:
		shape = qt.SizeVerCursor
	case platform.HTTOPLEFT, platform.HTBOTTOMRIGHT:
		shape = qt.SizeFDiagCursor
	case platform.HTTOPRIGHT, platform.HTBOTTOMLEFT:
		shape = qt.SizeBDiagCursor
	default:
		if w.resizeCursorSet {
			w.UnsetCursor()
			w.resizeCursorSet = false
		}
		return
	}
	if w.resizeCursorSet && w.resizeCursorShape == shape {
		return
	}
	w.resizeCursorShape, w.resizeCursorSet = shape, true
	cursor := qt.NewQCursor2(shape)
	w.SetCursor(cursor) // Qt copies the cursor
	cursor.Delete()
}

// enableMouseTrackingDeep turns on mouse tracking for the widget and every child
// widget, so the window keeps receiving moves (and can update the resize cursor) even
// where a child covers the border. Children created later are picked up by the next
// call; the resize hook refreshes it on every resize.
//
// QObject::children() also holds non-widget children (layouts, actions, timers), so
// every child is checked with Inherits before being touched as a widget.
func enableMouseTrackingDeep(w *qt.QWidget) {
	if w == nil {
		return
	}
	w.SetMouseTracking(true)
	for _, child := range w.Children() {
		if child == nil || !child.Inherits("QWidget") {
			continue
		}
		if cw := qt.UnsafeNewQWidget(child.UnsafePointer()); cw != nil {
			enableMouseTrackingDeep(cw)
		}
	}
}

// frameLog is a temporary diagnostic for the mouse-driven frame interactions: set
// GOFLUENT_DEBUG_FRAME=1 to see, on stderr, what the window's mouse handlers receive
// and which handover they trigger. It is read per call so a test process can set the
// variable itself.
func frameLog(format string, args ...interface{}) {
	if !frameEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "[frame] "+format+"\n", args...)
}

// frameEnabled reports whether the frame diagnostic is on (GOFLUENT_DEBUG_FRAME).
func frameEnabled() bool { return os.Getenv("GOFLUENT_DEBUG_FRAME") != "" }

// installMouseMoveDrag wires the Qt drag route. The decision of "is this the
// draggable part of the title bar" is HitTest, so an interactive control placed in
// the title bar keeps receiving its clicks (it consumes the press before it reaches
// the window) and the two routes can never disagree about the drag area. The same
// handlers implement the Qt resize route (SetMouseMoveResizeEnabled).
func (w *FramelessWindow) installMouseMoveDrag() {
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		if e.Button() != qt.LeftButton {
			return
		}
		pos := e.Pos() // GoGC-armed —do NOT Delete
		code := w.HitTest(pos.X(), pos.Y())
		frameLog("press at (%d,%d) -> hitTest=%d dragMode=%v resizeMode=%v",
			pos.X(), pos.Y(), code, w.dragByMouseMove, w.resizeByMouseMove)
		if w.dragByMouseMove && code == platform.HTCAPTION {
			w.mouseMoveDragging = true
			return
		}
		if w.resizeByMouseMove && isResizeHitCode(code) {
			g := e.GlobalPos() // GoGC-armed �?do NOT Delete
			w.startManualResize(code, g.X(), g.Y())
		}
	})

	w.OnMouseMoveEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		pos := e.Pos() // GoGC-armed — do NOT Delete
		// Keep the cursor telling the truth about the border: the frameless window
		// never gets Windows' own sizing cursor, because the hit test is answered by
		// whatever native child is under the cursor (the embedded browser) instead of
		// this window.
		w.useResizeCursor(w.HitTest(pos.X(), pos.Y()))
		frameLog("move: dragging=%v resizeCode=%d buttons=%d", w.mouseMoveDragging, w.manual.code, e.Buttons())
		if w.manual.code != 0 {
			if e.Buttons()&qt.LeftButton == 0 {
				w.manual.code = 0
				return
			}
			// Resize directly: the system's modal size loop needs a real button state
			// and does nothing when the resize starts from a Qt mouse event.
			g := e.GlobalPos() // GoGC-armed — do NOT Delete
			w.manualResizeTo(g.X(), g.Y())
			return
		}
		if !w.mouseMoveDragging {
			return
		}
		if e.Buttons()&qt.LeftButton == 0 {
			w.mouseMoveDragging = false
			return
		}
		// Hand over once; StartSystemMove blocks until the drag finishes.
		w.mouseMoveDragging = false
		w.StartSystemMove()
	})

	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		w.mouseMoveDragging = false
		w.manual.code = 0
	})

	// The native route gets double-click maximize from the caption; in the Qt route
	// the double click arrives as a client event, so handle it here too.
	w.OnMouseDoubleClickEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		if !w.dragByMouseMove || !w.doubleClickEnabled {
			return
		}
		pos := e.Pos() // GoGC-armed —do NOT Delete
		if w.HitTest(pos.X(), pos.Y()) != platform.HTCAPTION {
			return
		}
		if w.IsMaximized() {
			w.ShowNormal()
		} else {
			w.ShowMaximized()
		}
	})
}

// SetTitleBar sets the custom title bar widget, replacing (and destroying) any
// previous title bar. A window may set the title bar more than once (e.g.
// window.FluentWidget installs FluentWidgetTitleBar and window.FluentWindow
// then swaps in FluentTitleBar); without hiding the old bar both remain visible
// as stacked children and render duplicate icon/title/button groups.
func (w *FramelessWindow) SetTitleBar(titleBar *qt.QWidget) {
	if titleBar == nil {
		w.titleBar = nil
		return
	}
	if w.titleBar != nil && w.titleBar.UnsafePointer() != titleBar.UnsafePointer() {
		w.titleBar.Hide()
		w.titleBar.DeleteLater()
	}
	w.titleBar = titleBar
	w.positionTitleBar()
}

// positionTitleBar stretches the title bar across the top of the window so the
// system buttons sit against the right edge (mirrors qframelesswindow, which
// repositions the title bar to full width on every resize). window.FluentWindow
// and its variants shadow this method with a 46px left-gutter variant.
func (w *FramelessWindow) positionTitleBar() {
	if w.titleBar == nil {
		return
	}
	w.titleBar.Move(0, 0)
	w.titleBar.Resize(w.Width(), w.titleBar.Height())
}

// installResizeHook keeps the title bar glued to the top and full width while
// the window is resized. Subclasses that need different geometry register their
// own OnResizeEvent handler, which overrides this one.
func (w *FramelessWindow) installResizeHook() {
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		w.positionTitleBar()
		if w.resizeByMouseMove {
			enableMouseTrackingDeep(w.QWidget)
		}
		// The rounded corners follow the window size (and stop at the screen edges
		// while maximized).
		w.applyRoundedCorners()
	})
}

// TitleBar returns the custom title bar widget.
func (w *FramelessWindow) TitleBar() *qt.QWidget { return w.titleBar }

// KeepTitleBarOnTop gives the title bar its own native window and raises it above
// every other child of the window.
//
// A child widget that is itself a native window —for example one that hands its
// WinId() to an external player through `mplayer.exe -wid <hwnd>` —is composited
// by Windows above the (normally non-native) title bar and receives the mouse input
// over the whole area it covers. When such a child overlaps the title bar the window
// can no longer be dragged, even though the frameless hit test still answers
// HTCAPTION for the bar: the WM_NCHITTEST of the top-level window is never
// consulted for points that belong to a native child window.
//
// Calling this after embedding the external window makes the title bar a native
// sibling and pushes it to the top of the child z-order, so its drag area keeps its
// mouse input. The bar stays where the layout puts it; call it again after the bar
// is replaced.
func (w *FramelessWindow) KeepTitleBarOnTop() {
	if w.titleBar == nil {
		return
	}
	// Its own native window, so Qt can order it against other native children
	// instead of only painting it into the parent's backing store.
	w.titleBar.SetAttribute(qt.WA_NativeWindow)
	w.titleBar.Raise()
}

// SetTitleBarVisible toggles the title bar visibility.
func (w *FramelessWindow) SetTitleBarVisible(isVisible bool) {
	w.titleBarVisible = isVisible
	if w.titleBar != nil {
		w.titleBar.SetVisible(isVisible)
	}
}

// IsTitleBarVisible reports whether the title bar is visible.
func (w *FramelessWindow) IsTitleBarVisible() bool { return w.titleBarVisible }

// SetDoubleClickEnabled toggles double-click maximize.
func (w *FramelessWindow) SetDoubleClickEnabled(isEnabled bool) {
	w.doubleClickEnabled = isEnabled
}

// IsDoubleClickEnabled reports whether double-click maximize is enabled.
func (w *FramelessWindow) IsDoubleClickEnabled() bool { return w.doubleClickEnabled }

// SetMinimizeButtonEnabled toggles the minimize button.
func (w *FramelessWindow) SetMinimizeButtonEnabled(isEnabled bool) {}

// SetMaximizeButtonEnabled toggles the maximize button.
func (w *FramelessWindow) SetMaximizeButtonEnabled(isEnabled bool) {}

// SetCloseButtonEnabled toggles the close button.
func (w *FramelessWindow) SetCloseButtonEnabled(isEnabled bool) {}

// installNativeHook answers WM_NCHITTEST so the system provides native
// title-bar dragging and 8-way border resizing for the frameless window.
func (w *FramelessWindow) installNativeHook() {
	w.OnNativeEvent(func(super func(eventType []byte, message unsafe.Pointer, result *int64) bool,
		eventType []byte, message unsafe.Pointer, result *int64) bool {

		// Qt passes "windows_generic_MSG" for every native Windows message.
		if string(eventType) == "windows_generic_MSG" {
			msg := (*platform.MSG)(message)
			switch msg.Message {
			case platform.WM_NCHITTEST:
				if code := w.nativeHitTest(); code != 0 {
					// miqt Qt5 binds C's 32-bit long* as a Go *int64. Writing
					// all 8 bytes would overflow the 4-byte slot; only the low
					// 4 bytes are valid.
					*(*int32)(unsafe.Pointer(result)) = int32(code)
					return true
				}
			case platform.WM_NCCALCSIZE:
				// Adjust the maximized client rect so the invisible resize
				// border does not push the visible content past the work area
				// (under the taskbar / off-screen). Mirror qframelesswindow's
				// WM_NCCALCSIZE handler.
				code := w.adjustMaximizedClientRect(platform.HWND(msg.HWnd), msg.WParam, msg.LParam)
				*(*int32)(unsafe.Pointer(result)) = int32(code)
				return true
			}
		}
		return super(eventType, message, result)
	})
}

// installShowHook applies the native frame on first show and, once the native
// handle is fully initialized, runs the registered shown handler.
func (w *FramelessWindow) installShowHook() {
	w.OnShowEvent(func(super func(e *qt.QShowEvent), e *qt.QShowEvent) {
		super(e)
		w.applyNativeFrame()
		if w.shownHandler != nil {
			w.shownHandler()
		}
	})
}

// applyNativeFrame adds the native styles that enable 8-way resizing (via
// WM_NCHITTEST) and the system maximize/restore/minimize animation, extends the
// DWM frame into the client area (native shadow without a visible border) and
// rounds the window corners.
func (w *FramelessWindow) applyNativeFrame() {
	hwnd := platform.HWND(w.WinId())

	if w.resizeEnabled {
		style := platform.GetWindowLongPtr(hwnd, platform.GWL_STYLE)
		// WS_CAPTION is what makes Windows animate the maximize/restore
		// transition; the DWM frame extension below hides the caption visually
		// (mirrors qframelesswindow's addWindowAnimation).
		style |= platform.WS_THICKFRAME | platform.WS_MAXIMIZEBOX | platform.WS_MINIMIZEBOX | platform.WS_CAPTION
		platform.SetWindowLongPtr(hwnd, platform.GWL_STYLE, style)
	}
	_ = platform.DwmExtendFrameIntoClientArea(hwnd, &platform.MARGINS{Left: -1, Right: -1, Top: -1, Bottom: -1})
	w.rounded.valid = false
	w.applyRoundedCorners()
}

// SetCornerRadius sets the corner radius of the frameless window in logical pixels.
// 0 makes the window square. Windows 11 rounds the window through the DWM corner
// preference; Windows 10 ignores that attribute, so the same corners are cut with a
// window region (see applyRoundedCorners).
func (w *FramelessWindow) SetCornerRadius(radius int) {
	if radius < 0 {
		radius = 0
	}
	w.cornerRadius = radius
	w.applyRoundedCorners()
}

// CornerRadius returns the corner radius in logical pixels.
func (w *FramelessWindow) CornerRadius() int { return w.cornerRadius }

// DefaultCornerRadius returns the corner radius a frameless window starts with: 8
// logical pixels, which matches the Windows 11 rounded-corner preference (and is what
// the Windows 10 fallback cuts out of the window region).
func DefaultCornerRadius() int { return platform.DefaultCornerRadius }

// CornerRounding reports how the window corners are currently rounded (see
// CornerRoundingKind). It is meant for diagnostics.
func (w *FramelessWindow) CornerRounding() CornerRoundingKind {
	if !w.rounded.valid {
		return CornerRoundingNone
	}
	return w.rounded.kind
}

// RoundedByDWM reports whether the corners are rounded by the Windows 11 DWM corner
// preference (rather than by a window region on Windows 10 or a window mask on Linux).
func (w *FramelessWindow) RoundedByDWM() bool { return w.CornerRounding() == CornerRoundingDWM }

// RefreshRoundedCorners re-applies the rounded corners, ignoring the "nothing changed"
// cache. Call it when something the corners depend on changed outside the window: the
// forced Windows 10 mode (see window.SetForceWindows10) or a screen DPI change.
func (w *FramelessWindow) RefreshRoundedCorners() {
	w.rounded.valid = false
	w.applyRoundedCorners()
}

// applyRoundedCorners rounds the window corners for the current size and state, and
// records how (see CornerRoundingKind). A maximized or full-screen window is left
// square: its corners sit on the screen edges, where rounding would only clip the
// content.
//
// Every platform rounds the window with the mechanism its window manager offers:
//
//   - Windows 11: the DWM corner preference, which is smooth and keeps the native
//     shadow;
//   - Windows 10: a window region (the DWM ignores the preference there), so the
//     corners are cut out of the window and are not anti-aliased;
//   - Linux/X11: a widget mask (QWidget.SetMask), which the X server applies through the
//     Shape extension — also a 1 bit mask, and also not anti-aliased;
//   - macOS: nothing, the window server rounds every window by itself.
//
// The mask and the Windows region clip the window with all its content, so they work
// whatever the window paints.
func (w *FramelessWindow) applyRoundedCorners() {
	width, height := w.Width(), w.Height()
	if width <= 0 || height <= 0 {
		return
	}
	// Windows rounds the window through its native handle; the Linux mask is a widget
	// property and does not need one.
	if runtime.GOOS == "windows" && w.WinId() == 0 {
		return
	}

	radius := w.cornerRadius
	maximized := w.isFramelessMaximized()
	if maximized {
		radius = 0
	}
	// Windows rounds through the window itself (SetWindowRgn/CreateRoundRectRgn), which
	// works in physical pixels, so the radius is scaled there. Qt's mask on Linux is in
	// logical coordinates and Qt maps it to the device pixels of the screen.
	deviceRadius := radius
	if radius > 0 && runtime.GOOS == "windows" {
		if dpr := w.DevicePixelRatioF(); dpr > 0.5 {
			deviceRadius = int(math.Round(float64(radius) * dpr))
		}
	}

	// Nothing changed since the last call: skip the platform round trip (this runs on
	// every resize step).
	last := w.rounded
	if last.valid && last.width == width && last.height == height &&
		last.maximized == maximized && last.radius == deviceRadius {
		return
	}

	kind := CornerRoundingNone
	switch {
	case radius <= 0:
		w.clearRoundedCorners()
	case runtime.GOOS == "windows":
		if platform.EnableRoundedCorners(platform.HWND(w.WinId()), deviceRadius) {
			kind = CornerRoundingDWM
		} else {
			kind = CornerRoundingRegion
		}
	case runtime.GOOS == "linux":
		platform.SetWindowRoundedMask(w.QWidget, radius)
		kind = CornerRoundingMask
	}

	w.rounded = roundedState{
		valid: true, width: width, height: height,
		maximized: maximized, radius: deviceRadius, kind: kind,
	}
	frameLog("rounded corners: radius=%d max=%v kind=%v", deviceRadius, maximized, kind)
}

// isFramelessMaximized reports whether the window fills the screen, in which case the
// corners are not rounded. Windows reads the live WS_MAXIMIZE style bit: Qt's
// IsMaximized lags behind during a maximize/restore transition, which would leave the
// rounding applied to a window that already fills the screen.
func (w *FramelessWindow) isFramelessMaximized() bool {
	if runtime.GOOS == "windows" {
		if platform.IsZoomed(platform.HWND(w.WinId())) {
			return true
		}
	}
	return w.IsMaximized() || w.IsFullScreen()
}

// clearRoundedCorners removes whatever rounding was applied, making the window
// rectangular again (a maximized or full-screen window, or a radius of 0).
func (w *FramelessWindow) clearRoundedCorners() {
	if runtime.GOOS == "windows" {
		platform.ClearWindowRegion(platform.HWND(w.WinId()))
		return
	}
	if runtime.GOOS == "linux" {
		w.ClearMask()
	}
}

// adjustMaximizedClientRect insets the WM_NCCALCSIZE client rect by the resize
// border thickness while the window is maximized (and not full-screen), so the
// invisible DWM-extended borders do not make the visible client area overflow
// the monitor work area. It returns the WM_NCCALCSIZE LRESULT.
func (w *FramelessWindow) adjustMaximizedClientRect(hwnd platform.HWND, wParam, lParam uintptr) int32 {
	if !platform.IsZoomed(hwnd) || w.IsFullScreen() {
		if wParam != 0 {
			return platform.WVR_REDRAW
		}
		return 0
	}

	// SM_CXFRAME + SM_CXPADDEDBORDER (horizontal) / SM_CYFRAME +
	// SM_CXPADDEDBORDER (vertical) is the thickness of the invisible resize
	// border DWM paints for a WS_THICKFRAME window.
	tx := int32(platform.GetSystemMetrics(platform.SM_CXFRAME) + platform.GetSystemMetrics(platform.SM_CXPADDEDBORDER))
	ty := int32(platform.GetSystemMetrics(platform.SM_CYFRAME) + platform.GetSystemMetrics(platform.SM_CXPADDEDBORDER))
	if tx <= 0 {
		tx = 8
	}
	if ty <= 0 {
		ty = 8
	}

	if wParam != 0 {
		params := (*platform.NCCALCSIZE_PARAMS)(unsafe.Pointer(lParam))
		r := &params.Rgrc[0]
		r.Left += tx
		r.Right -= tx
		r.Top += ty
		r.Bottom -= ty
		return platform.WVR_REDRAW
	}

	rect := (*platform.RECT)(unsafe.Pointer(lParam))
	rect.Left += tx
	rect.Right -= tx
	rect.Top += ty
	rect.Bottom -= ty
	return 0
}

// nativeHitTest computes the WM_NCHITTEST hit code for the current cursor
// position. It maps Qt's global cursor position (device-independent pixels)
// into the window's client coordinates, mirroring qframelesswindow's
// ScreenToClient(hwnd, GetCursorPos()) + GetClientRect approach. The previous
// implementation compared the raw physical-pixel LParam against logical widget
// coordinates, which broke title-bar dragging on DPI-scaled displays.
func (w *FramelessWindow) nativeHitTest() int32 {
	global := qt.QCursor_Pos()       // GoGC-armed —do NOT Delete
	local := w.MapFromGlobal(global) // GoGC-armed —do NOT Delete
	return w.HitTest(local.X(), local.Y())
}

// HitTest returns the WM_NCHITTEST code a point in window coordinates would get:
// the border resize codes, HTCAPTION over the draggable part of the title bar, and
// HTCLIENT everywhere else. nativeHitTest() feeds it the real cursor position; it
// is exported so a caller can ask what a given point would do (and so the logic is
// testable) without moving the mouse.
func (w *FramelessWindow) HitTest(x, y int) int32 {
	if !w.resizeEnabled && !w.moveEnabled {
		return 0
	}

	cw, ch := w.Width(), w.Height()

	// Border / corner resizing. BORDER_WIDTH matches qframelesswindow's
	// WindowsFramelessWindow.BORDER_WIDTH (5 logical px); borders are disabled
	// while maximized or full-screen so a maximized window cannot be resized.
	if w.resizeEnabled && !w.IsMaximized() && !w.IsFullScreen() {
		const bw = 5
		onLeft := x < bw
		onRight := x > cw-bw
		onTop := y < bw
		onBottom := y > ch-bw

		switch {
		case onTop && onLeft:
			return platform.HTTOPLEFT
		case onTop && onRight:
			return platform.HTTOPRIGHT
		case onBottom && onLeft:
			return platform.HTBOTTOMLEFT
		case onBottom && onRight:
			return platform.HTBOTTOMRIGHT
		case onLeft:
			return platform.HTLEFT
		case onRight:
			return platform.HTRIGHT
		case onTop:
			return platform.HTTOP
		case onBottom:
			return platform.HTBOTTOM
		}
	}

	// Title-bar dragging.
	if w.moveEnabled && w.titleBarVisible && w.titleBar != nil {
		tb := w.titleBar
		if x >= tb.X() && x < tb.X()+tb.Width() && y >= tb.Y() && y < tb.Y()+tb.Height() {
			// Walk from the hit child up to the title bar. A control placed in the
			// title bar (system button, push button, line edit, tab bar, ...) must
			// receive its click, so hand the message back to Qt; every other part of
			// the title bar (background, icon and title labels, plain containers)
			// drags the window.
			for c := w.ChildAt(x, y); c != nil && c.UnsafePointer() != tb.UnsafePointer(); c = c.ParentWidget() {
				if isTitleBarInteractive(c) {
					return platform.HTCLIENT
				}
			}
			return platform.HTCAPTION
		}
	}

	return platform.HTCLIENT
}

// isTitleBarButton reports whether the widget is marked as a title bar button.
// window.TitleBarButton sets the "isTitleBarButton" dynamic property so this
// package can exclude those widgets from the draggable title-bar area without
// importing the window package (which would create an import cycle).
func isTitleBarButton(widget *qt.QWidget) bool {
	if widget == nil {
		return false
	}
	v := widget.Property("isTitleBarButton")
	// Property always returns a non-nil wrapper (GoGC-armed); an absent
	// property yields an invalid (null) QVariant.
	return v != nil && !v.IsNull() && v.ToBool()
}

// isTitleBarInteractive reports whether a widget inside the title bar wants mouse
// input: a focusable control (push button, line edit, combo box, tab bar, check
// box, ...) or anything that marks itself with the isTitleBarButton property, such
// as the hand-painted system buttons.
//
// Without this check every widget inside a custom title bar was treated as part of
// the drag area, so a button placed in the title bar moved the window instead of
// being clicked. A custom-painted clickable widget that takes no focus can opt out
// of dragging by setting the same property (see window.TitleBarButton).
func isTitleBarInteractive(widget *qt.QWidget) bool {
	if widget == nil {
		return false
	}
	if isTitleBarButton(widget) {
		return true
	}
	return widget.FocusPolicy() != qt.NoFocus
}
