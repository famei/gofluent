// Package examples embeds the source code of every runnable example so the
// gallery's expandable "Source code" area can derive a short call pseudo-code
// snippet (and fall back to the real Go source) without depending on the
// repository layout at runtime.
package examples

import (
	"embed"
	"strings"
)

// sourceFS embeds the example entry points at both supported depths:
//   - <name>/main.go            (e.g. hello/main.go)
//   - <category>/<name>/main.go (e.g. scroll/pips_pager/main.go)
//
// go:embed patterns cannot use "..", which is why this file lives directly
// inside gofluent/examples/.
//
//go:embed */main.go
//go:embed */*/main.go
var sourceFS embed.FS

// Source returns the embedded content of the example main.go at the given
// slash-separated relative path (e.g. "scroll/pips_pager/main.go"). The second
// result reports whether the file was found.
func Source(rel string) (string, bool) {
	rel = strings.ReplaceAll(rel, "\\", "/")
	rel = strings.TrimPrefix(rel, "./")
	data, err := sourceFS.ReadFile(rel)
	if err != nil {
		return "", false
	}
	return string(data), true
}
