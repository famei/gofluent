package settings

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/layout"
	qt "github.com/mappu/miqt/qt"
)

// SettingCardGroup is a titled container that stacks setting cards vertically.
//
// The card area uses the custom ExpandLayout (components/layout), exactly like
// the Python source, so that the group grows/shrinks when a card inside it
// expands or collapses.
type SettingCardGroup struct {
	*qt.QWidget
	titleLabel *qt.QLabel
	vBoxLayout *qt.QVBoxLayout
	cardLayout *layout.ExpandLayout
}

// NewSettingCardGroup builds a setting card group.
func NewSettingCardGroup(title string, parent *qt.QWidget) *SettingCardGroup {
	g := &SettingCardGroup{QWidget: qt.NewQWidget(parent)}
	g.SetObjectName("settingCardGroup")
	g.titleLabel = qt.NewQLabel5(title, g.QWidget)
	g.vBoxLayout = qt.NewQVBoxLayout(g.QWidget)
	g.cardLayout = layout.NewExpandLayout(nil)

	g.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	g.vBoxLayout.SetSpacing(0)
	g.cardLayout.SetContentsMargins(0, 0, 0, 0)
	g.cardLayout.SetSpacing(2)

	g.vBoxLayout.AddWidget(g.titleLabel.QWidget)
	g.vBoxLayout.AddSpacing(12)
	g.vBoxLayout.AddLayout2(g.cardLayout.QLayout, 1)

	common.FluentStyleSheet(common.FluentSettingCardGroup).Apply(g.QWidget, common.ThemeAuto)
	common.SetFont(g.titleLabel.QWidget, 20, int(qt.QFont__Normal))
	g.titleLabel.AdjustSize()
	return g
}

// AddSettingCard appends a setting card to the group.
func (g *SettingCardGroup) AddSettingCard(card *qt.QWidget) {
	card.SetParent(g.QWidget)
	g.cardLayout.AddWidget(card)
	g.AdjustSize()
}

// AddSettingCards appends multiple setting cards.
func (g *SettingCardGroup) AddSettingCards(cards []*qt.QWidget) {
	for _, card := range cards {
		g.AddSettingCard(card)
	}
}

// AdjustSize resizes the group to fit its cards. The +46 accounts for the
// title label and the 12px spacing below it, matching the Python port.
func (g *SettingCardGroup) AdjustSize() {
	// TotalHeightForWidth calls the ExpandLayout heightForWidth virtual (the
	// sum of the visible card heights), unlike SizeHint which returns the
	// maximum child minimum size.
	h := g.cardLayout.TotalHeightForWidth(g.Width()) + 46
	g.Resize(g.Width(), h)
}

// TitleLabel exposes the group title label.
func (g *SettingCardGroup) TitleLabel() *qt.QLabel { return g.titleLabel }

// CardLayout exposes the card layout.
func (g *SettingCardGroup) CardLayout() *layout.ExpandLayout { return g.cardLayout }
