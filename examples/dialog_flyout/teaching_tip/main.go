// Command teaching_tip migrates examples/dialog_flyout/teaching_tip/demo.py:
// teaching tips anchored to buttons with top/bottom/custom tail positions.
package main

import (
	_ "embed"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

//go:embed resource/Gyro.jpg
var gyroJPG []byte

// gyroImage holds the decoded Gyro.jpg for the whole demo lifetime. The
// TeachingTipView stores the QImage pointer without cloning, so the image must
// stay reachable while the demo runs.
var gyroImage *qt.QImage

func gyroQImage() *qt.QImage {
	if gyroImage == nil {
		img := qt.QImage_FromDataWithData(gyroJPG)
		if img == nil || img.IsNull() {
			panic("teaching_tip: decode resource/Gyro.jpg")
		}
		gyroImage = img
	}
	return gyroImage
}

// customFlyoutView is the Go equivalent of the Python CustomFlyoutView
// (FlyoutViewBase) subclass whose paintEvent is overridden to a no-op.
type customFlyoutView struct {
	*widgets.FlyoutViewBase
	vBoxLayout *qt.QVBoxLayout
	label      *widgets.BodyLabel
	button     *widgets.PrimaryPushButton
}

func newCustomFlyoutView(parent *qt.QWidget) *customFlyoutView {
	w := &customFlyoutView{FlyoutViewBase: widgets.NewFlyoutViewBase(parent)}
	w.SetDrawBackground(false) // mirrors the Python paintEvent(self, e): pass override
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
	w.Resize(700, 500)
	w.button1 = widgets.NewPushButtonText("Top", w.QWidget)
	w.button2 = widgets.NewPushButtonText("Bottom", w.QWidget)
	w.button3 = widgets.NewPushButtonText("Custom", w.QWidget)
	w.button1.SetFixedWidth(150)
	w.button2.SetFixedWidth(150)
	w.button3.SetFixedWidth(150)

	hBoxLayout := qt.NewQHBoxLayout(w.QWidget)
	hBoxLayout.AddWidget3(w.button2.QWidget, 0, qt.AlignHCenter)
	hBoxLayout.AddWidget3(w.button1.QWidget, 0, qt.AlignHCenter)
	hBoxLayout.AddWidget3(w.button3.QWidget, 0, qt.AlignHCenter)

	w.button1.OnClicked(w.showTopTip)
	w.button2.OnClicked(w.showBottomTip)
	w.button3.OnClicked(w.showCustomTip)
	return w
}

func (w *demoWindow) showTopTip() {
	position := widgets.TeachingTipTailBottom
	view := widgets.NewTeachingTipView(
		"Lesson 5",
		"最短的捷径就是绕远路，绕远路才是我的最短捷径。",
		nil,
		gyroQImage(),
		true,
		position,
		nil,
	)

	button := widgets.NewPushButtonText("Action", nil)
	button.SetFixedWidth(120)
	view.AddWidget(button.QWidget, 0, qt.AlignRight)

	tip := widgets.TeachingTipMake(view.FlyoutViewBase, w.button1.QWidget, -1, position, w.QWidget, false)
	view.SetOnClosed(func() { tip.Close() })
}

func (w *demoWindow) showBottomTip() {
	widgets.TeachingTipCreate(
		w.button2.QWidget,
		"Lesson 4",
		"表达敬意吧，表达出敬意，然后迈向回旋的另一个全新阶段！",
		widgets.InfoBarIconSuccess,
		nil,
		true,
		2000,
		widgets.TeachingTipTailTop,
		w.QWidget,
		false,
	)
}

func (w *demoWindow) showCustomTip() {
	tip := widgets.NewPopupTeachingTip(newCustomFlyoutView(nil).FlyoutViewBase, w.button3.QWidget, 2000, widgets.TeachingTipTailRight, w.QWidget, false)
	tip.Show()
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
