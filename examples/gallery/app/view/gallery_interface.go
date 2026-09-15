package view

import (
	"strings"

	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/layout"
	"github.com/famei/gofluent/components/settings"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// SeparatorWidget draws a thin vertical separator line (port of
// app/view/gallery_interface.py SeparatorWidget).
type SeparatorWidget struct {
	*qt.QWidget
}

// NewSeparatorWidget builds a separator widget.
func NewSeparatorWidget(parent *qt.QWidget) *SeparatorWidget {
	w := &SeparatorWidget{QWidget: qt.NewQWidget(parent)}
	w.SetFixedSize2(6, 16)
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		pen := qt.NewQPen()
		defer pen.Delete()
		pen.SetCosmetic(true)
		var c *qt.QColor
		if gcommon.IsDarkTheme() {
			c = qt.NewQColor11(255, 255, 255, 21)
		} else {
			c = qt.NewQColor11(0, 0, 0, 15)
		}
		defer c.Delete()
		pen.SetColor(c)
		painter.SetPenWithPen(pen)

		x := w.Width() / 2
		painter.DrawLine2(x, 0, x, w.Height())
		painter.End()
	})
	return w
}

// ToolBar is the gallery header (title, subtitle and action buttons).
type ToolBar struct {
	*qt.QWidget
	TitleLabel    *widgets.TitleLabel
	SubtitleLabel *widgets.CaptionLabel

	documentButton *widgets.PushButton
	sourceButton   *widgets.PushButton
	themeButton    *widgets.ToolButton
	separator      *SeparatorWidget
	supportButton  *widgets.ToolButton
	feedbackButton *widgets.ToolButton

	vBoxLayout   *qt.QVBoxLayout
	buttonLayout *qt.QHBoxLayout
}

// NewToolBar builds a gallery tool bar.
func NewToolBar(title, subtitle string, parent *qt.QWidget) *ToolBar {
	t := &ToolBar{QWidget: qt.NewQWidget(parent)}
	t.TitleLabel = widgets.NewTitleLabelText(title, t.QWidget)
	t.SubtitleLabel = widgets.NewCaptionLabelText(subtitle, t.QWidget)

	t.documentButton = widgets.NewPushButtonIcon(gcommon.Document, gallerycommon.Tr("ToolBar", "Documentation"), t.QWidget)
	t.sourceButton = widgets.NewPushButtonIcon(gcommon.Code, gallerycommon.Tr("ToolBar", "Source"), t.QWidget)
	t.themeButton = widgets.NewToolButtonIcon(gcommon.Light, t.QWidget)
	t.separator = NewSeparatorWidget(t.QWidget)
	t.supportButton = widgets.NewToolButtonIcon(gcommon.Heart, t.QWidget)
	t.feedbackButton = widgets.NewToolButtonIcon(gcommon.Feedback, t.QWidget)

	t.vBoxLayout = qt.NewQVBoxLayout(t.QWidget)
	t.buttonLayout = qt.NewQHBoxLayout2()

	t.SetObjectName("toolBar")
	t.initWidget()
	return t
}

func (t *ToolBar) initWidget() {
	t.SetFixedHeight(138)
	t.vBoxLayout.SetSpacing(0)
	t.vBoxLayout.SetContentsMargins(36, 22, 36, 12)
	t.vBoxLayout.AddWidget(t.TitleLabel.QWidget)
	t.vBoxLayout.AddSpacing(4)
	t.vBoxLayout.AddWidget(t.SubtitleLabel.QWidget)
	t.vBoxLayout.AddSpacing(4)
	t.vBoxLayout.AddLayout2(t.buttonLayout.QLayout, 1)

	t.buttonLayout.SetSpacing(4)
	t.buttonLayout.SetContentsMargins(0, 0, 0, 0)
	t.buttonLayout.AddWidget3(t.documentButton.QWidget, 0, qt.AlignLeft)
	t.buttonLayout.AddWidget3(t.sourceButton.QWidget, 0, qt.AlignLeft)
	t.buttonLayout.AddStretchWithStretch(1)
	t.buttonLayout.AddWidget3(t.themeButton.QWidget, 0, qt.AlignRight)
	t.buttonLayout.AddWidget3(t.separator.QWidget, 0, qt.AlignRight)
	t.buttonLayout.AddWidget3(t.supportButton.QWidget, 0, qt.AlignRight)
	t.buttonLayout.AddWidget3(t.feedbackButton.QWidget, 0, qt.AlignRight)

	widgets.NewToolTipFilter(t.themeButton.QWidget, 300, widgets.ToolTipPositionTop)
	widgets.NewToolTipFilter(t.supportButton.QWidget, 300, widgets.ToolTipPositionTop)
	widgets.NewToolTipFilter(t.feedbackButton.QWidget, 300, widgets.ToolTipPositionTop)
	t.themeButton.SetToolTip(gallerycommon.Tr("ToolBar", "Toggle theme"))
	t.supportButton.SetToolTip(gallerycommon.Tr("ToolBar", "Support me"))
	t.feedbackButton.SetToolTip(gallerycommon.Tr("ToolBar", "Send feedback"))

	t.themeButton.OnClicked(func() { gcommon.ToggleTheme(true, true) })
	t.supportButton.OnClicked(func() { gallerycommon.SignalBusInstance.EmitSupport() })
	t.documentButton.OnClicked(func() { qt.QDesktopServices_OpenUrl(qt.NewQUrl3(gallerycommon.HELP_URL)) })
	t.sourceButton.OnClicked(func() { qt.QDesktopServices_OpenUrl(qt.NewQUrl3(gallerycommon.EXAMPLE_URL)) })
	t.feedbackButton.OnClicked(func() { qt.QDesktopServices_OpenUrl(qt.NewQUrl3(gallerycommon.FEEDBACK_URL)) })

	t.SubtitleLabel.SetTextColor(qt.NewQColor3(96, 96, 96), qt.NewQColor3(216, 216, 216))
}

// ExampleCard wraps a demo widget with a source-code footer that expands
// downward to reveal a short call pseudo-code snippet. It reuses the same
// ExpandSettingCard component as the Settings page, and lays the demo widget +
// source card out with the settings' ExpandLayout so the expand/collapse height
// change resizes the parent instead of re-flowing it (which is what left ghosts
// on the sibling demo widget).
type ExampleCard struct {
	*qt.QWidget
	Widget     *qt.QWidget
	Card       *qt.QFrame
	TopLayout  *qt.QHBoxLayout
	topWidget  *qt.QWidget
	titleLabel *widgets.StrongBodyLabel
	sourceCard *settings.ExpandSettingCard
	codeEdit   *qt.QPlainTextEdit
	vBoxLayout *layout.ExpandLayout
	cardLayout *layout.ExpandLayout
	stretch    int
	sourcePath string
}

// NewExampleCard builds an example card.
func NewExampleCard(title string, widget *qt.QWidget, sourcePath string, stretch int, parent *qt.QWidget) *ExampleCard {
	card := &ExampleCard{QWidget: qt.NewQWidget(parent), Widget: widget, stretch: stretch, sourcePath: sourcePath}
	card.titleLabel = widgets.NewStrongBodyLabelText(title, card.QWidget)
	card.Card = qt.NewQFrame(card.QWidget)
	card.sourceCard = settings.NewExpandSettingCard(gcommon.Code, gallerycommon.Tr("ExampleCard", "Source code"), sourcePath, card.Card.QWidget)
	card.codeEdit = qt.NewQPlainTextEdit(card.sourceCard.View().QWidget)
	card.topWidget = qt.NewQWidget(card.Card.QWidget)

	card.vBoxLayout = layout.NewExpandLayout(card.QWidget)
	card.cardLayout = layout.NewExpandLayout(card.Card.QWidget)
	card.TopLayout = qt.NewQHBoxLayout(card.topWidget)

	card.initWidget()
	return card
}

func (c *ExampleCard) initWidget() {
	c.initCodeView()
	c.initLayout()
	c.SetObjectName("exampleCard")
	c.Card.SetObjectName("card")
}

func (c *ExampleCard) initLayout() {
	c.vBoxLayout.SetSpacing(12)
	c.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	c.TopLayout.SetContentsMargins(12, 12, 12, 12)
	c.cardLayout.SetContentsMargins(0, 0, 0, 0)
	c.cardLayout.SetSpacing(0)

	c.vBoxLayout.AddWidget(c.titleLabel.QWidget)
	c.vBoxLayout.AddWidget(c.Card.QWidget)

	c.cardLayout.AddWidget(c.topWidget)
	c.cardLayout.AddWidget(c.sourceCard.QWidget)

	c.Widget.SetParent(c.topWidget)
	if c.stretch == 0 {
		c.TopLayout.AddWidget(c.Widget)
		c.TopLayout.AddStretchWithStretch(1)
	} else {
		// stretch != 0 fills the card width so height-for-width demo widgets
		// (the FlowLayout cards) wrap at the gallery width instead of their
		// narrow sizeHint width, keeping the card compact.
		c.TopLayout.AddWidget2(c.Widget, 1)
	}
	c.Widget.Show()

	// Derive the demo area height from the widget's own height plus the actual
	// TopLayout margins (not a hardcoded +24) so cards that override the margins
	// (e.g. the indeterminate progress bar card) still fit their demo widget.
	m := c.TopLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	c.topWidget.SetFixedHeight(c.demoHeight() + m.Top() + m.Bottom())

	c.titleLabel.AdjustSize()
	c.AdjustSize()
}

// demoHeight returns the vertical space the demo widget needs, reproducing what
// QLayout::SetMinimumSize reserves for a widget item:
// QWidgetItemV2::minimumSize() = qSmartMinSize(sizeHint, minimumSizeHint,
// minimumSize, maximumSize, sizePolicy).
//
// Using w.SizeHint() alone is wrong: setFixedHeight/setFixedSize change
// minimumSize/maximumSize, NOT sizeHint(), so LineEdit(33), SpinBox(33),
// TextEdit(150), ListFrame/TreeFrame(380), TableFrame(440), SwitchButton(22),
// DatePicker/TimePicker(30) and CommandBar(34) all need their fixed height.
// That is recovered by the explicit-minimum override (minH) below.
func (c *ExampleCard) demoHeight() int {
	w := c.Widget

	sh := w.SizeHint().Height()         // GoGC-armed — do NOT Delete
	msh := w.MinimumSizeHint().Height() // GoGC-armed — do NOT Delete

	h := sh
	if msh > h {
		h = msh // qMax(sizeHint, minimumSizeHint)
	}
	if maxH := w.MaximumHeight(); h > maxH {
		h = maxH // boundedTo(maximumSize)
	}
	if minH := w.MinimumHeight(); minH > 0 {
		h = minH // explicit minimum size wins (setFixedSize/setFixedHeight)
	}

	// Height-for-width widgets (the FlowLayout cards) wrap according to their
	// width. Their sizeHint() is the height at their (narrow) sizeHint width —
	// e.g. 500px for the FlowLayout — but when the card stretches them to the
	// gallery width they collapse to a couple of rows. Use the height at a
	// typical gallery width (~600px) so the card matches the laid-out result.
	if w.HasHeightForWidth() {
		if hfw := w.HeightForWidth(600); hfw > h {
			h = hfw
		}
	}

	if h < 0 {
		h = 0
	}
	return h
}

// AdjustSize fixes the card's own height from its children (the demo widget +
// source card inside the Card frame, plus the title label), so the outer
// ExpandLayout can stack cards by their current heights. Resize (not
// SetFixedHeight) is used so the Card/ExampleCard can still grow/shrink when
// the source card expands/collapses.
func (c *ExampleCard) AdjustSize() {
	// Re-derive the demo area height so a later TopLayout margin change is
	// honored (the indeterminate progress bar card overrides the margins).
	m := c.TopLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	c.topWidget.SetFixedHeight(c.demoHeight() + m.Top() + m.Bottom())

	cardH := c.topWidget.Height() + c.sourceCard.Height()
	c.Card.Resize(c.Card.Width(), cardH)
	c.Resize(c.Width(), c.titleLabel.Height()+c.vBoxLayout.Spacing()+cardH)
}

// initCodeView builds the read-only pseudo-code editor inside the expand card.
func (c *ExampleCard) initCodeView() {
	c.codeEdit.SetReadOnly(true)
	c.codeEdit.SetLineWrapMode(qt.QPlainTextEdit__WidgetWidth)
	c.codeEdit.SetTextInteractionFlags(qt.TextSelectableByMouse | qt.TextSelectableByKeyboard)
	c.codeEdit.SetFrameShape(qt.QFrame__NoFrame)
	c.codeEdit.SetObjectName("sourceCodeEdit")
	c.codeEdit.SetPlainText(codeSnippet(c.sourcePath))
	c.codeEdit.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	c.codeEdit.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)

	font := qt.QFontDatabase_SystemFont(qt.QFontDatabase__FixedFont)
	font.SetPixelSize(13)
	font.SetWeight(int(qt.QFont__Normal))
	c.codeEdit.SetFont(font)

	// An opaque background so the code text repaints cleanly during the collapse
	// scroll (a transparent editor leaves a text trail); square, inset by the
	// view layout, so it does not add a second rounded corner.
	gcommon.SetCustomStyleSheet(c.codeEdit.QWidget,
		"QPlainTextEdit{background: rgb(252, 252, 252); color: rgb(30, 30, 30); border: none;}",
		"QPlainTextEdit{background: rgb(40, 40, 40); color: rgb(230, 230, 230); border: none;}")

	// Size the editor to the full wrapped content so long code expands the view
	// instead of clipping (no scrollbars).
	c.codeEdit.SetFixedHeight(c.snippetHeight())

	c.sourceCard.ViewLayout().SetContentsMargins(12, 0, 12, 12)
	c.sourceCard.ViewLayout().AddWidget(c.codeEdit.QWidget)
}

// ToggleCode toggles the source-code expand area (also used by tests).
func (c *ExampleCard) ToggleCode() {
	c.sourceCard.ToggleExpand()
}

// snippetHeight estimates the wrapped code height with a conservative
// characters-per-line heuristic (avoids mutating the live QTextDocument).
func (c *ExampleCard) snippetHeight() int {
	const lineH = 18
	const charsPerLine = 90
	total := 0
	for _, ln := range strings.Split(codeSnippet(c.sourcePath), "\n") {
		n := len(ln)/charsPerLine + 1
		total += n
	}
	h := total*lineH + 8
	if h < 62 {
		h = 62
	}
	return h
}

// GalleryInterface is the scrollable base class of every gallery page.
type GalleryInterface struct {
	*widgets.ScrollArea
	View       *qt.QWidget
	ToolBar    *ToolBar
	VBoxLayout *layout.ExpandLayout
}

// NewGalleryInterface builds a gallery interface.
func NewGalleryInterface(title, subtitle string, parent *qt.QWidget) *GalleryInterface {
	g := &GalleryInterface{ScrollArea: widgets.NewScrollArea(parent)}
	g.View = qt.NewQWidget(g.QWidget)
	g.ToolBar = NewToolBar(title, subtitle, g.QWidget)
	g.VBoxLayout = layout.NewExpandLayout(g.View)

	g.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	g.SetViewportMargins(0, g.ToolBar.Height(), 0, 0)
	g.SetWidget(g.View)
	g.SetWidgetResizable(true)

	g.VBoxLayout.SetSpacing(30)
	g.VBoxLayout.SetContentsMargins(36, 20, 36, 36)

	g.View.SetObjectName("view")
	// The QScrollArea viewport paints its own light palette background by default
	// (autoFillBackground = true), and the gallery QSS only styles #view (the
	// content widget) transparent — never the viewport — so in dark mode the
	// content area stayed white. Stop the viewport auto-fill so the dark
	// StackedWidget surface underneath shows through.
	if vp := g.Viewport(); vp != nil {
		vp.SetAutoFillBackground(false)
	}
	gallerycommon.StyleGalleryInterface.Apply(g.QWidget, gcommon.ThemeAuto)

	g.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		g.ToolBar.Resize(g.Width(), g.ToolBar.Height())
	})
	return g
}

// AddExampleCard adds an example card and returns it.
func (g *GalleryInterface) AddExampleCard(title string, widget *qt.QWidget, sourcePath string, stretch int) *ExampleCard {
	card := NewExampleCard(title, widget, sourcePath, stretch, g.View)
	g.VBoxLayout.AddWidget(card.QWidget)
	return card
}

// ScrollToCard scrolls the vertical scroll bar to the card at index.
func (g *GalleryInterface) ScrollToCard(index int) {
	w := g.VBoxLayout.WidgetAt(index)
	if w == nil {
		return
	}
	g.VerticalScrollBar().SetValue(w.Y())
}
