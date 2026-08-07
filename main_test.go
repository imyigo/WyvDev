package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeRepoName(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"github-mcp-server", false},
		{"my.repo-1.2.3", false},
		{"ECC", false},
		{"", true},
		{".", true},
		{"..", true},
		{"../etc", true},
		{"a/b", true},
		{`a\b`, true},
		{"./foo", true},
	}
	for _, c := range cases {
		err := safeRepoName(c.name)
		if c.wantErr && err == nil {
			t.Errorf("safeRepoName(%q): expected error, got nil", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("safeRepoName(%q): unexpected error: %v", c.name, err)
		}
	}
}

func TestIsLocalOrigin(t *testing.T) {
	cases := []struct {
		origin string
		want   bool
	}{
		{"http://127.0.0.1:47651", true},
		{"http://localhost:47651", true},
		{"http://127.0.0.1", true},
		{"https://evil.example", false},
		{"http://127.0.0.1.evil.example", false},
		{"http://127.0.0.1:9999", false}, // wrong port
		{"", false},
		{"not-a-url", false},
	}
	for _, c := range cases {
		got := isLocalOrigin(c.origin)
		if got != c.want {
			t.Errorf("isLocalOrigin(%q) = %v, want %v", c.origin, got, c.want)
		}
	}
}

func TestFindSkillMd(t *testing.T) {
	writeFile := func(t *testing.T, dir, rel, content string) {
		t.Helper()
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("root SKILL.md is found", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "SKILL.md", "description: test")
		if got := findSkillMd(dir); got == "" {
			t.Error("expected root SKILL.md to be found")
		}
	})

	t.Run("skills/*/SKILL.md is found", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "skills/foo/SKILL.md", "description: test")
		if got := findSkillMd(dir); got == "" {
			t.Error("expected skills/foo/SKILL.md to be found")
		}
	})

	// Regression: a repo's own internal Claude Code tooling (.claude/skills)
	// is not a distributable skill — this must NOT count, otherwise any
	// project that merely uses Claude Code for its own development (e.g.
	// the actual Dokploy PaaS repo) gets mislabeled as a skill package.
	t.Run(".claude/skills is NOT found", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, ".claude/skills/foo/SKILL.md", "description: test")
		if got := findSkillMd(dir); got != "" {
			t.Errorf(".claude/skills/foo/SKILL.md should not match, got %q", got)
		}
	})

	// Regression: no full-tree walk — a stray SKILL.md buried deep in an
	// unrelated subdirectory (docs example, vendored code) must not
	// mislabel the whole repo.
	t.Run("deeply nested unrelated SKILL.md is NOT found", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "vendor/some/example/docs/SKILL.md", "description: test")
		if got := findSkillMd(dir); got != "" {
			t.Errorf("deeply nested unrelated SKILL.md should not match, got %q", got)
		}
	})

	t.Run("no SKILL.md anywhere relevant", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "README.md", "hello")
		if got := findSkillMd(dir); got != "" {
			t.Errorf("expected no match, got %q", got)
		}
	})
}

func TestDetectEnvVarNames(t *testing.T) {
	dir := t.TempDir()
	jsFile := filepath.Join(dir, "index.js")
	if err := os.WriteFile(jsFile, []byte(`
		const url = process.env.FOO_BAR_URL;
		const nodeEnv = process.env.NODE_ENV; // noise, should be filtered
	`), 0644); err != nil {
		t.Fatal(err)
	}
	pyFile := filepath.Join(dir, "main.py")
	if err := os.WriteFile(pyFile, []byte(`
		import os
		token = os.getenv("BAZ_QUX_TOKEN")
	`), 0644); err != nil {
		t.Fatal(err)
	}

	got := detectEnvVarNames(dir)
	want := map[string]bool{"FOO_BAR_URL": true, "BAZ_QUX_TOKEN": true}
	found := map[string]bool{}
	for _, v := range got {
		found[v] = true
		if v == "NODE_ENV" {
			t.Error("NODE_ENV is noise and should have been filtered out")
		}
	}
	for w := range want {
		if !found[w] {
			t.Errorf("expected %q to be detected, got %v", w, got)
		}
	}
}

func TestSyncToIde_VSCodeFormat(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "mcp.json")
	ide := DetectedIde{ID: "vscode-workspace", Path: target}
	template := &McpConfigFile{McpServers: map[string]McpServerConfig{
		"my-server": {Command: "node", Args: []string{"server.js"}},
	}}

	if _, err := syncToIde(ide, template); err != nil {
		t.Fatalf("syncToIde: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	servers, ok := root["servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected top-level \"servers\" key, got: %s", data)
	}
	entry, ok := servers["my-server"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected servers.my-server, got: %s", data)
	}
	if entry["type"] != "stdio" {
		t.Errorf("expected type=stdio, got %v", entry["type"])
	}
}

func TestSyncToIde_ZedFormat(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "settings.json")
	ide := DetectedIde{ID: "zed", Path: target}
	template := &McpConfigFile{McpServers: map[string]McpServerConfig{
		"my-server": {Command: "node", Args: []string{"server.js"}},
	}}

	if _, err := syncToIde(ide, template); err != nil {
		t.Fatalf("syncToIde: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	if _, ok := root["context_servers"].(map[string]interface{}); !ok {
		t.Fatalf("expected top-level \"context_servers\" key for zed, got: %s", data)
	}
}
