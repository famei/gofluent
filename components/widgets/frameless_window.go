package widgets

import (
	"unsafe"

	"github.com/famei/gofluent/internal/win32"
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
	w.installNativeHook()
	w.installShowHook()
	w.installResizeHook()
	return w
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
	})
}

// TitleBar returns the custom title bar widget.
func (w *FramelessWindow) TitleBar() *qt.QWidget { return w.titleBar }

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
			msg := (*win32.MSG)(message)
			switch msg.Message {
			case win32.WM_NCHITTEST:
				if code := w.nativeHitTest(); code != 0 {
					// miqt Qt5 binds C's 32-bit long* as a Go *int64. Writing
					// all 8 bytes would overflow the 4-byte slot; only the low
					// 4 bytes are valid.
					*(*int32)(unsafe.Pointer(result)) = int32(code)
					return true
				}
			case win32.WM_NCCALCSIZE:
				// Adjust the maximized client rect so the invisible resize
				// border does not push the visible content past the work area
				// (under the taskbar / off-screen). Mirror qframelesswindow's
				// WM_NCCALCSIZE handler.
				code := w.adjustMaximizedClientRect(win32.HWND(msg.HWnd), msg.WParam, msg.LParam)
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
// requests the Windows 11 rounded-corner preference.
func (w *FramelessWindow) applyNativeFrame() {
	hwnd := win32.HWND(w.WinId())

	if w.resizeEnabled {
		style := win32.GetWindowLongPtr(hwnd, win32.GWL_STYLE)
		// WS_CAPTION is what makes Windows animate the maximize/restore
		// transition; the DWM frame extension below hides the caption visually
		// (mirrors qframelesswindow's addWindowAnimation).
		style |= win32.WS_THICKFRAME | win32.WS_MAXIMIZEBOX | win32.WS_MINIMIZEBOX | win32.WS_CAPTION
		win32.SetWindowLongPtr(hwnd, win32.GWL_STYLE, style)
	}
	_ = win32.DwmExtendFrameIntoClientArea(hwnd, &win32.MARGINS{Left: -1, Right: -1, Top: -1, Bottom: -1})
	win32.EnableRoundedCorners(hwnd)
}

// adjustMaximizedClientRect insets the WM_NCCALCSIZE client rect by the resize
// border thickness while the window is maximized (and not full-screen), so the
// invisible DWM-extended borders do not make the visible client area overflow
// the monitor work area. It returns the WM_NCCALCSIZE LRESULT.
func (w *FramelessWindow) adjustMaximizedClientRect(hwnd win32.HWND, wParam, lParam uintptr) int32 {
	if !win32.IsZoomed(hwnd) || w.IsFullScreen() {
		if wParam != 0 {
			return win32.WVR_REDRAW
		}
		return 0
	}

	// SM_CXFRAME + SM_CXPADDEDBORDER (horizontal) / SM_CYFRAME +
	// SM_CXPADDEDBORDER (vertical) is the thickness of the invisible resize
	// border DWM paints for a WS_THICKFRAME window.
	tx := int32(win32.GetSystemMetrics(win32.SM_CXFRAME) + win32.GetSystemMetrics(win32.SM_CXPADDEDBORDER))
	ty := int32(win32.GetSystemMetrics(win32.SM_CYFRAME) + win32.GetSystemMetrics(win32.SM_CXPADDEDBORDER))
	if tx <= 0 {
		tx = 8
	}
	if ty <= 0 {
		ty = 8
	}

	if wParam != 0 {
		params := (*win32.NCCALCSIZE_PARAMS)(unsafe.Pointer(lParam))
		r := &params.Rgrc[0]
		r.Left += tx
		r.Right -= tx
		r.Top += ty
		r.Bottom -= ty
		return win32.WVR_REDRAW
	}

	rect := (*win32.RECT)(unsafe.Pointer(lParam))
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
	if !w.resizeEnabled && !w.moveEnabled {
		return 0
	}

	global := qt.QCursor_Pos()       // GoGC-armed — do NOT Delete
	local := w.MapFromGlobal(global) // GoGC-armed — do NOT Delete
	x, y := local.X(), local.Y()
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
			return win32.HTTOPLEFT
		case onTop && onRight:
			return win32.HTTOPRIGHT
		case onBottom && onLeft:
			return win32.HTBOTTOMLEFT
		case onBottom && onRight:
			return win32.HTBOTTOMRIGHT
		case onLeft:
			return win32.HTLEFT
		case onRight:
			return win32.HTRIGHT
		case onTop:
			return win32.HTTOP
		case onBottom:
			return win32.HTBOTTOM
		}
	}

	// Title-bar dragging.
	if w.moveEnabled && w.titleBarVisible && w.titleBar != nil {
		tb := w.titleBar
		if x >= tb.X() && x < tb.X()+tb.Width() && y >= tb.Y() && y < tb.Y()+tb.Height() {
			// Walk from the hit child up to the title bar. A title bar button
			// (minimize/maximize/close) must receive its click, so hand the
			// message back to Qt; every other part of the title bar
			// (background, icon and title labels) drags the window.
			for c := w.ChildAt(x, y); c != nil && c.UnsafePointer() != tb.UnsafePointer(); c = c.ParentWidget() {
				if isTitleBarButton(c) {
					return win32.HTCLIENT
				}
			}
			return win32.HTCAPTION
		}
	}

	return win32.HTCLIENT
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
