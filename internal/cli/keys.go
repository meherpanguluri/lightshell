package cli

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// Keys handles the `lightshell keys` command.
func Keys(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: lightshell keys generate")
	}

	switch args[0] {
	case "generate":
		return keysGenerate()
	default:
		return fmt.Errorf("unknown keys subcommand: %s\n\nUsage: lightshell keys generate", args[0])
	}
}

func keysGenerate() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	keyDir := filepath.Join(home, ".lightshell")
	if err := os.MkdirAll(keyDir, 0o700); err != nil {
		return fmt.Errorf("could not create key directory: %w", err)
	}

	privKeyPath := filepath.Join(keyDir, "signing-key.pem")
	pubKeyPath := filepath.Join(keyDir, "signing-key.pub")

	// Check if keys already exist
	if _, err := os.Stat(privKeyPath); err == nil {
		return fmt.Errorf("signing key already exists at %s\n\nTo regenerate, delete the existing key files first:\n  rm %s %s", privKeyPath, privKeyPath, pubKeyPath)
	}

	// Generate Ed25519 keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Encode private key as PEM
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: priv.Seed(),
	})

	if err := os.WriteFile(privKeyPath, privPEM, 0o600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Encode public key as base64
	pubB64 := base64.StdEncoding.EncodeToString(pub)
	if err := os.WriteFile(pubKeyPath, []byte(pubB64+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	// Try to update lightshell.json if it exists in the current directory
	dir, err := os.Getwd()
	if err == nil {
		configPath := filepath.Join(dir, "lightshell.json")
		if _, err := os.Stat(configPath); err == nil {
			if updateErr := writePublicKeyToConfig(configPath, pubB64); updateErr != nil {
				fmt.Printf("Note: could not update lightshell.json: %v\n", updateErr)
			} else {
				fmt.Printf("Updated lightshell.json with public key\n")
			}
		}
	}

	fmt.Printf("Signing keys generated:\n")
	fmt.Printf("  Private key: %s (keep this secret!)\n", privKeyPath)
	fmt.Printf("  Public key:  %s\n\n", pubKeyPath)
	fmt.Printf("Public key (base64):\n  %s\n\n", pubB64)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Add the public key to your lightshell.json:\n")
	fmt.Printf("     \"updater\": { \"publicKey\": \"%s\" }\n\n", pubB64)
	fmt.Printf("  2. Use `lightshell release` to sign and publish updates\n")
	fmt.Printf("  3. Keep the private key safe — it's needed to sign releases\n")

	return nil
}

// writePublicKeyToConfig reads lightshell.json, adds the public key to the
// updater section, and writes it back.
func writePublicKeyToConfig(configPath, pubKey string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	updater, ok := cfg["updater"].(map[string]any)
	if !ok {
		updater = make(map[string]any)
	}
	updater["publicKey"] = pubKey
	cfg["updater"] = updater

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, append(out, '\n'), 0o644)
}

// loadPrivateKey loads the Ed25519 private key from the default location.
func loadPrivateKey() (ed25519.PrivateKey, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	privKeyPath := filepath.Join(home, ".lightshell", "signing-key.pem")
	data, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("signing key not found at %s\n\nRun `lightshell keys generate` to create one", privKeyPath)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "ED25519 PRIVATE KEY" {
		return nil, fmt.Errorf("invalid signing key format at %s", privKeyPath)
	}

	if len(block.Bytes) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid signing key size at %s", privKeyPath)
	}

	return ed25519.NewKeyFromSeed(block.Bytes), nil
}
