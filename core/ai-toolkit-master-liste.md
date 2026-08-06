# AI Toolkit — Master Liste (MCP + Skills)

Bu dosya Claude Code'a (veya Cursor/Antigravity terminaline) taşınıp oradan devam edilmek üzere hazırlandı.

---

## 1) MCP'ler (13)

| # | İsim | Tip | Kurulum | Notlar |
|---|------|-----|---------|--------|
| 1 | meta-ads | http | `"meta-ads": {"type":"http","url":"https://mcp.facebook.com/ads"}` | OAuth ilk kullanımda |
| 2 | github | http | `"github": {"type":"http","url":"https://api.githubcopilot.com/mcp/"}` | OAuth ilk kullanımda |
| 3 | huggingface | http | `"huggingface": {"type":"http","url":"https://huggingface.co/mcp"}` | OAuth ilk kullanımda |
| 4 | cloudflare | http | `"cloudflare": {"type":"http","url":"https://bindings.mcp.cloudflare.com/mcp"}` | OAuth ilk kullanımda |
| 5 | dokploy | stdio | `npx -y @dokploy/mcp` | env: `DOKPLOY_URL`, `DOKPLOY_API_KEY` (self-host panelinden) |
| 6 | supabase-selfhost | stdio | `npx -y selfhosted-supabase-mcp` | env: `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_ROLE_KEY` |
| 7 | n8n | stdio | `npx -y n8n-mcp` | env: `N8N_API_URL`, `N8N_API_KEY` |
| 8 | ssh | stdio | `npx -y ssh-mcp-server` | env: `SSH_HOST`, `SSH_USER`, `SSH_KEY_PATH` |
| 9 | yahoo-finance | stdio | `uvx mcp-yahoo-finance` | kişisel trading/izleme |
| 10 | google-analytics | stdio | `pipx run analytics-mcp` | Google OAuth gerekir |
| 11 | google-ads | stdio | `python -m google_ads_mcp` | Google OAuth gerekir |
| 12 | google-search-console | stdio | `npx -y mcp-gsc` | Google OAuth gerekir |
| 13 | google-tag-manager | http | `"google-tag-manager": {"type":"http","url":"https://gtm-mcp.stape.io/mcp"}` | OAuth ilk kullanımda |

**Tam config bloğu** (`.mcp.json` / `claude_desktop_config.json` / `.cursor/mcp.json`'a yapıştır):

```json
{
  "mcpServers": {
    "meta-ads": { "type": "http", "url": "https://mcp.facebook.com/ads" },
    "github": { "type": "http", "url": "https://api.githubcopilot.com/mcp/" },
    "huggingface": { "type": "http", "url": "https://huggingface.co/mcp" },
    "cloudflare": { "type": "http", "url": "https://bindings.mcp.cloudflare.com/mcp" },
    "dokploy": {
      "command": "npx", "args": ["-y", "@dokploy/mcp"],
      "env": { "DOKPLOY_URL": "https://SENIN-DOKPLOY-DOMAININ/api", "DOKPLOY_API_KEY": "SENIN_KEYIN" }
    },
    "supabase-selfhost": {
      "command": "npx", "args": ["-y", "selfhosted-supabase-mcp"],
      "env": { "SUPABASE_URL": "https://SENIN-SUPABASE-DOMAININ", "SUPABASE_ANON_KEY": "", "SUPABASE_SERVICE_ROLE_KEY": "" }
    },
    "n8n": {
      "command": "npx", "args": ["-y", "n8n-mcp"],
      "env": { "N8N_API_URL": "https://SENIN-N8N-DOMAININ", "N8N_API_KEY": "" }
    },
    "ssh": {
      "command": "npx", "args": ["-y", "ssh-mcp-server"],
      "env": { "SSH_HOST": "", "SSH_USER": "", "SSH_KEY_PATH": "" }
    },
    "yahoo-finance": { "command": "uvx", "args": ["mcp-yahoo-finance"] },
    "google-analytics": { "command": "pipx", "args": ["run", "analytics-mcp"] },
    "google-ads": { "command": "python", "args": ["-m", "google_ads_mcp"] },
    "google-search-console": { "command": "npx", "args": ["-y", "mcp-gsc"] },
    "google-tag-manager": { "type": "http", "url": "https://gtm-mcp.stape.io/mcp" }
  }
}
```

---

## 2) Skills (34) — kaynak repo + kurulum yöntemi

### A) `npx skills add` ile (global, otomatik tespit, doğrulanmış)

```bash
npx skills add nextlevelbuilder/ui-ux-pro-max-skill --skill ui-ux-pro-max -a claude-code -a cursor -a antigravity -g -y
npx skills add virgiliojr94/book-to-skill --all -a claude-code -a cursor -a antigravity -g -y
npx skills add K-Dense-AI/scientific-agent-skills --all -a claude-code -a cursor -a antigravity -g -y
npx skills add mukul975/Anthropic-Cybersecurity-Skills --all -a claude-code -a cursor -a antigravity -g -y
npx skills add twostraws/swiftui-agent-skill --all -a claude-code -a cursor -a antigravity -g -y
npx skills add chrisbanes/skills --all -a claude-code -a cursor -a antigravity -g -y
```

Zaten kurulanları güncellemek için: `npx skills update -g -y`

### B) Claude Code plugin marketplace (`claude` CLI içine yapıştır)

```
/plugin marketplace add nextlevelbuilder/ui-ux-pro-max-skill
/plugin install ui-ux-pro-max@ui-ux-pro-max-skill

/plugin marketplace add addyosmani/agent-skills
/plugin install agent-skills@addy-agent-skills

/plugin marketplace add AgriciDaniel/claude-seo
/plugin install claude-seo@agricidaniel-claude-seo

/plugin marketplace add coreyhaines31/marketingskills
/plugin install marketing-skills

/plugin marketplace add forrestchang/andrej-karpathy-skills
/plugin install andrej-karpathy-skills@karpathy-skills

/plugin marketplace add mvanhorn/last30days-skill
/plugin install last30days

/plugin marketplace add kepano/obsidian-skills
/plugin install obsidian@obsidian-skills

/plugin marketplace add Imbad0202/academic-research-skills
/plugin install academic-research-skills

/plugin marketplace add thedotmack/claude-mem
/plugin install claude-mem
```

### C) pip / npm (agent seçimi gerektirmez)

```bash
pip install graphifyy                  # graphify — bilgi grafiği
pip install "headroom-ai[all]"         # headroom — context sıkıştırma, sonra: headroom wrap claude
npx @santifer/career-ops init          # career-ops — DOĞRULANMAMIŞ, hata verirse repo teyit gerek
```

### D) Ajana mesaj olarak yapıştırılan kurulum

```
"bana Agent Reach'i kur: https://raw.githubusercontent.com/Panniantong/agent-reach/main/docs/install.md"
```

### E) Manuel indirme / şablon (komut değil)

```
cc-switch (masaüstü uygulama)          → https://github.com/farion1231/cc-switch/releases
ai-website-cloner-template (proje şablonu) → https://github.com/JCodesMore/ai-website-cloner-template  ("Use this template")
Claude-Code-Game-Studios (proje şablonu)   → git clone https://github.com/Donchitos/Claude-Code-Game-Studios.git
```

### F) Repo yolu doğrulanamadı — kurmadan önce README'den teyit et

```
microsoft/win-dev-skills
ComposioHQ/awesome-claude-skills   (tekil kurulum değil, dizin — içinden seç)
```

### G) Zaten elinde / ekstra kurulum gerekmez

```
caveman, security-review, mcp-builder, sentry-nextjs-sdk,
supabase/postgres-best-practices, cloudflare/wrangler,
react-native-best-practices, adspirer-ads-agent, marketing, engineering
```

---

## Kontrol listesi

- [ ] MCP JSON bloğunu `claude_desktop_config.json` / `.mcp.json` / `.cursor/mcp.json`'a ekle
- [ ] Dokploy/Supabase/n8n API key'lerini panellerden üret, doldur
- [ ] Google OAuth kurulumlarını tamamla (Analytics/Ads/GSC/GTM)
- [ ] Bölüm A: `npx skills add` komutlarını çalıştır (global, -g)
- [ ] Bölüm B: plugin marketplace komutlarını `claude` CLI içinde çalıştır
- [ ] Bölüm C: pip/npm paketlerini kur
- [ ] Bölüm D-E: manuel kurulumları yap
- [ ] Bölüm F: iki repo için güncel kurulum yolunu kendi README'sinden teyit et
