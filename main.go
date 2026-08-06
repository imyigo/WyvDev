package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	Status    string    `json:"status"`  // "running", "completed", "error"
	Message   string    `json:"message"` // detail or output snippet
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt,omitempty"`
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

type McpServerConfig struct {
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type McpConfigFile struct {
	McpServers map[string]McpServerConfig `json:"mcpServers"`
}

type DetectedIde struct {
	ID        string
	Name      string
	Path      string // MCP config file path
	SkillsDir string // global skills folder for this agent, e.g. ~/.claude/skills — empty if unknown
	Detected  bool
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

func detectIdes() []DetectedIde {
	home := getHomeDir()
	appData := getAppDataDir()
	baseDir := getBaseDir()

	candidates := []DetectedIde{
		{ID: "antigravity-app", Name: "Antigravity IDE & App", Path: filepath.Join(home, ".gemini", "antigravity", "mcp_config.json")},
		{ID: "antigravity-global", Name: "Antigravity Global Config", Path: filepath.Join(home, ".gemini", "config", "mcp_config.json")},
		{ID: "claude-desktop", Name: "Claude Desktop App", Path: filepath.Join(appData, "Claude", "claude_desktop_config.json")},
		{ID: "cursor", Name: "Cursor IDE", Path: filepath.Join(home, ".cursor", "mcp.json")},
		{ID: "cursor-cline", Name: "Cursor (Cline Extension)", Path: filepath.Join(appData, "Cursor", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json")},
		{ID: "claude-code", Name: "Claude Code CLI", Path: filepath.Join(home, ".claude.json"), SkillsDir: filepath.Join(home, ".claude", "skills")},
		{ID: "vscode-cline", Name: "VS Code (Cline Extension)", Path: filepath.Join(appData, "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json")},
		{ID: "vscode-roo", Name: "VS Code (Roo Code Extension)", Path: filepath.Join(appData, "Code", "User", "globalStorage", "rooveterinaryinc.roo-cline", "settings", "cline_mcp_settings.json")},
		{ID: "vscode-workspace", Name: "VS Code Workspace", Path: filepath.Join(baseDir, ".vscode", "mcp.json")},
		{ID: "windsurf", Name: "Windsurf IDE", Path: filepath.Join(appData, "Windsurf", "User", "globalStorage", "mcp_config.json")},
		{ID: "zed", Name: "Zed Editor", Path: filepath.Join(home, ".config", "zed", "settings.json")},
		{ID: "continue", Name: "Continue.dev", Path: filepath.Join(home, ".continue", "config.json")},
		{ID: "jetbrains", Name: "JetBrains IDEs", Path: filepath.Join(appData, "JetBrains", "mcp.json")},
	}

	for i := range candidates {
		dir := filepath.Dir(candidates[i].Path)
		if _, err := os.Stat(dir); err == nil {
			candidates[i].Detected = true
		} else if _, err := os.Stat(candidates[i].Path); err == nil {
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
			if len(e.Env) > 0 {
				sc.Env = e.Env
			}
		}
		cfg.McpServers[e.ID] = sc
	}
	return cfg
}

func syncToIde(targetPath string, template *McpConfigFile) (int, error) {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("klasor olusturulamadi: %v", err)
	}

	if _, err := os.Stat(targetPath); err == nil {
		backupPath := fmt.Sprintf("%s.bak-%s", targetPath, time.Now().Format("20060102-150405"))
		_ = copyFile(targetPath, backupPath)
	}

	var existing McpConfigFile
	existing.McpServers = make(map[string]McpServerConfig)

	if data, err := os.ReadFile(targetPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	if existing.McpServers == nil {
		existing.McpServers = make(map[string]McpServerConfig)
	}

	addedCount := 0
	for name, server := range template.McpServers {
		existing.McpServers[name] = server
		addedCount++
	}

	updatedData, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("JSON olusturma hatasi: %v", err)
	}

	if err := os.WriteFile(targetPath, updatedData, 0644); err != nil {
		return 0, fmt.Errorf("dosya yazma hatasi: %v", err)
	}

	return addedCount, nil
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
		count, err := syncToIde(ide.Path, template)
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

func loadState() (*StateBundle, error) {
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
	if err := os.MkdirAll(filepath.Join(getBaseDir(), "core"), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(), data, 0644)
}

// ---------- git tracking ----------

func repoDir(name string) string {
	return filepath.Join(getBaseDir(), "repo", name)
}

func autoCloneMissingRepos(s *StateBundle) []map[string]interface{} {
	results := []map[string]interface{}{}
	tracked := map[string]bool{}
	for _, t := range s.TrackedRepos {
		tracked[t.Repo] = true
	}

	// Anything with a "repo" (owner/name) field gets cloned into repo/<name> and
	// tracked — whether it arrived as a skill or as an MCP server (e.g. added
	// from the GitHub live search tab).
	var repoRefs []string
	for _, skill := range s.RecommendedRepos {
		if skill.Repo != "" {
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
		cmd := exec.Command("git", "clone", "https://github.com/"+repoRef+".git", localPath)
		out, err := cmd.CombinedOutput()
		result := map[string]interface{}{"repo": repoRef, "path": localPath}
		if err != nil {
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

func gitCheck(name string) (string, error) {
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("yerel klasor yok: %s", dir)
	}

	if out, err := exec.Command("git", "-C", dir, "fetch").CombinedOutput(); err != nil {
		return "error", fmt.Errorf("git fetch basarisiz: %s", strings.TrimSpace(string(out)))
	}

	localOut, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "error", fmt.Errorf("rev-parse HEAD: %v", err)
	}
	remoteOut, err := exec.Command("git", "-C", dir, "rev-parse", "@{u}").CombinedOutput()
	if err != nil {
		return "error", fmt.Errorf("upstream yok: %v", err)
	}

	local := strings.TrimSpace(string(localOut))
	remote := strings.TrimSpace(string(remoteOut))
	if local == remote {
		return "upToDate", nil
	}

	countOut, _ := exec.Command("git", "-C", dir, "rev-list", "--left-right", "--count", "HEAD...@{u}").CombinedOutput()
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

		// Pages that don't manage trackedRepos (index/skills/paths) push a bundle
		// without that field — preserve whatever is already on disk instead of wiping it.
		if len(incoming.TrackedRepos) == 0 {
			if existing, err := loadState(); err == nil {
				incoming.TrackedRepos = existing.TrackedRepos
			}
		}

		if err := saveState(&incoming); err != nil {
			writeErr(w, 500, err.Error())
			return
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

func handleRepoCheck(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	status, err := gitCheck(name)
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
	dir := repoDir(name)
	if _, err := os.Stat(dir); err != nil {
		writeErr(w, 404, "yerel klasor yok")
		return
	}
	out, err := exec.Command("git", "-C", dir, "pull").CombinedOutput()
	if err != nil {
		writeErr(w, 500, strings.TrimSpace(string(out)))
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "true", "output": strings.TrimSpace(string(out))})
}

func handleRepoDelete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeErr(w, 400, "repo adi belirtilmedi")
		return
	}

	dir := repoDir(name)
	deletedOnDisk := false
	if _, err := os.Stat(dir); err == nil {
		if err := os.RemoveAll(dir); err != nil {
			writeErr(w, 500, fmt.Sprintf("klasor silinemedi (%s): %v", dir, err))
			return
		}
		deletedOnDisk = true
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
	for _, m := range s.McpServers {
		parts := strings.Split(m.Repo, "/")
		repoName := parts[len(parts)-1]
		if m.ID != name && m.Name != name && m.Repo != name && repoName != name {
			stillMcp = append(stillMcp, m)
		}
	}
	s.McpServers = stillMcp

	pruned := reconcileState(s)
	_ = saveState(s)

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

	toolPath, err := exec.LookPath(toolName)
	if err != nil {
		writeErr(w, 424, fmt.Sprintf("%s bulunamadi - PATH'e kurulu olmali", toolName))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, toolPath, args...)
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

// ---------- real skill install (npx skills add ...) ----------

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

	npxPath, err := exec.LookPath("npx")
	if err != nil {
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

	cmd := exec.CommandContext(ctx, npxPath, args...)
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
	topOut, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").CombinedOutput()
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

	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").CombinedOutput()
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

	skillPath := filepath.Join(dir, "SKILL.md")
	if _, err := os.Stat(skillPath); err == nil {
		entry.HasSkill = true
		entry.SkillDesc = extractSkillDesc(skillPath)
	}

	pkgPath := filepath.Join(dir, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		entry.HasPackageJson = true
		var pkg struct {
			Name        string      `json:"name"`
			Description string      `json:"description"`
			Bin         interface{} `json:"bin"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			entry.PackageName = pkg.Name
			haystack := strings.ToLower(pkg.Name + " " + pkg.Description + " " + name)
			entry.LooksLikeMcp = pkg.Bin != nil || strings.Contains(haystack, "mcp") || strings.Contains(haystack, "dokploy")
		}
	}

	if !entry.LooksLikeMcp {
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "mcp") || strings.Contains(lowerName, "dokploy") {
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

					priorityScripts := []string{"dev", "dokploy:dev", "start", "serve", "preview", "watch", "build", "run", "cli"}
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
						for sName := range pkg.Scripts {
							if pkgManager == "pnpm" {
								entry.StartCommand = "pnpm run " + sName
							} else {
								entry.StartCommand = "npm run " + sName
							}
							break
						}
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
	} else if entry.HasPackageJson {
		entry.RepoType = "library"
		entry.RepoTypeLabel = "📦 Library"
	} else {
		entry.RepoType = "other"
		entry.RepoTypeLabel = "📁 Diğer"
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

func handleRepoStart(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
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

	parts := strings.Fields(startCmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir

	logActivity("repo-start", fmt.Sprintf("%s (%s)", name, startCmdStr))

	go func() {
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
		if err != nil {
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

	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		name := de.Name()
		status, checkErr := gitCheck(name)
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

	msg := fmt.Sprintf("%d repo kontrol edildi.", len(results))
	if behindCount > 0 {
		msg += fmt.Sprintf(" %d repo için güncelleme mevcut!", behindCount)
	}
	finishTask(task, "completed", msg)

	writeJSON(w, 200, map[string]interface{}{
		"ok":          true,
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
		filepath.Join(localAppData, "Programs", "Python"),
		filepath.Join(homeDir, ".cargo", "bin"),
		filepath.Join(homeDir, ".local", "bin"),
		filepath.Join(localAppData, "Programs", "Python", "Python312", "Scripts"),
		filepath.Join(localAppData, "Programs", "Python", "Python311", "Scripts"),
		filepath.Join(localAppData, "Programs", "Python", "Python310", "Scripts"),
		filepath.Join(localAppData, "Programs", "Python", "Scripts"),
		`C:\Program Files\Git\cmd`,
		`C:\Program Files\nodejs`,
		`C:\Program Files\Docker\Docker\resources\bin`,
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

func checkToolVersion(candidates ...[]string) (bool, string) {
	enhancedEnv := getEnhancedEnv()
	for _, cand := range candidates {
		if len(cand) == 0 {
			continue
		}
		cmd := exec.Command(cand[0], cand[1:]...)
		cmd.Env = enhancedEnv

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
		Candidates                             [][]string
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
		{"pnpm", "pnpm Paket Yöneticisi", "JavaScript", "Monorepo projeleri (ör. dokploy) paket yönetimi", "npm i -g pnpm", [][]string{
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

func handleSystemInstall(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		writeErr(w, 400, "geçersiz istek id")
		return
	}

	installCmds := map[string]struct {
		Name string
		Cmd  string
		Args []string
	}{
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

	toolInfo, ok := installCmds[body.ID]
	if !ok {
		writeErr(w, 400, "otomatik kurulum komutu bulunamadı")
		return
	}

	cmdStr := toolInfo.Cmd + " " + strings.Join(toolInfo.Args, " ")
	task := addTask("system-install", toolInfo.Name, fmt.Sprintf("%s otomatik kuruluyor (%s)...", toolInfo.Name, cmdStr))

	logActivity("system-install", fmt.Sprintf("%s (%s)", toolInfo.Name, cmdStr))

	go func() {
		cmd := exec.Command(toolInfo.Cmd, toolInfo.Args...)
		cmd.Env = getEnhancedEnv()
		out, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(out))

		if err != nil && (body.ID == "pipx" || body.ID == "uvx") {
			fbTarget := body.ID
			if body.ID == "uvx" {
				fbTarget = "uv"
			}
			fbCmd := exec.Command("pip", "install", fbTarget)
			fbCmd.Env = getEnhancedEnv()
			fbOut, fbErr := fbCmd.CombinedOutput()
			if fbErr == nil {
				err = nil
				outputStr = strings.TrimSpace(string(fbOut))
			}
		}

		if err != nil {
			finishTask(task, "error", fmt.Sprintf("%s kurulum hatası: %s", toolInfo.Name, outputStr))
		} else {
			if body.ID == "pipx" {
				_ = exec.Command("python", "-m", "pipx", "ensurepath").Run()
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

func handleAppKill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

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
	_ = exec.Command("docker", "stop", safeTag+"-run").Run()

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
		healCmd = exec.Command("python", "-m", "pip", "install", "--upgrade", "--force-reinstall", "-r", "requirements.txt")
	} else if containsRuntime(entry.Runtimes, "rust") {
		healActionStr = "cargo clean"
		healCmd = exec.Command("cargo", "clean")
	} else {
		healCmd = exec.Command("npm", "install", "--force")
	}

	healCmd.Dir = dir
	healCmd.Env = getEnhancedEnv()
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
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Env = getEnhancedEnv()

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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(204)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func openBrowser(url string) {
	time.Sleep(300 * time.Millisecond)
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

func runServer() {
	baseDir := getBaseDir()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/state", handleState)
	mux.HandleFunc("POST /api/state", handleState)
	mux.HandleFunc("POST /api/state/migrate", handleStateMigrate)
	mux.HandleFunc("GET /api/ides/detect", handleIdesDetect)
	mux.HandleFunc("POST /api/ides/backup", handleIdeBackup)
	mux.HandleFunc("POST /api/ides/{id}/danger-delete", handleDangerDelete)
	mux.HandleFunc("POST /api/repos/{name}/check", handleRepoCheck)
	mux.HandleFunc("POST /api/repos/{name}/pull", handleRepoPull)
	mux.HandleFunc("POST /api/repos/{name}/delete", handleRepoDelete)
	mux.HandleFunc("POST /api/repos/{name}/run", handleRepoRun)
	mux.HandleFunc("POST /api/repos/{name}/start", handleRepoStart)
	mux.HandleFunc("GET /api/repos/scan", handleReposScan)
	mux.HandleFunc("POST /api/repos/check-all", handleCheckAllRepos)
	mux.HandleFunc("GET /api/activity", handleActivity)
	mux.HandleFunc("GET /api/tasks", handleTasks)
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

	go openBrowser(url)

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
		count, err := syncToIde(ide.Path, template)
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
	runServer()
}
