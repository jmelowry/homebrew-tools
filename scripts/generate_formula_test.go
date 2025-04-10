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
	expectedOutput := filepath.Join("Formula", "kiosk.rb")

	// Clean old output
	_ = os.RemoveAll(outputDir)
	_ = os.MkdirAll(outputDir, 0755)

	err := processConfig(inputFile)
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
