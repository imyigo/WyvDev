# AI Toolkit - Global kurulum (workspace'e DEGIL, gercek IDE global yollarina kurar)
# install-skills.bat tarafindan cagrilir. Elle de calistirilabilir:
#   powershell -ExecutionPolicy Bypass -File install-skills.ps1

$ErrorActionPreference = 'Continue'
$RepoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$Home1 = $HOME
$AppData1 = $env:APPDATA

Write-Host "================================================================"
Write-Host " AI Toolkit - Global kurulum (workspace'e degil, IDE global yollarina)"
Write-Host "================================================================"
Write-Host ""

# ---------------------------------------------------------------------------
# 0) Bu makinede GERCEKTEN kurulu olan agent/IDE'leri tespit et.
#    Her agent'in global skills klasorunu skills CLI'nin kendi tablosundan
#    aldik (npx skills add --help / README), varsayilan/tahmini degil.
# ---------------------------------------------------------------------------
$AgentCandidates = @(
    @{ Id = 'claude-code';     Name = 'Claude Code';     Marker = Join-Path $Home1 '.claude' }
    @{ Id = 'claude-desktop';  Name = 'Claude Desktop';  Marker = Join-Path $AppData1 'Claude' }
    @{ Id = 'cursor';          Name = 'Cursor';           Marker = Join-Path $Home1 '.cursor' }
    @{ Id = 'antigravity';     Name = 'Antigravity';      Marker = Join-Path $Home1 '.gemini\antigravity' }
    @{ Id = 'antigravity-cli'; Name = 'Antigravity CLI';  Marker = Join-Path $Home1 '.gemini\antigravity-cli' }
    @{ Id = 'codex';           Name = 'Codex CLI';        Marker = Join-Path $Home1 '.codex' }
    @{ Id = 'gemini-cli';      Name = 'Gemini CLI';       Marker = Join-Path $Home1 '.gemini\settings.json' }
    @{ Id = 'opencode';        Name = 'OpenCode';         Marker = Join-Path $Home1 '.opencode' }
    @{ Id = 'windsurf';        Name = 'Windsurf';         Marker = Join-Path $Home1 '.codeium\windsurf' }
)

$Detected = $AgentCandidates | Where-Object { Test-Path $_.Marker }

if (-not $Detected -or $Detected.Count -eq 0) {
    Write-Host "Bu makinede desteklenen hicbir AI kodlama IDE'si bulunamadi (~/.claude, ~/.cursor, ~/.gemini/antigravity, ...)."
    Write-Host "Kurulum yapilacak bir yer yok, cikiliyor."
    exit 1
}

Write-Host "[0/7] Tespit edilen IDE'ler (bu gelistiricinin makinesinde gercekten kurulu olanlar):"
foreach ($a in $Detected) { Write-Host "   - $($a.Name)  (--agent $($a.Id))" }
Write-Host ""

$AgentArgs = @()
foreach ($a in $Detected) { 
    if ($a.Id -ne 'claude-desktop') { $AgentArgs += @('-a', $a.Id) }
}
$HasClaudeCode = $Detected.Id -contains 'claude-code'
$HasClaudeDesktop = $Detected.Id -contains 'claude-desktop' -or (Test-Path (Join-Path $AppData1 'Claude'))
$HasCursor = $Detected.Id -contains 'cursor'
$HasAntigravity = $Detected.Id -contains 'antigravity'

# Skill kurulumlarinin ASLA workspace'e sizmamasi icin notron bir dizinden calistir.
Set-Location $Home1

# ---------------------------------------------------------------------------
# 0b) Bu workspace'te daha once (eski/bugly calisma yuzunden) proje-scope
#     olarak kurulmus skill kalintilarini temizle (.agents/skills, .claude/skills,
#     skills-lock.json). Bunlar sadece bu proje icinde gecerliydi, ise yaramiyordu.
# ---------------------------------------------------------------------------
$StaleLock = Join-Path $RepoRoot 'skills-lock.json'
if (Test-Path $StaleLock) {
    Write-Host "[0b/7] Workspace'teki eski proje-scope skill kalintilari temizleniyor..."
    Push-Location $RepoRoot
    & npx --yes skills remove --all -a claude-code -a cursor -a antigravity -y 2>&1 | Out-Null
    Pop-Location
    foreach ($p in @('.agents', '.claude', 'skills-lock.json')) {
        $full = Join-Path $RepoRoot $p
        if (Test-Path $full) { Remove-Item -Recurse -Force $full -ErrorAction SilentlyContinue }
    }
    Write-Host "   -- temizlendi."
    Write-Host ""
}

# ---------------------------------------------------------------------------
# 1) Zaten kurulu global skill'leri guncelle (yuklu olmayanlara dokunmaz)
# ---------------------------------------------------------------------------
Write-Host "[1/7] Zaten kurulu olan global skiller guncelleniyor..."
& npx --yes skills update -g -y
Write-Host ""

# ---------------------------------------------------------------------------
# 2) Skill paketleri - sadece tespit edilen agent'lara, gercek global scope (-g)
# ---------------------------------------------------------------------------
$SkillPackages = @(
    @{ Step = '2/7'; Desc = 'ui-ux-pro-max';               Repo = 'nextlevelbuilder/ui-ux-pro-max-skill'; Extra = @('--skill', 'ui-ux-pro-max') }
    @{ Step = '3/7'; Desc = 'book-to-skill';                Repo = 'virgiliojr94/book-to-skill';           Extra = @('--all') }
    @{ Step = '4/7'; Desc = 'scientific-agent-skills';      Repo = 'K-Dense-AI/scientific-agent-skills';   Extra = @('--all') }
    @{ Step = '5/7'; Desc = 'Anthropic-Cybersecurity-Skills'; Repo = 'mukul975/Anthropic-Cybersecurity-Skills'; Extra = @('--all') }
    @{ Step = '6/7'; Desc = 'swiftui-agent-skill';          Repo = 'twostraws/swiftui-agent-skill';        Extra = @('--all') }
    @{ Step = '7/7'; Desc = 'chrisbanes/skills';            Repo = 'chrisbanes/skills';                    Extra = @('--all') }
)

foreach ($pkg in $SkillPackages) {
    Write-Host "[$($pkg.Step)] $($pkg.Desc) kuruluyor (global, zaten kuruluysa sessizce gecer)..."
    $argList = @('add', $pkg.Repo) + $pkg.Extra + $AgentArgs + @('-g', '-y')
    & npx --yes skills @argList
    Write-Host ""
}

# ---------------------------------------------------------------------------
# 3) Claude Code plugin marketplace (yalniz Claude Code tespit edildiyse)
#    Not: bunlar "claude" CLI'nin gercek alt komutlaridir, .bat icinden calisir -
#    slash-command olarak elle yapistirmaya gerek yok.
# ---------------------------------------------------------------------------
if ($HasClaudeCode) {
    Write-Host "[Plugin] Claude Code plugin marketplace kuruluyor..."
    $Plugins = @(
        @{ Market = 'nextlevelbuilder/ui-ux-pro-max-skill'; Plugin = 'ui-ux-pro-max@ui-ux-pro-max-skill' }
        @{ Market = 'addyosmani/agent-skills';               Plugin = 'agent-skills@addy-agent-skills' }
        @{ Market = 'AgriciDaniel/claude-seo';                Plugin = 'claude-seo@agricidaniel-claude-seo' }
        @{ Market = 'coreyhaines31/marketingskills';          Plugin = 'marketing-skills' }
        @{ Market = 'forrestchang/andrej-karpathy-skills';    Plugin = 'andrej-karpathy-skills@karpathy-skills' }
        @{ Market = 'mvanhorn/last30days-skill';              Plugin = 'last30days' }
        @{ Market = 'kepano/obsidian-skills';                 Plugin = 'obsidian@obsidian-skills' }
        @{ Market = 'Imbad0202/academic-research-skills';     Plugin = 'academic-research-skills' }
        @{ Market = 'thedotmack/claude-mem';                  Plugin = 'claude-mem' }
    )
    foreach ($p in $Plugins) {
        & claude plugin marketplace add $p.Market 2>&1 | Out-Null
        & claude plugin install $p.Plugin -s user 2>&1 | Out-Null
        Write-Host "   - $($p.Plugin)"
    }
    Write-Host ""
}

# ---------------------------------------------------------------------------
# 4) pip/npm paketleri (agent secimi gerektirmez)
# ---------------------------------------------------------------------------
Write-Host "================================================================"
Write-Host " pip/npm paketleri"
Write-Host "================================================================"
Write-Host "[graphify]"
pip install graphifyy
Write-Host ""
Write-Host "[headroom]"
pip install "headroom-ai[all]"
Write-Host "   -- kurulumdan sonra 'headroom wrap claude' ile aktif et"
Write-Host ""
Write-Host "[career-ops - DOGRULANMAMIS, hata verirse atla]"
& npx --yes @santifer/career-ops init
Write-Host ""

# ---------------------------------------------------------------------------
# 5) MCP sunucularini tespit edilen IDE'lerin GERCEK global config dosyalarina
#    yaz (mcp-config.json referans alinir, mevcut kayitlar asla ezilmez).
# ---------------------------------------------------------------------------
$McpConfigPath = Join-Path $RepoRoot 'mcp-config.json'
if (Test-Path $McpConfigPath) {
    Write-Host "================================================================"
    Write-Host " MCP sunuculari senkronize ediliyor"
    Write-Host "================================================================"
    $McpData = Get-Content $McpConfigPath -Raw | ConvertFrom-Json

    function Merge-McpFile {
        param([string]$Path, [pscustomobject]$Servers)
        $dir = Split-Path -Parent $Path
        if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
        if (Test-Path $Path) {
            $backup = "$Path.bak-$(Get-Date -Format yyyyMMdd-HHmmss)"
            Copy-Item $Path $backup -Force
            $existing = Get-Content $Path -Raw | ConvertFrom-Json
        } else {
            $existing = [pscustomobject]@{}
        }
        if (-not $existing.PSObject.Properties['mcpServers']) {
            $existing | Add-Member -NotePropertyName mcpServers -NotePropertyValue ([pscustomobject]@{}) -Force
        }
        $added = @()
        foreach ($name in $Servers.PSObject.Properties.Name) {
            if (-not $existing.mcpServers.PSObject.Properties[$name]) {
                $existing.mcpServers | Add-Member -NotePropertyName $name -NotePropertyValue $Servers.$name
                $added += $name
            }
        }
        ($existing | ConvertTo-Json -Depth 20) | Set-Content -Path $Path -Encoding utf8
        return $added
    }

    if ($HasCursor) {
        $added = Merge-McpFile -Path (Join-Path $Home1 '.cursor\mcp.json') -Servers $McpData.mcpServers
        Write-Host "[Cursor] ~/.cursor/mcp.json guncellendi. Yeni eklenenler: $($added -join ', ')"
    }
    if ($HasAntigravity) {
        $added1 = Merge-McpFile -Path (Join-Path $Home1 '.gemini\antigravity\mcp_config.json') -Servers $McpData.mcpServers
        $added2 = Merge-McpFile -Path (Join-Path $Home1 '.gemini\config\mcp_config.json') -Servers $McpData.mcpServers
        Write-Host "[Antigravity] ~/.gemini/antigravity/mcp_config.json & ~/.gemini/config/mcp_config.json guncellendi."
    }
    if ($HasClaudeDesktop) {
        $claudeDesktopPath = Join-Path $AppData1 'Claude\claude_desktop_config.json'
        $added = Merge-McpFile -Path $claudeDesktopPath -Servers $McpData.mcpServers
        Write-Host "[Claude Desktop] $claudeDesktopPath guncellendi. Yeni eklenenler: $($added -join ', ')"
    }
    if ($HasClaudeCode) {
        Write-Host "[Claude Code] MCP sunuculari 'claude mcp add-json ... -s user' ile ekleniyor..."
        foreach ($name in $McpData.mcpServers.PSObject.Properties.Name) {
            $json = ($McpData.mcpServers.$name | ConvertTo-Json -Depth 10 -Compress)
            & claude mcp add-json $name $json -s user 2>&1 | Out-Null
        }
        Write-Host "   -- mevcut olanlar atlandi, yeni olanlar eklendi."
    }
    Write-Host ""
    Write-Host "NOT: Dokploy/Supabase/n8n/SSH/Google servisleri icin API key'leri hala"
    Write-Host "     kendi panellerinden uretip ilgili config dosyasina elle doldurman gerekiyor."
    Write-Host ""
}

Write-Host "================================================================"
Write-Host " KAPSAM DISI (bu script'te yok):"
Write-Host " - cc-switch (masaustu app) -> https://github.com/farion1231/cc-switch/releases"
Write-Host " - ai-website-cloner-template -> https://github.com/JCodesMore/ai-website-cloner-template"
Write-Host " - Claude-Code-Game-Studios (proje sablonu) -> git clone ...Donchitos/Claude-Code-Game-Studios.git"
Write-Host " - Agent-Reach, win-dev-skills, awesome-claude-skills -> repo yolu dogrulanamadi/farkli mekanizma"
Write-Host " - caveman, security-review, mcp-builder ve digerleri -> zaten elinde"
Write-Host "================================================================"
