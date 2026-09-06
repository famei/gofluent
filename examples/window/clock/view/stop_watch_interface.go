package view

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
)

// StopWatchInterface is the "秒表" (stop watch) sub-interface.
type StopWatchInterface struct {
	*qt.QWidget

	timeLabel     *widgets.BodyLabel
	hourLabel     *widgets.TitleLabel
	minuteLabel   *widgets.TitleLabel
	secondLabel   *widgets.TitleLabel
	startButton   *widgets.PillToolButton
	flagButton    *widgets.PillToolButton
	restartButton *widgets.PillToolButton
}

// NewStopWatchInterface builds the stop watch interface widget.
func NewStopWatchInterface(parent *qt.QWidget) *StopWatchInterface {
	w := &StopWatchInterface{QWidget: qt.NewQWidget(parent)}
	w.SetObjectName("StopWatchInterface")

	vLayout2 := qt.NewQVBoxLayout(w.QWidget)
	vLayout := qt.NewQVBoxLayout2()
	vLayout.SetSpacing(0)

	vLayout.AddStretchWithStretch(1)

	lightGray := qt.NewQColor3(96, 96, 96)
	darkGray := qt.NewQColor3(206, 206, 206)

	w.timeLabel = widgets.NewBodyLabelText("", w.QWidget)
	w.timeLabel.SetObjectName("timeLabel")
	w.timeLabel.SetTextColor(lightGray, darkGray)
	w.timeLabel.SetPixelFontSize(100)
	vLayout.AddWidget3(w.timeLabel.QWidget, 0, qt.AlignHCenter)

	hLayout3 := qt.NewQHBoxLayout2()
	hLayout3.AddStretchWithStretch(1)
	w.hourLabel = widgets.NewTitleLabelText("", w.QWidget)
	w.hourLabel.SetObjectName("hourLabel")
	w.hourLabel.SetTextColor(lightGray, darkGray)
	hLayout3.AddWidget(w.hourLabel.QWidget)
	hLayout3.AddSpacing(60)
	w.minuteLabel = widgets.NewTitleLabelText("", w.QWidget)
	w.minuteLabel.SetObjectName("minuteLabel")
	w.minuteLabel.SetTextColor(lightGray, darkGray)
	hLayout3.AddWidget(w.minuteLabel.QWidget)
	hLayout3.AddSpacing(90)
	w.secondLabel = widgets.NewTitleLabelText("", w.QWidget)
	w.secondLabel.SetObjectName("secondLabel")
	w.secondLabel.SetTextColor(lightGray, darkGray)
	hLayout3.AddWidget(w.secondLabel.QWidget)
	hLayout3.AddStretchWithStretch(1)
	vLayout.AddLayout(hLayout3.QLayout)

	lightGray.Delete()
	darkGray.Delete()

	vLayout.AddSpacing(50)

	hLayout2 := qt.NewQHBoxLayout2()
	hLayout2.SetSpacing(24)
	hLayout2.AddStretchWithStretch(1)
	w.startButton = newPillButton(w.QWidget, "startButton", true, true)
	hLayout2.AddWidget(w.startButton.QWidget)
	w.flagButton = newPillButton(w.QWidget, "flagButton", false, false)
	hLayout2.AddWidget(w.flagButton.QWidget)
	w.restartButton = newPillButton(w.QWidget, "restartButton", false, false)
	hLayout2.AddWidget(w.restartButton.QWidget)
	hLayout2.AddStretchWithStretch(1)
	vLayout.AddLayout(hLayout2.QLayout)

	vLayout.AddStretchWithStretch(1)
	vLayout2.AddLayout(vLayout.QLayout)

	w.retranslate()

	w.startButton.SetIcon(common.PowerButton)
	w.flagButton.SetIcon(common.Flag)
	w.restartButton.SetIcon(common.Cancel)
	return w
}

func (w *StopWatchInterface) retranslate() {
	w.timeLabel.SetText("00:00:00")
	w.hourLabel.SetText("小时")
	w.minuteLabel.SetText("分钟")
	w.secondLabel.SetText("秒")
}

// newPillButton builds a 68x68 pill tool button with a 21x21 icon size.
func newPillButton(parent *qt.QWidget, objectName string, checked, enabled bool) *widgets.PillToolButton {
	b := widgets.NewPillToolButton(parent)
	b.SetMinimumSize2(68, 68)
	b.SetIconSize(qt.NewQSize2(21, 21))
	b.SetChecked(checked)
	b.SetEnabled(enabled)
	b.SetObjectName(objectName)
	return b
}
