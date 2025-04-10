package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Platform struct {
	OS     string `yaml:"os"`
	Arch   string `yaml:"arch"`
	URL    string `yaml:"url"`
	SHA256 string `yaml:"sha256"`
}

type FormulaSpec struct {
	Name      string     `yaml:"name"`
	Title     string     `yaml:"title"`
	Desc      string     `yaml:"desc"`
	Homepage  string     `yaml:"homepage"`
	Version   string     `yaml:"version"`
	Platforms []Platform `yaml:"platforms"`
}

func main() {
	matches, err := filepath.Glob("config/*.yaml")
	check(err, "finding config files")

	templatePath := "templates/formula.rb.tmpl" // Adjust this path if necessary

	for _, configPath := range matches {
		err := processConfig(configPath, templatePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error processing %s: %v\n", configPath, err)
			os.Exit(1)
		}
	}
}

func processConfig(configPath string, templatePath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	var spec FormulaSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("unmarshaling yaml: %w", err)
	}

	if len(spec.Platforms) == 0 {
		return fmt.Errorf("no platforms defined")
	}

	specMap := map[string]interface{}{
		"Name":               spec.Name,
		"Title":              spec.Title,
		"Desc":               spec.Desc,
		"Homepage":           spec.Homepage,
		"Version":            spec.Version,
		"Platforms":          spec.Platforms,
		"MacBuildFromSource": true,
	}

	tmpl, err := template.New("formula").Funcs(template.FuncMap{
		"title": title,
	}).ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "formula.rb.tmpl", specMap); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	outputPath := filepath.Join("Formula", spec.Name+".rb")
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing formula: %w", err)
	}

	cmd := exec.Command("ruby", "-c", outputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return fmt.Errorf("ruby syntax check failed")
	}

	return nil
}

func title(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(bytes.ToUpper([]byte{s[0]})) + s[1:]
}

func check(err error, context string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error %s: %v\n", context, err)
		os.Exit(1)
	}
}
