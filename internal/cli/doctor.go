package cli

import (
	"fmt"
	"os"

	"github.com/meherpanguluri/lightshell/internal/compat"
)

// Doctor runs compatibility checks on the project.
func Doctor() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	issues, err := compat.ScanProject(dir)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No compatibility issues found.")
		return nil
	}

	fmt.Println("LightShell Compatibility Report")
	fmt.Println("================================")
	fmt.Println()

	errors := 0
	warnings := 0
	autoFixed := 0

	currentFile := ""
	for _, issue := range issues {
		if issue.File != currentFile {
			if currentFile != "" {
				fmt.Println()
			}
			currentFile = issue.File
			fmt.Println(issue.File)
		}

		marker := "warning"
		if issue.Severity == "error" {
			marker = "error"
			errors++
		} else {
			warnings++
		}
		if issue.AutoFix {
			autoFixed++
		}

		fmt.Printf("  %s  line %d: %s\n", severityIcon(issue.Severity), issue.Line, issue.Title)
		if issue.Fix != "" {
			if issue.AutoFix {
				fmt.Printf("     -> Auto-polyfill: %s\n", issue.Fix)
			} else {
				fmt.Printf("     -> %s\n", issue.Fix)
			}
		}
		_ = marker
	}

	fmt.Println()
	fmt.Printf("Summary: %d error(s), %d warning(s)", errors, warnings)
	if autoFixed > 0 {
		fmt.Printf(" (%d auto-polyfilled)", autoFixed)
	}
	fmt.Println()

	return nil
}

func severityIcon(severity string) string {
	switch severity {
	case "error":
		return "X"
	case "warning":
		return "!"
	default:
		return "i"
	}
}
