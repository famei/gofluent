package common

import (
	"runtime"
	"strings"

	gcommon "github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// Language enumerates the supported UI locales. The order matches the option
// list of the language ComboBoxSettingCard in the settings interface
// (简体中文 / 繁體中文 / English / 使用系统设置).
type Language int

const (
	LanguageChineseSimplified Language = iota
	LanguageChineseTraditional
	LanguageEnglish
	LanguageAuto
)

// Name returns the locale name used by Qt ("zh_CN", "zh_HK", "en_US" or "").
func (l Language) Name() string {
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

// Locale builds the QLocale for this language. Explicit languages return
// caller-owned QLocale objects; AUTO returns the GoGC-armed system locale.
func (l Language) Locale() *qt.QLocale {
	switch l {
	case LanguageChineseSimplified:
		return qt.NewQLocale6(qt.QLocale__Chinese, qt.QLocale__China)
	case LanguageChineseTraditional:
		return qt.NewQLocale6(qt.QLocale__Chinese, qt.QLocale__HongKong)
	case LanguageEnglish:
		return qt.NewQLocale3(qt.QLocale__English)
	default:
		// Use QLocale.system() explicitly so AUTO detects a Chinese system and
		// auto-loads the Chinese translation (bug #33).
		return qt.QLocale_System() // GoGC-armed — do NOT Delete
	}
}

func languageFromName(name string) Language {
	switch strings.ToLower(name) {
	case "zh_cn":
		return LanguageChineseSimplified
	case "zh_hk", "zh_tw":
		return LanguageChineseTraditional
	case "en", "en_us":
		return LanguageEnglish
	default:
		return LanguageAuto
	}
}

// LanguageSerializer serializes a Language as its locale name ("Auto" for the
// system default), mirroring app/common/config.py LanguageSerializer.
type LanguageSerializer struct{ gcommon.DefaultSerializer }

// Serialize returns the locale name of the language, or "Auto".
func (LanguageSerializer) Serialize(value interface{}) interface{} {
	l, ok := value.(Language)
	if !ok || l == LanguageAuto {
		return "Auto"
	}
	return l.Name()
}

// Deserialize converts a locale name back to a Language.
func (LanguageSerializer) Deserialize(value interface{}) interface{} {
	s, ok := value.(string)
	if !ok {
		return LanguageAuto
	}
	if strings.EqualFold(s, "Auto") {
		return LanguageAuto
	}
	return languageFromName(s)
}

// Application metadata (port of app/common/config.py module constants).
const (
	YEAR           = 2026
	AUTHOR         = "famei"
	VERSION        = "0.1.0"
	HELP_URL       = "https://github.com/famei/gofluent/tree/main/docs"
	REPO_URL       = "https://github.com/famei/gofluent"
	EXAMPLE_URL    = "https://github.com/famei/gofluent/tree/main/examples"
	FEEDBACK_URL   = "https://github.com/famei/gofluent/issues"
	ZH_SUPPORT_URL = "https://github.com/famei/gofluent"
	EN_SUPPORT_URL = "https://github.com/famei/gofluent"
)

// IsWin11 reports whether the host is Windows 11 (build >= 22000). The build
// number cannot be read portably, so the check degrades to "is Windows".
func IsWin11() bool {
	return runtime.GOOS == "windows"
}

// Config holds the gallery-specific configuration items (port of the Config
// class in app/common/config.py). The theme/theme-color items are shared with
// the gofluent library singleton common.QConfigInstance.
type Config struct {
	// Folders
	MusicFolders   *gcommon.ConfigItem
	DownloadFolder *gcommon.ConfigItem

	// Main window
	MicaEnabled *gcommon.ConfigItem
	DpiScale    *gcommon.ConfigItem
	Language    *gcommon.ConfigItem

	// Material
	BlurRadius *gcommon.ConfigItem

	// Software update
	CheckUpdateAtStartUp *gcommon.ConfigItem
}

// NewConfig builds the gallery config with the same defaults as the Python app.
func NewConfig() *Config {
	return &Config{
		MusicFolders: gcommon.NewConfigItem(
			"Folders", "LocalMusic", []string{}, gcommon.FolderListValidator{}, gcommon.DefaultSerializer{}, false),
		DownloadFolder: gcommon.NewConfigItem(
			"Folders", "Download", "app/download", gcommon.FolderValidator{}, gcommon.DefaultSerializer{}, false),
		// Mica is disabled by default: the transparent Mica backdrop does not
		// render dark on this host (and follows the OS theme, not the gallery
		// toggle), which left the window white in dark mode. A solid background
		// is painted instead; users can still enable Mica in Settings.
		MicaEnabled: gcommon.NewConfigItem(
			"MainWindow", "MicaEnabled", false, gcommon.NewBoolValidator(), gcommon.DefaultSerializer{}, false),
		DpiScale: gcommon.NewConfigItem(
			"MainWindow", "DpiScale", "Auto",
			gcommon.NewOptionsValidator(1.0, 1.25, 1.5, 1.75, 2.0, "Auto"),
			gcommon.DefaultSerializer{}, true),
		Language: gcommon.NewConfigItem(
			"MainWindow", "Language", LanguageAuto,
			gcommon.NewOptionsValidator(LanguageChineseSimplified, LanguageChineseTraditional, LanguageEnglish, LanguageAuto),
			LanguageSerializer{}, false),
		BlurRadius: gcommon.NewConfigItem(
			"Material", "AcrylicBlurRadius", 15, &gcommon.RangeValidator{Min: 0, Max: 40}, gcommon.DefaultSerializer{}, false),
		CheckUpdateAtStartUp: gcommon.NewConfigItem(
			"Update", "CheckUpdateAtStartUp", true, gcommon.NewBoolValidator(), gcommon.DefaultSerializer{}, false),
	}
}

// ThemeMode returns the shared theme mode config item.
func (c *Config) ThemeMode() *gcommon.ConfigItem { return gcommon.QConfigInstance.ThemeMode() }

// ThemeColor returns the shared theme color config item.
func (c *Config) ThemeColor() *gcommon.ConfigItem { return gcommon.QConfigInstance.ThemeColorItem() }

// Get returns the value of a config item.
func (c *Config) Get(item *gcommon.ConfigItem) interface{} { return item.Value() }

// Set stores a value on a config item and persists it. Theme-related and
// restart-emitting items go through the shared gofluent config singleton.
func (c *Config) Set(item *gcommon.ConfigItem, value interface{}, save bool) {
	gcommon.QConfigInstance.Set(item, value, save)
}

// ConfigInstance is the process-wide gallery config (the Python module-level
// `cfg`). Its theme mode defaults to AUTO, matching config.py.
var ConfigInstance = NewConfig()

func init() {
	// Persist the language selection so the gallery restores the chosen locale
	// on the next launch, mirroring config.py's qconfig.load('app/config/config.json', cfg).
	// The language is also switched live via common.SwitchLanguage (restart=false
	// on the Language item keeps appRestartSig reserved for DpiScale, which does
	// require a restart).
	gcommon.QConfigInstance.RegisterItems(ConfigInstance.Language)
	gcommon.QConfigInstance.Load("")
}
