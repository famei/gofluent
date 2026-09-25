//go:build windows

package platform

import (
	"os"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"

	qt "github.com/mappu/miqt/qt"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	dwmapi = windows.NewLazySystemDLL("dwmapi.dll")
	gdi32  = windows.NewLazySystemDLL("gdi32.dll")

	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procGetWindowRect                 = user32.NewProc("GetWindowRect")
	procIsZoomed                      = user32.NewProc("IsZoomed")
	procGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
	procSendMessageW                  = user32.NewProc("SendMessageW")
	procReleaseCapture                = user32.NewProc("ReleaseCapture")
	procGetCursorPos                  = user32.NewProc("GetCursorPos")
	procRedrawWindow                  = user32.NewProc("RedrawWindow")
	procGetParent                     = user32.NewProc("GetParent")
	procSetWindowPos                  = user32.NewProc("SetWindowPos")
	procGetWindowLongPtrW             = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW             = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW               = user32.NewProc("CallWindowProcW")
	procSetWindowCompositionAttribute = user32.NewProc("SetWindowCompositionAttribute")
	procSetWindowRgn                  = user32.NewProc("SetWindowRgn")
	procGetWindowRgn                  = user32.NewProc("GetWindowRgn")
	procPtInRegion                    = gdi32.NewProc("PtInRegion")
	procCreateRoundRectRgn            = gdi32.NewProc("CreateRoundRectRgn")
	procCreateRectRgn                 = gdi32.NewProc("CreateRectRgn")
	procDeleteObject                  = gdi32.NewProc("DeleteObject")

	procDwmExtendFrameIntoClientArea = dwmapi.NewProc("DwmExtendFrameIntoClientArea")
	procDwmSetWindowAttribute        = dwmapi.NewProc("DwmSetWindowAttribute")
)

// windowsBuild caches the OS build number (0 = not read yet, 1 = unknown).
var windowsBuild uint32

// forceWindows10 makes IsWindows11 report false, so the Windows 10 code paths can be
// tested on a newer system.
var forceWindows10 atomic.Bool

// SetForceWindows10 forces the Windows 10 behaviour (no Mica backdrop, region based
// rounded corners) on any Windows version. It exists so the fallback can be tested on
// a Windows 11 machine; the GOFLUENT_FORCE_WIN10 environment variable sets the same
// flag before the first window is created.
func SetForceWindows10(force bool) { forceWindows10.Store(force) }

// ForceWindows10 reports whether the Windows 10 behaviour is forced.
func ForceWindows10() bool {
	if forceWindows10.Load() {
		return true
	}
	return os.Getenv("GOFLUENT_FORCE_WIN10") != ""
}

// WindowsBuild returns the Windows build number (0 when it cannot be read).
// RtlGetVersion is used instead of GetVersionEx: the latter reports 6.2 for every
// Windows 10+ process that is not manifested for the current OS.
func WindowsBuild() uint32 {
	if windowsBuild != 0 {
		return windowsBuild
	}
	info := windows.RtlGetVersion()
	if info == nil {
		windowsBuild = 1 // unknown, reported as "not Windows 11"
		return windowsBuild
	}
	windowsBuild = uint32(info.BuildNumber)
	if windowsBuild == 0 {
		windowsBuild = 1
	}
	return windowsBuild
}

// IsWindows11 reports whether the Windows 11 DWM features can be used (Mica backdrop
// and the DWMWA_WINDOW_CORNER_PREFERENCE rounded corners): build 22000 or newer.
func IsWindows11() bool {
	if ForceWindows10() {
		return false
	}
	return WindowsBuild() >= 22000
}

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

// RedrawWindow invalidates the window - optionally its frame and children -
// immediately. Used after moving a frameless window: DWM recomputes the frame and its
// shadow on a size change but not always on a move, which leaves a stale frame/shadow
// painted where the window used to be.
func RedrawWindow(hwnd HWND, flags uint32) {
	_, _, _ = procRedrawWindow.Call(uintptr(hwnd), 0, 0, uintptr(flags))
}

// GetParent returns the parent (owner) HWND of hwnd (0 when it has none).
func GetParent(hwnd HWND) HWND {
	r, _, _ := procGetParent.Call(uintptr(hwnd))
	return HWND(r)
}

// SetWindowPos moves/resizes a window. For a child window the coordinates are relative
// to its parent's client area; for an owned or top-level window they are screen
// coordinates.
func SetWindowPos(hwnd HWND, x, y, w, h int, flags uint32) {
	_, _, _ = procSetWindowPos.Call(uintptr(hwnd), 0, uintptr(int32(x)), uintptr(int32(y)),
		uintptr(int32(w)), uintptr(int32(h)), uintptr(flags))
}

// GetCursorPos returns the cursor position in physical screen coordinates, the form
// WM_NCLBUTTONDOWN expects in its lParam (Qt's QCursor::pos is in logical pixels and
// would be wrong on a scaled display).
func GetCursorPos() (x, y int32) {
	var p POINT
	r, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	if r == 0 {
		return 0, 0
	}
	return p.X, p.Y
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

// EnableMica applies the Mica backdrop. Only Windows 11 has it: the attribute is
// ignored by Windows 10, and the acrylic fallback is deliberately *not* used here,
// because the acrylic accent (SetWindowCompositionAttribute) re-composites the whole
// window and breaks an opaque frameless window on Windows 10 — it loses its shadow and
// leaves repaint artefacts. Call EnableAcrylic explicitly for a transparent window that
// wants blur behind it.
//
// It reports whether a backdrop was applied.
func EnableMica(hwnd HWND, dark bool) bool {
	if !IsWindows11() {
		return false
	}
	setBoolAttr(hwnd, DWMWA_USE_IMMERSIVE_DARK_MODE, dark)

	// Preferred: Windows 11 22H2+ official system backdrop (Mica).
	bt := uint32(DWMSBT_MAINWINDOW)
	if DwmSetWindowAttribute(hwnd, DWMWA_SYSTEMBACKDROP_TYPE,
		unsafe.Pointer(&bt), unsafe.Sizeof(bt)) == nil {
		return true
	}

	// Fallback: Windows 11 21H2 legacy Mica attribute.
	one := uint32(1)
	return DwmSetWindowAttribute(hwnd, DWMWA_MICA_EFFECT,
		unsafe.Pointer(&one), unsafe.Sizeof(one)) == nil
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

// EnableDWMCornerPreference rounds the window through the DWM corner preference, which
// only Windows 11 honours. Translucent windows that paint their own rounded shape
// (dialogs, flyouts) use this instead of EnableRoundedCorners: their shadow is drawn
// inside their own rectangle, where a window region would clip it.
//
// It reports whether the preference was accepted.
func EnableDWMCornerPreference(hwnd HWND) bool {
	if !IsWindows11() {
		return false
	}
	cp := uint32(DWMWCP_ROUND)
	return DwmSetWindowAttribute(hwnd, DWMWA_WINDOW_CORNER_PREFERENCE,
		unsafe.Pointer(&cp), unsafe.Sizeof(cp)) == nil
}

// EnableRoundedCorners rounds the window corners. Windows 11 has the DWM corner
// preference (DWMWA_WINDOW_CORNER_PREFERENCE); Windows 10 ignores that attribute, so
// the corners are cut with an equivalent window region there. radius is the corner
// radius in physical pixels.
//
// It reports whether the DWM preference was used (false = the region fallback).
func EnableRoundedCorners(hwnd HWND, radius int) bool {
	if radius <= 0 {
		ClearWindowRegion(hwnd)
		return false
	}

	if EnableDWMCornerPreference(hwnd) {
		// The DWM rounds the window; make sure no region is left over from an
		// earlier fallback (a forced Windows 10 run, or a downgrade).
		ClearWindowRegion(hwnd)
		return true
	}

	SetWindowRoundedRegion(hwnd, radius)
	return false
}

// SetWindowRoundedRegion clips the window to a rounded rectangle, the Windows 10 way
// of rounding window corners. The region is taken from the current window rectangle
// (physical pixels), so the caller does not have to track the size or the DPI.
//
// A window region is a 1-bit mask, so the corners are not anti-aliased; Windows 11 (and
// a forced Windows 10 run on it) can only be rounded this way because the DWM corner
// preference is unavailable there.
func SetWindowRoundedRegion(hwnd HWND, radius int) {
	var rect RECT
	if !GetWindowRect(hwnd, &rect) {
		return
	}
	w, h := int(rect.Right-rect.Left), int(rect.Bottom-rect.Top)
	if w <= 0 || h <= 0 {
		return
	}
	if radius < 1 {
		radius = 1
	}
	if max := w / 2; radius > max {
		radius = max
	}
	if max := h / 2; radius > max {
		radius = max
	}

	// CreateRoundRectRgn's right/bottom edges are exclusive, hence w+1/h+1.
	rgn, _, _ := procCreateRoundRectRgn.Call(0, 0, uintptr(w+1), uintptr(h+1),
		uintptr(radius*2), uintptr(radius*2))
	if rgn == 0 {
		return
	}
	// SetWindowRgn takes ownership of the region on success (it must not be deleted
	// afterwards); only a failed call leaves it to the caller.
	if r, _, _ := procSetWindowRgn.Call(uintptr(hwnd), rgn, 1); r == 0 {
		_, _, _ = procDeleteObject.Call(rgn)
	}
}

// ClearWindowRegion removes any window region, making the window rectangular again
// (used while a window is maximized or full-screen, where rounded corners would clip
// the content at the screen edges).
func ClearWindowRegion(hwnd HWND) {
	_, _, _ = procSetWindowRgn.Call(uintptr(hwnd), 0, 1)
}

// WindowRegionRoundsCorners reports whether the window currently has a region that
// excludes its top-left pixel, i.e. whether the rounded fallback is in effect. It is
// meant for diagnostics.
func WindowRegionRoundsCorners(hwnd HWND) bool {
	// GetWindowRgn copies the window's region into a region of the caller, so start
	// with the smallest possible one.
	rgn, _, _ := procCreateRectRgn.Call(0, 0, 1, 1)
	if rgn == 0 {
		return false
	}
	defer func() { _, _, _ = procDeleteObject.Call(rgn) }()

	// 0 (ERROR) means the window has no region at all.
	if r, _, _ := procGetWindowRgn.Call(uintptr(hwnd), rgn); r == 0 {
		return false
	}
	in, _, _ := procPtInRegion.Call(rgn, 1, 1)
	return in == 0
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

// SetWindowRoundedMask reports false on Windows: the corners of a native window are
// rounded by the window itself (the DWM corner preference, or a window region as its
// fallback - see EnableRoundedCorners), never by a Qt widget mask.
func SetWindowRoundedMask(w *qt.QWidget, radius int) bool { return false }