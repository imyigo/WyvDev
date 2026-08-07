package main

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ---------- Data model ----------

type TaskEntry struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`    // "clone", "pull", "install", "delete", "run"
	Name      string    `json:"name"`    // repo or folder name
	Status    string    `json:"status"`  // "running", "completed", "error", "cancelled"
	Message   string    `json:"message"` // detail or output snippet
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt,omitempty"`

	cancel context.CancelFunc // not serialized; lets /stop interrupt the in-flight command
}

var (
	tasksMutex sync.Mutex
	tasksMap   = make(map[string]*TaskEntry)
	taskOrder  = []string{}
)

func addTask(taskType, name, initialMsg string) *TaskEntry {
	tasksMutex.Lock()
	defer tasksMutex.Unlock()

	id := fmt.Sprintf("%s-%d", taskType, time.Now().UnixNano())
	task := &TaskEntry{
		ID:        id,
		Type:      taskType,
		Name:      name,
		Status:    "running",
		Message:   initialMsg,
		StartedAt: time.Now(),
	}
	tasksMap[id] = task
	taskOrder = append([]string{id}, taskOrder...)
	if len(taskOrder) > 30 {
		delete(tasksMap, taskOrder[len(taskOrder)-1])
		taskOrder = taskOrder[:30]
	}
	return task
}

func finishTask(task *TaskEntry, status, msg string) {
	if task == nil {
		return
	}
	tasksMutex.Lock()
	defer tasksMutex.Unlock()

	task.Status = status
	task.Message = msg
	task.EndedAt = time.Now()
	task.cancel = nil
}

// attachCancel lets a running task be interrupted later via POST /api/tasks/{id}/stop.
func (t *TaskEntry) attachCancel(cancel context.CancelFunc) {
	tasksMutex.Lock()
	t.cancel = cancel
	tasksMutex.Unlock()
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	tasksMutex.Lock()
	defer tasksMutex.Unlock()

	list := make([]*TaskEntry, 0, len(taskOrder))
	for _, id := range taskOrder {
		if t, ok := tasksMap[id]; ok {
			list = append(list, t)
		}
	}
	writeJSON(w, 200, list)
}

// handleTaskStop cancels a running task's underlying command (git clone, install,
// repo start, check-all) without removing it from the list — it settles into
// "cancelled" once the interrupted command actually returns.
func handleTaskStop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	tasksMutex.Lock()
	task, ok := tasksMap[id]
	var cancel context.CancelFunc
	var status, name, typ string
	if ok {
		cancel = task.cancel
		status = task.Status
		name = task.Name
		typ = task.Type
	}
	tasksMutex.Unlock()

	if !ok {
		writeErr(w, 404, "gorev bulunamadi")
		return
	}
	if status != "running" {
		writeJSON(w, 200, map[string]interface{}{"ok": true, "message": "Görev zaten sona ermiş."})
		return
	}
	if cancel == nil {
		writeErr(w, 400, "bu gorev turu durdurulamiyor")
		return
	}

	cancel()
	logActivity("task-stop", fmt.Sprintf("%s (%s) durduruldu", name, typ))
	writeJSON(w, 200, map[string]interface{}{
		"ok":      true,
		"message": fmt.Sprintf("🛑 '%s' görevi durduruluyor...", name),
	})
}

// handleTaskDelete removes a task from the visible list (running tasks are
// stopped first, same as handleTaskStop, then dropped from history).
func handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	tasksMutex.Lock()
	task, ok := tasksMap[id]
	var cancel context.CancelFunc
	if ok {
		cancel = task.cancel
		delete(tasksMap, id)
		for i, oid := range taskOrder {
			if oid == id {
				taskOrder = append(taskOrder[:i], taskOrder[i+1:]...)
				break
			}
		}
	}
	tasksMutex.Unlock()

	if !ok {
		writeErr(w, 404, "gorev bulunamadi")
		return
	}
	if cancel != nil {
		cancel()
	}
	writeJSON(w, 200, map[string]interface{}{"ok": true})
}

type McpServerConfig struct {
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type McpConfigFile struct {
	McpServers map[string]McpServerConfig `json:"mcpServers"`
}

type DetectedIde struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	SkillsDir string `json:"skillsDir,omitempty"`
	Detected  bool   `json:"detected"`
}

type MCPEntry struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	URL       string            `json:"url,omitempty"`
	Desc      string            `json:"desc,omitempty"`
	Category  string            `json:"category,omitempty"`
	Badge     string            `json:"badge,omitempty"`
	Icon      string            `json:"icon,omitempty"`
	IconColor string            `json:"iconColor,omitempty"`
	Auth      bool              `json:"auth,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Repo      string            `json:"repo,omitempty"`
	Cwd       string            `json:"cwd,omitempty"`
}

type SkillEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Repo     string `json:"repo,omitempty"`
	Category string `json:"category,omitempty"`
	Desc     string `json:"desc,omitempty"`
	Extra    string `json:"extra,omitempty"`
}

type IdePathEntry struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Status string `json:"status,omitempty"`
}

type TrackedRepo struct {
	Name        string `json:"name"`
	Repo        string `json:"repo"`
	LocalPath   string `json:"localPath"`
	Status      string `json:"status,omitempty"`
	LastChecked string `json:"lastChecked,omitempty"`
}

type ScanEntry struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	HasSkill       bool     `json:"hasSkill"`
	SkillDesc      string   `json:"skillDesc,omitempty"`
	HasPackageJson bool     `json:"hasPackageJson"`
	PackageName    string   `json:"packageName,omitempty"`
	LooksLikeMcp   bool     `json:"looksLikeMcp"`
	Repo           string   `json:"repo,omitempty"`
	Runtimes       []string `json:"runtimes,omitempty"` // "node", "python", "rust", "docker"
	GitStatus      string   `json:"gitStatus,omitempty"`
	IsInstalled    bool     `json:"isInstalled"`
	StartCommand   string   `json:"startCommand,omitempty"`
	AvailableCmds  []string `json:"availableCmds,omitempty"`
	HasStartError  bool     `json:"hasStartError,omitempty"`
	StartErrorMsg  string   `json:"startErrorMsg,omitempty"`
	IsRunning      bool     `json:"isRunning,omitempty"`
	RunningPort    string   `json:"runningPort,omitempty"`
	RunningPid     int      `json:"runningPid,omitempty"`
	RepoType       string   `json:"repoType"`
	RepoTypeLabel  string   `json:"repoTypeLabel"`
	RunMode        string   `json:"runMode,omitempty"`      // "local" or "docker" — the one primary way to run this repo
	MissingTool    string   `json:"missingTool,omitempty"`  // system tool RunMode needs but isn't installed (e.g. "docker")
	LocalCommand   string   `json:"localCommand,omitempty"` // absolute-path command for a real (not npx-guessed) MCP stdio config
	LocalArgs      []string `json:"localArgs,omitempty"`
	StartCommandGuessed bool `json:"startCommandGuessed,omitempty"` // true when StartCommand fell through to "pick any script" rather than matching a known launch-script name
}

type RunningApp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Command   string `json:"command"`
	PID       int    `json:"pid"`
	Port      string `json:"port,omitempty"`
	Status    string `json:"status"` // "running"
	StartedAt string `json:"startedAt"`
}

var (
	repoStartErrors    = map[string]string{}
	repoStartErrorsMux sync.RWMutex

	runningAppsMap = map[string]*RunningApp{}
	runningAppsMux sync.RWMutex
)

type LoopEngineConfig struct {
	AutoHealEnabled bool   `json:"autoHealEnabled"`
	AutoKillPort    bool   `json:"autoKillPort"`
	MaxRetries      int    `json:"maxRetries"`
	CacheStrategy   string `json:"cacheStrategy"`
}

type StateBundle struct {
	MasterIDE        string           `json:"masterIde,omitempty"`
	McpServers       []MCPEntry       `json:"mcpServers"`
	RecommendedRepos []SkillEntry     `json:"recommendedRepos"`
	IdePaths         []IdePathEntry   `json:"idePaths"`
	TrackedRepos     []TrackedRepo    `json:"trackedRepos"`
	DeletedIdeIDs    []string         `json:"deletedIdeIds,omitempty"`
	LoopConfig       LoopEngineConfig `json:"loopConfig,omitempty"`
}

// ---------- Path helpers (AI_TOOLKIT_TEST_HOME lets tests point at a scratch dir) ----------

func getHomeDir() string {
	if testHome := os.Getenv("AI_TOOLKIT_TEST_HOME"); testHome != "" {
		return testHome
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "C:\\Users\\Admin"
	}
	return home
}

func getAppDataDir() string {
	if testHome := os.Getenv("AI_TOOLKIT_TEST_HOME"); testHome != "" {
		return filepath.Join(testHome, "AppData", "Roaming")
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return filepath.Join(getHomeDir(), "AppData", "Roaming")
	}
	return appData
}

func getBaseDir() string {
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, "index.html")); err == nil {
			return wd
		}
	}
	if execPath, err := os.Executable(); err == nil {
		dir := filepath.Dir(execPath)
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			return dir
		}
	}
	return "."
}

// getAppSupportDir returns the per-OS root that Electron-based apps (VS Code,
// Cursor, Windsurf, Claude Desktop, JetBrains) use for user config: %APPDATA%
// on Windows, ~/Library/Application Support on macOS, ~/.config on Linux.
// detectIdes() used to assume the Windows shape unconditionally, so on
// macOS/Linux it pointed at a folder those apps never read from or write to
// — MCP sync looked like it worked (200 OK, "N sunucu yazıldı" logged) but
// silently wrote into a stray folder next to nothing.
func getAppSupportDir() string {
	if testHome := os.Getenv("AI_TOOLKIT_TEST_HOME"); testHome != "" {
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(testHome, "AppData", "Roaming")
		case "darwin":
			return filepath.Join(testHome, "Library", "Application Support")
		default:
			return filepath.Join(testHome, ".config")
		}
	}
	switch runtime.GOOS {
	case "windows":
		return getAppDataDir()
	case "darwin":
		return filepath.Join(getHomeDir(), "Library", "Application Support")
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return xdg
		}
		return filepath.Join(getHomeDir(), ".config")
	}
}

func detectIdes() []DetectedIde {
	home := getHomeDir()
	appData := getAppSupportDir()
	baseDir := getBaseDir()

	candidates := []DetectedIde{
		{ID: "antigravity-app", Name: "Antigravity IDE & App", Path: filepath.Join(home, ".gemini", "antigravity", "mcp_config.json"), SkillsDir: filepath.Join(home, ".gemini", "antigravity", "builtin", "skills")},
		{ID: "antigravity-global", Name: "Antigravity Global Config", Path: filepath.Join(home, ".gemini", "config", "mcp_config.json"), SkillsDir: filepath.Join(home, ".gemini", "config", "skills")},
		{ID: "claude-desktop", Name: "Claude Desktop App", Path: filepath.Join(appData, "Claude", "claude_desktop_config.json"), SkillsDir: filepath.Join(appData, "Claude", "skills")},
		{ID: "cursor", Name: "Cursor IDE", Path: filepath.Join(home, ".cursor", "mcp.json"), SkillsDir: filepath.Join(home, ".cursor", "skills")},
		{ID: "cursor-cline", Name: "Cursor (Cline Extension)", Path: filepath.Join(appData, "Cursor", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json"), SkillsDir: filepath.Join(appData, "Cursor", "User", "globalStorage", "saoudrizwan.claude-dev", "skills")},
		{ID: "claude-code", Name: "Claude Code CLI", Path: filepath.Join(home, ".claude.json"), SkillsDir: filepath.Join(home, ".claude", "skills")},
		{ID: "vscode-cline", Name: "VS Code (Cline Extension)", Path: filepath.Join(appData, "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json"), SkillsDir: filepath.Join(appData, "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "skills")},
		{ID: "vscode-roo", Name: "VS Code (Roo Code Extension)", Path: filepath.Join(appData, "Code", "User", "globalStorage", "rooveterinaryinc.roo-cline", "settings", "cline_mcp_settings.json"), SkillsDir: filepath.Join(appData, "Code", "User", "globalStorage", "rooveterinaryinc.roo-cline", "skills")},
		{ID: "vscode-workspace", Name: "VS Code Workspace", Path: filepath.Join(baseDir, ".vscode", "mcp.json"), SkillsDir: filepath.Join(baseDir, ".vscode", "skills")},
		{ID: "windsurf", Name: "Windsurf IDE", Path: filepath.Join(appData, "Windsurf", "User", "globalStorage", "mcp_config.json"), SkillsDir: filepath.Join(appData, "Windsurf", "User", "globalStorage", "skills")},
		{ID: "zed", Name: "Zed Editor", Path: filepath.Join(home, ".config", "zed", "settings.json"), SkillsDir: filepath.Join(home, ".config", "zed", "skills")},
		{ID: "continue", Name: "Continue.dev", Path: filepath.Join(home, ".continue", "config.json"), SkillsDir: filepath.Join(home, ".continue", "skills")},
		{ID: "jetbrains", Name: "JetBrains IDEs", Path: filepath.Join(appData, "JetBrains", "mcp.json"), SkillsDir: filepath.Join(appData, "JetBrains", "skills")},
		{ID: "universal-agents", Name: "Universal Agent Skills", Path: filepath.Join(home, ".agents", "mcp_config.json"), SkillsDir: filepath.Join(home, ".agents", "skills")},
	}

	for i := range candidates {
		dir := filepath.Dir(candidates[i].Path)
		skillsParent := filepath.Dir(candidates[i].SkillsDir)
		if _, err := os.Stat(dir); err == nil {
			candidates[i].Detected = true
		} else if _, err := os.Stat(candidates[i].Path); err == nil {
			candidates[i].Detected = true
		} else if _, err := os.Stat(skillsParent); err == nil {
			candidates[i].Detected = true
		} else if _, err := os.Stat(candidates[i].SkillsDir); err == nil {
			candidates[i].Detected = true
		}
	}

	return candidates
}

// ---------- MCP config sync (used by both --once CLI and the live server) ----------

func loadCoreTemplate() (*McpConfigFile, error) {
	s, err := loadState()
	if err == nil && len(s.McpServers) > 0 {
		return buildMcpConfigFile(s.McpServers), nil
	}

	baseDir := getBaseDir()
	templatePath := filepath.Join(baseDir, "core", "mcp-config.json")
	data, err := os.ReadFile(templatePath)
	if err == nil {
		var config McpConfigFile
		if err := json.Unmarshal(data, &config); err == nil {
			return &config, nil
		}
	}

	return &McpConfigFile{McpServers: make(map[string]McpServerConfig)}, nil
}

func buildMcpConfigFile(entries []MCPEntry) *McpConfigFile {
	cfg := &McpConfigFile{McpServers: make(map[string]McpServerConfig)}
	for _, e := range entries {
		sc := McpServerConfig{}
		if e.Type == "http" {
			sc.Type = "http"
			sc.URL = e.URL
			if len(e.Headers) > 0 {
				sc.Headers = e.Headers
			}
		} else {
			sc.Command = e.Command
			sc.Args = e.Args
			sc.Cwd = e.Cwd
			if len(e.Env) > 0 {
				sc.Env = e.Env
			}
		}
		cfg.McpServers[e.ID] = sc
	}
	return cfg
}

func syncToIde(ide DetectedIde, template *McpConfigFile) (int, error) {
	targetPath := ide.Path
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("klasor olusturulamadi: %v", err)
	}

	if _, err := os.Stat(targetPath); err == nil {
		backupPath := fmt.Sprintf("%s.bak-%s", targetPath, time.Now().Format("20060102-150405"))
		_ = copyFile(targetPath, backupPath)
	}

	var root map[string]interface{}
	if data, err := os.ReadFile(targetPath); err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = make(map[string]interface{})
	}

	var mcpMap map[string]interface{}
	if existingMcp, ok := root["mcpServers"].(map[string]interface{}); ok && existingMcp != nil {
		mcpMap = existingMcp
	} else {
		mcpMap = make(map[string]interface{})
	}

	addedCount := 0
	for name, server := range template.McpServers {
		mcpMap[name] = server
		addedCount++
	}
	root["mcpServers"] = mcpMap

	if ide.ID == "zed" || strings.Contains(strings.ToLower(targetPath), "zed") {
		var zedContext map[string]interface{}
		if existingZed, ok := root["context_servers"].(map[string]interface{}); ok && existingZed != nil {
			zedContext = existingZed
		} else {
			zedContext = make(map[string]interface{})
		}
		for name, server := range template.McpServers {
			if server.Type == "http" {
				zedContext[name] = map[string]interface{}{
					"url": server.URL,
				}
			} else {
				zedContext[name] = map[string]interface{}{
					"command": map[string]interface{}{
						"path": server.Command,
						"args": server.Args,
					},
					"env": server.Env,
				}
			}
		}
		root["context_servers"] = zedContext
	}

	if ide.ID == "continue" || strings.Contains(strings.ToLower(targetPath), "continue") {
		var expMap map[string]interface{}
		if existingExp, ok := root["experimental"].(map[string]interface{}); ok && existingExp != nil {
			expMap = existingExp
		} else {
			expMap = make(map[string]interface{})
		}
		var contServers []map[string]interface{}
		for name, server := range template.McpServers {
			if server.Type == "http" {
				contServers = append(contServers, map[string]interface{}{
					"name": name,
					"url":  server.URL,
				})
			} else {
				contServers = append(contServers, map[string]interface{}{
					"name": name,
					"transport": map[string]interface{}{
						"type":    "stdio",
						"command": server.Command,
						"args":    server.Args,
						"env":     server.Env,
					},
				})
			}
		}
		expMap["modelContextProtocolServers"] = contServers
		root["experimental"] = expMap
	}

	// VS Code's native MCP support (.vscode/mcp.json) does not read
	// "mcpServers" at all — it expects a top-level "servers" map with an
	// explicit "type": "stdio"|"http" per entry.
	if ide.ID == "vscode-workspace" {
		var vsServers map[string]interface{}
		if existingVs, ok := root["servers"].(map[string]interface{}); ok && existingVs != nil {
			vsServers = existingVs
		} else {
			vsServers = make(map[string]interface{})
		}
		for name, server := range template.McpServers {
			if server.Type == "http" {
				vsServers[name] = map[string]interface{}{
					"type": "http",
					"url":  server.URL,
				}
			} else {
				vsServers[name] = map[string]interface{}{
					"type":    "stdio",
					"command": server.Command,
					"args":    server.Args,
					"env":     server.Env,
				}
			}
		}
		root["servers"] = vsServers
	}

	updatedData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("JSON olusturma hatasi: %v", err)
	}

	if err := os.WriteFile(targetPath, updatedData, 0644); err != nil {
		return 0, fmt.Errorf("dosya yazma hatasi: %v", err)
	}

	return addedCount, nil
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func removeMcpIdsFromIde(ide DetectedIde, ids []string) {
	if len(ids) == 0 {
		return
	}
	targetPath := ide.Path
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil || root == nil {
		return
	}

	changed := false
	if mcpMap, ok := root["mcpServers"].(map[string]interface{}); ok && mcpMap != nil {
		for _, id := range ids {
			if _, ok := mcpMap[id]; ok {
				delete(mcpMap, id)
				changed = true
			}
		}
		root["mcpServers"] = mcpMap
	}

	if zedContext, ok := root["context_servers"].(map[string]interface{}); ok && zedContext != nil {
		for _, id := range ids {
			if _, ok := zedContext[id]; ok {
				delete(zedContext, id)
				changed = true
			}
		}
		root["context_servers"] = zedContext
	}

	if expMap, ok := root["experimental"].(map[string]interface{}); ok && expMap != nil {
		if contServers, ok := expMap["modelContextProtocolServers"].([]interface{}); ok {
			var newContServers []interface{}
			for _, item := range contServers {
				if mItem, ok := item.(map[string]interface{}); ok {
					name, _ := mItem["name"].(string)
					if !containsString(ids, name) {
						newContServers = append(newContServers, mItem)
					} else {
						changed = true
					}
				}
			}
			expMap["modelContextProtocolServers"] = newContServers
			root["experimental"] = expMap
		}
	}

	if vsServers, ok := root["servers"].(map[string]interface{}); ok && vsServers != nil {
		for _, id := range ids {
			if _, ok := vsServers[id]; ok {
				delete(vsServers, id)
				changed = true
			}
		}
		root["servers"] = vsServers
	}

	if !changed {
		return
	}

	updatedData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(targetPath, updatedData, 0644)
}

func removeMcpIdsFromAllIdes(ids []string) {
	if len(ids) == 0 {
		return
	}
	for _, ide := range detectIdes() {
		if ide.Detected {
			removeMcpIdsFromIde(ide, ids)
		}
	}
}

func syncAllIdes(entries []MCPEntry) []map[string]interface{} {
	template := buildMcpConfigFile(entries)
	ides := detectIdes()
	results := []map[string]interface{}{}
	for _, ide := range ides {
		if !ide.Detected {
			results = append(results, map[string]interface{}{"id": ide.ID, "name": ide.Name, "skipped": true})
			continue
		}
		count, err := syncToIde(ide, template)
		entry := map[string]interface{}{"id": ide.ID, "name": ide.Name, "path": ide.Path}
		if err != nil {
			entry["error"] = err.Error()
			logActivity("mcp-sync-error", fmt.Sprintf("%s: %v", ide.Name, err))
		} else {
			entry["written"] = count
			logActivity("mcp-sync", fmt.Sprintf("%s -> %d sunucu (%s)", ide.Name, count, ide.Path))
		}
		results = append(results, entry)
	}
	return results
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func ensureValidDir(path string) error {
	dir := path
	var parts []string
	for dir != "/" && dir != "." && dir != "" {
		parts = append([]string{dir}, parts...)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	for _, p := range parts {
		info, err := os.Lstat(p)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			if _, statErr := os.Stat(p); statErr != nil {
				_ = os.Remove(p)
			}
		}
		if err := os.MkdirAll(p, 0755); err != nil {
			return err
		}
	}
	return nil
}

// copyDirRecursive copies src into dst, skipping .git (large, irrelevant for a
// skill payload) and node_modules (regenerable, often huge).
func copyDirRecursive(src, dst string) error {
	if err := ensureValidDir(dst); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return ensureValidDir(dst)
		}
		base := filepath.Base(rel)
		if info.IsDir() && (base == ".git" || base == "node_modules") {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return ensureValidDir(target)
		}
		return copyFile(path, target)
	})
}

// ---------- state.json persistence ----------

func statePath() string {
	return filepath.Join(getBaseDir(), "core", "state.json")
}

func activityLogPath() string {
	return filepath.Join(getBaseDir(), "core", "activity.log")
}

// logActivity appends a timestamped line so every sync/install/clone/prune/delete
// leaves a permanent, inspectable trail — not just an ephemeral toast.
func logActivity(action, detail string) {
	if err := os.MkdirAll(filepath.Join(getBaseDir(), "core"), 0755); err != nil {
		return
	}
	f, err := os.OpenFile(activityLogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	line := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), action, detail)
	_, _ = f.WriteString(line)
}

func handleActivity(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	data, err := os.ReadFile(activityLogPath())
	if err != nil {
		writeJSON(w, 200, []string{})
		return
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if lines[0] == "" {
		lines = lines[:0]
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	// newest first
	reversed := make([]string, len(lines))
	for i, l := range lines {
		reversed[len(lines)-1-i] = l
	}
	writeJSON(w, 200, reversed)
}

// stateFileMux serializes every read-modify-write cycle against state.json.
// Without it, the 1s background sync ticker and an HTTP handler's own
// loadState→mutate→saveState sequence can interleave and clobber each
// other's write.
var stateFileMux sync.Mutex

func loadState() (*StateBundle, error) {
	stateFileMux.Lock()
	defer stateFileMux.Unlock()
	return loadStateLocked()
}

func loadStateLocked() (*StateBundle, error) {
	data, err := os.ReadFile(statePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &StateBundle{}, nil
		}
		return nil, err
	}
	var s StateBundle
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// reconcileState drops any mcpServers/recommendedRepos/trackedRepos entries whose
// backing repo/<name> folder was deleted outside the app (e.g. by hand in Explorer),
// so the panel never shows an item whose local copy no longer exists. Returns the
// list of repo full-names ("owner/name") that got pruned this way.
func reconcileState(s *StateBundle) []string {
	var pruned []string
	stillTracked := []TrackedRepo{}
	orphaned := map[string]bool{}

	for _, t := range s.TrackedRepos {
		if _, err := os.Stat(t.LocalPath); err != nil {
			orphaned[t.Repo] = true
			pruned = append(pruned, t.Repo)
			continue
		}
		stillTracked = append(stillTracked, t)
	}
	s.TrackedRepos = stillTracked

	if len(orphaned) == 0 {
		return pruned
	}

	filteredSkills := []SkillEntry{}
	for _, sk := range s.RecommendedRepos {
		if sk.Repo != "" && orphaned[sk.Repo] {
			continue
		}
		filteredSkills = append(filteredSkills, sk)
	}
	s.RecommendedRepos = filteredSkills

	filteredMcp := []MCPEntry{}
	for _, m := range s.McpServers {
		if m.Repo != "" && orphaned[m.Repo] {
			continue
		}
		filteredMcp = append(filteredMcp, m)
	}
	s.McpServers = filteredMcp

	for repo := range orphaned {
		logActivity("prune", fmt.Sprintf("%s (yerel repo/ klasoru silinmisti)", repo))
	}

	return pruned
}

func saveState(s *StateBundle) error {
	stateFileMux.Lock()
	defer stateFileMux.Unlock()

	if err := os.MkdirAll(filepath.Join(getBaseDir(), "core"), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Write-then-rename instead of a direct write: os.Rename is atomic on
	// both Windows and POSIX, so a reader never observes a half-written
	// file — the one place state.json actually got corrupted this session
	// was a race between this write and a concurrent read.
	target := statePath()
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

// ---------- git tracking ----------

func repoDir(name string) string {
	return filepath.Join(getBaseDir(), "repo", name)
}

// safeRepoName rejects anything that isn't a plain folder-name segment
// before it reaches repoDir(). Without this, a {name} path value of ".."
// (a single path segment is legal there) collapses filepath.Join(base,
// "repo", "..") back to base itself — turning handleRepoDelete's
// os.RemoveAll into a wipe of the app's own base directory.
func safeRepoName(name string) error {
	if name == "" {
		return fmt.Errorf("repo adi bos olamaz")
	}
	if name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("gecersiz repo adi")
	}
	if filepath.Clean(name) != name {
		return fmt.Errorf("gecersiz repo adi")
	}
	return nil
}

func autoCloneMissingRepos(s *StateBundle) []map[string]interface{} {
	results := []map[string]interface{}{}
	tracked := map[string]bool{}
	for _, t := range s.TrackedRepos {
		tracked[t.Repo] = true
	}

	// Only clone repos the user explicitly pulled in from the GitHub live search
	// tab (tagged category "GitHub" by downloadRepoFromGithub) or added as a
	// custom MCP server with a repo field — never the built-in default skill
	// catalog (DEFAULT_RECOMMENDED_REPOS), which is just a browsable list and
	// must not be auto-installed as a side effect of unrelated state saves.
	var repoRefs []string
	for _, skill := range s.RecommendedRepos {
		if skill.Repo != "" && skill.Category == "GitHub" {
			repoRefs = append(repoRefs, skill.Repo)
		}
	}
	for _, mcp := range s.McpServers {
		if mcp.Repo != "" {
			repoRefs = append(repoRefs, mcp.Repo)
		}
	}

	changed := false
	seen := map[string]bool{}
	for _, repoRef := range repoRefs {
		if seen[repoRef] {
			continue
		}
		seen[repoRef] = true

		parts := strings.Split(repoRef, "/")
		name := parts[len(parts)-1]
		localPath := repoDir(name)
		if _, err := os.Stat(localPath); err == nil {
			continue
		}

		task := addTask("clone", repoRef, fmt.Sprintf("%s deposu repo/%s klasörüne klonlanıyor...", repoRef, name))
		cloneCtx, cloneCancel := context.WithCancel(context.Background())
		task.attachCancel(cloneCancel)
		cmd := enhancedCommand(cloneCtx, "git", "clone", "https://github.com/"+repoRef+".git", localPath)
		out, err := cmd.CombinedOutput()
		wasCancelled := cloneCtx.Err() == context.Canceled
		cloneCancel()
		result := map[string]interface{}{"repo": repoRef, "path": localPath}
		if err != nil && wasCancelled {
			result["cancelled"] = true
			_ = os.RemoveAll(localPath)
			logActivity("clone-stop", repoRef)
			finishTask(task, "cancelled", fmt.Sprintf("%s klonlama işlemi durduruldu.", repoRef))
		} else if err != nil {
			errStr := strings.TrimSpace(string(out))
			result["error"] = errStr
			logActivity("clone-error", fmt.Sprintf("%s: %s", repoRef, errStr))
			finishTask(task, "error", fmt.Sprintf("%s klonlanırken hata: %s", repoRef, errStr))
		} else {
			result["cloned"] = true
			logActivity("clone", fmt.Sprintf("%s -> %s", repoRef, localPath))
			finishTask(task, "completed", fmt.Sprintf("%s deposu başarıyla repo/%s klasörüne indirildi.", repoRef, name))
			if !tracked[repoRef] {
				s.TrackedRepos = append(s.TrackedRepos, TrackedRepo{
					Name: name, Repo: repoRef, LocalPath: localPath,
					Status: "upToDate", LastChecked: time.Now().Format(time.RFC3339),
				})
				tracked[repoRef] = true
				changed = true
			}
		}
		results = append(results, result)
	}

	if changed {
		_ = saveState(s)
	}
	return results
}

func gitCheck(ctx context.Context, name string) (string, error) {
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("yerel klasor yok: %s", dir)
	}

	if out, err := enhancedCommand(ctx, "git", "-C", dir, "fetch").CombinedOutput(); err != nil {
		return "error", fmt.Errorf("git fetch basarisiz: %s", strings.TrimSpace(string(out)))
	}

	localOut, err := enhancedCommand(ctx, "git", "-C", dir, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "error", fmt.Errorf("rev-parse HEAD: %v", err)
	}
	remoteOut, err := enhancedCommand(ctx, "git", "-C", dir, "rev-parse", "@{u}").CombinedOutput()
	if err != nil {
		return "error", fmt.Errorf("upstream yok: %v", err)
	}

	local := strings.TrimSpace(string(localOut))
	remote := strings.TrimSpace(string(remoteOut))
	if local == remote {
		return "upToDate", nil
	}

	countOut, _ := enhancedCommand(ctx, "git", "-C", dir, "rev-list", "--left-right", "--count", "HEAD...@{u}").CombinedOutput()
	fields := strings.Fields(string(countOut))
	if len(fields) == 2 {
		ahead, behind := fields[0], fields[1]
		if behind != "0" {
			return "behind:" + behind, nil
		}
		if ahead != "0" {
			return "ahead:" + ahead, nil
		}
	}
	return "diverged", nil
}

// ---------- zip backup (used by Danger Zone before deleting a folder) ----------

func zipDir(srcDir, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		relPath = filepath.ToSlash(relPath)
		if info.IsDir() {
			_, err := zw.Create(relPath + "/")
			return err
		}
		fw, err := zw.Create(relPath)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(fw, f)
		return err
	})
}

// ---------- HTTP API ----------

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func handleState(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s, err := loadState()
		if err != nil {
			writeErr(w, 500, err.Error())
			return
		}

		pruned := reconcileState(s)
		if len(pruned) > 0 {
			_ = saveState(s)
		}

		resp := map[string]interface{}{
			"mcpServers":       s.McpServers,
			"recommendedRepos": s.RecommendedRepos,
			"idePaths":         s.IdePaths,
			"trackedRepos":     s.TrackedRepos,
		}
		if len(pruned) > 0 {
			resp["prunedRepos"] = pruned
		}
		writeJSON(w, 200, resp)

	case http.MethodPost:
		var incoming StateBundle
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			writeErr(w, 400, "gecersiz JSON: "+err.Error())
			return
		}

		previous, loadErr := loadState()

		// Pages that don't manage trackedRepos (index/skills/paths) push a bundle
		// without that field — preserve whatever is already on disk instead of wiping it.
		if len(incoming.TrackedRepos) == 0 && loadErr == nil {
			incoming.TrackedRepos = previous.TrackedRepos
		}

		// A save that wipes every configured MCP server (the "Varsayılan
		// Ayarlara Sıfırla" button, or any future bulk-clear action) is the
		// one state change with no built-in undo. Snapshot what's about to
		// be lost so it's recoverable from disk instead of gone for good.
		if loadErr == nil && len(previous.McpServers) > 0 && len(incoming.McpServers) == 0 {
			backupDir := filepath.Join(getBaseDir(), "core", "backups")
			if mkErr := os.MkdirAll(backupDir, 0755); mkErr == nil {
				backupPath := filepath.Join(backupDir, fmt.Sprintf("state-%s.json", time.Now().Format("20060102-150405")))
				_ = copyFile(statePath(), backupPath)
			}
		}

		if err := saveState(&incoming); err != nil {
			writeErr(w, 500, err.Error())
			return
		}

		// An id that was in the previous list but isn't in this one was just
		// removed (e.g. the trash icon on an MCP card) — drop it from every
		// real IDE config too, since syncAllIdes below only ever adds/updates.
		if loadErr == nil {
			stillPresent := map[string]bool{}
			for _, m := range incoming.McpServers {
				stillPresent[m.ID] = true
			}
			var removedIds []string
			for _, m := range previous.McpServers {
				if !stillPresent[m.ID] {
					removedIds = append(removedIds, m.ID)
				}
			}
			removeMcpIdsFromAllIdes(removedIds)
		}

		var syncResults []map[string]interface{}
		if len(incoming.McpServers) > 0 {
			syncResults = syncAllIdes(incoming.McpServers)
		}
		cloneResults := autoCloneMissingRepos(&incoming)

		writeJSON(w, 200, map[string]interface{}{
			"ok":     true,
			"sync":   syncResults,
			"cloned": cloneResults,
		})

	default:
		writeErr(w, 405, "method not allowed")
	}
}

func handleStateMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}

	var incoming struct {
		McpServers       []MCPEntry             `json:"mcpServers"`
		RecommendedRepos []SkillEntry           `json:"recommendedRepos"`
		IdePaths         []IdePathEntry         `json:"idePaths"`
		TrackedRepos     []TrackedRepo          `json:"trackedRepos"`
		DeletedIdeIDs    []string               `json:"deletedIdeIds,omitempty"`
		LoopConfig       LoopEngineConfig       `json:"loopConfig,omitempty"`
		UIPreferences    map[string]interface{} `json:"uiPreferences,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeErr(w, 400, "geçersiz yedek JSON dosyası: "+err.Error())
		return
	}

	s, err := loadState()
	if err != nil {
		s = &StateBundle{}
	}

	if len(incoming.McpServers) > 0 {
		s.McpServers = incoming.McpServers
	}
	if len(incoming.RecommendedRepos) > 0 {
		s.RecommendedRepos = incoming.RecommendedRepos
	}
	if len(incoming.TrackedRepos) > 0 {
		s.TrackedRepos = incoming.TrackedRepos
	}
	if len(incoming.DeletedIdeIDs) > 0 {
		s.DeletedIdeIDs = incoming.DeletedIdeIDs
	}
	if incoming.LoopConfig.MaxRetries > 0 {
		s.LoopConfig = incoming.LoopConfig
	}

	if err := saveState(s); err != nil {
		writeErr(w, 500, "durum kaydedilemedi: "+err.Error())
		return
	}

	newIdes := detectIdes()
	s.IdePaths = make([]IdePathEntry, 0, len(newIdes))
	for _, ide := range newIdes {
		if ide.Detected {
			s.IdePaths = append(s.IdePaths, IdePathEntry{
				ID:     ide.ID,
				Name:   ide.Name,
				Path:   ide.Path,
				Status: "✓ Algılandı",
			})
		}
	}
	_ = saveState(s)

	syncResults := []map[string]interface{}{}
	if len(s.McpServers) > 0 {
		syncResults = syncAllIdes(s.McpServers)
	}

	cloneResults := autoCloneMissingRepos(s)

	logActivity("state-migrate", fmt.Sprintf("Yedek paketi yüklendi: %d MCP, %d Repo klonlandı", len(s.McpServers), len(cloneResults)))

	writeJSON(w, 200, map[string]interface{}{
		"ok":            true,
		"syncedIdes":    syncResults,
		"clonedRepos":   cloneResults,
		"uiPreferences": incoming.UIPreferences,
		"message":       fmt.Sprintf("🎉 WyvDev yedek paketi yüklendi! %d MCP sunucusu yerel IDE'lere işlendi, %d repo otomatik klonlanıyor.", len(s.McpServers), len(cloneResults)),
	})
}

func handleIdesDetect(w http.ResponseWriter, r *http.Request) {
	resetAll := r.URL.Query().Get("reset") == "true"

	ides := detectIdes()
	s, err := loadState()
	if err != nil {
		s = &StateBundle{}
	}

	if resetAll {
		s.DeletedIdeIDs = nil
		s.IdePaths = nil
	}

	deletedSet := map[string]bool{}
	for _, id := range s.DeletedIdeIDs {
		deletedSet[id] = true
	}

	existingByID := map[string]IdePathEntry{}
	for _, p := range s.IdePaths {
		existingByID[p.ID] = p
	}

	merged := make([]IdePathEntry, 0, len(ides))
	for _, ide := range ides {
		if deletedSet[ide.ID] && !resetAll {
			continue
		}

		path := ide.Path
		if existing, ok := existingByID[ide.ID]; ok && existing.Path != "" {
			path = existing.Path
		}

		status := "❌ Klasör Bulunamadı"
		if _, err := os.Stat(filepath.Dir(path)); err == nil {
			status = "✅ Algılandı"
		} else if _, err := os.Stat(path); err == nil {
			status = "✅ Algılandı"
		}

		if status == "✅ Algılandı" || (existingByID[ide.ID].Path != "" && !deletedSet[ide.ID]) {
			merged = append(merged, IdePathEntry{ID: ide.ID, Name: ide.Name, Path: path, Status: status})
		}
	}

	s.IdePaths = merged
	_ = saveState(s)

	// A newly-installed IDE picked up by this scan shouldn't have to wait for
	// the next unrelated MCP edit to receive the existing config — push it
	// immediately so "detect" always leaves every IDE up to date too.
	if len(s.McpServers) > 0 {
		syncAllIdes(s.McpServers)
	}

	writeJSON(w, 200, merged)
}

func handleDangerDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body struct {
		ConfirmName string `json:"confirmName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "gecersiz JSON")
		return
	}

	var target *DetectedIde
	for _, ide := range detectIdes() {
		if ide.ID == id {
			t := ide
			target = &t
			break
		}
	}
	if target == nil {
		writeErr(w, 404, "IDE bulunamadi")
		return
	}
	if body.ConfirmName != target.Name {
		writeErr(w, 400, "onay adi eslesmiyor")
		return
	}

	dir := filepath.Dir(target.Path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		writeErr(w, 404, "klasor zaten yok: "+dir)
		return
	}

	backupDir := filepath.Join(getBaseDir(), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		writeErr(w, 500, "yedek klasoru olusturulamadi: "+err.Error())
		return
	}
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s-%s.zip", target.ID, time.Now().Format("20060102-150405")))
	if err := zipDir(dir, backupPath); err != nil {
		writeErr(w, 500, "zip yedek hatasi: "+err.Error())
		return
	}

	// Physically delete the IDE folder from volume / disk
	if err := os.RemoveAll(dir); err != nil {
		writeErr(w, 500, "silme hatasi (yedek alindi: "+backupPath+"): "+err.Error())
		return
	}

	// Update state & reconcile across all IDE config files
	s, loadErr := loadState()
	if loadErr == nil {
		var newIdePaths []IdePathEntry
		for _, p := range s.IdePaths {
			if p.ID != target.ID {
				newIdePaths = append(newIdePaths, p)
			}
		}
		s.IdePaths = newIdePaths

		alreadyDeleted := false
		for _, d := range s.DeletedIdeIDs {
			if d == target.ID {
				alreadyDeleted = true
				break
			}
		}
		if !alreadyDeleted {
			s.DeletedIdeIDs = append(s.DeletedIdeIDs, target.ID)
		}

		_ = saveState(s)
		syncAllIdes(s.McpServers)
	}

	logActivity("danger-delete", fmt.Sprintf("%s silindi, volume senkronize edildi, yedek: %s", dir, backupPath))
	writeJSON(w, 200, map[string]interface{}{
		"ok":      true,
		"backup":  backupPath,
		"deleted": dir,
		"message": fmt.Sprintf("'%s' IDE klasörü diskten silindi ve tüm IDE yapılandırmalarıyla tam senkronize edildi.", target.Name),
	})
}

// ---------- Master IDE Mirror Sync Logic ----------

func readMcpEntriesFromIdeConfig(ide *DetectedIde) ([]MCPEntry, error) {
	if ide == nil || ide.Path == "" {
		return nil, fmt.Errorf("master ide yolu geçersiz")
	}
	data, err := os.ReadFile(ide.Path)
	if err != nil {
		return nil, err
	}

	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	var mcpMap map[string]interface{}
	if m, ok := root["mcpServers"].(map[string]interface{}); ok && m != nil {
		mcpMap = m
	} else if m, ok := root["context_servers"].(map[string]interface{}); ok && m != nil {
		mcpMap = m
	}

	entries := []MCPEntry{}
	for name, v := range mcpMap {
		srv, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		entry := MCPEntry{
			ID:   name,
			Name: name,
		}
		if t, ok := srv["type"].(string); ok && t == "http" {
			entry.Type = "http"
			if urlStr, ok := srv["url"].(string); ok {
				entry.URL = urlStr
			}
		} else if urlStr, ok := srv["url"].(string); ok && urlStr != "" {
			entry.Type = "http"
			entry.URL = urlStr
		} else {
			entry.Type = "command"
			if cmd, ok := srv["command"].(string); ok {
				entry.Command = cmd
			} else if cmdObj, ok := srv["command"].(map[string]interface{}); ok {
				if p, ok := cmdObj["path"].(string); ok {
					entry.Command = p
				}
				if argsSlice, ok := cmdObj["args"].([]interface{}); ok {
					for _, a := range argsSlice {
						if str, ok := a.(string); ok {
							entry.Args = append(entry.Args, str)
						}
					}
				}
			}
			if argsSlice, ok := srv["args"].([]interface{}); ok {
				for _, a := range argsSlice {
					if str, ok := a.(string); ok {
						entry.Args = append(entry.Args, str)
					}
				}
			}
			if envMap, ok := srv["env"].(map[string]interface{}); ok {
				entry.Env = make(map[string]string)
				for ek, ev := range envMap {
					if str, ok := ev.(string); ok {
						entry.Env[ek] = str
					}
				}
			}
			if cwd, ok := srv["cwd"].(string); ok {
				entry.Cwd = cwd
			}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func syncSkillsFromMasterIde(masterIde *DetectedIde) (int, error) {
	if masterIde == nil || masterIde.SkillsDir == "" {
		return 0, nil
	}
	if _, err := os.Stat(masterIde.SkillsDir); err != nil {
		return 0, nil
	}
	skillFolders := findAllSkillFolders(masterIde.SkillsDir)
	if len(skillFolders) == 0 {
		return 0, nil
	}

	totalCopied := 0
	for _, ide := range detectIdes() {
		if !ide.Detected || ide.SkillsDir == "" || ide.ID == masterIde.ID {
			continue
		}
		if err := ensureValidDir(ide.SkillsDir); err != nil {
			continue
		}
		for _, folder := range skillFolders {
			sName := filepath.Base(folder)
			destDir := filepath.Join(ide.SkillsDir, sName)
			if err := copyDirRecursive(folder, destDir); err == nil {
				totalCopied++
			}
		}
	}
	return totalCopied, nil
}

func executeMasterIdeMirrorSync(masterID string) (map[string]interface{}, error) {
	ides := detectIdes()
	var masterIde *DetectedIde
	for _, ide := range ides {
		if ide.ID == masterID {
			t := ide
			masterIde = &t
			break
		}
	}
	if masterIde == nil {
		return nil, fmt.Errorf("master IDE bulunamadı: %s", masterID)
	}

	entries, err := readMcpEntriesFromIdeConfig(masterIde)
	if err != nil {
		logActivity("master-ide-read-warning", fmt.Sprintf("%s: %v", masterIde.Name, err))
	}

	s, _ := loadState()
	if s == nil {
		s = &StateBundle{}
	}
	s.MasterIDE = masterID
	if len(entries) > 0 {
		s.McpServers = entries
	}
	_ = saveState(s)

	syncResults := syncAllIdes(s.McpServers)
	skillsCopied, _ := syncSkillsFromMasterIde(masterIde)

	logActivity("master-ide-sync", fmt.Sprintf("Master (%s) -> %d MCP, %d Skill tüm IDE'lere aktarıldı", masterIde.Name, len(s.McpServers), skillsCopied))

	return map[string]interface{}{
		"ok":           true,
		"masterName":   masterIde.Name,
		"masterId":     masterID,
		"mcpCount":     len(s.McpServers),
		"skillsCopied": skillsCopied,
		"syncResults":  syncResults,
		"message":      fmt.Sprintf("✨ '%s' (Ana IDE) üzerindeki %d MCP ve %d Skill diğer tüm IDE'lere doğrudan aktarıldı!", masterIde.Name, len(s.McpServers), skillsCopied),
	}, nil
}

// ---------- MCP Connection Test ----------

func handleMcpTest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		writeErr(w, 400, "id alanı gerekli")
		return
	}

	s, err := loadState()
	if err != nil {
		s = &StateBundle{}
	}
	var entry *MCPEntry
	for i := range s.McpServers {
		if s.McpServers[i].ID == body.ID {
			entry = &s.McpServers[i]
			break
		}
	}
	if entry == nil {
		writeErr(w, 404, "MCP bulunamadı: "+body.ID)
		return
	}

	if entry.Type == "http" {
		if entry.URL == "" {
			writeJSON(w, 200, map[string]interface{}{"ok": false, "message": "⚠️ URL tanımlı değil"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		req, rErr := http.NewRequestWithContext(ctx, http.MethodGet, entry.URL, nil)
		if rErr != nil {
			writeJSON(w, 200, map[string]interface{}{"ok": false, "message": "⚠️ Geçersiz URL: " + rErr.Error()})
			return
		}
		for k, v := range entry.Headers {
			req.Header.Set(k, v)
		}
		resp, dErr := http.DefaultClient.Do(req)
		if dErr != nil {
			writeJSON(w, 200, map[string]interface{}{"ok": false, "message": "⚠️ Bağlantı hatası: " + dErr.Error()})
			return
		}
		defer resp.Body.Close()
		writeJSON(w, 200, map[string]interface{}{
			"ok":         true,
			"statusCode": resp.StatusCode,
			"message":    fmt.Sprintf("✅ Sunucuya ulaşıldı (HTTP %d)", resp.StatusCode),
		})
		return
	}

	if entry.Command == "" {
		writeJSON(w, 200, map[string]interface{}{"ok": false, "message": "⚠️ Komut tanımlı değil"})
		return
	}

	env := getEnhancedEnv()
	resolved := resolveInEnv(entry.Command, env)
	if resolved == "" {
		writeJSON(w, 200, map[string]interface{}{"ok": false, "message": fmt.Sprintf("⚠️ Komut bulunamadı: %s (PATH içinde yok)", entry.Command)})
		return
	}

	// Real MCP handshake: spawn the server and send an actual JSON-RPC
	// "initialize" request over stdin, then wait for its response on stdout.
	// A process merely *starting* proves nothing — many MCP servers exit
	// immediately when a required token/env var is missing, which this
	// would otherwise report as a false "success".
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, resolved, entry.Args...)
	cmd.Env = env
	if entry.Cwd != "" {
		cmd.Dir = entry.Cwd
	}
	for k, v := range entry.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, sErr := cmd.StdinPipe()
	if sErr != nil {
		writeJSON(w, 200, map[string]interface{}{"ok": false, "message": "⚠️ stdin açılamadı: " + sErr.Error()})
		return
	}
	stdout, oErr := cmd.StdoutPipe()
	if oErr != nil {
		writeJSON(w, 200, map[string]interface{}{"ok": false, "message": "⚠️ stdout açılamadı: " + oErr.Error()})
		return
	}
	stderrPipe, _ := cmd.StderrPipe()

	if startErr := cmd.Start(); startErr != nil {
		writeJSON(w, 200, map[string]interface{}{"ok": false, "message": "⚠️ Süreç başlatılamadı: " + startErr.Error()})
		return
	}

	var stderrBuf strings.Builder
	var stderrMux sync.Mutex
	if stderrPipe != nil {
		go func() {
			sc := bufio.NewScanner(stderrPipe)
			for sc.Scan() {
				stderrMux.Lock()
				if stderrBuf.Len() < 2000 {
					stderrBuf.WriteString(sc.Text())
					stderrBuf.WriteString("\n")
				}
				stderrMux.Unlock()
			}
		}()
	}

	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "wyvdev", "version": "1.0"},
		},
	}
	reqBytes, _ := json.Marshal(initReq)
	_, _ = stdin.Write(append(reqBytes, '\n'))

	type readResult struct {
		line string
		err  error
	}
	lineCh := make(chan readResult, 1)
	go func() {
		reader := bufio.NewReaderSize(stdout, 64*1024)
		for {
			line, rErr := reader.ReadString('\n')
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				lineCh <- readResult{line: trimmed}
				return
			}
			if rErr != nil {
				lineCh <- readResult{err: rErr}
				return
			}
		}
	}()

	stderrExcerpt := func() string {
		stderrMux.Lock()
		defer stderrMux.Unlock()
		return strings.TrimSpace(stderrBuf.String())
	}

	var resultOk bool
	var resultMsg string
	select {
	case res := <-lineCh:
		if res.err != nil {
			if detail := stderrExcerpt(); detail != "" {
				resultMsg = "⚠️ Sunucu yanıt vermeden kapandı: " + truncateStr(detail, 300)
			} else {
				resultMsg = "⚠️ Sunucudan yanıt alınamadı (stdout kapandı)"
			}
		} else {
			var parsed struct {
				Result *struct {
					ServerInfo *struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"serverInfo"`
				} `json:"result"`
				Error *struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if jErr := json.Unmarshal([]byte(res.line), &parsed); jErr != nil {
				resultMsg = "⚠️ Geçersiz JSON-RPC yanıtı: " + truncateStr(res.line, 200)
			} else if parsed.Error != nil {
				resultMsg = "⚠️ Sunucu hata döndü: " + parsed.Error.Message
			} else if parsed.Result != nil {
				resultOk = true
				if parsed.Result.ServerInfo != nil && parsed.Result.ServerInfo.Name != "" {
					resultMsg = fmt.Sprintf("✅ MCP handshake başarılı — %s v%s yanıt verdi", parsed.Result.ServerInfo.Name, parsed.Result.ServerInfo.Version)
				} else {
					resultMsg = "✅ MCP handshake başarılı — initialize yanıtı alındı"
				}
			} else {
				resultMsg = "⚠️ Beklenmeyen yanıt: " + truncateStr(res.line, 200)
			}
		}
	case <-ctx.Done():
		if detail := stderrExcerpt(); detail != "" {
			resultMsg = "⚠️ Zaman aşımı (8sn) — " + truncateStr(detail, 300)
		} else {
			resultMsg = "⚠️ Zaman aşımı (8sn) — sunucu initialize yanıtı vermedi"
		}
	}

	_ = cmd.Process.Kill()
	_ = cmd.Wait()

	writeJSON(w, 200, map[string]interface{}{"ok": resultOk, "message": resultMsg})
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ---------- Marketplace Handlers ----------

type MarketplaceItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Category string `json:"category"`
	Repo     string `json:"repo"`
	Desc     string `json:"desc"`
	Icon     string `json:"icon"`
	Stars    string `json:"stars"`
	Badge    string `json:"badge"`
}

func getMarketplaceCatalogItems() []MarketplaceItem {
	return []MarketplaceItem{
		{
			ID:       "vintlin-skill-flow",
			Name:     "SkillFlow Master Pack",
			Type:     "skill",
			Category: "📄 Prompt & Agent Skills",
			Repo:     "VintLin/skill-flow",
			Desc:     "Install, manage, and share skills across Claude Code, Cursor, Copilot, and Windsurf.",
			Icon:     "sparkles",
			Stars:    "4.8k",
			Badge:    "🔥 Popüler",
		},
		{
			ID:       "realzst-harnesskit",
			Name:     "HarnessKit Agent Rules & Memory",
			Type:     "skill",
			Category: "🧠 Agent Rules & Memory",
			Repo:     "RealZST/HarnessKit",
			Desc:     "Manage skills, MCP servers, plugins, hooks, CLIs, configs, memory & rules across AI coding agents.",
			Icon:     "brain",
			Stars:    "3.9k",
			Badge:    "⭐ Tavsiye",
		},
		{
			ID:       "mode-io-skill-manager",
			Name:     "SkillManager Cross-IDE Pack",
			Type:     "skill",
			Category: "📄 Prompt & Agent Skills",
			Repo:     "mode-io/skill-manager",
			Desc:     "Manage skills across Codex CLI, Claude Code, Cursor, OpenCode, and OpenClaw.",
			Icon:     "layers",
			Stars:    "2.9k",
			Badge:    "⚡ Hızlı",
		},
		{
			ID:       "wanghuan9-skilldock",
			Name:     "SkillDock Diff & Sync Pack",
			Type:     "skill",
			Category: "📄 Prompt & Agent Skills",
			Repo:     "wanghuan9/skilldock",
			Desc:     "Real-directory scanning and Git-aware Diff previews for AI coding tools.",
			Icon:     "git-compare",
			Stars:    "2.1k",
			Badge:    "📈 Trend",
		},
		{
			ID:       "modelcontextprotocol-servers",
			Name:     "MCP Official Reference Servers",
			Type:     "mcp",
			Category: "🔌 Protocol & Integration",
			Repo:     "modelcontextprotocol/servers",
			Desc:     "Official reference MCP servers: SQLite, Filesystem, GitHub, Git, Memory, Everything.",
			Icon:     "server",
			Stars:    "18.4k",
			Badge:    "Resmî MCP",
		},
		{
			ID:       "supabase-mcp",
			Name:     "Supabase Database MCP",
			Type:     "mcp",
			Category: "⚡ Database & Backend",
			Repo:     "supabase/mcp-server",
			Desc:     "Inspect database schemas, execute queries, and manage migrations via MCP protocol.",
			Icon:     "database",
			Stars:    "5.6k",
			Badge:    "Veritabanı",
		},
		{
			ID:       "puppeteer-mcp",
			Name:     "Puppeteer Web Automation MCP",
			Type:     "mcp",
			Category: "🌐 Web Automation",
			Repo:     "modelcontextprotocol/servers",
			Desc:     "Automate web browsing, screenshot capture, and web page DOM inspection.",
			Icon:     "globe",
			Stars:    "8.1k",
			Badge:    "Otomasyon",
		},
	}
}

func handleMarketplaceCatalog(w http.ResponseWriter, r *http.Request) {
	items := getMarketplaceCatalogItems()
	writeJSON(w, 200, items)
}

func handleMarketplaceInstall(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Repo string `json:"repo"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "geçersiz JSON")
		return
	}
	if body.Repo == "" {
		writeErr(w, 400, "repo alanı gerekli")
		return
	}

	repoName := body.Name
	if repoName == "" {
		parts := strings.Split(body.Repo, "/")
		repoName = parts[len(parts)-1]
	}

	task := addTask("clone", repoName, fmt.Sprintf("Marketplace paketi indiriliyor: %s", body.Repo))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		task.attachCancel(cancel)

		targetDir := repoDir(repoName)
		if _, err := os.Stat(targetDir); err == nil {
			_ = enhancedCommand(ctx, "git", "-C", targetDir, "pull").Run()
		} else {
			cmd := enhancedCommand(ctx, "git", "clone", "https://github.com/"+body.Repo+".git", targetDir)
			out, err := cmd.CombinedOutput()
			if err != nil {
				finishTask(task, "error", fmt.Sprintf("Klonlama hatası: %v (%s)", err, string(out)))
				return
			}
		}

		s, _ := loadState()
		if s == nil {
			s = &StateBundle{}
		}

		alreadyTracked := false
		for _, t := range s.TrackedRepos {
			if t.Repo == body.Repo || t.Name == repoName {
				alreadyTracked = true
				break
			}
		}
		if !alreadyTracked {
			s.TrackedRepos = append(s.TrackedRepos, TrackedRepo{
				Repo:        body.Repo,
				Name:        repoName,
				LocalPath:   targetDir,
				Status:      "✅ Güncel",
				LastChecked: time.Now().Format(time.RFC3339),
			})
			_ = saveState(s)
		}

		_, _, _ = enableSkillForRepo(repoName)
		if s.MasterIDE != "" {
			_, _ = executeMasterIdeMirrorSync(s.MasterIDE)
		} else {
			_, _ = executeMasterIdeMirrorSync("cursor")
		}

		finishTask(task, "completed", fmt.Sprintf("✅ '%s' Marketplace paketi kuruldu ve tüm IDE'lere aktarıldı!", repoName))
	}()

	writeJSON(w, 200, map[string]interface{}{
		"ok":      true,
		"task":    task,
		"message": fmt.Sprintf("🚀 '%s' paketi indirilip kuruluyor...", repoName),
	})
}


func handleRepoCheck(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	status, err := gitCheck(r.Context(), name)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}

	s, loadErr := loadState()
	if loadErr == nil {
		for i := range s.TrackedRepos {
			if s.TrackedRepos[i].Name == name {
				s.TrackedRepos[i].Status = status
				s.TrackedRepos[i].LastChecked = time.Now().Format(time.RFC3339)
			}
		}
		_ = saveState(s)
	}

	writeJSON(w, 200, map[string]string{"status": status})
}

func handleRepoPull(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		writeErr(w, 404, "yerel klasor yok")
		return
	}
	out, err := enhancedCommand(r.Context(), "git", "-C", dir, "pull").CombinedOutput()
	if err != nil {
		writeErr(w, 500, strings.TrimSpace(string(out)))
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "true", "output": strings.TrimSpace(string(out))})
}

func handleRepoDelete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}

	// Stop whatever is running against this folder before the folder itself
	// goes away — otherwise the process (and its port) is orphaned, still
	// running against a directory that no longer exists.
	killRunningApp(name)

	dir := repoDir(name)
	deletedOnDisk := false
	if _, err := os.Stat(dir); err == nil {
		if err := os.RemoveAll(dir); err != nil {
			writeErr(w, 500, fmt.Sprintf("klasor silinemedi (%s): %v", dir, err))
			return
		}
		deletedOnDisk = true
	}

	// Remove any copy of this repo's skill that enable-skill placed in an
	// IDE's skills folder — a deleted repo shouldn't leave a skill behind
	// that still claims to come from it.
	for _, ide := range detectIdes() {
		if ide.SkillsDir != "" {
			_ = os.RemoveAll(filepath.Join(ide.SkillsDir, name))
		}
	}

	s, err := loadState()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}

	var stillTracked []TrackedRepo
	for _, t := range s.TrackedRepos {
		parts := strings.Split(t.Repo, "/")
		repoName := parts[len(parts)-1]
		if t.Name != name && t.Repo != name && repoName != name {
			stillTracked = append(stillTracked, t)
		}
	}
	s.TrackedRepos = stillTracked

	var stillSkills []SkillEntry
	for _, sk := range s.RecommendedRepos {
		parts := strings.Split(sk.Repo, "/")
		repoName := parts[len(parts)-1]
		if sk.ID != name && sk.Name != name && sk.Repo != name && repoName != name {
			stillSkills = append(stillSkills, sk)
		}
	}
	s.RecommendedRepos = stillSkills

	var stillMcp []MCPEntry
	var removedMcpIds []string
	for _, m := range s.McpServers {
		parts := strings.Split(m.Repo, "/")
		repoName := parts[len(parts)-1]
		if m.ID != name && m.Name != name && m.Repo != name && repoName != name {
			stillMcp = append(stillMcp, m)
		} else {
			removedMcpIds = append(removedMcpIds, m.ID)
		}
	}
	s.McpServers = stillMcp

	pruned := reconcileState(s)
	_ = saveState(s)

	// stillMcp's own entries get (re)synced below; removedMcpIds is exactly
	// what's no longer supposed to exist anywhere, so strip it explicitly —
	// syncAllIdes only ever adds/updates, it never deletes a stale key.
	removeMcpIdsFromAllIdes(removedMcpIds)
	syncResults := syncAllIdes(s.McpServers)

	logActivity("delete-repo", fmt.Sprintf("%s (disk: %v)", name, deletedOnDisk))

	writeJSON(w, 200, map[string]interface{}{
		"ok":            true,
		"name":          name,
		"deletedOnDisk": deletedOnDisk,
		"pruned":        pruned,
		"sync":          syncResults,
		"state":         s,
	})
}

// ---------- real per-repo install/build/run (npm/pip/cargo/docker) ----------

func sanitizeDockerTag(name string) string {
	lower := strings.ToLower(name)
	var b strings.Builder
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	return b.String()
}

func handleRepoRun(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		writeErr(w, 404, "yerel klasor yok")
		return
	}

	var body struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "gecersiz JSON")
		return
	}

	safeTag := "ai-toolkit-" + sanitizeDockerTag(name)

	var toolName string
	var args []string
	switch body.Action {
	case "npm-install":
		toolName, args = "npm", []string{"install"}
	case "npm-repair":
		toolName, args = "npm", []string{"install", "--force"}
	case "pip-install":
		toolName, args = "pip", []string{"install", "-r", "requirements.txt"}
	case "pip-repair":
		toolName, args = "python", []string{"-m", "pip", "install", "--upgrade", "--force-reinstall", "-r", "requirements.txt"}
	case "cargo-build":
		toolName, args = "cargo", []string{"build", "--release"}
	case "cargo-repair":
		toolName, args = "cargo", []string{"clean"}
	case "docker-build":
		toolName, args = "docker", []string{"build", "-t", safeTag, "."}
	case "docker-repair":
		toolName, args = "docker", []string{"build", "--no-cache", "-t", safeTag, "."}
	case "docker-run":
		toolName, args = "docker", []string{"run", "--rm", "-d", "--name", safeTag + "-run", safeTag}
	default:
		writeErr(w, 400, "bilinmeyen aksiyon")
		return
	}

	if !commandExistsInEnv(toolName, getEnhancedEnv()) {
		writeErr(w, 424, fmt.Sprintf("%s bulunamadi. Sistem & Teşhis sayfasından kurabilirsiniz.", toolName))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	cmd := enhancedCommand(ctx, toolName, args...)
	cmd.Dir = dir
	out, runErr := cmd.CombinedOutput()

	if runErr == nil {
		repoStartErrorsMux.Lock()
		delete(repoStartErrors, name)
		repoStartErrorsMux.Unlock()
	}

	result := map[string]interface{}{
		"command": toolName + " " + strings.Join(args, " "),
		"output":  strings.TrimSpace(string(out)),
		"ok":      runErr == nil,
	}
	if runErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result["error"] = "islem zaman asimina ugradi (180sn)"
		} else {
			result["error"] = runErr.Error()
		}
		logActivity("repo-run-error", fmt.Sprintf("%s (%s): %v", name, body.Action, result["error"]))
	} else {
		logActivity("repo-run", fmt.Sprintf("%s (%s)", name, body.Action))
	}
	writeJSON(w, 200, result)
}

// ---------- Gelişmiş Kurulum: Bağımlılık Analizi ----------

type InstallStep struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Cmd      string `json:"cmd"`
	Required bool   `json:"required"`
}

type AnalyzeResult struct {
	Name            string        `json:"name"`
	Runtimes        []string      `json:"runtimes"`
	TotalDeps       int           `json:"totalDeps"`
	InstalledDeps   int           `json:"installedDeps"`
	InstallPercent  int           `json:"installPercent"`
	MissingFiles    []string      `json:"missingFiles"`
	InstallSteps    []InstallStep `json:"installSteps"`
	EnvVarsNeeded   []string      `json:"envVarsNeeded"`
	PortSuggestion  int           `json:"portSuggestion"`
	DiskEstimateMB  int           `json:"diskEstimateMB"`
	HasDockerfile   bool          `json:"hasDockerfile"`
	HasCompose      bool          `json:"hasCompose"`
	HasMakefile     bool          `json:"hasMakefile"`
}

func handleRepoAnalyze(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		writeErr(w, 404, "yerel klasör yok")
		return
	}

	result := AnalyzeResult{Name: name}
	runtimeSet := map[string]bool{}
	steps := []InstallStep{}

	// --- Node.js ---
	pkgJson := filepath.Join(dir, "package.json")
	if data, err := os.ReadFile(pkgJson); err == nil {
		runtimeSet["node"] = true
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			count := 0
			if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
				count += len(deps)
			}
			if dev, ok := pkg["devDependencies"].(map[string]interface{}); ok {
				count += len(dev)
			}
			result.TotalDeps += count
			result.DiskEstimateMB += count / 5 // rough estimate ~200KB per package
		}
		nmDir := filepath.Join(dir, "node_modules")
		if _, err := os.Stat(nmDir); err == nil {
			result.InstalledDeps += result.TotalDeps
		} else {
			result.MissingFiles = append(result.MissingFiles, "node_modules")
		}
		steps = append(steps, InstallStep{ID: "npm-install", Label: "npm install", Cmd: "npm install", Required: true})

		// Check for build scripts
		if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
			if _, hasBuild := scripts["build"]; hasBuild {
				steps = append(steps, InstallStep{ID: "npm-build", Label: "npm run build", Cmd: "npm run build", Required: false})
			}
			// Detect common port patterns
			if start, ok := scripts["start"].(string); ok {
				if strings.Contains(start, "3000") {
					result.PortSuggestion = 3000
				} else if strings.Contains(start, "8080") {
					result.PortSuggestion = 8080
				}
			}
		}
	}

	// --- Python ---
	reqTxt := filepath.Join(dir, "requirements.txt")
	pyproject := filepath.Join(dir, "pyproject.toml")
	if _, err := os.Stat(reqTxt); err == nil {
		runtimeSet["python"] = true
		if data, err := os.ReadFile(reqTxt); err == nil {
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			count := 0
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if l != "" && !strings.HasPrefix(l, "#") {
					count++
				}
			}
			result.TotalDeps += count
			result.DiskEstimateMB += count * 2
		}
		venvDir := filepath.Join(dir, "venv")
		venvDir2 := filepath.Join(dir, ".venv")
		if _, err1 := os.Stat(venvDir); err1 == nil {
			result.InstalledDeps += result.TotalDeps
		} else if _, err2 := os.Stat(venvDir2); err2 == nil {
			result.InstalledDeps += result.TotalDeps
		} else {
			result.MissingFiles = append(result.MissingFiles, "venv/")
			steps = append(steps, InstallStep{ID: "pip-install", Label: "pip install -r requirements.txt", Cmd: "pip install -r requirements.txt", Required: true})
		}
	} else if _, err := os.Stat(pyproject); err == nil {
		runtimeSet["python"] = true
		steps = append(steps, InstallStep{ID: "pip-install-pyproject", Label: "pip install -e .", Cmd: "pip install -e .", Required: true})
	}

	// --- Rust / Cargo ---
	cargoToml := filepath.Join(dir, "Cargo.toml")
	if _, err := os.Stat(cargoToml); err == nil {
		runtimeSet["rust"] = true
		targetDir := filepath.Join(dir, "target", "release")
		if _, err := os.Stat(targetDir); err == nil {
			result.InstalledDeps++
		} else {
			result.MissingFiles = append(result.MissingFiles, "target/")
			steps = append(steps, InstallStep{ID: "cargo-build", Label: "cargo build --release", Cmd: "cargo build --release", Required: true})
		}
		result.DiskEstimateMB += 200
	}

	// --- Go ---
	goMod := filepath.Join(dir, "go.mod")
	if data, err := os.ReadFile(goMod); err == nil {
		runtimeSet["go"] = true
		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(l), "require") {
				result.TotalDeps++
			}
		}
		steps = append(steps, InstallStep{ID: "go-build", Label: "go build ./...", Cmd: "go build ./...", Required: true})
	}

	// --- Docker ---
	dockerfile := filepath.Join(dir, "Dockerfile")
	composePaths := []string{
		filepath.Join(dir, "docker-compose.yml"),
		filepath.Join(dir, "docker-compose.yaml"),
		filepath.Join(dir, "compose.yml"),
	}
	if _, err := os.Stat(dockerfile); err == nil {
		runtimeSet["docker"] = true
		result.HasDockerfile = true
		steps = append(steps, InstallStep{ID: "docker-build", Label: "docker build", Cmd: "docker build -t " + name + " .", Required: true})
		steps = append(steps, InstallStep{ID: "docker-run", Label: "docker run", Cmd: "docker run --rm -d " + name, Required: false})
	}
	for _, cp := range composePaths {
		if _, err := os.Stat(cp); err == nil {
			result.HasCompose = true
			steps = append(steps, InstallStep{ID: "compose-up", Label: "docker compose up -d", Cmd: "docker compose up -d", Required: true})
			break
		}
	}

	// --- Makefile ---
	makefile := filepath.Join(dir, "Makefile")
	if _, err := os.Stat(makefile); err == nil {
		result.HasMakefile = true
	}

	// --- ENV variables ---
	envExample := filepath.Join(dir, ".env.example")
	envLocal := filepath.Join(dir, ".env")
	if data, err := os.ReadFile(envExample); err == nil {
		lines := strings.Split(string(data), "\n")
		envMissing := []string{}
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			parts := strings.SplitN(l, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := ""
			if len(parts) > 1 {
				val = strings.TrimSpace(parts[1])
			}
			if val == "" && key != "" {
				envMissing = append(envMissing, key)
			}
		}
		if _, err := os.Stat(envLocal); err != nil {
			result.EnvVarsNeeded = envMissing
			if len(envMissing) > 0 {
				steps = append(steps, InstallStep{ID: "copy-env", Label: "cp .env.example .env", Cmd: "cp .env.example .env", Required: false})
			}
		}
	}

	// Runtimes list
	for rt := range runtimeSet {
		result.Runtimes = append(result.Runtimes, rt)
	}

	// Install percentage
	if result.TotalDeps > 0 {
		result.InstallPercent = int(float64(result.InstalledDeps) / float64(result.TotalDeps) * 100)
	} else if len(result.MissingFiles) == 0 && len(steps) == 0 {
		result.InstallPercent = 100
	}

	result.InstallSteps = steps
	writeJSON(w, 200, result)
}

// ---------- Gelişmiş Kurulum: SSE Gerçek Zamanlı Log Akışı ----------

func handleRepoInstallStream(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		http.Error(w, "yerel klasör yok", 404)
		return
	}

	stepID := r.URL.Query().Get("step")
	if stepID == "" {
		http.Error(w, "step parametresi gerekli", 400)
		return
	}

	safeTag := "wyvdev-" + sanitizeDockerTag(name)
	var toolName string
	var args []string
	switch stepID {
	case "npm-install":
		toolName, args = "npm", []string{"install"}
	case "npm-build":
		toolName, args = "npm", []string{"run", "build"}
	case "pip-install":
		toolName, args = "pip", []string{"install", "-r", "requirements.txt"}
	case "pip-install-pyproject":
		toolName, args = "pip", []string{"install", "-e", "."}
	case "pip-repair":
		toolName, args = "python", []string{"-m", "pip", "install", "--upgrade", "--force-reinstall", "-r", "requirements.txt"}
	case "cargo-build":
		toolName, args = "cargo", []string{"build", "--release"}
	case "go-build":
		toolName, args = "go", []string{"build", "./..."}
	case "docker-build":
		toolName, args = "docker", []string{"build", "-t", safeTag, "."}
	case "compose-up":
		toolName, args = "docker", []string{"compose", "up", "-d"}
	case "copy-env":
		// Special: just copy .env.example to .env
		src := filepath.Join(dir, ".env.example")
		dst := filepath.Join(dir, ".env")
		data, err := os.ReadFile(src)
		if err != nil {
			http.Error(w, ".env.example bulunamadı", 404)
			return
		}
		_ = os.WriteFile(dst, data, 0644)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		fmt.Fprintf(w, "data: .env.example → .env kopyalandı ✅\n\n")
		fmt.Fprintf(w, "event: done\ndata: ok\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return
	default:
		http.Error(w, "bilinmeyen adım: "+stepID, 400)
		return
	}

	if !commandExistsInEnv(toolName, getEnhancedEnv()) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		fmt.Fprintf(w, "data: ❌ '%s' bulunamadı. Sistem & Teşhis sayfasından kurabilirsiniz.\n\n", toolName)
		fmt.Fprintf(w, "event: error\ndata: missing-tool:%s\n\n", toolName)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, canFlush := w.(http.Flusher)
	ctx := r.Context()

	sendEvent := func(line string) {
		line = strings.ReplaceAll(line, "\r", "")
		if line == "" {
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", line)
		if canFlush {
			flusher.Flush()
		}
	}

	sendEvent(fmt.Sprintf("▶ %s %s", toolName, strings.Join(args, " ")))
	sendEvent("---")

	cmd := enhancedCommand(ctx, toolName, args...)
	cmd.Dir = dir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		sendEvent("❌ Komut başlatılamadı: " + err.Error())
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		if canFlush {
			flusher.Flush()
		}
		return
	}

	// Stream stdout and stderr concurrently
	done := make(chan struct{}, 2)
	streamPipe := func(pipe io.Reader) {
		buf := make([]byte, 1)
		var line []byte
		for {
			n, err := pipe.Read(buf)
			if n > 0 {
				if buf[0] == '\n' {
					sendEvent(string(line))
					line = line[:0]
				} else {
					line = append(line, buf[0])
				}
			}
			if err != nil {
				if len(line) > 0 {
					sendEvent(string(line))
				}
				break
			}
		}
		done <- struct{}{}
	}

	go streamPipe(stdout)
	go streamPipe(stderr)

	<-done
	<-done

	err := cmd.Wait()
	if err != nil {
		if ctx.Err() != nil {
			sendEvent("⛔ Kullanıcı tarafından durduruldu.")
			fmt.Fprintf(w, "event: cancelled\ndata: cancelled\n\n")
		} else {
			sendEvent(fmt.Sprintf("❌ Hata: %v", err))
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		}
	} else {
		repoStartErrorsMux.Lock()
		delete(repoStartErrors, name)
		repoStartErrorsMux.Unlock()
		sendEvent("✅ Tamamlandı!")
		fmt.Fprintf(w, "event: done\ndata: ok\n\n")
	}
	if canFlush {
		flusher.Flush()
	}
}


// ---------- real skill install (npx skills add ...) ----------

// handleRepoEnableSkill copies a locally-cloned skill repo straight into every
// detected IDE's skills folder (currently just Claude Code CLI's ~/.claude/skills
// — the only one of the tracked IDEs that has a filesystem skills concept),
// instead of leaving "enable this skill" as a manual copy the user has to do.
func handleRepoEnableSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	copiedCount, results, err := enableSkillForRepo(name)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if copiedCount == 0 {
		for _, res := range results {
			if errStr, ok := res["error"].(string); ok {
				writeErr(w, 500, fmt.Sprintf("%s: %s", res["ide"], errStr))
				return
			}
		}
		writeErr(w, 424, "Skill kabul eden bir IDE tespit edilmedi (şu an sadece Claude Code CLI destekleniyor).")
		return
	}

	logActivity("skill-enable", fmt.Sprintf("%s -> %d IDE", name, copiedCount))
	writeJSON(w, 200, map[string]interface{}{
		"ok":          true,
		"copiedCount": copiedCount,
		"results":     results,
		"message":     fmt.Sprintf("✅ '%s' skill'i %d IDE'ye etkinleştirildi.", name, copiedCount),
	})
}

// handleSkillsSyncAll walks repo/ and enables every skill it finds into
// whatever IDEs are currently detected — the "sync all skills" bulk action.
func handleSkillsSyncAll(w http.ResponseWriter, r *http.Request) {
	rootDir := filepath.Join(getBaseDir(), "repo")
	dirEntries, err := os.ReadDir(rootDir)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}

	results := []map[string]interface{}{}
	skillsSynced := 0
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		name := de.Name()
		copiedCount, _, err := enableSkillForRepo(name)
		if err != nil || copiedCount == 0 {
			continue
		}
		skillsSynced++
		results = append(results, map[string]interface{}{"name": name, "copiedCount": copiedCount})
	}

	logActivity("skills-sync-all", fmt.Sprintf("%d skill -> IDE'lere", skillsSynced))
	writeJSON(w, 200, map[string]interface{}{
		"ok":           true,
		"skillsSynced": skillsSynced,
		"results":      results,
		"message":      fmt.Sprintf("✅ %d skill IDE'lere senkronize edildi.", skillsSynced),
	})
}

func handleSkillInstall(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Repo  string `json:"repo"`
		Extra string `json:"extra"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "gecersiz JSON")
		return
	}
	if body.Repo == "" {
		writeErr(w, 400, "repo alani gerekli")
		return
	}

	if !commandExistsInEnv("npx", getEnhancedEnv()) {
		writeErr(w, 424, "npx bulunamadi - kurulum icin Node.js gerekli (https://nodejs.org)")
		return
	}

	extra := strings.TrimSpace(body.Extra)
	if extra == "" {
		extra = "--all"
	}
	args := append([]string{"skills", "add", body.Repo}, strings.Fields(extra)...)
	args = append(args, "-g", "-y")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := enhancedCommand(ctx, "npx", args...)
	out, runErr := cmd.CombinedOutput()

	result := map[string]interface{}{
		"command": "npx " + strings.Join(args, " "),
		"output":  strings.TrimSpace(string(out)),
		"ok":      runErr == nil,
	}
	if runErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result["error"] = "kurulum zaman asimina ugradi (120sn)"
		} else {
			result["error"] = runErr.Error()
		}
		logActivity("skill-install-error", fmt.Sprintf("%s: %v", body.Repo, result["error"]))
	} else {
		logActivity("skill-install", body.Repo)
	}
	writeJSON(w, 200, result)
}

// ---------- repo/ library scan (auto-detect skill vs. MCP folders) ----------

func gitRemoteRepo(dir string) string {
	// git -C walks UP to find a .git if dir has none of its own — guard against
	// that returning the ai-toolkit workspace's own remote for a non-git folder.
	topOut, err := enhancedCommand(context.Background(), "git", "-C", dir, "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return ""
	}
	top := strings.TrimSpace(string(topOut))
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	if filepath.Clean(strings.ReplaceAll(top, "/", string(filepath.Separator))) != filepath.Clean(absDir) {
		return ""
	}

	out, err := enhancedCommand(context.Background(), "git", "-C", dir, "remote", "get-url", "origin").CombinedOutput()
	if err != nil {
		return ""
	}
	url := strings.TrimSpace(string(out))
	url = strings.TrimSuffix(url, ".git")
	// git@github.com:owner/name -> owner/name
	if idx := strings.Index(url, "github.com:"); idx != -1 {
		return url[idx+len("github.com:"):]
	}
	// https://github.com/owner/name -> owner/name
	if idx := strings.Index(url, "github.com/"); idx != -1 {
		return url[idx+len("github.com/"):]
	}
	return ""
}

// findSkillMd returns the path to this repo's SKILL.md, checking the root
// first, then the skills/<name>/ and plugins/<name>/skills/<name>/ bundle
// conventions — or "" if none of those shapes are present.
// findSkillMd looks for SKILL.md only in shapes that mean "this repo IS a
// distributable skill (bundle)": at the root, or nested under skills/,
// plugins/*/skills/, packages/*/skills/, cli/assets/skills/. It deliberately
// does NOT match .claude/skills/ (that's a repo's own internal Claude Code
// tooling, not something it ships to other users) and does NOT walk the full
// tree (a stray SKILL.md in docs/examples/vendored code anywhere in a large
// repo would otherwise mislabel an unrelated project as a skill package).
func findSkillMd(dir string) string {
	if p := filepath.Join(dir, "SKILL.md"); fileExists(p) {
		return p
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, "skills", "*", "SKILL.md")); len(matches) > 0 {
		return matches[0]
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, "plugins", "*", "skills", "*", "SKILL.md")); len(matches) > 0 {
		return matches[0]
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, "packages", "*", "skills", "*", "SKILL.md")); len(matches) > 0 {
		return matches[0]
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, "cli", "assets", "skills", "*", "SKILL.md")); len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func findAllSkillFolders(srcDir string) []string {
	if fileExists(filepath.Join(srcDir, "SKILL.md")) {
		return []string{srcDir}
	}
	var skillDirs []string
	seen := make(map[string]bool)
	_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == "node_modules" || base == "test" || base == "fixtures" {
				return filepath.SkipDir
			}
			if fileExists(filepath.Join(path, "SKILL.md")) {
				if !strings.Contains(path, "/fixtures/") && !strings.Contains(path, "/test/") && !strings.Contains(path, "/examples/") {
					sName := filepath.Base(path)
					if !seen[sName] {
						seen[sName] = true
						skillDirs = append(skillDirs, path)
					}
				}
				return filepath.SkipDir
			}
		}
		return nil
	})
	return skillDirs
}

// enableSkillForRepo copies each individual skill directory in repo/<name> directly
// into every detected IDE's skills folder (e.g. ~/.gemini/config/skills/<skill_name>/SKILL.md)
func enableSkillForRepo(name string) (int, []map[string]interface{}, error) {
	srcDir := repoDir(name)
	skillFolders := findAllSkillFolders(srcDir)
	if len(skillFolders) == 0 {
		return 0, nil, fmt.Errorf("bu klasörde SKILL.md bulunamadı")
	}

	results := []map[string]interface{}{}
	copiedCount := 0
	for _, ide := range detectIdes() {
		if !ide.Detected || ide.SkillsDir == "" {
			continue
		}
		if err := ensureValidDir(ide.SkillsDir); err != nil {
			results = append(results, map[string]interface{}{"ide": ide.Name, "error": err.Error()})
			continue
		}

		subCopied := 0
		for _, folder := range skillFolders {
			sName := filepath.Base(folder)
			destDir := filepath.Join(ide.SkillsDir, sName)
			if err := copyDirRecursive(folder, destDir); err != nil {
				results = append(results, map[string]interface{}{"ide": ide.Name, "error": fmt.Sprintf("%s: %v", sName, err)})
			} else {
				subCopied++
			}
		}
		if subCopied > 0 {
			copiedCount++
			results = append(results, map[string]interface{}{"ide": ide.Name, "skillsCount": subCopied, "ok": true})
		}
	}
	return copiedCount, results, nil
}

func syncSkillsToAllIdes() (int, []map[string]interface{}) {
	rootDir := filepath.Join(getBaseDir(), "repo")
	dirEntries, err := os.ReadDir(rootDir)
	if err != nil {
		return 0, nil
	}

	results := []map[string]interface{}{}
	skillsSynced := 0
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		name := de.Name()
		srcDir := repoDir(name)
		if findSkillMd(srcDir) == "" {
			continue
		}

		copiedCount, _, err := enableSkillForRepo(name)
		if err == nil && copiedCount > 0 {
			skillsSynced++
			results = append(results, map[string]interface{}{"name": name, "copiedCount": copiedCount})
		}
	}
	return skillsSynced, results
}

func pruneDeletedSkillsFromAllIdes() int {
	rootDir := filepath.Join(getBaseDir(), "repo")
	validSkills := make(map[string]bool)
	if dirEntries, err := os.ReadDir(rootDir); err == nil {
		for _, de := range dirEntries {
			if de.IsDir() {
				srcDir := repoDir(de.Name())
				if findSkillMd(srcDir) != "" {
					validSkills[de.Name()] = true
				}
			}
		}
	}

	prunedTotal := 0
	for _, ide := range detectIdes() {
		if ide.SkillsDir == "" {
			continue
		}
		entries, err := os.ReadDir(ide.SkillsDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				skillName := e.Name()
				if !validSkills[skillName] {
					target := filepath.Join(ide.SkillsDir, skillName)
					if err := os.RemoveAll(target); err == nil {
						prunedTotal++
						logActivity("skill-prune", fmt.Sprintf("Skill '%s' ide'den silindi: %s", skillName, ide.Name))
					}
				}
			}
		}
	}
	return prunedTotal
}

var (
	syncStateMux      sync.Mutex
	lastSyncStateHash string
	fullSyncCompleted bool
	lastSyncedMcpIds  map[string]bool
)

func computeSyncStateHash() string {
	var sb strings.Builder
	s, err := loadState()
	if err == nil && s != nil {
		sb.WriteString(fmt.Sprintf("mcp:%d;", len(s.McpServers)))
		for _, m := range s.McpServers {
			sb.WriteString(fmt.Sprintf("%s:%s;", m.ID, m.Command))
		}
	}
	rootDir := filepath.Join(getBaseDir(), "repo")
	if entries, err := os.ReadDir(rootDir); err == nil {
		sb.WriteString(fmt.Sprintf("repo:%d;", len(entries)))
		for _, e := range entries {
			if e.IsDir() {
				info, _ := e.Info()
				if info != nil {
					sb.WriteString(fmt.Sprintf("%s:%d;", e.Name(), info.ModTime().UnixNano()))
				}
			}
		}
	}
	return sb.String()
}

func runBackgroundSyncController() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			syncStateMux.Lock()
			currentHash := computeSyncStateHash()
			if currentHash == lastSyncStateHash && fullSyncCompleted {
				syncStateMux.Unlock()
				continue
			}

			s, err := loadState()
			if err == nil && s != nil {
				syncAllIdes(s.McpServers)

				// syncAllIdes only ever adds/updates keys — it never deletes
				// one, so an MCP entry removed from state.json outside the
				// normal HTTP flow (hand-edited, or by a future code path
				// that forgets to diff) would otherwise sit there forever,
				// resurrected on every tick. Diff against what this loop
				// itself last pushed and explicitly prune the difference.
				currentIds := map[string]bool{}
				for _, m := range s.McpServers {
					currentIds[m.ID] = true
				}
				if lastSyncedMcpIds != nil {
					var goneIds []string
					for id := range lastSyncedMcpIds {
						if !currentIds[id] {
							goneIds = append(goneIds, id)
						}
					}
					removeMcpIdsFromAllIdes(goneIds)
				}
				lastSyncedMcpIds = currentIds
			}

			_, _ = syncSkillsToAllIdes()
			_ = pruneDeletedSkillsFromAllIdes()

			lastSyncStateHash = currentHash
			fullSyncCompleted = true
			syncStateMux.Unlock()
		}
	}()
}

func extractSkillDesc(skillMdPath string) string {
	data, err := os.ReadFile(skillMdPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "description:") {
			return strings.TrimSpace(trimmed[len("description:"):])
		}
	}
	return ""
}

func scanRepoFolder(name string) ScanEntry {
	dir := repoDir(name)
	entry := ScanEntry{Name: name, Path: dir}

	// SKILL.md at the root covers a single-skill repo; multi-skill repos
	// (e.g. a plugin that bundles several skills) nest it under skills/<name>/
	// or plugins/<name>/skills/<name>/ instead — findSkillMd checks those
	// shapes too before giving up, so a bundle of skills doesn't get missed.
	if skillPath := findSkillMd(dir); skillPath != "" {
		entry.HasSkill = true
		entry.SkillDesc = extractSkillDesc(skillPath)
	}

	pkgPath := filepath.Join(dir, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		entry.HasPackageJson = true
		var pkg struct {
			Name            string            `json:"name"`
			Description     string            `json:"description"`
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			entry.PackageName = pkg.Name
			haystack := strings.ToLower(pkg.Name + " " + pkg.Description + " " + name)
			// A "bin" entry alone used to be treated as "looks like MCP", but
			// that's true of nearly every CLI tool (installers, linters,
			// scaffolders) — caveman-installer tripped this and got labeled
			// an MCP Server despite being a skill installer. Depending on the
			// actual MCP SDK is a real signal; a bin field isn't.
			_, hasMcpSdk := pkg.Dependencies["@modelcontextprotocol/sdk"]
			if !hasMcpSdk {
				_, hasMcpSdk = pkg.DevDependencies["@modelcontextprotocol/sdk"]
			}
			entry.LooksLikeMcp = hasMcpSdk || strings.Contains(haystack, "mcp")
		}
	}

	if !entry.LooksLikeMcp {
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "mcp") {
			entry.LooksLikeMcp = true
		} else if _, err := os.Stat(filepath.Join(dir, "mcp.json")); err == nil {
			entry.LooksLikeMcp = true
		} else if _, err := os.Stat(filepath.Join(dir, "mcp-config.json")); err == nil {
			entry.LooksLikeMcp = true
		}
	}

	if entry.HasPackageJson {
		entry.Runtimes = append(entry.Runtimes, "node")
	}
	pythonMarkers := []string{"requirements.txt", "pyproject.toml", "setup.py"}
	for _, marker := range pythonMarkers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			entry.Runtimes = append(entry.Runtimes, "python")
			break
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		entry.Runtimes = append(entry.Runtimes, "rust")
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		entry.Runtimes = append(entry.Runtimes, "go")
	}
	if _, err := os.Stat(filepath.Join(dir, "Gemfile")); err == nil {
		entry.Runtimes = append(entry.Runtimes, "ruby")
	}
	if _, err := os.Stat(filepath.Join(dir, "composer.json")); err == nil {
		entry.Runtimes = append(entry.Runtimes, "php")
	}
	for _, marker := range []string{"pom.xml", "build.gradle", "build.gradle.kts"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			entry.Runtimes = append(entry.Runtimes, "java")
			break
		}
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, "*.csproj")); len(matches) > 0 {
		entry.Runtimes = append(entry.Runtimes, "dotnet")
	} else if matches, _ := filepath.Glob(filepath.Join(dir, "*.sln")); len(matches) > 0 {
		entry.Runtimes = append(entry.Runtimes, "dotnet")
	}
	if _, err := os.Stat(filepath.Join(dir, "mix.exs")); err == nil {
		entry.Runtimes = append(entry.Runtimes, "elixir")
	}
	for _, marker := range []string{"deno.json", "deno.jsonc"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			entry.Runtimes = append(entry.Runtimes, "deno")
			break
		}
	}
	for _, marker := range []string{"Dockerfile", "dockerfile"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			entry.Runtimes = append(entry.Runtimes, "docker")
			break
		}
	}

	entry.Repo = gitRemoteRepo(dir)

	// Installation detection (node_modules, .venv, venv, target)
	if _, err := os.Stat(filepath.Join(dir, "node_modules")); err == nil {
		entry.IsInstalled = true
	}
	if _, err := os.Stat(filepath.Join(dir, ".venv")); err == nil {
		entry.IsInstalled = true
	}
	if _, err := os.Stat(filepath.Join(dir, "venv")); err == nil {
		entry.IsInstalled = true
	}
	if _, err := os.Stat(filepath.Join(dir, "target")); err == nil {
		entry.IsInstalled = true
	}

	// Start command detection
	if entry.HasPackageJson {
		if data, err := os.ReadFile(pkgPath); err == nil {
			var pkg struct {
				Main    string            `json:"main"`
				Scripts map[string]string `json:"scripts"`
				Bin     interface{}       `json:"bin"`
			}
			if json.Unmarshal(data, &pkg) == nil {
				pkgManager := "npm"
				if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
					pkgManager = "pnpm"
				} else if _, err := os.Stat(filepath.Join(dir, "pnpm-workspace.yaml")); err == nil {
					pkgManager = "pnpm"
				}

				if pkg.Scripts != nil {
					for scriptName := range pkg.Scripts {
						cmd := "npm run " + scriptName
						if pkgManager == "pnpm" {
							cmd = "pnpm run " + scriptName
						} else if scriptName == "start" {
							cmd = "npm start"
						}
						entry.AvailableCmds = append(entry.AvailableCmds, cmd)
					}

					// "build"/"run"/"cli"/"watch" used to be in this list, but
					// they're compile steps or too ambiguous to trust as "this
					// launches the thing" — a wrong guess here becomes a
					// silently-broken MCP entry someone else's IDE tries to run.
					priorityScripts := []string{"dev", "start", "serve", "preview"}
					for _, sName := range priorityScripts {
						if _, ok := pkg.Scripts[sName]; ok {
							if pkgManager == "pnpm" {
								entry.StartCommand = "pnpm run " + sName
							} else {
								if sName == "start" {
									entry.StartCommand = "npm start"
								} else {
									entry.StartCommand = "npm run " + sName
								}
							}
							break
						}
					}

					if entry.StartCommand == "" && len(pkg.Scripts) > 0 {
						// No known launch-script name matched — fall back to
						// *some* script rather than nothing, but pick it
						// deterministically (Go map iteration order is
						// randomized per-process) and flag it as a guess so
						// the UI doesn't present it with false confidence.
						scriptNames := make([]string, 0, len(pkg.Scripts))
						for sName := range pkg.Scripts {
							scriptNames = append(scriptNames, sName)
						}
						sort.Strings(scriptNames)
						sName := scriptNames[0]
						if pkgManager == "pnpm" {
							entry.StartCommand = "pnpm run " + sName
						} else {
							entry.StartCommand = "npm run " + sName
						}
						entry.StartCommandGuessed = true
					}
				}

				if entry.StartCommand == "" && pkg.Main != "" {
					entry.StartCommand = "node " + pkg.Main
					entry.AvailableCmds = append(entry.AvailableCmds, entry.StartCommand)
				}
			}
		}

		if entry.StartCommand == "" {
			for _, mainFile := range []string{"index.js", "server.js", "app.js", "main.js", "src/index.js", "dist/index.js", "bin/index.js"} {
				if _, err := os.Stat(filepath.Join(dir, mainFile)); err == nil {
					entry.StartCommand = "node " + mainFile
					entry.AvailableCmds = append(entry.AvailableCmds, entry.StartCommand)
					break
				}
			}
		}

		if entry.StartCommand == "" {
			entry.StartCommand = "npm start"
			entry.AvailableCmds = append(entry.AvailableCmds, "npm start")
		}
	}

	if containsRuntime(entry.Runtimes, "python") {
		for _, pyFile := range []string{"main.py", "app.py", "server.py", "run.py", "manage.py", "cli.py", "__main__.py", "src/main.py"} {
			if _, err := os.Stat(filepath.Join(dir, pyFile)); err == nil {
				cmdStr := "python " + pyFile
				entry.AvailableCmds = append(entry.AvailableCmds, cmdStr)
				if entry.StartCommand == "" {
					entry.StartCommand = cmdStr
				}
			}
		}
		if entry.StartCommand == "" {
			entry.StartCommand = "python main.py"
		}
	}
	if containsRuntime(entry.Runtimes, "rust") {
		entry.AvailableCmds = append(entry.AvailableCmds, "cargo run")
		if entry.StartCommand == "" {
			entry.StartCommand = "cargo run"
		}
	}
	if containsRuntime(entry.Runtimes, "docker") {
		entry.AvailableCmds = append(entry.AvailableCmds, "docker run")
		if entry.StartCommand == "" {
			entry.StartCommand = "docker run"
		}
	}

	// Local MCP command derivation: prefer a direct absolute-path invocation
	// (node/python3 <file>) over StartCommand's npm/pnpm wrapper, since an
	// external MCP client launches this process itself — a relative script
	// path or bare npm-script name only works by luck if the client's own
	// cwd happens to match repo/<name>.
	if entry.StartCommand != "" && entry.StartCommand != "docker run" {
		parts := strings.Fields(entry.StartCommand)
		if len(parts) >= 2 && (parts[0] == "node" || parts[0] == "python" || parts[0] == "python3") {
			scriptPath := parts[1]
			if !filepath.IsAbs(scriptPath) {
				scriptPath = filepath.Join(dir, scriptPath)
			}
			entry.LocalCommand = parts[0]
			if entry.LocalCommand == "python" {
				entry.LocalCommand = "python3"
			}
			entry.LocalArgs = append([]string{scriptPath}, parts[2:]...)
		} else if len(parts) >= 1 {
			entry.LocalCommand = parts[0]
			entry.LocalArgs = parts[1:]
		}
	}

	// RunMode + MissingTool: pick the one primary way to run this repo instead
	// of showing every detected runtime as an equal option, and cross-check
	// against what's actually installed system-wide (not just in this repo
	// folder) so the UI can point at Sistem & Teşhis instead of a doomed retry.
	enhancedEnv := getEnhancedEnv()
	switch {
	case entry.HasPackageJson:
		entry.RunMode = "local"
		if !commandExistsInEnv("node", enhancedEnv) {
			entry.MissingTool = "node"
		}
	case containsRuntime(entry.Runtimes, "python"):
		entry.RunMode = "local"
		if !commandExistsInEnv("python3", enhancedEnv) && !commandExistsInEnv("python", enhancedEnv) {
			entry.MissingTool = "python"
		}
	case containsRuntime(entry.Runtimes, "rust"):
		entry.RunMode = "local"
		if !commandExistsInEnv("cargo", enhancedEnv) {
			entry.MissingTool = "cargo"
		}
	case containsRuntime(entry.Runtimes, "docker"):
		entry.RunMode = "docker"
		if !commandExistsInEnv("docker", enhancedEnv) {
			entry.MissingTool = "docker"
		}
	case containsRuntime(entry.Runtimes, "go"):
		entry.RunMode = "local"
		if !commandExistsInEnv("go", enhancedEnv) {
			entry.MissingTool = "go"
		}
	case containsRuntime(entry.Runtimes, "ruby"):
		entry.RunMode = "local"
		if !commandExistsInEnv("ruby", enhancedEnv) {
			entry.MissingTool = "ruby"
		}
	case containsRuntime(entry.Runtimes, "php"):
		entry.RunMode = "local"
		if !commandExistsInEnv("php", enhancedEnv) {
			entry.MissingTool = "php"
		}
	case containsRuntime(entry.Runtimes, "java"):
		entry.RunMode = "local"
		if !commandExistsInEnv("java", enhancedEnv) {
			entry.MissingTool = "java"
		}
	case containsRuntime(entry.Runtimes, "dotnet"):
		entry.RunMode = "local"
		if !commandExistsInEnv("dotnet", enhancedEnv) {
			entry.MissingTool = "dotnet"
		}
	case containsRuntime(entry.Runtimes, "elixir"):
		entry.RunMode = "local"
		if !commandExistsInEnv("elixir", enhancedEnv) {
			entry.MissingTool = "elixir"
		}
	case containsRuntime(entry.Runtimes, "deno"):
		entry.RunMode = "local"
		if !commandExistsInEnv("deno", enhancedEnv) {
			entry.MissingTool = "deno"
		}
	}

	repoStartErrorsMux.RLock()
	if errStr, ok := repoStartErrors[name]; ok {
		entry.HasStartError = true
		entry.StartErrorMsg = errStr
	}
	repoStartErrorsMux.RUnlock()

	runningAppsMux.RLock()
	if app, ok := runningAppsMap[name]; ok {
		entry.IsRunning = true
		entry.RunningPort = app.Port
		entry.RunningPid = app.PID
	}
	runningAppsMux.RUnlock()

	// Auto detect RepoType
	if entry.LooksLikeMcp {
		entry.RepoType = "mcp"
		entry.RepoTypeLabel = "🔌 MCP Server"
	} else if entry.HasSkill {
		entry.RepoType = "skill"
		entry.RepoTypeLabel = "📄 Skill"
	} else if entry.IsInstalled || entry.StartCommand != "" || containsRuntime(entry.Runtimes, "docker") {
		entry.RepoType = "service"
		entry.RepoTypeLabel = "⚡ Service / App"
	} else if entry.HasPackageJson || len(entry.Runtimes) > 0 {
		entry.RepoType = "library"
		entry.RepoTypeLabel = "📦 Library"
	} else {
		entry.RepoType = "other"
		entry.RepoTypeLabel = "📁 Diğer"
	}

	// A skill/plugin repo (e.g. a bundle of agent skills with an installer
	// CLI) isn't "started" as a local service — its real action is enabling
	// the skill into an IDE, not npm-installing and running whatever script
	// happened to be its only one (often "test"). Drop the run-mode fields so
	// the UI doesn't offer a Kur & Başlat button that doesn't mean anything here.
	if entry.RepoType == "skill" {
		entry.RunMode = ""
		entry.MissingTool = ""
		entry.LocalCommand = ""
		entry.LocalArgs = nil
	}

	return entry
}

func containsRuntime(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// envVarPatterns cover the common "read an env var" idioms across the
// languages this app already recognizes as a runtime (node/python/go/rust/
// java/ruby/php) — language-agnostic by design, so a repo in any of these
// stacks gets its required config discovered instead of the user having to
// guess variable names from a runtime crash message.
var envVarPatterns = []*regexp.Regexp{
	regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`),
	regexp.MustCompile(`process\.env\[['"]([A-Z_][A-Z0-9_]*)['"]\]`),
	regexp.MustCompile(`os\.environ\[['"]([A-Z_][A-Z0-9_]*)['"]\]`),
	regexp.MustCompile(`os\.environ\.get\(['"]([A-Z_][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.getenv\(['"]([A-Z_][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.Getenv\(['"]([A-Z_][A-Z0-9_]*)['"]\)`),
	regexp.MustCompile(`os\.LookupEnv\(['"]([A-Z_][A-Z0-9_]*)['"]\)`),
	regexp.MustCompile(`env::var\(['"]([A-Z_][A-Z0-9_]*)['"]\)`),
	regexp.MustCompile(`System\.getenv\(['"]([A-Z_][A-Z0-9_]*)['"]\)`),
	regexp.MustCompile(`ENV\[['"]([A-Z_][A-Z0-9_]*)['"]\]`),
	regexp.MustCompile(`ENV\.fetch\(['"]([A-Z_][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`getenv\(['"]([A-Z_][A-Z0-9_]*)['"]\)`),
	regexp.MustCompile(`\$_ENV\[['"]([A-Z_][A-Z0-9_]*)['"]\]`),
}

var envVarScanExtensions = map[string]bool{
	".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true,
	".py": true, ".go": true, ".rs": true, ".java": true, ".kt": true,
	".rb": true, ".php": true, ".cs": true, ".ex": true, ".exs": true,
}

var envVarSkipDirs = map[string]bool{
	".git": true, "node_modules": true, "dist": true, "build": true,
	".venv": true, "venv": true, "target": true, "vendor": true,
	".next": true, "out": true, "bin": true, "obj": true,
}

// Common vars that show up in almost every codebase but aren't something a
// user needs to configure by hand for an MCP server.
var envVarNoise = map[string]bool{
	"NODE_ENV": true, "PATH": true, "HOME": true, "PWD": true, "SHELL": true,
	"TERM": true, "LANG": true, "USER": true, "USERPROFILE": true,
	"TEMP": true, "TMP": true, "TMPDIR": true, "CI": true,
}

// detectEnvVarNames scans a repo's source for env-var read patterns instead
// of making the user guess from a runtime crash message. Bounded by file
// count/size so it stays fast even on a large monorepo.
func detectEnvVarNames(dir string) []string {
	found := map[string]bool{}
	filesScanned := 0
	const maxFiles = 3000
	const maxFileSize = 256 * 1024

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if filesScanned >= maxFiles {
			return filepath.SkipAll
		}
		if info.IsDir() {
			if envVarSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !envVarScanExtensions[strings.ToLower(filepath.Ext(info.Name()))] {
			return nil
		}
		if info.Size() > maxFileSize {
			return nil
		}
		data, rErr := os.ReadFile(path)
		if rErr != nil {
			return nil
		}
		filesScanned++
		content := string(data)
		for _, re := range envVarPatterns {
			for _, m := range re.FindAllStringSubmatch(content, -1) {
				if len(m) > 1 && !envVarNoise[m[1]] {
					found[m[1]] = true
				}
			}
		}
		return nil
	})

	names := make([]string, 0, len(found))
	for k := range found {
		names = append(names, k)
	}
	sort.Strings(names)
	if len(names) > 40 {
		names = names[:40]
	}
	return names
}

func handleRepoEnvVars(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		writeErr(w, 404, "yerel klasor yok")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"vars": detectEnvVarNames(dir)})
}

func handleRepoStart(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := safeRepoName(name); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		writeErr(w, 404, "yerel klasor yok")
		return
	}

	var body struct {
		Command string `json:"command"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	entry := scanRepoFolder(name)
	startCmdStr := body.Command
	if startCmdStr == "" {
		startCmdStr = r.URL.Query().Get("cmd")
	}
	if startCmdStr == "" {
		startCmdStr = entry.StartCommand
	}
	if startCmdStr == "" {
		if entry.HasPackageJson {
			startCmdStr = "npm start"
		} else if containsRuntime(entry.Runtimes, "python") {
			startCmdStr = "python main.py"
		} else {
			startCmdStr = "go run ."
		}
	}

	task := addTask("start", name, fmt.Sprintf("%s projesi başlatılıyor (%s)...", name, startCmdStr))
	startCtx, startCancel := context.WithCancel(context.Background())
	task.attachCancel(startCancel)

	parts := strings.Fields(startCmdStr)
	cmd := enhancedCommand(startCtx, parts[0], parts[1:]...)
	cmd.Dir = dir

	logActivity("repo-start", fmt.Sprintf("%s (%s)", name, startCmdStr))

	go func() {
		defer startCancel()
		time.Sleep(500 * time.Millisecond)
		if cmd.Process != nil {
			pid := cmd.Process.Pid
			port := detectPortByPID(pid)
			appObj := &RunningApp{
				ID:        name,
				Name:      name,
				Command:   startCmdStr,
				PID:       pid,
				Port:      port,
				Status:    "running",
				StartedAt: time.Now().Format(time.RFC3339),
			}
			runningAppsMux.Lock()
			runningAppsMap[name] = appObj
			runningAppsMux.Unlock()
		}

		out, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(out))

		runningAppsMux.Lock()
		delete(runningAppsMap, name)
		runningAppsMux.Unlock()

		repoStartErrorsMux.Lock()
		if err != nil && startCtx.Err() == context.Canceled {
			delete(repoStartErrors, name)
			finishTask(task, "cancelled", fmt.Sprintf("%s durduruldu.", name))
		} else if err != nil {
			repoStartErrors[name] = outputStr
			finishTask(task, "error", fmt.Sprintf("%s hatası: %s", name, outputStr))
		} else {
			delete(repoStartErrors, name)
			finishTask(task, "completed", fmt.Sprintf("%s başarıyla çalıştırıldı.", name))
		}
		repoStartErrorsMux.Unlock()
	}()

	writeJSON(w, 200, map[string]interface{}{
		"ok":      true,
		"name":    name,
		"command": startCmdStr,
		"message": fmt.Sprintf("🚀 %s projesi '%s' komutuyla başlatıldı.", name, startCmdStr),
	})
}

func handleReposScan(w http.ResponseWriter, r *http.Request) {
	rootDir := filepath.Join(getBaseDir(), "repo")
	dirEntries, err := os.ReadDir(rootDir)
	if err != nil {
		writeJSON(w, 200, []ScanEntry{})
		return
	}

	s, _ := loadState()
	trackedMap := map[string]string{}
	if s != nil {
		for _, t := range s.TrackedRepos {
			trackedMap[t.Name] = t.Status
		}
	}

	results := []ScanEntry{}
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		entry := scanRepoFolder(de.Name())
		if status, ok := trackedMap[de.Name()]; ok {
			entry.GitStatus = status
		}
		results = append(results, entry)
	}
	writeJSON(w, 200, results)
}

func handleCheckAllRepos(w http.ResponseWriter, r *http.Request) {
	task := addTask("check-all", "Tüm Repolar", "Tüm yerel repoların GitHub güncellik durumları kontrol ediliyor...")
	ctx, cancel := context.WithCancel(r.Context())
	task.attachCancel(cancel)
	defer cancel()

	rootDir := filepath.Join(getBaseDir(), "repo")
	dirEntries, err := os.ReadDir(rootDir)
	if err != nil {
		finishTask(task, "error", err.Error())
		writeErr(w, 500, err.Error())
		return
	}

	s, _ := loadState()
	if s == nil {
		s = &StateBundle{}
	}

	trackedMap := map[string]*TrackedRepo{}
	for i := range s.TrackedRepos {
		trackedMap[s.TrackedRepos[i].Name] = &s.TrackedRepos[i]
	}

	type CheckResult struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}

	results := []CheckResult{}
	behindCount := 0

	stopped := false
	for _, de := range dirEntries {
		if ctx.Err() != nil {
			stopped = true
			break
		}
		if !de.IsDir() {
			continue
		}
		name := de.Name()
		status, checkErr := gitCheck(ctx, name)
		resItem := CheckResult{Name: name, Status: status}
		if checkErr != nil {
			resItem.Error = checkErr.Error()
		}
		if strings.HasPrefix(status, "behind") {
			behindCount++
		}
		if t, ok := trackedMap[name]; ok {
			t.Status = status
			t.LastChecked = time.Now().Format(time.RFC3339)
		} else {
			dir := repoDir(name)
			s.TrackedRepos = append(s.TrackedRepos, TrackedRepo{
				Name:        name,
				Repo:        gitRemoteRepo(dir),
				LocalPath:   dir,
				Status:      status,
				LastChecked: time.Now().Format(time.RFC3339),
			})
		}
		results = append(results, resItem)
	}

	_ = saveState(s)

	if stopped {
		finishTask(task, "cancelled", fmt.Sprintf("Kontrol durduruldu — %d repo tamamlandı.", len(results)))
	} else {
		msg := fmt.Sprintf("%d repo kontrol edildi.", len(results))
		if behindCount > 0 {
			msg += fmt.Sprintf(" %d repo için güncelleme mevcut!", behindCount)
		}
		finishTask(task, "completed", msg)
	}

	writeJSON(w, 200, map[string]interface{}{
		"ok":          true,
		"stopped":     stopped,
		"behindCount": behindCount,
		"results":     results,
	})
}

// ---------- GitHub live search proxy (server-side, no API key) ----------

type GithubRepo struct {
	FullName        string `json:"full_name"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	StargazersCount int    `json:"stargazers_count"`
	HTMLURL         string `json:"html_url"`
	Owner           struct {
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
}

type GithubRateLimitInfo struct {
	Limit        int   `json:"limit"`
	Remaining    int   `json:"remaining"`
	Used         int   `json:"used"`
	ResetUnix    int64 `json:"resetUnix"`
	ResetSeconds int64 `json:"resetSeconds"`
}

func parseRateLimitInfo(resp *http.Response) GithubRateLimitInfo {
	info := GithubRateLimitInfo{
		Limit:        10,
		Remaining:    0,
		Used:         10,
		ResetUnix:    time.Now().Unix() + 60,
		ResetSeconds: 60,
	}
	if resp == nil {
		return info
	}
	if v := resp.Header.Get("X-RateLimit-Limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			info.Limit = n
		}
	}
	if v := resp.Header.Get("X-RateLimit-Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			info.Remaining = n
		}
	}
	if v := resp.Header.Get("X-RateLimit-Used"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			info.Used = n
		}
	}
	if v := resp.Header.Get("X-RateLimit-Reset"); v != "" {
		if unix, err := strconv.ParseInt(v, 10, 64); err == nil {
			info.ResetUnix = unix
			diff := unix - time.Now().Unix()
			if diff > 0 {
				info.ResetSeconds = diff
			} else {
				info.ResetSeconds = 0
			}
		}
	}
	return info
}

var (
	discoverCache     []GithubRepo
	discoverCacheTime time.Time
	discoverCacheMux  sync.Mutex
)

func handleGithubSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	if q == "" {
		discoverCacheMux.Lock()
		if len(discoverCache) > 0 && time.Since(discoverCacheTime) < 5*time.Minute {
			items := discoverCache
			discoverCacheMux.Unlock()
			writeJSON(w, 200, map[string]interface{}{
				"items": items,
				"rateLimit": GithubRateLimitInfo{
					Limit: 10, Remaining: 10, Used: 0, ResetUnix: time.Now().Unix(), ResetSeconds: 0,
				},
			})
			return
		}
		discoverCacheMux.Unlock()

		q = "topic:mcp-server OR topic:claude-skill OR topic:agent-skill"
	}

	reqURL := "https://api.github.com/search/repositories?q=" + url.QueryEscape(q) + "&sort=stars&order=desc&per_page=20"
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ai-toolkit-hub")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	defer resp.Body.Close()

	rateInfo := parseRateLimitInfo(resp)

	if resp.StatusCode == 403 || resp.StatusCode == 429 {
		writeJSON(w, 429, map[string]interface{}{
			"error":     "rate_limited",
			"message":   "GitHub anlık arama limiti doldu.",
			"rateLimit": rateInfo,
		})
		return
	}

	if resp.StatusCode != 200 {
		writeErr(w, resp.StatusCode, fmt.Sprintf("github api hatasi: %d", resp.StatusCode))
		return
	}

	var result struct {
		Items []GithubRepo `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		writeErr(w, 500, err.Error())
		return
	}

	if r.URL.Query().Get("q") == "" {
		discoverCacheMux.Lock()
		discoverCache = result.Items
		discoverCacheTime = time.Now()
		discoverCacheMux.Unlock()
	}

	writeJSON(w, 200, map[string]interface{}{
		"items":     result.Items,
		"rateLimit": rateInfo,
	})
}

type RuntimeCheck struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Cmd         string `json:"cmd"`
	Category    string `json:"category"`
	Desc        string `json:"desc"`
	InstallHint string `json:"installHint"`
	Installed   bool   `json:"installed"`
	Version     string `json:"version,omitempty"`
}

func getEnhancedEnv() []string {
	env := os.Environ()
	homeDir, _ := os.UserHomeDir()
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")

	extraPaths := []string{
		filepath.Join(appData, "npm"),
		filepath.Join(localAppData, "Microsoft", "WindowsApps"),
		filepath.Join(localAppData, "Programs", "Python"),
		filepath.Join(homeDir, ".cargo", "bin"),
		filepath.Join(homeDir, ".local", "bin"),
		filepath.Join(homeDir, ".orbstack", "bin"),
		filepath.Join(homeDir, ".docker", "bin"),
		filepath.Join(localAppData, "Programs", "Python", "Python312", "Scripts"),
		filepath.Join(localAppData, "Programs", "Python", "Python311", "Scripts"),
		filepath.Join(localAppData, "Programs", "Python", "Python310", "Scripts"),
		filepath.Join(localAppData, "Programs", "Python", "Scripts"),
		`C:\Program Files\Git\cmd`,
		`C:\Program Files\nodejs`,
		`C:\Program Files\Docker\Docker\resources\bin`,
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
		"/Applications/Docker.app/Contents/Resources/bin",
	}

	if pyBins, _ := filepath.Glob(filepath.Join(homeDir, "Library", "Python", "*", "bin")); len(pyBins) > 0 {
		extraPaths = append(extraPaths, pyBins...)
	}
	if brewBins, _ := filepath.Glob("/opt/homebrew/Cellar/*/*/bin"); len(brewBins) > 0 {
		extraPaths = append(extraPaths, brewBins...)
	}
	if usrBrewBins, _ := filepath.Glob("/usr/local/Cellar/*/*/bin"); len(usrBrewBins) > 0 {
		extraPaths = append(extraPaths, usrBrewBins...)
	}

	pathEnv := os.Getenv("PATH")
	for _, p := range extraPaths {
		if p != "" && !strings.Contains(strings.ToLower(pathEnv), strings.ToLower(p)) {
			pathEnv = p + string(os.PathListSeparator) + pathEnv
		}
	}

	newEnv := []string{}
	for _, e := range env {
		if !strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			newEnv = append(newEnv, e)
		}
	}
	newEnv = append(newEnv, "PATH="+pathEnv)
	return newEnv
}

func resolveInEnv(name string, env []string) string {
	if strings.ContainsRune(name, filepath.Separator) {
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			return name
		}
		return ""
	}

	exts := []string{""}
	if runtime.GOOS == "windows" {
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".exe") && !strings.HasSuffix(lower, ".cmd") && !strings.HasSuffix(lower, ".bat") {
			exts = []string{"", ".exe", ".cmd", ".bat"}
		}
	}

	for _, e := range env {
		var dirs string
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			dirs = e[5:]
		} else {
			continue
		}
		for _, dir := range strings.Split(dirs, string(os.PathListSeparator)) {
			if dir == "" {
				continue
			}
			for _, ext := range exts {
				candidate := filepath.Join(dir, name+ext)
				if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
					return candidate
				}
			}
		}
	}
	return ""
}

// commandExistsInEnv reports whether name can be resolved within env's PATH.
func commandExistsInEnv(name string, env []string) bool {
	return resolveInEnv(name, env) != ""
}

// enhancedCommand builds an exec.Cmd for name resolved against the
// Homebrew/Windows-aware PATH from getEnhancedEnv (see resolveInEnv for why
// this is necessary), with that same PATH passed through to the child.
func enhancedCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	env := getEnhancedEnv()
	resolved := resolveInEnv(name, env)
	if resolved == "" {
		resolved = name
	}
	cmd := exec.CommandContext(ctx, resolved, args...)
	cmd.Env = env
	return cmd
}

func checkToolVersion(candidates ...[]string) (bool, string) {
	for _, cand := range candidates {
		if len(cand) == 0 {
			continue
		}
		cmd := enhancedCommand(context.Background(), cand[0], cand[1:]...)

		out, err := cmd.CombinedOutput()
		if err == nil {
			ver := strings.TrimSpace(string(out))
			lines := strings.Split(ver, "\n")
			if len(lines) > 0 {
				ver = strings.TrimSpace(lines[0])
			}
			if ver != "" && !strings.Contains(strings.ToLower(ver), "is not recognized") {
				return true, ver
			}
		}
	}
	return false, ""
}

func handleSystemHealth(w http.ResponseWriter, r *http.Request) {
	type ToolDef struct {
		ID, Name, Category, Desc, InstallHint string
		Candidates                            [][]string
	}

	tools := []ToolDef{
		{"git", "Git Sürüm Kontrolü", "Core", "Repo klonlama, takibi ve sürüm kontrolü", "https://git-scm.com", [][]string{
			{"git", "--version"},
			{`C:\Program Files\Git\cmd\git.exe`, "--version"},
		}},
		{"node", "Node.js Motoru", "JavaScript", "JS/TS MCP sunucuları ve betik çalıştırma motoru", "https://nodejs.org", [][]string{
			{"node", "--version"},
			{`C:\Program Files\nodejs\node.exe`, "--version"},
		}},
		{"npx", "npx Paket Çalıştırıcı", "JavaScript", "Global npx paketleri ve skills add komutları", "npm i -g npm", [][]string{
			{"npx", "--version"},
			{"npx.cmd", "--version"},
		}},
		{"pnpm", "pnpm Paket Yöneticisi", "JavaScript", "Monorepo projeleri için paket yönetimi", "npm i -g pnpm", [][]string{
			{"pnpm", "--version"},
			{"pnpm.cmd", "--version"},
			{"npx", "pnpm", "--version"},
		}},
		{"python", "Python 3 Motoru", "Python", "Python tabanlı MCP sunucuları ve scriptleri", "https://python.org", [][]string{
			{"python", "--version"},
			{"python3", "--version"},
			{"py", "-3", "--version"},
		}},
		{"pipx", "pipx İzolasyon Aracı", "Python", "Google Analytics vb. pipx MCP sunucuları çalıştırma", "pip install pipx", [][]string{
			{"pipx", "--version"},
			{"python", "-m", "pipx", "--version"},
			{"py", "-m", "pipx", "--version"},
		}},
		{"uvx", "uv / uvx Hızlı Çalıştırıcı", "Python", "Yahoo Finance vb. ultra hızlı uvx MCP sunucuları", "pip install uv / winget install astral-sh.uv", [][]string{
			{"uvx", "--version"},
			{"uv", "--version"},
			{"python", "-m", "uv", "--version"},
		}},
		{"cargo", "Rust / Cargo Derleyici", "Rust", "Rust projeleri ve hızlı MCP sunucusu derleme", "https://rustup.rs", [][]string{
			{"cargo", "--version"},
			{"rustc", "--version"},
		}},
		{"docker", "Docker Engine & CLI", "Containers", "Konteynırlı MCP sunucuları ve Docker Build/Run", "https://docker.com", [][]string{
			{"docker", "--version"},
		}},
		{"go", "Go Derleyici", "Go", "Go tabanlı MCP sunucuları ve servisleri", "https://go.dev/dl/", [][]string{
			{"go", "version"},
		}},
		{"ruby", "Ruby Yorumlayıcısı", "Ruby", "Ruby tabanlı MCP sunucuları ve scriptleri", "https://www.ruby-lang.org/en/downloads/", [][]string{
			{"ruby", "--version"},
		}},
		{"php", "PHP Yorumlayıcısı", "PHP", "PHP tabanlı MCP sunucuları ve scriptleri", "https://www.php.net/downloads", [][]string{
			{"php", "--version"},
		}},
		{"java", "Java (JDK)", "Java", "Java/Kotlin tabanlı MCP sunucuları (Maven/Gradle)", "https://adoptium.net", [][]string{
			{"java", "--version"},
			{"java", "-version"},
		}},
		{"dotnet", ".NET SDK", ".NET", "C#/.NET tabanlı MCP sunucuları", "https://dotnet.microsoft.com/download", [][]string{
			{"dotnet", "--version"},
		}},
		{"elixir", "Elixir Derleyici", "Elixir", "Elixir/Mix tabanlı MCP sunucuları", "https://elixir-lang.org/install.html", [][]string{
			{"elixir", "--version"},
		}},
		{"deno", "Deno Runtime", "Deno", "Deno tabanlı MCP sunucuları ve scriptleri", "https://deno.com", [][]string{
			{"deno", "--version"},
		}},
	}

	results := []RuntimeCheck{}
	installedCount := 0

	for _, t := range tools {
		ok, ver := checkToolVersion(t.Candidates...)
		if ok {
			installedCount++
		}
		results = append(results, RuntimeCheck{
			ID:          t.ID,
			Name:        t.Name,
			Cmd:         t.Candidates[0][0],
			Category:    t.Category,
			Desc:        t.Desc,
			InstallHint: t.InstallHint,
			Installed:   ok,
			Version:     ver,
		})
	}

	writeJSON(w, 200, map[string]interface{}{
		"total":     len(results),
		"installed": installedCount,
		"missing":   len(results) - installedCount,
		"runtimes":  results,
	})
}

type installCmdDef struct {
	Name string
	Cmd  string
	Args []string
}

// buildInstallCmds returns the 1-click install command table for the current
// OS. Windows (winget) and macOS (Homebrew) are first-class; Linux falls back
// to apt-get on a best-effort basis (it typically needs sudo, so it will
// often just surface a clear permission error rather than silently working).
func buildInstallCmds() map[string]installCmdDef {
	switch runtime.GOOS {
	case "darwin":
		return map[string]installCmdDef{
			"pipx":   {"pipx", "python3", []string{"-m", "pip", "install", "--user", "pipx", "--break-system-packages"}},
			"uvx":    {"uv / uvx", "python3", []string{"-m", "pip", "install", "uv", "--break-system-packages"}},
			"pnpm":   {"pnpm", "npm", []string{"install", "-g", "pnpm"}},
			"npx":    {"npm/npx", "npm", []string{"install", "-g", "npm"}},
			"git":    {"Git", "brew", []string{"install", "git"}},
			"node":   {"Node.js", "brew", []string{"install", "node"}},
			"python": {"Python 3", "brew", []string{"install", "python@3.12"}},
			"cargo":  {"Rust / Cargo", "brew", []string{"install", "rust"}},
			"docker": {"Docker Desktop", "brew", []string{"install", "--cask", "docker"}},
		}
	case "windows":
		return map[string]installCmdDef{
			"pipx":   {"pipx", "python", []string{"-m", "pip", "install", "--user", "pipx"}},
			"uvx":    {"uv / uvx", "python", []string{"-m", "pip", "install", "uv"}},
			"pnpm":   {"pnpm", "npm", []string{"install", "-g", "pnpm"}},
			"npx":    {"npm/npx", "npm", []string{"install", "-g", "npm"}},
			"git":    {"Git", "winget", []string{"install", "--id", "Git.Git", "-e", "--source", "winget", "--accept-package-agreements", "--accept-source-agreements"}},
			"node":   {"Node.js", "winget", []string{"install", "--id", "OpenJS.NodeJS.LTS", "-e", "--source", "winget", "--accept-package-agreements", "--accept-source-agreements"}},
			"python": {"Python 3", "winget", []string{"install", "--id", "Python.Python.3.12", "-e", "--source", "winget", "--accept-package-agreements", "--accept-source-agreements"}},
			"cargo":  {"Rustup", "winget", []string{"install", "--id", "Rustlang.Rustup", "-e", "--source", "winget", "--accept-package-agreements", "--accept-source-agreements"}},
			"docker": {"Docker Desktop", "winget", []string{"install", "--id", "Docker.DockerDesktop", "-e", "--source", "winget", "--accept-package-agreements", "--accept-source-agreements"}},
		}
	default: // linux (best-effort)
		return map[string]installCmdDef{
			"pipx":   {"pipx", "python3", []string{"-m", "pip", "install", "--user", "pipx"}},
			"uvx":    {"uv / uvx", "python3", []string{"-m", "pip", "install", "uv"}},
			"pnpm":   {"pnpm", "npm", []string{"install", "-g", "pnpm"}},
			"npx":    {"npm/npx", "npm", []string{"install", "-g", "npm"}},
			"git":    {"Git", "apt-get", []string{"install", "-y", "git"}},
			"node":   {"Node.js", "apt-get", []string{"install", "-y", "nodejs", "npm"}},
			"python": {"Python 3", "apt-get", []string{"install", "-y", "python3", "python3-pip"}},
			"cargo":  {"Rust / Cargo", "apt-get", []string{"install", "-y", "cargo"}},
			"docker": {"Docker Engine", "apt-get", []string{"install", "-y", "docker.io"}},
		}
	}
}

func handleSystemInstall(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		writeErr(w, 400, "geçersiz istek id")
		return
	}

	toolInfo, ok := buildInstallCmds()[body.ID]
	if !ok {
		writeErr(w, 400, "otomatik kurulum komutu bulunamadı")
		return
	}

	if toolInfo.Cmd == "brew" && !commandExistsInEnv("brew", getEnhancedEnv()) {
		writeErr(w, 424, "Homebrew kurulu değil. Önce https://brew.sh üzerinden Homebrew'i kurun, ardından bu aracı tekrar deneyin.")
		return
	}

	cmdStr := toolInfo.Cmd + " " + strings.Join(toolInfo.Args, " ")
	task := addTask("system-install", toolInfo.Name, fmt.Sprintf("%s otomatik kuruluyor (%s)...", toolInfo.Name, cmdStr))
	installCtx, installCancel := context.WithCancel(context.Background())
	task.attachCancel(installCancel)

	logActivity("system-install", fmt.Sprintf("%s (%s)", toolInfo.Name, cmdStr))

	go func() {
		defer installCancel()
		cmd := enhancedCommand(installCtx, toolInfo.Cmd, toolInfo.Args...)
		out, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(out))

		if err != nil && installCtx.Err() == context.Canceled {
			finishTask(task, "cancelled", fmt.Sprintf("%s kurulumu durduruldu.", toolInfo.Name))
			return
		}

		if err != nil && (body.ID == "pipx" || body.ID == "uvx") {
			fbTarget := body.ID
			if body.ID == "uvx" {
				fbTarget = "uv"
			}
			pyBin := "python"
			if runtime.GOOS != "windows" {
				pyBin = "python3"
			}
			fbCmd := enhancedCommand(context.Background(), pyBin, "-m", "pip", "install", fbTarget, "--break-system-packages")
			fbOut, fbErr := fbCmd.CombinedOutput()
			if fbErr == nil {
				err = nil
				outputStr = strings.TrimSpace(string(fbOut))
			} else {
				fbCmd2 := enhancedCommand(context.Background(), pyBin, "-m", "pip", "install", "--user", fbTarget)
				fbOut2, fbErr2 := fbCmd2.CombinedOutput()
				if fbErr2 == nil {
					err = nil
					outputStr = strings.TrimSpace(string(fbOut2))
				}
			}
		}

		if err != nil {
			finishTask(task, "error", fmt.Sprintf("%s kurulum hatası: %s", toolInfo.Name, outputStr))
		} else {
			if body.ID == "pipx" {
				pyBin := "python"
				if runtime.GOOS != "windows" {
					pyBin = "python3"
				}
				_ = enhancedCommand(context.Background(), pyBin, "-m", "pipx", "ensurepath").Run()
			}
			finishTask(task, "completed", fmt.Sprintf("✅ %s kurulumu tamamlandı!", toolInfo.Name))
		}
	}()

	writeJSON(w, 200, map[string]interface{}{
		"ok":      true,
		"id":      body.ID,
		"name":    toolInfo.Name,
		"command": cmdStr,
		"message": fmt.Sprintf("⚡ %s kurulumu arka planda başlatıldı (%s).", toolInfo.Name, cmdStr),
	})
}

func detectPortByPID(pid int) string {
	if pid <= 0 {
		return ""
	}
	out, err := exec.Command("netstat", "-ano", "-p", "tcp").CombinedOutput()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "LISTENING") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 5 {
				linePid := fields[len(fields)-1]
				if linePid == strconv.Itoa(pid) {
					localAddr := fields[1]
					parts := strings.Split(localAddr, ":")
					if len(parts) > 1 {
						return parts[len(parts)-1]
					}
				}
			}
		}
	}
	return ""
}

func handleRunningApps(w http.ResponseWriter, r *http.Request) {
	runningAppsMux.Lock()
	defer runningAppsMux.Unlock()

	results := []*RunningApp{}
	for _, app := range runningAppsMap {
		if app.PID > 0 && app.Port == "" {
			app.Port = detectPortByPID(app.PID)
		}
		results = append(results, app)
	}

	writeJSON(w, 200, results)
}

// killRunningApp stops name's tracked process (if any) and its docker
// container (if any). Shared by the explicit "Durdur" button and repo
// deletion, so removing a repo can never leave an orphaned process running
// against a directory that no longer exists.
func killRunningApp(name string) {
	runningAppsMux.Lock()
	app, ok := runningAppsMap[name]
	if ok {
		delete(runningAppsMap, name)
	}
	runningAppsMux.Unlock()

	if ok && app.PID > 0 {
		if runtime.GOOS == "windows" {
			_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(app.PID)).Run()
		} else {
			if proc, err := os.FindProcess(app.PID); err == nil {
				_ = proc.Kill()
			}
		}
	}

	safeTag := "ai-toolkit-" + sanitizeDockerTag(name)
	_ = enhancedCommand(context.Background(), "docker", "stop", safeTag+"-run").Run()
}

func handleAppKill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	killRunningApp(name)
	logActivity("app-kill", fmt.Sprintf("%s durduruldu", name))

	writeJSON(w, 200, map[string]interface{}{
		"ok":      true,
		"name":    name,
		"message": fmt.Sprintf("🛑 '%s' uygulaması durduruldu.", name),
	})
}

func handleIdeBackup(w http.ResponseWriter, r *http.Request) {
	s, err := loadState()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}

	backupDir := filepath.Join(getBaseDir(), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		writeErr(w, 500, "yedek klasörü oluşturulamadı: "+err.Error())
		return
	}

	timestamp := time.Now().Format("20060102-150405")
	zipPath := filepath.Join(backupDir, fmt.Sprintf("ide-backup-%s.zip", timestamp))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		writeErr(w, 500, "zip oluşturma hatası: "+err.Error())
		return
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	backedUpCount := 0

	stateData, _ := json.MarshalIndent(s, "", "  ")
	if fw, err := zw.Create("ai-toolkit-state.json"); err == nil {
		_, _ = fw.Write(stateData)
		backedUpCount++
	}

	for _, p := range s.IdePaths {
		if p.Path == "" {
			continue
		}
		if data, err := os.ReadFile(p.Path); err == nil {
			safeName := fmt.Sprintf("%s-%s", p.ID, filepath.Base(p.Path))
			if fw, err := zw.Create(safeName); err == nil {
				_, _ = fw.Write(data)
				backedUpCount++
			}
		}
	}

	_ = zw.Close()

	logActivity("ide-backup", fmt.Sprintf("%s (%d dosya)", zipPath, backedUpCount))

	writeJSON(w, 200, map[string]interface{}{
		"ok":            true,
		"backupPath":    zipPath,
		"filename":      fmt.Sprintf("ide-backup-%s.zip", timestamp),
		"backedUpCount": backedUpCount,
		"message":       fmt.Sprintf("💾 Tüm IDE konfigürasyonları yedeklendi! (%d dosya zip paketine kaydedildi)", backedUpCount),
	})
}

// ---------- WyvDev Agentic Loop Engine (Autonomous AI Agent Infrastructure) ----------

func handleLoopHeartbeat(w http.ResponseWriter, r *http.Request) {
	runningAppsMux.Lock()
	apps := []*RunningApp{}
	for _, app := range runningAppsMap {
		if app.PID > 0 && app.Port == "" {
			app.Port = detectPortByPID(app.PID)
		}
		apps = append(apps, app)
	}
	runningAppsMux.Unlock()

	s, err := loadState()
	if err != nil {
		s = &StateBundle{}
	}

	writeJSON(w, 200, map[string]interface{}{
		"agenticLoopEngine": "active",
		"version":           "v0.9.0-beta",
		"timestamp":         time.Now().Format(time.RFC3339),
		"activeAppsCount":   len(apps),
		"runningApps":       apps,
		"idePathsCount":     len(s.IdePaths),
		"mcpServersCount":   len(s.McpServers),
		"trackedReposCount": len(s.TrackedRepos),
	})
}

func handleLoopAutoHeal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string `json:"name"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeErr(w, 400, "geçersiz repo adı")
		return
	}

	name := body.Name
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		writeErr(w, 404, "repo klasörü bulunamadı: "+dir)
		return
	}

	runningAppsMux.Lock()
	if app, ok := runningAppsMap[name]; ok && app.PID > 0 {
		if runtime.GOOS == "windows" {
			_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(app.PID)).Run()
		} else {
			if proc, err := os.FindProcess(app.PID); err == nil {
				_ = proc.Kill()
			}
		}
		delete(runningAppsMap, name)
	}
	runningAppsMux.Unlock()

	repoStartErrorsMux.Lock()
	delete(repoStartErrors, name)
	repoStartErrorsMux.Unlock()

	entry := scanRepoFolder(name)
	var healCmd *exec.Cmd
	healActionStr := "npm install --force"

	if containsRuntime(entry.Runtimes, "python") {
		healActionStr = "python -m pip install --upgrade --force-reinstall -r requirements.txt"
		healCmd = enhancedCommand(context.Background(), "python", "-m", "pip", "install", "--upgrade", "--force-reinstall", "-r", "requirements.txt")
	} else if containsRuntime(entry.Runtimes, "rust") {
		healActionStr = "cargo clean"
		healCmd = enhancedCommand(context.Background(), "cargo", "clean")
	} else {
		healCmd = enhancedCommand(context.Background(), "npm", "install", "--force")
	}

	healCmd.Dir = dir
	out, healErr := healCmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))

	logActivity("loop-auto-heal", fmt.Sprintf("%s (%s - %v)", name, healActionStr, healErr == nil))

	writeJSON(w, 200, map[string]interface{}{
		"ok":         healErr == nil,
		"name":       name,
		"healAction": healActionStr,
		"output":     outStr,
		"message":    fmt.Sprintf("🌀 WyvDev Agentic Loop Engine '%s' projesini başarıyla iyileştirdi (%s).", name, healActionStr),
	})
}

func handleLoopVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string `json:"name"`
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Command == "" {
		writeErr(w, 400, "geçersiz doğrulama komutu")
		return
	}

	dir := getBaseDir()
	if body.Name != "" {
		targetDir := repoDir(body.Name)
		if _, err := os.Stat(targetDir); err == nil {
			dir = targetDir
		}
	}

	parts := strings.Fields(body.Command)
	if len(parts) == 0 {
		writeErr(w, 400, "geçersiz doğrulama komutu")
		return
	}
	cmd := enhancedCommand(r.Context(), parts[0], parts[1:]...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	writeJSON(w, 200, map[string]interface{}{
		"passed":   err == nil,
		"exitCode": exitCode,
		"command":  body.Command,
		"dir":      dir,
		"output":   outStr,
	})
}

func handleLoopConfig(w http.ResponseWriter, r *http.Request) {
	s, err := loadState()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}

	if r.Method == http.MethodPost {
		var cfg LoopEngineConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeErr(w, 400, "geçersiz ayarlar: "+err.Error())
			return
		}
		s.LoopConfig = cfg
		if err := saveState(s); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, map[string]interface{}{
			"ok":      true,
			"config":  s.LoopConfig,
			"message": "🌀 WyvDev Agentic Loop Engine ayarları kaydedildi.",
		})
		return
	}

	if s.LoopConfig.MaxRetries == 0 {
		s.LoopConfig = LoopEngineConfig{
			AutoHealEnabled: true,
			AutoKillPort:    true,
			MaxRetries:      3,
			CacheStrategy:   "force",
		}
	}

	writeJSON(w, 200, s.LoopConfig)
}

// isLocalOrigin reports whether an Origin/Referer header points at this
// app's own 127.0.0.1/localhost:<port> — the only origin that should ever
// be allowed to drive state-changing requests. A wildcard CORS origin on a
// server that deletes repos and spawns processes lets *any* web page the
// user happens to have open issue those requests silently.
func isLocalOrigin(raw string) bool {
	if raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return false
	}
	host := u.Hostname()
	if host != "127.0.0.1" && host != "localhost" && host != "[::1]" && host != "::1" {
		return false
	}
	port := u.Port()
	return port == "" || port == strconv.Itoa(defaultPort)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isLocalOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			// Non-browser callers (curl, the app's own --once CLI mode)
			// never send an Origin header — nothing to restrict for them.
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}

		// A browser always sets Origin (and usually Referer) truthfully —
		// it can't be spoofed by page JS — so this is a reliable gate
		// against a malicious site driving destructive requests even
		// though the response itself can't be read cross-origin (CSRF,
		// not just data leakage, since these endpoints run commands and
		// delete files as a side effect of the request alone).
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			ref := r.Header.Get("Referer")
			if (origin != "" && !isLocalOrigin(origin)) || (origin == "" && ref != "" && !isLocalOrigin(ref)) {
				writeErr(w, 403, "forbidden origin")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func openBrowser(url string, addr string) {
	for i := 0; i < 30; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

const defaultPort = 47651

// ---------- Startup Self-Check Engine ----------

type StartupCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "ok", "warn", "error"
	Message string `json:"message"`
}

type StartupReport struct {
	Ready    bool           `json:"ready"`
	Checks   []StartupCheck `json:"checks"`
	DiskUsed string         `json:"diskUsed"`
	BaseDir  string         `json:"baseDir"`
}

var startupReport StartupReport
var startupMux sync.Mutex

func handleStartupStatus(w http.ResponseWriter, r *http.Request) {
	startupMux.Lock()
	defer startupMux.Unlock()
	writeJSON(w, 200, startupReport)
}

func runStartupChecks(baseDir string) {
	checks := []StartupCheck{}
	allOk := true

	addCheck := func(name, status, msg string) {
		checks = append(checks, StartupCheck{Name: name, Status: status, Message: msg})
		if status == "error" {
			allOk = false
		}
		fmt.Printf("  [%s] %s: %s\n", strings.ToUpper(status), name, msg)
	}

	fmt.Println("\n[WyvDev] 🔍 Startup Self-Check başlıyor...")

	// 1. Kendi statik dosyalarını doğrula
	requiredFiles := []string{"index.html", "app.js", "style.css", "loop.html", "search.html", "paths.html", "settings.html"}
	missingFiles := []string{}
	for _, f := range requiredFiles {
		if _, err := os.Stat(filepath.Join(baseDir, f)); err != nil {
			missingFiles = append(missingFiles, f)
		}
	}
	if len(missingFiles) == 0 {
		addCheck("Statik Dosyalar", "ok", "Tüm HTML/JS/CSS dosyaları mevcut")
	} else {
		addCheck("Statik Dosyalar", "warn", "Eksik dosyalar: "+strings.Join(missingFiles, ", "))
	}

	// 2. state.json doğrula ve onar
	s, err := loadState()
	if err != nil {
		_ = saveState(&StateBundle{TrackedRepos: []TrackedRepo{}})
		addCheck("State (Durum Verisi)", "warn", "state.json bozuktu, sıfırlandı ve yeniden oluşturuldu")
		s = &StateBundle{}
	} else {
		addCheck("State (Durum Verisi)", "ok", fmt.Sprintf("state.json geçerli — %d MCP, %d Skill, %d IDE yolu kayıtlı", len(s.McpServers), len(s.RecommendedRepos), len(s.IdePaths)))
	}

	// 3. repo/ klasörü disk kullanımı
	repoDir := filepath.Join(baseDir, "repo")
	diskUsed := "0 MB"
	if _, err := os.Stat(repoDir); err == nil {
		var totalSize int64
		_ = filepath.Walk(repoDir, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		mb := totalSize / 1024 / 1024
		diskUsed = fmt.Sprintf("%d MB", mb)
		addCheck("Disk (repo/ Klasörü)", "ok", fmt.Sprintf("repo/ klasörü mevcut — %d MB kullanımda", mb))
	} else {
		_ = os.MkdirAll(repoDir, 0755)
		addCheck("Disk (repo/ Klasörü)", "ok", "repo/ klasörü oluşturuldu")
	}

	// 4. Kayıtlı repolar klonlu mu? Eksikleri arka planda clone et
	if len(s.TrackedRepos) > 0 {
		missingRepos := []string{}
		for _, tr := range s.TrackedRepos {
			repoPath := filepath.Join(baseDir, "repo", tr.Name)
			if _, err := os.Stat(repoPath); err != nil {
				missingRepos = append(missingRepos, tr.Name)
			}
		}
		if len(missingRepos) > 0 {
			addCheck("Kayıtlı Repolar", "warn", fmt.Sprintf("%d repo eksik, arka planda klonlanıyor: %s", len(missingRepos), strings.Join(missingRepos, ", ")))
			go autoCloneMissingRepos(s)
		} else {
			addCheck("Kayıtlı Repolar", "ok", fmt.Sprintf("Tüm %d repo klonlu ve mevcut", len(s.TrackedRepos)))
		}
	} else {
		addCheck("Kayıtlı Repolar", "ok", "Henüz takip edilen repo yok")
	}

	// 5. IDE yollarını doğrula ve geçersizleri temizle
	if len(s.IdePaths) > 0 {
		validPaths := []IdePathEntry{}
		removed := 0
		for _, ide := range s.IdePaths {
			dir := filepath.Dir(ide.Path)
			if _, err := os.Stat(dir); err == nil {
				validPaths = append(validPaths, ide)
			} else {
				removed++
			}
		}
		if removed > 0 {
			s.IdePaths = validPaths
			_ = saveState(s)
			addCheck("IDE Yolları", "warn", fmt.Sprintf("%d geçersiz IDE yolu temizlendi, %d yol geçerli", removed, len(validPaths)))
		} else {
			addCheck("IDE Yolları", "ok", fmt.Sprintf("Tüm %d IDE yolu geçerli", len(s.IdePaths)))
		}
	} else {
		addCheck("IDE Yolları", "ok", "IDE yolları state'den otomatik algılanacak")
	}

	// 6. Port çakışma kontrolü (zaten dinliyorsa bildir)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", defaultPort), 100*time.Millisecond)
	if err == nil {
		conn.Close()
		addCheck("Port Kontrolü", "warn", fmt.Sprintf("Port %d zaten kullanımda — mevcut instance'a bağlanılıyor", defaultPort))
	} else {
		addCheck("Port Kontrolü", "ok", fmt.Sprintf("Port %d müsait, sunucu başlatılıyor", defaultPort))
	}

	// 7. Kritik runtime araçları
	runtimes := map[string]string{
		"git":  "git --version",
		"node": "node --version",
		"go":   "go version",
	}
	missingRuntimes := []string{}
	for tool := range runtimes {
		versionArg := "--version"
		if tool == "go" {
			versionArg = "version"
		}
		if err := enhancedCommand(context.Background(), tool, versionArg).Run(); err != nil {
			missingRuntimes = append(missingRuntimes, tool)
		}
	}
	if len(missingRuntimes) == 0 {
		addCheck("Runtime Araçları", "ok", "git, node, go — tüm araçlar kurulu")
	} else {
		addCheck("Runtime Araçları", "warn", "Eksik araçlar: "+strings.Join(missingRuntimes, ", ")+" (Sistem & Teşhis sayfasından kurulabilir)")
	}

	startupMux.Lock()
	startupReport = StartupReport{
		Ready:    allOk,
		Checks:   checks,
		DiskUsed: diskUsed,
		BaseDir:  baseDir,
	}
	startupMux.Unlock()

	okCount := 0
	warnCount := 0
	for _, c := range checks {
		if c.Status == "ok" {
			okCount++
		} else if c.Status == "warn" {
			warnCount++
		}
	}
	fmt.Printf("[WyvDev] ✅ Startup kontrolü tamamlandı: %d OK, %d uyarı\n", okCount, warnCount)
	fmt.Println()
}

func runServer() {
	baseDir := getBaseDir()

	// Run startup checks in background before serving
	go runStartupChecks(baseDir)

	// Start 1-second background sync controller (smart idle)
	runBackgroundSyncController()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/state", handleState)
	mux.HandleFunc("POST /api/state", handleState)
	mux.HandleFunc("POST /api/state/migrate", handleStateMigrate)
	mux.HandleFunc("GET /api/startup/status", handleStartupStatus)
	mux.HandleFunc("GET /api/ides/detect", handleIdesDetect)
	mux.HandleFunc("POST /api/mcp/test", handleMcpTest)
	mux.HandleFunc("GET /api/marketplace/catalog", handleMarketplaceCatalog)
	mux.HandleFunc("POST /api/marketplace/install", handleMarketplaceInstall)
	mux.HandleFunc("POST /api/ides/backup", handleIdeBackup)
	mux.HandleFunc("POST /api/ides/{id}/danger-delete", handleDangerDelete)
	mux.HandleFunc("POST /api/repos/{name}/check", handleRepoCheck)
	mux.HandleFunc("POST /api/repos/{name}/pull", handleRepoPull)
	mux.HandleFunc("POST /api/repos/{name}/delete", handleRepoDelete)
	mux.HandleFunc("POST /api/repos/{name}/run", handleRepoRun)
	mux.HandleFunc("POST /api/repos/{name}/enable-skill", handleRepoEnableSkill)
	mux.HandleFunc("POST /api/skills/sync-all", handleSkillsSyncAll)
	mux.HandleFunc("POST /api/repos/{name}/start", handleRepoStart)
	mux.HandleFunc("GET /api/repos/{name}/env-vars", handleRepoEnvVars)
	mux.HandleFunc("GET /api/repos/{name}/analyze", handleRepoAnalyze)
	mux.HandleFunc("GET /api/repos/{name}/install/stream", handleRepoInstallStream)
	mux.HandleFunc("GET /api/repos/scan", handleReposScan)
	mux.HandleFunc("POST /api/repos/check-all", handleCheckAllRepos)
	mux.HandleFunc("GET /api/activity", handleActivity)
	mux.HandleFunc("GET /api/tasks", handleTasks)
	mux.HandleFunc("POST /api/tasks/{id}/stop", handleTaskStop)
	mux.HandleFunc("DELETE /api/tasks/{id}", handleTaskDelete)
	mux.HandleFunc("GET /api/apps/running", handleRunningApps)
	mux.HandleFunc("POST /api/apps/{name}/kill", handleAppKill)
	mux.HandleFunc("GET /api/system/health", handleSystemHealth)
	mux.HandleFunc("POST /api/system/install", handleSystemInstall)
	mux.HandleFunc("GET /api/loop/heartbeat", handleLoopHeartbeat)
	mux.HandleFunc("POST /api/loop/auto-heal", handleLoopAutoHeal)
	mux.HandleFunc("POST /api/loop/verify", handleLoopVerify)
	mux.HandleFunc("GET /api/loop/config", handleLoopConfig)
	mux.HandleFunc("POST /api/loop/config", handleLoopConfig)
	mux.HandleFunc("GET /api/github/search", handleGithubSearch)
	mux.HandleFunc("POST /api/skills/install", handleSkillInstall)
	mux.Handle("/", http.FileServer(http.Dir(baseDir)))

	addr := fmt.Sprintf("127.0.0.1:%d", defaultPort)
	url := fmt.Sprintf("http://%s/index.html", addr)

	fmt.Println("================================================================")
	fmt.Println(" WyvDev Hub — Local Developer PaaS & Universal MCP Orchestrator")
	fmt.Printf(" Sunucu dizini : %s\n", baseDir)
	fmt.Printf(" Adres         : %s\n", url)
	fmt.Println(" Kapatmak icin bu pencereyi kapatin (Ctrl+C).")
	fmt.Println("================================================================")

	go openBrowser(url, addr)

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		fmt.Printf("❌ Sunucu baslatilamadi: %v\n", err)
		waitForExit()
	}

}

// ---------- legacy one-shot CLI mode (ai-toolkit --once) ----------

func runSkillsScript() {
	fmt.Println("\n================================================================")
	fmt.Println(" [Skills] Global Skill Kurulumu Senkronize Ediliyor...")
	fmt.Println("================================================================")

	s, err := loadState()
	if err != nil || len(s.RecommendedRepos) == 0 {
		fmt.Println("✅ Saklı skill bulunamadı, tamamlandı.")
		return
	}

	for _, skill := range s.RecommendedRepos {
		if skill.Repo != "" {
			fmt.Printf("  - Skill kuruluyor: %s (%s)...\n", skill.Name, skill.Repo)
			_ = exec.Command("npx", "-y", "skills", "add", skill.Repo, "--all").Run()
		}
	}
	fmt.Println("✅ Global Skiller başarıyla senkronize edildi.")
}

func runOnceCli() {
	fmt.Println("================================================================")
	fmt.Println(" AI Toolkit Hub — Tek Seferlik Senkron (Go, --once)")
	fmt.Println("================================================================")
	fmt.Println()

	ides := detectIdes()
	fmt.Println("[1/3] Tespit Edilen AI IDE Konfigürasyon Yolları:")
	for _, ide := range ides {
		status := "❌ Klasör Bulunamadı"
		if ide.Detected {
			status = "✅ Algılandı"
		}
		fmt.Printf("  - %-25s : %s\n    └─ %s\n", ide.Name, status, ide.Path)
	}
	fmt.Println()

	template, err := loadCoreTemplate()
	if err != nil {
		fmt.Printf("❌ Şablon yüklenirken hata: %v\n", err)
		waitForExit()
		return
	}

	fmt.Printf("[2/3] core/mcp-config.json Şablonundaki MCP Sunucu Sayısı: %d\n", len(template.McpServers))
	fmt.Println("MCP'ler tüm tespit edilen IDE yollarına senkronize ediliyor...")
	fmt.Println()

	for _, ide := range ides {
		count, err := syncToIde(ide, template)
		if err != nil {
			fmt.Printf(" ⚠️ %s senkronize edilemedi: %v\n", ide.Name, err)
		} else {
			fmt.Printf(" ✅ %s -> %d MCP sunucusu yazıldı/güncellendi.\n    └─ %s\n", ide.Name, count, ide.Path)
		}
	}
	fmt.Println()

	fmt.Println("[3/3] Global Skiller Kurulsun mu?")
	fmt.Print("Global skill kurulum scriptini çalıştırmak istiyor musunuz? (E/H) [Varsayılan: E]: ")

	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToUpper(response))

	if response == "" || response == "E" || response == "Y" {
		runSkillsScript()
	}

	fmt.Println("\n================================================================")
	fmt.Println(" 🎉 İşlem Tamamlandı! Tüm IDE'leriniz MCP ve Skiller ile hazır.")
	fmt.Println("================================================================")
	waitForExit()
}

func waitForExit() {
	fmt.Println("\nÇıkmak için Enter tuşuna basın...")
	var input string
	fmt.Scanln(&input)
}

func main() {
	once := flag.Bool("once", false, "Tek seferlik CLI senkron modu (eski davranis)")
	flag.Parse()

	if *once {
		runOnceCli()
		return
	}

	// Singleton guard — eğer zaten çalışan bir instance varsa sadece tarayıcı aç ve çık
	addr := fmt.Sprintf("127.0.0.1:%d", defaultPort)
	url := fmt.Sprintf("http://%s/index.html", addr)
	conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
	if err == nil {
		conn.Close()
		fmt.Println("================================================================")
		fmt.Println(" WyvDev Hub — Zaten çalışıyor!")
		fmt.Printf(" Mevcut sunucuya bağlanılıyor: %s\n", url)
		fmt.Println("================================================================")
		openBrowser(url, addr)
		return
	}

	runServer()
}
