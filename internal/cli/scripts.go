package cli

import (
	_ "embed"
)

//go:embed scripts/polyfills.js
var polyfillsJS string

//go:embed scripts/lightshell.js
var clientJS string
