package dialog_box

import (
	"runtime"
	"unsafe"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/internal/platform"
	qt "github.com/mappu/miqt/qt"
)

// Frameless-window helpers shared by Dialog and a parent-less MessageBoxBase.
//
// Both are frameless (Qt.FramelessWindowHint | Qt.Dialog) translucent windows
// with no native title bar, so the window has to answer the two messages Windows
// uses to decide how the mouse interacts with it:
//
//   - WM_NCHITTEST: report HTCAPTION while the cursor is over the manual title
//     strip, which is what makes the window draggable (the strip must not contain
//     interactive children, or their clicks would start a move instead).
//   - WM_NCCALCSIZE: keep the full client rect, so the content reaches the window
//     edges instead of being inset by a native caption/border.

// installFramelessDrag makes win draggable by dragging titleWidget.
//
// Windows is answered through WM_NCHITTEST (the window manager then runs its own move loop).
// Everywhere else that message does not exist, so the drag is started from the mouse events
// of the title strip instead (installQtDrag).
func installFramelessDrag(win *qt.QDialog, titleWidget *qt.QWidget) {
	if runtime.GOOS != "windows" {
		installQtDrag(win, titleWidget)
		return
	}
	win.OnNativeEvent(func(super func(eventType []byte, message unsafe.Pointer, result *int64) bool,
		eventType []byte, message unsafe.Pointer, result *int64) bool {

		if string(eventType) == "windows_generic_MSG" {
			msg := (*platform.MSG)(message)
			switch msg.Message {
			case platform.WM_NCHITTEST:
				if code := framelessHitTest(win, titleWidget); code != 0 {
					// Writing the hit code through the low 4 bytes avoids
					// overflowing the 4-byte long slot.
					*(*int32)(unsafe.Pointer(result)) = int32(code)
					return true
				}
			case platform.WM_NCCALCSIZE:
				if msg.WParam != 0 {
					*(*int32)(unsafe.Pointer(result)) = int32(platform.WVR_REDRAW)
				} else {
					*(*int32)(unsafe.Pointer(result)) = 0
				}
				return true
			}
		}
		return super(eventType, message, result)
	})
}

// installQtDrag makes the title strip drag the dialog where WM_NCHITTEST does not exist: a
// left press on the strip arms the drag and the first move hands it to the platform's move
// loop through QWindow::startSystemMove, exactly like widgets.FramelessWindow does for its
// own title bar.
//
// A platform without a window manager move loop (WSLg's Weston, for one) accepts the call
// but never moves anything, so the moves that still arrive here drag the window directly
// instead. Only one of the two can act: as soon as the platform takes the drag over it grabs
// the pointer and no further move events reach the widget.
//
// The filter watches the strip widget itself, so a control placed on the strip (a button)
// still receives its own clicks instead of starting a move.
func installQtDrag(win *qt.QDialog, dragArea *qt.QWidget) {
	if dragArea == nil {
		return
	}
	filter := qt.NewQObject2(dragArea.QObject)
	dragArea.InstallEventFilter(filter)

	var (
		armed         bool
		dragging      bool
		startGlobalX  int
		startGlobalY  int
		startWinX     int
		startWinY     int
	)
	filter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if me := mouseEventOf(event); me != nil {
			switch me.Type() {
			case qt.QEvent__MouseButtonPress:
				armed = me.Button() == qt.LeftButton
			case qt.QEvent__MouseMove:
				if armed && me.Buttons()&qt.LeftButton != 0 {
					armed = false
					dragging = true
					g := me.GlobalPos()
					startGlobalX, startGlobalY = g.X(), g.Y()
					startWinX, startWinY = win.X(), win.Y()
					if handle := win.WindowHandle(); handle != nil {
						handle.StartSystemMove()
					}
				} else if dragging {
					g := me.GlobalPos()
					win.Move(startWinX+g.X()-startGlobalX, startWinY+g.Y()-startGlobalY)
				}
			case qt.QEvent__MouseButtonRelease:
				armed = false
				dragging = false
			}
		}
		return super(watched, event)
	})
}

// framelessHitTest returns HTCAPTION while the cursor is over titleWidget and
// HTCLIENT otherwise (the windows are fixed-size, so no resize codes).
func framelessHitTest(win *qt.QDialog, titleWidget *qt.QWidget) int32 {
	if titleWidget == nil {
		return platform.HTCLIENT
	}

	global := qt.QCursor_Pos()        // GoGC-armed — do NOT Delete
	local := win.MapFromGlobal(global) // GoGC-armed — do NOT Delete
	x, y := local.X(), local.Y()

	if x >= titleWidget.X() && x < titleWidget.X()+titleWidget.Width() &&
		y >= titleWidget.Y() && y < titleWidget.Y()+titleWidget.Height() {
		return platform.HTCAPTION
	}
	return platform.HTCLIENT
}

// installFramelessShadow applies the native shadow and the Windows 11 rounded-corner
// preference once the window's native handle is available.
//
// The dialogs are translucent and paint their own rounded shape with a shadow inside
// their rectangle, so only the DWM preference is requested: the Windows 10 window-region
// fallback would clip that shadow and make the smooth corners jagged.
//
// Linux has neither DWM nor a window manager that rounds the window for us, so there the
// corners are cut out of the window itself with a widget mask (the same mechanism
// widgets.FramelessWindow uses).
func installFramelessShadow(win *qt.QDialog) {
	win.OnShowEvent(func(super func(e *qt.QShowEvent), e *qt.QShowEvent) {
		super(e)
		if platform.SetWindowRoundedMask(win.QWidget, widgets.DefaultCornerRadius()) {
			return
		}
		hwnd := platform.HWND(win.WinId())
		style := platform.GetWindowLongPtr(hwnd, platform.GWL_STYLE)
		// WS_CAPTION keeps the native shadow/frame extension from leaving a
		// resize-border inset (the WM_NCCALCSIZE handler then returns the full
		// client rect), matching qframelesswindow's FramelessDialog.
		style |= platform.WS_THICKFRAME | platform.WS_CAPTION
		platform.SetWindowLongPtr(hwnd, platform.GWL_STYLE, style)
		_ = platform.DwmExtendFrameIntoClientArea(hwnd, &platform.MARGINS{Left: -1, Right: -1, Top: -1, Bottom: -1})
		platform.EnableDWMCornerPreference(hwnd)
	})
}
