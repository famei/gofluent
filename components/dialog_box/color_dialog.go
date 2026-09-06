package dialog_box

import (
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/resources"
	qt "github.com/mappu/miqt/qt"
)

// loadPixmap reads an embedded image asset into a QPixmap.
func loadPixmap(path string) *qt.QPixmap {
	data, err := fs.ReadFile(resources.Images, path)
	if err != nil {
		return qt.NewQPixmap()
	}
	pm := qt.NewQPixmap()
	pm.LoadFromDataWithData(data)
	return pm
}

// cloneColor copies a QColor.
func cloneColor(c *qt.QColor) *qt.QColor {
	if c == nil {
		return qt.NewQColor6("#000000")
	}
	return qt.NewQColor9(c)
}

// colorEquals compares two colors by their #AARRGGBB name.
func colorEquals(a, b *qt.QColor) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.NameWithFormat(qt.QColor__HexArgb) == b.NameWithFormat(qt.QColor__HexArgb)
}

// HuePanel is the 256x256 hue/saturation picker.
type HuePanel struct {
	*qt.QWidget
	color      *qt.QColor
	pickerPosX int
	pickerPosY int
	huePixmap  *qt.QPixmap

	OnColorChanged func(color *qt.QColor)
}

// NewHuePanel builds a hue panel.
func NewHuePanel(color *qt.QColor, parent *qt.QWidget) *HuePanel {
	p := &HuePanel{QWidget: qt.NewQWidget(parent)}
	p.SetFixedSize2(256, 256)
	p.huePixmap = loadPixmap("images/color_dialog/HuePanel.png")
	p.SetColor(color)
	p.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		pos := e.Pos()
		p.setPickerPosition(pos)

	})
	p.OnMouseMoveEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		pos := e.Pos()
		p.setPickerPosition(pos)

	})
	p.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		p.paint()
	})
	return p
}

func (p *HuePanel) setPickerPosition(pos *qt.QPoint) {
	p.pickerPosX = pos.X()
	p.pickerPosY = pos.Y()

	h := int(clamp01(float64(pos.X())/float64(p.Width())) * 359)
	s := int(clamp01(float64(p.Height()-pos.Y())/float64(p.Height())) * 255)
	p.color.SetHsv(h, s, 255)
	p.Update()
	if p.OnColorChanged != nil {
		p.OnColorChanged(p.color)
	}
}

// SetColor sets the current color and repositions the picker.
func (p *HuePanel) SetColor(color *qt.QColor) {
	p.color = cloneColor(color)
	p.color.SetHsv(p.color.Hue(), p.color.Saturation(), 255)
	p.pickerPosX = int(float64(p.color.Hue()) / 359 * float64(p.Width()))
	p.pickerPosY = int(float64(255-p.color.Saturation()) / 255 * float64(p.Height()))
	p.Update()
}

func (p *HuePanel) paint() {
	painter := qt.NewQPainter2(p.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

	brush := qt.NewQBrush7(p.huePixmap)
	painter.SetBrush(brush)
	brush.Delete()

	penColor := qt.NewQColor11(0, 0, 0, 15)
	pen := qt.NewQPen3(penColor)
	pen.SetWidthF(2.4)
	painter.SetPenWithPen(pen)
	pen.Delete()
	painter.DrawRoundedRect3(p.Rect(), 5.6, 5.6)

	var pickerColor *qt.QColor
	if p.color.Saturation() > 153 || (40 < p.color.Hue() && p.color.Hue() < 180) {
		pickerColor = qt.NewQColor2(qt.Black)
	} else {
		pickerColor = qt.NewQColor3(255, 253, 254)
	}
	pen = qt.NewQPen3(pickerColor)
	pen.SetWidth(3)
	painter.SetPenWithPen(pen)
	pen.Delete()
	painter.SetBrushWithStyle(qt.NoBrush)
	painter.DrawEllipse2(p.pickerPosX-8, p.pickerPosY-8, 16, 16)
	painter.End()
}

// BrightnessSlider is a clickable slider that controls the color value.
type BrightnessSlider struct {
	*widgets.ClickableSlider
	color *qt.QColor

	OnColorChanged func(color *qt.QColor)
}

// NewBrightnessSlider builds a brightness slider.
func NewBrightnessSlider(color *qt.QColor, parent *qt.QWidget) *BrightnessSlider {
	s := &BrightnessSlider{ClickableSlider: widgets.NewClickableSlider(parent)}
	// Match the reference BrightnessSlider(Qt.Horizontal, parent): the middle
	// "picker scrollbar" of the color dialog must be a horizontal Fluent gradient
	// slider. The QSS only defines QSlider:horizontal / ::groove:horizontal /
	// ::handle:horizontal rules, so if the orientation is not horizontal the
	// slider falls back to a native (non-Fluent) vertical rendering. Enforce the
	// orientation and the 332x24 geometry implied by the color_dialog.qss
	// min-width/min-height rules so it cannot regress to a narrow vertical bar.
	s.SetOrientation(qt.Horizontal)
	s.SetFixedSize2(332, 24)
	s.SetRange(0, 255)
	s.SetSingleStep(1)
	s.SetColor(color)
	s.OnValueChanged(s.onValueChanged)
	return s
}

// SetColor stores the color and refreshes the slider stylesheet.
func (s *BrightnessSlider) SetColor(color *qt.QColor) {
	s.color = cloneColor(color)
	s.SetValue(s.color.Value())
	qss := common.FluentStyleSheet(common.FluentColorDialog).Content(common.ThemeAuto)
	qss = strings.ReplaceAll(qss, "--slider-hue", strconv.Itoa(s.color.Hue()))
	qss = strings.ReplaceAll(qss, "--slider-saturation", strconv.Itoa(s.color.Saturation()))
	s.SetStyleSheet(qss)
}

func (s *BrightnessSlider) onValueChanged(value int) {
	s.color.SetHsv2(s.color.Hue(), s.color.Saturation(), value, s.color.Alpha())
	s.SetColor(s.color)
	if s.OnColorChanged != nil {
		s.OnColorChanged(s.color)
	}
}

// ColorCard draws a 44x128 color swatch (with a tiled background when alpha is
// enabled).
type ColorCard struct {
	*qt.QWidget
	color       *qt.QColor
	enableAlpha bool
	tiledPixmap *qt.QPixmap
}

// NewColorCard builds a color card.
func NewColorCard(color *qt.QColor, parent *qt.QWidget, enableAlpha bool) *ColorCard {
	c := &ColorCard{QWidget: qt.NewQWidget(parent), enableAlpha: enableAlpha}
	c.SetFixedSize2(44, 128)
	c.SetColor(color)
	c.tiledPixmap = c.createTiledBackground()
	c.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		c.paint()
	})
	return c
}

func (c *ColorCard) createTiledBackground() *qt.QPixmap {
	pm := qt.NewQPixmap2(8, 8)
	transparent := qt.NewQColor2(qt.Transparent)
	pm.FillWithFillColor(transparent)
	transparent.Delete()

	painter := qt.NewQPainter2(pm.QPaintDevice)
	r := 255
	if !common.IsDarkTheme() {
		r = 0
	}
	color := qt.NewQColor11(r, r, r, 26)
	painter.FillRect5(4, 0, 4, 4, color)
	painter.FillRect5(0, 4, 4, 4, color)
	painter.End()
	painter.Delete()
	return pm
}

// SetColor stores the swatch color.
func (c *ColorCard) SetColor(color *qt.QColor) {
	c.color = cloneColor(color)
	c.Update()
}

func (c *ColorCard) paint() {
	painter := qt.NewQPainter2(c.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	if c.enableAlpha {
		brush := qt.NewQBrush7(c.tiledPixmap)
		painter.SetBrush(brush)
		brush.Delete()
		painter.SetPen(qt.NewQColor11(0, 0, 0, 13))
		painter.DrawRoundedRect3(c.Rect(), 4, 4)
	}

	brush := qt.NewQBrush3(c.color)
	painter.SetBrush(brush)
	brush.Delete()
	painter.SetPen(qt.NewQColor11(0, 0, 0, 13))
	painter.DrawRoundedRect3(c.Rect(), 4, 4)
	painter.End()
}

// ColorLineEdit is a line edit that only emits OnValueChanged for integer
// values in [0, 255] (the Go equivalent of the QIntValidator).
type ColorLineEdit struct {
	*widgets.LineEdit
	validateFunc func(text string) bool

	OnValueChanged func(text string)
}

func newColorLineEdit(text string, parent *qt.QWidget) *ColorLineEdit {
	e := &ColorLineEdit{LineEdit: widgets.NewLineEdit(parent)}
	e.SetText(text)
	e.SetFixedSize2(136, 33)
	e.SetClearButtonEnabled(true)
	e.SetValidator(qt.NewQIntValidator2(0, 255).QValidator)
	e.validateFunc = func(t string) bool {
		v, err := strconv.Atoi(t)
		return err == nil && v >= 0 && v <= 255
	}
	e.OnTextEdited(e.onTextEdited)
	return e
}

// NewColorLineEdit builds a color channel line edit.
func NewColorLineEdit(value int, parent *qt.QWidget) *ColorLineEdit {
	return newColorLineEdit(strconv.Itoa(value), parent)
}

func (e *ColorLineEdit) onTextEdited(text string) {
	if e.validateFunc != nil && e.validateFunc(text) {
		if e.OnValueChanged != nil {
			e.OnValueChanged(text)
		}
	}
}

// HexColorLineEdit edits a 6 or 8 digit hex color (without the leading '#').
type HexColorLineEdit struct {
	*ColorLineEdit
	enableAlpha bool
	prefixLabel *qt.QLabel
	hexRe       *regexp.Regexp
}

// NewHexColorLineEdit builds a hex color line edit.
func NewHexColorLineEdit(color *qt.QColor, parent *qt.QWidget, enableAlpha bool) *HexColorLineEdit {
	var name string
	if enableAlpha {
		name = color.NameWithFormat(qt.QColor__HexArgb)
	} else {
		name = color.NameWithFormat(qt.QColor__HexRgb)
	}
	name = strings.TrimPrefix(name, "#")

	e := &HexColorLineEdit{ColorLineEdit: newColorLineEdit(name, parent), enableAlpha: enableAlpha}
	if enableAlpha {
		e.hexRe = regexp.MustCompile(`^[A-Fa-f0-9]{8}$`)
	} else {
		e.hexRe = regexp.MustCompile(`^[A-Fa-f0-9]{6}$`)
	}
	e.validateFunc = func(t string) bool { return e.hexRe.MatchString(t) }

	e.SetTextMargins(4, 0, 33, 0)
	e.prefixLabel = qt.NewQLabel5("#", e.QWidget)
	e.prefixLabel.Move(7, 2)
	e.prefixLabel.SetObjectName("prefixLabel")
	return e
}

// SetColor updates the hex text from a color.
func (e *HexColorLineEdit) SetColor(color *qt.QColor) {
	var name string
	if e.enableAlpha {
		name = color.NameWithFormat(qt.QColor__HexArgb)
	} else {
		name = color.NameWithFormat(qt.QColor__HexRgb)
	}
	e.SetText(strings.TrimPrefix(name, "#"))
}

// OpacityLineEdit edits a 0..100 percentage.
type OpacityLineEdit struct {
	*ColorLineEdit
	suffixLabel *qt.QLabel
}

// NewOpacityLineEdit builds an opacity line edit.
func NewOpacityLineEdit(value int, parent *qt.QWidget) *OpacityLineEdit {
	e := &OpacityLineEdit{ColorLineEdit: NewColorLineEdit(int(float64(value)/255*100), parent)}
	e.validateFunc = func(t string) bool {
		return regexp.MustCompile(`^[0-9][0-9]{0,1}$|^100$`).MatchString(t)
	}
	e.SetValidator(qt.NewQIntValidator2(0, 100).QValidator)
	e.SetTextMargins(4, 0, 33, 0)
	e.suffixLabel = qt.NewQLabel5("%", e.QWidget)
	e.suffixLabel.SetObjectName("suffixLabel")
	e.OnShowEvent(func(super func(event *qt.QShowEvent), event *qt.QShowEvent) {
		super(event)
		e.adjustSuffixPos()
	})
	// Reposition the "%" suffix on every text change (matches the reference's
	// textChanged.connect(self._adjustSuffixPos)) without overriding the
	// LineEdit's built-in OnTextChanged handler.
	e.SetOnTextChangedExtra(func(text string) { e.adjustSuffixPos() })
	e.adjustSuffixPos()
	return e
}

func (e *OpacityLineEdit) adjustSuffixPos() {
	fm := e.FontMetrics()
	x := fm.Width(e.Text()) + 18
	e.suffixLabel.Move(x, 2)
}

// ColorDialog is a color picker dialog with a hue panel, brightness slider and
// numeric/hex editors.
type ColorDialog struct {
	*MaskDialogBase
	enableAlpha     bool
	oldColor        *qt.QColor
	color           *qt.QColor
	scrollArea      *widgets.SingleDirectionScrollArea
	scrollWidget    *qt.QWidget
	buttonGroup     *qt.QFrame
	yesButton       *widgets.PrimaryPushButton
	cancelButton    *qt.QPushButton
	titleLabel      *qt.QLabel
	huePanel        *HuePanel
	newColorCard    *ColorCard
	oldColorCard    *ColorCard
	brightSlider    *BrightnessSlider
	editLabel       *qt.QLabel
	redLabel        *qt.QLabel
	blueLabel       *qt.QLabel
	greenLabel      *qt.QLabel
	opacityLabel    *qt.QLabel
	hexLineEdit     *HexColorLineEdit
	redLineEdit     *ColorLineEdit
	greenLineEdit   *ColorLineEdit
	blueLineEdit    *ColorLineEdit
	opacityLineEdit *OpacityLineEdit
	vBoxLayout      *qt.QVBoxLayout

	OnColorChanged func(color *qt.QColor)
}

// NewColorDialog builds a color dialog.
func NewColorDialog(color *qt.QColor, title string, parent *qt.QWidget, enableAlpha bool) *ColorDialog {
	d := &ColorDialog{MaskDialogBase: NewMaskDialogBase(parent)}
	d.enableAlpha = enableAlpha

	initial := cloneColor(color)
	if !enableAlpha {
		initial.SetAlpha(255)
	}
	d.oldColor = cloneColor(initial)
	d.color = cloneColor(initial)

	d.scrollArea = widgets.NewSingleDirectionScrollArea(d.widget.QWidget, qt.Vertical)
	d.scrollWidget = qt.NewQWidget(d.scrollArea.QWidget)

	d.buttonGroup = qt.NewQFrame(d.widget.QWidget)
	d.yesButton = widgets.NewPrimaryPushButtonText("OK", d.buttonGroup.QWidget)
	d.cancelButton = qt.NewQPushButton5("Cancel", d.buttonGroup.QWidget)

	d.titleLabel = qt.NewQLabel5(title, d.scrollWidget)
	d.huePanel = NewHuePanel(initial, d.scrollWidget)
	d.newColorCard = NewColorCard(initial, d.scrollWidget, enableAlpha)
	d.oldColorCard = NewColorCard(initial, d.scrollWidget, enableAlpha)
	d.brightSlider = NewBrightnessSlider(initial, d.scrollWidget)

	d.editLabel = qt.NewQLabel5("Edit Color", d.scrollWidget)
	d.redLabel = qt.NewQLabel5("Red", d.scrollWidget)
	d.blueLabel = qt.NewQLabel5("Blue", d.scrollWidget)
	d.greenLabel = qt.NewQLabel5("Green", d.scrollWidget)
	d.opacityLabel = qt.NewQLabel5("Opacity", d.scrollWidget)
	d.hexLineEdit = NewHexColorLineEdit(initial, d.scrollWidget, enableAlpha)
	d.redLineEdit = NewColorLineEdit(d.color.Red(), d.scrollWidget)
	d.greenLineEdit = NewColorLineEdit(d.color.Green(), d.scrollWidget)
	d.blueLineEdit = NewColorLineEdit(d.color.Blue(), d.scrollWidget)
	d.opacityLineEdit = NewOpacityLineEdit(d.color.Alpha(), d.scrollWidget)

	d.vBoxLayout = qt.NewQVBoxLayout(d.widget.QWidget)
	d.showEventHook = d.updateStyle
	d.initWidget()
	return d
}

func (d *ColorDialog) initWidget() {
	d.scrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	d.scrollArea.SetViewportMargins(48, 24, 0, 24)
	d.scrollArea.SetWidget(d.scrollWidget)

	extra := 0
	if d.enableAlpha {
		extra = 40
	}
	d.widget.SetMaximumSize2(488, 696+extra)
	d.widget.Resize(488, 696+extra)
	d.scrollWidget.Resize(440, 560+extra)
	d.buttonGroup.SetFixedSize2(486, 81)
	d.yesButton.SetFixedWidth(216)
	d.cancelButton.SetFixedWidth(216)

	d.SetShadowEffect(60, 0, 10, qt.NewQColor11(0, 0, 0, 80))
	d.SetMaskColor(qt.NewQColor11(0, 0, 0, 76))

	d.setQss()
	d.initLayout()
	d.connectSignalToSlot()
}

func (d *ColorDialog) initLayout() {
	d.huePanel.Move(0, 46)
	d.newColorCard.Move(288, 46)
	geo := d.newColorCard.Geometry() // borrowed (QWidget.Geometry const_cast) — do NOT Delete
	d.oldColorCard.Move(288, geo.Bottom()+1)
	d.brightSlider.Move(0, 324)

	d.editLabel.Move(0, 385)
	d.redLineEdit.Move(0, 426)
	d.greenLineEdit.Move(0, 470)
	d.blueLineEdit.Move(0, 515)
	d.redLabel.Move(144, 434)
	d.greenLabel.Move(144, 478)
	d.blueLabel.Move(144, 524)
	d.hexLineEdit.Move(196, 381)

	if d.enableAlpha {
		d.opacityLineEdit.Move(0, 560)
		d.opacityLabel.Move(144, 567)
	} else {
		d.opacityLineEdit.Hide()
		d.opacityLabel.Hide()
	}

	d.vBoxLayout.SetSpacing(0)
	d.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	// Match Python's setAlignment(Qt.AlignTop) on the layout itself. An item-level
	// AlignTop (AddWidget3 alignment) keeps the stretch=1 scroll area pinned to
	// its sizeHint instead of expanding, which hides the bottom half of the
	// color editors behind an empty gap.
	d.vBoxLayout.QLayoutItem.SetAlignment(qt.AlignTop)
	d.vBoxLayout.AddWidget2(d.scrollArea.QWidget, 1)
	d.vBoxLayout.AddWidget3(d.buttonGroup.QWidget, 0, qt.AlignBottom)

	d.yesButton.Move(24, 25)
	d.cancelButton.Move(250, 25)
}

func (d *ColorDialog) setQss() {
	d.editLabel.SetObjectName("editLabel")
	d.titleLabel.SetObjectName("titleLabel")
	d.yesButton.SetObjectName("yesButton")
	d.cancelButton.SetObjectName("cancelButton")
	d.buttonGroup.SetObjectName("buttonGroup")
	common.FluentStyleSheet(common.FluentColorDialog).Apply(d.QWidget, common.ThemeAuto)
	d.titleLabel.AdjustSize()
	d.editLabel.AdjustSize()
}

// SetColor updates every editor from a color.
func (d *ColorDialog) SetColor(color *qt.QColor, movePicker bool) {
	d.color = cloneColor(color)
	d.brightSlider.SetColor(color)
	d.newColorCard.SetColor(color)
	d.hexLineEdit.SetColor(color)
	d.redLineEdit.SetText(strconv.Itoa(color.Red()))
	d.blueLineEdit.SetText(strconv.Itoa(color.Blue()))
	d.greenLineEdit.SetText(strconv.Itoa(color.Green()))
	if movePicker {
		d.huePanel.SetColor(color)
	}
}

func (d *ColorDialog) onHueChanged(color *qt.QColor) {
	d.color.SetHsv2(color.Hue(), color.Saturation(), d.color.Value(), d.color.Alpha())
	d.SetColor(d.color, true)
}

func (d *ColorDialog) onBrightnessChanged(color *qt.QColor) {
	d.color.SetHsv2(d.color.Hue(), d.color.Saturation(), color.Value(), color.Alpha())
	d.SetColor(d.color, false)
}

func (d *ColorDialog) onRedChanged(red string) {
	v, _ := strconv.Atoi(red)
	d.color.SetRed(v)
	d.SetColor(d.color, true)
}

func (d *ColorDialog) onBlueChanged(blue string) {
	v, _ := strconv.Atoi(blue)
	d.color.SetBlue(v)
	d.SetColor(d.color, true)
}

func (d *ColorDialog) onGreenChanged(green string) {
	v, _ := strconv.Atoi(green)
	d.color.SetGreen(v)
	d.SetColor(d.color, true)
}

func (d *ColorDialog) onOpacityChanged(opacity string) {
	v, _ := strconv.Atoi(opacity)
	d.color.SetAlpha(int(float64(v) / 100 * 255))
	d.SetColor(d.color, true)
}

func (d *ColorDialog) onHexColorChanged(color string) {
	d.color.SetNamedColor("#" + color)
	d.SetColor(d.color, true)
}

func (d *ColorDialog) onYesButtonClicked() {
	d.Accept()
	if !colorEquals(d.color, d.oldColor) && d.OnColorChanged != nil {
		d.OnColorChanged(d.color)
	}
}

func (d *ColorDialog) updateStyle() {
	d.SetStyle(qt.QApplication_Style())
	d.titleLabel.AdjustSize()
	d.editLabel.AdjustSize()
	d.redLabel.AdjustSize()
	d.greenLabel.AdjustSize()
	d.blueLabel.AdjustSize()
	d.opacityLabel.AdjustSize()
}

func (d *ColorDialog) connectSignalToSlot() {
	d.cancelButton.OnClicked(func() { d.Reject() })
	d.yesButton.OnClicked(d.onYesButtonClicked)

	d.huePanel.OnColorChanged = d.onHueChanged
	d.brightSlider.OnColorChanged = d.onBrightnessChanged

	d.redLineEdit.OnValueChanged = d.onRedChanged
	d.blueLineEdit.OnValueChanged = d.onBlueChanged
	d.greenLineEdit.OnValueChanged = d.onGreenChanged
	d.hexLineEdit.OnValueChanged = d.onHexColorChanged
	d.opacityLineEdit.OnValueChanged = d.onOpacityChanged
}

func clamp01(f float64) float64 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}
