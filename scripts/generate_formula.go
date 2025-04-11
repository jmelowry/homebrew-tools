package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type Release struct {
	TagName string `json:"tag_name"`
}

func fetchLatestVersion(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func calculateSHA256(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("downloading asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download asset, status: %s", resp.Status)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, resp.Body); err != nil {
		return "", fmt.Errorf("calculating sha256: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
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

	// Fetch the latest version if "latest" is specified
	if spec.Version == "latest" {
		latestVersion, err := fetchLatestVersion("jmelowry/kiosk")
		if err != nil {
			return fmt.Errorf("fetching latest version: %w", err)
		}
		spec.Version = latestVersion
	}

	// Fetch URLs and SHA256 for each platform
	for i, platform := range spec.Platforms {
		assetName := fmt.Sprintf("kiosk-%s-%s-%s.tar.gz", spec.Version, platform.OS, platform.Arch)
		assetURL := fmt.Sprintf("https://github.com/jmelowry/kiosk/releases/download/%s/%s", spec.Version, assetName)

		// Download the asset to calculate its SHA256
		sha256, err := calculateSHA256(assetURL)
		if err != nil {
			return fmt.Errorf("calculating sha256 for %s: %w", assetURL, err)
		}

		spec.Platforms[i].URL = assetURL
		spec.Platforms[i].SHA256 = sha256
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

	// Ensure the Formula directory exists
	outputDir := "Formula"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, spec.Name+".rb")
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
