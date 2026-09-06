//go:build windows

package win32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	dwmapi = windows.NewLazySystemDLL("dwmapi.dll")

	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procGetWindowRect                 = user32.NewProc("GetWindowRect")
	procIsZoomed                      = user32.NewProc("IsZoomed")
	procGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
	procSendMessageW                  = user32.NewProc("SendMessageW")
	procReleaseCapture                = user32.NewProc("ReleaseCapture")
	procGetWindowLongPtrW             = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW             = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW               = user32.NewProc("CallWindowProcW")
	procSetWindowCompositionAttribute = user32.NewProc("SetWindowCompositionAttribute")

	procDwmExtendFrameIntoClientArea = dwmapi.NewProc("DwmExtendFrameIntoClientArea")
	procDwmSetWindowAttribute        = dwmapi.NewProc("DwmSetWindowAttribute")
)

// DefWindowProc calls the default window procedure for a message.
func DefWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

// GetWindowRect retrieves the bounding rectangle of a window in screen
// coordinates.
func GetWindowRect(hwnd HWND, rect *RECT) bool {
	r, _, _ := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)))
	return r != 0
}

// IsZoomed reports whether the window is currently maximized. Unlike Qt's
// QWidget::isMaximized (which lags behind during the WM_NCCALCSIZE of a
// maximize/restore transition), this reads the live WS_MAXIMIZE style bit.
func IsZoomed(hwnd HWND) bool {
	r, _, _ := procIsZoomed.Call(uintptr(hwnd))
	return r != 0
}

// GetSystemMetrics returns the value of the requested system metric.
func GetSystemMetrics(index int) int {
	r, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int(r)
}

// SendMessage sends a message to a window synchronously.
func SendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

// ReleaseCapture releases the mouse capture from the current thread's window.
func ReleaseCapture() {
	_, _, _ = procReleaseCapture.Call()
}

// GetWindowLongPtr retrieves a window-long value (style, ex-style, ...).
func GetWindowLongPtr(hwnd HWND, index int) uintptr {
	r, _, _ := procGetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index))
	return r
}

// SetWindowLongPtr sets a window-long value and returns the previous value.
func SetWindowLongPtr(hwnd HWND, index int, newLong uintptr) uintptr {
	r, _, _ := procSetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index), newLong)
	return r
}

// CallWindowProcW calls a previous window procedure (subclass chain).
func CallWindowProcW(prevWndProc uintptr, hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procCallWindowProcW.Call(prevWndProc, uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

// DwmExtendFrameIntoClientArea extends the DWM frame into the client area,
// producing a native shadow and rounded corners without a visible border.
func DwmExtendFrameIntoClientArea(hwnd HWND, margins *MARGINS) error {
	r, _, _ := procDwmExtendFrameIntoClientArea.Call(uintptr(hwnd), uintptr(unsafe.Pointer(margins)))
	return hresult(r)
}

// DwmSetWindowAttribute sets a DWM window attribute (Mica/Acrylic/dark mode/
// corner preference).
func DwmSetWindowAttribute(hwnd HWND, attr uint32, value unsafe.Pointer, size uintptr) error {
	r, _, _ := procDwmSetWindowAttribute.Call(uintptr(hwnd), uintptr(attr), uintptr(value), size)
	return hresult(r)
}

// SetWindowCompositionAttribute sets a Windows 10 composition attribute
// (acrylic / Aero backdrop).
func SetWindowCompositionAttribute(hwnd HWND, data *WINDOWCOMPOSITIONATTRIBDATA) error {
	r, _, _ := procSetWindowCompositionAttribute.Call(uintptr(hwnd), uintptr(unsafe.Pointer(data)))
	if r == 0 {
		return windows.GetLastError()
	}
	return nil
}

// EnableMica applies the Mica backdrop on Windows 11, degrading to the legacy
// Mica attribute and finally to Windows 10 acrylic on older systems.
func EnableMica(hwnd HWND, dark bool) {
	setBoolAttr(hwnd, DWMWA_USE_IMMERSIVE_DARK_MODE, dark)

	// Preferred: Windows 11 22H2+ official system backdrop (Mica).
	bt := uint32(DWMSBT_MAINWINDOW)
	if DwmSetWindowAttribute(hwnd, DWMWA_SYSTEMBACKDROP_TYPE,
		unsafe.Pointer(&bt), unsafe.Sizeof(bt)) == nil {
		return
	}

	// Fallback: Windows 11 21H2 legacy Mica attribute.
	one := uint32(1)
	if DwmSetWindowAttribute(hwnd, DWMWA_MICA_EFFECT,
		unsafe.Pointer(&one), unsafe.Sizeof(one)) == nil {
		return
	}

	// Fallback: Windows 10 acrylic.
	EnableAcrylic(hwnd, 0xCC000000 /* ABGR semi-transparent black */)
}

// EnableAcrylic applies the Windows 10 acrylic (or Aero) backdrop. The
// gradientColor is an ABGR value (0xAABBGGRR).
func EnableAcrylic(hwnd HWND, gradientColor uint32) {
	accent := ACCENT_POLICY{
		AccentState:   ACCENT_ENABLE_ACRYLICBLURBEHIND,
		GradientColor: gradientColor,
	}
	data := WINDOWCOMPOSITIONATTRIBDATA{
		Attrib: WCA_ACCENT_POLICY,
		Data:   unsafe.Pointer(&accent),
		Size:   unsafe.Sizeof(accent),
	}
	_ = SetWindowCompositionAttribute(hwnd, &data)
}

// RemoveBackdrop disables every backdrop effect applied to the window.
func RemoveBackdrop(hwnd HWND) {
	bt := uint32(DWMSBT_NONE)
	_ = DwmSetWindowAttribute(hwnd, DWMWA_SYSTEMBACKDROP_TYPE,
		unsafe.Pointer(&bt), unsafe.Sizeof(bt))
	zero := uint32(0)
	_ = DwmSetWindowAttribute(hwnd, DWMWA_MICA_EFFECT,
		unsafe.Pointer(&zero), unsafe.Sizeof(zero))

	accent := ACCENT_POLICY{AccentState: ACCENT_DISABLED}
	data := WINDOWCOMPOSITIONATTRIBDATA{
		Attrib: WCA_ACCENT_POLICY,
		Data:   unsafe.Pointer(&accent),
		Size:   unsafe.Sizeof(accent),
	}
	_ = SetWindowCompositionAttribute(hwnd, &data)
}

// EnableRoundedCorners requests the Windows 11 rounded-corner preference.
func EnableRoundedCorners(hwnd HWND) {
	cp := uint32(DWMWCP_ROUND)
	_ = DwmSetWindowAttribute(hwnd, DWMWA_WINDOW_CORNER_PREFERENCE,
		unsafe.Pointer(&cp), unsafe.Sizeof(cp))
}

func setBoolAttr(hwnd HWND, attr uint32, on bool) {
	v := uint32(0)
	if on {
		v = 1
	}
	_ = DwmSetWindowAttribute(hwnd, attr, unsafe.Pointer(&v), unsafe.Sizeof(v))
}

func hresult(r uintptr) error {
	// HRESULT: S_OK == 0; a non-zero value (high bit set) denotes failure.
	if r == 0 {
		return nil
	}
	return windows.Errno(r)
}
