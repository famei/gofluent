package main

import "embed"

// Settings example assets. The upstream settings demo loads these from disk at
// runtime; the Go port embeds them so the static cross-compiled binary carries
// the QSS and translation files (see examples/README.md §3).
//
//go:embed resource/qss/light/*.qss resource/qss/dark/*.qss
var settingsQSS embed.FS

//go:embed resource/i18n/*.qm resource/i18n/*.ts
var settingsI18n embed.FS
