<div align="center">

  <h1>⚡ WyvDev</h1>
  <p><strong>The Open Source Local Developer PaaS & Universal MCP Orchestrator</strong></p>

  <p>
    Manage multi-IDE configurations, orchestrate MCP servers, track live TCP ports, and run local applications with 1-click self-healing auto-repair.
  </p>

  <p>
    <a href="https://github.com/imyigo/WyvDev/stargazers"><img src="https://img.shields.io/github/stars/imyigo/WyvDev?style=for-the-badge&logo=github&color=gold" alt="GitHub Stars"></a>
    <a href="https://github.com/imyigo/WyvDev/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.style=for-the-badge" alt="License"></a>
    <a href="https://github.com/imyigo/WyvDev"><img src="https://img.shields.io/badge/Status-Public%20Beta%20v0.9-orange?style=for-the-badge&logo=rocket" alt="Public Beta"></a>
    <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"></a>
    <a href="#"><img src="https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue?style=for-the-badge" alt="Platform"></a>
    <a href="#"><img src="https://img.shields.io/badge/i18n-English%20%7C%20Türkçe-purple?style=for-the-badge" alt="i18n"></a>
  </p>

  <br />
</div>

> [!IMPORTANT]
> ⚠️ **WyvDev is currently in Public Beta (v0.9).** Active development is underway. Your feedback, bug reports, and pull requests are highly appreciated!

---

## 💡 Why WyvDev?

Managing local AI agent skills, MCP (Model Context Protocol) servers, and microservice runtimes across different IDEs often leads to configuration drift, port collisions, broken dependencies, and missing tools.

**WyvDev** solves this by providing a unified, local-first control plane inspired by modern PaaS platforms like Dokploy and Coolify:

* 🔌 **Zero Config Drift:** Edit your MCP servers once in WyvDev, and sync seamlessly across 13+ IDEs and extensions.
* ⚡ **Local Micro-PaaS:** Automatically scans your `repo/` directory, detects runtimes (Node.js, Python, Rust, Docker), parses scripts, and runs projects in the background.
* 🟢 **Live Port & Process Control:** Instantly detects listening TCP ports (`Port: 3000`, `Port: 5173`) and provides 1-click process termination.
* 🔧 **Self-Healing Auto-Repair:** Automatically detects runtime startup failures and offers 1-click clean reinstallations (`npm install --force`, `pip install --force-reinstall`, `cargo clean`).

---

## ✨ Features at a Glance

### 🔌 Universal MCP & Multi-IDE Sync
WyvDev supports automatic detection, configuration, and state syncing across:
- **Cursor IDE** (`.cursor/mcp.json` & Cline Extension)
- **Claude Desktop App** (`claude_desktop_config.json`)
- **VS Code** (Cline & Roo Code Extensions, `.vscode/mcp.json`)
- **Windsurf IDE** (`mcp_config.json`)
- **Zed Editor** (`settings.json`)
- **Continue.dev** (`config.json`)
- **JetBrains IDEs** (`mcp.json`)
- **Antigravity IDE & CLI** (`mcp_config.json`)

### ⚡ Local PaaS Engine
- **Smart Runtime Auto-Detection:** Automatically identifies Node.js, Python, Rust, Docker, and Monorepos (`pnpm`, `npm`).
- **Smart Script Prioritization:** Intelligently selects `dev`, `dokploy:dev`, `start`, `serve`, `preview`, `main.py`, or `Cargo.toml`.
- **Repo Type Classification:** Auto-tags repositories as `⚡ Service / App`, `🔌 MCP Server`, `📄 Skill`, `📦 Library`, or `📁 Other`.

### 🛠️ 1-Click Engine Diagnostics
Scans your environment for required runtime tools and offers 1-click silent background installations:
- **Languages & Engine:** Node.js, Python 3, Rust / Cargo, Docker Engine.
- **Package & Isolation Managers:** `uv` / `uvx`, `pipx`, `pnpm`, `npx`, `git`.

### 💾 Backup & Session Persistence
- **1-Click IDE Config Backup:** Archives all active IDE configurations into timestamped `.zip` packages.
- **F5 Session Persistence:** Preserves active filters, sorting preferences, search results, view modes (Table vs Cards), and scroll position across page reloads.

---

## 📊 Feature Comparison Matrix

| Feature | WyvDev | Manual Editing | Docker Compose | Generic Process Managers |
| :--- | :---: | :---: | :---: | :---: |
| **Multi-IDE Sync (13+ IDEs)** | ✅ **Yes** | ❌ No | ❌ No | ❌ No |
| **Live TCP Port Auto-Discovery** | ✅ **Yes** | ❌ No | ⚠️ Port mapping only | ❌ No |
| **Self-Healing Auto-Repair** | ✅ **Yes** | ❌ No | ❌ No | ❌ No |
| **1-Click Missing Tool Install** | ✅ **Yes** | ❌ No | ❌ No | ❌ No |
| **Zero-Dependency Single Binary** | ✅ **Yes** | N/A | ❌ No | ⚠️ Partial |
| **Bilingual UI (EN / TR)** | ✅ **Yes** | ❌ No | ❌ No | ❌ No |

---

## 🚀 Quick Start (30 Seconds)

### Installation

#### Option 1: Download Pre-built Binary
Download the latest `wyvdev.exe` (or `wyvdev` binary) from the [Releases](https://github.com/imyigo/WyvDev/releases) page.

#### Option 2: Build from Source
Ensure you have [Go 1.22+](https://go.dev) installed:

```bash
# Clone the repository
git clone https://github.com/imyigo/WyvDev.git
cd WyvDev

# Build binary
go build -o wyvdev.exe main.go

# Run live server
.\wyvdev.exe
```

The Web UI automatically opens in your default browser at `http://127.0.0.1:47651/index.html`.

---

## 📡 API Overview

WyvDev features a high-performance Go REST API:

* `GET /api/state` — Retrieves global state bundle (MCP servers, IDE paths, tracked repos).
* `GET /api/ides/detect` — Auto-detects installed IDE configurations and filters active paths.
* `POST /api/ides/backup` — Creates a timestamped `.zip` backup of all active IDE configurations.
* `GET /api/repos/scan` — Scans `repo/` directory and returns runtime engines & repo types.
* `GET /api/apps/running` — Returns list of active background apps with PIDs & TCP listening ports.
* `POST /api/apps/{name}/kill` — Terminates process tree (`taskkill` / `docker stop`) and frees port.
* `GET /api/system/health` — Runs diagnostic health checks on 9 core runtime tools.
* `POST /api/system/install` — Triggers 1-click background installation for missing runtime tools.

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/imyigo/WyvDev/issues).

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

Distributed under the MIT License. See [`LICENSE`](LICENSE) for more information.

<div align="center">
  <br />
  <p>Crafted with ❤️ for AI Developers & Systems Engineers.</p>
</div>
