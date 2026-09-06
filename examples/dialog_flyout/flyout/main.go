// Command flyout migrates examples/dialog_flyout/flyout/demo.py: three flyout
// variants — a default flyout with an info-bar icon, a FlyoutView with an
// embedded image, and a FlyoutViewBase subclass.
package main

import (
	_ "embed"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

//go:embed resource/SBR.jpg
var sbrJPG []byte

// sbrImage holds the decoded SBR.jpg for the whole demo lifetime. FlyoutView
// stores the QImage pointer without cloning, so the image must stay reachable
// while the demo runs (the process exits via demo.Run, so it is never freed).
var sbrImage *qt.QImage

func sbrQImage() *qt.QImage {
	if sbrImage == nil {
		img := qt.QImage_FromDataWithData(sbrJPG)
		if img == nil || img.IsNull() {
			panic("flyout: decode resource/SBR.jpg")
		}
		sbrImage = img
	}
	return sbrImage
}

// customFlyoutView is the Go equivalent of the Python CustomFlyoutView
// (FlyoutViewBase) subclass: a body label plus a primary button.
type customFlyoutView struct {
	*widgets.FlyoutViewBase
	vBoxLayout *qt.QVBoxLayout
	label      *widgets.BodyLabel
	button     *widgets.PrimaryPushButton
}

func newCustomFlyoutView(parent *qt.QWidget) *customFlyoutView {
	w := &customFlyoutView{FlyoutViewBase: widgets.NewFlyoutViewBase(parent)}
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.label = widgets.NewBodyLabelText("这是一场「试炼」，我认为这就是一场为了战胜过去的「试炼」，\n只有战胜了那些幼稚的过去，人才能有所成长。", w.QWidget)
	w.button = widgets.NewPrimaryPushButtonText("Action", w.QWidget)
	w.button.SetFixedWidth(140)
	w.vBoxLayout.SetSpacing(12)
	w.vBoxLayout.SetContentsMargins(20, 16, 20, 16)
	w.vBoxLayout.AddWidget(w.label.QWidget)
	w.vBoxLayout.AddWidget(w.button.QWidget)
	return w
}

type demoWindow struct {
	*qt.QWidget
	button1 *widgets.PushButton
	button2 *widgets.PushButton
	button3 *widgets.PushButton
}

func newDemo() *demoWindow {
	w := &demoWindow{QWidget: qt.NewQWidget2()}
	w.Resize(750, 550)
	w.button1 = widgets.NewPushButtonText("Click Me", w.QWidget)
	w.button2 = widgets.NewPushButtonText("Click Me", w.QWidget)
	w.button3 = widgets.NewPushButtonText("Click Me", w.QWidget)
	w.button1.SetFixedWidth(150)
	w.button2.SetFixedWidth(150)
	w.button3.SetFixedWidth(150)

	hBoxLayout := qt.NewQHBoxLayout(w.QWidget)
	hBoxLayout.AddWidget3(w.button1.QWidget, 0, qt.AlignBottom)
	hBoxLayout.AddWidget3(w.button2.QWidget, 0, qt.AlignBottom)
	hBoxLayout.AddWidget3(w.button3.QWidget, 0, qt.AlignBottom)
	hBoxLayout.SetContentsMargins(30, 50, 30, 50)

	w.button1.OnClicked(w.showFlyout1)
	w.button2.OnClicked(w.showFlyout2)
	w.button3.OnClicked(w.showFlyout3)
	return w
}

func (w *demoWindow) showFlyout1() {
	widgets.FlyoutCreate(
		"Lesson 4",
		"表达敬意吧，表达出敬意，然后迈向回旋的另一个全新阶段！",
		widgets.InfoBarIconSuccess,
		nil,
		true,
		w.button1.QWidget,
		w.QWidget,
		widgets.FlyoutAnimationPullUp,
		false,
	)
}

func (w *demoWindow) showFlyout2() {
	view := widgets.NewFlyoutView(
		"杰洛·齐贝林",
		"触网而起的网球会落到哪一侧，谁也无法知晓。\n如果那种时刻到来，我希望「女神」是存在的。\n这样的话，不管网球落到哪一边，我都会坦然接受的吧。",
		nil,
		sbrQImage(),
		true,
		nil,
	)

	button := widgets.NewPushButtonText("Action", nil)
	button.SetFixedWidth(120)
	view.AddWidget(button.QWidget, 0, qt.AlignRight)

	// NOTE: the Python demo additionally calls view.widgetLayout.insertSpacing(1, 5)
	// and addSpacing(5); gofluent's FlyoutView does not expose widgetLayout, so
	// those optional spacing tweaks are omitted.

	flyout := widgets.FlyoutMake(view.FlyoutViewBase, w.button2.QWidget, w.QWidget, widgets.FlyoutAnimationPullUp, false)
	view.SetOnClosed(func() { flyout.Close() })
}

func (w *demoWindow) showFlyout3() {
	widgets.FlyoutMake(newCustomFlyoutView(nil).FlyoutViewBase, w.button3.QWidget, w.QWidget, widgets.FlyoutAnimationDropDown, false)
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
