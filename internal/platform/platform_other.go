//go:build !windows

package platform

import (
	"unsafe"

	qt "github.com/mappu/miqt/qt"
)

// On non-Windows platforms the frameless-window effects are unsupported; every
// function is a no-op so the rest of the library keeps compiling.

func DefWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr { return 0 }

func GetWindowRect(hwnd HWND, rect *RECT) bool { return false }

func IsZoomed(hwnd HWND) bool { return false }

func GetSystemMetrics(index int) int { return 0 }

func SendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr { return 0 }

func ReleaseCapture() {}

// RedrawWindow is a no-op off Windows (it exists for the stale-frame repaint after a
// frameless window move).
func RedrawWindow(hwnd HWND, flags uint32) {}

// GetParent reports no parent off Windows.
func GetParent(hwnd HWND) HWND { return 0 }

// SetWindowPos is a no-op off Windows.
func SetWindowPos(hwnd HWND, x, y, w, h int, flags uint32) {}

// GetCursorPos is a no-op off Windows.
func GetCursorPos() (x, y int32) { return 0, 0 }

func GetWindowLongPtr(hwnd HWND, index int) uintptr { return 0 }

func SetWindowLongPtr(hwnd HWND, index int, newLong uintptr) uintptr { return 0 }

func CallWindowProcW(prevWndProc uintptr, hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	return 0
}

func DwmExtendFrameIntoClientArea(hwnd HWND, margins *MARGINS) error { return nil }

func DwmSetWindowAttribute(hwnd HWND, attr uint32, value unsafe.Pointer, size uintptr) error {
	return nil
}

func SetWindowCompositionAttribute(hwnd HWND, data *WINDOWCOMPOSITIONATTRIBDATA) error {
	return nil
}

func EnableMica(hwnd HWND, dark bool) bool { return false }

func EnableAcrylic(hwnd HWND, gradientColor uint32) {}

func RemoveBackdrop(hwnd HWND) {}

func EnableRoundedCorners(hwnd HWND, radius int) bool { return false }

func EnableDWMCornerPreference(hwnd HWND) bool { return false }

func SetWindowRoundedRegion(hwnd HWND, radius int) {}

func ClearWindowRegion(hwnd HWND) {}

func WindowRegionRoundsCorners(hwnd HWND) bool { return false }

// WindowsBuild reports 0 off Windows (no Windows build).
func WindowsBuild() uint32 { return 0 }

// IsWindows11 is false off Windows.
func IsWindows11() bool { return false }

// SetForceWindows10 is a no-op off Windows.
func SetForceWindows10(force bool) {}

// ForceWindows10 is always false off Windows.
func ForceWindows10() bool { return false }

// setWindowRoundedMask implements SetWindowRoundedMask on this system. The Linux build
// replaces it with the real mask (see platform_linux.go); every other system keeps the
// no-op, because rounding a window is either a Windows (DWM/region) or a Linux (X11 mask)
// concern.
var setWindowRoundedMask = func(w *qt.QWidget, radius int) bool { return false }

// SetWindowRoundedMask rounds the corners of a top-level window through whatever the
// system provides for it. It reports whether a widget mask is the mechanism here; on
// Windows the native window is rounded instead (the DWM corner preference, or a window
// region as its fallback), so callers keep their native path there.
func SetWindowRoundedMask(w *qt.QWidget, radius int) bool { return setWindowRoundedMask(w, radius) }
