package view

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/components"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// BannerWidget is the header banner of the home page (port of
// app/view/home_interface.py BannerWidget).
type BannerWidget struct {
	*qt.QWidget
	galleryLabel *qt.QLabel
	banner       *qt.QPixmap
	linkCardView *components.LinkCardView
	vBoxLayout   *qt.QVBoxLayout
}

// NewBannerWidget builds the banner.
func NewBannerWidget(parent *qt.QWidget) *BannerWidget {
	b := &BannerWidget{QWidget: qt.NewQWidget(parent)}
	b.SetFixedHeight(336)

	b.vBoxLayout = qt.NewQVBoxLayout(b.QWidget)
	b.galleryLabel = qt.NewQLabel5("Fluent Gallery", b.QWidget)
	b.banner = resource.Pixmap("header1.png")
	b.linkCardView = components.NewLinkCardView(b.QWidget)

	b.galleryLabel.SetObjectName("galleryLabel")
	b.SetObjectName("bannerWidget")

	b.vBoxLayout.SetSpacing(0)
	b.vBoxLayout.SetContentsMargins(0, 20, 0, 0)
	b.vBoxLayout.AddWidget(b.galleryLabel.QWidget)
	b.vBoxLayout.AddWidget3(b.linkCardView.QWidget, 1, qt.AlignBottom)

	b.linkCardView.AddCard(
		resource.Icon("logo.png"),
		"Getting started",
		"An overview of app development options and samples.",
		gallerycommon.HELP_URL,
	)
	b.linkCardView.AddCard(
		gcommon.GitHub,
		"GitHub repo",
		"The latest fluent design controls and styles for your applications.",
		gallerycommon.REPO_URL,
	)
	b.linkCardView.AddCard(
		gcommon.Code,
		"Code samples",
		"Find samples that demonstrate specific tasks, features and APIs.",
		gallerycommon.EXAMPLE_URL,
	)
	b.linkCardView.AddCard(
		gcommon.Feedback,
		"Send feedback",
		"Help us improve PyQt-Fluent-Widgets by providing feedback.",
		gallerycommon.FEEDBACK_URL,
	)

	b.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__SmoothPixmapTransform | qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		path := qt.NewQPainterPath()
		path.SetFillRule(qt.WindingFill)
		w := b.Width()
		h := b.Height()
		path.AddRoundedRect2(0, 0, float64(w), float64(h), 10, 10)
		path.AddRect2(0, float64(h-50), 50, 50)
		path.AddRect2(float64(w-50), 0, 50, 50)
		path.AddRect2(float64(w-50), float64(h-50), 50, 50)
		simplified := path.Simplified()
		path.Delete()

		gradient := qt.NewQLinearGradient3(0, 0, 0, float64(h))
		if !gcommon.IsDarkTheme() {
			gradient.SetColorAt(0, qt.NewQColor11(207, 216, 228, 255))
			gradient.SetColorAt(1, qt.NewQColor11(207, 216, 228, 0))
		} else {
			gradient.SetColorAt(0, qt.NewQColor11(0, 0, 0, 255))
			gradient.SetColorAt(1, qt.NewQColor11(0, 0, 0, 0))
		}
		gradientBrush := qt.NewQBrush10(gradient.QGradient)
		painter.FillPath(simplified, gradientBrush)
		gradientBrush.Delete()
		gradient.Delete()

		pixmap := b.banner.Scaled3(w, h, qt.IgnoreAspectRatio, qt.SmoothTransformation)
		pixmapBrush := qt.NewQBrush7(pixmap)
		painter.FillPath(simplified, pixmapBrush)
		pixmapBrush.Delete()
		painter.End()
	})
	return b
}

// HomeInterface is the "Home" gallery page (port of app/view/home_interface.py).
type HomeInterface struct {
	*widgets.ScrollArea
	banner     *BannerWidget
	view       *qt.QWidget
	vBoxLayout *qt.QVBoxLayout
}

// NewHomeInterface builds the home interface.
func NewHomeInterface(parent *qt.QWidget) *HomeInterface {
	i := &HomeInterface{ScrollArea: widgets.NewScrollArea(parent)}
	i.banner = NewBannerWidget(i.QWidget)
	i.view = qt.NewQWidget(i.QWidget)
	i.vBoxLayout = qt.NewQVBoxLayout(i.view)

	i.initWidget()
	i.loadSamples()
	return i
}

func (i *HomeInterface) initWidget() {
	i.view.SetObjectName("view")
	i.SetObjectName("homeInterface")
	gallerycommon.StyleHomeInterface.Apply(i.QWidget, gcommon.ThemeAuto)

	i.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	i.SetWidget(i.view)
	i.SetWidgetResizable(true)

	i.vBoxLayout.SetContentsMargins(0, 0, 0, 36)
	i.vBoxLayout.SetSpacing(40)
	i.vBoxLayout.AddWidget(i.banner.QWidget)
}

func (i *HomeInterface) loadSamples() {
	basicInputView := components.NewSampleCardView("Basic input samples", i.view)
	basicInputView.AddSampleCard(resource.Icon("controls/Button.png"), "Button",
		"A control that responds to user input and emit clicked signal.", "basicInputInterface", 0)
	basicInputView.AddSampleCard(resource.Icon("controls/Checkbox.png"), "CheckBox",
		"A control that a user can select or clear.", "basicInputInterface", 8)
	basicInputView.AddSampleCard(resource.Icon("controls/ComboBox.png"), "ComboBox",
		"A drop-down list of items a user can select from.", "basicInputInterface", 10)
	basicInputView.AddSampleCard(resource.Icon("controls/DropDownButton.png"), "DropDownButton",
		"A button that displays a flyout of choices when clicked.", "basicInputInterface", 12)
	basicInputView.AddSampleCard(resource.Icon("controls/HyperlinkButton.png"), "HyperlinkButton",
		"A button that appears as hyperlink text, and can navigate to a URI or handle a Click event.", "basicInputInterface", 18)
	basicInputView.AddSampleCard(resource.Icon("controls/RadioButton.png"), "RadioButton",
		"A control that allows a user to select a single option from a group of options.", "basicInputInterface", 19)
	basicInputView.AddSampleCard(resource.Icon("controls/Slider.png"), "Slider",
		"A control that lets the user select from a range of values by moving a Thumb control along a track.", "basicInputInterface", 20)
	basicInputView.AddSampleCard(resource.Icon("controls/SplitButton.png"), "SplitButton",
		"A two-part button that displays a flyout when its secondary part is clicked.", "basicInputInterface", 21)
	basicInputView.AddSampleCard(resource.Icon("controls/ToggleSwitch.png"), "SwitchButton",
		"A switch that can be toggled between 2 states.", "basicInputInterface", 25)
	basicInputView.AddSampleCard(resource.Icon("controls/ToggleButton.png"), "ToggleButton",
		"A button that can be switched between two states like a CheckBox.", "basicInputInterface", 26)
	i.vBoxLayout.AddWidget(basicInputView.QWidget)

	dateTimeView := components.NewSampleCardView("Date & time samples", i.view)
	dateTimeView.AddSampleCard(resource.Icon("controls/CalendarDatePicker.png"), "CalendarPicker",
		"A control that lets a user pick a date value using a calendar.", "dateTimeInterface", 0)
	dateTimeView.AddSampleCard(resource.Icon("controls/DatePicker.png"), "DatePicker",
		"A control that lets a user pick a date value.", "dateTimeInterface", 2)
	dateTimeView.AddSampleCard(resource.Icon("controls/TimePicker.png"), "TimePicker",
		"A configurable control that lets a user pick a time value.", "dateTimeInterface", 4)
	i.vBoxLayout.AddWidget(dateTimeView.QWidget)

	dialogView := components.NewSampleCardView("Dialog samples", i.view)
	dialogView.AddSampleCard(resource.Icon("controls/Flyout.png"), "Dialog",
		"A frameless message dialog.", "dialogInterface", 0)
	dialogView.AddSampleCard(resource.Icon("controls/ContentDialog.png"), "MessageBox",
		"A message dialog with mask.", "dialogInterface", 1)
	dialogView.AddSampleCard(resource.Icon("controls/ColorPicker.png"), "ColorDialog",
		"A dialog that allows user to select color.", "dialogInterface", 2)
	dialogView.AddSampleCard(resource.Icon("controls/Flyout.png"), "Flyout",
		"Shows contextual information and enables user interaction.", "dialogInterface", 3)
	dialogView.AddSampleCard(resource.Icon("controls/TeachingTip.png"), "TeachingTip",
		"A content-rich flyout for guiding users and enabling teaching moments.", "dialogInterface", 5)
	i.vBoxLayout.AddWidget(dialogView.QWidget)

	layoutView := components.NewSampleCardView("Layout samples", i.view)
	layoutView.AddSampleCard(resource.Icon("controls/Grid.png"), "FlowLayout",
		"A layout arranges components in a left-to-right flow, wrapping to the next row when the current row is full.", "layoutInterface", 0)
	i.vBoxLayout.AddWidget(layoutView.QWidget)

	menuView := components.NewSampleCardView("Menu & toolbars samples", i.view)
	menuView.AddSampleCard(resource.Icon("controls/MenuFlyout.png"), "RoundMenu",
		"Shows a contextual list of simple commands or options.", "menuInterface", 0)
	menuView.AddSampleCard(resource.Icon("controls/CommandBar.png"), "CommandBar",
		"Shows a contextual list of simple commands or options.", "menuInterface", 2)
	menuView.AddSampleCard(resource.Icon("controls/CommandBarFlyout.png"), "CommandBarFlyout",
		"A mini-toolbar displaying proactive commands, and an optional menu of commands.", "menuInterface", 3)
	i.vBoxLayout.AddWidget(menuView.QWidget)

	navigationView := components.NewSampleCardView("Navigation", i.view)
	navigationView.AddSampleCard(resource.Icon("controls/BreadcrumbBar.png"), "BreadcrumbBar",
		"Shows the trail of navigation taken to the current location.", "navigationViewInterface", 0)
	navigationView.AddSampleCard(resource.Icon("controls/Pivot.png"), "Pivot",
		"Presents information from different sources in a tabbed view.", "navigationViewInterface", 1)
	navigationView.AddSampleCard(resource.Icon("controls/TabView.png"), "TabView",
		"Presents information from different sources in a tabbed view.", "navigationViewInterface", 3)
	i.vBoxLayout.AddWidget(navigationView.QWidget)

	scrollView := components.NewSampleCardView("Scrolling samples", i.view)
	scrollView.AddSampleCard(resource.Icon("controls/ScrollViewer.png"), "ScrollArea",
		"A container control that lets the user pan and zoom its content smoothly.", "scrollInterface", 0)
	scrollView.AddSampleCard(resource.Icon("controls/PipsPager.png"), "PipsPager",
		"A control to let the user navigate through a paginated collection when the page numbers do not need to be visually known.", "scrollInterface", 3)
	i.vBoxLayout.AddWidget(scrollView.QWidget)

	stateInfoView := components.NewSampleCardView("Status & info samples", i.view)
	stateInfoView.AddSampleCard(resource.Icon("controls/ProgressRing.png"), "StateToolTip",
		"Shows the apps progress on a task, or that the app is performing ongoing work that does block user interaction.", "statusInfoInterface", 0)
	stateInfoView.AddSampleCard(resource.Icon("controls/InfoBadge.png"), "InfoBadge",
		"An non-intrusive Ul to display notifications or bring focus to an area.", "statusInfoInterface", 3)
	stateInfoView.AddSampleCard(resource.Icon("controls/InfoBar.png"), "InfoBar",
		"An inline message to display app-wide status change information.", "statusInfoInterface", 4)
	stateInfoView.AddSampleCard(resource.Icon("controls/ProgressBar.png"), "ProgressBar",
		"Shows the apps progress on a task, or that the app is performing ongoing work that doesn't block user interaction.", "statusInfoInterface", 8)
	stateInfoView.AddSampleCard(resource.Icon("controls/ProgressRing.png"), "ProgressRing",
		"Shows the apps progress on a task, or that the app is performing ongoing work that doesn't block user interaction.", "statusInfoInterface", 10)
	stateInfoView.AddSampleCard(resource.Icon("controls/ToolTip.png"), "ToolTip",
		"Displays information for an element in a pop-up window.", "statusInfoInterface", 1)
	i.vBoxLayout.AddWidget(stateInfoView.QWidget)

	textView := components.NewSampleCardView("Text samples", i.view)
	textView.AddSampleCard(resource.Icon("controls/TextBox.png"), "LineEdit",
		"A single-line plain text field.", "textInterface", 0)
	textView.AddSampleCard(resource.Icon("controls/PasswordBox.png"), "PasswordLineEdit",
		"A control for entering passwords.", "textInterface", 2)
	textView.AddSampleCard(resource.Icon("controls/NumberBox.png"), "SpinBox",
		"A text control used for numeric input and evaluation of algebraic equations.", "textInterface", 3)
	textView.AddSampleCard(resource.Icon("controls/RichEditBox.png"), "TextEdit",
		"A rich text editing control that supports formatted text, hyperlinks, and other rich content.", "textInterface", 8)
	i.vBoxLayout.AddWidget(textView.QWidget)

	collectionView := components.NewSampleCardView("View samples", i.view)
	collectionView.AddSampleCard(resource.Icon("controls/ListView.png"), "ListView",
		"A control that presents a collection of items in a vertical list.", "viewInterface", 0)
	collectionView.AddSampleCard(resource.Icon("controls/DataGrid.png"), "TableView",
		"The DataGrid control provides a flexible way to display a collection of data in rows and columns.", "viewInterface", 1)
	collectionView.AddSampleCard(resource.Icon("controls/TreeView.png"), "TreeView",
		"The TreeView control is a hierarchical list pattern with expanding and collapsing nodes that contain nested items.", "viewInterface", 2)
	collectionView.AddSampleCard(resource.Icon("controls/FlipView.png"), "FlipView",
		"Presents a collection of items that the user can flip through,one item at a time.", "viewInterface", 4)
	i.vBoxLayout.AddWidget(collectionView.QWidget)
}
