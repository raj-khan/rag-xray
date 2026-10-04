// Package data embeds the sample corpus so the app works with zero setup.
package data

import "embed"

//go:embed docs/*.md
var Docs embed.FS
