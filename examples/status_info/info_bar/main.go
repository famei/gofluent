// Command info_bar migrates examples/status_info/info_bar/demo.py.
package main

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

type infoBarDemo struct {
	*qt.QWidget
}

func newDemo() *infoBarDemo {
	w := &infoBarDemo{QWidget: qt.NewQWidget2()}

	hBoxLayout := qt.NewQHBoxLayout(w.QWidget)
	button1 := widgets.NewPushButtonText("Information", w.QWidget)
	button2 := widgets.NewPushButtonText("Success", w.QWidget)
	button3 := widgets.NewPushButtonText("Warning", w.QWidget)
	button4 := widgets.NewPushButtonText("Error", w.QWidget)
	button5 := widgets.NewPushButtonText("Custom", w.QWidget)
	button6 := widgets.NewPushButtonText("Desktop", w.QWidget)

	button1.OnClicked(w.createInfoInfoBar)
	button2.OnClicked(w.createSuccessInfoBar)
	button3.OnClicked(w.createWarningInfoBar)
	button4.OnClicked(w.createErrorInfoBar)
	button5.OnClicked(w.createCustomInfoBar)
	button6.OnClicked(w.createDeskTopBottomRightInfoBar)

	hBoxLayout.AddWidget(button1.QWidget)
	hBoxLayout.AddWidget(button2.QWidget)
	hBoxLayout.AddWidget(button3.QWidget)
	hBoxLayout.AddWidget(button4.QWidget)
	hBoxLayout.AddWidget(button5.QWidget)
	hBoxLayout.AddWidget(button6.QWidget)
	hBoxLayout.SetContentsMargins(30, 0, 30, 0)

	w.Resize(700, 700)
	return w
}

func (d *infoBarDemo) createInfoInfoBar() {
	content := "My name is kira yoshikake, 33 years old. Living in the villa area northeast of duwangting, unmarried. I work in Guiyou chain store. Every day I have to work overtime until 8 p.m. to go home. I don't smoke. The wine is only for a taste. Sleep at 11 p.m. for 8 hours a day. Before I go to bed, I must drink a cup of warm milk, then do 20 minutes of soft exercise, get on the bed, and immediately fall asleep. Never leave fatigue and stress until the next day. Doctors say I'm normal."
	bar := widgets.NewInfoBar(widgets.InfoBarIconInformation, "Title", content, qt.Vertical, true, 2000, widgets.InfoBarPositionTopRight, d.QWidget)
	bar.AddWidget(widgets.NewPushButtonText("Action", nil).QWidget, 0)
	bar.Show()
}

func (d *infoBarDemo) createSuccessInfoBar() {
	bar := widgets.NewInfoBar(widgets.InfoBarIconSuccess, "Lesson 4", "With respect, let's advance towards a new stage of the spin.", qt.Horizontal, true, 2000, widgets.InfoBarPositionTop, d.QWidget)
	bar.Show()
}

func (d *infoBarDemo) createWarningInfoBar() {
	bar := widgets.NewInfoBar(widgets.InfoBarIconWarning, "Lesson 3", "Believe in the spin, just keep believing!", qt.Horizontal, false, 2000, widgets.InfoBarPositionTopLeft, d.QWidget)
	bar.Show()
}

func (d *infoBarDemo) createErrorInfoBar() {
	bar := widgets.NewInfoBar(widgets.InfoBarIconError, "Lesson 5", "迂回路を行けば最短ルート。", qt.Horizontal, true, -1, widgets.InfoBarPositionBottomRight, d.QWidget)
	bar.Show()
}

func (d *infoBarDemo) createCustomInfoBar() {
	bar := widgets.NewInfoBar(common.GitHub, "Zeppeli", "人間讃歌は「勇気」の讃歌ッ！！ 人間のすばらしさは勇気のすばらしさ！！", qt.Horizontal, true, 2000, widgets.InfoBarPositionBottom, d.QWidget)
	bar.SetCustomBackgroundColor(qt.NewQColor6("white"), qt.NewQColor6("#202020"))
	bar.Show()
}

func (d *infoBarDemo) createDeskTopBottomRightInfoBar() {
	// InfoBar.desktopView() has no Go equivalent; a nil parent makes the bar a
	// top-level widget positioned against the screen (see MIGRATION_GUIDE §10).
	bar := widgets.NewInfoBar(widgets.InfoBarIconWarning, "Plugged Out Notify", "Battery is 64%", qt.Vertical, true, 1000, widgets.InfoBarPositionBottomRight, nil)
	bar.Show()
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
