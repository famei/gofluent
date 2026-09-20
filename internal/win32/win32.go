// Package win32 provides the small set of Win32 API bindings needed by the
// gofluent frameless-window implementation (WM_NCHITTEST hit-testing, DWM
// shadows/backdrops and the Windows 10 acrylic accent policy). The bindings are
// pure Go on top of golang.org/x/sys/windows and carry no CGO dependencies, so
// they cross-compile for windows/amd64 from WSL/MXE without extra link flags.
//
// Constants and structure types live in this build-agnostic file so the rest of
// the library can reference them on every platform; the actual function
// implementations are split between win32_windows.go and win32_other.go.
package win32

import "unsafe"

// HWND is a native window handle.
type HWND uintptr

// Window messages.
const (
	WM_NCHITTEST       = 0x0084
	WM_NCCALCSIZE      = 0x0083
	WM_NCLBUTTONDOWN   = 0x00A1

	// RedrawWindow flags.
	RDW_INVALIDATE = 0x0001
	RDW_UPDATENOW  = 0x0100
	RDW_ALLCHILDREN = 0x0080
	RDW_FRAME      = 0x0400

	// SetWindowPos flags.
	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010
	WM_NCLBUTTONUP     = 0x00A2
	WM_NCLBUTTONDBLCLK = 0x00A3
	WM_NCMOUSELEAVE    = 0x02A2
)

// WM_NCCALCSIZE return value: repaint the frame after the client rect is
// adjusted.
const (
	WVR_REDRAW = 0x0200
)

// WM_NCHITTEST return codes.
const (
	HTERROR       = -2
	HTTRANSPARENT = -1
	HTNOWHERE     = 0
	HTCLIENT      = 1
	HTCAPTION     = 2
	HTSYSMENU     = 3
	HTGROWBOX     = 4
	HTMENU        = 5
	HTMINBUTTON   = 8
	HTMAXBUTTON   = 9
	HTLEFT        = 10
	HTRIGHT       = 11
	HTTOP         = 12
	HTTOPLEFT     = 13
	HTTOPRIGHT    = 14
	HTBOTTOM      = 15
	HTBOTTOMLEFT  = 16
	HTBOTTOMRIGHT = 17
	HTBORDER      = 18
	HTCLOSE       = 20
)

// GetSystemMetrics indices.
const (
	SM_CXFRAME        = 32
	SM_CYFRAME        = 33
	SM_CXPADDEDBORDER = 92
)

// Window style / GetWindowLongPtr indices.
const (
	GWL_STYLE    = -16
	GWL_EXSTYLE  = -20
	GWLP_WNDPROC = -4

	WS_THICKFRAME  = 0x00040000
	WS_CAPTION     = 0x00C00000
	WS_MAXIMIZEBOX = 0x00010000
	WS_MINIMIZEBOX = 0x00020000
	WS_EX_LAYERED  = 0x00080000
)

// DWM window attributes (DWMWINDOWATTRIBUTE).
const (
	DWMWA_USE_IMMERSIVE_DARK_MODE  = 20
	DWMWA_WINDOW_CORNER_PREFERENCE = 33
	DWMWA_BORDER_COLOR             = 34
	DWMWA_SYSTEMBACKDROP_TYPE      = 38
	DWMWA_MICA_EFFECT              = 1029
)

// DWM_SYSTEMBACKDROP_TYPE values.
const (
	DWMSBT_AUTO            = 0
	DWMSBT_NONE            = 1
	DWMSBT_MAINWINDOW      = 2 // Mica
	DWMSBT_TRANSIENTWINDOW = 3 // Acrylic
	DWMSBT_TABBEDWINDOW    = 4 // Mica Alt
)

// DWM_WINDOW_CORNER_PREFERENCE values.
const (
	DWMWCP_DEFAULT    = 0
	DWMWCP_DONOTROUND = 1
	DWMWCP_ROUND      = 2
	DWMWCP_ROUNDSMALL = 3
)

// SetWindowCompositionAttribute attribute and accent state values.
const (
	WCA_ACCENT_POLICY                 = 19
	ACCENT_DISABLED                   = 0
	ACCENT_ENABLE_GRADIENT            = 1
	ACCENT_ENABLE_TRANSPARENTGRADIENT = 2
	ACCENT_ENABLE_BLURBEHIND          = 3 // Windows 7 Aero
	ACCENT_ENABLE_ACRYLICBLURBEHIND   = 4 // Windows 10 acrylic
	ACCENT_ENABLE_HOSTBACKDROP        = 5
)

// MSG mirrors the Win32 MSG structure (x64 layout: 48 bytes).
type MSG struct {
	HWnd    uintptr
	Message uint32
	_       uint32 // padding aligning WParam to 8 bytes
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// POINT mirrors the Win32 POINT structure.
type POINT struct {
	X int32
	Y int32
}

// RECT mirrors the Win32 RECT structure.
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// MARGINS mirrors the DWM MARGINS structure (cxLeftWidth, cxRightWidth,
// cyTopHeight, cyBottomHeight).
type MARGINS struct {
	Left   int32
	Right  int32
	Top    int32
	Bottom int32
}

// NCCALCSIZE_PARAMS mirrors the Win32 NCCALCSIZE_PARAMS structure. Only the
// first RECT (the proposed new window rectangle, in screen coordinates) is
// used by the WM_NCCALCSIZE handler; lppos is kept for correct structure size.
type NCCALCSIZE_PARAMS struct {
	Rgrc  [3]RECT
	Lppos uintptr
}

// ACCENT_POLICY mirrors the Windows 10 ACCENTPOLICY structure.
type ACCENT_POLICY struct {
	AccentState   int32
	AccentFlags   int32
	GradientColor uint32 // ABGR: 0xAABBGGRR
	AnimationId   int32
}

// WINDOWCOMPOSITIONATTRIBDATA mirrors the Windows 10
// WINDOWCOMPOSITIONATTRIBDATA structure.
type WINDOWCOMPOSITIONATTRIBDATA struct {
	Attrib int32
	_      int32 // padding aligning Data to 8 bytes
	Data   unsafe.Pointer
	Size   uintptr
}
