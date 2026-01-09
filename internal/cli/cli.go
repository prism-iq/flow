package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"flow/internal/codegen"
	"flow/internal/compiler"
	"flow/internal/config"
	"flow/internal/parser"
)

const version = "0.4.0"

func Execute() error {
	if len(os.Args) < 2 {
		return printUsage()
	}

	cmd := os.Args[1]

	switch cmd {
	case "run":
		return runCmd()
	case "build":
		return buildCmd()
	case "show":
		return showCmd()
	case "version", "-v", "--version":
		fmt.Printf("flow %s\n", version)
		return nil
	case "help", "-h", "--help":
		return printUsage()
	default:
		// If it's a .flow file, run it
		if strings.HasSuffix(cmd, ".flow") {
			os.Args = append([]string{os.Args[0], "run"}, os.Args[1:]...)
			return runCmd()
		}
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func printUsage() error {
	fmt.Println(`Flow - Human syntax, native performance.

Usage:
  flow <command> [arguments]

Commands:
  run <file.flow>     Parse, compile, and run
  build <file.flow>   Parse and compile (creates binary)
  show <file.flow>    Show generated C++ code

Options:
  --keep              Keep intermediate files (.cpp)
  --debug             Show debug output

Examples:
  flow run hello.flow
  flow build hello.flow --keep
  flow show hello.flow

Environment:
  FLOW_COMPILER       C++ compiler (default: g++)
  FLOW_STD            C++ standard (default: c++17)
  FLOW_DEBUG          Enable debug output

Version: ` + version + ` (fully local, no API)`)
	return nil
}

func runCmd() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: flow run <file.flow>")
	}

	filename := os.Args[2]
	opts := parseOptions(os.Args[3:])

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	if cfg.Debug || opts.debug {
		fmt.Printf("[DEBUG] Running %s\n", filename)
	}

	// Read Flow source
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	// Parse Flow
	ast, err := parser.Parse(string(source))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Generate C++
	cppCode, err := codegen.GenerateCode(ast)
	if err != nil {
		return fmt.Errorf("codegen error: %w", err)
	}

	if cfg.Debug || opts.debug {
		fmt.Println("[DEBUG] Generated C++:")
		fmt.Println(cppCode)
		fmt.Println("[DEBUG] ---")
	}

	// Compile and run
	c := compiler.New(cfg)
	output, err := c.CompileAndRun(cppCode, filename)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

func buildCmd() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: flow build <file.flow>")
	}

	filename := os.Args[2]
	opts := parseOptions(os.Args[3:])

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Read Flow source
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	// Parse Flow
	ast, err := parser.Parse(string(source))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Generate C++
	cppCode, err := codegen.GenerateCode(ast)
	if err != nil {
		return fmt.Errorf("codegen error: %w", err)
	}

	// Compile to binary
	c := compiler.New(cfg)
	binaryPath, err := c.Compile(cppCode, filename, opts.keep)
	if err != nil {
		return err
	}

	fmt.Printf("Built: %s\n", binaryPath)
	return nil
}

func showCmd() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: flow show <file.flow>")
	}

	filename := os.Args[2]

	// Read Flow source
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	// Parse Flow
	ast, err := parser.Parse(string(source))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Generate C++
	cppCode, err := codegen.GenerateCode(ast)
	if err != nil {
		return fmt.Errorf("codegen error: %w", err)
	}

	// Get base name without extension
	base := strings.TrimSuffix(filepath.Base(filename), ".flow")

	fmt.Printf("// Generated from %s\n", filename)
	fmt.Printf("// Save as: %s.cpp\n\n", base)
	fmt.Println(cppCode)

	return nil
}

type options struct {
	keep  bool
	debug bool
}

func parseOptions(args []string) options {
	opts := options{}
	for _, arg := range args {
		switch arg {
		case "--keep":
			opts.keep = true
		case "--debug":
			opts.debug = true
		}
	}
	return opts
}
