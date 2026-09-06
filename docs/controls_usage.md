# gofluent Controls Usage Guide

> gofluent is a Go + miqt (Qt5) port of PyQt-Fluent-Widgets. The component library lives in `gofluent/components/` and is organized by function into 6 packages: `widgets`, `navigation`, `dialog_box`, `settings`, `date_time`, `layout`. This document details each control's **function, parameters, calling methods, and effects**.

## Table of Contents

1. [widgets package](#1-widgets-package) — buttons, cards, input boxes, lists, menus, progress, scrolling, tab views, tables, trees, tooltips, etc.
2. [navigation package](#2-navigation-package) — navigation panel, breadcrumb, Pivot, segmented controls, etc.
3. [dialog_box package](#3-dialog_box-package) — message dialogs, mask dialogs, color picker, folder list, etc.
4. [settings package](#4-settings-package) — the setting card series
5. [date_time package](#5-date_time-package) — date / time / calendar pickers
6. [layout package](#6-layout-package) — flow layout, vertical layout, expand layout

## General Conventions

- Every control embeds a `*qt.QWidget` (field name `QWidget`); put it into a layout or use it as a parent container via `.QWidget`, e.g. `layout.AddWidget(btn.QWidget)`.
- `parent *qt.QWidget` denotes the parent widget; it is almost always the last parameter of a constructor and may be `nil`.
- The `icon interface{}` icon source parameter accepts `*qt.QIcon`, a string path, `common.FluentIconBase` (such as `common.Folder`), or `nil`.
- Callbacks are usually exposed as fields (e.g. `OnClicked func()`); just assign them directly. Qt-style signal methods (e.g. `OnClicked(func)`) are also kept.

---

# 1. widgets Package

## button.go

### `PushButton`
- **Function**: A Fluent-style ordinary button used to trigger a single click action.
- **Constructor**:
  - `NewPushButton(parent *qt.QWidget)` / `NewPushButtonText(text string, parent *qt.QWidget)` / `NewPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)`
  - `icon`: icon source; `text`: button text; `parent`: parent widget.
- **Key methods**: `SetIcon(icon interface{})`, `Icon()/GetIcon() *qt.QIcon`; inherits `QPushButton`'s `SetText`, `OnClicked(func())`, `SetIconSize(*qt.QSize)`, `SetEnabled`, `SetCheckable`, `SetChecked`, etc.
- **Example**:
  ```go
  btn := widgets.NewPushButtonIcon(common.Accept, "确定", parent)
  btn.SetIconSize(qt.NewQSize2(16, 16))
  btn.OnClicked(func() { /* handle click */ })
  ```
- **Effect**: A rounded button with theme-color border/background and hover/pressed feedback; the icon is drawn to the left of the text.

### `PrimaryPushButton`
- **Function**: A primary-color (accent) filled button, used to highlight the main action.
- **Constructor**: `NewPrimaryPushButton(parent)` / `NewPrimaryPushButtonText(text, parent)` / `NewPrimaryPushButtonIcon(icon, text, parent)`.
- **Key methods**: Inherits `PushButton`.
- **Effect**: The button body is filled with the theme primary color; the icon/text are in inverted color; hover/press changes brightness.

### `TransparentPushButton`
- **Function**: A transparent-background button, suited to toolbars or scenes where the border only shows on hover.
- **Constructor**: `NewTransparentPushButton(parent)` / `...Text(text, parent)` / `...Icon(icon, text, parent)`.
- **Effect**: No background by default; a faint background only shows on hover/press.

### `ToggleButton`
- **Function**: A button with a toggle (checkable) state, used for on/off style options.
- **Constructor**: `NewToggleButton(parent)` / `...Text(text, parent)` / `...Icon(icon, text, parent)`.
- **Key methods**: Inherits `PushButton`; `SetChecked(bool)` / `IsChecked() bool`.
- **Effect**: Clicking toggles between selected/unselected; when selected it highlights with the theme color.

### `TransparentTogglePushButton`
- **Function**: A toggle button with a transparent background.
- **Constructor**: `NewTransparentTogglePushButton(parent)` / `...Text(text, parent)` / `...Icon(icon, text, parent)`.
- **Effect**: Transparent background, highlighted when selected; used in command bars/tool areas.

### `HyperlinkButton`
- **Function**: A link-styled button; clicking opens the URL in the system default browser.
- **Constructor**: `NewHyperlinkButton(parent)` / `NewHyperlinkButtonURL(url, text string, parent)` / `NewHyperlinkButtonIcon(icon interface{}, url, text string, parent)`.
- **Key methods**: `SetUrl(url string)` / `Url() *qt.QUrl`; inherits `PushButton`.
- **Effect**: Hand cursor, theme-color text/icon; clicking opens the link.

### `RadioButton`
- **Function**: A Fluent-style radio button; mutually exclusive within the same parent container.
- **Constructor**: `NewRadioButton(parent)` / `NewRadioButtonText(text string, parent)`.
- **Key methods**: `SetTextColor(light, dark *qt.QColor)`, `SetIndicatorColor(light, dark *qt.QColor)`; inherits `QRadioButton`'s `SetChecked`, `IsChecked`, `OnToggled(func(bool))`, `OnClicked(func())`.
- **Effect**: A circular indicator with different hover/press/selected colors; when selected, the interior is filled with the theme color.

### `ToolButton`
- **Function**: A Fluent-style tool button that shows only an icon (no text), usually used in toolbars.
- **Constructor**: `NewToolButton(parent)` / `NewToolButtonIcon(icon interface{}, parent)`.
- **Key methods**: `SetIcon(icon interface{})`, `Icon() *qt.QIcon`, `SetDrawIconFunc(func(icon interface{}, painter *qt.QPainter, rect *qt.QRectF))`; inherits `QToolButton`'s `OnClicked` etc.
- **Effect**: A centered icon button, transparent background by default with faint hover/press feedback.

### `TransparentToolButton` / `PrimaryToolButton`
- **Function**: Transparent-background / primary-filled tool buttons.
- **Constructor**: `NewTransparentToolButton(parent)` / `NewPrimaryToolButton(parent)` and the `...Icon(icon, parent)` variants.
- **Effect**: Transparent no-background / primary rounded background + inverted icon.

### `ToggleToolButton` / `TransparentToggleToolButton`
- **Function**: Tool buttons with a toggle (checkable) state.
- **Constructor**: `NewToggleToolButton(parent)` / `NewTransparentToggleToolButton(parent)` and the `...Icon(icon, parent)` variants.
- **Key methods**: Inherits `ToolButton`; `SetChecked(bool)` / `IsChecked() bool`.
- **Effect**: Clicking toggles the selected state; when selected, the icon is inverted and highlighted with the primary color.

### `DropDownPushButton`
- **Function**: A button with a drop-down arrow and popup menu; clicking the arrow on the right opens the menu.
- **Constructor**: `NewDropDownPushButton(parent)` / `...Text(text, parent)` / `...Icon(icon, text, parent)`.
- **Key methods**: `SetMenu(menu *RoundMenu)` / `Menu() *RoundMenu`; inherits `PushButton`.
- **Effect**: There's a downward arrow at the right end of the button; clicking the arrow first depresses and rebounds, then pops up the menu centered.

### `TransparentDropDownPushButton` / `PrimaryDropDownPushButton`
- **Function**: Transparent / primary-background drop-down buttons.
- **Constructor**: `NewTransparentDropDownPushButton(parent)` / `NewPrimaryDropDownPushButton(parent)` and the `...Text` / `...Icon` variants.

### `DropDownToolButton` / `TransparentDropDownToolButton` / `PrimaryDropDownToolButton`
- **Function**: Tool buttons with a drop-down arrow and menu (normal / transparent / primary).
- **Constructor**: `NewDropDownToolButton(parent)` etc. and the `...Icon(icon, parent)` variants.
- **Key methods**: `SetMenu(menu *RoundMenu)` / `Menu() *RoundMenu`.

### `SplitPushButton`
- **Function**: A split button composed of "main button + right drop-down arrow"; the main button triggers the primary action and the arrow pops a menu/flyout.
- **Constructor**: `NewSplitPushButton(parent)` / `...Text(text, parent)` / `...Icon(icon, text, parent)`.
- **Key methods**: `OnClicked(f func())`, `OnDropDownClicked(f func())`, `SetFlyout(flyout interface{})`, `SetText/SetIcon/SetIconSize`, `SetDropIcon/SetDropIconSize`.
- **Effect**: Left main button + right narrow arrow button; the two respond to clicks independently.

### `PrimarySplitPushButton`
- **Function**: A primary-filled split button.
- **Constructor**: `NewPrimarySplitPushButton(parent)` and the `...Text` / `...Icon` variants.

### `SplitToolButton` / `PrimarySplitToolButton`
- **Function**: Split tool buttons composed of "icon main tool button + drop-down arrow" (normal / primary).
- **Constructor**: `NewSplitToolButton(parent)` / `NewPrimarySplitToolButton(parent)` and the `...Icon(icon, parent)` variants.

### `PillPushButton` / `PillToolButton`
- **Function**: Capsule-shaped (fully rounded) toggle button / tool button.
- **Constructor**: `NewPillPushButton(parent)` / `NewPillToolButton(parent)` and the `...Text` / `...Icon` variants.
- **Key methods**: Inherits `ToggleButton` (including `SetChecked` / `IsChecked`).
- **Effect**: A capsule look with fully rounded sides; filled with the primary color when selected.

## card_widget.go

### `CardWidget`
- **Function**: A rounded card container with themed background and hover/press feedback; it can hold any child widget.
- **Constructor**: `NewCardWidget(parent *qt.QWidget)`.
- **Key methods**: `OnClicked(f func())`, `SetClickEnabled(bool)` / `IsClickEnabled()`, `SetBorderRadius(int)` / `BorderRadius()`.
- **Effect**: A semi-transparent white rounded card; brightens on hover, darkens on press; clickable.

### `SimpleCardWidget`
- **Function**: A flat-background card (hover/press look the same as the normal background), suited to static panels.
- **Constructor**: `NewSimpleCardWidget(parent)`.
- **Effect**: A rounded card with no interactive state change, with a thin border.

### `ElevatedCardWidget`
- **Function**: A card with a drop shadow on hover (no movement animation).
- **Constructor**: `NewElevatedCardWidget(parent)`.
- **Effect**: A soft shadow appears beneath the card on mouse hover.

### `CardSeparator`
- **Function**: A thin divider (3px high) used inside cards.
- **Constructor**: `NewCardSeparator(parent)`.

### `HeaderCardWidget`
- **Function**: A card with title bar + divider + content area.
- **Constructor**: `NewHeaderCardWidget(parent)` / `NewHeaderCardWidgetTitle(title string, parent)`.
- **Key methods**: `SetTitle(string)` / `Title()`.
- **Effect**: A bold title (15px) at the top, a divider and content area below, with rounded corners overall.

### `CardGroupWidget`
- **Function**: A row inside a grouped card (icon + title + secondary text + optional widget).
- **Constructor**: `NewCardGroupWidget(icon interface{}, title, content string, parent)`.
- **Key methods**: `SetTitle/Title`, `SetContent/Content`, `SetIcon/Icon`, `SetIconSize(*qt.QSize)`, `SetSeparatorVisible(bool)`, `AddWidget(widget, stretch int)`.
- **Effect**: A single settings row: an icon on the left, title + gray secondary text on the right; can attach a widget and a divider.

### `GroupHeaderCardWidget`
- **Function**: A titled grouped card that can hold multiple `CardGroupWidget` rows.
- **Constructor**: `NewGroupHeaderCardWidget(parent)`.
- **Key methods**: `AddGroup(icon, title, content, widget, stretch int) *CardGroupWidget`, `GroupCount()`, `SetTitle/Title`.
- **Effect**: Multiple settings rows arranged vertically under the title; a divider is automatically shown between rows.

## check_box.go

### `CheckBox`
- **Function**: A Fluent-style checkbox supporting selected / partially-checked (tri-state) display.
- **Constructor**: `NewCheckBox(parent)` / `NewCheckBoxText(text string, parent)`.
- **Key methods**: `SetCheckedColor(light, dark *qt.QColor)`, `SetTextColor(light, dark *qt.QColor)`; inherits `QCheckBox`'s `SetChecked`, `IsChecked`, `SetTristate`, `OnStateChanged(func(int))`, `OnToggled(func(bool))`.
- **Effect**: A checkmark (selected) or short dash (partially checked) is drawn inside the square indicator, with separate colors for hover/press/disabled states.

## combo_box.go

### `ComboBox`
- **Function**: A Fluent-style drop-down selector, implemented on top of a button + popup menu.
- **Constructor**: `NewComboBox(parent *qt.QWidget)`.
- **Key methods**:
  - Add/remove: `AddItem(text string, icon interface{}, userData interface{})`, `AddItems([]string)`, `InsertItem(index, text, icon, userData)`, `RemoveItem(index)`, `Clear()`.
  - Selection: `CurrentIndex()/SetCurrentIndex`, `CurrentText()/SetCurrentText`, `CurrentData()`.
  - Items: `ItemText/ItemData/ItemIcon`, `SetItemText/SetItemData/SetItemIcon/SetItemEnabled`, `FindText/FindData`, `Count()`.
  - Appearance: `SetPlaceholderText(text)`, `SetMaxVisibleItems(num)`.
  - Signals: `OnCurrentIndexChanged(func(int))`, `OnCurrentTextChanged(func(string))`, `OnActivated`, `OnTextActivated`.
- **Effect**: Clicking expands a centered popup menu; after selecting, the button text updates and the callback fires; it has a placeholder-text state and a drop-down arrow animation.

### `EditableComboBox`
- **Function**: An editable drop-down box implemented on top of `LineEdit`, allowing manual input and adding new items by pressing Enter.
- **Constructor**: `NewEditableComboBox(parent)`.
- **Key methods**: Essentially the same as `ComboBox`.
- **Effect**: An editable input box with a drop-down arrow on the right; if the entered text does not exist and Enter is pressed, it is automatically appended as an item.

## command_bar.go

### `CommandButton`
- **Function**: A single button in the command bar, self-drawing its icon and text.
- **Constructor**: `NewCommandButton(parent)`.
- **Key methods**: `SetText/Text`, `SetIcon/Icon`, `SetTight(bool)`, `SetToolButtonStyle(qt.ToolButtonStyle)`, `SetAction(*qt.QAction)`.
- **Effect**: A small button that can show icon-only / text-only / icon+text / icon-above-text layouts; inverted when selected.

### `MoreActionsButton`
- **Function**: The "more actions" button ("…", 40×34) shown when the command bar runs out of space.
- **Constructor**: `NewMoreActionsButton(parent)`.

### `CommandSeparator`
- **Function**: A vertical divider (9×34) between buttons in the command bar.
- **Constructor**: `NewCommandSeparator(parent)`.

### `CommandMenu`
- **Function**: The rounded menu (item height 32, icon 16×16) popped by the command bar's "more actions" button.
- **Constructor**: `NewCommandMenu(parent)`, inherits `RoundMenu`.

### `CommandViewMenu`
- **Function**: The "more actions" menu of `CommandBarView`, styled like a command bar.
- **Constructor**: `NewCommandViewMenu(parent)`; `SetDropDown(down, long bool)`.

### `CommandBar`
- **Function**: A horizontal command button bar; when space runs out, overflowing buttons are folded into a "more" menu.
- **Constructor**: `NewCommandBar(parent)`.
- **Key methods**: `AddAction(*qt.QAction) *CommandButton`, `AddActions`, `AddHiddenAction`, `InsertAction`, `RemoveAction`, `AddWidget`, `AddSeparator`, `SetSpacing`, `SetToolButtonStyle`, `SetButtonTight`, `SetIconSize`, `SetFont`, `SetMenuDropDown`, `SuitableWidth()`, `ResizeToSuitableWidth()`.
- **Effect**: Buttons are arranged horizontally; when width is insufficient the remaining buttons collapse into "…", which pops a menu when clicked.

### `CommandViewBar`
- **Function**: The command bar used inside `CommandBarView`; its "more" menu is attached to the outer view.
- **Constructor**: `NewCommandViewBar(parent)`, inherits `CommandBar`.

### `CommandBarView`
- **Function**: A self-drawn rounded-background command bar container that can be embedded in a `Flyout`.
- **Constructor**: `NewCommandBarView(parent)`.
- **Key methods**: Similar to `CommandBar` (`AddWidget/AddAction/AddActions/AddSeparator/RemoveAction/SetToolButtonStyle/SetButtonTight/SetIconSize/SetFont/SetMenuDropDown/SuitableWidth/ResizeToSuitableWidth` etc.), plus `Actions() []*qt.QAction`, `SetMenuVisible(bool)`.
- **Effect**: A command bar flyout with rounded light/dark background + thin border, suited as a contextual toolbar.

## cycle_list_widget.go

### `CycleListWidget`
- **Function**: A cyclically scrolling list with up/down scroll buttons and a centered selected row (mostly used for date/time pickers).
- **Constructor**: `NewCycleListWidget(items []string, itemSize *qt.QSize, align int, parent *qt.QWidget)`.
  - `items`: item texts; `itemSize`: single item size; `align`: text alignment (e.g. `int(qt.AlignCenter)`); `parent`: parent widget.
- **Key methods**: `SetItems([]string)`, `CurrentIndex()/SetCurrentIndex`, `CurrentItem()`, `SetSelectedItem(text)`, `ScrollUp/ScrollDown`, `SetScrollButtonRepeatEnabled(bool)`, `OnCurrentItemChanged(func(*qt.QListWidgetItem))`, `SetOnScrollChanged(func())`.
- **Effect**: A scrolling list with a centered highlight; items repeat beyond the visible area to support cyclic scrolling; up/down arrows appear on hover; smooth scrolling animation.

## flip_view.go

### `FlipView`
- **Function**: An image carousel/page-flip view, supporting horizontal or vertical paging and previous/next arrows.
- **Constructor**: `NewFlipView(parent)` (horizontal) / `NewFlipViewOrientation(orientation qt.Orientation, parent)`.
- **Key methods**: `AddImage(image interface{})`, `AddImages([]interface{})`, `SetItemImage(index, image)`, `Image(index) *qt.QImage`, `CurrentIndex()/SetCurrentIndex`, `ScrollPrevious/ScrollNext`, `OnCurrentIndexChanged(func(int))`, `SetItemSize(*qt.QSize)`, `SetBorderRadius(int)`, `SetAspectRatioMode(qt.AspectRatioMode)`, `IsHorizontal() bool`.
- **Effect**: Images are displayed at item size; previous/next arrows appear on mouse hover; wheel or arrow triggers smooth paging.

### `HorizontalFlipView` / `VerticalFlipView`
- **Function**: Horizontal / vertical flip views (convenient subclasses of `FlipView`).
- **Constructor**: `NewHorizontalFlipView(parent)` / `NewVerticalFlipView(parent)`.

## flyout.go

### `FlyoutIconWidget`
- **Function**: A small widget that draws a flyout icon in a 36×54 area.
- **Constructor**: `NewFlyoutIconWidget(icon interface{}, parent)`.

### `FlyoutViewBase`
- **Function**: Base class of flyout views, responsible for drawing the rounded background and border.
- **Constructor**: `NewFlyoutViewBase(parent)`.
- **Key methods**: `SetDrawBackground(bool)`, `BackgroundColor()`, `BorderColor()`.

### `FlyoutView`
- **Function**: The default flyout content view (title + content + icon + image + close button).
- **Constructor**: `NewFlyoutView(title, content string, icon, image interface{}, isClosable bool, parent)`.
- **Key methods**: `AddWidget(widget, stretch int, align qt.AlignmentFlag)`, `SetOnClosed(f func())`.
- **Effect**: Icon, title and body laid out vertically, with an optional top image and a top-right close button; text wraps automatically.

### `Flyout`
- **Function**: A popup container that hosts a `FlyoutViewBase` with a shadow and slide-in/fade-in animations.
- **Constructor**: `NewFlyout(view *FlyoutViewBase, parent *qt.QWidget, isDeleteOnClose bool)`.
- **Key methods**: `SetShadowEffect(blurRadius, dx, dy float64)`, `OnClosed(f func())`, `Exec(pos *qt.QPoint, aniType FlyoutAnimationType)`.
- **Effect**: A borderless popup window with a projection, appearing according to the chosen animation (drop-down / pop-up / slide left / slide right / fade in).

## frameless_window.go

### `FramelessWindow`
- **Function**: A resizable, draggable window without a system title bar (on Windows it implements native border resizing / title-bar dragging and DWM shadow).
- **Constructor**: `NewFramelessWindow(parent *qt.QWidget)`.
- **Key methods**: `SetResizeEnabled(bool)`, `SetMoveEnabled(bool)`, `SetTitleBar(*qt.QWidget)`, `SetTitleBarVisible(bool)`, `SetDoubleClickEnabled(bool)`, `SetShownHandler(f func())`.
- **Effect**: No native title bar, with system shadow/rounded corners (Win11); resizable from all eight edges and draggable via the title bar.

## icon_widget.go

### `IconWidget`
- **Function**: A general-purpose icon widget that draws an icon centered in its own area.
- **Constructor**: `NewIconWidget(parent)` / `NewIconWidgetIcon(icon interface{}, parent)`.
- **Key methods**: `SetIcon(icon interface{})`, `GetIcon()/Icon() *qt.QIcon`.

## info_badge.go

### `InfoBadge`
- **Function**: A small rounded label (badge) showing a count or status.
- **Constructor**: `NewInfoBadge(parent, level InfoLevel)` / `NewInfoBadgeText(text string, parent, level InfoLevel)`; `level` takes `InfoLevelInformation/Success/Attention/Warning/Error`.
- **Key methods**: `SetLevel(level)` / `Level()`, `SetCustomBackgroundColor(light, dark *qt.QColor)`; inherits `QLabel`'s `SetText`.
- **Effect**: A capsule/circle color block + centered text; color changes with level; a single character is automatically circular.

### `DotInfoBadge`
- **Function**: A small dot badge without text (4×4).
- **Constructor**: `NewDotInfoBadge(parent, level InfoLevel)`.

### `IconInfoBadge`
- **Function**: A small circular badge with an icon (16×16).
- **Constructor**: `NewIconInfoBadge(parent, level)` / `NewIconInfoBadgeIcon(icon interface{}, parent, level)`.
- **Key methods**: `SetIcon(icon)` / `Icon()`, `SetIconSize(*qt.QSize)` / `IconSize()`.

## info_bar.go

### `InfoIconWidget`
- **Function**: A small widget that draws an info-bar icon in a 36×36 area.
- **Constructor**: `NewInfoIconWidget(icon interface{}, parent)`.

### `InfoBar`
- **Function**: A temporary info bar with icon, title, body and an optional close button; can auto-dismiss.
- **Constructor**: `NewInfoBar(icon interface{}, title, content string, orient qt.Orientation, isClosable bool, duration int, position InfoBarPosition, parent *qt.QWidget)`.
  - `orient`: title/body orientation (`qt.Horizontal/Vertical`); `duration`: milliseconds before auto-close (`<0` means no auto-close); `position`: `InfoBarPositionTop/Bottom/TopLeft/...`.
- **Key methods**: `Show()` (slides in), `AddWidget(widget, stretch int)`, `SetCustomBackgroundColor(light, dark *qt.QColor)`, `OnClosed func()`.
- **Convenience functions**: `InfoBarInfo/InfoBarSuccess/InfoBarWarning/InfoBarError(title, content string, parent)` create and show an info bar of the corresponding level directly.
- **Effect**: An info bar slides in from the target position, fades out and closes after the duration; multiple bars at the same position stack automatically.

## label.go

### `PixmapLabel`
- **Function**: A label that renders high-DPI pixmaps.
- **Constructor**: `NewPixmapLabel(parent)`; `SetPixmap(*qt.QPixmap)`, `Pixmap()`.

### `CaptionLabel` / `BodyLabel` / `StrongBodyLabel`
- **Function**: 12px caption text / 14px body text / 14px bold body text labels.
- **Constructor**: `NewCaptionLabel(parent)` / `NewBodyLabel(parent)` / `NewStrongBodyLabel(parent)` and the `...Text(text, parent)` variants.
- **Key methods**: Inherits `FluentLabelBase`'s `SetText/Text`, `SetTextColor(light, dark *qt.QColor)`, `SetPixelFontSize(int)`, `SetStrikeOut/SetUnderline(bool)` etc.

### `SubtitleLabel` / `TitleLabel` / `LargeTitleLabel` / `DisplayLabel`
- **Function**: 20px subtitle / 28px title / 40px large title / 68px display-level title labels.
- **Constructor**: `NewSubtitleLabel(parent)` / `NewTitleLabel(parent)` / `NewLargeTitleLabel(parent)` / `NewDisplayLabel(parent)` and the `...Text` variants.

### `ImageLabel`
- **Function**: A label that renders an image, with support for independent per-corner rounding and a click callback.
- **Constructor**: `NewImageLabel(parent)` / `NewImageLabelImage(image interface{}, parent)` (`image`: string path, `*qt.QImage`, or `*qt.QPixmap`).
- **Key methods**: `SetImage(image)` / `Image()` / `IsNull()`, `SetBorderRadius(topLeft, topRight, bottomLeft, bottomRight int)`, `OnClicked(f func())`, `ScaledToWidth/Height(int)`, `SetScaledSize(*qt.QSize)`.

### `AvatarWidget`
- **Function**: A circular avatar widget supporting image avatars or text-initial avatars.
- **Constructor**: `NewAvatarWidget(parent)` / `NewAvatarWidgetImage(image interface{}, parent)`.
- **Key methods**: `SetRadius(radius int)` / `Radius()`, `SetImage(image)`, `SetBackgroundColor(light, dark *qt.QColor)` (text-avatar background color).

### `HyperlinkLabel`
- **Function**: A link-styled label button; clicking opens the URL.
- **Constructor**: `NewHyperlinkLabel(parent)` / `NewHyperlinkLabelText(text, parent)` / `NewHyperlinkLabelURL(url, text string, parent)`.
- **Key methods**: `SetUrl(url)` / `Url()`, `SetUnderlineVisible(bool)` / `IsUnderlineVisible()`.

## line_edit.go

### `LineEditButton`
- **Function**: An icon button (31×23) for use inside input boxes.
- **Constructor**: `NewLineEditButton(icon interface{}, parent)`.

### `LineEdit`
- **Function**: A Fluent-style single-line input box with an internal clear button and a focus underline.
- **Constructor**: `NewLineEdit(parent *qt.QWidget)`.
- **Key methods**: `SetClearButtonEnabled(bool)`, `SetError(bool)` / `IsError()`, `SetCustomFocusedBorderColor(light, dark *qt.QColor)`, `AddAction(action *qt.QAction, position int)` (`LineEditLeadingPosition` / `LineEditTrailingPosition`), `SetCompleter(*qt.QCompleter)`; inherits `QLineEdit`'s `Text/SetText/OnTextChanged/OnReturnPressed`.
- **Effect**: A rounded input box; a theme-color underline shows at the bottom when focused; a clear button shows when focused and it has content.

### `SearchLineEdit`
- **Function**: An input box with a search button.
- **Constructor**: `NewSearchLineEdit(parent)`.
- **Key methods**: `SearchSignal func(string)` (clicking search / pressing Enter), `ClearSignal func()`.

### `PasswordLineEdit`
- **Function**: A password input box with a "show/hide" view button.
- **Constructor**: `NewPasswordLineEdit(parent)`.
- **Key methods**: `SetPasswordVisible(bool)` / `IsPasswordVisible()`, `SetViewPasswordButtonVisible(bool)`.
- **Effect**: Input is masked with dots; holding the eye button on the right temporarily shows the plain text.

### `TextEdit` / `PlainTextEdit` / `TextBrowser`
- **Function**: Fluent-style multi-line rich text / plain text / read-only rich text browse boxes.
- **Constructor**: `NewTextEdit(parent)` / `NewPlainTextEdit(parent)` / `NewTextBrowser(parent)`, inheriting `QTextEdit` / `QPlainTextEdit` / `QTextBrowser` respectively.

## list_view.go

### `ListWidget`
- **Function**: A Fluent-style list widget with built-in smooth scrolling, hover/selected rounded highlighting and a left selection indicator bar.
- **Constructor**: `NewListWidget(parent *qt.QWidget)`.
- **Key methods**: `SetCurrentItem(*qt.QListWidgetItem)`, `SetCurrentRow(int)`, `SetCheckedColor(light, dark *qt.QColor)`, `SetSelectRightClickedRow(bool)`; inherits `QListWidget`'s `AddItem/AddItems/Count/CurrentRow` etc.

### `ListView`
- **Function**: A Fluent-style list view (based on `QListView`, used with a data model).
- **Constructor**: `NewListView(parent)`; `SetModel`, `SetSelectionMode`, etc.

## menu.go

### `RoundMenu`
- **Function**: A Fluent rounded popup menu with popup animation, submenus, icons, shortcuts, separators and custom widgets.
- **Constructor**: `NewRoundMenu(title string, parent *qt.QWidget)`.
- **Key methods**: `AddAction(*qt.QAction)`, `AddActionText(text) *qt.QAction`, `AddActionIcon(icon, text)`, `AddActions([]*qt.QAction)`, `InsertAction(before, action)`, `RemoveAction`, `AddMenu(*RoundMenu)`, `AddSeparator()`, `AddWidget(*qt.QWidget)`, `SetDefaultAction`, `SetIcon/SetTitle`, `SetItemHeight/SetMaxVisibleItems`, `Exec(pos *qt.QPoint, aniType MenuAnimationType)`, `ExecAt/ExecAtCentered`, `Clear()`, `OnClosed(func())`.
- **Effect**: Pops a rounded, shadowed menu from the trigger point with drop-down/pop-up/fade animations; expands on hover and shows a right arrow when it contains submenus.

### `CheckableMenu`
- **Function**: A rounded menu whose items are checkable; selected items show a check or radio indicator.
- **Constructor**: `NewCheckableMenu(title string, parent *qt.QWidget, indicatorType MenuIndicatorType)` (`MenuIndicatorCheck` / `MenuIndicatorRadio`).

## model_combo_box.go

### `ModelComboBox`
- **Function**: A model-driven combo box (the Go port is backed by `[]ComboItem`), supporting showing/hiding the current item's icon.
- **Constructor**: `NewModelComboBox(parent)`.
- **Key methods**: `SetIconVisible(bool)` / `IsIconVisible()`, `SetCurrentIndex(index)`, `SetItemIcon(index, icon)`, `Clear()`; inherits `ComboBox`.

### `EditableModelComboBox`
- **Function**: An editable model combo box (with a text input + drop-down list).
- **Constructor**: `NewEditableModelComboBox(parent)`; inherits `EditableComboBox`.

## pips_pager.go

### `PipsPager`
- **Function**: A paging indicator that represents the current page with a row/column of dots, with smooth scrolling and optional previous/next page buttons.
- **Constructor**: `NewPipsPager(parent)` (horizontal) / `NewPipsPagerOrientation(orientation qt.Orientation, parent)`.
- **Key methods**: `SetPageNumber(n int)` / `GetPageNumber()`, `SetCurrentIndex(index)` / `CurrentIndex()`, `ScrollNext/ScrollPrevious`, `SetVisibleNumber(n)`, `SetPreviousButtonDisplayMode/SetNextButtonDisplayMode(mode PipsScrollButtonDisplayMode)`, `OnCurrentIndexChanged(func(int))`.
- **Effect**: A row of dots represents the total page count; the current page's dot is enlarged and centered; switching scrolls smoothly.

### `HorizontalPipsPager` / `VerticalPipsPager`
- **Function**: Horizontal / vertical convenient subclasses of `PipsPager`.
- **Constructor**: `NewHorizontalPipsPager(parent)` / `NewVerticalPipsPager(parent)`.

## progress_bar.go

### `ProgressBar`
- **Function**: A Fluent-style progress bar supporting animation, pause/error states and custom colors.
- **Constructor**: `NewProgressBar(parent *qt.QWidget, useAni bool)`.
- **Key methods**: `SetVal(float64)` / `Val()`, `SetUseAni(bool)`, `SetCustomBarColor(light, dark *qt.QColor)`, `SetCustomBackgroundColor(light, dark *qt.QColor)`, `Pause/Resume/SetPaused/IsPaused`, `Error/SetError/IsError`; inherits `QProgressBar`'s `SetRange/SetValue/SetFormat`.
- **Effect**: A 4px thin progress bar filled with the theme color according to progress; turns yellow on pause, red on error.

### `IndeterminateProgressBar`
- **Function**: An indeterminate progress bar showing a back-and-forth scanning animation.
- **Constructor**: `NewIndeterminateProgressBar(parent *qt.QWidget, start bool)`.
- **Key methods**: `Start/Stop/IsStarted`, `Pause/Resume/SetPaused/IsPaused`, `Error/SetError/IsError`, `SetCustomBarColor(light, dark *qt.QColor)`.

## progress_ring.go

### `ProgressRing`
- **Function**: A circular progress indicator whose arc fills according to progress, optionally showing a percentage text in the center.
- **Constructor**: `NewProgressRing(parent *qt.QWidget, useAni bool)`.
- **Key methods**: `SetStrokeWidth(int)` / `StrokeWidth()`; inherits `ProgressBar`'s `SetRange/SetValue/SetTextVisible/SetFormat/Pause/Error/SetCustomBarColor` etc.
- **Effect**: A fixed 100×100 ring; the arc fills clockwise from the top.

### `IndeterminateProgressRing`
- **Function**: A rotating indeterminate progress ring (80×80).
- **Constructor**: `NewIndeterminateProgressRing(parent *qt.QWidget, start bool)`.
- **Key methods**: `Start/Stop`, `SetStrokeWidth(int)`, `SetCustomBarColor(light, dark *qt.QColor)`, `SetCustomBackgroundColor(light, dark *qt.QColor)`.

## scroll_area.go

### `ScrollArea`
- **Function**: A scroll area with smooth scrolling and Fluent overlay scrollbars.
- **Constructor**: `NewScrollArea(parent *qt.QWidget)`.
- **Key methods**: `SetVerticalScrollBarPolicy/SetHorizontalScrollBarPolicy(policy qt.ScrollBarPolicy)`, `SetSmoothMode(mode common.SmoothMode, orientation qt.Orientation)`, `EnableTransparentBackground()`, `ScrollDelegate()`; inherits `QScrollArea`'s `SetWidget/SetWidgetResizable`.

### `SingleDirectionScrollArea`
- **Function**: A scroll area that scrolls in a single direction only.
- **Constructor**: `NewSingleDirectionScrollArea(parent *qt.QWidget, orient qt.Orientation)`.

### `SmoothScrollArea`
- **Function**: A smooth scroll area using an animated scrollbar (value changes have easing animation).
- **Constructor**: `NewSmoothScrollArea(parent *qt.QWidget)`.
- **Key methods**: `SetScrollAnimation(orient qt.Orientation, duration int)`, `EnableTransparentBackground()`.

## scroll_bar.go

### `ScrollBar`
- **Function**: A Fluent custom scrollbar independent of the native one; synchronizes value/range bidirectionally with the target scroll area.
- **Constructor**: `NewScrollBar(orient qt.Orientation, parent *qt.QAbstractScrollArea)`.
- **Key methods**: `SetValue/Value`, `SetRange/SetMinimum/SetMaximum`, `SetPageStep/SetSingleStep`, `SetHandleColor/SetArrowColor/SetGrooveColor(light, dark *qt.QColor)`, `SetHandleDisplayMode(mode ScrollBarHandleDisplayMode)`, `SetForceHidden(bool)`, `Expand/Collapse`; signals `OnValueChanged(func(int))`, `OnRangeChanged`, `OnSliderPressed/Released/Moved`.
- **Effect**: A thin scrollbar floating on the edge of the scroll area; expands the groove and shows arrows/handle on hover.

### `SmoothScrollBar`
- **Function**: A scrollbar whose value changes have OutCubic easing animation (the smooth version of `ScrollBar`).
- **Constructor**: `NewSmoothScrollBar(orient qt.Orientation, parent *qt.QAbstractScrollArea)`.
- **Key methods**: `SetValue(value)` (animated scroll), `ScrollValue(value)`, `ScrollTo(value)`, `ResetValue(value)`, `SetScrollAnimation(duration int)`.

## separator.go

### `HorizontalSeparator` / `VerticalSeparator`
- **Function**: Thin horizontal / vertical lines (fixed 3px) used to separate content.
- **Constructor**: `NewHorizontalSeparator(parent)` / `NewVerticalSeparator(parent)`.

## slider.go

### `Slider`
- **Function**: A Fluent-style slider; clicking/dragging the track jumps directly, with a round handle.
- **Constructor**: `NewSlider(parent)` (horizontal) / `NewSliderOrientation(orientation qt.Orientation, parent)`.
- **Key methods**: `OnClicked(func(value int))`, `SetThemeColor(light, dark *qt.QColor)`, `SetOrientation(orientation qt.Orientation)`; inherits `QSlider`'s `SetRange/SetValue/Value/OnValueChanged`.
- **Effect**: Rounded track + round handle; the traversed part fills with the theme color; clicking the track jumps directly to the corresponding value.

### `ClickableSlider`
- **Function**: A slider that preserves native mouse handling and jumps on click.
- **Constructor**: `NewClickableSlider(parent)`; `OnClicked(func(value int))`.

## spin_box.go

### `SpinBox` / `DoubleSpinBox`
- **Function**: Fluent-style integer / floating-point input boxes with inline increment/decrement buttons on the right.
- **Constructor**: `NewSpinBox(parent)` / `NewDoubleSpinBox(parent)`.
- **Key methods**: `SetAccelerated(bool)`, `SetReadOnly(bool)`, `SetSymbolVisible(bool)`, `SetError(bool)` / `IsError()`, `SetCustomFocusedBorderColor(light, dark *qt.QColor)`; inherits `QSpinBox` / `QDoubleSpinBox`'s `SetRange/SetValue/Value/SetDecimals` etc.

### `DateEdit` / `DateTimeEdit` / `TimeEdit`
- **Function**: Fluent-style date / date-time / time input boxes with inline increment/decrement buttons on the right.
- **Constructor**: `NewDateEdit(parent)` / `NewDateTimeEdit(parent)` / `NewTimeEdit(parent)`, inheriting `QDateEdit` / `QDateTimeEdit` / `QTimeEdit` respectively.

### `CompactSpinBox` / `CompactDoubleSpinBox` / `CompactDateEdit` / `CompactDateTimeEdit` / `CompactTimeEdit`
- **Function**: Compact input boxes, with a single combined up/down arrow button on the right (upper half increments, lower half decrements).
- **Constructor**: `NewCompactSpinBox(parent)` etc.

## stacked_widget.go

### `OpacityAniStackedWidget`
- **Function**: A stacked container whose new page fades in when switching.
- **Constructor**: `NewOpacityAniStackedWidget(parent)`; `AddWidget(widget)`, `SetCurrentIndex(index)`, `SetCurrentWidget(widget)`.

### `PopUpAniStackedWidget`
- **Function**: A stacked container whose new page fades in when switching (WinUI pop-up approximated as fade-in), keeping animation start/finish callbacks.
- **Constructor**: `NewPopUpAniStackedWidget(parent)`.
- **Key methods**: `AddWidget(widget, deltaX, deltaY int)`, `RemoveWidget`, `SetCurrentIndex/SetCurrentWidget`, `SetAnimationEnabled(bool)`, `OnAniStart/OnAniFinished(func())`.

### `TransitionStackedWidget`
- **Function**: Base class of transition-animation stacked containers (approximated as fade-in).
- **Constructor**: `NewTransitionStackedWidget(parent)`.
- **Key methods**: `AddWidget(widget) int`, `InsertWidget(index, widget) int`, `SetCurrentIndex(index, duration int, isBack bool)`, `SetCurrentWidget(widget, duration, isBack)`, `SetAnimationEnabled(bool)`, `OnAniStart/OnAniFinished`.

### `EntranceTransitionStackedWidget` / `DrillInTransitionStackedWidget`
- **Function**: Entrance / drill-in transition stacked containers (approximated as fade-in).
- **Constructor**: `NewEntranceTransitionStackedWidget(parent)` / `NewDrillInTransitionStackedWidget(parent)`; inherit `TransitionStackedWidget`.

## state_tool_tip.go

### `StateToolTip`
- **Function**: A status tooltip shown in a window corner; shows a spinner while in progress, and a checkmark with auto fade-out on completion.
- **Constructor**: `NewStateToolTip(title, content string, parent *qt.QWidget)`.
- **Key methods**: `SetTitle(title)`, `SetContent(content)`, `SetState(isDone bool)` (when true, fades out and destroys after 1 second), `OnClosed(func())`, `GetSuitablePos() *qt.QPoint`.
- **Effect**: A small card appears in the top-right corner of the window; a spinner on the left indicates processing; it turns into a checkmark and fades out on completion.

## switch_button.go

### `SwitchButton`
- **Function**: A Fluent switch button with an optional text label.
- **Constructor**: `NewSwitchButton(parent *qt.QWidget, indicatorPos IndicatorPosition)` / `NewSwitchButtonText(text string, parent, indicatorPos)`; `indicatorPos` takes `LEFT` / `RIGHT`.
- **Key methods**: `IsChecked/SetChecked/ToggleChecked`, `SetText/Text`, `SetOnText/OnText`, `SetOffText/OffText`, `SetSpacing(int)`, `SetTextColor(light, dark *qt.QColor)`, `SetCheckedIndicatorColor(light, dark *qt.QColor)`, `OnCheckedChanged(func(bool))`.
- **Effect**: A capsule-shaped switch with sliding-thumb animation; the side text toggles between on/off text with the state.

## tab_view.go

### `TabItem`
- **Function**: A single tab button with a close button, selected state, icon and text.
- **Constructor**: `NewTabItem(text string, parent *qt.QWidget, icon interface{})`.
- **Key methods**: `SetRouteKey(key)` / `RouteKey()`, `SetSelected(bool)`, `SetBorderRadius(int)`, `SetShadowEnabled(bool)`, `SetCloseButtonDisplayMode(mode TabCloseButtonDisplayMode)`, `SetTextColor(*qt.QColor)`, `SetSelectedBackgroundColor(light, dark *qt.QColor)`, `SetOnClosed(func())`, `SetOnDoubleClicked(func())`.

### `TabBar`
- **Function**: A scrollable, drag-reorderable tab bar hosting multiple `TabItem`s.
- **Constructor**: `NewTabBar(parent *qt.QWidget)`.
- **Key methods**: `AddTab(routeKey, text string, icon interface{}, onClick func()) *TabItem`, `InsertTab`, `RemoveTab/RemoveTabByKey`, `SetCurrentIndex/SetCurrentTab/CurrentIndex/CurrentTab`, `SetAddButtonVisible`, `SetMovable/SetScrollable/SetTabsClosable`, `SetCloseButtonDisplayMode`, `SetTabIcon/SetTabText/SetTabToolTip/SetTabEnabled/SetTabVisible/SetTabData`, `SetTabSelectedBackgroundColor/SetTabShadowEnabled`, `SetTabMaximumWidth/SetTabMinimumWidth`, `TabItem(index)/Tab(routeKey)/Count/Clear`; signals `OnCurrentChanged/OnTabBarClicked/OnTabBarDoubleClicked/OnTabCloseRequested/OnTabAddRequested/OnTabMoved`.
- **Effect**: A horizontally scrollable tab bar; tabs can be drag-reordered and closed; the selected tab is highlighted; there's an add button at the far left.

### `TabWidget`
- **Function**: A complete tab control combining a tab bar with stacked pages.
- **Constructor**: `NewTabWidget(parent *qt.QWidget)`.
- **Key methods**: `AddTab(widget *qt.QWidget, label string, icon interface{}, routeKey string) int`, `AddPage`, `InsertTab`, `RemoveTab/Clear`, `SetCurrentIndex/SetCurrentWidget/CurrentIndex/CurrentWidget`, `Widget(index)/Count`; forwards the various `SetXxx` to `TabBar`; signals `OnCurrentChanged/OnTabBarClicked/OnTabCloseRequested/OnTabAddRequested/OnTabBarDoubleClicked`.
- **Effect**: A tab bar on top + stacked pages below; clicking a tab automatically switches to the corresponding page; tabs can be closed/dragged.

## table_view.go

### `TableWidget`
- **Function**: A Fluent-style table widget (based on `QTableWidget`) with rounded row highlighting and a left selection indicator bar.
- **Constructor**: `NewTableWidget(parent *qt.QWidget)`.
- **Key methods**: `SetBorderVisible(bool)`, `SetBorderRadius(int)`, `SetCheckedColor(light, dark *qt.QColor)`, `SetCurrentCell(row, column)`, `SetSelectRightClickedRow(bool)`; inherits `QTableWidget`'s `SetRowCount/SetColumnCount/SetItem/SetHorizontalHeaderLabels`.
- **Effect**: A Fluent table without grid lines; the entire row is selected and a theme-color indicator bar shows on the left.

### `TableView`
- **Function**: A Fluent-style table view (based on `QTableView`, used with a data model).
- **Constructor**: `NewTableView(parent *qt.QWidget)`; `SetModel`, etc.

## teaching_tip.go

### `TeachingTip`
- **Function**: A teaching-tip bubble anchored to a target widget, with an adjustable tail and auto fade-out.
- **Constructor**: `NewTeachingTip(view *FlyoutViewBase, target *qt.QWidget, duration int, tailPosition TeachingTipTailPosition, parent *qt.QWidget, isDeleteOnClose bool)`.
  - `tailPosition`: `TeachingTipTailTop/Bottom/Left/Right/...` or `TeachingTipTailNone`.
- **Key methods**: `SetView(view)` / `View()`, `SetShadowEffect(blurRadius, dx, dy)`, `OnClosed(func())`.
- **Convenience functions**: `TeachingTipMake(view, target, duration, tailPosition, parent, isDeleteOnClose)`, `TeachingTipCreate(target, title, content, icon, image, isClosable, duration, tailPosition, parent, isDeleteOnClose)`.
- **Effect**: A bubble with a tail pops up beside the target widget; it fades in then fades out automatically after the duration, with the tail pointing at the target.

### `PopupTeachingTip`
- **Function**: A `TeachingTip` displayed as a Popup window.
- **Constructor**: `NewPopupTeachingTip(...)` with the same parameters as `TeachingTip`.

## tool_tip.go

### `ToolTip`
- **Function**: A Fluent-style tooltip popup, rounded, shadowed, auto-hidable.
- **Constructor**: `NewToolTip(text string, parent *qt.QWidget)`.
- **Key methods**: `SetText(text)` / `Text()`, `SetDuration(int)` / `Duration()`, `AdjustPos(widget *qt.QWidget, position ToolTipPosition)` (`ToolTipPositionTop/Bottom/Left/Right/TopLeft/...`); inherits `QFrame`'s `Show/Hide/Move`.

## tree_view.go

### `TreeWidget`
- **Function**: A Fluent-style tree widget (based on `QTreeWidget`) with rounded highlighting and theme-color expand arrows.
- **Constructor**: `NewTreeWidget(parent *qt.QWidget)`.
- **Key methods**: `SetBorderVisible(bool)`, `SetBorderRadius(int)`, `SetCheckedColor(light, dark *qt.QColor)`; inherits `QTreeWidget`'s `SetHeaderLabels/AddTopLevelItem/SetItemWidget`.

### `TreeView`
- **Function**: A Fluent-style tree view (based on `QTreeView`, used with a data model).
- **Constructor**: `NewTreeView(parent *qt.QWidget)`; `SetModel`, etc.

---

# 2. navigation Package

## breadcrumb.go

### `ElideButton`
- **Function**: The "…" ellipsis button shown when breadcrumb content overflows; clicking pops a menu of hidden items.
- **Constructor**: `NewElideButton(parent *qt.QWidget)`.
- **Key methods**: `ClearState()`, `OnClicked(func())`.

### `BreadcrumbItem`
- **Function**: A clickable hierarchical label item in breadcrumb navigation.
- **Constructor**: `NewBreadcrumbItem(routeKey, text string, index int, parent *qt.QWidget)`.
- **Key methods**: `RouteKey()`, `Text()/SetText`, `IsRoot()`, `SetSelected(bool)`, `SetFont(*qt.QFont)`, `SetSpacing(int)`, `OnClicked(func())`.
- **Effect**: The root item shows its text directly; non-root items have a gray ">" arrow in front, brighten on hover, darken on click.

### `BreadcrumbBar`
- **Function**: A horizontal breadcrumb navigation bar that shows the route path in hierarchical order, supporting click-back and overflow collapsing.
- **Constructor**: `NewBreadcrumbBar(parent *qt.QWidget)`.
- **Key methods**: `OnCurrentItemChanged(func(string))`, `OnCurrentIndexChanged(func(int))`, `AddItem(routeKey, text string)`, `SetCurrentIndex(index)`, `SetCurrentItem(routeKey)`, `SetItemText(routeKey, text)`, `Item(routeKey)/ItemAt(index)`, `CurrentIndex()/CurrentItem()`, `Clear()`, `PopItem()`, `Count()`, `SetFont`, `SetSpacing/Spacing`.
- **Effect**: Shows "Home > Settings > Profile" from left to right; clicking any item selects it and removes the levels after it; when the total width overflows, middle items collapse into "…".

## navigation_bar.go

### `NavigationBarPushButton`
- **Function**: A large button in the bottom navigation bar; icon on top, text below; on selection the icon moves down and an indicator bar appears.
- **Constructor**: `NewNavigationBarPushButton(icon interface{}, text string, isSelectable bool, selectedIcon interface{}, parent *qt.QWidget)`.
- **Key methods**: `SetSelectedColor(light, dark interface{})`, `SetSelectedIcon(icon)`, `SetSelectedTextVisible(bool)`, `IndicatorRect() *qt.QRectF`, `SetSelected(bool)`; inherits `NavigationPushButton`'s `Text/SetText/SetIcon/OnClicked(func(bool))`.
- **Effect**: Fixed 64×58; unselected icon is semi-transparent; on selection the rounded background highlights, a 4×24 indicator bar appears on the left, the icon moves down 6px and switches to the selected icon.

### `NavigationBar`
- **Function**: A vertical navigation bar (often used for bottom-navigation layouts) with a top area, scrollable area and bottom area, plus a sliding selection indicator.
- **Constructor**: `NewNavigationBar(parent *qt.QWidget)`.
- **Key methods**: `AddItem(routeKey, icon, text string, onClick func(bool), selectable bool, selectedIcon interface{}, position NavigationItemPosition)`, `AddWidget`, `InsertItem/InsertWidget`, `RemoveWidget`, `SetCurrentItem(routeKey)`, `SetFont`, `SetSelectedTextVisible`, `SetSelectedColor`, `Buttons()`, `SetIndicatorAnimationEnabled(bool)`.
- **Effect**: Buttons are arranged vertically by `Top/Scroll/Bottom` position; clicking a selectable item slides the indicator smoothly; the selected button's icon moves down and highlights.

## navigation_interface.go

### `NavigationInterface`
- **Function**: A high-level wrapper over `NavigationPanel` providing convenient APIs for adding navigation items, separators, group headers, user cards, etc.
- **Constructor**: `NewNavigationInterface(parent *qt.QWidget, showMenuButton, showReturnButton, collapsible bool)`.
- **Key methods**: `OnDisplayModeChanged(func(NavigationDisplayMode))`, `AddItem(routeKey, icon, text, onClick func(bool), selectable bool, position, tooltip, parentRouteKey string)`, `AddWidget`, `InsertItem/InsertWidget`, `AddSeparator/InsertSeparator`, `AddItemHeader/InsertItemHeader`, `AddUserCard(...)`, `RemoveWidget/SetCurrentItem/Widget`, `Expand/Toggle/SetExpandWidth/SetMinimumExpandWidth`, `SetMenuButtonVisible/SetReturnButtonVisible/SetCollapsible`, `IsAcrylicEnabled/SetAcrylicEnabled` (no-op), `SetIndicatorAnimationEnabled` etc.
- **Effect**: A collapsible side navigation panel; clicking the menu button switches between "expanded / collapsed / menu flyout"; the return button enables/disables with route history; the selected item has a sliding indicator.

## navigation_panel.go

### `NavigationPanel`
- **Function**: A `QFrame`-based collapsible navigation panel (the low-level implementation of `NavigationInterface`), supporting multi-level tree items and expand/collapse animations.
- **Constructor**: `NewNavigationPanel(parent *qt.QWidget, isMinimalEnabled bool)`.
- **Key methods**: `OnDisplayModeChanged`, `Widget(routeKey)`, `AddItem/AddWidget/InsertItem/InsertWidget`, `AddSeparator/InsertSeparator`, `AddItemHeader/InsertItemHeader`, `RemoveWidget`, `SetMenuButtonVisible/SetReturnButtonVisible/SetCollapsible`, `SetExpandWidth/SetMinimumExpandWidth`, `SetAcrylicEnabled/IsAcrylicEnabled` (no-op), `Expand/Collapse/Toggle/IsCollapsed`, `SetCurrentItem`, `SetIndicatorAnimationEnabled`, `LayoutMinHeight()`.
- **Effect**: Default compact state is 48px wide (icons only); text appears after expanding; child items with `parentRouteKey` are attached to the parent as tree nodes; pops as a "menu flyout" mode when the window is too narrow.

## navigation_widget.go

### `NavigationPushButton`
- **Function**: A selectable navigation item with icon and text in side navigation (the base control of all navigation items).
- **Constructor**: `NewNavigationPushButton(icon interface{}, text string, isSelectable bool, parent *qt.QWidget)`.
- **Key methods**: `Text/SetText`, `Icon/SetIcon`, `OnClicked(func(bool))`, `OnSelectedChanged(func(bool))`, `SetSelected(bool)`, `SetAboutSelected(bool)`, `SetCompacted(bool)`, `SetTextColor(light, dark interface{})`, `SetIndicatorColor(light, dark interface{})`.
- **Effect**: Compact state shows only the icon; expanded state shows text to the right of the icon; a rounded background appears on hover; a 3×16 indicator bar shows on the left when selected.

### `NavigationToolButton`
- **Function**: A square tool button (40×36) used for the panel's menu/return buttons.
- **Constructor**: `NewNavigationToolButton(icon interface{}, parent)`.

### `NavigationSeparator`
- **Function**: A thin horizontal divider between navigation items.
- **Constructor**: `NewNavigationSeparator(parent)`.

### `NavigationItemHeader`
- **Function**: A non-clickable group header used to group navigation items.
- **Constructor**: `NewNavigationItemHeader(text string, parent)`.
- **Key methods**: `Text/SetText`, `SetCompacted(bool)`.

### `NavigationTreeWidget`
- **Function**: A collapsible multi-level tree node in the navigation panel; clicking expands/collapses children.
- **Constructor**: `NewNavigationTreeWidget(icon interface{}, text string, isSelectable bool, parent *qt.QWidget)`.
- **Key methods**: `OnExpanded(func())`, `AddChild/InsertChild/RemoveChild`, `ChildItems()`, `Text/SetText`, `Icon/SetIcon`, `SetTextColor/SetIndicatorColor`, `SetFont`, `SetExpanded(bool, ani bool)`, `IsRoot/IsLeaf`, `SetSelected/SetCompacted/SetAboutSelected`, `SetRememberExpandState/SaveExpandState/RestoreExpandState`, `Clone()`, `SuitableWidth()`.
- **Effect**: Items with children show a rotating arrow on the right; children are indented by depth; when the panel is collapsed, clicking pops a flyout submenu.

### `NavigationAvatarWidget`
- **Function**: A navigation control showing an avatar and an optional name.
- **Constructor**: `NewNavigationAvatarWidget(name string, avatar interface{}, parent)`.
- **Key methods**: `SetName(name)`, `SetAvatar(avatar)`.

### `NavigationUserCard`
- **Function**: A user card with avatar, title and subtitle; avatar radius and text opacity animate together between compact/expanded states.
- **Constructor**: `NewNavigationUserCard(parent *qt.QWidget)`.
- **Key methods**: `SetAvatarIcon(icon)`, `SetAvatarBackgroundColor(light, dark *qt.QColor)`, `Title/SetTitle`, `Subtitle/SetSubtitle`, `SetTitleFontSize/SetSubtitleFontSize`, `SetAnimationDuration(int)`, `SetCompacted(bool)`.
- **Effect**: Compact state shows a 40×36 circular avatar; after expanding, the avatar enlarges to radius 32, with a bold title and subtitle on the right.

### `NavigationIndicator`
- **Function**: The sliding selection indicator bar (a small 3×16 rounded bar) in the navigation panel/bar.
- **Constructor**: `NewNavigationIndicator(parent *qt.QWidget)`.
- **Key methods**: `OnAniFinished(func())`, `StartAnimation(startRect, endRect *qt.QRectF, useCrossFade bool)`, `StopAnimation()`, `SetIndicatorColor(light, dark interface{})`.

### `NavigationFlyoutMenu`
- **Function**: A flyout submenu popped when a collapsed root node is clicked, showing the cloned subtree.
- **Constructor**: `NewNavigationFlyoutMenu(tree *NavigationTreeWidget, parent)`.
- **Key methods**: `OnExpanded(func())`; inherits `widgets.ScrollArea`.

## pivot.go

### `PivotItem`
- **Function**: A selectable Pivot segment item based on `PushButton` (large 18px font).
- **Constructor**: `NewPivotItem(text string, parent)`.
- **Key methods**: `OnItemClicked(func(bool))`, `SetSelected(bool)`.

### `Pivot`
- **Function**: A horizontal navigation control (like WinUI Pivot tabs) with a sliding indicator at the bottom.
- **Constructor**: `NewPivot(parent *qt.QWidget)`.
- **Key methods**: `OnCurrentItemChanged(func(string))`, `AddItem(routeKey, text string, onClick func(bool), icon interface{})`, `AddWidget`, `InsertItem/InsertWidget`, `RemoveWidget/Clear`, `CurrentItem/CurrentRouteKey`, `SetCurrentItem(routeKey)`, `SetIndicatorLength(int)`, `SetItemFontSize(int)`, `SetItemText(routeKey, text)`, `SetIndicatorColor(light, dark interface{})`.
- **Effect**: Horizontally arranged tabs with a 16×3 indicator bar at the bottom, sliding with a squash-and-stretch animation when switching.

## segmented_widget.go

### `SegmentedItem` / `SegmentedToolItem` / `SegmentedToggleToolItem`
- **Function**: A segmented item (14px font), an icon-only tool item, and a checkable tool item.
- **Constructor**: `NewSegmentedItem(text string, parent)` / `NewSegmentedToolItem(icon interface{}, parent)` / `NewSegmentedToggleToolItem(icon interface{}, parent)`.

### `SegmentedWidget`
- **Function**: A segmented control (subclass of `Pivot`); the selected background capsule and the indicator slide horizontally.
- **Constructor**: `NewSegmentedWidget(parent *qt.QWidget)`.
- **Key methods**: `AddItem(routeKey, text string, onClick func(bool), icon interface{})`, `InsertItem`, `SetCurrentItem(routeKey)`, `CurrentIndicatorGeometry()`; inherits `Pivot`.
- **Effect**: Equal-width segments share a rounded selected capsule background; the capsule and indicator slide horizontally when switching.

### `SegmentedToolWidget` / `SegmentedToggleToolWidget`
- **Function**: Segmented controls made of icon tool buttons (normal / checkable).
- **Constructor**: `NewSegmentedToolWidget(parent)` / `NewSegmentedToggleToolWidget(parent)`; `AddItem(routeKey, icon interface{}, onClick func(bool))`, `InsertItem`.

---

# 3. dialog_box Package

## dialog.go

### `Dialog`
- **Function**: A borderless message dialog with a title, body and OK/Cancel buttons.
- **Constructor**: `NewDialog(title, content string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnYes func()`, `OnCancel func()`, `SetTitleBarVisible(bool)`, `SetContentCopyable(bool)`, `HideYesButton()/HideCancelButton()`.
- **Effect**: A fixed-size, borderless modal dialog; the title bar can drag the window; the body wraps automatically according to the window width.

## message_box_base.go

### `MessageBox`
- **Function**: A message box with a semi-transparent full-screen mask; a centered card shows title/body and OK/Cancel buttons.
- **Constructor**: `NewMessageBox(title, content string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnYes func()`, `OnCancel func()`, `SetContentCopyable(bool)`, `HideYesButton/HideCancelButton`; inherits `MaskDialogBase`.
- **Effect**: A semi-transparent mask over the parent window with a rounded, shadowed card floating in the center; supports closing and dragging by clicking the mask.

### `MessageBoxBase`
- **Function**: The base class of mask dialogs with a title/OK/Cancel button bar; the content area is left for subclasses to fill via `ViewLayout()`.
- **Constructor**: `NewMessageBoxBase(parent *qt.QWidget)`.
- **Key methods/signals**: `ValidateFunc func() bool`, `Validate() bool`, `ViewLayout() *qt.QVBoxLayout`, `HideYesButton/HideCancelButton`.

### `MessageDialog`
- **Function**: A Win10-style mask message dialog.
- **Constructor**: `NewMessageDialog(title, content string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnYes func()`, `OnCancel func()`.

## mask_dialog_base.go

### `MaskDialogBase`
- **Function**: The base class of all mask dialogs — borderless, semi-transparent background, covering the entire parent window, with the content placed in a centered `widget` frame.
- **Constructor**: `NewMaskDialogBase(parent *qt.QWidget)`.
- **Key methods**: `SetShadowEffect(blurRadius, offsetX, offsetY float64, color *qt.QColor)`, `SetMaskColor(color)`, `SetClosableOnMaskClicked(bool)`, `SetDraggable(bool)`, `Widget() *qt.QFrame`, `WindowMask() *qt.QWidget`, `HBoxLayout() *qt.QHBoxLayout`.
- **Effect**: A transparent mask dialog covering the parent window, with centered content having fade-in/fade-out animation and a projection.

## color_dialog.go

### `HuePanel`
- **Function**: A 256×256 hue/saturation picker panel; click or drag to choose a color.
- **Constructor**: `NewHuePanel(color *qt.QColor, parent *qt.QWidget)`.
- **Key methods/signals**: `OnColorChanged func(color *qt.QColor)`, `SetColor(color)`.

### `BrightnessSlider`
- **Function**: A horizontal clickable slider (range 0–255) controlling color brightness (Value).
- **Constructor**: `NewBrightnessSlider(color *qt.QColor, parent *qt.QWidget)`.
- **Key methods/signals**: `OnColorChanged func(color *qt.QColor)`, `SetColor(color)`.

### `ColorCard`
- **Function**: A 44×128 color swatch used to preview colors (with a checkerboard transparent base when alpha is enabled).
- **Constructor**: `NewColorCard(color *qt.QColor, parent *qt.QWidget, enableAlpha bool)`.
- **Key methods/signals**: `SetColor(color)`.

### `ColorLineEdit`
- **Function**: A color-channel input box restricted to integers in the range 0–255.
- **Constructor**: `NewColorLineEdit(value int, parent *qt.QWidget)`.
- **Key methods/signals**: `OnValueChanged func(text string)` (fired for valid integers).

### `HexColorLineEdit`
- **Function**: An input box for editing a 6-digit (RGB) or 8-digit (ARGB) hex color (without the `#` prefix).
- **Constructor**: `NewHexColorLineEdit(color *qt.QColor, parent *qt.QWidget, enableAlpha bool)`.
- **Key methods/signals**: `SetColor(color)`, `OnValueChanged func(text string)`.

### `OpacityLineEdit`
- **Function**: An input box for editing opacity as a 0–100 percentage.
- **Constructor**: `NewOpacityLineEdit(value int, parent *qt.QWidget)` (`value` is 0–255 alpha; internally converted to a percentage).
- **Key methods/signals**: `OnValueChanged func(text string)`.

### `ColorDialog`
- **Function**: A complete color-selection dialog integrating a hue panel, brightness slider, RGB/Hex/opacity inputs and old/new color previews.
- **Constructor**: `NewColorDialog(color *qt.QColor, title string, parent *qt.QWidget, enableAlpha bool)`.
- **Key methods/signals**: `OnColorChanged func(color *qt.QColor)` (fired when OK is clicked and the color changed), `SetColor(color *qt.QColor, movePicker bool)`.
- **Effect**: A color-selection card with a scroll area pops over the mask; clicking OK confirms the new color and fires the callback; Cancel discards.

## folder_list_dialog.go

### `FolderCard`
- **Function**: A clickable card (292×72, with a delete icon) showing a folder's name and full path.
- **Constructor**: `NewFolderCard(folderPath string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnClicked func()`, `FolderPath() string`, `FolderName() string`.

### `AddFolderCard`
- **Function**: An "add folder" card with a `+` icon (292×72).
- **Constructor**: `NewAddFolderCard(parent *qt.QWidget)`.
- **Key methods/signals**: `OnClicked func()`.

### `FolderListDialog`
- **Function**: A dialog for editing a folder list, supporting adding/removing folders and reporting changes via callback.
- **Constructor**: `NewFolderListDialog(folderPaths []string, title, content string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnFolderChanged func(folders []string)`, `FolderPaths() []string`.
- **Effect**: The mask dialog shows a list of folder cards and a "+" add card; clicking `+` pops the system directory picker, clicking a card pops a delete confirmation.

---

# 4. settings Package

> In this package, every `icon interface{}` parameter accepts `*qt.QIcon`, a string path, or `common.FluentIconBase` (such as `common.Folder`), and may also be `nil`.

## setting_card.go

### `SettingIconWidget`
- **Function**: A small icon widget (16×16) with a "reduce opacity when disabled" effect.
- **Constructor**: `NewSettingIconWidget(icon interface{}, parent *qt.QWidget)`.
- **Key methods**: `SetIcon(icon)`, `IconSource()`.

### `SettingCard`
- **Function**: The basic settings card showing "icon + title + description text"; the base class of all settings cards.
- **Constructor**: `NewSettingCard(icon interface{}, title, content string, parent *qt.QWidget)`.
- **Key methods**: `SetTitle(title)`, `SetContent(content)`, `SetIconSize(width, height int)`, `TitleLabel()/ContentLabel()`, `HBoxLayout()/VBoxLayout()`, `SetValue(value interface{})` (empty implementation in the base class).

### `SwitchSettingCard`
- **Function**: A settings card with an on/off switch, for boolean config items.
- **Constructor**: `NewSwitchSettingCard(icon interface{}, title, content string, configItem *common.ConfigItem, parent *qt.QWidget)`.
- **Key methods/signals**: `OnCheckedChanged func(bool)`, `SetChecked(bool)` / `SetValue(interface{})`, `IsChecked()`, `SwitchButton()`.

### `RangeSettingCard`
- **Function**: A settings card with a slider, for integer-range numeric config.
- **Constructor**: `NewRangeSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, parent *qt.QWidget)` (`configItem` must not be `nil`).
- **Key methods/signals**: `OnValueChanged func(int)`, `SetValue(interface{})`, `Slider()`.

### `PushSettingCard` / `PrimaryPushSettingCard`
- **Function**: Settings cards with a button / primary button, for triggering a one-time action.
- **Constructor**: `NewPushSettingCard(text string, icon interface{}, title, content string, parent)` / `NewPrimaryPushSettingCard(...)`.
- **Key methods/signals**: `OnClicked func()`, `Button() *qt.QPushButton`.

### `HyperlinkCard`
- **Function**: A settings card with a hyperlink button, for jumping to external links.
- **Constructor**: `NewHyperlinkCard(url, text string, icon interface{}, title, content string, parent)`.
- **Key methods/signals**: `LinkButton() *widgets.HyperlinkButton`.

### `ColorPickerButton`
- **Function**: A color-picker button (96×32) showing the current color and popping `ColorDialog` on click.
- **Constructor**: `NewColorPickerButton(color *qt.QColor, title string, parent *qt.QWidget, enableAlpha bool)`.
- **Key methods/signals**: `OnColorChanged func(color *qt.QColor)`, `SetColor(color)`, `Color()`.

### `ColorSettingCard`
- **Function**: A settings card with a color-picker button, for color config items.
- **Constructor**: `NewColorSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, parent *qt.QWidget, enableAlpha bool)`.
- **Key methods/signals**: `OnColorChanged func(color *qt.QColor)`, `SetValue(interface{})`, `ColorPicker()`.

### `ComboBoxSettingCard`
- **Function**: A settings card with a drop-down box, for enumerated config items.
- **Constructor**: `NewComboBoxSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, texts []string, parent *qt.QWidget)`.
- **Key methods/signals**: `SetValue(interface{})`, `ComboBox()`.

## setting_card_group.go

### `SettingCardGroup`
- **Function**: A titled vertical container for stacking multiple settings cards into a group.
- **Constructor**: `NewSettingCardGroup(title string, parent *qt.QWidget)`.
- **Key methods**: `AddSettingCard(card *qt.QWidget)`, `AddSettingCards([]*qt.QWidget)`, `AdjustSize()`, `TitleLabel()`, `CardLayout()`.
- **Effect**: Multiple cards are arranged vertically under the group title; the group height auto-adjusts when cards expand/collapse.

## expand_setting_card.go

### `ExpandButton`
- **Function**: A rotatable arrow button (30×30) used as the toggle of an expandable card (the arrow rotates 180° when expanded).
- **Constructor**: `NewExpandButton(parent *qt.QWidget)`.
- **Key methods**: `SetExpand(bool)`, `SetClickedHook(func())`, `SetHover/SetPressed(bool)`, `Angle() float64`.

### `HeaderSettingCard`
- **Function**: A settings card with an expand button on the right, used as the header of `ExpandSettingCard`.
- **Constructor**: `NewHeaderSettingCard(icon interface{}, title, content string, parent *qt.QWidget)`.
- **Key methods**: `ExpandButton()`, `AddWidget(widget *qt.QWidget)`.

### `ExpandSettingCard`
- **Function**: An expandable/collapsible settings card based on `QScrollArea`; expanding reveals more content downward.
- **Constructor**: `NewExpandSettingCard(icon interface{}, title, content string, parent *qt.QWidget)`.
- **Key methods**: `SetExpand(bool)`, `ToggleExpand()`, `IsExpand()`, `AddWidget(widget)`, `Card()`, `View() *qt.QFrame` / `ViewLayout() *qt.QVBoxLayout`, `SetValue(interface{})`.

### `GroupWidget`
- **Function**: A single row (icon + title + description + trailing widget) in an expand group; it is one row of `ExpandGroupSettingCard`.
- **Constructor**: `NewGroupWidget(icon interface{}, title, content string, widget *qt.QWidget, stretch int, parent *qt.QWidget)`.
- **Key methods**: `SetTitle/SetContent`, `SetIcon(icon)`, `SetIconSize(width, height int)`, `Widget()`.

### `ExpandGroupSettingCard`
- **Function**: An expandable settings card whose expanded content is multiple `GroupWidget` rows (dividers automatically added between rows).
- **Constructor**: `NewExpandGroupSettingCard(icon interface{}, title, content string, parent *qt.QWidget)`.
- **Key methods**: `AddGroup(icon, title, content, widget, stretch int) *GroupWidget`, `AddGroupWidget(widget)`, `RemoveGroupWidget(widget)`, `Widgets()`.

### `SimpleExpandGroupSettingCard`
- **Function**: A simplified version of `ExpandGroupSettingCard` (expanded height uses the layout sizeHint).
- **Constructor**: `NewSimpleExpandGroupSettingCard(icon interface{}, title, content string, parent *qt.QWidget)`.

## options_setting_card.go

### `OptionsSettingCard`
- **Function**: An expandable settings card whose expanded content is a group of radio buttons, for single-choice enumerated config.
- **Constructor**: `NewOptionsSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, texts []string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnOptionChanged func()`, `SetValue(interface{})`, `Buttons() []*widgets.RadioButton`, `ButtonGroup() *qt.QButtonGroup`.

## custom_color_setting_card.go

### `CustomColorSettingCard`
- **Function**: An expandable settings card that switches between "default color" and "custom color"; the custom color is picked via `ColorDialog`.
- **Constructor**: `NewCustomColorSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, parent *qt.QWidget, enableAlpha bool)`.
- **Key methods/signals**: `OnColorChanged func(color *qt.QColor)`, `ConfigItem()`.

## folder_list_setting_card.go

### `FolderItem`
- **Function**: A single row in a folder list, showing the folder path with a delete button.
- **Constructor**: `NewFolderItem(folder string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnRemoved func(item *FolderItem)`, `Folder() string`.

### `FolderListSettingCard`
- **Function**: An expandable settings card that maintains a folder list; the header provides an "add folder" button and each row can be deleted.
- **Constructor**: `NewFolderListSettingCard(configItem *common.ConfigItem, title, content, directory string, parent *qt.QWidget)`.
- **Key methods/signals**: `OnFolderChanged func(folders []string)`, `Folders() []string`.

---

# 5. date_time Package

> The base classes `PickerBase` / `DatePickerBase` / `TimePickerBase` and internal helper types are not listed.

## date_picker.go

### `DatePicker`
- **Function**: A Fluent-style date picker button; clicking pops a three-column year/month/day panel that scrolls up and down.
- **Constructor**: `NewDatePicker(parent *qt.QWidget, format int, isMonthTight bool)`.
  - `format`: `DatePickerMMDDYYYY` (month-day-year) or `DatePickerYYYYMMDD` (year-month-day); `isMonthTight`: whether the month column is compact.
- **Key methods/signals**: `Date() *qt.QDate`, `SetDate(date *qt.QDate)`, `SetDateFormat(format int)`, `SetMonthTight(bool)`, `SetYearFormatter/SetMonthFormatter/SetDayFormatter(f PickerColumnFormatter)`, `OnDateChanged func(*qt.QDate)`, `Reset()`; inherits `PickerBase`'s `SetColumnWidth/SetColumnVisible/SetResetEnabled/Value()` etc.
- **Effect**: The button text is the current date; clicking pops a three-column panel that scrolls up/down, with the selected row highlighted at the top and OK/reset/cancel buttons at the bottom.

### `ZhDatePicker`
- **Function**: A Chinese date picker that displays and selects in the "YYYY年MM月DD日" format.
- **Constructor**: `NewZhDatePicker(parent *qt.QWidget)`.
- **Key methods/signals**: Inherits `DatePicker`; the default Chinese formatter carries the 年/月/日 suffixes.

## time_picker.go

### `TimePicker`
- **Function**: A 24-hour Fluent time picker; clicking pops an hour/minute (optional second) column panel.
- **Constructor**: `NewTimePicker(parent *qt.QWidget, showSeconds bool)`.
- **Key methods/signals**: `Time() *qt.QTime`, `SetTime(time *qt.QTime)`, `SetSecondVisible(bool)` / `IsSecondVisible()`, `OnTimeChanged func(*qt.QTime)`, `Reset()`.
- **Effect**: The button shows the current time; clicking pops a scroll-column panel; hours 0–23, minutes/seconds 0–59 (zero-padded to two digits).

### `AMTimePicker`
- **Function**: A 12-hour (AM/PM) Fluent time picker, with an additional AM/PM column.
- **Constructor**: `NewAMTimePicker(parent *qt.QWidget, showSeconds bool)`.
- **Key methods/signals**: Same as `TimePicker`; `Time()` is internally a 24-hour value.
- **Effect**: The button shows 12-hour time; the popup panel contains hour (1–12), minute (optional second) and AM/PM columns.

## calendar_picker.go

### `CalendarPicker`
- **Function**: A Fluent calendar picker button; clicking pops a month calendar view and the button shows the selected date.
- **Constructor**: `NewCalendarPicker(parent *qt.QWidget)`.
- **Key methods/signals**: `Date() *qt.QDate`, `SetDate(date)`, `Reset()`, `DateFormat()/SetDateFormat(format string)`, `SetResetEnabled(bool)` / `IsResetEnabled()`, `OnDateChanged func(*qt.QDate)`.
- **Effect**: A drop-down button with a calendar icon; shows the "Pick a date" placeholder when nothing is selected, and the formatted date after selection; clicking pops the calendar view, which auto-collapses after a day is selected.

### `FastCalendarPicker`
- **Function**: The "Pro" version of `CalendarPicker`, showing the calendar view in a Flyout.
- **Constructor**: `NewFastCalendarPicker(parent *qt.QWidget)`.
- **Key methods/signals**: Same as `CalendarPicker`; also `SetFlyoutAnimationType(aniType int)`.

## calendar_view.go / fast_calendar_view.go

### `CalendarView`
- **Function**: The popup calendar view used by `CalendarPicker`, showing day/month/year three-level grids, supporting title switching, paging up/down and an optional reset button.
- **Constructor**: `NewCalendarView(parent *qt.QWidget)`.
- **Key methods/signals**: `SetDate(date *qt.QDate)`, `SetResetEnabled(bool)` / `IsResetEnabled()`, `Exec(pos *qt.QPoint, ani bool)`, `OnDateChanged func(*qt.QDate)`, `OnResetted func()`.
- **Effect**: A borderless, semi-transparent, shadowed popup card (314×355); the top title button cycles through "month + year → year → decade range"; wheel or up/down buttons page with a 300ms animation.

### `FastCalendarView`
- **Function**: The "Pro" version of `CalendarView`, for use by `FastCalendarPicker` in a Flyout.
- **Constructor**: `NewFastCalendarView(parent *qt.QWidget)`.
- **Key methods/signals**: Inherits `CalendarView`.

---

# 6. layout Package

## flow_layout.go

### `FlowLayout`
- **Function**: A flow layout; widgets are arranged left-to-right, wrapping automatically when the available width is exceeded, with support for layout-reflow animation.
- **Constructor**: `NewFlowLayout(parent *qt.QWidget, needAni, isTight bool)`.
  - `needAni`: whether to animate widget position changes during reflow; `isTight`: compact mode (skip hidden widgets).
- **Key methods**: `AddWidget(w)`, `InsertWidget(index, w)`, `InsertItem(index, item)`, `AddItem(item)`, `RemoveWidget(w)`, `RemoveAllWidgets()`, `TakeAllWidgets()`, `Count()`, `ItemAt(index)`, `TakeAt(index)`, `SetAnimation(duration int)`, `SetVerticalSpacing/SetHorizontalSpacing(int)`.
- **Effect**: Widgets are laid out left-to-right, wrapping to the next line when one line is full; with `needAni` enabled, widgets move smoothly when added/removed or resized.

### `AdaptiveFlowLayout`
- **Function**: An adaptive flow layout; based on `FlowLayout`, it stretches each row's cards to uniform width and computes the per-row count automatically from the minimum/maximum card width.
- **Constructor**: `NewAdaptiveFlowLayout(parent *qt.QWidget, needAni, isTight bool)`.
- **Key methods**: `SetWidgetMinimumWidth(int)` / `WidgetMinimumWidth()`, `SetWidgetMaximumWidth(int)` / `WidgetMaximumWidth()`; inherits `FlowLayout`.
- **Effect**: Cards in each row have the same width and fill the available width; as the container widens each row holds more cards, and shrinks/wraps when narrowed.

## v_box_layout.go

### `VBoxLayout`
- **Function**: A lightweight wrapper over `QVBoxLayout` that stacks widgets vertically and additionally records all added widgets for batch removal or deletion.
- **Constructor**: `NewVBoxLayout(parent *qt.QWidget)`.
- **Key methods**: `AddWidget(widget, stretch int, alignment qt.AlignmentFlag)`, `AddWidgets([]*qt.QWidget, stretch, alignment)`, `RemoveWidget(widget)`, `DeleteWidget(widget)`, `RemoveAllWidgets()`.
- **Effect**: Widgets are arranged vertically from top to bottom in the order added; `stretch` controls the share of extra space a widget receives, and `alignment` controls alignment.

## expand_layout.go

### `ExpandLayout`
- **Function**: A vertical-only layout that stacks child widgets top-to-bottom at their natural height and automatically re-lays out when the parent geometry or a child's height changes.
- **Constructor**: `NewExpandLayout(parent *qt.QWidget)`.
- **Key methods**: `AddWidget(widget)`, `AddItem(item)`, `Count()`, `ItemAt(index)`, `WidgetAt(index) *qt.QWidget`, `TakeAt(index)`.
- **Effect**: Child widgets stack vertically at natural height; when a child (such as an expandable settings card) changes height, the layout adjusts the parent height in sync and pushes subsequent content downward.
