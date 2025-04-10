package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateFormulaFromYAML(t *testing.T) {
	outputDir := "Formula"
	inputFile := "../testdata/kiosk.yaml"
	expectedOutput := filepath.Join(outputDir, "kiosk.rb")
	templatePath := "../templates/formula.rb.tmpl"

	// Verify the input file exists
	if _, err := os.Stat(inputFile); err != nil {
		t.Fatalf("input file not found: %s", inputFile)
	}

	// Verify the template file exists
	if _, err := os.Stat(templatePath); err != nil {
		t.Fatalf("template file not found: %s", templatePath)
	}

	// Clean old output
	_ = os.RemoveAll(outputDir)
	_ = os.MkdirAll(outputDir, 0755)

	// Ensure cleanup after the test
	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	err := processConfig(inputFile, templatePath)
	if err != nil {
		t.Fatalf("processConfig failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(expectedOutput); err != nil {
		t.Fatalf("expected output file not found: %s", expectedOutput)
	}

	// Check Ruby syntax
	cmd := exec.Command("ruby", "-c", expectedOutput)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("ruby syntax check failed: %s", string(output))
	} else {
		t.Logf("ruby syntax check succeeded: %s", string(output))
	}

	// Validate content includes expected fields
	content, err := os.ReadFile(expectedOutput)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	data := string(content)
	if !strings.Contains(data, "url") {
		t.Error("formula missing 'url'")
	}
	if !strings.Contains(data, "sha256") {
		t.Error("formula missing 'sha256'")
	}
}
