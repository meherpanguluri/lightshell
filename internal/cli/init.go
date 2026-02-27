package cli

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:templates/default
var defaultTemplate embed.FS

//go:embed all:templates/react
var reactTemplate embed.FS

//go:embed all:templates/svelte
var svelteTemplate embed.FS

// Init creates a new LightShell project.
func Init(name string, templateName string) error {
	if name == "" {
		name = "my-lightshell-app"
	}

	if templateName == "" {
		templateName = "default"
	}

	// Validate name
	if strings.ContainsAny(name, " /\\") {
		return fmt.Errorf("project name cannot contain spaces or slashes: %q", name)
	}

	// Select template
	var tmplFS embed.FS
	var tmplRoot string
	switch templateName {
	case "default":
		tmplFS = defaultTemplate
		tmplRoot = "templates/default"
	case "react":
		tmplFS = reactTemplate
		tmplRoot = "templates/react"
	case "svelte":
		tmplFS = svelteTemplate
		tmplRoot = "templates/svelte"
	default:
		return fmt.Errorf("unknown template %q. Available templates: default, react, svelte", templateName)
	}

	dir, err := filepath.Abs(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("directory already exists: %s", dir)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	// Copy template files
	err = fs.WalkDir(tmplFS, tmplRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path from template root
		relPath, _ := filepath.Rel(tmplRoot, path)
		destPath := filepath.Join(dir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		data, err := tmplFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Replace template variables
		content := string(data)
		content = strings.ReplaceAll(content, "{{NAME}}", name)
		content = strings.ReplaceAll(content, "{{TITLE}}", formatTitle(name))

		return os.WriteFile(destPath, []byte(content), 0o644)
	})

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Created %s", name)
	if templateName != "default" {
		fmt.Printf(" (template: %s)", templateName)
	}
	fmt.Println()
	fmt.Println()
	fmt.Printf("  cd %s\n", name)
	if templateName != "default" {
		fmt.Println("  npm install")
	}
	fmt.Println("  lightshell dev")
	fmt.Println()

	return nil
}

func formatTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
