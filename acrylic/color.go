package acrylic

import (
	"strconv"
	"strings"
)

// parseHexColor parses a CSS-style hex colour string into 0..255 components.
//
// Accepted forms (the leading '#' is optional):
//
//	#RGB       #RGBA
//	#RRGGBB    #RRGGBBAA
//
// When no alpha is given, a is 255. It returns ok == false for malformed input,
// in which case the caller keeps its previous colour.
func parseHexColor(s string) (r, g, b, a int, ok bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")

	switch len(s) {
	case 3: // RGB -> RRGGBB
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]}) + "ff"
	case 4: // RGBA -> RRGGBBAA
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2], s[3], s[3]})
	case 6: // RRGGBB
		s += "ff"
	case 8: // RRGGBBAA
	default:
		return 0, 0, 0, 0, false
	}

	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, 0, false
	}
	r = int((v >> 24) & 0xff)
	g = int((v >> 16) & 0xff)
	b = int((v >> 8) & 0xff)
	a = int(v & 0xff)
	return r, g, b, a, true
}

// clamp255 limits v to the 0..255 range.
func clamp255(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
