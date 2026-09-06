package common

import (
	qt "github.com/mappu/miqt/qt"
)

// FluentThemeColor is the Fluent 2 accent colour palette
// (https://www.figma.com/file/iM7EPX8Jn37zjeSezb43cF). Each constant carries
// its hex value; Color() converts it to a *qt.QColor.
type FluentThemeColor string

// Hex returns the "#RRGGBB" value of the color.
func (c FluentThemeColor) Hex() string { return string(c) }

// Color converts the color to a *qt.QColor. The caller owns the result and
// should defer Delete() when it is a temporary value.
func (c FluentThemeColor) Color() *qt.QColor { return qt.NewQColor6(string(c)) }

const (
	YellowGold       FluentThemeColor = "#FFB900"
	Gold             FluentThemeColor = "#FF8C00"
	OrangeBright     FluentThemeColor = "#F7630C"
	OrangeDark       FluentThemeColor = "#CA5010"
	Rust             FluentThemeColor = "#DA3B01"
	PaleRust         FluentThemeColor = "#EF6950"
	BrickRed         FluentThemeColor = "#D13438"
	ModRed           FluentThemeColor = "#FF4343"
	PaleRed          FluentThemeColor = "#E74856"
	Red              FluentThemeColor = "#E81123"
	RoseBright       FluentThemeColor = "#EA005E"
	Rose             FluentThemeColor = "#C30052"
	PlumLight        FluentThemeColor = "#E3008C"
	Plum             FluentThemeColor = "#BF0077"
	OrchidLight      FluentThemeColor = "#BF0077"
	Orchid           FluentThemeColor = "#9A0089"
	DefaultBlue      FluentThemeColor = "#0078D7"
	NavyBlue         FluentThemeColor = "#0063B1"
	PurpleShadow     FluentThemeColor = "#8E8CD8"
	PurpleShadowDark FluentThemeColor = "#6B69D6"
	IrisPastel       FluentThemeColor = "#8764B8"
	IrisSpring       FluentThemeColor = "#744DA9"
	VioletRedLight   FluentThemeColor = "#B146C2"
	VioletRed        FluentThemeColor = "#881798"
	CoolBlueBright   FluentThemeColor = "#0099BC"
	CoolBlue         FluentThemeColor = "#2D7D9A"
	Seafoam          FluentThemeColor = "#00B7C3"
	SeafoamTeal      FluentThemeColor = "#038387"
	MintLight        FluentThemeColor = "#00B294"
	MintDark         FluentThemeColor = "#018574"
	TurfGreen        FluentThemeColor = "#00CC6A"
	SportGreen       FluentThemeColor = "#10893E"
	Gray             FluentThemeColor = "#7A7574"
	GrayBrown        FluentThemeColor = "#5D5A58"
	SteelBlue        FluentThemeColor = "#68768A"
	MetalBlue        FluentThemeColor = "#515C6B"
	PaleMoss         FluentThemeColor = "#567C73"
	Moss             FluentThemeColor = "#486860"
	MeadowGreen      FluentThemeColor = "#498205"
	Green            FluentThemeColor = "#107C10"
	Overcast         FluentThemeColor = "#767676"
	Storm            FluentThemeColor = "#4C4A48"
	BlueGray         FluentThemeColor = "#69797E"
	GrayDark         FluentThemeColor = "#4A5459"
	LiddyGreen       FluentThemeColor = "#647C64"
	Sage             FluentThemeColor = "#525E54"
	CamouflageDesert FluentThemeColor = "#847545"
	Camouflage       FluentThemeColor = "#7E735F"
)

// FluentSystemColor pairs a light and a dark foreground/background color used
// for success/caution/critical message accents.
type FluentSystemColor struct {
	Light string
	Dark  string
}

// Color returns the color for the given theme (ThemeAuto follows the current
// theme).
func (s FluentSystemColor) Color(theme Theme) *qt.QColor {
	if IsDarkThemeMode(theme) {
		return qt.NewQColor6(s.Dark)
	}
	return qt.NewQColor6(s.Light)
}

var (
	SuccessForeground  = FluentSystemColor{Light: "#0f7b0f", Dark: "#6ccb5f"}
	CautionForeground  = FluentSystemColor{Light: "#9d5d00", Dark: "#fce100"}
	CriticalForeground = FluentSystemColor{Light: "#c42b1c", Dark: "#ff99a4"}

	SuccessBackground  = FluentSystemColor{Light: "#dff6dd", Dark: "#393d1b"}
	CautionBackground  = FluentSystemColor{Light: "#fff4ce", Dark: "#433519"}
	CriticalBackground = FluentSystemColor{Light: "#fde7e9", Dark: "#442726"}
)

// ValidColor returns color if it is valid, otherwise the default color.
func ValidColor(color, def *qt.QColor) *qt.QColor {
	if color != nil && color.IsValid() {
		return color
	}
	return def
}

// FallbackThemeColor returns color if valid, otherwise the current theme color.
func FallbackThemeColor(color *qt.QColor) *qt.QColor {
	if color != nil && color.IsValid() {
		return color
	}
	return themeColor()
}

// AutoFallbackThemeColor picks dark or light by the current theme and falls
// back to the theme color when invalid.
func AutoFallbackThemeColor(light, dark *qt.QColor) *qt.QColor {
	if IsDarkTheme() {
		return FallbackThemeColor(dark)
	}
	return FallbackThemeColor(light)
}
