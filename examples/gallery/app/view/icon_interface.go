package view

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// Icon card geometry. The grid cell is a little larger than the painted card so
// neighbouring cards keep a visible gap (the delegate insets the cell by
// iconCardMargin on every side). A card shows the glyph and the enum member
// only; the code point and the tags live in the info panel.
const (
	iconCardWidth  = 104
	iconCardHeight = 92
	iconCardMargin = 4
	iconCardRadius = 6

	iconGlyphSize = 28
	// The glyph is drawn as a font glyph in a iconGlyphSize box, and the Segoe
	// Fluent Icons glyphs fill that box, so the card content is centered on the
	// glyph box plus the name line below it (the name ink starts right at the top
	// of its line box, hence the extra offset).
	iconGlyphTop = 22
	iconNameTop  = 58
	iconNameH    = 16

	iconInfoPanelWidth = 232

	// zeroWidthSpace is an invisible break opportunity used to wrap long enum
	// member names inside the info panel.
	zeroWidthSpace = '\u200b'
)

// iconListViewQss keeps the icon grid transparent: the rounded surface and the
// border come from the #iconView frame (gallery qss/icon_interface.qss) and the
// cards are painted by iconGridDelegate. The stylesheet is theme independent, so
// the light and dark variants are identical.
const iconListViewQss = "#iconListView{background:transparent;border:none;outline:0;padding:0px;}" +
	"#iconListView::item{background:transparent;border:0px;padding:0px;}"

// ---------------------------------------------------------------------------
// icon library: IconsData.json
// ---------------------------------------------------------------------------

// iconEntry is one icon of the browser: one enum member of
// common/SegoeFluentIcons.go with its metadata. The embedded IconsData.json (the
// WinUI Gallery iconography list) is the enumeration source — it holds every
// SegoeFluentIcon member, where "Code" is the enum value (the code point of the
// glyph, e.g. "E721") and "Name" is the member name (e.g. "Search").
type iconEntry struct {
	icon gcommon.SegoeFluentIcon // enum value — the Segoe Fluent Icons glyph
	name string                  // enum member / icon name, IconsData.json "Name"
	code string                  // code point in hex, IconsData.json "Code"
	tags []string                // IconsData.json "Tags"

	// Pre-lower-cased copies of the searchable fields: filtering happens on
	// every keystroke, so it only does substring scans and never re-lower-cases
	// or re-joins anything.
	nameKey string
	tagKey  string
	codeKey string
}

func (e *iconEntry) buildSearchKeys() {
	e.nameKey = strings.ToLower(e.name)
	e.tagKey = strings.ToLower(strings.Join(e.tags, " "))
	e.codeKey = strings.ToLower(e.code) + " u+" + strings.ToLower(e.code)
}

// matches reports whether the entry satisfies every query term. A term matches
// when it is contained in the enum member (which is also the icon name), in one
// of the tags or in the code point — i.e. one keyword finds the icon through any
// of the three search modes the browser offers.
func (e *iconEntry) matches(terms []string) bool {
	for _, term := range terms {
		if !strings.Contains(e.nameKey, term) &&
			!strings.Contains(e.tagKey, term) &&
			!strings.Contains(e.codeKey, term) {
			return false
		}
	}
	return true
}

// enumMember returns the enum member as it is written in Go code.
func (e *iconEntry) enumMember() string { return "common." + e.name }

// enumMemberText is the enum member as displayed by the info panel: the same
// text with zero-width spaces at the natural break points of the identifier
// (after the package qualifier and in front of every camelCase hump). The
// characters are invisible; they only give the word-wrapped label the chance to
// move a long member onto the next line instead of running into the copy button.
// Copied text comes from enumMember(), so it never contains them.
func (e *iconEntry) enumMemberText() string {
	runes := []rune(e.enumMember())
	var b strings.Builder
	b.Grow(len(runes) + len(runes)/2)
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if prev := runes[i-1]; (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') {
				b.WriteRune(zeroWidthSpace)
			}
		}
		b.WriteRune(r)
		if r == '.' {
			b.WriteRune(zeroWidthSpace)
		}
	}
	return b.String()
}

var (
	iconLibraryOnce sync.Once
	iconLibrary     []*iconEntry
)

// allIconEntries returns the icon library, built once on first use from the
// embedded IconsData.json (one record per enum member, in code point order).
func allIconEntries() []*iconEntry {
	iconLibraryOnce.Do(func() { iconLibrary = buildIconLibrary() })
	return iconLibrary
}

func buildIconLibrary() []*iconEntry {
	records := resource.GetIconsData()
	entries := make([]*iconEntry, 0, len(records))
	for i := range records {
		glyph, ok := glyphOfCode(records[i].Code)
		if !ok {
			continue // a record without a usable code point is not an icon
		}
		e := &iconEntry{
			icon: gcommon.SegoeFluentIcon(glyph),
			name: records[i].Name,
			code: strings.ToUpper(strings.TrimSpace(records[i].Code)),
			tags: records[i].Tags,
		}
		e.buildSearchKeys()
		entries = append(entries, e)
	}
	return entries
}

// glyphOfCode converts an IconsData.json "Code" — the enum value — into the
// glyph rendered by the Segoe Fluent Icons font.
func glyphOfCode(code string) (string, bool) {
	v, err := strconv.ParseUint(strings.TrimSpace(code), 16, 32)
	if err != nil || v == 0 || v > 0x10FFFF {
		return "", false
	}
	return string(rune(v)), true
}

// ---------------------------------------------------------------------------
// model: the icon library as a list model
// ---------------------------------------------------------------------------

// iconListModel exposes the icon library as a flat QAbstractListModel. It is
// the source model of iconFilterProxy and is never modified after construction:
// searching only changes which rows the proxy lets through, so the view is never
// rebuilt and no widget is ever created per icon.
type iconListModel struct {
	*qt.QAbstractListModel
	entries []*iconEntry

	// miqt hands the *QVariant returned by data() to C++, which copies the value
	// but never frees the Go wrapper, so allocating a fresh QVariant per call
	// would leak one per painted cell. The cache bounds that to one wrapper per
	// distinct string.
	variantCache map[string]*qt.QVariant
	empty        *qt.QVariant
	destroyed    bool
}

// newIconListModel builds the icon list model.
func newIconListModel(entries []*iconEntry, parent *qt.QObject) *iconListModel {
	m := &iconListModel{
		QAbstractListModel: qt.NewQAbstractListModel2(parent),
		entries:            entries,
		variantCache:       make(map[string]*qt.QVariant, 16),
		empty:              qt.NewQVariant(),
	}

	m.OnRowCount(func(parent *qt.QModelIndex) int {
		if parent != nil && parent.IsValid() {
			return 0 // flat list: no children
		}
		return len(m.entries)
	})
	m.OnData(m.itemData)

	// The model is owned by the view (Qt deletes it), so the cached QVariants
	// are released here rather than leaking with the Go wrapper.
	m.QObject.OnDestroyed(func() {
		m.destroyed = true
		for _, v := range m.variantCache {
			v.Delete()
		}
		m.variantCache = nil
		m.empty.Delete()
		m.empty = nil
	})
	return m
}

// entryAt returns the entry of a source row.
func (m *iconListModel) entryAt(row int) *iconEntry {
	if row < 0 || row >= len(m.entries) {
		return nil
	}
	return m.entries[row]
}

// itemData implements QAbstractItemModel::data. Only the display role is served:
// iconGridDelegate paints the cards from the Go entries, and the grid shows no
// tooltip at all (the details live in the info panel).
func (m *iconListModel) itemData(index *qt.QModelIndex, role int) *qt.QVariant {
	if index == nil || !index.IsValid() {
		return m.emptyVariant()
	}
	e := m.entryAt(index.Row())
	if e == nil {
		return m.emptyVariant()
	}
	switch qt.ItemDataRole(role) {
	case qt.DisplayRole:
		return m.cachedText(e.name)
	}
	return m.emptyVariant()
}

// emptyVariant returns the shared "no data" variant.
func (m *iconListModel) emptyVariant() *qt.QVariant {
	if m.destroyed {
		return qt.NewQVariant()
	}
	return m.empty
}

// cachedText returns a QVariant for text, reusing the wrapper of a previous call
// (the value is copied by Qt, so one wrapper can serve every request).
func (m *iconListModel) cachedText(text string) *qt.QVariant {
	if m.destroyed {
		return qt.NewQVariant14(text)
	}
	if v, ok := m.variantCache[text]; ok {
		return v
	}
	v := qt.NewQVariant14(text)
	m.variantCache[text] = v
	return v
}

// ---------------------------------------------------------------------------
// filter: multi-dimensional search
// ---------------------------------------------------------------------------

// iconFilterProxy is the QSortFilterProxyModel between iconListModel and the
// view. filterAcceptsRow is implemented in Go because a query has to match the
// enum member, the icon name, the tags and the code point at once — Qt's
// built-in filter only ever looks at a single role.
type iconFilterProxy struct {
	*qt.QSortFilterProxyModel
	source *iconListModel
	terms  []string

	// visible holds the source rows that pass the current query, in source
	// order. filterAcceptsRow applies exactly the same predicate to the same
	// (never sorted, never reset) source model, so visible[row] is always the
	// source row behind proxy row `row`. That lets the delegate and the
	// selection handler look an entry up in O(1) instead of mapping every
	// index through Qt while painting.
	visible []int
}

// newIconFilterProxy builds the filter proxy around source.
func newIconFilterProxy(source *iconListModel, parent *qt.QObject) *iconFilterProxy {
	p := &iconFilterProxy{QSortFilterProxyModel: qt.NewQSortFilterProxyModel2(parent), source: source}
	p.SetSourceModel(source.QAbstractItemModel)
	p.SetDynamicSortFilter(true)
	p.rebuildVisible()
	p.OnFilterAcceptsRow(func(_ func(sourceRow int, sourceParent *qt.QModelIndex) bool, sourceRow int, _ *qt.QModelIndex) bool {
		e := source.entryAt(sourceRow)
		return e != nil && e.matches(p.terms)
	})
	return p
}

// SetQuery applies a new search query. It runs on every keystroke, so all it
// does is one pass over the in-memory entries (a few thousand substring tests)
// plus an incremental proxy invalidate — no Qt model reset, no widget creation
// and therefore no blocked UI thread.
func (p *iconFilterProxy) SetQuery(query string) {
	terms := strings.Fields(strings.ToLower(query))
	if slices.Equal(p.terms, terms) {
		return
	}
	p.terms = terms
	p.rebuildVisible()
	p.InvalidateFilter()
}

// Count returns the number of icons passing the current query.
func (p *iconFilterProxy) Count() int { return len(p.visible) }

// entryAt returns the entry behind a proxy index (nil for an invalid index).
func (p *iconFilterProxy) entryAt(index *qt.QModelIndex) *iconEntry {
	if index == nil || !index.IsValid() {
		return nil
	}
	row := index.Row()
	if row < 0 || row >= len(p.visible) {
		return nil
	}
	return p.source.entryAt(p.visible[row])
}

func (p *iconFilterProxy) rebuildVisible() {
	p.visible = p.visible[:0]
	for row, e := range p.source.entries {
		if e.matches(p.terms) {
			p.visible = append(p.visible, row)
		}
	}
}

// ---------------------------------------------------------------------------
// delegate: paints one icon card per grid cell
// ---------------------------------------------------------------------------

// iconCardPalette holds every color the delegate needs for one theme, so the
// paint path never allocates a QColor.
type iconCardPalette struct {
	background *qt.QColor
	border     *qt.QColor
	hover      *qt.QColor
	name       *qt.QColor
	glyph      *qt.QColor

	selectedBackground *qt.QColor
	selectedBorder     *qt.QColor
	selectedGlyph      *qt.QColor
}

// newIconCardPalette resolves the card colors. They mirror the IconCard rules of
// the gallery's qss/icon_interface.qss (light: rgb(251,251,251) on
// rgb(229,229,229); dark: rgb(43,43,43) on rgb(29,29,29); selected: the theme
// color with inverted label text), since the cards are painted and not styled.
func newIconCardPalette(dark bool, primary *qt.QColor) *iconCardPalette {
	p := &iconCardPalette{
		selectedBackground: qt.NewQColor9(primary), // copy: the config keeps the original
		selectedBorder:     qt.NewQColor9(primary),
	}
	if dark {
		p.background = qt.NewQColor3(43, 43, 43)
		p.border = qt.NewQColor3(29, 29, 29)
		p.hover = qt.NewQColor3(52, 52, 52)
		p.name = qt.NewQColor3(207, 207, 207)
		p.glyph = qt.NewQColor3(255, 255, 255)
		p.selectedGlyph = qt.NewQColor3(0, 0, 0)
	} else {
		p.background = qt.NewQColor3(251, 251, 251)
		p.border = qt.NewQColor3(229, 229, 229)
		p.hover = qt.NewQColor3(244, 244, 244)
		p.name = qt.NewQColor3(96, 96, 96)
		p.glyph = qt.NewQColor3(0, 0, 0)
		p.selectedGlyph = qt.NewQColor3(255, 255, 255)
	}
	return p
}

// Delete releases every color of the palette.
func (p *iconCardPalette) Delete() {
	colors := []*qt.QColor{
		p.background, p.border, p.hover, p.name, p.glyph,
		p.selectedBackground, p.selectedBorder, p.selectedGlyph,
	}
	for _, c := range colors {
		if c != nil {
			c.Delete()
		}
	}
}

// iconGridDelegate paints the icon cards of the grid: the rounded card surface,
// the Segoe Fluent Icons glyph and the enum member. Painting (instead of one
// widget per icon) is what keeps the browser usable with the full 1500+ icon
// library.
type iconGridDelegate struct {
	*qt.QStyledItemDelegate
	view  *qt.QListView
	proxy *iconFilterProxy

	// cardSize is shared by every sizeHint call: Qt takes the returned QSize by
	// value, while miqt copies it out of the Go pointer, so a fresh QSize per
	// call would leak.
	cardSize *qt.QSize

	nameFont    *qt.QFont
	nameMetrics *qt.QFontMetrics

	hoverRow int

	palette    *iconCardPalette
	paletteKey string
}

// newIconGridDelegate builds the grid delegate.
func newIconGridDelegate(view *qt.QListView, proxy *iconFilterProxy) *iconGridDelegate {
	d := &iconGridDelegate{
		QStyledItemDelegate: qt.NewQStyledItemDelegate2(view.QObject),
		view:                view,
		proxy:               proxy,
		cardSize:            qt.NewQSize2(iconCardWidth, iconCardHeight),
		nameFont:            qt.NewQFont2("Segoe UI"),
		hoverRow:            -1,
	}
	d.nameFont.SetPixelSize(11)
	// The metrics are built against the view's paint device so that eliding uses
	// the same device pixel ratio the cards are painted with.
	if vp := view.Viewport(); vp != nil {
		d.nameMetrics = qt.NewQFontMetrics2(d.nameFont, vp.QPaintDevice)
	} else {
		d.nameMetrics = qt.NewQFontMetrics(d.nameFont)
	}
	d.installPaint()

	d.QObject.OnDestroyed(func() {
		d.nameMetrics.Delete()
		d.nameFont.Delete()
		d.cardSize.Delete()
		if d.palette != nil {
			d.palette.Delete()
			d.palette = nil
		}
	})
	return d
}

// installPaint overrides paint and sizeHint. The base paint is deliberately not
// called: an icon card has no text, decoration or selection chrome from the
// style, everything on it is drawn here.
func (d *iconGridDelegate) installPaint() {
	d.OnPaint(func(_ func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		e := d.proxy.entryAt(index)
		if e == nil {
			return
		}
		d.paintCard(painter, option, index, e)
	})

	d.OnSizeHint(func(_ func(option *qt.QStyleOptionViewItem, index *qt.QModelIndex) *qt.QSize, _ *qt.QStyleOptionViewItem, _ *qt.QModelIndex) *qt.QSize {
		return d.cardSize
	})
}

// SetHoverRow records the hovered row and repaints the grid.
func (d *iconGridDelegate) SetHoverRow(row int) {
	if d.hoverRow == row {
		return
	}
	d.hoverRow = row
	if vp := d.view.Viewport(); vp != nil {
		vp.Update()
	}
}

// cardPalette returns the palette of the active theme, rebuilding it when the
// theme or the theme color changed (the gallery can switch both at run time).
func (d *iconGridDelegate) cardPalette() *iconCardPalette {
	// ThemeColor returns the color stored in the config (not a copy) — do NOT
	// Delete it.
	primary := gcommon.QConfigInstance.ThemeColor()
	key := gcommon.CurrentTheme().String() + "|" + primary.Name()
	if d.palette != nil && d.paletteKey == key {
		return d.palette
	}
	if d.palette != nil {
		d.palette.Delete()
	}
	d.paletteKey = key
	d.palette = newIconCardPalette(gcommon.IsDarkTheme(), primary)
	return d.palette
}

func (d *iconGridDelegate) paintCard(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex, e *iconEntry) {
	pal := d.cardPalette()

	isSelected := option.State()&qt.QStyle__State_Selected != 0
	isHover := !isSelected &&
		(d.hoverRow == index.Row() || option.State()&qt.QStyle__State_MouseOver != 0)

	cell := option.Rect() // GoGC-armed value copy — do NOT Delete

	nameColor := pal.name
	glyphColor := pal.glyph
	background := pal.background
	border := pal.border
	switch {
	case isSelected:
		background, border = pal.selectedBackground, pal.selectedBorder
		nameColor, glyphColor = pal.selectedGlyph, pal.selectedGlyph
	case isHover:
		background = pal.hover
	}

	painter.Save()
	defer painter.Restore()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)

	// Card surface. The half-pixel offset keeps the 1px border crisp; the brush
	// and the pen are copied by setBrush / setPen, so the temporaries are freed
	// immediately.
	card := qt.NewQRectF4(
		float64(cell.X()+iconCardMargin)+0.5,
		float64(cell.Y()+iconCardMargin)+0.5,
		float64(cell.Width()-2*iconCardMargin-1),
		float64(cell.Height()-2*iconCardMargin-1),
	)
	brush := qt.NewQBrush3(background)
	pen := qt.NewQPen3(border)
	painter.SetBrush(brush)
	painter.SetPenWithPen(pen)
	brush.Delete()
	pen.Delete()
	painter.DrawRoundedRect(card, iconCardRadius, iconCardRadius)
	card.Delete()

	// Glyph: RenderGlyph fills the Segoe Fluent Icons character as a path, so it
	// stays crisp at any device pixel ratio.
	glyph := qt.NewQRectF4(
		float64(cell.X()+(cell.Width()-iconGlyphSize)/2),
		float64(cell.Y()+iconGlyphTop),
		iconGlyphSize,
		iconGlyphSize,
	)
	e.icon.RenderGlyph(painter, glyph, gcommon.ThemeAuto, glyphColor)
	glyph.Delete()

	// Enum member (the card shows the member only, elided to the card width).
	// It is clipped to the card so a metrics / device-DPI rounding difference can
	// never bleed into the next card.
	clip := qt.NewQRect4(cell.X()+iconCardMargin, cell.Y()+iconCardMargin, cell.Width()-2*iconCardMargin, cell.Height()-2*iconCardMargin)
	painter.SetClipping(true)
	painter.SetClipRectWithQRect(clip)
	clip.Delete()

	painter.SetFont(d.nameFont)
	painter.SetPen(nameColor)
	nameRect := qt.NewQRect4(cell.X()+4, cell.Y()+iconNameTop, cell.Width()-8, iconNameH)
	painter.DrawText6(nameRect, int(qt.AlignHCenter|qt.AlignVCenter),
		d.nameMetrics.ElidedText(e.name, qt.ElideRight, nameRect.Width()))
	nameRect.Delete()
}

// ---------------------------------------------------------------------------
// info panel
// ---------------------------------------------------------------------------

// iconInfoPanel shows the details of the current icon.
type iconInfoPanel struct {
	*qt.QFrame
	nameLabel          *qt.QLabel
	iconWidget         *widgets.IconWidget
	iconNameTitleLabel *qt.QLabel
	iconNameLabel      *qt.QLabel
	enumNameTitleLabel *qt.QLabel
	enumNameLabel      *qt.QLabel
	enumCopyButton     *widgets.PushButton
	codeTitleLabel     *qt.QLabel
	codeLabel          *qt.QLabel
	tagsTitleLabel     *qt.QLabel
	tagsLabel          *qt.QLabel
	vBoxLayout         *qt.QVBoxLayout

	entry *iconEntry
}

// newIconInfoPanel builds an icon info panel.
func newIconInfoPanel(parent *qt.QWidget) *iconInfoPanel {
	p := &iconInfoPanel{QFrame: qt.NewQFrame(parent)}
	p.SetObjectName("iconInfoPanel")
	p.nameLabel = qt.NewQLabel5("", p.QWidget)
	p.iconWidget = widgets.NewIconWidget(p.QWidget)
	p.iconNameTitleLabel = qt.NewQLabel5(gallerycommon.Tr("IconInfoPanel", "Icon name"), p.QWidget)
	p.iconNameLabel = qt.NewQLabel5("", p.QWidget)
	p.enumNameTitleLabel = qt.NewQLabel5(gallerycommon.Tr("IconInfoPanel", "Enum member"), p.QWidget)
	p.enumNameLabel = qt.NewQLabel5("", p.QWidget)
	p.enumCopyButton = widgets.NewPushButtonText(gallerycommon.Tr("IconInfoPanel", "Copy"), p.QWidget)
	p.enumCopyButton.SetObjectName("iconEnumCopyButton")
	p.codeTitleLabel = qt.NewQLabel5(gallerycommon.Tr("IconInfoPanel", "Code"), p.QWidget)
	p.codeLabel = qt.NewQLabel5("", p.QWidget)
	p.tagsTitleLabel = qt.NewQLabel5(gallerycommon.Tr("IconInfoPanel", "Tags"), p.QWidget)
	p.tagsLabel = qt.NewQLabel5("", p.QWidget)

	p.vBoxLayout = qt.NewQVBoxLayout(p.QWidget)
	p.vBoxLayout.SetContentsMargins(16, 20, 16, 20)
	p.vBoxLayout.SetSpacing(0)

	p.vBoxLayout.AddWidget(p.nameLabel.QWidget)
	p.vBoxLayout.AddSpacing(16)
	p.vBoxLayout.AddWidget(p.iconWidget.QWidget)
	p.vBoxLayout.AddSpacing(45)
	p.addField(p.iconNameTitleLabel, p.iconNameLabel)
	p.vBoxLayout.AddSpacing(20)
	p.addFieldWithAction(p.enumNameTitleLabel, p.enumNameLabel, p.enumCopyButton.QWidget)
	p.vBoxLayout.AddSpacing(20)
	p.addField(p.codeTitleLabel, p.codeLabel)
	p.vBoxLayout.AddSpacing(20)
	p.addField(p.tagsTitleLabel, p.tagsLabel)
	p.vBoxLayout.AddStretchWithStretch(1)

	p.iconWidget.SetFixedSize2(48, 48)
	p.SetFixedWidth(iconInfoPanelWidth)

	p.nameLabel.SetObjectName("nameLabel")
	p.iconNameTitleLabel.SetObjectName("subTitleLabel")
	p.enumNameTitleLabel.SetObjectName("subTitleLabel")
	p.codeTitleLabel.SetObjectName("subTitleLabel")
	p.tagsTitleLabel.SetObjectName("subTitleLabel")
	p.enumNameLabel.SetObjectName("iconEnumLabel")
	p.iconNameLabel.SetObjectName("iconNameLabel")
	p.codeLabel.SetObjectName("iconCodeLabel")
	p.tagsLabel.SetObjectName("iconTagsLabel")

	// Long names and the tag list must wrap inside the fixed-width panel instead
	// of widening it.
	p.nameLabel.SetWordWrap(true)
	p.iconNameLabel.SetWordWrap(true)
	p.enumNameLabel.SetWordWrap(true)
	p.tagsLabel.SetWordWrap(true)
	p.tagsLabel.SetAlignment(qt.AlignLeft | qt.AlignTop)

	// The copy button sits right behind the enum member and puts it on the
	// clipboard in the form used in Go code (e.g. "common.Cancel"). The fluent
	// tooltip filter has to be installed before SetToolTip — it replaces the
	// native tooltip with the styled one.
	widgets.NewToolTipFilter(p.enumCopyButton.QWidget, 500, widgets.ToolTipPositionTop)
	p.enumCopyButton.SetToolTip(gallerycommon.Tr("IconInfoPanel", "Copy the enum member for the current icon"))
	p.enumCopyButton.SetFixedSize2(58, 26)
	p.enumCopyButton.OnClicked(p.copyEnumMember)
	return p
}

// addField appends a title/value pair to the panel.
func (p *iconInfoPanel) addField(title, value *qt.QLabel) {
	p.vBoxLayout.AddWidget(title.QWidget)
	p.vBoxLayout.AddSpacing(5)
	p.vBoxLayout.AddWidget(value.QWidget)
}

// addFieldWithAction appends a title plus a value row that carries the action
// button right behind the value (the copy button of the enum member). A value
// that is too long for the remaining width wraps onto the next line, because the
// label is word wrapped and its text carries break opportunities (see
// iconEntry.enumMemberText).
//
// The row is a nested *layout*, not a container widget: the label must stay a
// direct child of the panel, because that is what the gallery stylesheet matches
// ("IconInfoPanel > QLabel" → the fluent font and the light/dark text color).
// Inside a wrapper widget the label would silently fall back to the default Qt
// font and palette.
func (p *iconInfoPanel) addFieldWithAction(title, value *qt.QLabel, action *qt.QWidget) {
	p.vBoxLayout.AddWidget(title.QWidget)
	p.vBoxLayout.AddSpacing(5)

	row := qt.NewQHBoxLayout2()
	row.SetContentsMargins(0, 0, 0, 0)
	row.SetSpacing(8)
	row.AddWidget2(value.QWidget, 1)
	row.AddWidget3(action, 0, qt.AlignTop|qt.AlignRight)
	p.vBoxLayout.AddLayout(row.QLayout)
}

// copyEnumMember copies the enum member of the current icon to the clipboard.
func (p *iconInfoPanel) copyEnumMember() {
	if p.entry == nil {
		return
	}
	if cb := qt.QGuiApplication_Clipboard(); cb != nil {
		cb.SetText(p.entry.enumMember())
	}
}

// setEntry shows the details of an icon.
func (p *iconInfoPanel) setEntry(e *iconEntry) {
	p.entry = e
	p.iconWidget.SetIcon(e.icon)
	p.nameLabel.SetText(e.name)
	p.iconNameLabel.SetText(e.name)
	p.enumNameLabel.SetText(e.enumMemberText())
	p.codeLabel.SetText("U+" + e.code + "  (0x" + e.code + ")")
	if len(e.tags) == 0 {
		p.tagsLabel.SetText("-")
	} else {
		p.tagsLabel.SetText(strings.Join(e.tags, ", "))
	}
}

// ---------------------------------------------------------------------------
// the icon browser
// ---------------------------------------------------------------------------

// iconCardView hosts the searchable icon grid: a QListView in icon mode backed
// by iconListModel → iconFilterProxy, the iconGridDelegate that paints the cards
// and the info panel of the current icon.
type iconCardView struct {
	*qt.QWidget
	titleLabel     *widgets.StrongBodyLabel
	searchLineEdit *widgets.SearchLineEdit
	countLabel     *widgets.CaptionLabel
	searchWidget   *qt.QWidget
	view           *qt.QFrame
	listView       *qt.QListView
	listModel      *iconListModel
	filterModel    *iconFilterProxy
	delegate       *iconGridDelegate
	scrollDelegate *widgets.SmoothScrollDelegate
	infoPanel      *iconInfoPanel

	vBoxLayout   *qt.QVBoxLayout
	hBoxLayout   *qt.QHBoxLayout
	searchLayout *qt.QHBoxLayout

	entries []*iconEntry
}

// newIconCardView builds the icon browser.
func newIconCardView(parent *qt.QWidget) *iconCardView {
	v := &iconCardView{QWidget: qt.NewQWidget(parent), entries: allIconEntries()}
	v.SetObjectName("iconCardView")

	v.titleLabel = widgets.NewStrongBodyLabelText(gallerycommon.Tr("IconCardView", "Fluent Icons Library"), v.QWidget)
	v.searchLineEdit = widgets.NewSearchLineEdit(v.QWidget)
	v.searchLineEdit.SetObjectName("iconSearchLineEdit")
	v.countLabel = widgets.NewCaptionLabel(v.QWidget)
	v.countLabel.SetObjectName("iconCountLabel")
	v.searchWidget = qt.NewQWidget(v.QWidget)

	v.view = qt.NewQFrame(v.QWidget)
	v.view.SetObjectName("iconView")
	// Expanding so the icon grid (and info panel) fill the remaining vertical
	// space instead of collapsing to the info panel's sizeHint height.
	v.view.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)

	v.listView = qt.NewQListView(v.view.QWidget)
	v.listView.SetObjectName("iconListView")
	v.listView.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)
	v.infoPanel = newIconInfoPanel(v.QWidget)

	v.vBoxLayout = qt.NewQVBoxLayout(v.QWidget)
	v.hBoxLayout = qt.NewQHBoxLayout(v.view.QWidget)
	v.searchLayout = qt.NewQHBoxLayout(v.searchWidget)

	v.initModels()
	v.initWidget()
	return v
}

// initModels wires the model / proxy / delegate chain of the icon grid.
func (v *iconCardView) initModels() {
	// Every Qt object below is owned by the list view (it owns the model and the
	// delegate), so only the views of "who owns what" matter here: the Go
	// wrappers are kept alive by the fields of this struct.
	v.listModel = newIconListModel(v.entries, v.listView.QObject)
	v.filterModel = newIconFilterProxy(v.listModel, v.listView.QObject)
	v.delegate = newIconGridDelegate(v.listView, v.filterModel)

	v.listView.SetModel(v.filterModel.QAbstractItemModel)
	v.listView.SetItemDelegate(v.delegate.QAbstractItemDelegate)

	// IconMode with a fixed grid makes the view lay out a wrapping grid of
	// equally sized cells and recycle nothing but paint calls.
	v.listView.SetViewMode(qt.QListView__IconMode)
	v.listView.SetFlow(qt.QListView__LeftToRight)
	v.listView.SetWrapping(true)
	v.listView.SetResizeMode(qt.QListView__Adjust)
	v.listView.SetMovement(qt.QListView__Static)
	v.listView.SetUniformItemSizes(true)
	v.listView.SetSpacing(0)
	v.listView.SetSelectionMode(qt.QAbstractItemView__SingleSelection)
	v.listView.SetVerticalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
	v.listView.SetEditTriggers(qt.QAbstractItemView__NoEditTriggers)
	v.listView.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	v.listView.SetMouseTracking(true)

	grid := qt.NewQSize2(iconCardWidth, iconCardHeight)
	v.listView.SetGridSize(grid) // setGridSize copies the size
	grid.Delete()

	// The rounded surface comes from the #iconView frame, so the view and its
	// viewport must stay transparent.
	v.listView.SetStyleSheet(iconListViewQss)
	if vp := v.listView.Viewport(); vp != nil {
		vp.SetAutoFillBackground(false)
		vp.SetMouseTracking(true)
	}

	// Fluent overlay scroll bar (the native one is hidden by the delegate).
	v.scrollDelegate = widgets.NewSmoothScrollDelegate(v.listView.QAbstractScrollArea, false)

	v.installItemTracking()
}

// installItemTracking keeps the hovered row of the delegate in sync so the card
// under the cursor is highlighted.
func (v *iconCardView) installItemTracking() {
	v.listView.OnEntered(func(index *qt.QModelIndex) {
		v.delegate.SetHoverRow(index.Row())
	})

	// The viewport is owned by Qt, so its leave event is caught with an event
	// filter instead of a virtual override.
	filter := qt.NewQObject2(v.listView.QObject)
	if vp := v.listView.Viewport(); vp != nil {
		vp.InstallEventFilter(filter)
	}
	filter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		switch event.Type() {
		case qt.QEvent__ToolTip:
			// The cards carry no tooltip: the item view would otherwise answer
			// this with the model's ToolTipRole in a native popup, so the request
			// is swallowed here.
			return true
		case qt.QEvent__Leave:
			v.delegate.SetHoverRow(-1)
		}
		return super(watched, event)
	})

	if sm := v.listView.SelectionModel(); sm != nil {
		sm.OnCurrentChanged(func(current *qt.QModelIndex, _ *qt.QModelIndex) {
			// A filtered-out icon leaves no current index; the panel then keeps
			// showing the last valid icon instead of blanking out.
			if e := v.filterModel.entryAt(current); e != nil {
				v.infoPanel.setEntry(e)
			}
		})
	}
}

func (v *iconCardView) initWidget() {
	v.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	v.vBoxLayout.SetSpacing(12)
	v.vBoxLayout.AddWidget(v.titleLabel.QWidget)
	v.vBoxLayout.AddWidget(v.searchWidget)
	v.vBoxLayout.AddWidget(v.view.QWidget)

	v.searchLayout.SetContentsMargins(0, 0, 0, 0)
	v.searchLayout.SetSpacing(10)
	v.searchLayout.AddWidget2(v.searchLineEdit.QWidget, 1)
	v.searchLayout.AddWidget3(v.countLabel.QWidget, 0, qt.AlignRight|qt.AlignVCenter)

	v.hBoxLayout.SetSpacing(0)
	v.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	v.hBoxLayout.AddWidget2(v.listView.QWidget, 1)
	v.hBoxLayout.AddWidget3(v.infoPanel.QWidget, 0, qt.AlignRight)

	v.setQss()

	v.searchLineEdit.SetPlaceholderText(gallerycommon.Tr("LineEdit", "Search icons"))
	// Live filtering: every keystroke re-runs the proxy filter.
	v.searchLineEdit.OnTextChanged(func(text string) { v.search(text) })
	v.searchLineEdit.SearchSignal = func(text string) { v.search(text) }
	v.searchLineEdit.ClearSignal = func() { v.search("") }

	v.updateCountLabel()
	v.selectFirstResult()
}

func (v *iconCardView) setQss() {
	gallerycommon.StyleIconInterface.Apply(v.QWidget, gcommon.ThemeAuto)
}

// search applies a query to the grid. It runs on every keystroke: only the
// proxy filter is invalidated, the model and the widgets are left untouched.
func (v *iconCardView) search(keyWord string) {
	v.filterModel.SetQuery(keyWord)
	v.updateCountLabel()
	// Keep the highlighted card and the info panel in sync after the result set
	// changed: the first hit becomes the current icon (nothing happens for an
	// empty result set, the panel then keeps the last icon).
	v.selectFirstResult()
}

// updateCountLabel refreshes the "n icons" counter next to the search box.
func (v *iconCardView) updateCountLabel() {
	n := v.filterModel.Count()
	if n == 1 {
		v.countLabel.SetText(gallerycommon.Tr("IconCardView", "1 icon"))
		return
	}
	v.countLabel.SetText(fmt.Sprintf(gallerycommon.Tr("IconCardView", "%d icons"), n))
}

// selectFirstResult selects the first icon of the current result set so the info
// panel always shows something (also right after a search narrowed the grid).
func (v *iconCardView) selectFirstResult() {
	if v.filterModel.Count() == 0 {
		return
	}
	// miqt's index() glue dereferences the parent argument, so a nil parent
	// would crash the process: an explicit invalid root index is passed instead.
	root := qt.NewQModelIndex()
	defer root.Delete()

	index := v.filterModel.Index(0, 0, root) // GoGC-armed — do NOT Delete
	v.listView.SetCurrentIndex(index)        // selects the index (ClearAndSelect)
	if e := v.filterModel.entryAt(index); e != nil {
		v.infoPanel.setEntry(e)
	}
}

// ---------------------------------------------------------------------------
// page
// ---------------------------------------------------------------------------

// IconInterface is the "Icons" gallery page: the full Segoe Fluent Icons library
// with multi-dimensional search.
type IconInterface struct {
	*GalleryInterface
}

// NewIconInterface builds the icon interface.
func NewIconInterface(parent *qt.QWidget) *IconInterface {
	t := gallerycommon.NewTranslator()
	i := &IconInterface{GalleryInterface: NewGalleryInterface(t.Icons, "github.com/famei/gofluent/common", parent)}
	i.SetObjectName("iconInterface")

	// Parent the icon view to the content widget (i.View), matching how
	// AddExampleCard parents cards to g.View.
	iconView := newIconCardView(i.View)
	// Activate the icon view's internal layout before sizing, so its
	// title/search/grid are laid out (not overlapping at y=0).
	if l := iconView.Layout(); l != nil {
		l.Activate()
	}
	iconView.AdjustSize() // ExpandLayout positions by current Height()
	i.VBoxLayout.AddWidget(iconView.QWidget)

	// Fill the whole content area with the icon view (its QFrame view is
	// Expanding). The list view inside it is laid out by its own hBoxLayout, so
	// only the frame itself has to be sized here. Keep it synced on resize.
	fillIconView := func() {
		m := i.VBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
		h := i.View.Height() - m.Top() - m.Bottom()
		if h <= 0 {
			return
		}
		iconView.Resize(i.View.Width(), h)

		headerH := iconView.titleLabel.Height() + iconView.searchWidget.Height() + iconView.vBoxLayout.Spacing()*2
		innerH := iconView.Height() - headerH
		if innerH > 0 {
			iconView.view.Resize(iconView.Width(), innerH)
		}
	}
	fillIconView()
	i.View.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		fillIconView()
	})
	return i
}
