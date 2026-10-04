// Package docs embeds the lesson Markdown so the ragxray binary is self-contained.
package docs

import "embed"

//go:embed lessons/*.md
var Lessons embed.FS
