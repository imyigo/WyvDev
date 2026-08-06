# ⚡ WyvDev — Next-Gen Local Developer PaaS & Universal MCP Orchestrator

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue?style=for-the-badge)]()
[![i18n](https://img.shields.io/badge/i18n-English%20%7C%20Türkçe-purple?style=for-the-badge)]()

**WyvDev** is a powerful, lightweight, local-first Developer PaaS and Universal Model Context Protocol (MCP) Orchestrator. It automatically scans your repositories, manages multi-IDE configurations, detects live TCP ports, and provides 1-click installation, execution, and self-healing dependency repairs.

---

## 🔥 Key Features

* **🔌 Multi-IDE Configuration Sync:** Seamlessly syncs MCP server definitions across Cursor, Claude Desktop, VS Code (Cline/Roo Code), Windsurf, Zed Editor, Continue.dev, JetBrains, and Antigravity.
* **⚡ Dokploy-Style Local PaaS:** Automatically scans `repo/` projects, detects runtimes (Node.js, Python, Rust, Docker), parses start scripts (`dev`, `dokploy:dev`, `start`, `main.py`), and executes them in the background.
* **🟢 Live TCP Port & Process Monitoring:** Continuously tracks listening TCP ports (e.g. `Port: 3000`, `Port: 5173`) and provides 1-click process termination (`taskkill` / `docker stop`).
* **🔧 Self-Healing Auto-Repair:** If a project fails to start or crashes, WyvDev automatically reveals a `🔧 Onar (Repair)` button that performs clean forced reinstallations (`npm install --force`, `pip install --force-reinstall`, `cargo clean`).
* **🔍 System Health & Engine Diagnostics:** Tests 9 core runtime tools and provides 1-click background installations for missing dependencies (`pipx`, `uv`, `pnpm`, `winget`).
* **💾 IDE Config Backup:** 1-click zipped archiving of all detected IDE configuration files into timestamped backups.
* **🌐 Bilingual UI:** Full Turkish & English (TR/EN) toggle support with session state preservation across page reloads (F5).

---

## 🚀 Quick Start

### Prerequisites
* [Go 1.22+](https://go.dev) installed on your system.

### Build and Run

```bash
# Clone the repository
git clone https://github.com/imyigo/WyvDev.git
cd WyvDev

# Build the executable
go build -o wyvdev.exe main.go

# Run the live server
.\wyvdev.exe
```

The GUI web panel will automatically launch at `http://127.0.0.1:47651/index.html`.

---

## 🗂️ Project Structure

```
WyvDev/
├── main.go               # High-performance Go HTTP API & process manager
├── index.html            # Main Local Library & PaaS dashboard
├── search.html           # GitHub Live Search & Catalog
├── paths.html            # Dynamic IDE Paths & Backup manager
├── settings.html         # System Health & Engine Diagnostics
├── app.js                # Frontend state, i18n, & real-time polling
├── style.css             # Glassmorphism design tokens & styles
└── repo/                 # Local git repositories & projects directory
```

---

## 📄 License

This project is open-source and licensed under the [MIT License](LICENSE).
