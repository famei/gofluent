package common

import (
	"strings"
	"unicode"
)

// CharType classifies a character for text wrapping.
type CharType int

const (
	CharSpace CharType = iota
	CharAsian
	CharLatin
)

// TextWrap is a namespace of text wrapping helpers (ported from the Python
// TextWrap class, whose methods are classmethods).
type TextWrap struct{}

// CharWidth returns the display width of a character: East Asian wide
// characters (Fullwidth/Wide) count as 2, everything else as 1.
func CharWidth(ch rune) int {
	if isWideRune(ch) {
		return 2
	}
	return 1
}

// TextWidth returns the total display width of text.
func TextWidth(text string) int {
	w := 0
	for _, ch := range text {
		w += CharWidth(ch)
	}
	return w
}

// CharTypeOf returns the type of a character.
func CharTypeOf(ch rune) CharType {
	if unicode.IsSpace(ch) {
		return CharSpace
	}
	if CharWidth(ch) == 1 {
		return CharLatin
	}
	return CharAsian
}

// ProcessTextWhitespace collapses whitespace runs to single spaces and trims.
func ProcessTextWhitespace(text string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(text), " "))
}

// splitLongToken splits a token into chunks of the given width.
func splitLongToken(token string, width int) []string {
	runes := []rune(token)
	if width <= 0 {
		return []string{string(runes)}
	}
	out := make([]string, 0, (len(runes)+width-1)/width)
	for i := 0; i < len(runes); i += width {
		end := i + width
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}

// tokenizer yields tokens split on char-type boundaries.
func tokenizer(text string) []string {
	var out []string
	var buf strings.Builder
	var last CharType
	first := true

	for _, ch := range text {
		ct := CharTypeOf(ch)
		if !first && (ct != last || ct != CharLatin) {
			if buf.Len() > 0 {
				out = append(out, buf.String())
				buf.Reset()
			}
		}
		buf.WriteRune(ch)
		last = ct
		first = false
	}
	if buf.Len() > 0 {
		out = append(out, buf.String())
	}
	return out
}

// Wrap wraps text according to display width. width is measured in "Chinese
// characters = 2" units; once reports only the first line break.
func Wrap(text string, width int, once bool) (string, bool) {
	lines := strings.Split(text, "\n")
	isWrapped := false
	var wrappedLines []string

	for _, line := range lines {
		line = ProcessTextWhitespace(line)
		if TextWidth(line) > width {
			wrapped, ok := wrapLine(line, width, once)
			wrappedLines = append(wrappedLines, wrapped)
			isWrapped = isWrapped || ok
			if once {
				return strings.Join(wrappedLines, ""), isWrapped
			}
		} else {
			wrappedLines = append(wrappedLines, line)
		}
	}

	return strings.Join(wrappedLines, "\n"), isWrapped
}

func wrapLine(text string, width int, once bool) (string, bool) {
	var lineBuffer strings.Builder
	var wrappedLines []string
	currentWidth := 0

	for _, token := range tokenizer(text) {
		tokenWidth := TextWidth(token)

		if token == " " && currentWidth == 0 {
			continue
		}

		if currentWidth+tokenWidth <= width {
			lineBuffer.WriteString(token)
			currentWidth += tokenWidth
			if currentWidth == width {
				wrappedLines = append(wrappedLines, strings.TrimRight(lineBuffer.String(), " "))
				lineBuffer.Reset()
				currentWidth = 0
			}
		} else {
			if currentWidth != 0 {
				wrappedLines = append(wrappedLines, strings.TrimRight(lineBuffer.String(), " "))
			}
			chunks := splitLongToken(token, width)
			for _, chunk := range chunks[:len(chunks)-1] {
				wrappedLines = append(wrappedLines, strings.TrimRight(chunk, " "))
			}
			lineBuffer.Reset()
			lineBuffer.WriteString(chunks[len(chunks)-1])
			currentWidth = TextWidth(chunks[len(chunks)-1])
		}
	}

	if currentWidth != 0 {
		wrappedLines = append(wrappedLines, strings.TrimRight(lineBuffer.String(), " "))
	}

	if once {
		return strings.Join([]string{wrappedLines[0], strings.Join(wrappedLines[1:], " ")}, "\n"), true
	}
	return strings.Join(wrappedLines, "\n"), true
}

// isWideRune reports whether a rune has East Asian Wide or Fullwidth width.
// It covers the blocks used in practice (CJK, Hangul, fullwidth forms, etc.).
func isWideRune(ch rune) bool {
	switch {
	case ch >= 0x1100 && ch <= 0x115F: // Hangul Jamo
		return true
	case ch >= 0x2E80 && ch <= 0x303E: // CJK Radicals, Kangxi, Symbols & Punctuation
		return true
	case ch >= 0x3041 && ch <= 0x33FF: // Hiragana, Katakana, Bopomofo, CJK compat
		return true
	case ch >= 0x3400 && ch <= 0x4DBF: // CJK Ext A
		return true
	case ch >= 0x4E00 && ch <= 0x9FFF: // CJK Unified Ideographs
		return true
	case ch >= 0xA000 && ch <= 0xA4CF: // Yi
		return true
	case ch >= 0xAC00 && ch <= 0xD7A3: // Hangul Syllables
		return true
	case ch >= 0xF900 && ch <= 0xFAFF: // CJK Compatibility Ideographs
		return true
	case ch >= 0xFE30 && ch <= 0xFE4F: // CJK Compatibility Forms
		return true
	case ch >= 0xFF00 && ch <= 0xFF60: // Fullwidth Forms
		return true
	case ch >= 0xFFE0 && ch <= 0xFFE6: // Fullwidth Signs
		return true
	case ch >= 0x20000 && ch <= 0x2FFFD: // CJK Ext B+
		return true
	case ch >= 0x30000 && ch <= 0x3FFFD: // CJK Ext G+
		return true
	}
	return false
}
