// Command stacked_widget ports PyQt-Fluent-Widgets' examples/navigation/
// stacked_widget demo: EntranceTransitionStackedWidget and
// DrillInTransitionStackedWidget driven by a radio-button control panel.
package main

import (
	"fmt"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

// LOREM is the shared filler text used by the two sample pages (kept verbatim).
const LOREM = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et " +
	"dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip " +
	"ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu " +
	"fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
	"mollit anim id est laborum."

// ColorBlock is a QFrame painted with a solid background color.
type ColorBlock struct {
	*qt.QFrame
}

func newColorBlock(color *qt.QColor, parent *qt.QWidget) *ColorBlock {
	b := &ColorBlock{QFrame: qt.NewQFrame(parent)}
	b.SetStyleSheet(fmt.Sprintf("background: %s;", color.Name()))
	return b
}

// SamplePage1 is the grid-layout sample page.
type SamplePage1 struct {
	*qt.QWidget
}

func newSamplePage1(parent *qt.QWidget) *SamplePage1 {
	p := &SamplePage1{QWidget: qt.NewQWidget(parent)}
	layout := qt.NewQGridLayout(p.QWidget)
	layout.SetContentsMargins(6, 6, 6, 6)
	layout.SetSpacing(6)

	// blue accent block (left, spans 2 rows).
	accent := newColorBlock(common.ThemeColorValue(), p.QWidget)
	accent.SetMinimumWidth(200)
	layout.AddWidget3(accent.QWidget, 0, 0, 2, 1)

	// gray blocks (right side 2x2 grid).
	darkGray := newColorBlock(qt.NewQColor3(128, 128, 128), p.QWidget)
	lightGray1 := newColorBlock(qt.NewQColor3(192, 192, 192), p.QWidget)
	lightGray2 := newColorBlock(qt.NewQColor3(192, 192, 192), p.QWidget)
	darkGray2 := newColorBlock(qt.NewQColor3(160, 160, 160), p.QWidget)

	for _, blk := range []*ColorBlock{darkGray, lightGray1, lightGray2, darkGray2} {
		blk.SetMinimumHeight(120)
	}

	layout.AddWidget2(darkGray.QWidget, 0, 1)
	layout.AddWidget2(lightGray1.QWidget, 0, 2)
	layout.AddWidget2(lightGray2.QWidget, 1, 1)
	layout.AddWidget2(darkGray2.QWidget, 1, 2)

	// text at bottom.
	lbl := widgets.NewBodyLabelText(LOREM, p.QWidget)
	lbl.SetWordWrap(true)
	layout.AddWidget3(lbl.QWidget, 2, 0, 1, 3)

	layout.SetColumnStretch(1, 1)
	layout.SetColumnStretch(2, 1)
	layout.SetRowStretch(0, 1)
	layout.SetRowStretch(1, 1)
	return p
}

// SamplePage2 is the horizontal-layout sample page.
type SamplePage2 struct {
	*qt.QWidget
}

func newSamplePage2(parent *qt.QWidget) *SamplePage2 {
	p := &SamplePage2{QWidget: qt.NewQWidget(parent)}
	layout := qt.NewQHBoxLayout(p.QWidget)
	layout.SetContentsMargins(6, 6, 6, 6)
	layout.SetSpacing(16)

	// blue accent block (left).
	accent := newColorBlock(common.ThemeColorValue(), p.QWidget)
	accent.SetFixedSize2(140, 180)
	layout.AddWidget3(accent.QWidget, 0, qt.AlignTop)

	// text content (right).
	vbox := qt.NewQVBoxLayout2()
	vbox.SetSpacing(8)
	vbox.SetContentsMargins(0, 0, 0, 0)

	title := widgets.NewTitleLabelText("Lorem ipsum dolor sit amet, consectetur adipiscing elit", p.QWidget)
	title.SetWordWrap(true)
	vbox.AddWidget(title.QWidget)

	body := widgets.NewBodyLabelText(LOREM, p.QWidget)
	body.SetWordWrap(true)
	vbox.AddWidget(body.QWidget)
	vbox.AddStretchWithStretch(1)

	layout.AddLayout2(vbox.QLayout, 1)
	return p
}

// Window is the transition stacked widget demo window.
type Window struct {
	*qt.QWidget

	backStack []int

	stackedWidget *qt.QStackedWidget
	hBoxLayout    *qt.QHBoxLayout

	entranceStackedWidget *widgets.EntranceTransitionStackedWidget
	drillInStackedWidget  *widgets.DrillInTransitionStackedWidget

	ctrlPanel   *qt.QWidget
	buttonGroup *qt.QButtonGroup
	fwdBtn      *widgets.PushButton
	bwdBtn      *widgets.PushButton
}

func newWindow() *Window {
	w := &Window{QWidget: qt.NewQWidget2()}

	w.stackedWidget = qt.NewQStackedWidget(w.QWidget)
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)

	// entrance.
	w.entranceStackedWidget = widgets.NewEntranceTransitionStackedWidget(nil)
	w.entranceStackedWidget.SetMinimumHeight(500)
	w.entranceStackedWidget.AddWidget(newSamplePage1(nil).QWidget)
	w.entranceStackedWidget.AddWidget(newSamplePage2(nil).QWidget)
	w.stackedWidget.AddWidget(w.entranceStackedWidget.QWidget)

	// drill in.
	w.drillInStackedWidget = widgets.NewDrillInTransitionStackedWidget(nil)
	w.drillInStackedWidget.SetMinimumHeight(500)
	w.drillInStackedWidget.AddWidget(newSamplePage1(nil).QWidget)
	w.drillInStackedWidget.AddWidget(newSamplePage2(nil).QWidget)
	w.stackedWidget.AddWidget(w.drillInStackedWidget.QWidget)

	// control panel.
	w.createControlPanel()

	w.hBoxLayout.AddWidget2(w.stackedWidget.QWidget, 1)
	w.hBoxLayout.AddWidget2(w.ctrlPanel, 0)

	w.fwdBtn.OnClicked(w.onForward)
	w.bwdBtn.OnClicked(w.onBackward)
	w.buttonGroup.OnIdClicked(func(int) {
		checked := w.buttonGroup.CheckedButton()
		if checked != nil {
			w.stackedWidget.SetCurrentIndex(checked.Property("index").ToInt())
		}
	})

	w.Resize(800, 700)
	return w
}

func (w *Window) createControlPanel() {
	w.ctrlPanel = qt.NewQWidget(w.QWidget)
	w.buttonGroup = qt.NewQButtonGroup2(w.QObject)

	w.ctrlPanel.SetFixedWidth(260)

	layout := qt.NewQVBoxLayout(w.ctrlPanel)
	layout.AddWidget(widgets.NewSubtitleLabelText("Transition modes", w.ctrlPanel).QWidget)

	// transition type selection.
	for i, name := range []string{"Entrance", "DrillIn"} {
		button := widgets.NewRadioButtonText(name, w.ctrlPanel)
		button.SetProperty("index", qt.NewQVariant7(i))
		w.buttonGroup.AddButton(button.QAbstractButton)

		layout.AddWidget(button.QWidget)
		if i == 0 {
			button.SetChecked(true)
		}
	}

	layout.AddSpacing(16)
	layout.AddWidget(widgets.NewSubtitleLabelText("Navigate", w.ctrlPanel).QWidget)
	layout.SetContentsMargins(16, 16, 16, 16)
	layout.SetSpacing(8)

	// navigation buttons.
	w.fwdBtn = widgets.NewPushButtonText("Navigate Forward", w.ctrlPanel)
	w.bwdBtn = widgets.NewPushButtonText("Navigate Backward", w.ctrlPanel)
	layout.AddWidget(w.fwdBtn.QWidget)
	layout.AddWidget(w.bwdBtn.QWidget)
	layout.AddStretchWithStretch(1)
}

func (w *Window) onForward() {
	stack := w.currentStack()
	w.backStack = append(w.backStack, stack.CurrentIndex())
	stack.SetCurrentIndex((stack.CurrentIndex()+1)%stack.Count(), 0, false)
}

func (w *Window) onBackward() {
	if len(w.backStack) == 0 {
		return
	}
	stack := w.currentStack()
	idx := w.backStack[len(w.backStack)-1]
	w.backStack = w.backStack[:len(w.backStack)-1]
	stack.SetCurrentIndex(idx, 0, true)
}

// currentStack returns the transition stacked widget currently selected by the
// control panel.
func (w *Window) currentStack() *widgets.TransitionStackedWidget {
	if w.stackedWidget.CurrentIndex() == 0 {
		return w.entranceStackedWidget.TransitionStackedWidget
	}
	return w.drillInStackedWidget.TransitionStackedWidget
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
