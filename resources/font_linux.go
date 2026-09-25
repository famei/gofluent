//go:build linux

package resources

import _ "embed"

// MicrosoftYaHei embeds Microsoft YaHei (msyh.ttc, the regular and the UI face) so that a
// Linux system without any Chinese font can still render CJK text instead of drawing empty
// boxes. common.ensureCJKFont registers it, and only when the platform does not provide the
// family already.
//
// The font is embedded on Linux only: Windows and macOS have their own copy of the family,
// and the collection is 18.8 MB, which has no business in their binaries (see
// font_other.go).
//
//go:embed fonts/msyh.ttc
var MicrosoftYaHei []byte
