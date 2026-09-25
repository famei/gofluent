//go:build !linux

package resources

// MicrosoftYaHei is empty off Linux: the CJK font is embedded for Linux only (see
// font_linux.go), because Windows and macOS ship Microsoft YaHei themselves and the
// collection would add 18.8 MB to every binary for nothing. common.ensureCJKFont treats the
// empty slice as "nothing to register".
var MicrosoftYaHei []byte
