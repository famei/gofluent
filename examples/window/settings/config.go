// Package settings is the multi-file settings example (config + setting
// interface + demo). It mirrors PyQt-Fluent-Widgets' examples/window/settings.
//
// The upstream Config subclasses QConfig and adds a large set of typed config
// items, then qconfig.load() persists them to config/config.json. The Go
// common.QConfig only persists its three built-in items (theme mode, theme
// color, font families), so this port keeps the extra items in memory using
// common.ConfigItem (value validation + change callbacks still work); the
// json persistence of the extra items is intentionally omitted.
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
)

// SongQuality enumerates the online song quality levels.
type SongQuality int

const (
	SongQualityStandard SongQuality = iota
	SongQualityHigh
	SongQualitySuper
	SongQualityLossless
)

// String returns the user-facing quality name.
func (q SongQuality) String() string {
	return [...]string{"Standard quality", "High quality", "Super quality", "Lossless quality"}[q]
}

// MvQuality enumerates the online MV quality levels.
type MvQuality int

const (
	MvQualityFullHD MvQuality = iota
	MvQualityHD
	MvQualitySD
	MvQualityLD
)

// String returns the user-facing MV quality name.
func (q MvQuality) String() string {
	return [...]string{"Full HD", "HD", "SD", "LD"}[q]
}

// Language enumerates the supported UI languages.
type Language int

const (
	LanguageChineseSimplified Language = iota
	LanguageChineseTraditional
	LanguageEnglish
	LanguageAuto
)

// LocaleName returns the Qt locale name for the language ("" for auto).
func (l Language) LocaleName() string {
	switch l {
	case LanguageChineseSimplified:
		return "zh_CN"
	case LanguageChineseTraditional:
		return "zh_HK"
	case LanguageEnglish:
		return "en_US"
	default:
		return ""
	}
}

// App metadata used by the about cards (kept in Go for the demo).
const (
	AppYear     = 2023
	AppAuthor   = "zhiyiYo"
	AppVersion  = "1.0.0"
	AppHelpURL  = "https://pyqt-fluent-widgets.readthedocs.io"
	AppFeedback = "https://github.com/zhiyiYo/PyQt-Fluent-Widgets/issues"
	AppRelease  = "https://github.com/zhiyiYo/PyQt-Fluent-Widgets/releases/latest"
)

// Config holds the application config items. The theme mode and theme color
// reuse the process-wide common.QConfigInstance built-ins so the setting cards
// drive the global theme system.
type Config struct {
	MusicFolders            *common.ConfigItem
	DownloadFolder          *common.ConfigItem
	OnlineSongQuality       *common.ConfigItem
	OnlinePageSize          *common.ConfigItem
	OnlineMvQuality         *common.ConfigItem
	EnableAcrylicBackground *common.ConfigItem
	MinimizeToTray          *common.ConfigItem
	PlayBarColor            *common.ConfigItem
	RecentPlaysNumber       *common.ConfigItem
	DpiScale                *common.ConfigItem
	Language                *common.ConfigItem
	DeskLyricHighlightColor *common.ConfigItem
	DeskLyricFontSize       *common.ConfigItem
	DeskLyricStrokeSize     *common.ConfigItem
	DeskLyricStrokeColor    *common.ConfigItem
	DeskLyricFontFamily     *common.ConfigItem
	DeskLyricAlignment      *common.ConfigItem
	CheckUpdateAtStartUp    *common.ConfigItem
}

// newConfig builds the config items with the same defaults as the Python port.
func newConfig() *Config {
	c := &Config{
		MusicFolders: common.NewConfigItem(
			"Folders", "LocalMusic", []string{}, common.FolderListValidator{}, common.DefaultSerializer{}, false),
		DownloadFolder: common.NewConfigItem(
			"Folders", "Download", "download", common.FolderValidator{}, common.DefaultSerializer{}, false),

		OnlineSongQuality: common.NewConfigItem(
			"Online", "SongQuality", SongQualityStandard,
			common.NewOptionsValidator(SongQualityStandard, SongQualityHigh, SongQualitySuper, SongQualityLossless),
			common.DefaultSerializer{}, false),
		OnlinePageSize: common.NewConfigItem(
			"Online", "PageSize", 30, &common.RangeValidator{Min: 0, Max: 50}, common.DefaultSerializer{}, false),
		OnlineMvQuality: common.NewConfigItem(
			"Online", "MvQuality", MvQualityFullHD,
			common.NewOptionsValidator(MvQualityFullHD, MvQualityHD, MvQualitySD, MvQualityLD),
			common.DefaultSerializer{}, false),

		EnableAcrylicBackground: common.NewConfigItem(
			"MainWindow", "EnableAcrylicBackground", false, common.NewBoolValidator(), common.DefaultSerializer{}, false),
		MinimizeToTray: common.NewConfigItem(
			"MainWindow", "MinimizeToTray", true, common.NewBoolValidator(), common.DefaultSerializer{}, false),
		PlayBarColor: common.NewConfigItem(
			"MainWindow", "PlayBarColor", "#225C7F", common.NewColorValidator("#225C7F"), common.ColorSerializer{}, false),
		RecentPlaysNumber: common.NewConfigItem(
			"MainWindow", "RecentPlayNumbers", 300, &common.RangeValidator{Min: 10, Max: 300}, common.DefaultSerializer{}, false),
		DpiScale: common.NewConfigItem(
			"MainWindow", "DpiScale", "Auto",
			common.NewOptionsValidator(float64(1), 1.25, 1.5, 1.75, float64(2), "Auto"),
			common.DefaultSerializer{}, true),
		Language: common.NewConfigItem(
			"MainWindow", "Language", LanguageAuto,
			common.NewOptionsValidator(LanguageChineseSimplified, LanguageChineseTraditional, LanguageEnglish, LanguageAuto),
			common.DefaultSerializer{}, true),

		DeskLyricHighlightColor: common.NewConfigItem(
			"DesktopLyric", "HighlightColor", "#0099BC", common.NewColorValidator("#0099BC"), common.ColorSerializer{}, false),
		DeskLyricFontSize: common.NewConfigItem(
			"DesktopLyric", "FontSize", 50, &common.RangeValidator{Min: 15, Max: 50}, common.DefaultSerializer{}, false),
		DeskLyricStrokeSize: common.NewConfigItem(
			"DesktopLyric", "StrokeSize", 5, &common.RangeValidator{Min: 0, Max: 20}, common.DefaultSerializer{}, false),
		DeskLyricStrokeColor: common.NewConfigItem(
			"DesktopLyric", "StrokeColor", qt.NewQColor3(0, 0, 0), common.NewColorValidator("#000000"), common.ColorSerializer{}, false),
		DeskLyricFontFamily: common.NewConfigItem(
			"DesktopLyric", "FontFamily", "Microsoft YaHei", common.DefaultValidator{}, common.DefaultSerializer{}, false),
		DeskLyricAlignment: common.NewConfigItem(
			"DesktopLyric", "Alignment", "Center",
			common.NewOptionsValidator("Center", "Left", "Right"),
			common.DefaultSerializer{}, false),

		CheckUpdateAtStartUp: common.NewConfigItem(
			"Update", "CheckUpdateAtStartUp", true, common.NewBoolValidator(), common.DefaultSerializer{}, false),
	}
	return c
}

// ThemeMode returns the built-in theme mode config item.
func (c *Config) ThemeMode() *common.ConfigItem { return common.QConfigInstance.ThemeMode() }

// ThemeColor returns the built-in theme color config item.
func (c *Config) ThemeColor() *common.ConfigItem { return common.QConfigInstance.ThemeColorItem() }

// DownloadFolderPath returns the current download folder as a string.
func (c *Config) DownloadFolderPath() string {
	if v, ok := c.DownloadFolder.Value().(string); ok {
		return v
	}
	return "download"
}

// cfg is the process-wide settings example config.
var cfg = newConfig()
