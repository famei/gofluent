package dialog_box

import (
	"path/filepath"
	"sort"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// ClickableWindow is a frameless 292x72 clickable card with hover/press paint
// states.
type ClickableWindow struct {
	*qt.QWidget
	isPressed bool
	isEnter   bool

	OnClicked func()
}

func newClickableWindow(parent *qt.QWidget) *ClickableWindow {
	w := &ClickableWindow{QWidget: qt.NewQWidget(parent)}
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetWindowFlags(qt.FramelessWindowHint)
	w.SetFixedSize2(292, 72)

	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		super(e)
		w.isEnter = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		super(e)
		w.isEnter = false
		w.Update()
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		w.isPressed = false
		w.Update()
		if e.Button() == qt.LeftButton && w.OnClicked != nil {
			w.OnClicked()
		}
	})
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		w.isPressed = true
		w.Update()
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		w.paintBackground()
	})
	return w
}

func (w *ClickableWindow) paintBackground() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	isDark := common.IsDarkTheme()
	bg := 51
	if !isDark {
		bg = 204
	}
	painter.SetPenWithStyle(qt.NoPen)

	if !w.isEnter {
		brush := qt.NewQBrush3(qt.NewQColor3(bg, bg, bg))
		painter.SetBrush(brush)
		brush.Delete()
		painter.DrawRoundedRect3(w.Rect(), 4, 4)
	} else {
		pen := qt.NewQPen3(qt.NewQColor3(bg, bg, bg))
		pen.SetWidth(2)
		painter.SetPenWithPen(pen)
		pen.Delete()
		painter.DrawRect2(1, 1, w.Width()-2, w.Height()-2)
		painter.SetPenWithStyle(qt.NoPen)

		if !w.isPressed {
			bg = 24
			if !isDark {
				bg = 230
			}
			brush := qt.NewQBrush3(qt.NewQColor3(bg, bg, bg))
			painter.SetBrush(brush)
			brush.Delete()
			painter.DrawRect2(2, 2, w.Width()-4, w.Height()-4)
		} else {
			brush := qt.NewQBrush3(qt.NewQColor3(153, 153, 153))
			painter.SetBrush(brush)
			brush.Delete()
			painter.DrawRoundedRect2(5, 1, w.Width()-10, w.Height()-2, 2, 2)
		}
	}
	painter.End()
}

// FolderCard is a clickable card showing a folder name and path.
type FolderCard struct {
	*ClickableWindow
	folderPath string
	folderName string
	closeIcon  *qt.QPixmap
}

// NewFolderCard builds a folder card.
func NewFolderCard(folderPath string, parent *qt.QWidget) *FolderCard {
	c := &FolderCard{ClickableWindow: newClickableWindow(parent), folderPath: folderPath}
	c.folderName = filepath.Base(folderPath)
	color := common.GetIconColor(common.ThemeAuto, false)
	pm := loadPixmap("images/folder_list_dialog/Close_" + color + ".png")
	c.closeIcon = pm.Scaled3(12, 12, qt.KeepAspectRatio, qt.SmoothTransformation)
	pm.Delete()

	c.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		c.paintBackground()
		c.paintForeground()
	})
	return c
}

// FolderPath returns the full folder path.
func (c *FolderCard) FolderPath() string { return c.folderPath }

// FolderName returns the basename of the folder path.
func (c *FolderCard) FolderName() string { return c.folderName }

func (c *FolderCard) paintForeground() {
	painter := qt.NewQPainter2(c.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__TextAntialiasing | qt.QPainter__SmoothPixmapTransform | qt.QPainter__Antialiasing)

	var color *qt.QColor
	if common.IsDarkTheme() {
		color = qt.NewQColor2(qt.White)
	} else {
		color = qt.NewQColor2(qt.Black)
	}
	painter.SetPen(color)

	if c.isPressed {
		c.drawText(painter, 12, 8, 12, 7)
		painter.DrawPixmap9(c.Width()-26, 18, c.closeIcon)
	} else {
		c.drawText(painter, 10, 9, 10, 8)
		painter.DrawPixmap9(c.Width()-24, 20, c.closeIcon)
	}
	painter.End()
}

func (c *FolderCard) drawText(painter *qt.QPainter, x1, fontSize1, x2, fontSize2 int) {
	font := qt.NewQFont2("Microsoft YaHei")
	font.SetPixelSize(fontSize1)
	font.SetWeight(75)
	painter.SetFont(font)
	fm := qt.NewQFontMetrics(font)
	name := fm.ElidedText(c.folderName, qt.ElideRight, c.Width()-48)
	fm.Delete()
	p := qt.NewQPoint2(x1, 30)
	painter.DrawText2(p, name)
	p.Delete()

	font2 := qt.NewQFont2("Microsoft YaHei")
	font2.SetPixelSize(fontSize2)
	painter.SetFont(font2)
	fm2 := qt.NewQFontMetrics(font2)
	path := fm2.ElidedText(c.folderPath, qt.ElideRight, c.Width()-24)
	fm2.Delete()
	rect := qt.NewQRect4(x2, 37, c.Width()-16, 18)
	painter.DrawText6(rect, int(qt.AlignLeft), path)
	rect.Delete()

	font.Delete()
	font2.Delete()
}

// AddFolderCard is the "+" card that opens the folder picker.
type AddFolderCard struct {
	*ClickableWindow
	iconPix *qt.QPixmap
}

// NewAddFolderCard builds an add folder card.
func NewAddFolderCard(parent *qt.QWidget) *AddFolderCard {
	c := &AddFolderCard{ClickableWindow: newClickableWindow(parent)}
	color := common.GetIconColor(common.ThemeAuto, false)
	pm := loadPixmap("images/folder_list_dialog/Add_" + color + ".png")
	c.iconPix = pm.Scaled3(22, 22, qt.KeepAspectRatio, qt.SmoothTransformation)
	pm.Delete()

	c.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		c.paintBackground()
		painter := qt.NewQPainter2(c.QPaintDevice)
		defer painter.Delete()
		w, h := c.Width(), c.Height()
		pw, ph := c.iconPix.Width(), c.iconPix.Height()
		if !c.isPressed {
			painter.DrawPixmap9(int(w/2-pw/2), int(h/2-ph/2), c.iconPix)
		} else {
			painter.DrawPixmap11(int(w/2-(pw-4)/2), int(h/2-(ph-4)/2), pw-4, ph-4, c.iconPix)
		}
		painter.End()
	})
	return c
}

// FolderListDialog is a dialog that edits a list of folders.
type FolderListDialog struct {
	*MaskDialogBase
	title          string
	content        string
	originalPaths  []string
	folderPaths    []string
	vBoxLayout     *qt.QVBoxLayout
	titleLabel     *qt.QLabel
	contentLabel   *qt.QLabel
	scrollArea     *widgets.SingleDirectionScrollArea
	scrollWidget   *qt.QWidget
	completeButton *qt.QPushButton
	addFolderCard  *AddFolderCard
	folderCards    []*FolderCard
	scrollLayout   *qt.QVBoxLayout

	OnFolderChanged func(folders []string)
}

// NewFolderListDialog builds a folder list dialog.
func NewFolderListDialog(folderPaths []string, title, content string, parent *qt.QWidget) *FolderListDialog {
	d := &FolderListDialog{MaskDialogBase: NewMaskDialogBase(parent)}
	d.title = title
	d.content = content
	d.originalPaths = append([]string(nil), folderPaths...)
	d.folderPaths = append([]string(nil), folderPaths...)

	d.vBoxLayout = qt.NewQVBoxLayout(d.widget.QWidget)
	d.titleLabel = qt.NewQLabel5(title, d.widget.QWidget)
	d.contentLabel = qt.NewQLabel5(content, d.widget.QWidget)
	d.scrollArea = widgets.NewSingleDirectionScrollArea(d.widget.QWidget, qt.Vertical)
	d.scrollWidget = qt.NewQWidget(d.scrollArea.QWidget)
	d.completeButton = qt.NewQPushButton5("Done", d.widget.QWidget)
	d.addFolderCard = NewAddFolderCard(d.scrollWidget)
	for _, path := range folderPaths {
		d.folderCards = append(d.folderCards, NewFolderCard(path, d.scrollWidget))
	}
	d.initWidget()
	return d
}

func (d *FolderListDialog) initWidget() {
	d.setQss()

	w := maxInt(d.titleLabel.Width()+48, d.contentLabel.Width()+48)
	w = maxInt(w, 352)
	d.widget.SetFixedWidth(w)
	d.scrollArea.Resize(294, 72)
	d.scrollWidget.Resize(292, 72)
	d.scrollArea.SetFixedWidth(294)
	d.scrollWidget.SetFixedWidth(292)
	d.scrollArea.SetMaximumHeight(400)
	d.scrollArea.SetViewportMargins(0, 0, 0, 0)
	d.scrollArea.SetWidgetResizable(true)
	d.scrollArea.SetWidget(d.scrollWidget)
	d.scrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)

	d.initLayout()

	d.addFolderCard.OnClicked = d.showFileDialog
	d.completeButton.OnClicked(d.onButtonClicked)
	for _, card := range d.folderCards {
		card := card
		card.OnClicked = func() { d.showDeleteFolderCardDialog(card) }
	}
}

func (d *FolderListDialog) initLayout() {
	d.vBoxLayout.SetContentsMargins(24, 24, 24, 24)
	d.vBoxLayout.SetSizeConstraint(qt.QLayout__SetFixedSize)
	d.vBoxLayout.SetSpacing(0)

	layout1 := qt.NewQVBoxLayout2()
	layout1.SetContentsMargins(0, 0, 0, 0)
	layout1.SetSpacing(6)
	layout1.AddWidget3(d.titleLabel.QWidget, 0, qt.AlignTop)
	layout1.AddWidget3(d.contentLabel.QWidget, 0, qt.AlignTop)
	d.vBoxLayout.AddLayout2(layout1.QLayout, 0)
	d.vBoxLayout.AddSpacing(12)

	layout2 := qt.NewQHBoxLayout2()
	layout2.SetContentsMargins(4, 0, 4, 0)
	layout2.AddWidget3(d.scrollArea.QWidget, 0, qt.AlignCenter)
	d.vBoxLayout.AddLayout2(layout2.QLayout, 1)
	d.vBoxLayout.AddSpacing(24)

	d.scrollLayout = qt.NewQVBoxLayout(d.scrollWidget)
	d.scrollLayout.SetContentsMargins(0, 0, 0, 0)
	d.scrollLayout.SetSpacing(8)
	d.scrollLayout.AddWidget3(d.addFolderCard.QWidget, 0, qt.AlignTop)
	for _, card := range d.folderCards {
		d.scrollLayout.AddWidget3(card.QWidget, 0, qt.AlignTop)
	}

	layout3 := qt.NewQHBoxLayout2()
	layout3.SetContentsMargins(0, 0, 0, 0)
	layout3.AddStretchWithStretch(1)
	layout3.AddWidget(d.completeButton.QWidget)
	d.vBoxLayout.AddLayout2(layout3.QLayout, 0)

	d.adjustWidgetSize()
}

func (d *FolderListDialog) showFileDialog() {
	path := qt.QFileDialog_GetExistingDirectory3(d.QWidget, "Choose folder", "./")
	if path == "" {
		return
	}
	for _, p := range d.folderPaths {
		if p == path {
			return
		}
	}

	card := NewFolderCard(path, d.scrollWidget)
	d.scrollLayout.AddWidget3(card.QWidget, 0, qt.AlignTop)
	card.OnClicked = func() { d.showDeleteFolderCardDialog(card) }
	card.Show()

	d.folderPaths = append(d.folderPaths, path)
	d.folderCards = append(d.folderCards, card)
	d.adjustWidgetSize()
}

func (d *FolderListDialog) showDeleteFolderCardDialog(folderCard *FolderCard) {
	title := "Are you sure you want to delete the folder?"
	content := "If you delete the \"" + folderCard.folderName + "\" folder and remove it from the list, the folder will no longer appear in the list, but will not be deleted."
	dialog := NewDialog(title, content, d.Window())
	dialog.OnYes = func() { d.deleteFolderCard(folderCard) }
	dialog.Exec()
}

func (d *FolderListDialog) deleteFolderCard(folderCard *FolderCard) {
	d.scrollLayout.RemoveWidget(folderCard.QWidget)
	index := -1
	for i, c := range d.folderCards {
		if c.QWidget.UnsafePointer() == folderCard.QWidget.UnsafePointer() {
			index = i
			break
		}
	}
	if index < 0 {
		return
	}
	d.folderCards = append(d.folderCards[:index], d.folderCards[index+1:]...)
	d.folderPaths = append(d.folderPaths[:index], d.folderPaths[index+1:]...)
	folderCard.DeleteLater()
	d.adjustWidgetSize()
}

func (d *FolderListDialog) setQss() {
	d.titleLabel.SetObjectName("titleLabel")
	d.contentLabel.SetObjectName("contentLabel")
	d.completeButton.SetObjectName("completeButton")
	d.scrollWidget.SetObjectName("scrollWidget")

	common.FluentStyleSheet(common.FluentFolderListDialog).Apply(d.QWidget, common.ThemeAuto)
	d.SetStyle(qt.QApplication_Style())

	d.titleLabel.AdjustSize()
	d.contentLabel.AdjustSize()
	d.completeButton.AdjustSize()
}

func (d *FolderListDialog) onButtonClicked() {
	if !stringSlicesEqual(d.originalPaths, d.folderPaths) {
		d.SetEnabled(false)
		qt.QCoreApplication_ProcessEvents()
		if d.OnFolderChanged != nil {
			d.OnFolderChanged(d.folderPaths)
		}
	}
	d.Close()
}

func (d *FolderListDialog) adjustWidgetSize() {
	n := len(d.folderCards)
	h := 72*(n+1) + 8*n
	if h > 400 {
		h = 400
	}
	d.scrollArea.SetFixedHeight(h)
}

// FolderPaths returns the current folder list.
func (d *FolderListDialog) FolderPaths() []string { return d.folderPaths }

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as := append([]string(nil), a...)
	bs := append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}
