// Command card_widget migrates examples/view/card_widget/demo.py: a Microsoft
// Store style window that showcases CardWidget and friends, plus two mica-style
// windows (emoji gallery and app cards).
//
// Documented degradations (library limits, see MIGRATION_GUIDE §10):
//   - The qframelesswindow AcrylicWindow/Mica effect is dropped (FramelessWindow
//     port keeps a solid background only).
//   - HeaderCardWidget does not expose its header/view layouts, so GalleryCard,
//     DescriptionCard and SystemRequirementCard keep only their titles.
//   - LightBox is omitted (it depends on the GalleryCard flip view and the
//     widget-backed AcrylicBrush grab).
//   - NavigationBar "addItem" has no exported accessor on MSFluentWindow, so the
//     two placeholder navigation items ("编辑"/"设置") are not added.
package main

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/layout"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/resources"
	"github.com/famei/gofluent/window"
	qt "github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/svg"
)

//go:embed all:resource
var resourceFS embed.FS

// svgIcon renders an embedded SVG into a QIcon (the static Qt build does not
// register the qsvg image-format plugin, so QPixmap.LoadFromData cannot be used;
// QSvgRenderer is used directly instead).
func svgIcon(fsys fs.FS, name string) *qt.QIcon {
	renderer := svg.NewQSvgRenderer3(asset.Bytes(fsys, name))
	defer renderer.Delete()

	const size = 32
	pm := qt.NewQPixmap2(size, size)
	transparent := qt.NewQColor11(0, 0, 0, 0)
	pm.FillWithFillColor(transparent)
	transparent.Delete()

	painter := qt.NewQPainter2(pm.QPaintDevice)
	renderer.Render2(painter, qt.NewQRectF4(0, 0, size, size))
	painter.End()
	painter.Delete()

	icon := qt.NewQIcon2(pm)
	pm.Delete()
	return icon
}

// ctrlIcon loads a control sprite from the gofluent resources.
func ctrlIcon(name string) *qt.QIcon {
	return asset.QIcon(resources.Images, "images/controls/"+name)
}

// ---------------------------------------------------------------------------
// AppCard

type appCard struct {
	*widgets.CardWidget
	moreButton *widgets.TransparentToolButton
}

func newAppCard(icon interface{}, title, content string, parent *qt.QWidget) *appCard {
	card := &appCard{CardWidget: widgets.NewCardWidget(parent)}

	iconWidget := widgets.NewIconWidgetIcon(icon, nil)
	titleLabel := widgets.NewBodyLabelText(title, card.QWidget)
	contentLabel := widgets.NewCaptionLabelText(content, card.QWidget)
	openButton := widgets.NewPushButtonText("打开", card.QWidget)
	card.moreButton = widgets.NewTransparentToolButtonIcon(common.More, card.QWidget)

	hBoxLayout := qt.NewQHBoxLayout(card.QWidget)
	vBoxLayout := qt.NewQVBoxLayout2()

	card.SetFixedHeight(73)
	iconWidget.SetFixedSize2(48, 48)
	contentLabel.SetTextColor(qt.NewQColor6("#606060"), qt.NewQColor6("#d2d2d2"))
	openButton.SetFixedWidth(120)

	hBoxLayout.SetContentsMargins(20, 11, 11, 11)
	hBoxLayout.SetSpacing(15)
	hBoxLayout.AddWidget(iconWidget.QWidget)

	vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	vBoxLayout.SetSpacing(0)
	vBoxLayout.AddWidget3(titleLabel.QWidget, 0, qt.AlignVCenter)
	vBoxLayout.AddWidget3(contentLabel.QWidget, 0, qt.AlignVCenter)
	hBoxLayout.AddLayout(vBoxLayout.QLayout)

	hBoxLayout.AddStretchWithStretch(1)
	hBoxLayout.AddWidget3(openButton.QWidget, 0, qt.AlignRight)
	hBoxLayout.AddWidget3(card.moreButton.QWidget, 0, qt.AlignRight)

	card.moreButton.SetFixedSize2(32, 32)
	card.moreButton.OnClicked(card.onMoreButtonClicked)
	return card
}

func (c *appCard) onMoreButtonClicked() {
	menu := widgets.NewRoundMenu("", c.QWidget)
	menu.AddAction(common.NewActionFluentIcon(common.Share, "共享", c.QObject).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.ChatBubbles, "写评论", c.QObject).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Pin, "固定到任务栏", c.QObject).QAction)

	x := (c.moreButton.Width()-menu.Width())/2 + 10
	p := qt.NewQPoint2(x, c.moreButton.Height())
	pos := c.moreButton.MapToGlobal(p)
	p.Delete()
	menu.ExecAt(pos)
}

// ---------------------------------------------------------------------------
// EmojiCard

func newEmojiCard(name string, parent *qt.QWidget) *widgets.ElevatedCardWidget {
	card := widgets.NewElevatedCardWidget(parent)

	iconWidget := widgets.NewImageLabelImage(asset.QImage(resourceFS, name), card.QWidget)
	stem := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	label := widgets.NewCaptionLabelText(stem, card.QWidget)

	iconWidget.ScaledToHeight(68)

	vBoxLayout := qt.NewQVBoxLayout(card.QWidget)
	vBoxLayout.AddStretchWithStretch(1)
	vBoxLayout.AddWidget3(iconWidget.QWidget, 0, qt.AlignCenter)
	vBoxLayout.AddStretchWithStretch(1)
	vBoxLayout.AddWidget3(label.QWidget, 0, qt.AlignHCenter|qt.AlignBottom)

	card.SetFixedSize2(168, 176)
	return card
}

// ---------------------------------------------------------------------------
// StatisticsWidget

func newStatisticsWidget(title, value string, parent *qt.QWidget) *qt.QWidget {
	w := qt.NewQWidget(parent)
	titleLabel := widgets.NewCaptionLabelText(title, w)
	valueLabel := widgets.NewBodyLabelText(value, w)
	vBoxLayout := qt.NewQVBoxLayout(w)

	vBoxLayout.SetContentsMargins(16, 0, 16, 0)
	vBoxLayout.AddWidget3(valueLabel.QWidget, 0, qt.AlignTop)
	vBoxLayout.AddWidget3(titleLabel.QWidget, 0, qt.AlignBottom)

	common.SetFont(valueLabel.QWidget, 18, int(qt.QFont__DemiBold))
	titleLabel.SetTextColor(qt.NewQColor3(96, 96, 96), qt.NewQColor3(206, 206, 206))
	return w
}

// ---------------------------------------------------------------------------
// AppInfoCard

func newAppInfoCard(parent *qt.QWidget) *widgets.SimpleCardWidget {
	card := widgets.NewSimpleCardWidget(parent)

	iconLabel := widgets.NewImageLabelImage(asset.QImage(resources.Images, "images/logo.png"), card.QWidget)
	iconLabel.SetBorderRadius(8, 8, 8, 8)
	iconLabel.ScaledToWidth(120)

	nameLabel := widgets.NewTitleLabelText("QFluentWidgets", card.QWidget)
	installButton := widgets.NewPrimaryPushButtonText("安装", card.QWidget)
	companyLabel := widgets.NewHyperlinkLabelURL("https://qfluentwidgets.com", "Shokokawaii Inc.", card.QWidget)
	installButton.SetFixedWidth(160)

	scoreWidget := newStatisticsWidget("平均", "5.0", card.QWidget)
	separator := widgets.NewVerticalSeparator(card.QWidget)
	commentWidget := newStatisticsWidget("评论数", "3K", card.QWidget)

	descriptionLabel := widgets.NewBodyLabelText(
		"PyQt-Fluent-Widgets 是一个基于 PyQt/PySide 的 Fluent Design 风格组件库，包含许多美观实用的组件，支持亮暗主题无缝切换和自定义主题色，帮助开发者快速实现美观优雅的现代化界面。", card.QWidget)
	descriptionLabel.SetWordWrap(true)

	tagButton := widgets.NewPillPushButtonText("组件库", card.QWidget)
	tagButton.SetCheckable(false)
	common.SetFont(tagButton.QWidget, 12, 400)
	tagButton.SetFixedSize2(80, 32)

	shareButton := widgets.NewTransparentToolButtonIcon(common.Share, card.QWidget)
	shareButton.SetFixedSize2(32, 32)
	shareButton.SetIconSize(qt.NewQSize2(14, 14))

	hBoxLayout := qt.NewQHBoxLayout(card.QWidget)
	vBoxLayout := qt.NewQVBoxLayout2()
	topLayout := qt.NewQHBoxLayout2()
	statisticsLayout := qt.NewQHBoxLayout2()
	buttonLayout := qt.NewQHBoxLayout2()

	hBoxLayout.SetSpacing(30)
	hBoxLayout.SetContentsMargins(34, 24, 24, 24)
	hBoxLayout.AddWidget(iconLabel.QWidget)
	hBoxLayout.AddLayout(vBoxLayout.QLayout)

	vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	vBoxLayout.SetSpacing(0)

	vBoxLayout.AddLayout(topLayout.QLayout)
	topLayout.SetContentsMargins(0, 0, 0, 0)
	topLayout.AddWidget(nameLabel.QWidget)
	topLayout.AddWidget3(installButton.QWidget, 0, qt.AlignRight)

	vBoxLayout.AddSpacing(3)
	vBoxLayout.AddWidget(companyLabel.QWidget)

	vBoxLayout.AddSpacing(20)
	vBoxLayout.AddLayout(statisticsLayout.QLayout)
	statisticsLayout.SetContentsMargins(0, 0, 0, 0)
	statisticsLayout.SetSpacing(10)
	statisticsLayout.AddWidget(scoreWidget)
	statisticsLayout.AddWidget(separator.QWidget)
	statisticsLayout.AddWidget(commentWidget)

	vBoxLayout.AddSpacing(20)
	vBoxLayout.AddWidget(descriptionLabel.QWidget)

	vBoxLayout.AddSpacing(12)
	buttonLayout.SetContentsMargins(0, 0, 0, 0)
	vBoxLayout.AddLayout(buttonLayout.QLayout)
	buttonLayout.AddWidget3(tagButton.QWidget, 0, qt.AlignLeft)
	buttonLayout.AddWidget3(shareButton.QWidget, 0, qt.AlignRight)

	card.SetBorderRadius(8)
	return card
}

// ---------------------------------------------------------------------------
// SettinsCard (GroupHeaderCardWidget is fully exposed via AddGroup)

func newSettingsCard(parent *qt.QWidget) *widgets.GroupHeaderCardWidget {
	card := widgets.NewGroupHeaderCardWidget(parent)
	card.SetTitle("基本设置")
	card.SetBorderRadius(8)

	chooseButton := widgets.NewPushButtonText("选择", nil)
	comboBox := widgets.NewComboBox(nil)
	lineEdit := widgets.NewSearchLineEdit(nil)

	chooseButton.SetFixedWidth(120)
	lineEdit.SetFixedWidth(320)
	comboBox.SetFixedWidth(320)
	comboBox.AddItems([]string{"始终显示（首次打包时建议启用）", "始终隐藏"})
	lineEdit.SetPlaceholderText("输入入口脚本的路径")

	card.AddGroup(svgIcon(resourceFS, "resource/Rocket.svg"), "构建目录", "选择 Nuitka 的输出目录", chooseButton.QWidget, 0)
	card.AddGroup(svgIcon(resourceFS, "resource/Joystick.svg"), "运行终端", "设置是否显示命令行终端", comboBox.QWidget, 0)
	card.AddGroup(svgIcon(resourceFS, "resource/Python.svg"), "入口脚本", "选择软件的入口脚本", lineEdit.QWidget, 0)
	return card
}

// ---------------------------------------------------------------------------
// Mica window (frameless window with an MSFluentTitleBar)

func newMicaWindow() *window.FluentWidget {
	w := window.NewFluentWidget(nil)
	w.SetTitleBar(window.NewMSFluentTitleBar(w.QWidget).QWidget)
	return w
}

// Demo1: emoji gallery built on a FlowLayout.
func newDemo1() *window.FluentWidget {
	w := newMicaWindow()
	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.SetWindowTitle("Fluent Emoji gallery")

	flowLayout := layout.NewFlowLayout(w.QWidget, true, false)

	w.Resize(580, 680)
	flowLayout.SetSpacing(6)
	flowLayout.SetContentsMargins(30, 60, 30, 30)

	names, err := fs.Glob(resourceFS, "resource/*.png")
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		flowLayout.AddWidget(newEmojiCard(name, w.QWidget).QWidget)
	}
	return w
}

// Demo2: app cards stacked in a vertical layout.
func newDemo2() *window.FluentWidget {
	w := newMicaWindow()
	w.Resize(600, 600)

	vBoxLayout := qt.NewQVBoxLayout(w.QWidget)
	vBoxLayout.SetSpacing(6)
	vBoxLayout.SetContentsMargins(30, 60, 30, 30)

	addCard := func(icon interface{}, title, content string) {
		vBoxLayout.AddWidget3(newAppCard(icon, title, content, w.QWidget).QWidget, 0, qt.AlignTop)
	}

	addCard(asset.QIcon(resources.Images, "images/logo.png"), "PyQt-Fluent-Widgets", "Shokokawaii Inc.")
	addCard(ctrlIcon("TitleBar.png"), "PyQt-Frameless-Window", "Shokokawaii Inc.")
	addCard(ctrlIcon("RatingControl.png"), "反馈中心", "Microsoft Corporation")
	addCard(ctrlIcon("Checkbox.png"), "Microsoft 使用技巧", "Microsoft Corporation")
	addCard(ctrlIcon("Pivot.png"), "MSN 天气", "Microsoft Corporation")
	addCard(ctrlIcon("MediaPlayerElement.png"), "电影和电视", "Microsoft Corporation")
	addCard(ctrlIcon("PersonPicture.png"), "照片", "Microsoft Corporation")
	return w
}

// ---------------------------------------------------------------------------
// AppInterface (scrollable interface shown inside Demo3)

func newAppInterface(parent *qt.QWidget) *widgets.ScrollArea {
	area := widgets.NewScrollArea(parent)
	area.SetObjectName("appInterface")

	view := qt.NewQWidget(area.QWidget)
	vBoxLayout := qt.NewQVBoxLayout(view)

	appCard := newAppInfoCard(area.QWidget)

	// HeaderCardWidget does not expose its view layout, so these three cards are
	// title-only degradations of GalleryCard / DescriptionCard / SystemRequirementCard.
	galleryCard := widgets.NewHeaderCardWidgetTitle("屏幕截图", area.QWidget)
	galleryCard.SetBorderRadius(8)
	descriptionCard := widgets.NewHeaderCardWidgetTitle("描述", area.QWidget)
	descriptionCard.SetBorderRadius(8)
	systemCard := widgets.NewHeaderCardWidgetTitle("系统要求", area.QWidget)
	systemCard.SetBorderRadius(8)

	settingCard := newSettingsCard(area.QWidget)

	area.SetWidget(view)
	area.SetWidgetResizable(true)

	vBoxLayout.SetSpacing(10)
	vBoxLayout.SetContentsMargins(0, 0, 10, 30)
	vBoxLayout.AddWidget3(appCard.QWidget, 0, qt.AlignTop)
	vBoxLayout.AddWidget3(galleryCard.QWidget, 0, qt.AlignTop)
	vBoxLayout.AddWidget3(settingCard.QWidget, 0, qt.AlignTop)
	vBoxLayout.AddWidget3(descriptionCard.QWidget, 0, qt.AlignTop)
	vBoxLayout.AddWidget3(systemCard.QWidget, 0, qt.AlignTop)

	area.EnableTransparentBackground()
	return area
}

// Demo3: the Microsoft Store style window.
func newDemo3() *window.MSFluentWindow {
	w := window.NewMSFluentWindow(nil)

	appInterface := newAppInterface(w.QWidget)

	w.AddSubInterface(appInterface.QWidget, common.Library, "库", common.Library, navigation.NavigationItemPositionTop, true)

	w.Resize(880, 760)
	w.SetWindowTitle("PyQt-Fluent-Widgets")
	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.TitleBar().Raise()
	return w
}

func main() {
	demo.Run(
		func() *qt.QWidget { return newDemo1().QWidget },
		func() *qt.QWidget { return newDemo2().QWidget },
		func() *qt.QWidget { return newDemo3().QWidget },
	)
}
