package dialog_box

import (
	"unsafe"

	"github.com/famei/gofluent/internal/win32"
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
func installFramelessDrag(win *qt.QDialog, titleWidget *qt.QWidget) {
	win.OnNativeEvent(func(super func(eventType []byte, message unsafe.Pointer, result *int64) bool,
		eventType []byte, message unsafe.Pointer, result *int64) bool {

		if string(eventType) == "windows_generic_MSG" {
			msg := (*win32.MSG)(message)
			switch msg.Message {
			case win32.WM_NCHITTEST:
				if code := framelessHitTest(win, titleWidget); code != 0 {
					// Writing the hit code through the low 4 bytes avoids
					// overflowing the 4-byte long slot.
					*(*int32)(unsafe.Pointer(result)) = int32(code)
					return true
				}
			case win32.WM_NCCALCSIZE:
				if msg.WParam != 0 {
					*(*int32)(unsafe.Pointer(result)) = int32(win32.WVR_REDRAW)
				} else {
					*(*int32)(unsafe.Pointer(result)) = 0
				}
				return true
			}
		}
		return super(eventType, message, result)
	})
}

// framelessHitTest returns HTCAPTION while the cursor is over titleWidget and
// HTCLIENT otherwise (the windows are fixed-size, so no resize codes).
func framelessHitTest(win *qt.QDialog, titleWidget *qt.QWidget) int32 {
	if titleWidget == nil {
		return win32.HTCLIENT
	}

	global := qt.QCursor_Pos()        // GoGC-armed — do NOT Delete
	local := win.MapFromGlobal(global) // GoGC-armed — do NOT Delete
	x, y := local.X(), local.Y()

	if x >= titleWidget.X() && x < titleWidget.X()+titleWidget.Width() &&
		y >= titleWidget.Y() && y < titleWidget.Y()+titleWidget.Height() {
		return win32.HTCAPTION
	}
	return win32.HTCLIENT
}

// installFramelessShadow applies the native shadow and the Windows 11
// rounded-corner preference once the window's native handle is available.
func installFramelessShadow(win *qt.QDialog) {
	win.OnShowEvent(func(super func(e *qt.QShowEvent), e *qt.QShowEvent) {
		super(e)
		hwnd := win32.HWND(win.WinId())
		style := win32.GetWindowLongPtr(hwnd, win32.GWL_STYLE)
		// WS_CAPTION keeps the native shadow/frame extension from leaving a
		// resize-border inset (the WM_NCCALCSIZE handler then returns the full
		// client rect), matching qframelesswindow's FramelessDialog.
		style |= win32.WS_THICKFRAME | win32.WS_CAPTION
		win32.SetWindowLongPtr(hwnd, win32.GWL_STYLE, style)
		_ = win32.DwmExtendFrameIntoClientArea(hwnd, &win32.MARGINS{Left: -1, Right: -1, Top: -1, Bottom: -1})
		win32.EnableRoundedCorners(hwnd)
	})
}
