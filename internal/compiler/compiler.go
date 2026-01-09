package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"flow/internal/config"
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

	// Compile
	if err := c.compile(cppFile, binFile); err != nil {
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

	// Compile
	if err := c.compile(cppFile, binFile); err != nil {
		return "", err
	}

	return binFile, nil
}

func (c *Compiler) compile(cppFile, binFile string) error {
	cmd := exec.Command(c.cfg.Compiler,
		"-std="+c.cfg.CppStd,
		"-o", binFile,
		cppFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compilation failed:\n%s", string(output))
	}

	return nil
}
