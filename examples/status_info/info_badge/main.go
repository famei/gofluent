// Command info_badge migrates examples/status_info/info_badge/demo.py.
package main

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	vBoxLayout := qt.NewQVBoxLayout(w)

	// Info badges.
	hBoxLayout1 := qt.NewQHBoxLayout2()
	hBoxLayout1.SetSpacing(20)
	hBoxLayout1.SetSizeConstraint(qt.QLayout__SetMinimumSize)

	hBoxLayout1.AddStretchWithStretch(1)
	hBoxLayout1.AddWidget(widgets.InfoBadgeMake(1, nil, widgets.InfoLevelInformation, nil, 0).QWidget)
	hBoxLayout1.AddWidget(widgets.InfoBadgeMake(10, nil, widgets.InfoLevelSuccess, nil, 0).QWidget)
	hBoxLayout1.AddWidget(widgets.InfoBadgeMake(100, nil, widgets.InfoLevelAttention, nil, 0).QWidget)
	hBoxLayout1.AddWidget(widgets.InfoBadgeMake(1000, nil, widgets.InfoLevelWarning, nil, 0).QWidget)
	hBoxLayout1.AddWidget(widgets.InfoBadgeMake(10000, nil, widgets.InfoLevelError, nil, 0).QWidget)
	customBadge := widgets.InfoBadgeMake("1w+", nil, widgets.InfoLevelInformation, nil, 0)
	customBadge.SetCustomBackgroundColor(qt.NewQColor6("#005fb8"), qt.NewQColor6("#60cdff"))
	hBoxLayout1.AddWidget(customBadge.QWidget)
	hBoxLayout1.AddStretchWithStretch(1)
	vBoxLayout.AddLayout(hBoxLayout1.QLayout)

	// Dot info badges.
	hBoxLayout2 := qt.NewQHBoxLayout2()
	hBoxLayout2.SetSpacing(20)
	hBoxLayout2.SetSizeConstraint(qt.QLayout__SetMinimumSize)

	hBoxLayout2.AddStretchWithStretch(1)
	hBoxLayout2.AddWidget(widgets.DotInfoBadgeMake(nil, widgets.InfoLevelInformation, nil, 0).QWidget)
	hBoxLayout2.AddWidget(widgets.DotInfoBadgeMake(nil, widgets.InfoLevelSuccess, nil, 0).QWidget)
	hBoxLayout2.AddWidget(widgets.DotInfoBadgeMake(nil, widgets.InfoLevelAttention, nil, 0).QWidget)
	hBoxLayout2.AddWidget(widgets.DotInfoBadgeMake(nil, widgets.InfoLevelWarning, nil, 0).QWidget)
	hBoxLayout2.AddWidget(widgets.DotInfoBadgeMake(nil, widgets.InfoLevelError, nil, 0).QWidget)
	customDot := widgets.DotInfoBadgeMake(nil, widgets.InfoLevelInformation, nil, 0)
	customDot.SetCustomBackgroundColor(qt.NewQColor6("#005fb8"), qt.NewQColor6("#60cdff"))
	hBoxLayout2.AddWidget(customDot.QWidget)
	hBoxLayout2.AddStretchWithStretch(1)
	vBoxLayout.AddLayout(hBoxLayout2.QLayout)

	// Icon info badges.
	hBoxLayout3 := qt.NewQHBoxLayout2()
	hBoxLayout3.SetSpacing(20)
	hBoxLayout3.SetSizeConstraint(qt.QLayout__SetMinimumSize)

	hBoxLayout3.AddStretchWithStretch(1)
	hBoxLayout3.AddWidget(widgets.IconInfoBadgeMake(common.AcceptMedium, nil, widgets.InfoLevelInformation, nil, 0).QWidget)
	hBoxLayout3.AddWidget(widgets.IconInfoBadgeMake(common.AcceptMedium, nil, widgets.InfoLevelSuccess, nil, 0).QWidget)
	hBoxLayout3.AddWidget(widgets.IconInfoBadgeMake(common.AcceptMedium, nil, widgets.InfoLevelAttention, nil, 0).QWidget)
	hBoxLayout3.AddWidget(widgets.IconInfoBadgeMake(common.CancelMedium, nil, widgets.InfoLevelWarning, nil, 0).QWidget)
	hBoxLayout3.AddWidget(widgets.IconInfoBadgeMake(common.CancelMedium, nil, widgets.InfoLevelError, nil, 0).QWidget)

	iconBadge := widgets.IconInfoBadgeMake(common.Ringer, nil, widgets.InfoLevelInformation, nil, 0)
	iconBadge.SetCustomBackgroundColor(qt.NewQColor6("#005fb8"), qt.NewQColor6("#60cdff"))
	iconBadge.SetFixedSize2(32, 32)
	iconBadge.SetIconSize(qt.NewQSize2(16, 16))
	hBoxLayout3.AddWidget(iconBadge.QWidget)

	hBoxLayout3.AddStretchWithStretch(1)
	vBoxLayout.AddLayout(hBoxLayout3.QLayout)

	// Using an InfoBadge anchored to another control.
	button := widgets.NewToolButtonIcon(common.LeafTwo, w)
	vBoxLayout.AddWidget3(button.QWidget, 0, qt.AlignHCenter)
	widgets.InfoBadgeMake(1, w, widgets.InfoLevelSuccess, button.QWidget, widgets.InfoBadgePositionTopRight)

	w.Resize(450, 400)
	return w
}

func main() {
	demo.Run(newDemo)
}
