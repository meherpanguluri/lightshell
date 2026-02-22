package main

import (
	"fmt"
	"os"

	"github.com/lightshell-dev/lightshell/internal/cli"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("lightshell %s\n", version)
	case "init":
		name := ""
		if len(os.Args) > 2 {
			name = os.Args[2]
		}
		if err := cli.Init(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "dev":
		if err := cli.Dev(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "build":
		if err := cli.Build(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "doctor":
		if err := cli.Doctor(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "keys":
		if err := cli.Keys(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "release":
		if err := cli.Release(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "config":
		if err := cli.Config(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "mcp":
		if err := cli.MCP(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`LightShell — Build desktop apps with JavaScript

Usage:
  lightshell <command> [options]

Commands:
  init [name]    Create a new LightShell project
  dev            Run app with hot reload (dev mode)
  build          Build app for current platform
  doctor         Check for cross-platform compatibility issues
  keys           Manage signing keys (keys generate)
  release        Sign and publish a release
  config         Get/set global config (config get/set <key> [value])
  mcp            Run MCP server for AI-assisted development
  version        Print version

Run 'lightshell help' for more information.`)
}
