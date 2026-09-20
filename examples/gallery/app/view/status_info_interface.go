package view

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// StatusInfoInterface is the "Status & info" gallery page (port of
// app/view/status_info_interface.py).
type StatusInfoInterface struct {
	*GalleryInterface
	stateTooltip *widgets.StateToolTip
	stateButton  *widgets.PushButton
}

// NewStatusInfoInterface builds the status info interface.
func NewStatusInfoInterface(parent *qt.QWidget) *StatusInfoInterface {
	t := gallerycommon.NewTranslator()
	i := &StatusInfoInterface{GalleryInterface: NewGalleryInterface(t.StatusInfo, "github.com/famei/gofluent/components/widgets", parent)}
	i.SetObjectName("statusInfoInterface")

	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }

	i.stateButton = widgets.NewPushButtonText(tr("Show StateToolTip"), nil)
	i.stateButton.OnClicked(i.onStateButtonClicked)
	i.AddExampleCard(tr("State tool tip"), i.stateButton.QWidget,
		"components/widgets/state_tool_tip.go", codeStateToolTip, 0)

	button := widgets.NewPushButtonText(tr("Button with a simple ToolTip"), nil)
	widgets.NewToolTipFilter(button.QWidget, 300, widgets.ToolTipPositionTop)
	button.SetToolTip(tr("Simple ToolTip"))
	i.AddExampleCard(tr("A button with a simple ToolTip"), button.QWidget,
		"components/widgets/tool_tip.go", codeToolTipButton, 0)

	label := widgets.NewPixmapLabel(nil)
	pm := resource.Pixmap("kunkun.png")
	scaled := pm.Scaled3(160, 160, qt.KeepAspectRatio, qt.SmoothTransformation)
	pm.Delete()
	label.SetPixmap(scaled)
	widgets.NewToolTipFilter(label.QWidget, 500, widgets.ToolTipPositionTop)
	label.SetToolTip(tr("Label with a ToolTip"))
	label.SetToolTipDuration(2000)
	label.SetFixedSize2(160, 160)
	i.AddExampleCard(tr("A label with a ToolTip"), label.QWidget,
		"components/widgets/tool_tip.go", codeToolTipLabel, 0)

	badgeWidget := qt.NewQWidget2()
	badgeLayout := qt.NewQHBoxLayout(badgeWidget)
	badgeLayout.AddWidget(widgets.NewInfoBadgeText("1", nil, widgets.InfoLevelInformation).QWidget)
	badgeLayout.AddWidget(widgets.NewInfoBadgeText("10", nil, widgets.InfoLevelSuccess).QWidget)
	badgeLayout.AddWidget(widgets.NewInfoBadgeText("100", nil, widgets.InfoLevelAttention).QWidget)
	badgeLayout.AddWidget(widgets.NewInfoBadgeText("1000", nil, widgets.InfoLevelWarning).QWidget)
	badgeLayout.AddWidget(widgets.NewInfoBadgeText("10000", nil, widgets.InfoLevelError).QWidget)
	customBadge := widgets.NewInfoBadgeText("1w+", nil, widgets.InfoLevelInformation)
	customBadge.SetCustomBackgroundColor(qt.NewQColor6("#005fb8"), qt.NewQColor6("#60cdff"))
	badgeLayout.AddWidget(customBadge.QWidget)
	badgeLayout.SetSpacing(20)
	badgeLayout.SetContentsMargins(0, 10, 0, 10)
	i.AddExampleCard(tr("InfoBadge in different styles"), badgeWidget,
		"components/widgets/info_badge.go", codeInfoBadge, 0)

	infoBar := widgets.NewInfoBar(widgets.InfoBarIconSuccess, tr("Success"),
		tr("The Anthem of man is the Anthem of courage."),
		qt.Horizontal, true, -1, widgets.InfoBarPositionNone, i.QWidget)
	i.AddExampleCard(tr("A closable InfoBar"), infoBar.QWidget,
		"components/widgets/info_bar.go", codeInfoBar, 0)

	content := tr("My name is kira yoshikake, 33 years old. Living in the villa area northeast of duwangting, unmarried. I work in Guiyou chain store. Every day I have to work overtime until 8 p.m. to go home. I don't smoke. The wine is only for a taste. Sleep at 11 p.m. for 8 hours a day. Before I go to bed, I must drink a cup of warm milk, then do 20 minutes of soft exercise, get on the bed, and immediately fall asleep. Never leave fatigue and stress until the next day. Doctors say I'm normal.")
	infoBar = widgets.NewInfoBar(widgets.InfoBarIconWarning, tr("Warning"), content,
		qt.Vertical, true, -1, widgets.InfoBarPositionNone, i.QWidget)
	i.AddExampleCard(tr("A closable InfoBar with long message"), infoBar.QWidget,
		"components/widgets/info_bar.go", codeInfoBarLongMessage, 0)

	infoBar = widgets.NewInfoBar(gcommon.Code, tr("GitHub"),
		tr("When you look long into an abyss, the abyss looks into you."),
		qt.Horizontal, true, -1, widgets.InfoBarPositionNone, i.QWidget)
	infoBar.AddWidget(widgets.NewPushButtonText(tr("Action"), nil).QWidget, 0)
	infoBar.SetCustomBackgroundColor(qt.NewQColor6("white"), qt.NewQColor6("#2a2a2a"))
	i.AddExampleCard(tr("An InfoBar with custom icon, background color and widget."), infoBar.QWidget,
		"components/widgets/info_bar.go", codeInfoBarCustom, 0)

	w := qt.NewQWidget(i.QWidget)
	hBoxLayout := qt.NewQHBoxLayout(w)
	button1 := widgets.NewPushButtonText(tr("Top right"), w)
	button2 := widgets.NewPushButtonText(tr("Top"), w)
	button3 := widgets.NewPushButtonText(tr("Top left"), w)
	button4 := widgets.NewPushButtonText(tr("Bottom right"), w)
	button5 := widgets.NewPushButtonText(tr("Bottom"), w)
	button6 := widgets.NewPushButtonText(tr("Bottom left"), w)
	button1.OnClicked(i.createTopRightInfoBar)
	button2.OnClicked(i.createTopInfoBar)
	button3.OnClicked(i.createTopLeftInfoBar)
	button4.OnClicked(i.createBottomRightInfoBar)
	button5.OnClicked(i.createBottomInfoBar)
	button6.OnClicked(i.createBottomLeftInfoBar)
	hBoxLayout.AddWidget(button1.QWidget)
	hBoxLayout.AddWidget(button2.QWidget)
	hBoxLayout.AddWidget(button3.QWidget)
	hBoxLayout.AddWidget(button4.QWidget)
	hBoxLayout.AddWidget(button5.QWidget)
	hBoxLayout.AddWidget(button6.QWidget)
	hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	hBoxLayout.SetSpacing(15)
	i.AddExampleCard(tr("InfoBar with different pop-up locations"), w,
		"components/widgets/info_bar.go", codeInfoBarPosition, 0)

	bar := widgets.NewIndeterminateProgressBar(i.QWidget, true)
	bar.SetFixedWidth(200)
	card := i.AddExampleCard(tr("An indeterminate progress bar"), bar.QWidget,
		"components/widgets/progress_bar.go", codeIndeterminateProgressBar, 0)
	card.TopLayout.SetContentsMargins(12, 24, 12, 24)
	card.AdjustSize()

	determinateBar := widgets.NewProgressBar(i.QWidget, true)
	determinateBar.SetFixedWidth(200)
	i.AddExampleCard(tr("An determinate progress bar"), NewProgressWidget(determinateBar.QProgressBar, i.QWidget).QWidget,
		"components/widgets/progress_bar.go", codeProgressBar, 0)

	indeterminateRing := widgets.NewIndeterminateProgressRing(i.QWidget, true)
	indeterminateRing.SetFixedSize2(70, 70)
	i.AddExampleCard(tr("An indeterminate progress ring"), indeterminateRing.QWidget,
		"components/widgets/progress_ring.go", codeIndeterminateProgressRing, 0)

	determinateRing := widgets.NewProgressRing(i.QWidget, true)
	determinateRing.SetFixedSize2(80, 80)
	determinateRing.SetTextVisible(true)
	i.AddExampleCard(tr("An determinate progress ring"), NewProgressWidget(determinateRing.QProgressBar, i.QWidget).QWidget,
		"components/widgets/progress_ring.go", codeProgressRing, 0)
	return i
}

func (i *StatusInfoInterface) onStateButtonClicked() {
	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }
	if i.stateTooltip != nil {
		i.stateTooltip.SetContent(tr("The model training is complete!") + " 😆")
		i.stateButton.SetText(tr("Show StateToolTip"))
		i.stateTooltip.SetState(true)
		i.stateTooltip = nil
	} else {
		i.stateTooltip = widgets.NewStateToolTip(tr("Training model"), tr("Please wait patiently"), i.Window())
		i.stateButton.SetText(tr("Hide StateToolTip"))
		pos := i.stateTooltip.GetSuitablePos()
		i.stateTooltip.Move(pos.X(), pos.Y())
		pos.Delete()
		i.stateTooltip.Show()
	}
}

func (i *StatusInfoInterface) createTopRightInfoBar() {
	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }
	b := widgets.NewInfoBar(widgets.InfoBarIconInformation, tr("Lesson 3"),
		tr("Believe in the spin, just keep believing!"), qt.Horizontal, true, 2000, widgets.InfoBarPositionTopRight, i.QWidget)
	b.Show()
}

func (i *StatusInfoInterface) createTopInfoBar() {
	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }
	b := widgets.NewInfoBar(widgets.InfoBarIconSuccess, tr("Lesson 4"),
		tr("With respect, let's advance towards a new stage of the spin."), qt.Horizontal, true, 2000, widgets.InfoBarPositionTop, i.QWidget)
	b.Show()
}

func (i *StatusInfoInterface) createTopLeftInfoBar() {
	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }
	b := widgets.NewInfoBar(widgets.InfoBarIconWarning, tr("Lesson 5"),
		tr("The shortest shortcut is to take a detour."), qt.Horizontal, false, 2000, widgets.InfoBarPositionTopLeft, i.QWidget)
	b.Show()
}

func (i *StatusInfoInterface) createBottomRightInfoBar() {
	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }
	b := widgets.NewInfoBar(widgets.InfoBarIconError, tr("No Internet"),
		tr("An error message which won't disappear automatically."), qt.Horizontal, true, -1, widgets.InfoBarPositionBottomRight, i.QWidget)
	b.Show()
}

func (i *StatusInfoInterface) createBottomInfoBar() {
	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }
	b := widgets.NewInfoBar(widgets.InfoBarIconSuccess, tr("Lesson 1"),
		tr("Don't have any strange expectations of me."), qt.Horizontal, true, 2000, widgets.InfoBarPositionBottom, i.QWidget)
	b.Show()
}

func (i *StatusInfoInterface) createBottomLeftInfoBar() {
	tr := func(s string) string { return gallerycommon.Tr("StatusInfoInterface", s) }
	b := widgets.NewInfoBar(widgets.InfoBarIconWarning, tr("Lesson 2"),
		tr("Don't let your muscles notice."), qt.Horizontal, true, 1500, widgets.InfoBarPositionBottomLeft, i.QWidget)
	b.Show()
}

// ProgressWidget pairs a progress indicator with a spin box (port of
// app/view/status_info_interface.py ProgressWidget).
type ProgressWidget struct {
	*qt.QWidget
	spinBox *widgets.SpinBox
}

// NewProgressWidget builds a progress widget.
func NewProgressWidget(widget *qt.QProgressBar, parent *qt.QWidget) *ProgressWidget {
	w := &ProgressWidget{QWidget: qt.NewQWidget(parent)}
	hBoxLayout := qt.NewQHBoxLayout(w.QWidget)

	w.spinBox = widgets.NewSpinBox(w.QWidget)
	w.spinBox.OnValueChanged(func(value int) { widget.SetValue(value) })
	w.spinBox.SetRange(0, 100)

	hBoxLayout.AddWidget(widget.QWidget)
	hBoxLayout.AddSpacing(50)
	progressLabel := qt.NewQLabel5(gallerycommon.Tr("ProgressWidget", "Progress"), w.QWidget)
	hBoxLayout.AddWidget(progressLabel.QWidget)
	hBoxLayout.AddSpacing(5)
	hBoxLayout.AddWidget(w.spinBox.QWidget)
	hBoxLayout.SetContentsMargins(0, 0, 0, 0)

	w.spinBox.SetValue(0)
	return w
}
