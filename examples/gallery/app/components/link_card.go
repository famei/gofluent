package components

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// LinkCard is a fixed-size card that opens a URL when clicked (port of
// app/components/link_card.py).
type LinkCard struct {
	*qt.QFrame
	url          *qt.QUrl
	iconWidget   *widgets.IconWidget
	titleLabel   *qt.QLabel
	contentLabel *qt.QLabel
	urlWidget    *widgets.IconWidget
	vBoxLayout   *qt.QVBoxLayout
}

// NewLinkCard builds a link card. icon may be a *qt.QIcon, a string path or a
// gcommon.FluentIconBase.
func NewLinkCard(icon interface{}, title, content, url string, parent *qt.QWidget) *LinkCard {
	card := &LinkCard{QFrame: qt.NewQFrame(parent), url: qt.NewQUrl3(url)}
	card.SetFixedSize2(198, 220)
	card.iconWidget = widgets.NewIconWidgetIcon(icon, card.QWidget)
	card.titleLabel = qt.NewQLabel5(title, card.QWidget)
	wrapped, _ := gcommon.Wrap(content, 28, false)
	card.contentLabel = qt.NewQLabel5(wrapped, card.QWidget)
	card.urlWidget = widgets.NewIconWidgetIcon(gcommon.Link, card.QWidget)
	card.initWidget()
	return card
}

func (c *LinkCard) initWidget() {
	c.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))

	c.iconWidget.SetFixedSize2(54, 54)
	c.urlWidget.SetFixedSize2(16, 16)

	c.vBoxLayout = qt.NewQVBoxLayout(c.QWidget)
	c.vBoxLayout.SetSpacing(0)
	c.vBoxLayout.SetContentsMargins(24, 24, 0, 13)
	c.vBoxLayout.AddWidget(c.iconWidget.QWidget)
	c.vBoxLayout.AddSpacing(16)
	c.vBoxLayout.AddWidget(c.titleLabel.QWidget)
	c.vBoxLayout.AddSpacing(8)
	c.vBoxLayout.AddWidget(c.contentLabel.QWidget)
	c.urlWidget.Move(170, 192)

	c.titleLabel.SetObjectName("titleLabel")
	c.contentLabel.SetObjectName("contentLabel")
	c.SetObjectName("linkCard")

	c.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		qt.QDesktopServices_OpenUrl(c.url)
	})
}

// LinkCardView is a horizontally scrolling row of link cards.
type LinkCardView struct {
	*widgets.SingleDirectionScrollArea
	view       *qt.QWidget
	hBoxLayout *qt.QHBoxLayout
}

// NewLinkCardView builds a link card view.
func NewLinkCardView(parent *qt.QWidget) *LinkCardView {
	v := &LinkCardView{SingleDirectionScrollArea: widgets.NewSingleDirectionScrollArea(parent, qt.Horizontal)}
	v.view = qt.NewQWidget(v.QWidget)
	v.hBoxLayout = qt.NewQHBoxLayout(v.view)

	v.hBoxLayout.SetContentsMargins(36, 0, 0, 0)
	v.hBoxLayout.SetSpacing(12)

	v.SetWidget(v.view)
	v.SetWidgetResizable(true)
	v.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	v.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)

	v.view.SetObjectName("view")
	gallerycommon.StyleLinkCard.Apply(v.QWidget, gcommon.ThemeAuto)
	return v
}

// AddCard adds a link card to the view.
func (v *LinkCardView) AddCard(icon interface{}, title, content, url string) {
	card := NewLinkCard(icon, title, content, url, v.view)
	v.hBoxLayout.AddWidget(card.QWidget)
}
