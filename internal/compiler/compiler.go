package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"flow/internal/config"
	"flow/internal/transpiler"
)

type Compiler struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Compiler {
	return &Compiler{cfg: cfg}
}

func (c *Compiler) CompileAndRun(cppCode, flowFile string) (string, error) {
	// Create temp directory for compilation
	tmpDir, err := os.MkdirTemp("", "flow-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Get base name
	base := strings.TrimSuffix(filepath.Base(flowFile), ".flow")
	cppFile := filepath.Join(tmpDir, base+".cpp")
	binFile := filepath.Join(tmpDir, base)

	// Write C++ file
	if err := os.WriteFile(cppFile, []byte(cppCode), 0644); err != nil {
		return "", fmt.Errorf("failed to write cpp file: %w", err)
	}

	// Read original Flow source for feedback loop
	flowSource, _ := os.ReadFile(flowFile)

	// Compile with feedback loop
	if err := c.compileWithFeedback(cppFile, binFile, string(flowSource), cppCode); err != nil {
		return "", err
	}

	// Run the binary
	cmd := exec.Command(binFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execution failed: %w\n%s", err, string(output))
	}

	return string(output), nil
}

func (c *Compiler) Compile(cppCode, flowFile string, keepCpp bool) (string, error) {
	// Get base name and paths
	dir := filepath.Dir(flowFile)
	base := strings.TrimSuffix(filepath.Base(flowFile), ".flow")
	cppFile := filepath.Join(dir, base+".cpp")
	binFile := filepath.Join(dir, base)

	// Write C++ file
	if err := os.WriteFile(cppFile, []byte(cppCode), 0644); err != nil {
		return "", fmt.Errorf("failed to write cpp file: %w", err)
	}

	// Clean up cpp file if not keeping
	if !keepCpp {
		defer os.Remove(cppFile)
	}

	// Read original Flow source for feedback loop
	flowSource, _ := os.ReadFile(flowFile)

	// Compile with feedback loop
	if err := c.compileWithFeedback(cppFile, binFile, string(flowSource), cppCode); err != nil {
		return "", err
	}

	return binFile, nil
}

func (c *Compiler) compileWithFeedback(cppFile, binFile, flowSource, cppCode string) error {
	var lastError string
	currentCpp := cppCode

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("[FLOW] Fixing error (attempt %d/%d)...\n", attempt, c.cfg.MaxRetries)

			// Use transpiler to fix the error
			t := transpiler.New(c.cfg)
			fixedCpp, err := t.TranspileWithFeedback(flowSource, currentCpp, lastError)
			if err != nil {
				return fmt.Errorf("failed to fix error: %w", err)
			}

			currentCpp = fixedCpp

			// Write the fixed C++ code
			if err := os.WriteFile(cppFile, []byte(currentCpp), 0644); err != nil {
				return fmt.Errorf("failed to write fixed cpp file: %w", err)
			}

			if c.cfg.Debug {
				fmt.Println("[DEBUG] Fixed C++:")
				fmt.Println(currentCpp)
				fmt.Println("[DEBUG] ---")
			}
		}

		// Try to compile
		cmd := exec.Command(c.cfg.Compiler,
			"-std="+c.cfg.CppStd,
			"-o", binFile,
			cppFile,
		)

		output, err := cmd.CombinedOutput()
		if err == nil {
			// Success!
			if attempt > 0 {
				fmt.Printf("[FLOW] Fixed after %d attempt(s)\n", attempt)
			}
			return nil
		}

		lastError = string(output)

		if c.cfg.Debug {
			fmt.Printf("[DEBUG] Compile error:\n%s\n", lastError)
		}
	}

	return fmt.Errorf("compilation failed after %d attempts:\n%s", c.cfg.MaxRetries, lastError)
}
