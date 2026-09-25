package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// PixmapLabel is a QLabel that renders a high-dpi pixmap with antialiasing.
type PixmapLabel struct {
	*qt.QLabel
	pixmap *qt.QPixmap
}

// NewPixmapLabel builds a pixmap label.
func NewPixmapLabel(parent *qt.QWidget) *PixmapLabel {
	w := &PixmapLabel{QLabel: qt.NewQLabel(parent), pixmap: qt.NewQPixmap()}
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		if w.pixmap.IsNull() {
			super(event)
			return
		}
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)
		painter.SetPenWithStyle(qt.NoPen)
		painter.DrawPixmap10(w.Rect(), w.pixmap)
		painter.End()
	})
	return w
}

// SetPixmap sets the pixmap and fixes the label size to match.
func (w *PixmapLabel) SetPixmap(pixmap *qt.QPixmap) {
	w.pixmap = pixmap
	size := pixmap.Size()

	w.SetFixedSize(size)
	w.Update()
}

// Pixmap returns the current pixmap.
func (w *PixmapLabel) Pixmap() *qt.QPixmap { return w.pixmap }

// FluentLabelBase is the base class of all fluent labels. It applies the
// label stylesheet, the fluent font and theme-aware text color.
type FluentLabelBase struct {
	*qt.QLabel
	lightColor *qt.QColor
	darkColor  *qt.QColor
}

func newFluentLabelBase(parent *qt.QWidget, fontSize, weight int) *FluentLabelBase {
	w := &FluentLabelBase{QLabel: qt.NewQLabel(parent)}
	common.FluentStyleSheet(common.FluentLabel).Apply(w.QWidget, common.ThemeAuto)
	common.SetFont(w.QWidget, fontSize, weight)
	w.lightColor = qt.NewQColor3(0, 0, 0)
	w.darkColor = qt.NewQColor3(255, 255, 255)
	w.SetTextColor(w.lightColor, w.darkColor)
	// The theme listener is tied to the label: the registry drops it when the
	// label is destroyed, so a theme switch can never call into freed widget
	// memory (a Go closure is not a Qt slot and would not be disconnected).
	common.QConfigInstance.OnThemeChangedFor(w.QObject, func(common.Theme) {
		w.SetTextColor(w.lightColor, w.darkColor)
	})
	w.OnContextMenuEvent(func(super func(event *qt.QContextMenuEvent), event *qt.QContextMenuEvent) {
		menu := NewLabelContextMenu(w.QLabel)
		globalPos := w.MapToGlobal(event.Pos())
		menu.execAt(globalPos)
	})
	return w
}

// SetText sets the label text.
func (w *FluentLabelBase) SetText(text string) { w.QLabel.SetText(text) }

// Text returns the label text.
func (w *FluentLabelBase) Text() string { return w.QLabel.Text() }

// SetTextColor sets the light/dark text colors and applies a custom stylesheet.
func (w *FluentLabelBase) SetTextColor(light, dark *qt.QColor) {
	w.lightColor = cloneColor(light)
	w.darkColor = cloneColor(dark)
	common.SetCustomStyleSheet(w.QWidget,
		"FluentLabelBase{color:"+w.lightColor.NameWithFormat(qt.QColor__HexArgb)+"}",
		"FluentLabelBase{color:"+w.darkColor.NameWithFormat(qt.QColor__HexArgb)+"}")
}

// LightColor returns the light text color.
func (w *FluentLabelBase) LightColor() *qt.QColor { return w.lightColor }

// DarkColor returns the dark text color.
func (w *FluentLabelBase) DarkColor() *qt.QColor { return w.darkColor }

// PixelFontSize returns the current font pixel size.
func (w *FluentLabelBase) PixelFontSize() int { return w.Font().PixelSize() }

// SetPixelFontSize sets the font pixel size.
func (w *FluentLabelBase) SetPixelFontSize(size int) {
	font := w.Font()
	font.SetPixelSize(size)
	w.SetFont(font)
}

// IsStrikeOut reports whether the font is struck out.
func (w *FluentLabelBase) IsStrikeOut() bool { return w.Font().StrikeOut() }

// SetStrikeOut sets the strike-out style.
func (w *FluentLabelBase) SetStrikeOut(isStrikeOut bool) {
	font := w.Font()
	font.SetStrikeOut(isStrikeOut)
	w.SetFont(font)
}

// IsUnderline reports whether the font is underlined.
func (w *FluentLabelBase) IsUnderline() bool { return w.Font().Underline() }

// SetUnderline sets the underline style.
func (w *FluentLabelBase) SetUnderline(isUnderline bool) {
	font := w.Font()
	font.SetUnderline(isUnderline)
	w.SetFont(font)
}

// CaptionLabel is a caption text label (12px).
type CaptionLabel struct{ *FluentLabelBase }

// NewCaptionLabel builds a caption label.
func NewCaptionLabel(parent *qt.QWidget) *CaptionLabel {
	return &CaptionLabel{FluentLabelBase: newFluentLabelBase(parent, 12, int(qt.QFont__Normal))}
}

// NewCaptionLabelText builds a caption label with text.
func NewCaptionLabelText(text string, parent *qt.QWidget) *CaptionLabel {
	w := NewCaptionLabel(parent)
	w.SetText(text)
	return w
}

// BodyLabel is a body text label (14px).
type BodyLabel struct{ *FluentLabelBase }

// NewBodyLabel builds a body label.
func NewBodyLabel(parent *qt.QWidget) *BodyLabel {
	return &BodyLabel{FluentLabelBase: newFluentLabelBase(parent, 14, int(qt.QFont__Normal))}
}

// NewBodyLabelText builds a body label with text.
func NewBodyLabelText(text string, parent *qt.QWidget) *BodyLabel {
	w := NewBodyLabel(parent)
	w.SetText(text)
	return w
}

// StrongBodyLabel is a strong body text label (14px DemiBold).
type StrongBodyLabel struct{ *FluentLabelBase }

// NewStrongBodyLabel builds a strong body label.
func NewStrongBodyLabel(parent *qt.QWidget) *StrongBodyLabel {
	return &StrongBodyLabel{FluentLabelBase: newFluentLabelBase(parent, 14, int(qt.QFont__DemiBold))}
}

// NewStrongBodyLabelText builds a strong body label with text.
func NewStrongBodyLabelText(text string, parent *qt.QWidget) *StrongBodyLabel {
	w := NewStrongBodyLabel(parent)
	w.SetText(text)
	return w
}

// SubtitleLabel is a subtitle text label (20px DemiBold).
type SubtitleLabel struct{ *FluentLabelBase }

// NewSubtitleLabel builds a subtitle label.
func NewSubtitleLabel(parent *qt.QWidget) *SubtitleLabel {
	return &SubtitleLabel{FluentLabelBase: newFluentLabelBase(parent, 20, int(qt.QFont__DemiBold))}
}

// NewSubtitleLabelText builds a subtitle label with text.
func NewSubtitleLabelText(text string, parent *qt.QWidget) *SubtitleLabel {
	w := NewSubtitleLabel(parent)
	w.SetText(text)
	return w
}

// TitleLabel is a title text label (28px DemiBold).
type TitleLabel struct{ *FluentLabelBase }

// NewTitleLabel builds a title label.
func NewTitleLabel(parent *qt.QWidget) *TitleLabel {
	return &TitleLabel{FluentLabelBase: newFluentLabelBase(parent, 28, int(qt.QFont__DemiBold))}
}

// NewTitleLabelText builds a title label with text.
func NewTitleLabelText(text string, parent *qt.QWidget) *TitleLabel {
	w := NewTitleLabel(parent)
	w.SetText(text)
	return w
}

// LargeTitleLabel is a large title text label (40px DemiBold).
type LargeTitleLabel struct{ *FluentLabelBase }

// NewLargeTitleLabel builds a large title label.
func NewLargeTitleLabel(parent *qt.QWidget) *LargeTitleLabel {
	return &LargeTitleLabel{FluentLabelBase: newFluentLabelBase(parent, 40, int(qt.QFont__DemiBold))}
}

// NewLargeTitleLabelText builds a large title label with text.
func NewLargeTitleLabelText(text string, parent *qt.QWidget) *LargeTitleLabel {
	w := NewLargeTitleLabel(parent)
	w.SetText(text)
	return w
}

// DisplayLabel is a display text label (68px DemiBold).
type DisplayLabel struct{ *FluentLabelBase }

// NewDisplayLabel builds a display label.
func NewDisplayLabel(parent *qt.QWidget) *DisplayLabel {
	return &DisplayLabel{FluentLabelBase: newFluentLabelBase(parent, 68, int(qt.QFont__DemiBold))}
}

// NewDisplayLabelText builds a display label with text.
func NewDisplayLabelText(text string, parent *qt.QWidget) *DisplayLabel {
	w := NewDisplayLabel(parent)
	w.SetText(text)
	return w
}

// ImageLabel is a QLabel that renders an image with independently rounded
// corners and emits an onClicked callback.
type ImageLabel struct {
	*qt.QLabel
	image             *qt.QImage
	topLeftRadius     int
	topRightRadius    int
	bottomLeftRadius  int
	bottomRightRadius int
	onClicked         func()
}

// NewImageLabel builds an empty image label.
func NewImageLabel(parent *qt.QWidget) *ImageLabel {
	w := &ImageLabel{QLabel: qt.NewQLabel(parent), image: qt.NewQImage()}
	w.installEvents()
	return w
}

// NewImageLabelImage builds an image label from a string path, QImage or QPixmap.
func NewImageLabelImage(image interface{}, parent *qt.QWidget) *ImageLabel {
	w := NewImageLabel(parent)
	// NewImageLabel allocated an owned empty QImage (qt.NewQImage) that is only
	// a placeholder; free it before SetImage replaces w.image so it does not
	// leak. It is owned (no finalizer), so Delete() is safe here.
	w.image.Delete()
	w.SetImage(image)
	return w
}

func (w *ImageLabel) installEvents() {
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		if w.onClicked != nil {
			w.onClicked()
		}
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		if w.image.IsNull() {
			return
		}
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		path := qt.NewQPainterPath()
		defer path.Delete()
		wbuild, hbuild := w.Width(), w.Height()
		buildRoundedPath(path, w.topLeftRadius, w.topRightRadius, w.bottomLeftRadius, w.bottomRightRadius, wbuild, hbuild)

		scaled := w.image.Scaled3(wbuild, hbuild, qt.IgnoreAspectRatio, qt.SmoothTransformation) // GoGC-armed — do NOT Delete

		painter.SetPenWithStyle(qt.NoPen)
		painter.SetClipPath(path)
		painter.DrawImage6(w.Rect(), scaled)
		painter.End()
	})
}

// OnClicked registers a callback invoked on mouse release.
func (w *ImageLabel) OnClicked(f func()) { w.onClicked = f }

// SetBorderRadius sets the four corner radii.
func (w *ImageLabel) SetBorderRadius(topLeft, topRight, bottomLeft, bottomRight int) {
	w.topLeftRadius = topLeft
	w.topRightRadius = topRight
	w.bottomLeftRadius = bottomLeft
	w.bottomRightRadius = bottomRight
	w.Update()
}

// SetImage sets the image from a string path, *qt.QImage or *qt.QPixmap.
func (w *ImageLabel) SetImage(image interface{}) {
	switch v := image.(type) {
	case nil:
		w.image = qt.NewQImage()
	case string:
		w.image = qt.NewQImage8(v)
	case *qt.QImage:
		w.image = v
	case *qt.QPixmap:
		w.image = v.ToImage()
	}
	size := w.image.Size()

	w.SetFixedSize(size)
	w.Update()
}

// Image returns the underlying image.
func (w *ImageLabel) Image() *qt.QImage { return w.image }

// IsNull reports whether the image is null.
func (w *ImageLabel) IsNull() bool { return w.image.IsNull() }

// ScaledToWidth scales the label to the given width preserving aspect ratio.
func (w *ImageLabel) ScaledToWidth(width int) {
	if w.IsNull() {
		return
	}
	h := int(float64(width) / float64(w.image.Width()) * float64(w.image.Height()))
	w.SetFixedSize2(width, h)
}

// ScaledToHeight scales the label to the given height preserving aspect ratio.
func (w *ImageLabel) ScaledToHeight(height int) {
	if w.IsNull() {
		return
	}
	wd := int(float64(height) / float64(w.image.Height()) * float64(w.image.Width()))
	w.SetFixedSize2(wd, height)
}

// SetScaledSize fixes the label size.
func (w *ImageLabel) SetScaledSize(size *qt.QSize) {
	if w.IsNull() {
		return
	}
	w.SetFixedSize(size)
}

// SetPixmap sets the image from a pixmap.
func (w *ImageLabel) SetPixmap(pixmap *qt.QPixmap) { w.SetImage(pixmap) }

// Pixmap returns the image as a pixmap.
func (w *ImageLabel) Pixmap() *qt.QPixmap { return qt.QPixmap_FromImage(w.image) }

// TopLeftRadius returns the top-left corner radius.
func (w *ImageLabel) TopLeftRadius() int { return w.topLeftRadius }

// TopRightRadius returns the top-right corner radius.
func (w *ImageLabel) TopRightRadius() int { return w.topRightRadius }

// BottomLeftRadius returns the bottom-left corner radius.
func (w *ImageLabel) BottomLeftRadius() int { return w.bottomLeftRadius }

// BottomRightRadius returns the bottom-right corner radius.
func (w *ImageLabel) BottomRightRadius() int { return w.bottomRightRadius }

// buildRoundedPath constructs a rounded rectangle path with independent radii.
func buildRoundedPath(path *qt.QPainterPath, tl, tr, bl, br, w, h int) {
	path.MoveTo2(float64(tl), 0)
	path.LineTo2(float64(w-tr), 0)
	d := float64(tr * 2)
	path.ArcTo2(float64(w)-d, 0, d, d, 90, -90)
	path.LineTo2(float64(w), float64(h-br))
	d = float64(br * 2)
	path.ArcTo2(float64(w)-d, float64(h)-d, d, d, 0, -90)
	path.LineTo2(float64(bl), float64(h))
	d = float64(bl * 2)
	path.ArcTo2(0, float64(h)-d, d, d, -90, -90)
	path.LineTo2(0, float64(tl))
	d = float64(tl * 2)
	path.ArcTo2(0, 0, d, d, -180, -90)
}

// AvatarWidget is an image label with a circular avatar (image or text).
type AvatarWidget struct {
	*ImageLabel
	radius               int
	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor
}

// NewAvatarWidget builds an avatar widget.
func NewAvatarWidget(parent *qt.QWidget) *AvatarWidget {
	w := &AvatarWidget{ImageLabel: NewImageLabel(parent)}
	w.SetRadius(48)
	w.lightBackgroundColor = qt.NewQColor11(0, 0, 0, 50)
	w.darkBackgroundColor = qt.NewQColor11(255, 255, 255, 50)
	w.installAvatarPaint()
	return w
}

// NewAvatarWidgetImage builds an avatar widget with an image.
func NewAvatarWidgetImage(image interface{}, parent *qt.QWidget) *AvatarWidget {
	w := NewAvatarWidget(parent)
	w.SetImage(image)
	return w
}

// Radius returns the avatar radius.
func (w *AvatarWidget) Radius() int { return w.radius }

// SetRadius sets the avatar radius (and fixed size).
func (w *AvatarWidget) SetRadius(radius int) {
	w.radius = radius
	common.SetFont(w.QWidget, radius, int(qt.QFont__Normal))
	w.SetFixedSize2(2*radius, 2*radius)
	w.Update()
}

// SetImage overrides to re-apply the radius after loading.
func (w *AvatarWidget) SetImage(image interface{}) {
	w.ImageLabel.SetImage(image)
	w.SetRadius(w.radius)
}

// SetBackgroundColor sets the light/dark text-avatar background colors.
func (w *AvatarWidget) SetBackgroundColor(light, dark *qt.QColor) {
	w.lightBackgroundColor = cloneColor(light)
	w.darkBackgroundColor = cloneColor(dark)
	w.Update()
}

func (w *AvatarWidget) installAvatarPaint() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		if !w.IsNull() {
			w.drawImageAvatar(painter)
		} else {
			w.drawTextAvatar(painter)
		}
		painter.End()
	})
}

func (w *AvatarWidget) drawImageAvatar(painter *qt.QPainter) {
	image := w.image.Scaled3(w.Width(), w.Height(), qt.KeepAspectRatioByExpanding, qt.SmoothTransformation) // GoGC-armed — do NOT Delete
	iw, ih := image.Width(), image.Height()
	d := w.radius * 2
	x := (iw - d) / 2
	y := (ih - d) / 2
	cropped := image.Copy2(x, y, d, d) // GoGC-armed — do NOT Delete

	path := qt.NewQPainterPath()
	defer path.Delete()
	rect := qt.NewQRectF5(w.Rect())
	defer rect.Delete()
	path.AddEllipse(rect)

	painter.SetPenWithStyle(qt.NoPen)
	painter.SetClipPath(path)
	painter.DrawImage6(w.Rect(), cropped)
}

func (w *AvatarWidget) drawTextAvatar(painter *qt.QPainter) {
	text := w.Text()
	if text == "" {
		return
	}
	brush := qt.NewQBrush3(w.backgroundForTheme())
	defer brush.Delete()
	painter.SetBrush(brush)
	painter.SetPenWithStyle(qt.NoPen)
	rect := qt.NewQRectF5(w.Rect())
	defer rect.Delete()
	painter.DrawEllipse(rect)

	painter.SetFont(w.Font())
	if common.IsDarkTheme() {
		painter.SetPen(qt.NewQColor3(255, 255, 255))
	} else {
		painter.SetPen(qt.NewQColor3(0, 0, 0))
	}
	first := string([]rune(text)[0])
	painter.DrawText6(w.Rect(), int(qt.AlignCenter), string(unicodeUpper(first)))
}

func (w *AvatarWidget) backgroundForTheme() *qt.QColor {
	if common.IsDarkTheme() {
		return w.darkBackgroundColor
	}
	return w.lightBackgroundColor
}

func unicodeUpper(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
}

// HyperlinkLabel is a label-like button that opens a URL when clicked.
type HyperlinkLabel struct {
	*qt.QPushButton
	url                *qt.QUrl
	isUnderlineVisible bool
}

// NewHyperlinkLabel builds a hyperlink label.
func NewHyperlinkLabel(parent *qt.QWidget) *HyperlinkLabel {
	w := &HyperlinkLabel{QPushButton: qt.NewQPushButton(parent), url: qt.NewQUrl()}
	w.SetObjectName("hyperlinkLabel")
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	w.SetUnderlineVisible(false)
	common.FluentStyleSheet(common.FluentLabel).Apply(w.QWidget, common.ThemeAuto)
	w.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))
	w.OnClicked(w.onClicked)
	return w
}

// NewHyperlinkLabelText builds a hyperlink label with text.
func NewHyperlinkLabelText(text string, parent *qt.QWidget) *HyperlinkLabel {
	w := NewHyperlinkLabel(parent)
	w.SetText(text)
	return w
}

// NewHyperlinkLabelURL builds a hyperlink label with url and text.
func NewHyperlinkLabelURL(url, text string, parent *qt.QWidget) *HyperlinkLabel {
	w := NewHyperlinkLabel(parent)
	w.SetText(text)
	w.SetUrl(url)
	return w
}

// Url returns the target URL.
func (w *HyperlinkLabel) Url() *qt.QUrl { return w.url }

// SetUrl sets the target URL from a string.
func (w *HyperlinkLabel) SetUrl(url string) { w.url = qt.NewQUrl3(url) }

// IsUnderlineVisible reports whether the underline is visible.
func (w *HyperlinkLabel) IsUnderlineVisible() bool { return w.isUnderlineVisible }

// SetUnderlineVisible toggles the underline style.
func (w *HyperlinkLabel) SetUnderlineVisible(isVisible bool) {
	w.isUnderlineVisible = isVisible
	w.SetProperty("underline", qt.NewQVariant11(isVisible))
	w.SetStyle(qt.QApplication_Style())
}

func (w *HyperlinkLabel) onClicked() {
	if w.url.IsValid() {
		qt.QDesktopServices_OpenUrl(w.url)
	}
}
