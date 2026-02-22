package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// --- Parameter extraction helpers ---

func getString(params map[string]any, key string, defaultVal string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getInt(params map[string]any, key string, defaultVal int) int {
	if v, ok := params[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return defaultVal
}

func getBool(params map[string]any, key string, defaultVal bool) bool {
	if v, ok := params[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func getMap(params map[string]any, key string) map[string]any {
	if v, ok := params[key]; ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return nil
}

// validProjectName checks that a project name is lowercase alphanumeric with hyphens.
var validProjectName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// getProjectDir returns the current project directory, safely reading with the mutex.
func (s *Server) getProjectDir() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.projectDir
}

// safePath resolves a relative path within the project directory and ensures it
// doesn't escape via traversal. Returns the absolute path or an error.
func (s *Server) safePath(relPath string) (string, error) {
	projDir := s.getProjectDir()
	if relPath == "" {
		return projDir, nil
	}

	// Join and clean the path
	abs := filepath.Join(projDir, relPath)
	abs = filepath.Clean(abs)

	// Resolve symlinks if the path exists
	if _, err := os.Lstat(abs); err == nil {
		real, err := filepath.EvalSymlinks(abs)
		if err != nil {
			return "", fmt.Errorf("cannot resolve path: %s", relPath)
		}
		abs = real
	} else {
		// Path doesn't exist yet. Resolve the existing parent directory
		// to catch symlinks in the parent chain pointing outside the project.
		parent := filepath.Dir(abs)
		if _, statErr := os.Lstat(parent); statErr == nil {
			realParent, evalErr := filepath.EvalSymlinks(parent)
			if evalErr != nil {
				return "", fmt.Errorf("cannot resolve parent path: %s", relPath)
			}
			abs = filepath.Join(realParent, filepath.Base(abs))
		}
	}

	// Ensure it's within the project directory
	cleanProjDir := filepath.Clean(projDir)
	if abs != cleanProjDir && !strings.HasPrefix(abs, cleanProjDir+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal blocked: %s resolves outside project directory", relPath)
	}

	return abs, nil
}

// requireDevRunning returns an error if the dev process is not running.
func (s *Server) requireDevRunning() error {
	if !s.devProcess.IsRunning() {
		return fmt.Errorf("dev process is not running — call lightshell_dev_start first")
	}
	return nil
}

// registerTools registers all 16 MCP tools on the server.
func (s *Server) registerTools() {
	s.registerCreateProject()
	s.registerWriteFile()
	s.registerReadFile()
	s.registerListFiles()
	s.registerDevStart()
	s.registerDevStop()
	s.registerScreenshot()
	s.registerGetConsole()
	s.registerBuild()
	s.registerGetDOM()
	s.registerExecuteJS()
	s.registerGetConfig()
	s.registerUpdateConfig()
	s.registerDoctor()
	s.registerHotReload()
	s.registerPackage()
}

// --- Tool 1: lightshell_create_project ---

func (s *Server) registerCreateProject() {
	s.registerTool(Tool{
		Name:        "lightshell_create_project",
		Description: "Create a new LightShell desktop app project with the standard directory structure (lightshell.json, src/index.html, src/app.js, src/style.css). This initializes everything needed to run 'lightshell dev'.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Project name (lowercase, hyphens allowed, e.g. 'my-app')",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Window title displayed in the title bar (defaults to name)",
				},
				"width": map[string]any{
					"type":        "number",
					"description": "Initial window width in pixels (default 1024)",
				},
				"height": map[string]any{
					"type":        "number",
					"description": "Initial window height in pixels (default 768)",
				},
			},
			"required": []string{"name"},
		},
		Handler: s.handleCreateProject,
	})
}

func (s *Server) handleCreateProject(params map[string]any) (any, error) {
	name := getString(params, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if !validProjectName.MatchString(name) {
		return nil, fmt.Errorf("invalid project name %q: must be lowercase alphanumeric with hyphens, starting with a letter", name)
	}

	title := getString(params, "title", name)
	width := getInt(params, "width", 1024)
	height := getInt(params, "height", 768)

	// Determine where to create the project. If s.projectDir already is a
	// LightShell project (has lightshell.json), create as a sibling.
	baseDir := s.projectDir
	if _, err := os.Stat(filepath.Join(s.projectDir, "lightshell.json")); err == nil {
		baseDir = filepath.Dir(s.projectDir)
	}

	projDir := filepath.Join(baseDir, name)

	// Check if directory already exists
	if _, err := os.Stat(projDir); err == nil {
		return nil, fmt.Errorf("directory %q already exists", projDir)
	}

	// Create directory structure
	srcDir := filepath.Join(projDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Write lightshell.json
	config := map[string]any{
		"name":    name,
		"version": "1.0.0",
		"entry":   "src/index.html",
		"window": map[string]any{
			"title":     title,
			"width":     width,
			"height":    height,
			"minWidth":  400,
			"minHeight": 300,
			"resizable": true,
			"frameless": false,
		},
		"permissions": []string{"fs", "dialog", "clipboard", "shell", "notification"},
		"tray":        false,
		"build": map[string]any{
			"icon":  "",
			"appId": "com.lightshell." + name,
		},
	}
	configBytes, _ := json.MarshalIndent(config, "", "  ")

	files := map[string]string{
		"lightshell.json": string(configBytes),
		"src/index.html": `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>` + title + `</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <main>
    <h1>` + title + `</h1>
    <p>Your LightShell app is running.</p>
    <div id="info"></div>
  </main>
  <script src="app.js"></script>
</body>
</html>`,
		"src/app.js": `async function init() {
  const platform = await lightshell.system.platform()
  const arch = await lightshell.system.arch()
  const info = document.getElementById('info')
  info.textContent = ` + "`Running on ${platform}/${arch}`" + `
}

init()`,
		"src/style.css": `* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
               Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  background: #f5f5f7;
  color: #1d1d1f;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
}

main {
  text-align: center;
  padding: 2rem;
}

h1 {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

p {
  color: #6e6e73;
  margin-bottom: 1rem;
}

#info {
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 0.875rem;
  color: #86868b;
  background: #e8e8ed;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  display: inline-block;
}`,
	}

	var createdFiles []string
	for relPath, content := range files {
		absPath := filepath.Join(projDir, relPath)
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", relPath, err)
		}
		createdFiles = append(createdFiles, relPath)
	}

	// Stop dev process if running before changing project directory
	if s.devProcess.IsRunning() {
		s.devProcess.Stop()
	}

	// Update server project directory to the new project (protected by mutex)
	s.mu.Lock()
	s.projectDir = projDir
	s.mu.Unlock()

	return map[string]any{
		"projectPath": projDir,
		"files":       createdFiles,
	}, nil
}

// --- Tool 2: lightshell_write_file ---

func (s *Server) registerWriteFile() {
	s.registerTool(Tool{
		Name:        "lightshell_write_file",
		Description: "Write or overwrite a file in the LightShell project. Creates parent directories automatically. Path is relative to the project root.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path relative to project root (e.g. 'src/index.html')",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "File content to write",
				},
			},
			"required": []string{"path", "content"},
		},
		Handler: s.handleWriteFile,
	})
}

func (s *Server) handleWriteFile(params map[string]any) (any, error) {
	relPath := getString(params, "path", "")
	content := getString(params, "content", "")
	if relPath == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := s.safePath(relPath)
	if err != nil {
		return nil, err
	}

	// Check if the file already exists (for the created flag)
	_, statErr := os.Stat(absPath)
	existed := statErr == nil

	// Create parent directories
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]any{
		"path":    absPath,
		"size":    len(content),
		"created": !existed,
	}, nil
}

// --- Tool 3: lightshell_read_file ---

func (s *Server) registerReadFile() {
	s.registerTool(Tool{
		Name:        "lightshell_read_file",
		Description: "Read the contents of a file in the LightShell project. Path is relative to the project root.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path relative to project root (e.g. 'src/index.html')",
				},
			},
			"required": []string{"path"},
		},
		Handler: s.handleReadFile,
	})
}

func (s *Server) handleReadFile(params map[string]any) (any, error) {
	relPath := getString(params, "path", "")
	if relPath == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := s.safePath(relPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", relPath)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]any{
		"path":    absPath,
		"content": string(data),
		"size":    len(data),
	}, nil
}

// --- Tool 4: lightshell_list_files ---

func (s *Server) registerListFiles() {
	s.registerTool(Tool{
		Name:        "lightshell_list_files",
		Description: "List all files and directories in the LightShell project (or a subdirectory). Excludes hidden files, node_modules, and dist/.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Subdirectory to list, relative to project root (default: '.')",
				},
			},
		},
		Handler: s.handleListFiles,
	})
}

func (s *Server) handleListFiles(params map[string]any) (any, error) {
	relPath := getString(params, "path", ".")

	absPath, err := s.safePath(relPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", relPath)
		}
		return nil, fmt.Errorf("failed to stat: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", relPath)
	}

	var files []map[string]any

	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip errors, keep walking
		}

		// Get path relative to project root for display
		rel, _ := filepath.Rel(s.projectDir, path)
		if rel == "." {
			return nil // skip the root itself
		}

		name := d.Name()

		// Skip hidden files/dirs
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common non-project dirs
		if d.IsDir() && (name == "node_modules" || name == "dist" || name == "__pycache__") {
			return filepath.SkipDir
		}

		// Skip symlinks to prevent leaking paths outside the project
		if d.Type()&os.ModeSymlink != 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return nil // skip on error
		}

		files = append(files, map[string]any{
			"path":  rel,
			"size":  info.Size(),
			"isDir": d.IsDir(),
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if files == nil {
		files = []map[string]any{}
	}

	return map[string]any{
		"files": files,
	}, nil
}

// --- Tool 5: lightshell_dev_start ---

func (s *Server) registerDevStart() {
	s.registerTool(Tool{
		Name:        "lightshell_dev_start",
		Description: "Start the LightShell dev server. This launches the app window with hot-reload enabled and opens a socket for MCP commands (screenshot, console, DOM inspection, JS execution).",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: s.handleDevStart,
	})
}

func (s *Server) handleDevStart(params map[string]any) (any, error) {
	// Verify we have a valid project
	if _, err := os.Stat(filepath.Join(s.projectDir, "lightshell.json")); err != nil {
		return nil, fmt.Errorf("no lightshell.json found in %s — create a project first with lightshell_create_project", s.projectDir)
	}

	// Create a fresh DevProcessManager pointing at the current projectDir
	s.devProcess = NewDevProcessManager(s.projectDir)

	if err := s.devProcess.Start(); err != nil {
		return nil, fmt.Errorf("failed to start dev server: %w", err)
	}

	return map[string]any{
		"status":     "running",
		"projectDir": s.projectDir,
	}, nil
}

// --- Tool 6: lightshell_dev_stop ---

func (s *Server) registerDevStop() {
	s.registerTool(Tool{
		Name:        "lightshell_dev_stop",
		Description: "Stop the running LightShell dev server and close the app window.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: s.handleDevStop,
	})
}

func (s *Server) handleDevStop(params map[string]any) (any, error) {
	if err := s.devProcess.Stop(); err != nil {
		return nil, fmt.Errorf("failed to stop dev server: %w", err)
	}

	return map[string]any{
		"status": "stopped",
	}, nil
}

// --- Tool 7: lightshell_screenshot ---

func (s *Server) registerScreenshot() {
	s.registerTool(Tool{
		Name:        "lightshell_screenshot",
		Description: "Capture a PNG screenshot of the running LightShell app window. Returns the image directly so you can see what the user's app looks like.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"delay": map[string]any{
					"type":        "number",
					"description": "Milliseconds to wait before capturing (default 500). Useful for animations or async rendering.",
				},
			},
		},
		Handler: s.handleScreenshot,
	})
}

func (s *Server) handleScreenshot(params map[string]any) (any, error) {
	if err := s.requireDevRunning(); err != nil {
		return nil, err
	}

	delay := getInt(params, "delay", 500)

	resp, err := s.devProcess.SendCommand(MCPCommand{
		Cmd:   "screenshot",
		Delay: delay,
	})
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}

	if resp.Image == "" {
		return nil, fmt.Errorf("screenshot returned empty image")
	}

	// Return pre-formatted MCP content with image type
	return map[string]any{
		"content": []map[string]any{
			{
				"type":     "image",
				"data":     resp.Image,
				"mimeType": "image/png",
			},
		},
	}, nil
}

// --- Tool 8: lightshell_get_console ---

func (s *Server) registerGetConsole() {
	s.registerTool(Tool{
		Name:        "lightshell_get_console",
		Description: "Retrieve console log entries (console.log, console.error, etc.) from the running LightShell app. Useful for debugging JavaScript errors.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"lines": map[string]any{
					"type":        "number",
					"description": "Number of log entries to return (default 50, max 200)",
				},
				"level": map[string]any{
					"type":        "string",
					"description": "Filter by level: 'all', 'log', 'warn', 'error', 'info' (default 'all')",
					"enum":        []string{"all", "log", "warn", "error", "info"},
				},
				"clear": map[string]any{
					"type":        "boolean",
					"description": "Clear the console buffer after reading (default false)",
				},
			},
		},
		Handler: s.handleGetConsole,
	})
}

func (s *Server) handleGetConsole(params map[string]any) (any, error) {
	if err := s.requireDevRunning(); err != nil {
		return nil, err
	}

	lines := getInt(params, "lines", 50)
	if lines > 200 {
		lines = 200
	}
	if lines < 1 {
		lines = 50
	}
	level := getString(params, "level", "all")
	clear := getBool(params, "clear", false)

	resp, err := s.devProcess.SendCommand(MCPCommand{
		Cmd:   "console",
		Lines: lines,
		Level: level,
		Clear: clear,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get console: %w", err)
	}

	entries := resp.Entries
	if entries == nil {
		entries = []ConsoleEntry{}
	}

	return map[string]any{
		"entries": entries,
		"count":   len(entries),
	}, nil
}

// --- Tool 9: lightshell_build ---

func (s *Server) registerBuild() {
	s.registerTool(Tool{
		Name:        "lightshell_build",
		Description: "Build the LightShell app for production. Creates a native .app bundle (macOS) or AppImage (Linux). Stops the dev server first if it's running.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target": map[string]any{
					"type":        "string",
					"description": "Build target: 'default' (native bundle), 'dmg', 'deb', 'rpm', 'all' (default: 'default')",
					"enum":        []string{"default", "dmg", "deb", "rpm", "all"},
				},
			},
		},
		Handler: s.handleBuild,
	})
}

func (s *Server) handleBuild(params map[string]any) (any, error) {
	// Verify we have a valid project
	if _, err := os.Stat(filepath.Join(s.projectDir, "lightshell.json")); err != nil {
		return nil, fmt.Errorf("no lightshell.json found in %s — create a project first", s.projectDir)
	}

	// Stop dev server if running
	if s.devProcess.IsRunning() {
		s.devProcess.Stop()
	}

	target := getString(params, "target", "default")

	// Run the build by executing lightshell build as a subprocess
	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not find lightshell binary: %w", err)
	}

	args := []string{"build"}
	if target != "" && target != "default" {
		args = append(args, "--target", target)
	}

	cmd := exec.Command(selfPath, args...)
	cmd.Dir = s.projectDir
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return nil, fmt.Errorf("build failed: %s\n%s", err, outputStr)
	}

	// Try to find the output path in the dist/ directory
	distDir := filepath.Join(s.projectDir, "dist")
	var outputPath string
	var outputSize int64

	if entries, err := os.ReadDir(distDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			// Look for the .app bundle or other output
			info, err := e.Info()
			if err == nil {
				outputPath = filepath.Join(distDir, e.Name())
				outputSize = info.Size()
			}
		}
		// If no dirs found, look for files (like .dmg, .deb, .rpm)
		if outputPath == "" {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				info, err := e.Info()
				if err == nil {
					outputPath = filepath.Join(distDir, e.Name())
					outputSize = info.Size()
				}
			}
		}
	}

	return map[string]any{
		"target":     target,
		"outputPath": outputPath,
		"size":       outputSize,
		"output":     outputStr,
	}, nil
}

// --- Tool 10: lightshell_get_dom ---

func (s *Server) registerGetDOM() {
	s.registerTool(Tool{
		Name:        "lightshell_get_dom",
		Description: "Inspect the DOM tree of the running LightShell app. Returns the HTML structure for a given CSS selector, useful for understanding app layout and debugging rendering.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"selector": map[string]any{
					"type":        "string",
					"description": "CSS selector to query (default: 'body')",
				},
				"depth": map[string]any{
					"type":        "number",
					"description": "Maximum depth of DOM tree to return (default 5)",
				},
			},
		},
		Handler: s.handleGetDOM,
	})
}

func (s *Server) handleGetDOM(params map[string]any) (any, error) {
	if err := s.requireDevRunning(); err != nil {
		return nil, err
	}

	selector := getString(params, "selector", "body")
	depth := getInt(params, "depth", 5)

	resp, err := s.devProcess.SendCommand(MCPCommand{
		Cmd:      "dom",
		Selector: selector,
		Depth:    depth,
	})
	if err != nil {
		return nil, fmt.Errorf("DOM inspection failed: %w", err)
	}

	return map[string]any{
		"html":     resp.HTML,
		"selector": selector,
	}, nil
}

// --- Tool 11: lightshell_execute_js ---

func (s *Server) registerExecuteJS() {
	s.registerTool(Tool{
		Name:        "lightshell_execute_js",
		Description: "Execute JavaScript code in the running LightShell app's webview context. The code runs in the page and can access the DOM, lightshell APIs, and all page-level variables. Returns the result of the last expression.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code": map[string]any{
					"type":        "string",
					"description": "JavaScript code to execute in the webview",
				},
			},
			"required": []string{"code"},
		},
		Handler: s.handleExecuteJS,
	})
}

func (s *Server) handleExecuteJS(params map[string]any) (any, error) {
	if err := s.requireDevRunning(); err != nil {
		return nil, err
	}

	code := getString(params, "code", "")
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	resp, err := s.devProcess.SendCommand(MCPCommand{
		Cmd:  "eval",
		Code: code,
	})
	if err != nil {
		return nil, fmt.Errorf("JS execution failed: %w", err)
	}

	// resp.Result is json.RawMessage — parse it to return as native value
	var result any
	if resp.Result != nil {
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			// If it can't be parsed, return as string
			result = string(resp.Result)
		}
	}

	return map[string]any{
		"result": result,
	}, nil
}

// --- Tool 12: lightshell_get_config ---

func (s *Server) registerGetConfig() {
	s.registerTool(Tool{
		Name:        "lightshell_get_config",
		Description: "Read the current lightshell.json configuration file for the project. Returns the full config object including window settings, permissions, build options, etc.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: s.handleGetConfig,
	})
}

func (s *Server) handleGetConfig(params map[string]any) (any, error) {
	configPath := filepath.Join(s.projectDir, "lightshell.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no lightshell.json found in %s — create a project first", s.projectDir)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse lightshell.json: %w", err)
	}

	return config, nil
}

// --- Tool 13: lightshell_update_config ---

func (s *Server) registerUpdateConfig() {
	s.registerTool(Tool{
		Name:        "lightshell_update_config",
		Description: "Update the lightshell.json configuration. Provide a partial config object — it will be shallow-merged with the existing config. Set a key to null to delete it.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"patch": map[string]any{
					"type":        "object",
					"description": "Partial config to merge into lightshell.json. Keys with null values are deleted.",
				},
			},
			"required": []string{"patch"},
		},
		Handler: s.handleUpdateConfig,
	})
}

func (s *Server) handleUpdateConfig(params map[string]any) (any, error) {
	patch := getMap(params, "patch")
	if patch == nil {
		return nil, fmt.Errorf("patch is required and must be an object")
	}

	configPath := filepath.Join(s.projectDir, "lightshell.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no lightshell.json found in %s — create a project first", s.projectDir)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse lightshell.json: %w", err)
	}

	// Deep merge patch into config
	deepMerge(config, patch)

	// Write back
	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}

	return config, nil
}

// deepMerge merges src into dst. For nested maps, it recurses. If a src value
// is nil (JSON null), the key is deleted from dst.
func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if v == nil {
			// null value means delete the key
			delete(dst, k)
			continue
		}

		srcMap, srcIsMap := v.(map[string]any)
		dstVal, dstExists := dst[k]

		if srcIsMap && dstExists {
			if dstMap, dstIsMap := dstVal.(map[string]any); dstIsMap {
				deepMerge(dstMap, srcMap)
				continue
			}
		}

		dst[k] = v
	}
}

// --- Tool 14: lightshell_doctor ---

func (s *Server) registerDoctor() {
	s.registerTool(Tool{
		Name:        "lightshell_doctor",
		Description: "Run diagnostics on the LightShell project. Checks for common issues like missing dependencies, invalid config, cross-platform compatibility problems, and unsupported API usage.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: s.handleDoctor,
	})
}

func (s *Server) handleDoctor(params map[string]any) (any, error) {
	// Verify we have a valid project
	if _, err := os.Stat(filepath.Join(s.projectDir, "lightshell.json")); err != nil {
		return nil, fmt.Errorf("no lightshell.json found in %s — create a project first", s.projectDir)
	}

	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not find lightshell binary: %w", err)
	}

	cmd := exec.Command(selfPath, "doctor")
	cmd.Dir = s.projectDir
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	// doctor may exit non-zero if it finds issues — that's not an error for us
	return map[string]any{
		"output": outputStr,
		"passed": err == nil,
	}, nil
}

// --- Tool 15: lightshell_hot_reload ---

func (s *Server) registerHotReload() {
	s.registerTool(Tool{
		Name:        "lightshell_hot_reload",
		Description: "Trigger an immediate hot reload of the running LightShell app. The webview will reload the page with the latest file changes. Useful after writing files to see changes instantly.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: s.handleHotReload,
	})
}

func (s *Server) handleHotReload(params map[string]any) (any, error) {
	if err := s.requireDevRunning(); err != nil {
		return nil, err
	}

	_, err := s.devProcess.SendCommand(MCPCommand{
		Cmd: "reload",
	})
	if err != nil {
		return nil, fmt.Errorf("hot reload failed: %w", err)
	}

	return map[string]any{
		"status": "reloaded",
	}, nil
}

// --- Tool 16: lightshell_package ---

func (s *Server) registerPackage() {
	s.registerTool(Tool{
		Name:        "lightshell_package",
		Description: "Package the LightShell app into a distributable format. Supports DMG (macOS), .deb (Debian/Ubuntu), .rpm (Fedora), or all formats for the current OS. Optionally code-sign on macOS.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target": map[string]any{
					"type":        "string",
					"description": "Package format: 'dmg', 'deb', 'rpm', 'all'",
					"enum":        []string{"dmg", "deb", "rpm", "all"},
				},
				"sign": map[string]any{
					"type":        "boolean",
					"description": "Code-sign the package (macOS only, requires signing identity in config)",
				},
			},
			"required": []string{"target"},
		},
		Handler: s.handlePackage,
	})
}

func (s *Server) handlePackage(params map[string]any) (any, error) {
	// Verify we have a valid project
	if _, err := os.Stat(filepath.Join(s.projectDir, "lightshell.json")); err != nil {
		return nil, fmt.Errorf("no lightshell.json found in %s — create a project first", s.projectDir)
	}

	// Stop dev server if running
	if s.devProcess.IsRunning() {
		s.devProcess.Stop()
	}

	target := getString(params, "target", "")
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}
	sign := getBool(params, "sign", false)

	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not find lightshell binary: %w", err)
	}

	args := []string{"build", "--target", target}
	if sign {
		args = append(args, "--sign")
	}

	cmd := exec.Command(selfPath, args...)
	cmd.Dir = s.projectDir
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return nil, fmt.Errorf("packaging failed: %s\n%s", err, outputStr)
	}

	// Collect output files from dist/
	distDir := filepath.Join(s.projectDir, "dist")
	var packages []map[string]any

	if entries, err := os.ReadDir(distDir); err == nil {
		for _, e := range entries {
			info, infoErr := e.Info()
			if infoErr != nil {
				continue
			}
			name := e.Name()
			// Include recognized package formats and app bundles
			if strings.HasSuffix(name, ".dmg") ||
				strings.HasSuffix(name, ".deb") ||
				strings.HasSuffix(name, ".rpm") ||
				strings.HasSuffix(name, ".AppImage") ||
				strings.HasSuffix(name, ".app") {
				packages = append(packages, map[string]any{
					"path": filepath.Join(distDir, name),
					"size": info.Size(),
					"name": name,
				})
			}
		}
	}

	if packages == nil {
		packages = []map[string]any{}
	}

	return map[string]any{
		"target":   target,
		"signed":   sign,
		"packages": packages,
		"output":   outputStr,
	}, nil
}
