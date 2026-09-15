package widgets

import (
	"fmt"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// InfoLevel enumerates the accent levels of an info badge.
type InfoLevel string

const (
	// InfoLevelInformation is the neutral information level.
	InfoLevelInformation InfoLevel = "Info"
	// InfoLevelSuccess is the success level.
	InfoLevelSuccess InfoLevel = "Success"
	// InfoLevelAttention is the attention (theme color) level.
	InfoLevelAttention InfoLevel = "Attension"
	// InfoLevelWarning is the warning level.
	InfoLevelWarning InfoLevel = "Warning"
	// InfoLevelError is the error level.
	InfoLevelError InfoLevel = "Error"
)

// InfoBadgePosition enumerates where a badge is placed relative to its target.
type InfoBadgePosition int

const (
	InfoBadgePositionTopRight InfoBadgePosition = iota
	InfoBadgePositionBottomRight
	InfoBadgePositionRight
	InfoBadgePositionTopLeft
	InfoBadgePositionBottomLeft
	InfoBadgePositionLeft
	InfoBadgePositionNavigationItem
)

// InfoBadge is a small rounded label used to show a count or status.
type InfoBadge struct {
	*qt.QLabel
	level                InfoLevel
	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor
	target               *qt.QWidget
	position             InfoBadgePosition
}

// NewInfoBadge builds an info badge.
func NewInfoBadge(parent *qt.QWidget, level InfoLevel) *InfoBadge {
	b := &InfoBadge{QLabel: qt.NewQLabel(parent)}
	b.SetLevel(level)
	b.SetObjectName("InfoBadge")
	common.SetFont(b.QWidget, 11, 400)
	// Center the badge text both horizontally and vertically. The QSS only
	// sets symmetric padding, so QLabel's default left alignment would leave
	// single-character badges (which get a circular minimum width) off-center.
	b.SetAlignment(qt.AlignCenter)
	b.SetAttribute(qt.WA_TranslucentBackground)
	b.SetAttribute(qt.WA_TransparentForMouseEvents)
	b.SetSizePolicy2(qt.QSizePolicy__Minimum, qt.QSizePolicy__Fixed)
	common.FluentStyleSheet(common.FluentInfoBadge).Apply(b.QWidget, common.ThemeAuto)

	b.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		color := b.backgroundColor()
		brush := qt.NewQBrush3(color)
		color.Delete()
		defer brush.Delete()
		painter.SetBrush(brush)

		rect := qt.NewQRectF4(0, 0, float64(b.Width()), float64(b.Height()))
		defer rect.Delete()
		r := float64(b.Height()) / 2
		painter.DrawRoundedRect(rect, r, r)
		painter.End()
		super(event)
	})
	return b
}

// NewInfoBadgeText builds an info badge with text.
func NewInfoBadgeText(text string, parent *qt.QWidget, level InfoLevel) *InfoBadge {
	b := NewInfoBadge(parent, level)
	b.SetText(text)
	b.ensureCircularMinimum()
	return b
}

// ensureCircularMinimum makes a single-character badge at least as wide as it
// is tall, so it renders as a perfect circle instead of a narrow vertical
// ellipse (matching the reference rounded-rect geometry whose effective radius
// is min(width, height)/2). Multi-character badges remain wider pills because
// their width already exceeds this minimum.
func (b *InfoBadge) ensureCircularMinimum() {
	h := b.SizeHint().Height() // GoGC-armed — do NOT Delete
	if h > 0 {
		b.SetMinimumWidth(h)
	}
}

// Level returns the badge level.
func (b *InfoBadge) Level() InfoLevel { return b.level }

// SetLevel sets the badge level.
func (b *InfoBadge) SetLevel(level InfoLevel) {
	if level == b.level {
		return
	}
	b.level = level
	b.SetProperty("level", qt.NewQVariant14(string(level)))
	b.Update()
}

// SetCustomBackgroundColor sets the custom background color for light/dark mode.
func (b *InfoBadge) SetCustomBackgroundColor(light, dark *qt.QColor) {
	b.lightBackgroundColor = cloneColor(light)
	b.darkBackgroundColor = cloneColor(dark)
	b.Update()
}

func (b *InfoBadge) backgroundColor() *qt.QColor {
	isDark := common.IsDarkTheme()
	if b.lightBackgroundColor != nil && b.lightBackgroundColor.IsValid() {
		if isDark {
			return cloneColor(b.darkBackgroundColor)
		}
		return cloneColor(b.lightBackgroundColor)
	}

	switch b.level {
	case InfoLevelInformation:
		if isDark {
			return qt.NewQColor3(157, 157, 157)
		}
		return qt.NewQColor3(138, 138, 138)
	case InfoLevelSuccess:
		if isDark {
			return qt.NewQColor3(108, 203, 95)
		}
		return qt.NewQColor3(15, 123, 15)
	case InfoLevelAttention:
		return common.ThemeColorPrimary.Color()
	case InfoLevelWarning:
		if isDark {
			return qt.NewQColor3(255, 244, 206)
		}
		return qt.NewQColor3(157, 93, 0)
	default:
		if isDark {
			return qt.NewQColor3(255, 153, 164)
		}
		return qt.NewQColor3(196, 43, 28)
	}
}

// InfoBadgeMake builds and positions a badge relative to a target widget.
func InfoBadgeMake(text interface{}, parent *qt.QWidget, level InfoLevel, target *qt.QWidget, position InfoBadgePosition) *InfoBadge {
	b := NewInfoBadgeText(fmt.Sprint(text), parent, level)
	b.AdjustSize()
	anchorBadge(b, target, position)
	return b
}

// DotInfoBadge is a small circular badge with no text.
type DotInfoBadge struct {
	*InfoBadge
}

// NewDotInfoBadge builds a dot info badge.
func NewDotInfoBadge(parent *qt.QWidget, level InfoLevel) *DotInfoBadge {
	b := &DotInfoBadge{InfoBadge: NewInfoBadge(parent, level)}
	b.SetFixedSize2(4, 4)
	b.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		color := b.backgroundColor()
		brush := qt.NewQBrush3(color)
		color.Delete()
		defer brush.Delete()
		painter.SetBrush(brush)

		rect := qt.NewQRectF4(0, 0, float64(b.Width()), float64(b.Height()))
		defer rect.Delete()
		painter.DrawEllipse(rect)
		painter.End()
	})
	return b
}

// DotInfoBadgeMake builds and positions a dot badge relative to a target widget.
func DotInfoBadgeMake(parent *qt.QWidget, level InfoLevel, target *qt.QWidget, position InfoBadgePosition) *DotInfoBadge {
	b := NewDotInfoBadge(parent, level)
	anchorBadge(b.InfoBadge, target, position)
	return b
}

// IconInfoBadge is a small circular badge with an icon.
type IconInfoBadge struct {
	*InfoBadge
	iconSource interface{}
	iconSize   *qt.QSize
}

// NewIconInfoBadge builds an icon info badge.
func NewIconInfoBadge(parent *qt.QWidget, level InfoLevel) *IconInfoBadge {
	b := &IconInfoBadge{InfoBadge: NewInfoBadge(parent, level), iconSize: qt.NewQSize2(8, 8)}
	b.SetFixedSize2(16, 16)
	b.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		color := b.backgroundColor()
		brush := qt.NewQBrush3(color)
		color.Delete()
		defer brush.Delete()
		painter.SetBrush(brush)

		bgRect := qt.NewQRectF4(0, 0, float64(b.Width()), float64(b.Height()))
		defer bgRect.Delete()
		painter.DrawEllipse(bgRect)

		iw, ih := b.iconSize.Width(), b.iconSize.Height()
		iconRect := qt.NewQRectF4(float64(b.Width()-iw)/2, float64(b.Height()-ih)/2, float64(iw), float64(ih))
		defer iconRect.Delete()

		if b.iconSource != nil {
			renderFluentIcon(b.iconSource, painter, iconRect, common.ThemeAuto)
		}
		painter.End()
	})
	return b
}

// NewIconInfoBadgeIcon builds an icon info badge from an icon source.
func NewIconInfoBadgeIcon(icon interface{}, parent *qt.QWidget, level InfoLevel) *IconInfoBadge {
	b := NewIconInfoBadge(parent, level)
	b.SetIcon(icon)
	return b
}

// SetIcon sets the badge icon.
func (b *IconInfoBadge) SetIcon(icon interface{}) {
	b.iconSource = icon
	b.Update()
}

// Icon returns the badge icon as a QIcon.
func (b *IconInfoBadge) Icon() *qt.QIcon { return common.ToQIcon(b.iconSource) }

// IconSize returns the icon size.
func (b *IconInfoBadge) IconSize() *qt.QSize { return b.iconSize }

// SetIconSize sets the icon size.
func (b *IconInfoBadge) SetIconSize(size *qt.QSize) {
	b.iconSize = size
	b.Update()
}

// IconInfoBadgeMake builds and positions an icon badge relative to a target.
func IconInfoBadgeMake(icon interface{}, parent *qt.QWidget, level InfoLevel, target *qt.QWidget, position InfoBadgePosition) *IconInfoBadge {
	b := NewIconInfoBadgeIcon(icon, parent, level)
	anchorBadge(b.InfoBadge, target, position)
	return b
}

// anchorBadge attaches the badge to target and positions it relative to the
// target. The reference InfoBadgeManager installs an event filter on the target
// for live repositioning, but that crashes in miqt (InstallEventFilter uses the
// promoted QObject while OnEventFilter overrides the directly-constructed
// QLabel). Instead, the badge is positioned once the target has been laid out,
// via a 0ms single-shot timer that runs on the next event-loop iteration.
func anchorBadge(b *InfoBadge, target *qt.QWidget, position InfoBadgePosition) {
	if target == nil {
		return
	}
	b.target = target
	b.position = position

	timer := qt.NewQTimer2(b.QObject)
	timer.SetSingleShot(true)
	timer.OnTimeout(func() {
		b.reposition()
		timer.Delete()
	})
	timer.Start(0)
}

// reposition moves the badge to the current position of its target.
func (b *InfoBadge) reposition() {
	if b.target == nil {
		return
	}
	x, y := badgePosFor(b.position, b.target, b.QWidget)
	b.Move(x, y)
}

// badgePosFor computes the top-left position of badge relative to target for
// the given position (the InfoBadgeManager positioning, inlined).
func badgePosFor(position InfoBadgePosition, target *qt.QWidget, badge *qt.QWidget) (int, int) {
	g := target.Geometry() // borrowed (QWidget.Geometry const_cast) — do NOT Delete

	switch position {
	case InfoBadgePositionTopRight:
		p := g.TopRight() // GoGC-armed — do NOT Delete
		return p.X() - badge.Width()/2, p.Y() - badge.Height()/2
	case InfoBadgePositionRight:
		c := g.Center() // GoGC-armed — do NOT Delete
		return g.Right() - badge.Width()/2, c.Y() - badge.Height()/2
	case InfoBadgePositionBottomRight:
		p := g.BottomRight() // GoGC-armed — do NOT Delete
		return p.X() - badge.Width()/2, p.Y() - badge.Height()/2
	case InfoBadgePositionTopLeft:
		return target.X() - badge.Width()/2, target.Y() - badge.Height()/2
	case InfoBadgePositionLeft:
		c := g.Center() // GoGC-armed — do NOT Delete
		return target.X() - badge.Width()/2, c.Y() - badge.Height()/2
	case InfoBadgePositionBottomLeft:
		p := g.BottomLeft() // GoGC-armed — do NOT Delete
		return p.X() - badge.Width()/2, p.Y() - badge.Height()/2
	case InfoBadgePositionNavigationItem:
		// Mirrors the reference NavigationItemInfoBadgeManager.position(): a
		// compacted navigation item pins the badge to its top-right corner,
		// while an expanded item places it near the right edge. Compacted state
		// is detected from the item width (40px compacted vs 312px expanded);
		// the leaf/non-leaf dx distinction (10 vs 35) isn't available in this
		// package, so the leaf offset is used.
		if target.Width() <= 48 {
			return g.Right() - badge.Width() - 2, g.Top() + 2
		}
		return g.Right() - badge.Width() - 10, g.Top() + 18 - badge.Height()/2
	default:
		return 0, 0
	}
}
