package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/settings"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
)

// SettingInterface is the settings page. The Python port subclasses ScrollArea
// and lays the cards out with an ExpandLayout; the Go port uses widgets.ScrollArea
// and a plain QVBoxLayout (matching the ExpandLayout downgrade used by the
// settings package).
type SettingInterface struct {
	*widgets.ScrollArea

	scrollWidget *qt.QWidget
	settingLabel *qt.QLabel

	musicInThisPCGroup          *settings.SettingCardGroup
	musicFolderCard             *settings.FolderListSettingCard
	downloadFolderCard          *settings.PushSettingCard
	personalGroup               *settings.SettingCardGroup
	enableAcrylicCard           *settings.SwitchSettingCard
	themeCard                   *settings.OptionsSettingCard
	themeColorCard              *settings.CustomColorSettingCard
	zoomCard                    *settings.OptionsSettingCard
	languageCard                *settings.ComboBoxSettingCard
	onlineMusicGroup            *settings.SettingCardGroup
	onlinePageSizeCard          *settings.RangeSettingCard
	onlineMusicQualityCard      *settings.OptionsSettingCard
	onlineMvQualityCard         *settings.OptionsSettingCard
	deskLyricGroup              *settings.SettingCardGroup
	deskLyricFontCard           *settings.PushSettingCard
	deskLyricHighlightColorCard *settings.ColorSettingCard
	deskLyricStrokeColorCard    *settings.ColorSettingCard
	deskLyricStrokeSizeCard     *settings.RangeSettingCard
	deskLyricAlignmentCard      *settings.OptionsSettingCard
	mainPanelGroup              *settings.SettingCardGroup
	minimizeToTrayCard          *settings.SwitchSettingCard
	updateSoftwareGroup         *settings.SettingCardGroup
	updateOnStartUpCard         *settings.SwitchSettingCard
	aboutGroup                  *settings.SettingCardGroup
	helpCard                    *settings.HyperlinkCard
	feedbackCard                *settings.PrimaryPushSettingCard
	aboutCard                   *settings.PrimaryPushSettingCard

	// External signals (the Python pyqtSignal declarations). The demo does not
	// connect these, but the fields are exported for downstream users.
	CheckUpdateSig        func()
	MusicFoldersChanged   func(folders []string)
	AcrylicEnableChanged  func(enabled bool)
	DownloadFolderChanged func(folder string)
	MinimizeToTrayChanged func(enabled bool)
}

// NewSettingInterface builds the settings interface.
func NewSettingInterface(parent *qt.QWidget) *SettingInterface {
	s := &SettingInterface{ScrollArea: widgets.NewScrollArea(parent)}
	s.scrollWidget = qt.NewQWidget2()
	s.settingLabel = qt.NewQLabel5("Settings", nil)

	s.buildGroups()
	s.initWidget()
	s.initLayout()
	s.connectSignalToSlot()
	return s
}

func (s *SettingInterface) buildGroups() {
	// music on this PC
	s.musicInThisPCGroup = settings.NewSettingCardGroup("Music on this PC", s.scrollWidget)
	s.musicFolderCard = settings.NewFolderListSettingCard(
		cfg.MusicFolders, "Local music library", "", "download", s.musicInThisPCGroup.QWidget)
	s.downloadFolderCard = settings.NewPushSettingCard(
		"Choose folder", common.Download, "Download directory", cfg.DownloadFolderPath(), s.musicInThisPCGroup.QWidget)

	// personalization
	s.personalGroup = settings.NewSettingCardGroup("Personalization", s.scrollWidget)
	s.enableAcrylicCard = settings.NewSwitchSettingCard(
		common.SquareSparkle, "Use Acrylic effect",
		"Acrylic effect has better visual experience, but it may cause the window to become stuck",
		cfg.EnableAcrylicBackground, s.personalGroup.QWidget)
	s.themeCard = settings.NewOptionsSettingCard(
		cfg.ThemeMode(), common.Eyedropper, "Application theme", "Change the appearance of your application",
		[]string{"Light", "Dark", "Use system setting"}, s.personalGroup.QWidget)
	s.themeColorCard = settings.NewCustomColorSettingCard(
		cfg.ThemeColor(), common.Color, "Theme color", "Change the theme color of you application",
		s.personalGroup.QWidget, false)
	s.zoomCard = settings.NewOptionsSettingCard(
		cfg.DpiScale, common.Zoom, "Interface zoom", "Change the size of widgets and fonts",
		[]string{"100%", "125%", "150%", "175%", "200%", "Use system setting"}, s.personalGroup.QWidget)
	s.languageCard = settings.NewComboBoxSettingCard(
		cfg.Language, common.LocaleLanguage, "Language", "Set your preferred language for UI",
		[]string{"简体中文", "繁體中文", "English", "Use system setting"}, s.personalGroup.QWidget)

	// online music
	s.onlineMusicGroup = settings.NewSettingCardGroup("Online Music", s.scrollWidget)
	s.onlinePageSizeCard = settings.NewRangeSettingCard(
		cfg.OnlinePageSize, common.Search, "Number of online music displayed on each page", "", s.onlineMusicGroup.QWidget)
	s.onlineMusicQualityCard = settings.NewOptionsSettingCard(
		cfg.OnlineSongQuality, common.Audio, "Online music quality", "",
		[]string{"Standard quality", "High quality", "Super quality", "Lossless quality"}, s.onlineMusicGroup.QWidget)
	s.onlineMvQualityCard = settings.NewOptionsSettingCard(
		cfg.OnlineMvQuality, common.Video, "Online MV quality", "",
		[]string{"Full HD", "HD", "SD", "LD"}, s.onlineMusicGroup.QWidget)

	// desktop lyric
	s.deskLyricGroup = settings.NewSettingCardGroup("Desktop Lyric", s.scrollWidget)
	s.deskLyricFontCard = settings.NewPushSettingCard(
		"Choose font", common.Font, "Font", "", s.deskLyricGroup.QWidget)
	s.deskLyricHighlightColorCard = settings.NewColorSettingCard(
		cfg.DeskLyricHighlightColor, common.Color, "Foreground color", "", s.deskLyricGroup.QWidget, false)
	s.deskLyricStrokeColorCard = settings.NewColorSettingCard(
		cfg.DeskLyricStrokeColor, common.PencilInk, "Stroke color", "", s.deskLyricGroup.QWidget, false)
	s.deskLyricStrokeSizeCard = settings.NewRangeSettingCard(
		cfg.DeskLyricStrokeSize, common.Highlight, "Stroke size", "", s.deskLyricGroup.QWidget)
	s.deskLyricAlignmentCard = settings.NewOptionsSettingCard(
		cfg.DeskLyricAlignment, common.Alignment, "Alignment", "",
		[]string{"Center aligned", "Left aligned", "Right aligned"}, s.deskLyricGroup.QWidget)

	// main panel
	s.mainPanelGroup = settings.NewSettingCardGroup("Main Panel", s.scrollWidget)
	s.minimizeToTrayCard = settings.NewSwitchSettingCard(
		common.Minimize, "Minimize to tray after closing",
		"PyQt-Fluent-Widgets will continue to run in the background",
		cfg.MinimizeToTray, s.mainPanelGroup.QWidget)

	// software update
	s.updateSoftwareGroup = settings.NewSettingCardGroup("Software update", s.scrollWidget)
	s.updateOnStartUpCard = settings.NewSwitchSettingCard(
		common.Sync, "Check for updates when the application starts",
		"The new version will be more stable and have more features",
		cfg.CheckUpdateAtStartUp, s.updateSoftwareGroup.QWidget)

	// about
	s.aboutGroup = settings.NewSettingCardGroup("About", s.scrollWidget)
	s.helpCard = settings.NewHyperlinkCard(
		AppHelpURL, "Open help page", common.Help, "Help",
		"Discover new features and learn useful tips about gofluent", s.aboutGroup.QWidget)
	s.feedbackCard = settings.NewPrimaryPushSettingCard(
		"Provide feedback", common.Feedback, "Provide feedback",
		"Help us improve gofluent by providing feedback", s.aboutGroup.QWidget)
	s.aboutCard = settings.NewPrimaryPushSettingCard(
		"Check update", common.Info, "About",
		"© Copyright 2023, zhiyiYo. Version 1.0.0", s.aboutGroup.QWidget)
}

func (s *SettingInterface) initWidget() {
	s.Resize(1000, 800)
	s.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	s.SetViewportMargins(0, 120, 0, 20)
	s.SetWidget(s.scrollWidget)
	s.SetWidgetResizable(true)
	s.setQss()
}

func (s *SettingInterface) initLayout() {
	// add cards to groups
	s.musicInThisPCGroup.AddSettingCard(s.musicFolderCard.QWidget)
	s.musicInThisPCGroup.AddSettingCard(s.downloadFolderCard.QWidget)

	s.personalGroup.AddSettingCard(s.enableAcrylicCard.QWidget)
	s.personalGroup.AddSettingCard(s.themeCard.QWidget)
	s.personalGroup.AddSettingCard(s.themeColorCard.QWidget)
	s.personalGroup.AddSettingCard(s.zoomCard.QWidget)
	s.personalGroup.AddSettingCard(s.languageCard.QWidget)

	s.onlineMusicGroup.AddSettingCard(s.onlinePageSizeCard.QWidget)
	s.onlineMusicGroup.AddSettingCard(s.onlineMusicQualityCard.QWidget)
	s.onlineMusicGroup.AddSettingCard(s.onlineMvQualityCard.QWidget)

	s.deskLyricGroup.AddSettingCard(s.deskLyricFontCard.QWidget)
	s.deskLyricGroup.AddSettingCard(s.deskLyricHighlightColorCard.QWidget)
	s.deskLyricGroup.AddSettingCard(s.deskLyricStrokeColorCard.QWidget)
	s.deskLyricGroup.AddSettingCard(s.deskLyricStrokeSizeCard.QWidget)
	s.deskLyricGroup.AddSettingCard(s.deskLyricAlignmentCard.QWidget)

	s.mainPanelGroup.AddSettingCard(s.minimizeToTrayCard.QWidget)
	s.updateSoftwareGroup.AddSettingCard(s.updateOnStartUpCard.QWidget)

	s.aboutGroup.AddSettingCard(s.helpCard.QWidget)
	s.aboutGroup.AddSettingCard(s.feedbackCard.QWidget)
	s.aboutGroup.AddSettingCard(s.aboutCard.QWidget)

	// stack the groups in a plain vertical layout (ExpandLayout downgrade)
	layout := qt.NewQVBoxLayout(s.scrollWidget)
	layout.SetContentsMargins(60, 63, 60, 0)
	layout.SetSpacing(28)
	layout.AddWidget(s.settingLabel.QWidget)
	layout.AddWidget(s.musicInThisPCGroup.QWidget)
	layout.AddWidget(s.personalGroup.QWidget)
	layout.AddWidget(s.onlineMusicGroup.QWidget)
	layout.AddWidget(s.deskLyricGroup.QWidget)
	layout.AddWidget(s.mainPanelGroup.QWidget)
	layout.AddWidget(s.updateSoftwareGroup.QWidget)
	layout.AddWidget(s.aboutGroup.QWidget)
	layout.AddStretchWithStretch(1)
}

func (s *SettingInterface) setQss() {
	s.scrollWidget.SetObjectName("scrollWidget")
	s.settingLabel.SetObjectName("settingLabel")

	dir := "light"
	if common.IsDarkTheme() {
		dir = "dark"
	}
	s.SetStyleSheet(asset.String(settingsQSS, "resource/qss/"+dir+"/setting_interface.qss"))
}

func (s *SettingInterface) showRestartTooltip() {
	widgets.InfoBarWarning("", "Configuration takes effect after restart", s.Window())
}

func (s *SettingInterface) onDownloadFolderCardClicked() {
	folder := qt.QFileDialog_GetExistingDirectory3(s.QWidget, "Choose folder", "./")
	if folder == "" || folder == cfg.DownloadFolderPath() {
		return
	}
	common.QConfigInstance.Set(cfg.DownloadFolder, folder, true)
	s.downloadFolderCard.SetContent(folder)
	if s.DownloadFolderChanged != nil {
		s.DownloadFolderChanged(folder)
	}
}

func (s *SettingInterface) connectSignalToSlot() {
	common.QConfigInstance.OnAppRestart(func() { s.showRestartTooltip() })
	common.QConfigInstance.OnThemeChanged(func(t common.Theme) {
		// Mirror the Python __onThemeChanged: call setTheme(theme) so every
		// registered fluent widget re-applies its QSS for the new theme (and
		// themeChangedFinished fires), then re-apply the demo's own QSS. The
		// theme card drives the switch through QConfigInstance.Set directly,
		// which only emits themeChanged — without setTheme the fluent widgets
		// (cards, groups, ...) keep the old theme's QSS until a restart.
		common.SetTheme(t, false, false)
		s.setQss()
	})

	s.musicFolderCard.OnFolderChanged = func(folders []string) {
		if s.MusicFoldersChanged != nil {
			s.MusicFoldersChanged(folders)
		}
	}
	s.downloadFolderCard.OnClicked = s.onDownloadFolderCardClicked

	s.enableAcrylicCard.OnCheckedChanged = func(isChecked bool) {
		if s.AcrylicEnableChanged != nil {
			s.AcrylicEnableChanged(isChecked)
		}
	}
	s.themeColorCard.OnColorChanged = func(color *qt.QColor) {
		common.SetThemeColor(color, false, false)
	}

	// The Python port opens a QFontDialog here; QFontDialog is not wired in the
	// Go port, so the font card remains a no-op.
	s.deskLyricFontCard.OnClicked = func() {}

	s.minimizeToTrayCard.OnCheckedChanged = func(isChecked bool) {
		if s.MinimizeToTrayChanged != nil {
			s.MinimizeToTrayChanged(isChecked)
		}
	}

	s.aboutCard.OnClicked = func() {
		if s.CheckUpdateSig != nil {
			s.CheckUpdateSig()
		}
	}
	s.feedbackCard.OnClicked = func() {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3(AppFeedback))
	}
}
