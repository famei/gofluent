// Command login ports PyQt-Fluent-Widgets' examples/window/login demo: a
// FluentWidget with a split login form whose left half shows a full-bleed
// background image (scaled on resize) and whose right half hosts an FTP-style
// login form. The Ui_LoginWindow.py layout is translated directly to Go widget
// construction, and the resource.qrc images are loaded with go:embed.
package main

import (
	"embed"
	_ "embed"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	gfwindow "github.com/famei/gofluent/window"
)

//go:embed resource/images/*.jpg resource/images/*.png
var loginFS embed.FS

// LoginWindow is the login demo window.
type LoginWindow struct {
	*gfwindow.FluentWidget

	backgroundLabel *qt.QLabel
	logoLabel       *qt.QLabel
	formWidget      *qt.QWidget

	lineEdit   *widgets.LineEdit
	lineEdit2  *widgets.LineEdit
	lineEdit3  *widgets.LineEdit
	lineEdit4  *widgets.LineEdit
	hostLabel  *widgets.BodyLabel
	portLabel  *widgets.BodyLabel
	userLabel  *widgets.BodyLabel
	passLabel  *widgets.BodyLabel
	checkBox   *widgets.CheckBox
	pushButton *widgets.PrimaryPushButton
	linkButton *widgets.HyperlinkButton
}

func newLoginWindow() *LoginWindow {
	w := &LoginWindow{FluentWidget: gfwindow.NewFluentWidget(nil)}
	w.SetObjectName("Form")
	w.Resize(1250, 809)
	w.SetMinimumSize2(700, 500)

	w.setupUI()
	common.SetThemeColor(qt.NewQColor6("#28afe9"), false, false)

	// place the title bar on the top layer
	w.TitleBar().Raise()

	w.backgroundLabel.SetScaledContents(false)

	w.SetWindowTitle("PyQt-Fluent-Widget")
	w.SetWindowIcon(asset.QIcon(loginFS, "resource/images/logo.png"))

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)

	w.scaleBackground()
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		// Overriding OnResizeEvent replaces FramelessWindow's resize hook (miqt
		// virtuals do not chain to the previous Go handler), so re-glue the title
		// bar across the full width here — mirroring FramelessWindow.positionTitleBar.
		// Without this the system buttons keep the construction-time width and drift
		// toward the middle of the window on resize/show.
		if tb := w.TitleBar(); tb != nil {
			tb.Move(0, 0)
			tb.Resize(w.Width(), tb.Height())
		}
		w.scaleBackground()
	})
	return w
}

func (w *LoginWindow) setupUI() {
	hLayout := qt.NewQHBoxLayout(w.QWidget)
	hLayout.SetContentsMargins(0, 0, 0, 0)
	hLayout.SetSpacing(0)

	// left background image
	w.backgroundLabel = qt.NewQLabel(w.QWidget)
	w.backgroundLabel.SetText("")
	background := asset.QPixmap(loginFS, "resource/images/background.jpg")
	w.backgroundLabel.SetPixmap(background)
	background.Delete()
	w.backgroundLabel.SetScaledContents(true)
	w.backgroundLabel.SetObjectName("label")
	hLayout.AddWidget(w.backgroundLabel.QWidget)

	// right form panel
	w.formWidget = qt.NewQWidget(w.QWidget)
	w.formWidget.SetMinimumSize2(360, 0)
	w.formWidget.SetMaximumSize2(360, 16777215)
	w.formWidget.SetStyleSheet("QLabel{font: 13px 'Microsoft YaHei'}")
	w.formWidget.SetObjectName("widget")

	vLayout2 := qt.NewQVBoxLayout(w.formWidget)
	vLayout2.SetContentsMargins(20, 20, 20, 20)
	vLayout2.SetSpacing(9)

	vLayout2.AddStretchWithStretch(1)

	w.logoLabel = qt.NewQLabel(w.formWidget)
	w.logoLabel.SetMinimumSize2(100, 100)
	w.logoLabel.SetMaximumSize2(100, 100)
	logo := asset.QPixmap(loginFS, "resource/images/logo.png")
	w.logoLabel.SetPixmap(logo)
	logo.Delete()
	w.logoLabel.SetScaledContents(true)
	w.logoLabel.SetObjectName("label_2")
	vLayout2.AddWidget3(w.logoLabel.QWidget, 0, qt.AlignHCenter)

	vLayout2.AddSpacing(15)

	grid := qt.NewQGridLayout2()
	grid.SetHorizontalSpacing(4)
	grid.SetVerticalSpacing(9)

	w.lineEdit = widgets.NewLineEdit(w.formWidget)
	w.lineEdit.SetClearButtonEnabled(true)
	w.lineEdit.SetObjectName("lineEdit")
	grid.AddWidget3(w.lineEdit.QWidget, 1, 0, 1, 1)

	w.hostLabel = widgets.NewBodyLabelText("", w.formWidget)
	w.hostLabel.SetObjectName("label_3")
	grid.AddWidget3(w.hostLabel.QWidget, 0, 0, 1, 1)

	w.lineEdit2 = widgets.NewLineEdit(w.formWidget)
	w.lineEdit2.SetClearButtonEnabled(true)
	w.lineEdit2.SetObjectName("lineEdit_2")
	grid.AddWidget3(w.lineEdit2.QWidget, 1, 1, 1, 1)

	w.portLabel = widgets.NewBodyLabelText("", w.formWidget)
	w.portLabel.SetObjectName("label_4")
	grid.AddWidget3(w.portLabel.QWidget, 0, 1, 1, 1)

	grid.SetColumnStretch(0, 2)
	grid.SetColumnStretch(1, 1)
	vLayout2.AddLayout(grid.QLayout)

	w.userLabel = widgets.NewBodyLabelText("", w.formWidget)
	w.userLabel.SetObjectName("label_5")
	vLayout2.AddWidget(w.userLabel.QWidget)

	w.lineEdit3 = widgets.NewLineEdit(w.formWidget)
	w.lineEdit3.SetClearButtonEnabled(true)
	w.lineEdit3.SetObjectName("lineEdit_3")
	vLayout2.AddWidget(w.lineEdit3.QWidget)

	w.passLabel = widgets.NewBodyLabelText("", w.formWidget)
	w.passLabel.SetObjectName("label_6")
	vLayout2.AddWidget(w.passLabel.QWidget)

	w.lineEdit4 = widgets.NewLineEdit(w.formWidget)
	w.lineEdit4.SetEchoMode(qt.QLineEdit__Password)
	w.lineEdit4.SetClearButtonEnabled(true)
	w.lineEdit4.SetObjectName("lineEdit_4")
	vLayout2.AddWidget(w.lineEdit4.QWidget)

	vLayout2.AddSpacing(5)

	w.checkBox = widgets.NewCheckBox(w.formWidget)
	w.checkBox.SetChecked(true)
	w.checkBox.SetObjectName("checkBox")
	vLayout2.AddWidget(w.checkBox.QWidget)

	vLayout2.AddSpacing(5)

	w.pushButton = widgets.NewPrimaryPushButton(w.formWidget)
	w.pushButton.SetObjectName("pushButton")
	vLayout2.AddWidget(w.pushButton.QWidget)

	vLayout2.AddSpacing(6)

	w.linkButton = widgets.NewHyperlinkButton(w.formWidget)
	w.linkButton.SetObjectName("pushButton_2")
	vLayout2.AddWidget(w.linkButton.QWidget)

	vLayout2.AddStretchWithStretch(1)

	hLayout.AddWidget(w.formWidget)

	w.retranslate()
}

func (w *LoginWindow) retranslate() {
	w.lineEdit.SetPlaceholderText("ftp.example.com")
	w.hostLabel.SetText("主机")
	w.lineEdit2.SetText("21")
	w.portLabel.SetText("端口")
	w.userLabel.SetText("用户名")
	w.lineEdit3.SetPlaceholderText("example@example.com")
	w.passLabel.SetText("密码")
	w.lineEdit4.SetPlaceholderText("••••••••••••")
	w.checkBox.SetText("记住密码")
	w.pushButton.SetText("登录")
	w.linkButton.SetText("找回密码")
}

// scaleBackground re-scales the background image to the label size (the Python
// resizeEvent implementation).
func (w *LoginWindow) scaleBackground() {
	if w.backgroundLabel.Width() <= 0 || w.backgroundLabel.Height() <= 0 {
		return
	}
	background := asset.QPixmap(loginFS, "resource/images/background.jpg")
	defer background.Delete()
	scaled := background.Scaled3(w.backgroundLabel.Width(), w.backgroundLabel.Height(), qt.KeepAspectRatioByExpanding, qt.SmoothTransformation)
	w.backgroundLabel.SetPixmap(scaled)
}

func main() {
	demo.Run(func() *qt.QWidget { return newLoginWindow().QWidget })
}
