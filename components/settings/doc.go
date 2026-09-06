// Package settings ports qfluentwidgets/components/settings to Go + miqt.
//
// It provides the fluent setting cards (SettingCard, SwitchSettingCard,
// RangeSettingCard, PushSettingCard, ColorSettingCard, ComboBoxSettingCard,
// HyperlinkCard), the expandable cards (ExpandSettingCard and its group
// variants), the custom color / options / folder-list cards and the
// SettingCardGroup container.
//
// Dependencies (all strictly one-way):
//   - gofluent/common      (config, style sheets, icons, fonts, text wrap)
//   - gofluent/components/widgets  (buttons, combo box, slider, switch, icon widget)
//   - gofluent/components/dialog_box (ColorDialog, Dialog)
//
// Signals from the Python source (pyqtSignal) are expressed as Go callback
// fields on each card (OnCheckedChanged, OnValueChanged, OnClicked, ...).
// Custom meta-object property animations (the expand-button rotation and the
// expand-card scroll animation) are driven frame-by-frame with
// common.ProgressAnimation (custom pyqtProperty values cannot be registered
// from Go, see MIGRATION_GUIDE §12).
package settings
