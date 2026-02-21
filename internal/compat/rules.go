package compat

// CompatRule defines a known cross-platform compatibility issue.
type CompatRule struct {
	ID          string
	Severity    string // "error" | "warning" | "info"
	Title       string
	Description string
	Platforms   []string
	Patterns    []string // regex patterns to match in user code
	FileTypes   []string // "css", "js", "html"
	Fix         string
	AutoFix     bool
}

// Issue represents a detected compatibility problem in user code.
type Issue struct {
	File     string
	Line     int
	Rule     CompatRule
	Severity string
	Title    string
	Fix      string
	AutoFix  bool
}

// Rules is the database of known cross-platform issues.
var Rules = []CompatRule{
	{
		ID:        "CSS-001",
		Severity:  "warning",
		Title:     "backdrop-filter — limited on Linux (WebKitGTK)",
		Platforms: []string{"linux"},
		Patterns:  []string{`backdrop-filter`},
		FileTypes: []string{"css", "html"},
		Fix:       "fallback background injected at runtime",
		AutoFix:   true,
	},
	{
		ID:        "CSS-002",
		Severity:  "warning",
		Title:     "system-ui font — renders differently across platforms",
		Platforms: []string{"linux"},
		Patterns:  []string{`font-family:.*system-ui`},
		FileTypes: []string{"css", "html"},
		Fix:       "Use explicit font stack or bundle a web font",
		AutoFix:   false,
	},
	{
		ID:        "CSS-003",
		Severity:  "warning",
		Title:     ":has() selector — limited support on older WebKitGTK",
		Platforms: []string{"linux"},
		Patterns:  []string{`:has\(`},
		FileTypes: []string{"css", "html"},
		Fix:       "Use JavaScript or alternative CSS selectors for broader support",
		AutoFix:   false,
	},
	{
		ID:        "CSS-004",
		Severity:  "warning",
		Title:     "color-mix() — not supported on older WebKitGTK",
		Platforms: []string{"linux"},
		Patterns:  []string{`color-mix\(`},
		FileTypes: []string{"css", "html"},
		Fix:       "Use pre-computed color values instead",
		AutoFix:   false,
	},
	{
		ID:        "CSS-005",
		Severity:  "warning",
		Title:     "CSS nesting — requires WebKitGTK 2.42+",
		Platforms: []string{"linux"},
		Patterns:  []string{`&\s*[.#\[]`},
		FileTypes: []string{"css"},
		Fix:       "Use flat CSS selectors for broader WebKitGTK support",
		AutoFix:   false,
	},
	{
		ID:        "CSS-006",
		Severity:  "warning",
		Title:     "Container Queries — version-dependent WebKitGTK support",
		Platforms: []string{"linux"},
		Patterns:  []string{`@container`},
		FileTypes: []string{"css", "html"},
		Fix:       "Use media queries or resize observers as fallback",
		AutoFix:   false,
	},
	{
		ID:        "CSS-007",
		Severity:  "warning",
		Title:     "View Transitions API — not available in WebKitGTK",
		Platforms: []string{"linux"},
		Patterns:  []string{`view-transition`},
		FileTypes: []string{"css", "html"},
		Fix:       "Use CSS transitions/animations instead",
		AutoFix:   false,
	},
	{
		ID:        "JS-001",
		Severity:  "warning",
		Title:     "structuredClone() — missing on WebKitGTK < 2.40",
		Platforms: []string{"linux"},
		Patterns:  []string{`structuredClone\(`},
		FileTypes: []string{"js", "html"},
		Fix:       "JSON-based clone injected at runtime",
		AutoFix:   true,
	},
	{
		ID:        "JS-002",
		Severity:  "warning",
		Title:     "Intl.Segmenter — not available in WebKitGTK",
		Platforms: []string{"linux"},
		Patterns:  []string{`Intl\.Segmenter`},
		FileTypes: []string{"js", "html"},
		Fix:       "Use a polyfill library or alternative text segmentation",
		AutoFix:   false,
	},
	{
		ID:        "JS-003",
		Severity:  "error",
		Title:     "Navigation API — not available in webviews",
		Platforms: []string{"darwin", "linux"},
		Patterns:  []string{`navigation\.navigate`, `navigation\.back`, `navigation\.addEventListener`},
		FileTypes: []string{"js", "html"},
		Fix:       "Use standard History API or lightshell.window instead",
		AutoFix:   false,
	},
	{
		ID:        "JS-004",
		Severity:  "error",
		Title:     "showOpenFilePicker — not available in webviews",
		Platforms: []string{"darwin", "linux"},
		Patterns:  []string{`showOpenFilePicker`, `showSaveFilePicker`, `showDirectoryPicker`},
		FileTypes: []string{"js", "html"},
		Fix:       "Use lightshell.dialog.open() / lightshell.dialog.save() instead",
		AutoFix:   false,
	},
	{
		ID:        "JS-005",
		Severity:  "error",
		Title:     "Web USB API — not available in webviews",
		Platforms: []string{"darwin", "linux"},
		Patterns:  []string{`navigator\.usb`},
		FileTypes: []string{"js", "html"},
		Fix:       "Web USB is not supported in system webviews",
		AutoFix:   false,
	},
	{
		ID:        "JS-006",
		Severity:  "error",
		Title:     "Web Bluetooth API — not available in webviews",
		Platforms: []string{"darwin", "linux"},
		Patterns:  []string{`navigator\.bluetooth`},
		FileTypes: []string{"js", "html"},
		Fix:       "Web Bluetooth is not supported in system webviews",
		AutoFix:   false,
	},
	{
		ID:        "JS-007",
		Severity:  "error",
		Title:     "Web Serial API — not available in webviews",
		Platforms: []string{"darwin", "linux"},
		Patterns:  []string{`navigator\.serial`},
		FileTypes: []string{"js", "html"},
		Fix:       "Web Serial is not supported in system webviews",
		AutoFix:   false,
	},
	{
		ID:        "JS-008",
		Severity:  "error",
		Title:     "Node.js require() — not available in LightShell",
		Platforms: []string{"darwin", "linux"},
		Patterns:  []string{`require\(['"]\w`},
		FileTypes: []string{"js"},
		Fix:       "Use lightshell.* APIs instead of Node.js modules",
		AutoFix:   false,
	},
	{
		ID:        "JS-009",
		Severity:  "error",
		Title:     "Node.js process global — not available in LightShell",
		Platforms: []string{"darwin", "linux"},
		Patterns:  []string{`process\.env`, `process\.argv`, `process\.cwd`},
		FileTypes: []string{"js"},
		Fix:       "Use lightshell.system.* APIs for environment info",
		AutoFix:   false,
	},
}
