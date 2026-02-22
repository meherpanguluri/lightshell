package cli

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	lsruntime "github.com/lightshell-dev/lightshell/internal/runtime"
)

// ReleaseFlags holds the parsed flags for the release command.
type ReleaseFlags struct {
	Platform  string // target platform (e.g., "darwin-arm64", "linux-x64")
	Notes     string // release notes text
	NotesFile string // path to release notes file
	Draft     bool   // mark as draft release
	DryRun    bool   // do everything except upload
	NoBuild   bool   // skip the build step, use existing dist/
	Server    string // release server URL (overrides config)
	Token     string // auth token (overrides config)
}

// Release handles the `lightshell release` command.
func Release(args []string) error {
	flags, err := parseReleaseFlags(args)
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := lsruntime.LoadConfig(dir)
	if err != nil {
		return err
	}

	// Resolve server and token from flags or config
	server := flags.Server
	token := flags.Token

	if server == "" {
		server = loadConfigValue("releaseServer")
	}
	if token == "" {
		token = loadConfigValue("releaseToken")
	}

	if server == "" && !flags.DryRun {
		return fmt.Errorf("no release server configured\n\nSet one with:\n  lightshell config set releaseServer https://releases.example.com\n\nOr pass --server on the command line")
	}

	// Determine platform
	platform := flags.Platform
	if platform == "" {
		platform = runtime.GOOS + "-" + runtime.GOARCH
	}

	// Normalize platform names
	platform = normalizePlatform(platform)

	// Build if needed
	if !flags.NoBuild {
		fmt.Println("Building...")
		if err := Build(); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}

	// Find the built artifact in dist/
	distDir := filepath.Join(dir, "dist")
	artifact, err := findArtifact(distDir, cfg.Name)
	if err != nil {
		return fmt.Errorf("could not find built artifact in dist/: %w\n\nRun `lightshell build` first, or use --no-build if you have a pre-built artifact", err)
	}

	fmt.Printf("Artifact: %s\n", artifact)

	// Compute SHA256
	hash, err := computeSHA256(artifact)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}
	fmt.Printf("SHA256: %s\n", hash)

	// Resolve release notes
	notes := flags.Notes
	if notes == "" && flags.NotesFile != "" {
		data, err := os.ReadFile(flags.NotesFile)
		if err != nil {
			return fmt.Errorf("could not read notes file: %w", err)
		}
		notes = string(data)
	}
	if notes == "" {
		notes = fmt.Sprintf("Release %s", cfg.Version)
	}

	// Load signing key
	privKey, err := loadPrivateKey()
	if err != nil {
		return err
	}

	// Create release manifest — set URL before signing
	platformArtifact := PlatformArtifact{SHA256: hash}
	if server != "" {
		platformArtifact.URL = fmt.Sprintf("%s/releases/v%s/%s", strings.TrimSuffix(server, "/"), cfg.Version, filepath.Base(artifact))
	}

	manifest := ReleaseManifest{
		Version: cfg.Version,
		Notes:   notes,
		PubDate: time.Now().UTC().Format(time.RFC3339),
		Draft:   flags.Draft,
		Platforms: map[string]PlatformArtifact{
			platform: platformArtifact,
		},
	}

	// Sign the manifest (after all fields are set, excluding Signature itself)
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}

	signature := ed25519.Sign(privKey, manifestJSON)
	manifest.Signature = base64.StdEncoding.EncodeToString(signature)

	// Print manifest
	manifestOut, _ := json.MarshalIndent(manifest, "", "  ")
	fmt.Printf("\nRelease manifest:\n%s\n\n", string(manifestOut))

	if flags.DryRun {
		fmt.Println("Dry run — skipping upload")

		// Write manifest to dist/ for inspection
		manifestPath := filepath.Join(distDir, "latest.json")
		if err := os.WriteFile(manifestPath, append(manifestOut, '\n'), 0o644); err != nil {
			return fmt.Errorf("failed to write manifest: %w", err)
		}
		fmt.Printf("Manifest written to: %s\n", manifestPath)
		return nil
	}

	// Upload to server
	fmt.Printf("Uploading to %s...\n", server)
	if err := uploadRelease(server, token, artifact, manifest); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf("Released %s v%s for %s\n", cfg.Name, cfg.Version, platform)
	return nil
}

// ReleaseManifest is the JSON manifest published for the auto-updater.
type ReleaseManifest struct {
	Version   string                       `json:"version"`
	Notes     string                       `json:"notes"`
	PubDate   string                       `json:"pub_date"`
	Draft     bool                         `json:"draft,omitempty"`
	Signature string                       `json:"signature,omitempty"`
	Platforms map[string]PlatformArtifact  `json:"platforms"`
}

// PlatformArtifact describes a platform-specific release artifact.
type PlatformArtifact struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

func parseReleaseFlags(args []string) (ReleaseFlags, error) {
	var flags ReleaseFlags

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--platform":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--platform requires a value")
			}
			i++
			flags.Platform = args[i]
		case "--notes":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--notes requires a value")
			}
			i++
			flags.Notes = args[i]
		case "--notes-file":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--notes-file requires a value")
			}
			i++
			flags.NotesFile = args[i]
		case "--draft":
			flags.Draft = true
		case "--dry-run":
			flags.DryRun = true
		case "--no-build":
			flags.NoBuild = true
		case "--server":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--server requires a value")
			}
			i++
			flags.Server = args[i]
		case "--token":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--token requires a value")
			}
			i++
			flags.Token = args[i]
		default:
			return flags, fmt.Errorf("unknown flag: %s\n\nUsage: lightshell release [--platform darwin-arm64] [--notes \"...\"] [--notes-file NOTES.md] [--draft] [--dry-run] [--no-build] [--server URL] [--token TOKEN]", args[i])
		}
	}

	return flags, nil
}

func normalizePlatform(platform string) string {
	// Normalize common aliases
	replacer := strings.NewReplacer(
		"amd64", "x64",
		"x86_64", "x64",
		"aarch64", "arm64",
	)
	return replacer.Replace(platform)
}

func findArtifact(distDir, appName string) (string, error) {
	// Look for known artifact patterns in order of preference
	patterns := []string{
		filepath.Join(distDir, appName+".tar.gz"),
		filepath.Join(distDir, appName+".zip"),
		filepath.Join(distDir, "*.app"),
		filepath.Join(distDir, appName+"*.dmg"),
		filepath.Join(distDir, appName+"*.deb"),
		filepath.Join(distDir, appName+"*.rpm"),
		filepath.Join(distDir, appName+"*.AppImage"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			return matches[0], nil
		}
	}

	// Fall back to any .app directory
	entries, err := os.ReadDir(distDir)
	if err != nil {
		return "", fmt.Errorf("could not read dist directory: %w", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".app") {
			return filepath.Join(distDir, e.Name()), nil
		}
	}

	return "", fmt.Errorf("no artifact found")
}

func computeSHA256(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		// For directories (like .app bundles), hash a tar of the contents
		return computeDirSHA256(path)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func computeDirSHA256(dir string) (string, error) {
	h := sha256.New()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		// Include path name and size in hash for determinism
		fmt.Fprintf(h, "%s:%d:%d\n", rel, info.Size(), info.Mode())
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(h, f); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func uploadRelease(server, token, artifactPath string, manifest ReleaseManifest) error {
	// Create multipart request with manifest + artifact
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add manifest as JSON field
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}
	if err := writer.WriteField("manifest", string(manifestJSON)); err != nil {
		return err
	}

	// Add artifact file
	info, err := os.Stat(artifactPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// For directory artifacts (.app), skip file upload — server handles it differently
		if err := writer.WriteField("artifact_name", filepath.Base(artifactPath)); err != nil {
			return err
		}
	} else {
		part, err := writer.CreateFormFile("artifact", filepath.Base(artifactPath))
		if err != nil {
			return err
		}
		f, err := os.Open(artifactPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(part, f); err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	// Send request
	uploadURL := strings.TrimSuffix(server, "/") + "/api/releases"
	req, err := http.NewRequest("POST", uploadURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
