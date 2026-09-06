package components

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/layout"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// SampleCard is a clickable card that switches to a sample (port of
// app/components/sample_card.py).
type SampleCard struct {
	*widgets.CardWidget
	index        int
	routeKey     string
	iconWidget   *widgets.IconWidget
	titleLabel   *qt.QLabel
	contentLabel *qt.QLabel
	hBoxLayout   *qt.QHBoxLayout
	vBoxLayout   *qt.QVBoxLayout
}

// NewSampleCard builds a sample card.
func NewSampleCard(icon interface{}, title, content, routeKey string, index int, parent *qt.QWidget) *SampleCard {
	card := &SampleCard{CardWidget: widgets.NewCardWidget(parent), index: index, routeKey: routeKey}
	card.iconWidget = widgets.NewIconWidgetIcon(icon, card.QWidget)
	card.titleLabel = qt.NewQLabel5(title, card.QWidget)
	wrapped, _ := gcommon.Wrap(content, 45, false)
	card.contentLabel = qt.NewQLabel5(wrapped, card.QWidget)

	card.hBoxLayout = qt.NewQHBoxLayout(card.QWidget)
	card.vBoxLayout = qt.NewQVBoxLayout2()

	card.SetFixedSize2(360, 90)
	card.iconWidget.SetFixedSize2(48, 48)

	card.hBoxLayout.SetSpacing(28)
	card.hBoxLayout.SetContentsMargins(20, 0, 0, 0)
	card.vBoxLayout.SetSpacing(2)
	card.vBoxLayout.SetContentsMargins(0, 0, 0, 0)

	card.hBoxLayout.AddWidget(card.iconWidget.QWidget)
	card.hBoxLayout.AddLayout(card.vBoxLayout.QLayout)
	card.vBoxLayout.AddStretchWithStretch(1)
	card.vBoxLayout.AddWidget(card.titleLabel.QWidget)
	card.vBoxLayout.AddWidget(card.contentLabel.QWidget)
	card.vBoxLayout.AddStretchWithStretch(1)

	card.titleLabel.SetObjectName("titleLabel")
	card.contentLabel.SetObjectName("contentLabel")

	card.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		gallerycommon.SignalBusInstance.EmitSwitchToSample(card.routeKey, card.index)
	})
	return card
}

// SampleCardView is a titled section of sample cards arranged in a flow layout.
type SampleCardView struct {
	*qt.QWidget
	titleLabel *qt.QLabel
	vBoxLayout *qt.QVBoxLayout
	flowLayout *layout.FlowLayout
}

// NewSampleCardView builds a sample card view.
func NewSampleCardView(title string, parent *qt.QWidget) *SampleCardView {
	v := &SampleCardView{QWidget: qt.NewQWidget(parent)}
	v.titleLabel = qt.NewQLabel5(title, v.QWidget)
	v.vBoxLayout = qt.NewQVBoxLayout(v.QWidget)
	v.flowLayout = layout.NewFlowLayout(nil, false, false)

	v.vBoxLayout.SetContentsMargins(36, 0, 36, 0)
	v.vBoxLayout.SetSpacing(10)
	v.flowLayout.SetContentsMargins(0, 0, 0, 0)
	v.flowLayout.SetHorizontalSpacing(12)
	v.flowLayout.SetVerticalSpacing(12)

	v.vBoxLayout.AddWidget(v.titleLabel.QWidget)
	v.vBoxLayout.AddLayout2(v.flowLayout.QLayout, 1)

	v.titleLabel.SetObjectName("viewTitleLabel")
	gallerycommon.StyleSampleCard.Apply(v.QWidget, gcommon.ThemeAuto)
	return v
}

// AddSampleCard adds a sample card to the view.
func (v *SampleCardView) AddSampleCard(icon interface{}, title, content, routeKey string, index int) {
	card := NewSampleCard(icon, title, content, routeKey, index, v.QWidget)
	v.flowLayout.AddWidget(card.QWidget)
}
