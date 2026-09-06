package view

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/layout"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// allFluentIcons is the full Fluent icon library shown by the icon gallery
// (the Python demo iterates FluentIcon._member_map_.values()).
var allFluentIcons = gcommon.AllFluentIcons()

// iconCard is a clickable icon preview (port of app/view/icon_interface.py
// IconCard).
type iconCard struct {
	*qt.QFrame
	icon       gcommon.FluentIcon
	isSelected bool
	iconWidget *widgets.IconWidget
	nameLabel  *qt.QLabel
	vBoxLayout *qt.QVBoxLayout
	OnClicked  func(icon gcommon.FluentIcon)
}

// newIconCard builds an icon card.
func newIconCard(icon gcommon.FluentIcon, parent *qt.QWidget) *iconCard {
	card := &iconCard{QFrame: qt.NewQFrame(parent), icon: icon}
	card.SetObjectName("iconCard")
	card.iconWidget = widgets.NewIconWidgetIcon(icon, card.QWidget)
	card.nameLabel = qt.NewQLabel(card.QWidget)
	card.vBoxLayout = qt.NewQVBoxLayout(card.QWidget)

	card.SetFixedSize2(96, 96)
	card.vBoxLayout.SetSpacing(0)
	card.vBoxLayout.SetContentsMargins(8, 28, 8, 0)
	card.iconWidget.SetFixedSize2(28, 28)
	card.vBoxLayout.AddWidget3(card.iconWidget.QWidget, 0, qt.AlignHCenter)
	card.vBoxLayout.AddSpacing(14)
	card.vBoxLayout.AddWidget3(card.nameLabel.QWidget, 0, qt.AlignHCenter)

	text := card.nameLabel.FontMetrics().ElidedText(string(icon), qt.ElideRight, 90)
	card.nameLabel.SetText(text)

	card.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		if card.isSelected {
			return
		}
		if card.OnClicked != nil {
			card.OnClicked(card.icon)
		}
	})
	return card
}

func (c *iconCard) setSelected(isSelected bool, force bool) {
	if isSelected == c.isSelected && !force {
		return
	}
	c.isSelected = isSelected

	if !isSelected {
		c.iconWidget.SetIcon(c.icon)
	} else {
		if gcommon.IsDarkTheme() {
			c.iconWidget.SetIcon(c.icon.Icon(gcommon.ThemeLight))
		} else {
			c.iconWidget.SetIcon(c.icon.Icon(gcommon.ThemeDark))
		}
	}

	c.SetProperty("isSelected", qt.NewQVariant11(isSelected))
	c.SetStyle(qt.QApplication_Style())
}

// iconInfoPanel shows the details of the selected icon.
type iconInfoPanel struct {
	*qt.QFrame
	nameLabel          *qt.QLabel
	iconWidget         *widgets.IconWidget
	iconNameTitleLabel *qt.QLabel
	iconNameLabel      *qt.QLabel
	enumNameTitleLabel *qt.QLabel
	enumNameLabel      *qt.QLabel
	vBoxLayout         *qt.QVBoxLayout
}

// newIconInfoPanel builds an icon info panel.
func newIconInfoPanel(icon gcommon.FluentIcon, parent *qt.QWidget) *iconInfoPanel {
	p := &iconInfoPanel{QFrame: qt.NewQFrame(parent)}
	p.SetObjectName("iconInfoPanel")
	p.nameLabel = qt.NewQLabel5(string(icon), p.QWidget)
	p.iconWidget = widgets.NewIconWidgetIcon(icon, p.QWidget)
	p.iconNameTitleLabel = qt.NewQLabel5(gallerycommon.Tr("IconInfoPanel", "Icon name"), p.QWidget)
	p.iconNameLabel = qt.NewQLabel5(string(icon), p.QWidget)
	p.enumNameTitleLabel = qt.NewQLabel5(gallerycommon.Tr("IconInfoPanel", "Enum member"), p.QWidget)
	p.enumNameLabel = qt.NewQLabel5("FluentIcon."+string(icon), p.QWidget)

	p.vBoxLayout = qt.NewQVBoxLayout(p.QWidget)
	p.vBoxLayout.SetContentsMargins(16, 20, 16, 20)
	p.vBoxLayout.SetSpacing(0)

	p.vBoxLayout.AddWidget(p.nameLabel.QWidget)
	p.vBoxLayout.AddSpacing(16)
	p.vBoxLayout.AddWidget(p.iconWidget.QWidget)
	p.vBoxLayout.AddSpacing(45)
	p.vBoxLayout.AddWidget(p.iconNameTitleLabel.QWidget)
	p.vBoxLayout.AddSpacing(5)
	p.vBoxLayout.AddWidget(p.iconNameLabel.QWidget)
	p.vBoxLayout.AddSpacing(34)
	p.vBoxLayout.AddWidget(p.enumNameTitleLabel.QWidget)
	p.vBoxLayout.AddSpacing(5)
	p.vBoxLayout.AddWidget(p.enumNameLabel.QWidget)

	p.iconWidget.SetFixedSize2(48, 48)
	p.SetFixedWidth(216)

	p.nameLabel.SetObjectName("nameLabel")
	p.iconNameTitleLabel.SetObjectName("subTitleLabel")
	p.enumNameTitleLabel.SetObjectName("subTitleLabel")
	return p
}

func (p *iconInfoPanel) setIcon(icon gcommon.FluentIcon) {
	p.iconWidget.SetIcon(icon)
	p.nameLabel.SetText(string(icon))
	p.iconNameLabel.SetText(string(icon))
	p.enumNameLabel.SetText("FluentIcon." + string(icon))
}

// iconCardView hosts the searchable icon grid (port of IconCardView).
type iconCardView struct {
	*qt.QWidget
	trie             *gallerycommon.Trie
	iconLibraryLabel *widgets.StrongBodyLabel
	searchLineEdit   *widgets.SearchLineEdit
	view             *qt.QFrame
	scrollArea       *widgets.SmoothScrollArea
	scrollWidget     *qt.QWidget
	infoPanel        *iconInfoPanel
	vBoxLayout       *qt.QVBoxLayout
	hBoxLayout       *qt.QHBoxLayout
	flowLayout       *layout.FlowLayout

	cards        []*iconCard
	icons        []gcommon.FluentIcon
	currentIndex int
}

// newIconCardView builds the icon card view.
func newIconCardView(parent *qt.QWidget) *iconCardView {
	v := &iconCardView{QWidget: qt.NewQWidget(parent), trie: gallerycommon.NewTrie(), currentIndex: -1}
	v.SetObjectName("iconCardView")
	v.iconLibraryLabel = widgets.NewStrongBodyLabelText(gallerycommon.Tr("IconCardView", "Fluent Icons Library"), v.QWidget)
	v.searchLineEdit = widgets.NewSearchLineEdit(v.QWidget)

	v.view = qt.NewQFrame(v.QWidget)
	// Expanding so the icon grid (and info panel) fill the remaining vertical
	// space instead of collapsing to the info panel's sizeHint height.
	v.view.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)
	v.scrollArea = widgets.NewSmoothScrollArea(v.view.QWidget)
	// The QScrollArea's default vertical policy collapses it to its sizeHint in
	// the hBoxLayout; Expanding makes it fill the view's height (and therefore
	// show the whole icon grid).
	v.scrollArea.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)
	v.scrollWidget = qt.NewQWidget(v.scrollArea.QWidget)
	v.infoPanel = newIconInfoPanel(gcommon.Menu, v.QWidget)

	v.vBoxLayout = qt.NewQVBoxLayout(v.QWidget)
	v.hBoxLayout = qt.NewQHBoxLayout(v.view.QWidget)
	v.flowLayout = layout.NewFlowLayout(v.scrollWidget, false, true)

	v.initWidget()
	return v
}

func (v *iconCardView) initWidget() {
	v.scrollArea.SetWidget(v.scrollWidget)
	v.scrollArea.SetViewportMargins(0, 5, 0, 5)
	v.scrollArea.SetWidgetResizable(true)
	v.scrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)

	v.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	v.vBoxLayout.SetSpacing(12)
	v.vBoxLayout.AddWidget(v.iconLibraryLabel.QWidget)
	v.vBoxLayout.AddWidget(v.searchLineEdit.QWidget)
	v.vBoxLayout.AddWidget(v.view.QWidget)

	v.hBoxLayout.SetSpacing(0)
	v.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	v.hBoxLayout.AddWidget(v.scrollArea.QWidget)
	v.hBoxLayout.AddWidget3(v.infoPanel.QWidget, 0, qt.AlignRight)

	v.flowLayout.SetVerticalSpacing(8)
	v.flowLayout.SetHorizontalSpacing(8)
	v.flowLayout.SetContentsMargins(8, 3, 8, 8)

	v.setQss()
	v.searchLineEdit.SetPlaceholderText(gallerycommon.Tr("LineEdit", "Search icons"))
	v.searchLineEdit.ClearSignal = v.showAllIcons
	v.searchLineEdit.SearchSignal = v.search
	v.searchLineEdit.OnTextChanged(func(text string) { v.search(text) })

	for _, icon := range allFluentIcons {
		v.addIcon(icon)
	}

	v.setSelectedIcon(v.icons[0])
}

func (v *iconCardView) addIcon(icon gcommon.FluentIcon) {
	card := newIconCard(icon, v.QWidget)
	card.OnClicked = v.setSelectedIcon

	v.trie.Insert(string(icon), len(v.cards))
	v.cards = append(v.cards, card)
	v.icons = append(v.icons, icon)
	v.flowLayout.AddWidget(card.QWidget)
}

func (v *iconCardView) setSelectedIcon(icon gcommon.FluentIcon) {
	index := -1
	for i, ic := range v.icons {
		if ic == icon {
			index = i
			break
		}
	}
	if index < 0 {
		return
	}

	if v.currentIndex >= 0 {
		v.cards[v.currentIndex].setSelected(false, false)
	}

	v.currentIndex = index
	v.cards[index].setSelected(true, false)
	v.infoPanel.setIcon(icon)
}

func (v *iconCardView) setQss() {
	v.view.SetObjectName("iconView")
	v.scrollWidget.SetObjectName("scrollWidget")

	gallerycommon.StyleIconInterface.Apply(v.QWidget, gcommon.ThemeAuto)
	gallerycommon.StyleIconInterface.Apply(v.scrollWidget, gcommon.ThemeAuto)

	if v.currentIndex >= 0 {
		v.cards[v.currentIndex].setSelected(true, true)
	}
}

func (v *iconCardView) search(keyWord string) {
	items := v.trie.Items(keyWord)
	indexes := map[int]bool{}
	for _, item := range items {
		if idx, ok := item[1].(int); ok {
			indexes[idx] = true
		}
	}
	v.flowLayout.RemoveAllWidgets()

	for i, card := range v.cards {
		isVisible := indexes[i]
		card.SetVisible(isVisible)
		if isVisible {
			v.flowLayout.AddWidget(card.QWidget)
		}
	}
}

func (v *iconCardView) showAllIcons() {
	v.flowLayout.RemoveAllWidgets()
	for _, card := range v.cards {
		card.Show()
		v.flowLayout.AddWidget(card.QWidget)
	}
}

// IconInterface is the "Icons" gallery page (port of app/view/icon_interface.py).
type IconInterface struct {
	*GalleryInterface
}

// NewIconInterface builds the icon interface.
func NewIconInterface(parent *qt.QWidget) *IconInterface {
	t := gallerycommon.NewTranslator()
	i := &IconInterface{GalleryInterface: NewGalleryInterface(t.Icons, "qfluentwidgets.common.icon", parent)}
	i.SetObjectName("iconInterface")

	// Parent the icon view to the content widget (i.View), matching how
	// AddExampleCard parents cards to g.View. The previous i.QWidget parent kept
	// the icon view as a child of the QScrollArea, so the ExpandLayout's
	// geometry (relative to the View) landed at the ScrollArea origin and the
	// icon grid overlapped the title.
	iconView := newIconCardView(i.View)
	// Activate the icon card view's internal layout before sizing, so its
	// title/search/grid are laid out (not overlapping at y=0).
	if l := iconView.Layout(); l != nil {
		l.Activate()
	}
	iconView.AdjustSize() // ExpandLayout positions by current Height()
	i.VBoxLayout.AddWidget(iconView.QWidget)

	// Fill the whole content area with the icon view (its QFrame view is
	// Expanding), so the icon grid shows all icons instead of a short, clipped
	// strip. Keep it synced when the window resizes.
	fillIconView := func() {
		m := i.VBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
		h := i.View.Height() - m.Top() - m.Bottom()
		if h <= 0 {
			return
		}
		iconView.Resize(i.View.Width(), h)

		// The nested QScrollArea does not fill the QFrame view via its size
		// policy in miqt, so size the inner frame + scroll area explicitly from
		// the remaining height below the title + search row.
		titleH := iconView.iconLibraryLabel.Height() + iconView.searchLineEdit.Height() + iconView.vBoxLayout.Spacing()*2
		innerH := iconView.Height() - titleH
		if innerH > 0 {
			iconView.view.Resize(iconView.Width(), innerH)
			iconView.scrollArea.Resize(iconView.view.Width()-iconView.infoPanel.Width(), innerH)
			// SetWidgetResizable does not re-run updateWidgetPosition after this
			// manual resize in miqt, so size the scroll widget to the viewport
			// explicitly (the FlowLayout then re-flows to the wider width).
			iconView.scrollWidget.Resize(iconView.scrollArea.Width(), iconView.scrollArea.Height())
		}
	}
	fillIconView()
	i.View.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		fillIconView()
	})
	return i
}
