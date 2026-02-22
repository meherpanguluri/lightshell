package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config handles the `lightshell config` command.
func Config(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: lightshell config <get|set> <key> [value]\n\nKeys:\n  releaseServer    URL of the release server\n  releaseToken     Auth token for the release server")
	}

	switch args[0] {
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: lightshell config get <key>")
		}
		return configGet(args[1])
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: lightshell config set <key> <value>")
		}
		return configSet(args[1], args[2])
	default:
		return fmt.Errorf("unknown config subcommand: %s\n\nUsage: lightshell config <get|set> <key> [value]", args[0])
	}
}

// Valid config keys
var validConfigKeys = map[string]bool{
	"releaseServer": true,
	"releaseToken":  true,
}

func configGet(key string) error {
	if !validConfigKeys[key] {
		return fmt.Errorf("unknown config key: %q\n\nValid keys: releaseServer, releaseToken", key)
	}

	value := loadConfigValue(key)
	if value == "" {
		fmt.Printf("%s: (not set)\n", key)
	} else {
		fmt.Printf("%s: %s\n", key, value)
	}
	return nil
}

func configSet(key, value string) error {
	if !validConfigKeys[key] {
		return fmt.Errorf("unknown config key: %q\n\nValid keys: releaseServer, releaseToken", key)
	}

	cfg, err := loadGlobalConfig()
	if err != nil {
		cfg = make(map[string]string)
	}

	cfg[key] = value

	if err := saveGlobalConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s\n", key)
	return nil
}

func globalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".lightshell", "config.json")
}

func loadGlobalConfig() (map[string]string, error) {
	path := globalConfigPath()
	if path == "" {
		return nil, fmt.Errorf("could not determine home directory")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	var cfg map[string]string
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}
	return cfg, nil
}

func saveGlobalConfig(cfg map[string]string) error {
	path := globalConfigPath()
	if path == "" {
		return fmt.Errorf("could not determine home directory")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// loadConfigValue reads a single value from the global config.
func loadConfigValue(key string) string {
	cfg, err := loadGlobalConfig()
	if err != nil {
		return ""
	}
	return cfg[key]
}
