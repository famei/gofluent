// Command rounded_corner is a test window for the rounded corners of the frameless
// windows. Windows 11 rounds a window through the DWM corner preference
// (DWMWA_WINDOW_CORNER_PREFERENCE); Windows 10 ignores that attribute, so the corners
// are cut with an equivalent window region there (SetWindowRgn). The window shows which
// OS build was detected, which of the two paths rounds it, and lets both be inspected:
//
//   - 圆角 switches the radius between 8px and 0 (square);
//   - 强制 Windows 10 方式 skips every Windows 11 API, which is how the Windows 10 path
//     can be checked on a newer system (the same as GOFLUENT_FORCE_WIN10=1 at startup);
//   - 演示菜单 adds a menu, a flyout and a dialog, whose own rounded shapes must keep
//     working next to the window corners (they are translucent and draw their own
//     shadow, so they deliberately do not use the window-region fallback).
//
// Maximizing the window drops the corners on purpose: they sit on the screen edges,
// where a rounded region would only clip the content.
package main

import (
	"fmt"
	"runtime"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	gfwindow "github.com/famei/gofluent/window"
)

// RoundedCornerWindow is the test window.
type RoundedCornerWindow struct {
	*gfwindow.FluentWidget

	status       *widgets.CaptionLabel
	cornerSwitch *widgets.SwitchButton
	legacySwitch *widgets.SwitchButton
	tip          *widgets.BodyLabel
}

func newWindow() *RoundedCornerWindow {
	w := &RoundedCornerWindow{FluentWidget: gfwindow.NewFluentWidget(nil)}
	w.SetObjectName("RoundedCornerWindow")
	w.Resize(720, 480)
	w.SetMinimumSize2(520, 360)
	w.SetWindowTitle("Rounded corner test")

	w.setupUI()
	// The corners are only applied once the window is shown (the native handle does not
	// exist before that), so the first status text is written shortly after the show.
	timer := qt.NewQTimer2(w.QObject)
	timer.SetSingleShot(true)
	timer.OnTimeout(w.updateStatus)
	timer.Start(300)
	return w
}

func (w *RoundedCornerWindow) setupUI() {
	vBox := qt.NewQVBoxLayout(w.QWidget)
	vBox.SetContentsMargins(28, 24, 28, 24)
	vBox.SetSpacing(14)

	title := widgets.NewTitleLabelText("无边框窗口圆角", w.QWidget)
	vBox.AddWidget(title.QWidget)

	tip := "Windows 11 用 DWM 圆角偏好（平滑）；Windows 10 忽略该属性，改用窗口区域（SetWindowRgn）裁出同样的圆角。"
	if runtime.GOOS != "windows" {
		tip = "Linux/X11 用控件掩码（QWidget.SetMask，X 服务器通过 Shape 扩展裁窗口）裁出圆角；Wayland 会话下由 XWayland 应用该形状。"
	}
	w.tip = widgets.NewBodyLabelText(tip+
		"\n两条回退路径（窗口区域/控件掩码）都是 1 位掩码，所以圆角边缘没有抗锯齿；DWM 圆角是平滑的。", w.QWidget)
	w.tip.SetWordWrap(true)
	vBox.AddWidget(w.tip.QWidget)

	form := qt.NewQHBoxLayout2()
	form.SetContentsMargins(0, 0, 0, 0)
	form.SetSpacing(12)

	cornerLabel := widgets.NewBodyLabelText("圆角 8px", w.QWidget)
	form.AddWidget(cornerLabel.QWidget)

	w.cornerSwitch = widgets.NewSwitchButtonText("", w.QWidget, widgets.RIGHT)
	w.cornerSwitch.SetChecked(w.CornerRadius() > 0)
	w.cornerSwitch.OnCheckedChanged(func(checked bool) {
		if checked {
			w.SetCornerRadius(widgets.DefaultCornerRadius())
		} else {
			w.SetCornerRadius(0)
		}
		w.updateStatus()
	})
	form.AddWidget(w.cornerSwitch.QWidget)

	legacyLabel := widgets.NewBodyLabelText("强制 Windows 10 方式", w.QWidget)
	form.AddWidget(legacyLabel.QWidget)

	w.legacySwitch = widgets.NewSwitchButtonText("", w.QWidget, widgets.RIGHT)
	w.legacySwitch.SetChecked(!gfwindow.IsWindows11())
	w.legacySwitch.OnCheckedChanged(func(checked bool) {
		gfwindow.SetForceWindows10(checked)
		// The corners of an open window were already applied: re-apply them for the
		// new mode, and let the FluentWidget re-evaluate its backdrop.
		w.RefreshRoundedCorners()
		w.SetMicaEffectEnabled(true)
		w.updateStatus()
	})
	// The forced mode only exists on Windows: Linux rounds through the window mask and
	// has nothing to force.
	legacyLabel.SetVisible(runtime.GOOS == "windows")
	w.legacySwitch.SetVisible(runtime.GOOS == "windows")
	form.AddWidget(w.legacySwitch.QWidget)

	form.AddStretch()

	menuButton := widgets.NewPushButtonText("演示菜单", w.QWidget)
	menuButton.SetFixedHeight(34)
	menuButton.OnClicked(func() { w.showDemoMenu(menuButton.QWidget) })
	form.AddWidget(menuButton.QWidget)

	themeButton := widgets.NewPrimaryPushButtonText("切换主题", w.QWidget)
	themeButton.SetFixedHeight(34)
	themeButton.OnClicked(func() { common.ToggleTheme(true, false) })
	form.AddWidget(themeButton.QWidget)

	vBox.AddLayout2(form.QLayout, 0)

	// The status prints what the window is actually doing: the OS build, the rounding
	// path in use and whether a Mica backdrop was applied.
	w.status = widgets.NewCaptionLabelText("", w.QWidget)
	w.status.SetWordWrap(true)
	vBox.AddWidget(w.status.QWidget)

	card := widgets.NewCardWidget(w.QWidget)
	card.SetBorderRadius(8)
	cardVBox := qt.NewQVBoxLayout(card.QWidget)
	cardVBox.SetContentsMargins(20, 18, 20, 18)
	cardVBox.SetSpacing(10)
	cardVBox.AddWidget(widgets.NewStrongBodyLabelText("检查要点", card.QWidget).QWidget)
	dragTip := "• 拖动标题栏、拖四边缩放仍然正常（圆角不影响 WM_NCHITTEST）；"
	forcedTip := "• 切换主题、切换“强制 Windows 10 方式”后圆角与背景色都保持正确。"
	if runtime.GOOS != "windows" {
		dragTip = "• 拖动标题栏、拖四边缩放仍然正常（走 Qt 的 QWindow.startSystemMove 与窗口自绘缩放，非 Windows 没有 WM_NCHITTEST）；"
		forcedTip = "• 切换主题后圆角与背景色都保持正确。"
	}
	for _, line := range []string{
		"• 窗口四角是圆角，标题栏贴住窗口上沿；",
		dragTip,
		"• 最大化后四角变直角（贴住屏幕边缘），还原后恢复圆角；",
		forcedTip,
	} {
		label := widgets.NewBodyLabelText(line, card.QWidget)
		label.SetWordWrap(true)
		cardVBox.AddWidget(label.QWidget)
	}
	vBox.AddWidget2(card.QWidget, 1)

	// Re-apply the status text when the theme changes (the window itself re-applies the
	// corners by itself, on every show/resize).
	common.QConfigInstance.OnThemeChangedFor(w.QObject, func(common.Theme) { w.updateStatus() })
}

// updateStatus refreshes the diagnostic caption.
func (w *RoundedCornerWindow) updateStatus() {
	build := gfwindow.WindowsBuild()
	version := fmt.Sprintf("Windows 10 及更早 (build %d)", build)
	switch {
	case runtime.GOOS != "windows":
		version = fmt.Sprintf("Linux (平台插件 %s)", qt.QGuiApplication_PlatformName())
	case build >= 22000:
		version = fmt.Sprintf("Windows 11 (build %d)", build)
	}

	path := "未启用圆角"
	switch w.CornerRounding() {
	case widgets.CornerRoundingDWM:
		path = "DWM 圆角偏好 (Windows 11，边缘平滑)"
	case widgets.CornerRoundingRegion:
		path = "窗口区域 SetWindowRgn (Windows 10 回退，边缘无抗锯齿)"
	case widgets.CornerRoundingMask:
		path = "控件掩码 SetMask (Linux/X11 Shape，边缘无抗锯齿)"
	}

	// The Mica backdrop only exists on Windows 11, and this library keeps the window
	// background solid even then (the backdrop cannot show through the opaque Qt surface,
	// see window.FluentWidget.normalBackgroundColor), so the line reports the request.
	mica := "未请求"
	switch {
	case !w.IsMicaEffectEnabled():
	case gfwindow.IsWindows11():
		mica = "已请求（Windows 11 Mica；背景按本库策略保持纯色）"
	default:
		mica = "不可用（Windows 10 没有 Mica，保持纯色背景）"
	}

	line := fmt.Sprintf("系统：%s　|　圆角：%dpx，%s　|　Mica：%s",
		version, w.CornerRadius(), path, mica)
	if runtime.GOOS == "windows" {
		line += fmt.Sprintf("　|　强制 Win10：%v", !gfwindow.IsWindows11())
	}
	w.status.SetText(line)
}

// showDemoMenu opens a round menu (with a sub menu) so the popup's own rounded shape
// and shadow can be compared with the window corners.
func (w *RoundedCornerWindow) showDemoMenu(anchor *qt.QWidget) {
	menu := widgets.NewRoundMenu("", w.QWidget)
	menu.AddAction(common.NewActionFluentIcon(common.Copy, "复制", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Cut, "剪切", nil).QAction)

	sub := widgets.NewRoundMenu("更多", w.QWidget)
	sub.AddAction(common.NewActionFluentIcon(common.Settings, "设置", nil).QAction)
	sub.AddAction(common.NewActionFluentIcon(common.Help, "帮助", nil).QAction)
	menu.AddMenu(sub)

	pos := anchor.MapToGlobal(qt.NewQPoint2(0, anchor.Height()))
	menu.Exec(pos, widgets.MenuAnimationDropDown)
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
