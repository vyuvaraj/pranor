package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServPluginPull(t *testing.T) {
	tmpDir := t.TempDir()
	outPluginPath := filepath.Join(tmpDir, "jwt-auth.wasm")

	// Test pulling a plugin to custom path
	pullPlugin("jwt-auth", outPluginPath, "v0.1.0")

	if _, err := os.Stat(outPluginPath); os.IsNotExist(err) {
		t.Fatalf("Expected plugin file at %s, but it was not created", outPluginPath)
	}

	content, err := os.ReadFile(outPluginPath)
	if err != nil {
		t.Fatalf("Failed to read pulled plugin file: %v", err)
	}

	// Verify WASM magic bytes: \x00asm
	if len(content) < 8 {
		t.Fatalf("Plugin binary size too small: %d bytes", len(content))
	}

	if string(content[0:4]) != "\x00asm" {
		t.Fatalf("Header does not match WASM magic bytes \\x00asm: %v", content[0:4])
	}
}

func TestServPluginListAndSearch(t *testing.T) {
	if len(AvailablePlugins) < 5 {
		t.Fatalf("Expected at least 5 available plugins, found %d", len(AvailablePlugins))
	}

	foundJWT := false
	for _, p := range AvailablePlugins {
		if p.Name == "jwt-auth" {
			foundJWT = true
			if p.Component != "Pranor Gate" {
				t.Errorf("Expected jwt-auth component to be Pranor Gate, got %s", p.Component)
			}
		}
	}

	if !foundJWT {
		t.Errorf("jwt-auth plugin missing from AvailablePlugins list")
	}

	// Test search filter
	matches := 0
	query := "scrubber"
	for _, p := range AvailablePlugins {
		if strings.Contains(p.Name, query) || strings.Contains(p.Description, query) {
			matches++
		}
	}

	if matches == 0 {
		t.Errorf("Expected search match for query '%s'", query)
	}
}
