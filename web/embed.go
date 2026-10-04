// Package web embeds the browser UI (plain HTML, CSS and JS, no build step).
package web

import "embed"

//go:embed static
var Static embed.FS
