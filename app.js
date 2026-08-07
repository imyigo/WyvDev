// WyvDev — Visual MCP Store & Local Developer PaaS Engine

// ---------- WyvDev i18n Translation Engine (TR / EN) ----------
const I18N_DICTIONARY = {
  tr: {
    nav_library: "Yerel Kütüphane",
    nav_loop: "🌀 Loop Engine",
    nav_search: "GitHub Arama",
    nav_paths: "IDE Dosya Yolları",
    nav_settings: "Sistem & Teşhis",
    backend_connected: "🟢 Backend Bağlı",
    backend_disconnected: "🔴 Backend Bağlı Değil",
    add_mcp_btn: "Yeni MCP Ekle",
    json_edit_btn: "JSON Düzenle / İndir",
    reset_defaults_btn: "Varsayılan Ayarlara Sıfırla",
    type_filter_label: "Tür:",
    sort_label: "Sırala:",
    filter_all: "Tüm Türler",
    filter_service: "⚡ Service & App",
    filter_mcp: "🔌 MCP Server",
    filter_skill: "📄 Skill",
    filter_lib: "📦 Library",
    filter_other: "📁 Diğer",
    sort_name: "İsme Göre (A-Z)",
    sort_running: "🟢 Önce Çalışanlar",
    sort_installed: "✓ Önce Yüklü Olanlar",
    sort_type: "Türe Göre",
    view_table: "Tablo",
    view_grid: "Kartlar",
    btn_start: "🚀 Başlat",
    btn_stop: "🛑 Durdur",
    btn_repair: "🔧 Onar",
    btn_install: "⚡ Kur",
    btn_installed: "✓ Yüklü",
    status_running: "🟢 Çalışıyor",
    status_stopped: "⚪ Durduruldu",
    status_error: "⚠️ Hata",
    toast_lang_changed: "🌐 Uygulama dili Türkçe olarak ayarlandı."
  },
  en: {
    nav_library: "Local Library",
    nav_loop: "🌀 Loop Engine",
    nav_search: "GitHub Search",
    nav_paths: "IDE File Paths",
    nav_settings: "System & Health",
    backend_connected: "🟢 Backend Connected",
    backend_disconnected: "🔴 Backend Disconnected",
    add_mcp_btn: "Add New MCP",
    json_edit_btn: "Edit / Download JSON",
    reset_defaults_btn: "Reset to Defaults",
    type_filter_label: "Type:",
    sort_label: "Sort By:",
    filter_all: "All Types",
    filter_service: "⚡ Service & App",
    filter_mcp: "🔌 MCP Server",
    filter_skill: "📄 Skill",
    filter_lib: "📦 Library",
    filter_other: "📁 Other",
    sort_name: "By Name (A-Z)",
    sort_running: "🟢 Running First",
    sort_installed: "✓ Installed First",
    sort_type: "By Type",
    view_table: "Table",
    view_grid: "Cards",
    btn_start: "🚀 Start",
    btn_stop: "🛑 Stop",
    btn_repair: "🔧 Repair",
    btn_install: "⚡ Install",
    btn_installed: "✓ Installed",
    status_running: "🟢 Running",
    status_stopped: "⚪ Stopped",
    status_error: "⚠️ Error",
    toast_lang_changed: "🌐 Switched language to English."
  }
};

let currentLang = localStorage.getItem('wyvdev_lang') || 'tr';

function t(key) {
  const dict = I18N_DICTIONARY[currentLang] || I18N_DICTIONARY.tr;
  return dict[key] || key;
}

function toggleAppLanguage() {
  currentLang = currentLang === 'tr' ? 'en' : 'tr';
  localStorage.setItem('wyvdev_lang', currentLang);
  updateLanguageUI();
  showToast(t('toast_lang_changed'));
}

function updateLanguageUI() {
  const langTextEl = document.getElementById('current-lang-text');
  if (langTextEl) langTextEl.innerText = currentLang.toUpperCase();

  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (key) {
      const val = t(key);
      if (val) el.textContent = val;
    }
  });

  if (document.getElementById('library-scan-body') || document.getElementById('library-view-grid-wrapper')) {
    renderLibraryScan();
  }
  if (document.getElementById('mcp-grid')) renderMcps();
  if (document.getElementById('skills-grid')) renderSkills();
  if (document.getElementById('ide-paths-grid')) renderIdePaths();
}

const DEFAULT_MCP_SERVERS = [
  { id: 'dokploy', name: 'Dokploy PaaS', type: 'stdio', command: 'npx', args: ['-y', '@dokploy/mcp'], desc: 'Dokploy PaaS self-host uygulama ve veritabanı paneli yönetimi.', category: 'DevOps', badge: 'Self-Host', icon: 'cloud-cog', iconColor: 'text-indigo-400', auth: true, env: { DOKPLOY_URL: 'https://dokploy.domain.com/api', DOKPLOY_API_KEY: '' } },
  { id: 'supabase-selfhost', name: 'Supabase Database', type: 'stdio', command: 'npx', args: ['-y', 'selfhosted-supabase-mcp'], desc: 'Self-hosted Supabase PostgreSQL, Auth ve Storage yönetimi.', category: 'Database', badge: 'Self-Host', icon: 'database', iconColor: 'text-emerald-400', auth: true, env: { SUPABASE_URL: '', SUPABASE_ANON_KEY: '', SUPABASE_SERVICE_ROLE_KEY: '' } },
  { id: 'n8n', name: 'n8n Automation', type: 'stdio', command: 'npx', args: ['-y', 'n8n-mcp'], desc: 'n8n workflow otomasyon senaryolarını tetikleme ve yönetme.', category: 'Automation', badge: 'Self-Host', icon: 'workflow', iconColor: 'text-amber-400', auth: true, env: { N8N_API_URL: '', N8N_API_KEY: '' } },
  { id: 'github', name: 'GitHub Copilot', type: 'http', url: 'https://api.githubcopilot.com/mcp/', desc: 'GitHub repoları, issue ve pull request işlemleri.', category: 'DevOps', badge: 'Resmi', icon: 'github', iconColor: 'text-gray-200', auth: true, headers: { Authorization: '' } },
  { id: 'cloudflare', name: 'Cloudflare Bindings', type: 'http', url: 'https://bindings.mcp.cloudflare.com/mcp', desc: 'Cloudflare Workers, KV ve D1 veritabanı yönetimi.', category: 'DevOps', badge: 'Resmi', icon: 'cloud-lightning', iconColor: 'text-orange-400', auth: true, headers: { Authorization: '' } },
  { id: 'meta-ads', name: 'Meta Ads Manager', type: 'http', url: 'https://mcp.facebook.com/ads', desc: 'Facebook ve Meta reklam kampanya metrikleri ve bütçeleri.', category: 'Marketing', badge: 'Resmi', icon: 'megaphone', iconColor: 'text-blue-400', auth: true, headers: { Authorization: '' } },
  { id: 'google-analytics', name: 'Google Analytics 4', type: 'stdio', command: 'pipx', args: ['run', 'analytics-mcp'], desc: 'GA4 ziyaretçi trafiği ve dönüşüm performans raporları.', category: 'Marketing', badge: 'Resmi', icon: 'bar-chart-3', iconColor: 'text-amber-400', auth: true, env: { GOOGLE_APPLICATION_CREDENTIALS: '' } },
  { id: 'google-search-console', name: 'Search Console GSC', type: 'stdio', command: 'npx', args: ['-y', 'mcp-gsc'], desc: 'Google Arama tıklama analizi ve indeksleme takibi.', category: 'Marketing', badge: 'Resmi', icon: 'search-code', iconColor: 'text-red-400', auth: true, env: { GOOGLE_APPLICATION_CREDENTIALS: '' } },
  { id: 'huggingface', name: 'HuggingFace Hub', type: 'http', url: 'https://huggingface.co/mcp', desc: 'HuggingFace açık kaynak LLM modelleri ve dataset araması.', category: 'AI & Models', badge: 'Resmi', icon: 'brain', iconColor: 'text-yellow-400', auth: true, headers: { Authorization: '' } },
  { id: 'ssh', name: 'SSH Remote Server', type: 'stdio', command: 'npx', args: ['-y', 'ssh-mcp-server'], desc: 'Uzak sunucularda güvenli terminal komutları çalıştırma.', category: 'DevOps', badge: 'Topluluk', icon: 'terminal', iconColor: 'text-purple-400', auth: true, env: { SSH_HOST: '', SSH_USER: '', SSH_KEY_PATH: '' } },
  { id: 'yahoo-finance', name: 'Yahoo Finance Borsa', type: 'stdio', command: 'uvx', args: ['mcp-yahoo-finance'], desc: 'Hisse senedi, kripto, döviz ve borsa canlı verileri.', category: 'Finance', badge: 'Topluluk', icon: 'line-chart', iconColor: 'text-emerald-400', auth: false }
];

const DEFAULT_RECOMMENDED_REPOS = [
  { id: 'ui-ux-pro-max', name: 'UI/UX Pro Max', repo: 'nextlevelbuilder/ui-ux-pro-max-skill', category: 'UI/UX', desc: 'İleri düzey tasarım sistemi, renk paletleri ve component rehberi.', extra: '--skill ui-ux-pro-max' },
  { id: 'book-to-skill', name: 'Book to Skill', repo: 'virgiliojr94/book-to-skill', category: 'Research', desc: 'Kitap ve dokümanlardan AI rehberleri üretme.', extra: '--all' },
  { id: 'scientific-agent-skills', name: 'Scientific Agent Skills', repo: 'K-Dense-AI/scientific-agent-skills', category: 'Research', desc: 'Akademik analiz, matematik ve veri analitiği.', extra: '--all' },
  { id: 'cybersecurity-skills', name: 'Anthropic Cybersecurity', repo: 'mukul975/Anthropic-Cybersecurity-Skills', category: 'Security', desc: 'Güvenlik analizi ve kod zaafiyet taraması.', extra: '--all' },
  { id: 'swiftui-agent-skill', name: 'SwiftUI Agent Skill', repo: 'twostraws/swiftui-agent-skill', category: 'Engineering', desc: 'iOS ve macOS için SwiftUI mimarisi ve bileşenleri.', extra: '--all' },
  { id: 'claude-seo', name: 'Claude SEO Expert', repo: 'AgriciDaniel/claude-seo', category: 'SEO', desc: 'Arama motoru optimizasyonu ve meta analizi.', extra: '--all' },
  { id: 'agent-skills', name: 'Agent Skills Collection', repo: 'addyosmani/agent-skills', category: 'Engineering', desc: 'Addy Osmani agent geliştirme araçları.', extra: '--all' },
  { id: 'andrej-karpathy-skills', name: 'Andrej Karpathy Skills', repo: 'forrestchang/andrej-karpathy-skills', category: 'Research', desc: 'Karpathy AI öğrenme yetenekleri.', extra: '--all' },
  { id: 'claude-mem', name: 'Claude Memory Hub', repo: 'thedotmack/claude-mem', category: 'Utility', desc: 'Sürekli hafıza ve context yönetim yeteneği.', extra: '--all' },
  { id: 'last30days-skill', name: 'Last 30 Days Skill', repo: 'mvanhorn/last30days-skill', category: 'Research', desc: 'Son 30 günlük trend ve gelişmeleri özetleme.', extra: '--all' },
  { id: 'obsidian-skills', name: 'Obsidian Notes Skill', repo: 'kepano/obsidian-skills', category: 'Utility', desc: 'Obsidian notları ve bilgi grafiklerini işleme.', extra: '--all' },
  { id: 'academic-research-skills', name: 'Academic Research Skills', repo: 'Imbad0202/academic-research-skills', category: 'Research', desc: 'Akademik makale ve literatür taraması.', extra: '--all' }
];

const DEFAULT_IDE_PATHS = [
  { id: 'antigravity-local', name: 'Antigravity IDE & CLI (Yerel)', path: 'C:\\Users\\Admin\\.gemini\\antigravity\\mcp_config.json', status: '✓ Algılandı' },
  { id: 'antigravity-global', name: 'Antigravity IDE & CLI (Global)', path: 'C:\\Users\\Admin\\.gemini\\config\\mcp_config.json', status: '✓ Algılandı' },
  { id: 'claude-desktop', name: 'Claude Desktop App (Windows)', path: 'C:\\Users\\Admin\\AppData\\Roaming\\Claude\\claude_desktop_config.json', status: '✓ Algılandı' },
  { id: 'cursor-ide', name: 'Cursor IDE (Windows)', path: 'C:\\Users\\Admin\\.cursor\\mcp.json', status: '✓ Algılandı' },
  { id: 'claude-code', name: 'Claude Code CLI (Windows)', path: 'C:\\Users\\Admin\\.claude.json', status: '✓ Algılandı' }
];

let selectedCategory = 'All';

function loadMcpServers() {
  const saved = localStorage.getItem('aitoolkit_mcp_servers');
  if (saved) { try { return JSON.parse(saved); } catch (e) {} }
  // No built-in external MCP catalog — the app only ever knows about MCPs
  // the user explicitly added (from a scanned repo or the "Yeni MCP Ekle"
  // modal). DEFAULT_MCP_SERVERS used to seed this on first load, which meant
  // 11 unrelated servers (Dokploy, Supabase, n8n, Meta Ads, ...) got synced
  // into every real IDE config without the user ever asking for them.
  return [];
}

function saveMcpServers(servers) {
  localStorage.setItem('aitoolkit_mcp_servers', JSON.stringify(servers));
  pushStateToBackend();
}

function loadRecommendedRepos() {
  const saved = localStorage.getItem('aitoolkit_recommended_repos');
  if (saved) { try { return JSON.parse(saved); } catch (e) {} }
  return DEFAULT_RECOMMENDED_REPOS;
}

function saveRecommendedRepos(repos) {
  localStorage.setItem('aitoolkit_recommended_repos', JSON.stringify(repos));
  pushStateToBackend();
}

function loadIdePaths() {
  const saved = localStorage.getItem('aitoolkit_ide_paths');
  if (saved) { try { return JSON.parse(saved); } catch (e) {} }
  // DEFAULT_IDE_PATHS is hardcoded Windows paths (C:\Users\Admin\...) — real
  // detection (autoDetectIdes / GET /api/state) is what actually populates
  // this; an empty starting point beats phantom paths that exist nowhere.
  return [];
}

function saveIdePaths(paths) {
  localStorage.setItem('aitoolkit_ide_paths', JSON.stringify(paths));
  pushStateToBackend();
}

let activeMcpServers = loadMcpServers();
let activeRecommendedRepos = loadRecommendedRepos();
let activeIdePaths = loadIdePaths();

// ---------- Go backend bridge ----------
// ai-toolkit.exe serves this page itself (same-origin) once launched. If the
// page is still opened via file:// (legacy), fall back to the fixed local port.
const API_BASE = (window.location.protocol === 'http:' || window.location.protocol === 'https:') ? '' : 'http://127.0.0.1:47651';

function setBackendStatus(online) {
  const pill = document.getElementById('backend-status-pill');
  if (!pill) return;
  if (online) {
    pill.textContent = t('backend_connected');
    pill.className = 'text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/15 text-emerald-400 border border-emerald-500/30';
  } else {
    pill.textContent = t('backend_disconnected');
    pill.className = 'text-[10px] font-bold px-2 py-0.5 rounded-full bg-red-500/15 text-red-400 border border-red-500/30';
  }
}

async function hydrateFromBackend(retryCount = 0) {
  try {
    const res = await fetch(`${API_BASE}/api/state`);
    if (!res.ok) throw new Error('bad status: ' + res.status);
    const data = await res.json();
    setBackendStatus(true);

    if (Array.isArray(data.mcpServers) && data.mcpServers.length) {
      activeMcpServers = data.mcpServers;
      localStorage.setItem('aitoolkit_mcp_servers', JSON.stringify(activeMcpServers));
    }
    if (Array.isArray(data.recommendedRepos) && data.recommendedRepos.length) {
      activeRecommendedRepos = data.recommendedRepos;
      localStorage.setItem('aitoolkit_recommended_repos', JSON.stringify(activeRecommendedRepos));
    }
    if (Array.isArray(data.idePaths) && data.idePaths.length) {
      activeIdePaths = data.idePaths;
      localStorage.setItem('aitoolkit_ide_paths', JSON.stringify(activeIdePaths));
    }

    try { if (typeof renderMcps === 'function') renderMcps(); } catch (err) {}
    try { if (typeof renderSkills === 'function') renderSkills(); } catch (err) {}
    try { if (typeof renderIdePaths === 'function') renderIdePaths(); } catch (err) {}
    try { if (document.getElementById('library-scan-body')) loadLibraryScan(); } catch (err) {}

    if (Array.isArray(data.prunedRepos) && data.prunedRepos.length) {
      showToast(`🧹 ${data.prunedRepos.length} öğe kaldırıldı — yerel repo/ klasörü silinmişti: ${data.prunedRepos.join(', ')}`);
    }

    if (window.lucide) lucide.createIcons();
  } catch (e) {
    if (retryCount < 5) {
      setTimeout(() => hydrateFromBackend(retryCount + 1), 400);
    } else {
      setBackendStatus(false);
    }
  }
}

let pushDebounceTimer = null;
function pushStateToBackend() {
  clearTimeout(pushDebounceTimer);
  pushDebounceTimer = setTimeout(async () => {
    try {
      const bundle = {
        mcpServers: activeMcpServers,
        // Backend auto-clones anything sent here with a repo field — only send
        // repos the user explicitly pulled from GitHub Arama (tagged 'GitHub'),
        // never the built-in default skill catalog, which is browse-only.
        recommendedRepos: activeRecommendedRepos.filter(s => s.category === 'GitHub'),
        idePaths: activeIdePaths
      };
      const res = await fetch(`${API_BASE}/api/state`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(bundle)
      });
      if (!res.ok) throw new Error('push failed');
      setBackendStatus(true);

      const data = await res.json();
      summarizeSyncResult(data);
    } catch (e) {
      setBackendStatus(false);
    }
  }, 400);
}

// Every panel edit auto-syncs to real IDE config files and (if a repo changed) may
// trigger a clone — this makes that otherwise-invisible background work visible.
function summarizeSyncResult(data) {
  const msgs = [];
  if (Array.isArray(data.sync) && data.sync.length) {
    const written = data.sync.filter(s => s.written).length;
    const skipped = data.sync.filter(s => s.skipped).length;
    const failed = data.sync.filter(s => s.error).length;
    if (written) msgs.push(`📡 ${written} IDE güncellendi`);
    if (skipped) msgs.push(`${skipped} atlandı (bulunamadı)`);
    if (failed) msgs.push(`⚠️ ${failed} yazma hatası`);
  }
  if (Array.isArray(data.cloned) && data.cloned.length) {
    const clonedCount = data.cloned.filter(c => c.cloned).length;
    if (clonedCount) msgs.push(`📦 ${clonedCount} repo klonlandı`);
  }
  if (msgs.length) showToast(msgs.join(' · '));
  if (document.getElementById('library-scan-body')) loadLibraryScan();
}

async function backupAllIdeConfigs(btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Yedekleniyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/ides/backup`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Yedekleme hatası');
    showToast(data.message || `💾 IDE ayarları yedeklendi.`);
  } catch (e) {
    showToast('⚠️ Yedekleme hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function autoDetectIdes(reset = false) {
  try {
    const url = reset ? `${API_BASE}/api/ides/detect?reset=true` : `${API_BASE}/api/ides/detect`;
    const res = await fetch(url);
    if (!res.ok) throw new Error('detect failed');
    const paths = await res.json();
    activeIdePaths = paths;
    localStorage.setItem('aitoolkit_ide_paths', JSON.stringify(activeIdePaths));
    renderIdePaths();
    setBackendStatus(true);
    showToast(reset ? '🔍 Tüm IDE\'ler yeniden tarandı.' : `🔍 Aktif IDE'ler güncellendi (${paths.length} aktif IDE).`);
  } catch (e) {
    setBackendStatus(false);
    showToast('⚠️ Otomatik tespit için backend sunucusuna ulaşılamadı. ai-toolkit.exe çalışıyor mu?');
  }
}

async function dangerDeleteIde(id, name, typedName) {
  if (typedName !== name) {
    showToast('⚠️ Yazdığınız isim IDE adıyla eşleşmiyor.');
    return;
  }
  if (!confirm(`'${name}' klasörünü KALICI OLARAK silmek üzeresiniz. Silmeden önce otomatik zip yedeği alınacak. Devam edilsin mi?`)) {
    return;
  }
  try {
    const res = await fetch(`${API_BASE}/api/ides/${id}/danger-delete`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ confirmName: name })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'silme başarısız');
    showToast(`🗑️ '${name}' klasörü silindi. Yedek: ${data.backup}`);
    autoDetectIdes();
  } catch (e) {
    showToast('⚠️ Silme hatası: ' + e.message);
  }
}

function showToast(msg) {
  let toast = document.getElementById('toast-msg');
  if (!toast) {
    toast = document.createElement('div');
    toast.id = 'toast-msg';
    toast.className = 'fixed bottom-6 right-6 z-50 bg-gray-900 border border-cyan-500/40 text-cyan-200 px-4 py-3 rounded-xl shadow-2xl flex items-center gap-3 animate-bounce text-xs font-medium';
    document.body.appendChild(toast);
  }
  toast.innerText = msg;
  toast.classList.remove('hidden');
  setTimeout(() => toast.classList.add('hidden'), 3500);
}

function filterMcpCategory(cat) {
  selectedCategory = cat;
  renderMcps();
}

async function deleteRepoFolder(name) {
  if (!confirm(`'${name}' klasörü VE tüm kayıtları yerel diskten (repo/${name}) KALICI OLARAK silinecek ve IDE konfigürasyonları güncellenecek. Devam edilsin mi?`)) {
    return;
  }
  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/delete`, { method: 'POST' });
    const text = await res.text();
    let data;
    try {
      data = JSON.parse(text);
    } catch (e) {
      data = { error: text || `Sunucu hatası (${res.status})` };
    }
    if (!res.ok) throw new Error(data.error || 'Silme başarısız');
    showToast(`🗑️ '${name}' klasörü diskten silindi ve tüm IDE'ler senkronize edildi.`);
    await hydrateFromBackend();
    if (document.getElementById('library-scan-body')) loadLibraryScan();
    if (document.getElementById('tracked-repos-body')) renderTrackedRepos();
  } catch (e) {
    showToast('⚠️ Silme hatası: ' + e.message);
  }
}

function removeRecommendedRepo(repoId) {
  const repoObj = activeRecommendedRepos.find(s => s.id === repoId);
  const repoName = repoObj && repoObj.repo ? repoObj.repo.split('/').pop() : (repoObj ? repoObj.name : repoId);
  deleteRepoFolder(repoName);
}

// Arama Kataloğu tek arama kutusu — backend'in /api/github/search proxy'sine gider.
// Boş sorguda backend claude-skill/mcp-server gibi etiketlerden güncel bir "keşfet" akışı döner.
let lastGithubResults = [];
let githubSearchDebounceTimer = null;

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str || '';
  return div.innerHTML;
}

function performLocalSearch(presetQuery) {
  const input = document.getElementById('gh-search-input');
  if (presetQuery !== undefined && input) input.value = presetQuery;
  const rawQuery = (presetQuery !== undefined ? presetQuery : (input ? input.value : '')) || '';

  clearTimeout(githubSearchDebounceTimer);
  githubSearchDebounceTimer = setTimeout(() => performGithubSearch(rawQuery.trim()), 400);
}

async function performGithubSearch(query) {
  const container = document.getElementById('gh-search-results');
  if (!container) return;

  container.innerHTML = `<p class="text-xs text-gray-400 col-span-full text-center py-8">${query ? 'GitHub aranıyor...' : 'Güncel repolar getiriliyor...'}</p>`;

  try {
    const res = await fetch(`${API_BASE}/api/github/search?q=${encodeURIComponent(query)}`);
    const data = await res.json().catch(() => ({}));

    if (res.status === 429 || res.status === 403 || data.error === 'rate_limited') {
      const rate = data.rateLimit || {};
      const resetSec = rate.resetSeconds !== undefined ? rate.resetSeconds : 60;
      const resetMin = Math.ceil(resetSec / 60);
      const used = rate.used !== undefined ? rate.used : (rate.limit || 10);
      const limit = rate.limit || 10;
      const remaining = rate.remaining !== undefined ? rate.remaining : 0;

      container.innerHTML = `
        <div class="col-span-full p-6 rounded-2xl bg-red-950/40 border border-red-500/40 text-red-200 text-center space-y-4 shadow-2xl animate-fade-in">
          <div class="flex items-center justify-center gap-2 font-bold text-sm text-red-300">
            <i data-lucide="shield-alert" class="w-5 h-5 text-red-400"></i>
            <span>GitHub Canlı Arama Limiti Doldu</span>
          </div>
          <p class="text-xs text-red-200/80 leading-relaxed max-w-lg mx-auto">
            GitHub anahtarsız genel aramada dakikada en fazla <strong>${limit} istek</strong> izni vermektedir. Kısa sürede çok fazla istek atıldığı için kota doldu.
          </p>

          <div class="inline-flex flex-wrap items-center justify-center gap-4 p-3.5 rounded-xl bg-gray-900/90 border border-red-500/30 font-mono text-xs text-gray-200 shadow-inner">
            <div>
              <span class="text-[10px] text-gray-400 block uppercase font-sans">Yapılan Arama</span>
              <span class="font-bold text-red-400 text-sm">${used} / ${limit}</span>
            </div>
            <div class="h-6 w-px bg-gray-800 hidden sm:block"></div>
            <div>
              <span class="text-[10px] text-gray-400 block uppercase font-sans">Kalan İstek</span>
              <span class="font-bold text-amber-400 text-sm">${remaining}</span>
            </div>
            <div class="h-6 w-px bg-gray-800 hidden sm:block"></div>
            <div>
              <span class="text-[10px] text-gray-400 block uppercase font-sans">Sıfırlanmaya Kalan Süre</span>
              <span class="font-bold text-cyan-400 text-sm">${resetSec} sn (${resetMin} dk)</span>
            </div>
          </div>

          <p class="text-[11px] text-gray-400">
            ⏱️ Kotanın sıfırlanması için yaklaşık <strong class="text-cyan-300">${resetSec} saniye (${resetMin} dakika)</strong> beklemeniz gerekmektedir.
          </p>
        </div>
      `;
      if (window.lucide) lucide.createIcons();
      return;
    }

    if (!res.ok) throw new Error(`Arama sunucusu hatası: ${res.status}`);

    const items = Array.isArray(data) ? data : (data.items || []);
    lastGithubResults = items;
    setBackendStatus(true);

    container.innerHTML = lastGithubResults.length ? lastGithubResults.map((repo, i) => {
      const repoId = repo.full_name.toLowerCase().replace(/[^a-z0-9]+/g, '-');
      const isInstalled = activeRecommendedRepos.some(s => s.repo === repo.full_name || s.id === repoId) ||
                          activeMcpServers.some(s => s.repo === repo.full_name || s.id === repoId) ||
                          lastScanResults.some(s => s.name === repo.name || (s.repo && s.repo === repo.full_name));

      return `
        <div class="glass-card flex flex-col justify-between space-y-3 border-gray-800 hover:border-purple-500/40">
          <div>
            <div class="flex items-center justify-between mb-2 gap-2">
              <div class="flex items-center gap-2 min-w-0">
                <img src="${repo.owner.avatar_url}" class="w-5 h-5 rounded-full shrink-0" alt="" />
                <h3 class="font-semibold text-sm text-gray-100 truncate" title="${escapeHtml(repo.full_name)}">${escapeHtml(repo.full_name)}</h3>
              </div>
              <span class="text-[9px] font-semibold px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20 shrink-0">⭐ ${repo.stargazers_count}</span>
            </div>
            <p class="text-xs text-gray-400 leading-relaxed mb-2">${repo.description ? escapeHtml(repo.description) : 'Açıklama yok.'}</p>
          </div>
          <div class="pt-3 border-t border-gray-800 flex items-center gap-2">
            ${isInstalled ? `
              <span class="flex-1 py-2 px-3 rounded-xl text-xs font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex items-center justify-center gap-1.5 select-none">
                <i data-lucide="check-circle-2" class="w-4 h-4 text-emerald-400"></i> Kuruldu
              </span>
            ` : `
              <button onclick="downloadRepoFromGithub(${i})" class="flex-1 py-2 px-3 rounded-xl text-xs font-semibold bg-purple-600 hover:bg-purple-500 text-white flex items-center justify-center gap-1.5 cursor-pointer shadow-lg transition-all active:scale-95">
                <i data-lucide="download" class="w-4 h-4"></i> İndir
              </button>
            `}
            <a href="${repo.html_url}" target="_blank" rel="noreferrer" class="p-2 rounded-xl text-xs bg-gray-800 hover:bg-gray-700 text-gray-300 flex items-center justify-center cursor-pointer">
              <i data-lucide="external-link" class="w-4 h-4"></i>
            </a>
          </div>
        </div>
      `;
    }).join('') : `<p class="text-xs text-gray-400 col-span-full text-center py-8">Sonuç bulunamadı.</p>`;

    if (window.lucide) lucide.createIcons();
  } catch (err) {
    setBackendStatus(false);
    container.innerHTML = `<p class="text-xs text-red-400 col-span-full text-center py-8">Arama sunucusuna ulaşılamıyor. ai-toolkit.exe çalışıyor mu?</p>`;
  }
}

function downloadRepoFromGithub(index) {
  const repo = lastGithubResults[index];
  if (!repo) return;
  const id = repo.full_name.toLowerCase().replace(/[^a-z0-9]+/g, '-');
  if (activeRecommendedRepos.some(s => s.repo === repo.full_name || s.id === id)) {
    showToast(`'${repo.name}' zaten katalogda ekli ve indirilmiş!`);
    return;
  }
  activeRecommendedRepos.push({
    id, name: repo.name, repo: repo.full_name,
    category: 'GitHub', desc: repo.description || `${repo.name} deposu.`, extra: '--all'
  });
  saveRecommendedRepos(activeRecommendedRepos);
  renderSkills();
  showToast(`🚀 '${repo.name}' indiriliyor — repo/ klasörüne klonlanıp takip edilecek.`);
}

function addSkillFromGithub(index) {
  downloadRepoFromGithub(index);
}

function addMcpFromGithub(index) {
  downloadRepoFromGithub(index);
}

function renderMcps() {
  const container = document.getElementById('mcp-grid');
  const countBadge = document.getElementById('mcp-count-badge');
  if (countBadge) countBadge.innerText = activeMcpServers.length;

  const filterBar = document.getElementById('mcp-category-filter');
  if (filterBar) {
    const categories = ['All', ...new Set(activeMcpServers.map(s => s.category).filter(Boolean))];
    filterBar.innerHTML = categories.map(cat => `
      <button onclick="filterMcpCategory('${cat}')" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold cursor-pointer transition-all ${selectedCategory === cat ? 'bg-cyan-600 text-white' : 'bg-gray-800 text-gray-400 hover:text-gray-200'}">${cat === 'All' ? 'Tümü' : escapeHtml(cat)}</button>
    `).join('');
  }

  if (!container) return;

  const filtered = selectedCategory === 'All' ? activeMcpServers : activeMcpServers.filter(s => s.category === selectedCategory);

  container.innerHTML = filtered.map(s => {
    const fields = s.env ? Object.keys(s.env) : (s.headers ? Object.keys(s.headers) : []);
    const hasKeys = fields.length > 0;
    const badgeColor = s.badge === 'Resmi' ? 'bg-cyan-500/15 text-cyan-400 border-cyan-500/30' : (s.badge === 'Self-Host' ? 'bg-indigo-500/15 text-indigo-300 border-indigo-500/30' : 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30');

    return `
      <div id="mcp-card-${s.id}" class="glass-card flex flex-col justify-between space-y-3">
        <div>
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center gap-2">
              <i data-lucide="${s.icon || 'server'}" class="w-5 h-5 ${s.iconColor || 'text-cyan-400'}"></i>
              <h3 class="font-semibold text-sm text-gray-100">
                ${s.name}
              </h3>
            </div>
            <div class="flex items-center gap-1.5">
              <span class="text-[9px] font-semibold px-2 py-0.5 rounded uppercase border ${badgeColor}">${s.badge || s.type}</span>
              <button onclick="removeMcp('${s.id}')" title="MCP Sunucusunu Sil" class="p-1 text-gray-500 hover:text-red-400 hover:bg-red-500/10 rounded cursor-pointer transition-colors">
                <i data-lucide="trash-2" class="w-3.5 h-3.5"></i>
              </button>
            </div>
          </div>
          <p class="text-xs text-gray-400 leading-relaxed mb-2">${s.desc || 'Açıklama girilmedi.'}</p>
          
          <div class="text-[11px] font-mono text-gray-500 border-t border-gray-800/80 pt-2 truncate">
            ${s.url ? `URL: <span class="text-cyan-400">${s.url}</span>` : `Komut: <span class="text-emerald-400">${s.command || ''} ${(s.args || []).join(' ')}</span>`}
          </div>
        </div>

        ${hasKeys ? `
          <div class="p-3 rounded-xl bg-gray-950/60 border border-amber-500/20 space-y-2 mt-2">
            <p class="text-[10px] font-semibold text-amber-400 flex items-center gap-1">
              <i data-lucide="key-round" class="w-3 h-3"></i> API Key & Parametreler
            </p>
            ${fields.map(f => {
              const val = s.env ? (s.env[f] || '') : (s.headers[f] || '');
              return `
                <div class="space-y-1">
                  <label class="text-[9px] font-mono text-gray-400 block">${f}:</label>
                  <input type="text" value="${val}" onchange="updateKey('${s.id}', '${f}', this.value)" placeholder="Anahtar veya URL..." class="glass-input text-xs py-1" />
                </div>
              `;
            }).join('')}
          </div>
        ` : ''}
      </div>
    `;
  }).join('');
  if (window.lucide) lucide.createIcons();

  const scrollTargetId = sessionStorage.getItem('wyvdev_scroll_to_mcp');
  if (scrollTargetId) {
    const card = document.getElementById(`mcp-card-${scrollTargetId}`);
    if (card) {
      sessionStorage.removeItem('wyvdev_scroll_to_mcp');
      card.scrollIntoView({ behavior: 'smooth', block: 'center' });
      card.classList.add('ring-2', 'ring-cyan-400', 'ring-offset-2', 'ring-offset-gray-950');
      setTimeout(() => card.classList.remove('ring-2', 'ring-cyan-400', 'ring-offset-2', 'ring-offset-gray-950'), 2500);
    }
  }
}

// Jumps from a repo card's "MCP ekli — Düzenle" to its variable editor on
// the IDE Dosya Yolları page instead of leaving it as a dead-end label.
// Opens an in-place popup with the MCP's editable variables — MCPs are only
// ever added from the repo scan or the header modal; editing an existing one
// happens right here, not on a separate IDE Dosya Yolları page.
function goToMcpConfig(id) {
  openMcpEditModal(id);
}

async function testMcpConnection(id, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3 h-3 animate-spin"></i> Test ediliyor...';
    if (window.lucide) lucide.createIcons();
  }
  try {
    const res = await fetch(`${API_BASE}/api/mcp/test`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Test başarısız');
    showToast(data.ok ? (data.message || '✅ Bağlantı başarılı') : (data.message || '⚠️ Bağlantı başarısız'));
  } catch (e) {
    showToast('⚠️ Test hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

function openMcpEditModal(id) {
  const s = activeMcpServers.find(m => m.id === id);
  const modal = document.getElementById('mcp-edit-modal');
  const body = document.getElementById('mcp-edit-body');
  const title = document.getElementById('mcp-edit-title');
  if (!s || !modal || !body) return;

  if (title) title.textContent = s.name || s.id;
  modal.dataset.mcpId = id;

  // Repo-detected MCPs (addMcpFromScan) don't ship a known env schema — we
  // can't know a package needs e.g. DOKPLOY_URL until it actually throws for
  // a missing one. So always offer an "add variable" row, not just editing
  // whatever env keys happen to already exist.
  const fields = s.env ? Object.keys(s.env) : (s.headers ? Object.keys(s.headers) : []);
  body.innerHTML = `
    <p class="text-gray-400 leading-relaxed">${escapeHtml(s.desc || 'Açıklama girilmedi.')}</p>
    <div class="text-[11px] font-mono text-gray-500 border-t border-gray-800/80 pt-2">
      ${s.url ? `URL: <span class="text-cyan-400">${escapeHtml(s.url)}</span>` : `Komut: <span class="text-emerald-400">${escapeHtml(s.command || '')} ${escapeHtml((s.args || []).join(' '))}</span>`}
      ${s.cwd ? `<div class="mt-1">Dizin: <span class="text-gray-400">${escapeHtml(s.cwd)}</span></div>` : ''}
    </div>
    <div class="p-3 rounded-xl bg-gray-950/60 border border-amber-500/20 space-y-2 mt-2">
      <p class="text-[10px] font-semibold text-amber-400 flex items-center gap-1">
        <i data-lucide="key-round" class="w-3 h-3"></i> API Key & Parametreler
      </p>
      ${fields.length ? fields.map(f => {
        const val = s.env ? (s.env[f] || '') : (s.headers[f] || '');
        return `
          <div class="flex items-end gap-1.5">
            <div class="flex-1 space-y-1">
              <label class="text-[9px] font-mono text-gray-400 block">${escapeHtml(f)}:</label>
              <input type="text" data-env-key="${escapeHtml(f)}" value="${escapeHtml(val)}" onchange="updateKey('${s.id}', '${f}', this.value)" placeholder="Anahtar veya URL..." class="glass-input text-xs py-1" />
            </div>
            <button onclick="removeMcpEnvVar('${s.id}', '${f}')" title="Bu değişkeni sil" class="p-1.5 rounded-lg text-gray-500 hover:text-red-400 hover:bg-red-500/10 cursor-pointer"><i data-lucide="x" class="w-3.5 h-3.5"></i></button>
          </div>
        `;
      }).join('') : `<p class="text-[11px] text-gray-500">Henüz değişken eklenmedi — paket çalışma zamanında hata veriyorsa (ör. "Environment variable X is not defined"), o ismi aşağıya ekleyin.</p>`}
      <div id="mcp-env-suggestions"></div>
      <div class="flex items-end gap-1.5 pt-2 border-t border-gray-800/60">
        <div class="flex-1 space-y-1">
          <label class="text-[9px] font-mono text-gray-400 block">Yeni değişken adı:</label>
          <input id="mcp-new-var-key" type="text" placeholder="ör. API_KEY" class="glass-input text-xs py-1 font-mono" />
        </div>
        <div class="flex-1 space-y-1">
          <label class="text-[9px] font-mono text-gray-400 block">Değer:</label>
          <input id="mcp-new-var-value" type="text" placeholder="değer..." class="glass-input text-xs py-1" />
        </div>
        <button onclick="addMcpEnvVar('${s.id}')" title="Değişken ekle" class="p-1.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white cursor-pointer"><i data-lucide="plus" class="w-3.5 h-3.5"></i></button>
      </div>
    </div>
  `;
  modal.classList.remove('hidden');
  if (window.lucide) lucide.createIcons();

  const repoName = repoNameFromCwd(s.cwd);
  if (repoName) loadMcpEnvSuggestions(repoName, id);
}

// Repo-detected MCPs know their source folder (cwd) but not which env vars
// the package actually reads — scan the repo's own source for env-read
// idioms (process.env.X, os.getenv(), etc.) instead of making the user
// guess names blind from a runtime crash message.
function repoNameFromCwd(cwd) {
  if (!cwd) return null;
  const parts = cwd.split(/[\\/]/).filter(Boolean);
  return parts.length ? parts[parts.length - 1] : null;
}

async function loadMcpEnvSuggestions(repoName, mcpId) {
  const el = document.getElementById('mcp-env-suggestions');
  if (!el) return;
  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(repoName)}/env-vars`);
    if (!res.ok) return;
    const data = await res.json();
    const server = activeMcpServers.find(m => m.id === mcpId);
    const existing = server && server.env ? Object.keys(server.env) : [];
    const suggestions = (data.vars || []).filter(v => !existing.includes(v));
    if (!suggestions.length) return;
    el.innerHTML = `
      <p class="text-[10px] text-gray-500 mt-1 mb-1">Kaynak kodda tespit edilen olası değişkenler (tıkla, eklensin):</p>
      <div class="flex flex-wrap gap-1">
        ${suggestions.map(v => `<button onclick="addMcpEnvVar('${mcpId}', '${v}')" class="px-2 py-0.5 rounded-md text-[10px] font-mono bg-indigo-900/40 hover:bg-indigo-800/60 text-indigo-300 border border-indigo-500/30 cursor-pointer">+ ${escapeHtml(v)}</button>`).join('')}
      </div>
    `;
    if (window.lucide) lucide.createIcons();
  } catch (e) {
    // best-effort only — silently leave the manual-entry row as the fallback
  }
}

function addMcpEnvVar(id, presetKey) {
  const keyInput = document.getElementById('mcp-new-var-key');
  const valueInput = document.getElementById('mcp-new-var-value');
  const key = (presetKey || keyInput?.value || '').trim();
  if (!key) {
    showToast('⚠️ Değişken adı boş olamaz.');
    return;
  }
  const server = activeMcpServers.find(s => s.id === id);
  if (!server) return;
  if (!server.env) server.env = {};
  server.env[key] = presetKey ? (server.env[key] || '') : (valueInput?.value || '');
  saveMcpServers(activeMcpServers);
  showToast(`'${key}' değişkeni eklendi.`);
  openMcpEditModal(id);
}

function removeMcpEnvVar(id, key) {
  const server = activeMcpServers.find(s => s.id === id);
  if (!server || !server.env) return;
  delete server.env[key];
  saveMcpServers(activeMcpServers);
  openMcpEditModal(id);
}

function closeMcpEditModal() {
  const modal = document.getElementById('mcp-edit-modal');
  if (modal) modal.classList.add('hidden');
}

// Explicit "Kaydet" — flushes every field (in case one's still focused and
// hasn't fired its own onchange yet), then pushes immediately and shows the
// real per-IDE sync result, instead of relying on the silent debounced push.
async function saveMcpEditModal(id) {
  const server = activeMcpServers.find(s => s.id === id);
  if (!server) return;

  document.querySelectorAll('#mcp-edit-body input[data-env-key]').forEach(input => {
    if (!server.env) server.env = {};
    server.env[input.dataset.envKey] = input.value;
  });
  localStorage.setItem('aitoolkit_mcp_servers', JSON.stringify(activeMcpServers));

  const btn = document.querySelector('#mcp-edit-modal button[onclick*="saveMcpEditModal"]');
  const originalLabel = btn ? btn.innerHTML : null;
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Kaydediliyor...';
    if (window.lucide) lucide.createIcons();
  }
  try {
    const syncData = await pushStateNow();
    closeMcpEditModal();
    showToast(`✓ ${server.name || server.id} MCP ayarları kaydedildi.`);
    summarizeSyncResult(syncData);
  } catch (e) {
    showToast('⚠️ Kaydetme hatası: ' + e.message);
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

// Immediate (non-debounced) push — used wherever the user takes an explicit
// "save"/"sync" action and expects to see the real result right away, unlike
// pushStateToBackend's silent 400ms-debounced background push.
async function pushStateNow() {
  const res = await fetch(`${API_BASE}/api/state`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      mcpServers: activeMcpServers,
      recommendedRepos: activeRecommendedRepos.filter(s => s.category === 'GitHub'),
      idePaths: activeIdePaths
    })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Senkronizasyon başarısız');
  return data;
}

async function syncMcpsToIdes(btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Senkronize ediliyor...';
    if (window.lucide) lucide.createIcons();
  }
  try {
    summarizeSyncResult(await pushStateNow());
  } catch (e) {
    showToast('⚠️ MCP senkronizasyon hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function syncSkillsToIdes(btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Senkronize ediliyor...';
    if (window.lucide) lucide.createIcons();
  }
  try {
    const res = await fetch(`${API_BASE}/api/skills/sync-all`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Senkronizasyon başarısız');
    showToast(data.message || `✅ ${data.skillsSynced} skill senkronize edildi.`);
  } catch (e) {
    showToast('⚠️ Skill senkronizasyon hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function installSkillCommand(repo, extra = '--all', btnEl) {
  if (!repo) {
    showToast('Bu skill yerel bir klasör — npx kurulumu gerekmiyor, repo/ altında zaten hazır.');
    return;
  }

  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Kuruluyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/skills/install`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ repo, extra })
    });
    const data = await res.json();
    showInstallResultModal(repo, data.command, data.output, !!data.ok, data.error);
  } catch (e) {
    showInstallResultModal(repo, `npx skills add ${repo} ${extra} -g -y`, '', false, 'Backend sunucusuna ulaşılamadı. ai-toolkit.exe çalışıyor mu?');
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

function showInstallResultModal(repo, command, output, ok, error) {
  const modal = document.getElementById('install-result-modal');
  if (!modal) {
    showToast(ok ? `✅ '${repo}' kuruldu.` : `⚠️ '${repo}' kurulamadı: ${error || ''}`);
    return;
  }
  document.getElementById('install-result-title').innerHTML = `<i data-lucide="package" class="w-4 h-4"></i> ${ok ? `✅ '${repo}' kuruldu` : `⚠️ '${repo}' kurulamadı`}`;
  document.getElementById('install-result-command').textContent = command || '';
  document.getElementById('install-result-output').textContent = [output, error].filter(Boolean).join('\n\n') || '(çıktı yok)';
  modal.classList.remove('hidden');
  if (window.lucide) lucide.createIcons();
}

function closeInstallResultModal() {
  const modal = document.getElementById('install-result-modal');
  if (modal) modal.classList.add('hidden');
}

// ============================================================
// 🚀 GELİŞMİŞ KURULUM MERKEZİ
// ============================================================

let _aimCurrentRepo = null;
let _aimAnalysis = null;
let _aimStepStates = {};       // stepId → 'idle' | 'running' | 'done' | 'error'
let _aimCurrentSSE = null;     // active EventSource
let _aimPipelineQueue = [];    // steps waiting in pipeline
let _aimRunningPipeline = false;

const RUNTIME_META = {
  node:     { icon: '🟩', label: 'Node.js',       color: 'text-green-400' },
  python:   { icon: '🐍', label: 'Python',         color: 'text-yellow-400' },
  rust:     { icon: '🦀', label: 'Rust',           color: 'text-orange-400' },
  go:       { icon: '🐹', label: 'Go',             color: 'text-cyan-400' },
  docker:   { icon: '🐳', label: 'Docker',         color: 'text-blue-400' },
  composer: { icon: '🐘', label: 'PHP Composer',   color: 'text-indigo-400' },
  gem:      { icon: '💎', label: 'Ruby Gems',      color: 'text-red-400' },
  maven:    { icon: '☕', label: 'Maven',           color: 'text-orange-500' },
  gradle:   { icon: '🐘', label: 'Gradle',         color: 'text-teal-400' },
  make:     { icon: '⚙️', label: 'Makefile',       color: 'text-gray-400' },
};

const PM_META = {
  pnpm: { icon: '⚡', label: 'pnpm', color: 'text-yellow-300' },
  yarn: { icon: '🧶', label: 'Yarn', color: 'text-blue-300' },
  bun:  { icon: '🍞', label: 'Bun',  color: 'text-amber-300' },
  npm:  { icon: '📦', label: 'npm',  color: 'text-red-300' },
};

async function showAdvancedInstallModal(repoName) {
  _aimCurrentRepo = repoName;
  _aimAnalysis = null;
  _aimStepStates = {};
  _aimPipelineQueue = [];
  _aimRunningPipeline = false;
  if (_aimCurrentSSE) { _aimCurrentSSE.close(); _aimCurrentSSE = null; }

  const modal = document.getElementById('advanced-install-modal');
  if (!modal) { showToast('Kurulum Merkezi modal bulunamadı.'); return; }

  // Reset UI
  document.getElementById('aim-repo-badge').textContent = repoName;
  document.getElementById('aim-loading').classList.remove('hidden');
  document.getElementById('aim-info').classList.add('hidden');
  document.getElementById('aim-steps').innerHTML = '<div class="text-xs text-gray-600 italic">Analiz bekleniyor...</div>';
  document.getElementById('aim-terminal').innerHTML = '<span class="text-gray-600 italic">Kurulum başlatmak için bir adım seçin...</span>';
  document.getElementById('aim-status-text').textContent = '';
  document.getElementById('aim-run-all-btn').disabled = false;
  document.getElementById('aim-stop-btn').classList.add('hidden');

  modal.classList.remove('hidden');
  if (window.lucide) lucide.createIcons();

  // Fetch analysis
  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(repoName)}/analyze`);
    if (!res.ok) throw new Error(await res.text());
    _aimAnalysis = await res.json();
    renderAimAnalysis(_aimAnalysis);
  } catch (e) {
    document.getElementById('aim-loading').innerHTML =
      `<div class="text-xs text-red-400 text-center">Analiz başarısız:<br>${escapeHtml(e.message)}</div>`;
  }
}

function renderAimAnalysis(a) {
  document.getElementById('aim-loading').classList.add('hidden');
  document.getElementById('aim-info').classList.remove('hidden');

  // Runtimes
  const rtContainer = document.getElementById('aim-runtimes');
  const rtItems = (a.runtimes || []).map(rt => {
    const m = RUNTIME_META[rt] || { icon: '⚙️', label: rt, color: 'text-gray-300' };
    return `<div class="flex items-center gap-1.5 text-xs ${m.color}"><span>${m.icon}</span><span class="font-semibold">${m.label}</span></div>`;
  });
  // Show package manager badge if Node.js detected
  if (a.packageManager && (a.runtimes || []).includes('node')) {
    const pm = PM_META[a.packageManager] || { icon: '📦', label: a.packageManager, color: 'text-gray-300' };
    rtItems.push(`<div class="flex items-center gap-1.5 text-xs ${pm.color} border border-gray-700 rounded px-1.5 py-0.5 w-fit">${pm.icon} <span class="font-mono">${pm.label}</span></div>`);
  }
  rtContainer.innerHTML = rtItems.length ? rtItems.join('') : '<div class="text-xs text-gray-600 italic">Tespit edilemedi</div>';

  // Deps
  document.getElementById('aim-deps').textContent =
    a.totalDeps > 0 ? `${a.installedDeps} / ${a.totalDeps} kurulu` : 'Sayılabilir bağımlılık yok';
  document.getElementById('aim-progress-bar').style.width = `${Math.min(a.installPercent, 100)}%`;
  document.getElementById('aim-progress-label').textContent = `%${a.installPercent} kurulu`;

  // Disk
  if (a.diskEstimateMB > 0) {
    document.getElementById('aim-disk-row').classList.remove('hidden');
    document.getElementById('aim-disk').textContent = `~${a.diskEstimateMB} MB`;
  }

  // Port
  if (a.portSuggestion > 0) {
    document.getElementById('aim-port-row').classList.remove('hidden');
    document.getElementById('aim-port').textContent = `:${a.portSuggestion}`;
  }

  // Extras
  const extras = document.getElementById('aim-extras');
  const extraItems = [];
  if (a.hasDockerfile) extraItems.push(`<div class="text-[10px] text-blue-400 flex items-center gap-1">🐳 Dockerfile mevcut</div>`);
  if (a.hasCompose)    extraItems.push(`<div class="text-[10px] text-blue-400 flex items-center gap-1">🐳 Compose mevcut</div>`);
  if (a.hasMakefile)   extraItems.push(`<div class="text-[10px] text-gray-400 flex items-center gap-1">⚙️ Makefile mevcut</div>`);
  extras.innerHTML = extraItems.join('');

  // ENV
  if (a.envVarsNeeded && a.envVarsNeeded.length > 0) {
    document.getElementById('aim-env-section').classList.remove('hidden');
    document.getElementById('aim-env-list').innerHTML = a.envVarsNeeded.map(k =>
      `<div class="text-[10px] font-mono text-amber-400 truncate" title="${escapeHtml(k)}">${escapeHtml(k)}</div>`
    ).join('');
  }

  // Steps
  renderAimSteps(a.installSteps || []);
}

function renderAimSteps(steps) {
  const container = document.getElementById('aim-steps');
  if (!steps.length) {
    container.innerHTML = '<div class="text-xs text-emerald-400">✅ Kurulum adımı gerekmiyor — zaten hazır!</div>';
    return;
  }

  container.innerHTML = steps.map(step => {
    const state = _aimStepStates[step.id] || 'idle';
    const stateIcon = { idle: '⏳', running: '⚙️', done: '✅', error: '❌' }[state] || '⏳';
    const stateCls  = { idle: 'border-gray-700 text-gray-300', running: 'border-indigo-500 text-indigo-200 bg-indigo-950/30', done: 'border-emerald-700 text-emerald-300 bg-emerald-950/20', error: 'border-red-700 text-red-300 bg-red-950/20' }[state] || 'border-gray-700 text-gray-300';
    const badge = step.required
      ? `<span class="text-[9px] px-1 py-0.5 rounded bg-rose-900/50 text-rose-300 border border-rose-700/50">Zorunlu</span>`
      : `<span class="text-[9px] px-1 py-0.5 rounded bg-gray-800 text-gray-500 border border-gray-700">Opsiyonel</span>`;
    return `
      <div class="flex items-center justify-between gap-2 px-2.5 py-2 rounded-lg border ${stateCls} transition-all">
        <div class="flex items-center gap-2 min-w-0">
          <span class="text-sm shrink-0">${stateIcon}</span>
          <div class="min-w-0">
            <div class="text-xs font-semibold truncate">${escapeHtml(step.label)}</div>
            <div class="text-[10px] font-mono text-gray-500 truncate">${escapeHtml(step.cmd)}</div>
          </div>
        </div>
        <div class="flex items-center gap-1.5 shrink-0">
          ${badge}
          ${state !== 'running' ? `<button onclick="runSingleInstallStep('${escapeHtml(step.id)}', '${escapeHtml(step.label)}')" class="px-2 py-0.5 rounded text-[10px] font-bold bg-gray-800 hover:bg-indigo-900/60 text-gray-300 hover:text-indigo-200 border border-gray-700 cursor-pointer transition-all">▶ Çalıştır</button>` : `<span class="text-[10px] text-indigo-400 animate-pulse">Çalışıyor...</span>`}
        </div>
      </div>`;
  }).join('');
}

function aimTerminalAppend(text, cls = '') {
  const term = document.getElementById('aim-terminal');
  if (!term) return;
  // Clear placeholder
  const placeholder = term.querySelector('span.text-gray-600.italic');
  if (placeholder) placeholder.remove();

  const line = document.createElement('div');
  if (cls) line.className = cls;

  // Color error/warning lines
  const lc = text.toLowerCase();
  if (lc.includes('error') || lc.includes('❌') || lc.startsWith('err ')) {
    line.className = 'text-red-400';
  } else if (lc.includes('warn') || lc.includes('⚠️')) {
    line.className = 'text-yellow-400';
  } else if (lc.includes('✅') || lc.includes('done') || lc.includes('success')) {
    line.className = 'text-emerald-400';
  }

  line.textContent = text;
  term.appendChild(line);
  term.scrollTop = term.scrollHeight;
}

function clearAimTerminal() {
  const term = document.getElementById('aim-terminal');
  if (term) term.innerHTML = '<span class="text-gray-600 italic">Terminal temizlendi.</span>';
}

function runSingleInstallStep(stepId, stepLabel) {
  if (_aimCurrentSSE) {
    _aimCurrentSSE.close();
    _aimCurrentSSE = null;
  }

  _aimStepStates[stepId] = 'running';
  if (_aimAnalysis) renderAimSteps(_aimAnalysis.installSteps);

  const statusEl = document.getElementById('aim-status-text');
  if (statusEl) statusEl.textContent = `⚙️ ${stepLabel} çalışıyor...`;

  document.getElementById('aim-stop-btn').classList.remove('hidden');
  document.getElementById('aim-run-all-btn').disabled = true;

  aimTerminalAppend(`\n▶ Adım: ${stepLabel}`, 'text-indigo-300 font-bold');

  const url = `${API_BASE}/api/repos/${encodeURIComponent(_aimCurrentRepo)}/install/stream?step=${encodeURIComponent(stepId)}`;
  const sse = new EventSource(url);
  _aimCurrentSSE = sse;

  sse.onmessage = (e) => { aimTerminalAppend(e.data); };

  sse.addEventListener('done', () => {
    _aimStepStates[stepId] = 'done';
    sse.close(); _aimCurrentSSE = null;
    if (_aimAnalysis) renderAimSteps(_aimAnalysis.installSteps);
    if (statusEl) statusEl.textContent = `✅ ${stepLabel} tamamlandı`;
    document.getElementById('aim-stop-btn').classList.add('hidden');
    document.getElementById('aim-run-all-btn').disabled = false;
    // Continue pipeline if running
    _continueAimPipeline();
  });

  sse.addEventListener('error', (e) => {
    _aimStepStates[stepId] = 'error';
    sse.close(); _aimCurrentSSE = null;
    if (_aimAnalysis) renderAimSteps(_aimAnalysis.installSteps);
    if (statusEl) statusEl.textContent = `❌ ${stepLabel} hata ile tamamlandı`;
    document.getElementById('aim-stop-btn').classList.add('hidden');
    document.getElementById('aim-run-all-btn').disabled = false;
    _aimRunningPipeline = false;
    _aimPipelineQueue = [];
  });

  sse.addEventListener('cancelled', () => {
    _aimStepStates[stepId] = 'idle';
    sse.close(); _aimCurrentSSE = null;
    if (_aimAnalysis) renderAimSteps(_aimAnalysis.installSteps);
    if (statusEl) statusEl.textContent = '⛔ Durduruldu';
    document.getElementById('aim-stop-btn').classList.add('hidden');
    document.getElementById('aim-run-all-btn').disabled = false;
    _aimRunningPipeline = false;
    _aimPipelineQueue = [];
  });

  sse.onerror = () => {
    if (sse.readyState === EventSource.CLOSED) return;
    _aimStepStates[stepId] = 'error';
    sse.close(); _aimCurrentSSE = null;
    aimTerminalAppend('❌ SSE bağlantısı koptu.', 'text-red-400');
    if (_aimAnalysis) renderAimSteps(_aimAnalysis.installSteps);
    document.getElementById('aim-stop-btn').classList.add('hidden');
    document.getElementById('aim-run-all-btn').disabled = false;
    _aimRunningPipeline = false;
  };
}

function _continueAimPipeline() {
  if (!_aimRunningPipeline || _aimPipelineQueue.length === 0) {
    _aimRunningPipeline = false;
    const statusEl = document.getElementById('aim-status-text');
    if (statusEl && _aimPipelineQueue.length === 0 && _aimRunningPipeline === false) {
      statusEl.textContent = '🎉 Tüm adımlar tamamlandı!';
    }
    return;
  }
  const next = _aimPipelineQueue.shift();
  runSingleInstallStep(next.id, next.label);
}

function runAllInstallSteps() {
  if (!_aimAnalysis || !_aimAnalysis.installSteps || _aimAnalysis.installSteps.length === 0) {
    showToast('Kurulum adımı bulunamadı.');
    return;
  }
  // Reset all states
  _aimStepStates = {};
  _aimPipelineQueue = [..._aimAnalysis.installSteps];
  _aimRunningPipeline = true;
  aimTerminalAppend('\n🚀 Pipeline başlatılıyor — tüm adımlar sırayla çalışacak', 'text-indigo-300 font-bold');
  _continueAimPipeline();
}

function stopCurrentInstall() {
  if (_aimCurrentSSE) {
    _aimCurrentSSE.close();
    _aimCurrentSSE = null;
  }
  _aimRunningPipeline = false;
  _aimPipelineQueue = [];
  aimTerminalAppend('⛔ Kullanıcı tarafından durduruldu.', 'text-red-400');
  document.getElementById('aim-stop-btn').classList.add('hidden');
  document.getElementById('aim-run-all-btn').disabled = false;
  const statusEl = document.getElementById('aim-status-text');
  if (statusEl) statusEl.textContent = '⛔ Durduruldu';
}

function closeAdvancedInstallModal() {
  stopCurrentInstall();
  const modal = document.getElementById('advanced-install-modal');
  if (modal) modal.classList.add('hidden');
}


function updateIdePath(pathId, newPath) {
  const target = activeIdePaths.find(p => p.id === pathId);
  if (target) {
    target.path = newPath;
    target.status = '✎ Özel Yol';
    saveIdePaths(activeIdePaths);
    showToast(`IDE dosya yolu güncellendi.`);
  }
}

function copyPathToClipboard(path) {
  navigator.clipboard.writeText(path);
  showToast(`Dosya yolu panoya kopyalandı:\n${path}`);
}

function renderIdePaths() {
  const container = document.getElementById('ide-paths-grid');
  if (!container) return;

  container.innerHTML = activeIdePaths.map(p => `
    <div class="glass-card space-y-3 flex flex-col justify-between">
      <div>
        <div class="flex items-center justify-between mb-2">
          <h3 class="font-semibold text-sm text-gray-100 flex items-center gap-2">
            <i data-lucide="folder-git-2" class="w-4 h-4 text-indigo-400"></i> ${p.name}
          </h3>
          <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">${p.status}</span>
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-mono text-gray-400">Hedef Konfigürasyon Dosya Adresi:</label>
          <input
            type="text"
            value="${p.path}"
            onchange="updateIdePath('${p.id}', this.value)"
            class="glass-input font-mono text-xs py-1.5"
            placeholder="C:\\Path\\To\\mcp_config.json"
          />
        </div>
      </div>

      <div class="pt-3 border-t border-gray-800/80 flex items-center gap-2">
        <button onclick="copyPathToClipboard('${p.path.replace(/\\/g, '\\\\')}')" class="flex-1 py-1.5 px-3 rounded-xl text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white flex items-center justify-center gap-1.5 cursor-pointer transition-all">
          <i data-lucide="copy" class="w-3.5 h-3.5"></i> Yolu Kopyala
        </button>
        <button onclick="openJsonModal()" class="py-1.5 px-3 rounded-xl text-xs font-semibold bg-cyan-600 hover:bg-cyan-500 text-white flex items-center justify-center gap-1.5 cursor-pointer">
          <i data-lucide="file-json" class="w-3.5 h-3.5"></i> JSON Al
        </button>
      </div>

      <details class="rounded-xl border border-red-500/30 bg-red-950/20 overflow-hidden">
        <summary class="px-3 py-2 text-[11px] font-bold text-red-400 cursor-pointer flex items-center gap-1.5 select-none">
          <i data-lucide="triangle-alert" class="w-3.5 h-3.5"></i> Danger Zone
        </summary>
        <div class="p-3 pt-1 space-y-2 border-t border-red-500/20">
          <p class="text-[10px] text-red-300/80 leading-relaxed">
            Bu, sadece config dosyasını değil <strong>tüm "${p.name}" klasörünü</strong> kalıcı olarak siler. Silmeden önce otomatik zip yedeği alınır. Onaylamak için aşağıya tam olarak <code class="font-mono">${p.name}</code> yazın.
          </p>
          <input
            id="danger-confirm-${p.id}"
            type="text"
            placeholder="${p.name}"
            class="glass-input text-xs py-1.5 border-red-500/30"
          />
          <button
            onclick='dangerDeleteIde("${p.id}", ${JSON.stringify(p.name)}, document.getElementById("danger-confirm-${p.id}").value)'
            class="w-full py-1.5 px-3 rounded-xl text-xs font-semibold bg-red-600 hover:bg-red-500 text-white flex items-center justify-center gap-1.5 cursor-pointer transition-all"
          >
            <i data-lucide="trash-2" class="w-3.5 h-3.5"></i> Klasörü Kalıcı Sil
          </button>
        </div>
      </details>
    </div>
  `).join('');

  if (window.lucide) lucide.createIcons();
}

function renderSkills() {
  const container = document.getElementById('skills-grid');
  const countBadge = document.getElementById('skill-count-badge');
  if (countBadge) countBadge.innerText = activeRecommendedRepos.length;
  if (!container) return;

  if (activeRecommendedRepos.length === 0) {
    container.innerHTML = '<p class="text-xs text-gray-400 col-span-full text-center py-8">Kaldırılmamış önerilen repo kalmadı. İstediğiniz zaman varsayılan repolara sıfırlayabilirsiniz.</p>';
    return;
  }

  container.innerHTML = activeRecommendedRepos.map(s => `
    <div class="glass-card space-y-3 flex flex-col justify-between relative group">
      <div>
        <div class="flex items-center justify-between mb-2">
          <h3 class="font-semibold text-sm text-emerald-300 flex items-center gap-2">
            <i data-lucide="sparkles" class="w-4 h-4"></i> ${s.name}
          </h3>
          <div class="flex items-center gap-1.5">
            <span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">${s.category}</span>
            <button onclick="removeRecommendedRepo('${s.id}')" title="Bu Önerilen Repoyu Kapat / Sil" class="p-1 text-gray-500 hover:text-red-400 hover:bg-red-500/10 rounded cursor-pointer transition-colors">
              <i data-lucide="x" class="w-3.5 h-3.5"></i>
            </button>
          </div>
        </div>
        <p class="text-xs text-gray-400 leading-relaxed mb-3">${s.desc}</p>
        <p class="text-[10px] font-mono text-gray-500 bg-gray-950/80 p-2 rounded border border-gray-800 break-all">
          ${s.repo || '— (yerel klasör, git repo değil)'}
        </p>
      </div>

      <div class="pt-3 border-t border-gray-800/80 flex items-center gap-2">
        ${s.repo ? `
          <button onclick="installSkillCommand('${s.repo}', '${s.extra || '--all'}', this)" class="flex-1 py-2 px-3 rounded-xl text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white flex items-center justify-center gap-2 cursor-pointer transition-all active:scale-95 shadow-lg">
            <i data-lucide="download" class="w-3.5 h-3.5"></i> Global Kur (npx)
          </button>
          <a href="https://github.com/${s.repo}" target="_blank" rel="noreferrer" class="p-2 rounded-xl text-xs bg-gray-800 hover:bg-gray-700 text-gray-300 flex items-center justify-center cursor-pointer">
            <i data-lucide="external-link" class="w-4 h-4"></i>
          </a>
        ` : `
          <span class="flex-1 text-center py-2 px-3 rounded-xl text-xs font-semibold bg-gray-800 text-gray-400">repo/${s.id} içinde yerel</span>
        `}
      </div>
    </div>
  `).join('');
  if (window.lucide) lucide.createIcons();
}

function updateKey(serverId, field, value) {
  const server = activeMcpServers.find(s => s.id === serverId);
  if (server) {
    if (server.env) server.env[field] = value;
    if (server.headers) server.headers[field] = value;
    saveMcpServers(activeMcpServers);
    showToast(`'${field}' parametresi kaydedildi.`);
  }
}

function removeMcp(serverId) {
  if (confirm(`'${serverId}' MCP sunucusunu silmek istediğinize emin misiniz?`)) {
    activeMcpServers = activeMcpServers.filter(s => s.id !== serverId);
    saveMcpServers(activeMcpServers);
    renderMcps();
    showToast(`'${serverId}' MCP sunucusu silindi.`);
  }
}

function openAddMcpModal() {
  const modal = document.getElementById('add-mcp-modal');
  if (modal) modal.classList.remove('hidden');
}

function closeAddMcpModal() {
  const modal = document.getElementById('add-mcp-modal');
  if (modal) modal.classList.add('hidden');
}

function saveNewMcp() {
  const id = document.getElementById('new-mcp-id').value.trim().toLowerCase();
  const name = document.getElementById('new-mcp-name').value.trim();
  const type = document.getElementById('new-mcp-type').value;
  const target = document.getElementById('new-mcp-target').value.trim();
  const desc = document.getElementById('new-mcp-desc').value.trim();
  const category = document.getElementById('new-mcp-category').value;

  if (!id || !name || !target) {
    alert('Lütfen ID, İsim ve URL/Komut alanlarını doldurun.');
    return;
  }

  if (activeMcpServers.some(s => s.id === id)) {
    alert('Bu ID zaten kullanımda!');
    return;
  }

  const newServer = {
    id: id,
    name: name,
    type: type,
    desc: desc || `${name} MCP sunucusu.`,
    category: category || 'Utility',
    badge: 'Özel',
    icon: 'server',
    iconColor: 'text-cyan-400',
    auth: false
  };

  if (type === 'http') {
    newServer.url = target;
  } else {
    const parts = target.split(' ');
    newServer.command = parts[0];
    newServer.args = parts.slice(1);
  }

  activeMcpServers.push(newServer);
  saveMcpServers(activeMcpServers);
  renderMcps();
  closeAddMcpModal();
  showToast(`Yeni MCP sunucusu eklendi: ${name}`);

  document.getElementById('new-mcp-id').value = '';
  document.getElementById('new-mcp-name').value = '';
  document.getElementById('new-mcp-target').value = '';
  document.getElementById('new-mcp-desc').value = '';
}

function buildConfigObject() {
  const config = { mcpServers: {} };
  activeMcpServers.forEach(s => {
    if (s.type === 'http') {
      config.mcpServers[s.id] = { type: 'http', url: s.url };
      if (s.headers && Object.values(s.headers).some(v => v)) {
        config.mcpServers[s.id].headers = s.headers;
      }
    } else {
      config.mcpServers[s.id] = { command: s.command, args: s.args || [] };
      if (s.cwd) {
        config.mcpServers[s.id].cwd = s.cwd;
      }
      if (s.env) {
        config.mcpServers[s.id].env = s.env;
      }
    }
  });
  return config;
}

function openJsonModal() {
  const output = document.getElementById('json-output');
  const modal = document.getElementById('json-modal');
  if (output) output.value = JSON.stringify(buildConfigObject(), null, 2);
  if (modal) modal.classList.remove('hidden');
}

function closeModal() {
  const modal = document.getElementById('json-modal');
  if (modal) modal.classList.add('hidden');
}

function applyRawJson() {
  const rawText = document.getElementById('json-output').value;
  try {
    const parsed = JSON.parse(rawText);
    if (!parsed.mcpServers) {
      alert("Hata: JSON içinde 'mcpServers' anahtarı bulunamadı!");
      return;
    }

    const updatedList = [];
    for (const [id, s] of Object.entries(parsed.mcpServers)) {
      const existing = activeMcpServers.find(item => item.id === id);
      const isHttp = s.type === 'http' || !!s.url;
      updatedList.push({
        id: id,
        name: existing ? existing.name : id.toUpperCase(),
        type: isHttp ? 'http' : 'stdio',
        url: s.url || '',
        command: s.command || '',
        args: s.args || [],
        env: s.env || (existing ? existing.env : undefined),
        headers: s.headers || (existing ? existing.headers : undefined),
        desc: existing ? existing.desc : `${id} MCP sunucusu.`,
        category: existing ? existing.category : 'DevOps',
        badge: existing ? existing.badge : 'Özel',
        icon: existing ? existing.icon : 'server',
        iconColor: existing ? existing.iconColor : 'text-cyan-400',
        auth: !!(s.env || s.headers || (existing && existing.auth))
      });
    }

    activeMcpServers = updatedList;
    saveMcpServers(activeMcpServers);
    renderMcps();
    closeModal();
    showToast('Ham JSON başarıyla uygulandı ve kaydedildi!');
  } catch (err) {
    alert('Geçersiz JSON formatı: ' + err.message);
  }
}

function copyToClipboard() {
  const output = document.getElementById('json-output');
  if (!output) return;
  navigator.clipboard.writeText(output.value);
  showToast('MCP JSON konfigürasyonu panoya kopyalandı!');
}

function downloadJsonFile() {
  const output = document.getElementById('json-output');
  if (!output) return;
  const blob = new Blob([output.value], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'mcp_config.json';
  a.click();
}

function resetToDefaultMcps() {
  if (confirm('Tüm MCP sunucularını silmek, IDE yollarını yeniden taramak ve skill kataloğunu varsayılana döndürmek istiyor musunuz?')) {
    activeMcpServers = [];
    activeRecommendedRepos = DEFAULT_RECOMMENDED_REPOS;
    saveMcpServers(activeMcpServers);
    saveRecommendedRepos(activeRecommendedRepos);
    renderMcps();
    renderSkills();
    autoDetectIdes(true);
    showToast('MCP sunucuları temizlendi, IDE yolları yeniden tarandı.');
  }
}

// Initial Render
// ---------- Otomatik Kurulum Modu (Auto Install Mode) ----------

async function loadAutoInstallConfig() {
  try {
    const res = await fetch(`${API_BASE}/api/auto-install/config`);
    if (!res.ok) return;
    const cfg = await res.json();
    applyAutoInstallConfigToUI(cfg);
  } catch (e) {
    // backend not reachable
  }
}

function applyAutoInstallConfigToUI(cfg) {
  const setCheck = (id, val) => { const el = document.getElementById(id); if (el) el.checked = !!val; };
  const setBadge = () => {
    const badge = document.getElementById('aim-status-badge');
    if (!badge) return;
    const enabled = document.getElementById('aim-enabled')?.checked;
    badge.textContent = enabled ? '● AKTİF' : '○ Pasif';
    badge.className = `text-[10px] font-bold px-2.5 py-1 rounded-full border ${
      enabled ? 'bg-purple-500/20 text-purple-300 border-purple-500/40' : 'bg-gray-700 text-gray-400 border-gray-600'
    }`;
    const sub = document.getElementById('aim-suboptions');
    if (sub) {
      sub.classList.toggle('opacity-50', !enabled);
      sub.classList.toggle('pointer-events-none', !enabled);
    }
  };

  setCheck('aim-enabled', cfg.enabled);
  setCheck('aim-trigger-clone', cfg.triggerOnClone);
  setCheck('aim-trigger-scan', cfg.triggerOnScan);
  setCheck('aim-trigger-start', cfg.triggerOnStart);
  setCheck('aim-only-required', cfg.onlyRequired);
  setCheck('aim-allow-shell', cfg.allowShellScripts);

  // Runtimes
  const runtimes = cfg.enabledRuntimes || [];
  document.querySelectorAll('.aim-runtime').forEach(el => {
    el.checked = runtimes.includes(el.value);
  });

  // Timeout
  const timeoutEl = document.getElementById('aim-timeout');
  const timeoutVal = document.getElementById('aim-timeout-val');
  if (timeoutEl && cfg.timeoutPerStep) {
    timeoutEl.value = cfg.timeoutPerStep;
    if (timeoutVal) timeoutVal.textContent = cfg.timeoutPerStep + ' sn';
  }

  // Error behavior
  const onError = cfg.onError || 'continue';
  document.querySelectorAll('input[name="aim-onerror"]').forEach(el => {
    el.checked = el.value === onError;
  });

  setBadge();
}

function aimOnChange() {
  const enabled = document.getElementById('aim-enabled')?.checked;
  const badge = document.getElementById('aim-status-badge');
  if (badge) {
    badge.textContent = enabled ? '● AKTİF' : '○ Pasif';
    badge.className = `text-[10px] font-bold px-2.5 py-1 rounded-full border ${
      enabled ? 'bg-purple-500/20 text-purple-300 border-purple-500/40' : 'bg-gray-700 text-gray-400 border-gray-600'
    }`;
  }
  const sub = document.getElementById('aim-suboptions');
  if (sub) {
    sub.classList.toggle('opacity-50', !enabled);
    sub.classList.toggle('pointer-events-none', !enabled);
  }
}

function readAutoInstallConfigFromUI() {
  const getCheck = (id) => !!document.getElementById(id)?.checked;
  const runtimes = [];
  document.querySelectorAll('.aim-runtime:checked').forEach(el => runtimes.push(el.value));
  const onError = document.querySelector('input[name="aim-onerror"]:checked')?.value || 'continue';
  const timeout = parseInt(document.getElementById('aim-timeout')?.value || '300', 10);
  return {
    enabled: getCheck('aim-enabled'),
    triggerOnClone: getCheck('aim-trigger-clone'),
    triggerOnScan: getCheck('aim-trigger-scan'),
    triggerOnStart: getCheck('aim-trigger-start'),
    onlyRequired: getCheck('aim-only-required'),
    allowShellScripts: getCheck('aim-allow-shell'),
    enabledRuntimes: runtimes,
    timeoutPerStep: timeout,
    onError: onError,
  };
}

async function saveAutoInstallConfig() {
  const cfg = readAutoInstallConfigFromUI();
  const fb = document.getElementById('aim-save-feedback');
  try {
    const res = await fetch(`${API_BASE}/api/auto-install/config`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(cfg),
    });
    const data = await res.json();
    if (data.ok) {
      if (fb) { fb.textContent = '✅ Kaydedildi'; setTimeout(() => { if (fb) fb.textContent = ''; }, 3000); }
      showToast('Otomatik Kurulum ayarları kaydedildi.');
    } else {
      if (fb) fb.textContent = '❌ Kayıt hatası';
    }
  } catch (e) {
    if (fb) fb.textContent = '❌ Backend bağlantısı yok';
  }
}

async function triggerAutoInstallScan(btnEl) {
  const origHtml = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Taranıyor...';
    if (window.lucide) lucide.createIcons();
  }
  const fb = document.getElementById('aim-save-feedback');
  try {
    const res = await fetch(`${API_BASE}/api/auto-install/scan`, { method: 'POST' });
    const data = await res.json();
    if (data.ok) {
      const count = (data.queued || []).length;
      const msg = `✅ ${count} repo kuyruğa alındı. Kurulumlar arka planda devam ediyor — Aktivite Günlüğü'nden takip edin.`;
      if (fb) { fb.textContent = msg; setTimeout(() => { if (fb) fb.textContent = ''; }, 6000); }
      showToast(msg);
    }
  } catch (e) {
    if (fb) fb.textContent = '❌ Backend bağlantısı yok';
    showToast('Backend bağlantısı yok.');
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = origHtml;
      if (window.lucide) lucide.createIcons();
    }
  }
}

// Auto-load on settings page
if (document.getElementById('aim-enabled')) {
  loadAutoInstallConfig();
}

// ---------- Ayarlar (settings.html): export/import + git repo tracking ----------

async function exportState() {
  try {
    const res = await fetch(`${API_BASE}/api/state`);
    const stateData = res.ok ? await res.json() : {};

    let loopConfig = {};
    try {
      const loopRes = await fetch(`${API_BASE}/api/loop/config`);
      if (loopRes.ok) loopConfig = await loopRes.json();
    } catch (e) {}

    const backupPackage = {
      version: 'wyvdev-full-backup-v1',
      exportedAt: new Date().toISOString(),
      mcpServers: activeMcpServers,
      recommendedRepos: activeRecommendedRepos,
      idePaths: activeIdePaths,
      trackedRepos: stateData.trackedRepos || [],
      deletedIdeIds: stateData.deletedIdeIds || [],
      loopConfig: loopConfig,
      uiPreferences: {
        repo_type_filter: SessionStore.get('repo_type_filter', 'all'),
        repo_sort_by: SessionStore.get('repo_sort_by', 'name'),
        library_view: localStorage.getItem('aitoolkit_library_view') || 'table',
        wyvdev_lang: localStorage.getItem('wyvdev_lang') || 'tr'
      }
    };

    const blob = new Blob([JSON.stringify(backupPackage, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `wyvdev-full-backup-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    showToast('📦 WyvDev tam yedek ve göç paketi (Full Backup) indirildi.');
  } catch (e) {
    showToast('⚠️ Dışa aktarma hatası: ' + e.message);
  }
}

function importState(file) {
  if (!file) return;
  const reader = new FileReader();
  reader.onload = async () => {
    try {
      const data = JSON.parse(reader.result);
      if (!confirm('İçe aktarılan WyvDev yedek paketi; mevcut tüm MCP, Skills, IDE yolları ve proje ayarlarının üzerine yazılacak. Eksik olan repolar otomatik klonlanacaktır.\n\nDevam edilsin mi?')) return;

      if (Array.isArray(data.mcpServers)) {
        activeMcpServers = data.mcpServers;
        localStorage.setItem('aitoolkit_mcp_servers', JSON.stringify(activeMcpServers));
      }
      if (Array.isArray(data.recommendedRepos)) {
        activeRecommendedRepos = data.recommendedRepos;
        localStorage.setItem('aitoolkit_recommended_repos', JSON.stringify(activeRecommendedRepos));
      }
      if (Array.isArray(data.idePaths)) {
        activeIdePaths = data.idePaths;
        localStorage.setItem('aitoolkit_ide_paths', JSON.stringify(activeIdePaths));
      }

      if (data.uiPreferences) {
        if (data.uiPreferences.repo_type_filter) SessionStore.set('repo_type_filter', data.uiPreferences.repo_type_filter);
        if (data.uiPreferences.repo_sort_by) SessionStore.set('repo_sort_by', data.uiPreferences.repo_sort_by);
        if (data.uiPreferences.library_view) localStorage.setItem('aitoolkit_library_view', data.uiPreferences.library_view);
        if (data.uiPreferences.wyvdev_lang) localStorage.setItem('wyvdev_lang', data.uiPreferences.wyvdev_lang);
      }

      const res = await fetch(`${API_BASE}/api/state/migrate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
      const result = await res.ok ? await res.json() : {};

      renderMcps();
      renderSkills();
      renderIdePaths();
      if (document.getElementById('library-scan-body')) loadLibraryScan();
      if (document.getElementById('tracked-repos-body')) renderTrackedRepos();

      const clonedCount = (result.clonedRepos || []).length;
      showToast(result.message || `🎉 WyvDev yedek paketi başarıyla yüklendi (${clonedCount} repo klonlanıyor).`);
    } catch (err) {
      showToast('⚠️ Geçersiz yedek JSON dosyası: ' + err.message);
    }
  };
  reader.readAsText(file);
}

function statusToBadge(status) {
  if (!status || status === 'upToDate') {
    return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">✓ Güncel</span>`;
  }
  if (status.startsWith('behind')) {
    return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">⏳ Geride (${status.split(':')[1]})</span>`;
  }
  if (status.startsWith('ahead')) {
    return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-cyan-500/15 text-cyan-400 border border-cyan-500/30">↑ İlerde (${status.split(':')[1]})</span>`;
  }
  return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-red-500/15 text-red-400 border border-red-500/30">⚠️ ${status}</span>`;
}

async function renderTrackedRepos() {
  const tbody = document.getElementById('tracked-repos-body');
  if (!tbody) return;
  try {
    const res = await fetch(`${API_BASE}/api/state`);
    if (!res.ok) throw new Error('backend yok');
    const data = await res.json();
    setBackendStatus(true);
    const repos = data.trackedRepos || [];
    if (!repos.length) {
      tbody.innerHTML = `<tr><td colspan="4" class="text-xs text-gray-400 text-center py-6">Henüz takip edilen repo yok. Skills Kataloğuna "repo" alanlı bir şey eklediğinizde (ör. GitHub Canlı Arama'dan) burada görünecek.</td></tr>`;
      return;
    }
    tbody.innerHTML = repos.map(r => `
      <tr class="border-t border-gray-800/60">
        <td class="py-2 px-3 text-xs font-mono text-gray-200">${r.repo}</td>
        <td class="py-2 px-3">${statusToBadge(r.status)}</td>
        <td class="py-2 px-3 text-[10px] text-gray-500">${r.lastChecked ? new Date(r.lastChecked).toLocaleString('tr-TR') : '—'}</td>
        <td class="py-2 px-3 text-right space-x-2">
          <button onclick="checkRepo('${r.name}')" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-indigo-600 hover:bg-indigo-500 text-white cursor-pointer">Şimdi Kontrol Et</button>
          <button onclick="pullRepo('${r.name}')" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-emerald-600 hover:bg-emerald-500 text-white cursor-pointer">Güncelle (pull)</button>
          <button onclick="deleteRepoFolder('${r.name}')" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-red-600 hover:bg-red-500 text-white cursor-pointer">Klasörü Sil</button>
        </td>
      </tr>
    `).join('');
  } catch (e) {
    setBackendStatus(false);
    tbody.innerHTML = `<tr><td colspan="4" class="text-xs text-red-400 text-center py-6">Backend'e ulaşılamıyor. ai-toolkit.exe çalışıyor mu?</td></tr>`;
  }
}

async function checkRepo(name) {
  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/check`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'kontrol hatası');
    showToast(`🔎 '${name}' durumu: ${data.status}`);
    renderTrackedRepos();
  } catch (e) {
    showToast('⚠️ ' + e.message);
  }
}

async function pullRepo(name) {
  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/pull`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'pull hatası');
    showToast(`⬇️ '${name}' güncellendi.`);
    renderTrackedRepos();
  } catch (e) {
    showToast('⚠️ ' + e.message);
  }
}

// ---------- Yerel Kütüphane: repo/ klasörünü tara, tip tespit et, tek tık ekle ----------

let lastScanResults = [];

// ---------- Aktivite Günlüğü: kalıcı işlem geçmişi (settings.html) ----------

async function loadActivityLog() {
  const container = document.getElementById('activity-log-output');
  if (!container) return;
  try {
    const res = await fetch(`${API_BASE}/api/activity?limit=50`);
    if (!res.ok) throw new Error('backend yok');
    const lines = await res.json();
    setBackendStatus(true);
    container.textContent = lines.length ? lines.join('\n') : 'Henüz kayıtlı işlem yok.';
  } catch (e) {
    setBackendStatus(false);
    container.textContent = "Backend'e ulaşılamıyor. ai-toolkit.exe çalışıyor mu?";
  }
}

async function loadLibraryScan() {
  const tbody = document.getElementById('library-scan-body');
  if (!tbody) return;
  try {
    const res = await fetch(`${API_BASE}/api/repos/scan`);
    if (!res.ok) throw new Error('backend yok');
    lastScanResults = await res.json();
    setBackendStatus(true);
    renderLibraryScan();
    updateRepoCountBadge(lastScanResults.length);
  } catch (e) {
    setBackendStatus(false);
    tbody.innerHTML = `<tr><td colspan="3" class="text-xs text-red-400 text-center py-6">Backend'e ulaşılamıyor. ai-toolkit.exe çalışıyor mu?</td></tr>`;
  }
}

// Sidebar badge always reflects what's actually cloned into repo/ — not the
// (unrelated) MCP catalog or skill catalog counts that used to sit here.
async function updateRepoCountBadge(knownCount) {
  const badge = document.getElementById('repo-count-badge');
  if (!badge) return;
  if (typeof knownCount === 'number') {
    badge.innerText = knownCount;
    return;
  }
  try {
    const res = await fetch(`${API_BASE}/api/repos/scan`);
    if (!res.ok) return;
    const repos = await res.json() || [];
    badge.innerText = repos.length;
  } catch (e) {}
}

async function startRepoProject(name, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Başlatılıyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/start`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Başlatılamadı');
    showToast(data.message || `🚀 '${name}' başlatıldı.`);
  } catch (e) {
    showToast('⚠️ Başlatma hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function repairRepoProject(name, runtimes, btnEl) {
  let action = 'npm-repair';
  if (runtimes && runtimes.includes('python')) action = 'pip-repair';
  else if (runtimes && runtimes.includes('rust')) action = 'cargo-repair';
  else if (runtimes && runtimes.includes('docker')) action = 'docker-repair';

  await runRepoAction(name, action, btnEl);
  setTimeout(() => {
    loadLibraryScan();
  }, 2500);
}

function goToSystemDiagnostics(tool) {
  showToast(`⚠️ '${tool}' kurulu değil — Sistem & Teşhis sayfasına yönlendiriliyorsunuz.`);
  setTimeout(() => { window.location.href = 'settings.html'; }, 900);
}

// Single-click install+start: chains the right install action for this repo's
// runtime, then starts it — replaces the old "Kur (npm)" + separate "Başlat"
// two-step flow.
async function installAndStart(name, runtimes, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  const setLabel = (text) => {
    if (!btnEl) return;
    btnEl.innerHTML = `<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> ${text}`;
    if (window.lucide) lucide.createIcons();
  };
  if (btnEl) btnEl.disabled = true;

  const action = runtimes.includes('node') ? 'npm-install'
    : runtimes.includes('python') ? 'pip-install'
    : runtimes.includes('rust') ? 'cargo-build'
    : null;

  try {
    if (action) {
      setLabel('Kuruluyor...');
      const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/run`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action })
      });
      const data = await res.json();
      if (!res.ok || data.ok === false) throw new Error(data.error || 'Kurulum başarısız');
    }
    setLabel('Başlatılıyor...');
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/start`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Başlatılamadı');
    showToast(data.message || `🚀 '${name}' başlatıldı.`);
  } catch (e) {
    showToast('⚠️ ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
    loadLibraryScan();
  }
}

// Docker-mode equivalent of installAndStart: build the image, then run it.
async function dockerInstallAndRun(name, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  const setLabel = (text) => {
    if (!btnEl) return;
    btnEl.innerHTML = `<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> ${text}`;
    if (window.lucide) lucide.createIcons();
  };
  if (btnEl) btnEl.disabled = true;

  try {
    setLabel('Docker Build...');
    let res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/run`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ action: 'docker-build' })
    });
    let data = await res.json();
    if (!res.ok || data.ok === false) throw new Error(data.error || 'Docker build başarısız');

    setLabel('Docker Başlatılıyor...');
    res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/run`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ action: 'docker-run' })
    });
    data = await res.json();
    if (!res.ok || data.ok === false) throw new Error(data.error || 'Docker run başarısız');
    showToast(`🐳 '${name}' docker konteynerinde başlatıldı.`);
  } catch (e) {
    showToast('⚠️ ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
    loadLibraryScan();
  }
}

// Copies the skill straight into every IDE that has a skills folder (today:
// Claude Code CLI) — the real "enable" action, not just adding a catalog entry.
async function enableSkillToIdes(name, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Etkinleştiriliyor...';
    if (window.lucide) lucide.createIcons();
  }
  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/enable-skill`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Etkinleştirilemedi');
    showToast(data.message || `✅ '${name}' skill'i etkinleştirildi.`);
  } catch (e) {
    showToast('⚠️ ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

let libraryViewMode = localStorage.getItem('aitoolkit_library_view') || 'table';

function setLibraryViewMode(mode) {
  libraryViewMode = mode;
  localStorage.setItem('aitoolkit_library_view', mode);

  const tableWrapper = document.getElementById('library-view-table-wrapper');
  const gridWrapper = document.getElementById('library-view-grid-wrapper');
  const tableBtn = document.getElementById('view-mode-table-btn');
  const gridBtn = document.getElementById('view-mode-grid-btn');

  if (tableWrapper && gridWrapper) {
    if (mode === 'grid') {
      tableWrapper.classList.add('hidden');
      gridWrapper.classList.remove('hidden');
      if (tableBtn) tableBtn.className = 'px-2.5 py-1 rounded-lg text-xs font-semibold text-gray-400 hover:text-white flex items-center gap-1 cursor-pointer';
      if (gridBtn) gridBtn.className = 'px-2.5 py-1 rounded-lg text-xs font-semibold bg-gray-800 text-cyan-300 flex items-center gap-1 cursor-pointer';
    } else {
      tableWrapper.classList.remove('hidden');
      gridWrapper.classList.add('hidden');
      if (tableBtn) tableBtn.className = 'px-2.5 py-1 rounded-lg text-xs font-semibold bg-gray-800 text-cyan-300 flex items-center gap-1 cursor-pointer';
      if (gridBtn) gridBtn.className = 'px-2.5 py-1 rounded-lg text-xs font-semibold text-gray-400 hover:text-white flex items-center gap-1 cursor-pointer';
    }
  }
  renderLibraryScan();
}

function renderLibraryScan() {
  const tbody = document.getElementById('library-scan-body');
  const gridContainer = document.getElementById('library-view-grid-wrapper');
  if (!tbody && !gridContainer) return;

  const typeFilter = document.getElementById('repo-type-filter')?.value || 'all';
  const sortBy = document.getElementById('repo-sort-by')?.value || 'grouped';

  SessionStore.set('repo_type_filter', typeFilter);
  SessionStore.set('repo_sort_by', sortBy);

  // 1. Filter
  let filtered = [...lastScanResults];
  if (typeFilter !== 'all') {
    filtered = filtered.filter(e => entryCapabilities(e).includes(typeFilter));
  }

  if (!filtered.length) {
    const emptyHtml = `<div class="text-xs text-gray-400 text-center py-8 col-span-full">Filtreye uygun proje veya repo bulunamadı.</div>`;
    if (tbody) tbody.innerHTML = `<tr><td colspan="4" class="text-xs text-gray-400 text-center py-8">Filtreye uygun proje veya repo bulunamadı.</td></tr>`;
    if (gridContainer) gridContainer.innerHTML = emptyHtml;
    return;
  }

  const RUNTIME_BADGE = {
    node: '🟢 Node.js',
    python: '🐍 Python',
    rust: '🦀 Rust',
    go: '🐹 Go',
    ruby: '💎 Ruby',
    php: '🐘 PHP',
    java: '☕ Java',
    dotnet: '🔷 .NET',
    elixir: '💧 Elixir',
    deno: '🦕 Deno',
    docker: '🐳 Docker'
  };

  const TYPE_BADGE_STYLE = {
    service: 'bg-amber-500/15 text-amber-300 border-amber-500/30',
    mcp: 'bg-cyan-500/15 text-cyan-300 border-cyan-500/30',
    skill: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
    library: 'bg-purple-500/15 text-purple-300 border-purple-500/30',
    other: 'bg-gray-700/40 text-gray-400 border-gray-600/40'
  };

  const GROUPS = [
    { key: 'mcp', title: '🔌 MCP Sunucuları (MCP Servers)', color: 'text-cyan-400', icon: 'server' },
    { key: 'skill', title: '📄 AI Skills', color: 'text-emerald-400', icon: 'sparkles' },
    { key: 'service', title: '⚡ Servis ve Uygulamalar', color: 'text-amber-400', icon: 'zap' },
    { key: 'library', title: '📦 Kütüphaneler', color: 'text-purple-400', icon: 'folder-git-2' },
    { key: 'other', title: '📁 Diğer Projeler', color: 'text-gray-400', icon: 'folder' }
  ];

  // A repo can genuinely have more than one capability (e.g. it ships both
  // an MCP server and a skill bundle) — return every one it has instead of
  // the single best-guess RepoType, so it can show up under each relevant
  // group with only that group's own actions instead of everything blended
  // into one row.
  function entryCapabilities(entry) {
    const caps = [];
    if (entry.looksLikeMcp) caps.push('mcp');
    if (entry.hasSkill) caps.push('skill');
    if (entry.runMode) caps.push('service');
    if (!caps.length) caps.push(entry.hasPackageJson || (entry.runtimes || []).length ? 'library' : 'other');
    return caps;
  }

  function primaryRuntimeBadge(entry) {
    const rt = entry.runMode === 'docker' ? 'docker'
      : (entry.runtimes || []).find(r => r !== 'docker') || (entry.runtimes || [])[0];
    if (!rt) return '';
    return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-gray-700/40 text-gray-300 border border-gray-600/40">${RUNTIME_BADGE[rt] || rt}</span>`;
  }

  function statusBadge(entry) {
    if (entry.isRunning) {
      const portTxt = entry.runningPort ? `Port: ${entry.runningPort}` : 'Aktif';
      return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 animate-pulse">🟢 Çalışıyor (${portTxt})</span>`;
    }
    if (entry.hasStartError) {
      return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-red-500/15 text-red-400 border border-red-500/30" title="${escapeHtml(entry.startErrorMsg || '')}">⚠️ Çalıştırma Hatası</span>`;
    }
    if (entry.missingTool) {
      return `<button onclick="goToSystemDiagnostics('${entry.missingTool}')" title="Sistem & Teşhis sayfasına git" class="text-[10px] font-semibold px-2 py-0.5 rounded bg-amber-500/15 text-amber-300 border border-amber-500/30 cursor-pointer hover:bg-amber-500/25">⚠️ ${escapeHtml(entry.missingTool)} kurulu değil</button>`;
    }
    if (entry.isInstalled) {
      return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">✓ Hazır</span>`;
    }
    if (entry.runMode) {
      return `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">⚠️ Kurulum Gerekli</span>`;
    }
    return '';
  }

  function primaryActionButton(entry, extraClass) {
    const runtimes = entry.runtimes || [];
    const cls = `rounded-xl text-[11px] font-bold shadow-lg cursor-pointer transition-all active:scale-95 flex items-center justify-center gap-1.5 ${extraClass || 'px-3 py-1'}`;
    if (entry.isRunning) {
      return `<button onclick="killRunningApp('${entry.name}', this)" class="${cls} bg-red-600 hover:bg-red-500 text-white"><i data-lucide="square" class="w-3.5 h-3.5 fill-current"></i> Durdur</button>`;
    }
    if (entry.hasStartError) {
      return `<button onclick="repairRepoProject('${entry.name}', ${JSON.stringify(runtimes).replace(/"/g, '&quot;')}, this)" title="Bağımlılıkları zorla yenile" class="${cls} bg-gradient-to-r from-amber-600 to-red-600 hover:from-amber-500 hover:to-red-500 text-white"><i data-lucide="wrench" class="w-3.5 h-3.5 fill-current"></i> Onar</button>`;
    }
    if (entry.missingTool) {
      return `<button onclick="goToSystemDiagnostics('${entry.missingTool}')" title="Sistem & Teşhis sayfasından kurun" class="${cls} bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-500 hover:to-orange-500 text-white"><i data-lucide="download" class="w-3.5 h-3.5"></i> ${escapeHtml(entry.missingTool)} Kurulu Değil</button>`;
    }
    if (entry.isInstalled) {
      return `<button onclick="startRepoProject('${entry.name}', this)" title="Çalıştır: ${escapeHtml(entry.startCommand || '')}" class="${cls} bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white"><i data-lucide="play" class="w-3.5 h-3.5 fill-current"></i> Başlat</button>`;
    }
    if (entry.runMode === 'docker') {
      return `<button onclick="showAdvancedInstallModal('${entry.name}')" class="${cls} bg-gradient-to-r from-blue-700 to-cyan-700 hover:from-blue-600 hover:to-cyan-600 text-white"><i data-lucide="package-2" class="w-3.5 h-3.5"></i> Kurulum Merkezi</button>`;
    }
    if (entry.runMode === 'local') {
      return `<button onclick="showAdvancedInstallModal('${entry.name}')" class="${cls} bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white"><i data-lucide="package-2" class="w-3.5 h-3.5"></i> Kurulum Merkezi</button>`;
    }
    return '';
  }

  // groupKey scopes which actions render: 'mcp' → config/test only (no
  // Skill/Başlat), 'skill' → enable only, 'service' → run/install only.
  // null (flat sort modes, where there's no group boundary) keeps the old
  // blended behavior — every applicable action on one row.
  function renderTableRow(entry, i, groupKey) {
    const typeStyle = TYPE_BADGE_STYLE[entry.repoType] || TYPE_BADGE_STYLE.other;
    const typeBadge = `<span class="text-[10px] font-bold px-2 py-0.5 rounded border ${typeStyle}">${escapeHtml(entry.repoTypeLabel || '📁 Diğer')}</span>`;

    const badges = [typeBadge, entry.gitStatus ? statusToBadge(entry.gitStatus) : '', primaryRuntimeBadge(entry)].filter(Boolean).join(' ');

    const alreadySkill = activeRecommendedRepos.some(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));
    const matchedMcp = activeMcpServers.find(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));

    // Her grup yalnızca kendi aksiyonunu gösterir
    const isMcpGroup   = groupKey === 'mcp';
    const isSkillGroup = groupKey === 'skill';
    const isServiceGroup = groupKey === 'service' || groupKey === 'other' || !groupKey;

    const secondary = [];

    // 🔌 MCP grubu: sadece MCP yapılandır / düzenle / test
    if (isMcpGroup) {
      secondary.push(matchedMcp
        ? `<button onclick="goToMcpConfig('${matchedMcp.id}')" title="Değişkenlerini düzenle" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-gray-800 hover:bg-cyan-900/60 text-cyan-300 border border-cyan-500/20 cursor-pointer flex items-center gap-1"><i data-lucide="sliders-horizontal" class="w-3 h-3"></i> MCP ekli — Düzenle</button>`
        : `<button onclick="addMcpFromScan(${i})" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-cyan-600 hover:bg-cyan-500 text-white cursor-pointer">MCP Yapılandır</button>`);
      if (matchedMcp) secondary.push(`<button onclick="testMcpConnection('${matchedMcp.id}', this)" title="Bağlantıyı test et" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-gray-800 hover:bg-indigo-900/60 text-indigo-300 border border-indigo-500/20 cursor-pointer flex items-center gap-1"><i data-lucide="plug-zap" class="w-3 h-3"></i> Test</button>`);
    }

    // 📄 Skill grubu: sadece Skill Etkinleştir
    if (isSkillGroup) {
      secondary.push(alreadySkill
        ? `<span class="text-[10px] text-emerald-500 flex items-center gap-1"><i data-lucide="check-circle" class="w-3 h-3"></i> Skill ekli</span>`
        : `<button onclick="enableSkillToIdes('${entry.name}', this)" title="Bu skill'i tespit edilen IDE'lere kopyalar" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-emerald-600 hover:bg-emerald-500 text-white cursor-pointer">Skill'i Etkinleştir</button>`);
    }

    // ⚡ Servis / Uygulama grubu veya flat görünüm: Başlat / Kur
    const primaryBtn = isServiceGroup ? primaryActionButton(entry) : '';

    // 🔧 Her grupta ortak: git pull + sil
    if (entry.repo) secondary.push(`<button onclick="pullRepo('${entry.name}')" title="Güncelle (git pull)" class="p-1.5 rounded-lg text-gray-400 hover:text-cyan-300 hover:bg-cyan-500/10 cursor-pointer"><i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i></button>`);
    secondary.push(`<button onclick="deleteRepoFolder('${entry.name}')" title="Klasörü sil" class="p-1.5 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-500/10 cursor-pointer"><i data-lucide="trash-2" class="w-3.5 h-3.5"></i></button>`);

    return `
      <tr class="border-t border-gray-800/60 hover:bg-gray-900/40">
        <td class="py-2.5 px-3 text-xs font-mono text-gray-200 align-middle">
          <div class="font-bold text-gray-100">${escapeHtml(entry.name)}</div>
          <div class="text-[10px] text-gray-500 truncate">${escapeHtml(entry.packageName || entry.path)}</div>
        </td>
        <td class="py-2.5 px-3 space-x-1 space-y-1 align-middle">${badges}</td>
        <td class="py-2.5 px-3 align-middle">${statusBadge(entry)}</td>
        <td class="py-2.5 px-3 text-right align-middle">
          <div class="flex items-center justify-end gap-1.5 flex-wrap">
            ${primaryBtn}
            ${secondary.join('')}
          </div>
        </td>
      </tr>
    `;
  }

  function renderGridCard(entry, i, groupKey) {
    const typeStyle = TYPE_BADGE_STYLE[entry.repoType] || TYPE_BADGE_STYLE.other;
    const alreadySkill = activeRecommendedRepos.some(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));
    const matchedMcp = activeMcpServers.find(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));

    const isMcpGroup   = groupKey === 'mcp';
    const isSkillGroup = groupKey === 'skill';
    const isServiceGroup = groupKey === 'service' || groupKey === 'other' || !groupKey;

    return `
      <div class="glass-card flex flex-col justify-between space-y-3.5 relative overflow-hidden ${entry.isRunning ? 'border-emerald-500/50 bg-emerald-950/10' : ''}">
        <div class="space-y-2">
          <div class="flex items-start justify-between gap-2">
            <div>
              <h3 class="font-bold text-sm text-gray-100 truncate" title="${escapeHtml(entry.name)}">${escapeHtml(entry.name)}</h3>
              <p class="text-[10px] text-gray-400 font-mono truncate">${escapeHtml(entry.packageName || entry.path)}</p>
            </div>
            <span class="text-[10px] font-bold px-2 py-0.5 rounded border shrink-0 ${typeStyle}">${escapeHtml(entry.repoTypeLabel || '📁 Diğer')}</span>
          </div>
          <div class="flex items-center gap-1.5 flex-wrap">
            ${statusBadge(entry)}
            ${primaryRuntimeBadge(entry)}
          </div>
        </div>

        <div class="pt-2 border-t border-gray-800/80 space-y-2">
          ${isServiceGroup ? primaryActionButton(entry, 'w-full py-1.5 px-3 text-xs') : ''}

          <div class="flex items-center justify-between text-[11px] pt-1 flex-wrap gap-1.5">
            ${isSkillGroup
              ? (alreadySkill
                  ? `<span class="text-[10px] text-emerald-500 flex items-center gap-1"><i data-lucide="check-circle" class="w-3 h-3"></i> Skill ekli</span>`
                  : `<button onclick="enableSkillToIdes('${entry.name}', this)" class="text-emerald-400 hover:underline cursor-pointer font-semibold">+ Skill'i Etkinleştir</button>`)
              : ''}
            ${isMcpGroup
              ? (matchedMcp
                  ? `<button onclick="goToMcpConfig('${matchedMcp.id}')" class="text-cyan-400 hover:underline cursor-pointer font-semibold flex items-center gap-1"><i data-lucide="sliders-horizontal" class="w-3 h-3"></i> MCP ekli — Düzenle</button>`
                  : `<button onclick="addMcpFromScan(${i})" class="text-cyan-400 hover:underline cursor-pointer font-semibold">+ MCP Yapılandır</button>`)
              : ''}
            ${isMcpGroup && matchedMcp ? `<button onclick="testMcpConnection('${matchedMcp.id}', this)" class="text-indigo-400 hover:underline cursor-pointer font-semibold">Test</button>` : ''}
            ${entry.repo ? `<button onclick="pullRepo('${entry.name}')" title="git pull" class="text-gray-400 hover:text-cyan-300 cursor-pointer"><i data-lucide="refresh-cw" class="w-3 h-3"></i></button>` : ''}
            <button onclick="deleteRepoFolder('${entry.name}')" class="text-red-400 hover:underline cursor-pointer">Sil</button>
          </div>
        </div>
      </div>
    `;
  }

  // Render Logic: Grouped or Flat
  if (sortBy === 'grouped' || sortBy === 'type') {
    let tableHtml = '';
    let gridHtml = '';

    GROUPS.forEach(grp => {
      const grpItems = filtered.filter(e => entryCapabilities(e).includes(grp.key));

      if (!grpItems.length) return;

      // Sort items within group by name
      grpItems.sort((a, b) => a.name.localeCompare(b.name));

      tableHtml += `
        <tr class="bg-gray-950/90 border-y border-gray-800">
          <td colspan="4" class="py-2.5 px-3">
            <div class="flex items-center gap-2 font-bold text-xs ${grp.color}">
              <i data-lucide="${grp.icon}" class="w-4 h-4"></i> ${grp.title}
              <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-gray-800 text-gray-300 border border-gray-700">${grpItems.length}</span>
            </div>
          </td>
        </tr>
      ` + grpItems.map(entry => renderTableRow(entry, lastScanResults.indexOf(entry), grp.key)).join('');

      gridHtml += `
        <div class="col-span-full pt-4 pb-2 border-b border-gray-800/80 flex items-center gap-2 font-bold text-xs ${grp.color}">
          <i data-lucide="${grp.icon}" class="w-4 h-4"></i> ${grp.title}
          <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-gray-800 text-gray-300 border border-gray-700">${grpItems.length}</span>
        </div>
      ` + grpItems.map(entry => renderGridCard(entry, lastScanResults.indexOf(entry), grp.key)).join('');
    });

    if (tbody) tbody.innerHTML = tableHtml;
    if (gridContainer) gridContainer.innerHTML = gridHtml;
  } else {
    filtered.sort((a, b) => {
      if (sortBy === 'name') return a.name.localeCompare(b.name);
      if (sortBy === 'running') return (b.isRunning ? 1 : 0) - (a.isRunning ? 1 : 0);
      if (sortBy === 'installed') return (b.isInstalled ? 1 : 0) - (a.isInstalled ? 1 : 0);
      return 0;
    });

    if (tbody) tbody.innerHTML = filtered.map(entry => renderTableRow(entry, lastScanResults.indexOf(entry))).join('');
    if (gridContainer) gridContainer.innerHTML = filtered.map(entry => renderGridCard(entry, lastScanResults.indexOf(entry))).join('');
  }

  if (window.lucide) lucide.createIcons();
}

// Generic runner for repo/<name> — npm install, pip install, cargo build, docker build/run.
// Same transparency pattern as installSkillCommand: real command + real output in a modal.
async function runRepoAction(name, action, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Çalışıyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/repos/${encodeURIComponent(name)}/run`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action })
    });
    const data = await res.json();
    showInstallResultModal(name, data.command, data.output, !!data.ok, data.error);
  } catch (e) {
    showInstallResultModal(name, action, '', false, "Backend sunucusuna ulaşılamadı. ai-toolkit.exe çalışıyor mu?");
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

function addSkillFromScan(index) {
  const entry = lastScanResults[index];
  if (!entry) return;
  const id = entry.repo ? entry.repo.toLowerCase().replace(/[^a-z0-9]+/g, '-') : entry.name;
  if (activeRecommendedRepos.some(s => s.id === id)) {
    showToast(`'${entry.name}' zaten Skills Kataloğunda ekli!`);
    return;
  }
  activeRecommendedRepos.push({
    id, name: entry.name, repo: entry.repo || undefined,
    category: 'Yerel Kütüphane', desc: entry.skillDesc || `${entry.name} skill.`, extra: '--all'
  });
  saveRecommendedRepos(activeRecommendedRepos);
  renderSkills();
  renderLibraryScan();
  showToast(`🚀 '${entry.name}' Skills Kataloğuna eklendi.`);
}

function addMcpFromScan(index) {
  const entry = lastScanResults[index];
  if (!entry) return;
  const id = entry.repo ? entry.repo.toLowerCase().replace(/[^a-z0-9]+/g, '-') : entry.name;
  if (activeMcpServers.some(s => s.id === id)) {
    showToast(`'${entry.name}' zaten MCP panelinde ekli!`);
    return;
  }
  // Prefer a real local invocation (node/python3 <abs path>, cwd=repo folder)
  // derived by the backend scan over guessing an npm-registry package name —
  // repo/<name> is a local clone, usually unpublished, so "npx -y <name>"
  // mostly just 404s against the registry.
  const hasLocalCmd = entry.runMode === 'local' && entry.localCommand && entry.localArgs && entry.localArgs.length;
  const command = hasLocalCmd ? entry.localCommand : 'npx';
  const args = hasLocalCmd ? entry.localArgs : ['-y', entry.packageName || entry.repo || entry.name];
  let desc;
  if (!hasLocalCmd) {
    desc = `${entry.name} MCP sunucusu (repo/${entry.name} içinden tespit edildi — komutu kontrol edin).`;
  } else if (entry.startCommandGuessed) {
    desc = `${entry.name} MCP sunucusu — repo/${entry.name} için bilinen bir başlatma script'i (dev/start/serve) bulunamadı, "${command} ${args.join(' ')}" rastgele bir script'ten tahmin edildi. Gerçekten bir MCP sunucusu başlatmayabilir — kontrol edin.`;
  } else {
    desc = `${entry.name} MCP sunucusu — repo/${entry.name} içinden yerel olarak çalıştırılır.`;
  }
  activeMcpServers.push({
    id, name: entry.name, type: 'stdio', command, args, cwd: entry.path || undefined, repo: entry.repo || undefined,
    desc, category: 'Yerel Kütüphane', badge: 'Kütüphane', icon: 'folder-git-2', iconColor: 'text-indigo-400', auth: false
  });
  saveMcpServers(activeMcpServers);
  renderMcps();
  renderLibraryScan();
  showToast(`🛒 '${entry.name}' MCP panosuna eklendi.`);
}

// ---------- Canlı İşlem & İndirme Durumu Takibi (Real-time Live Tasks) ----------

let activeTasksList = [];
let activeRunningApps = [];

async function pollLiveTasks() {
  try {
    const [tasksRes, appsRes] = await Promise.all([
      fetch(`${API_BASE}/api/tasks`),
      fetch(`${API_BASE}/api/apps/running`)
    ]);
    if (tasksRes.ok) activeTasksList = await tasksRes.json() || [];
    if (appsRes.ok) activeRunningApps = await appsRes.json() || [];
    renderLiveTasksWidget();
  } catch (e) {}
}

async function stopLiveTask(id, btnEl) {
  if (btnEl) btnEl.disabled = true;
  try {
    const res = await fetch(`${API_BASE}/api/tasks/${encodeURIComponent(id)}/stop`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Durdurulamadı');
    showToast(data.message || '🛑 Görev durduruldu.');
    pollLiveTasks();
  } catch (e) {
    showToast('⚠️ Görev durdurulamadı: ' + e.message);
    if (btnEl) btnEl.disabled = false;
  }
}

async function deleteLiveTask(id, btnEl) {
  if (btnEl) btnEl.disabled = true;
  activeTasksList = activeTasksList.filter(t => t.id !== id);
  renderLiveTasksWidget();
  try {
    const res = await fetch(`${API_BASE}/api/tasks/${encodeURIComponent(id)}`, { method: 'DELETE' });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || 'Silinemedi');
    }
  } catch (e) {
    showToast('⚠️ Görev silinemedi: ' + e.message);
    pollLiveTasks();
  }
}

async function killRunningApp(name, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3 h-3 animate-spin"></i> Durduruluyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/apps/${encodeURIComponent(name)}/kill`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Durdurulamadı');
    showToast(data.message || `🛑 '${name}' durduruldu.`);
    pollLiveTasks();
    if (document.getElementById('library-scan-body')) loadLibraryScan();
  } catch (e) {
    showToast('⚠️ Durdurma hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

function renderLiveTasksWidget() {
  let widget = document.getElementById('live-tasks-widget');
  if (!widget) {
    widget = document.createElement('div');
    widget.id = 'live-tasks-widget';
    widget.className = 'fixed bottom-5 left-5 z-40 max-w-sm space-y-2 pointer-events-none';
    document.body.appendChild(widget);
  }

  const now = Date.now();
  const runningTasks = activeTasksList.filter(t => t.status === 'running');
  const recentCompleted = activeTasksList.filter(t => {
    if (t.status === 'running') return false;
    if (!t.endedAt) return false;
    const endedTime = new Date(t.endedAt).getTime();
    return (now - endedTime) < 4000;
  }).slice(0, 2);

  if (!runningTasks.length && !recentCompleted.length && !activeRunningApps.length) {
    widget.innerHTML = '';
    return;
  }

  widget.innerHTML = `
    <div class="pointer-events-auto bg-gray-900/95 border border-purple-500/40 backdrop-blur-md p-3.5 rounded-2xl shadow-2xl space-y-2.5 text-xs animate-fade-in">
      <div class="flex items-center justify-between border-b border-gray-800 pb-2">
        <span class="font-bold text-gray-200 flex items-center gap-1.5 text-[11px]">
          ${activeRunningApps.length ? `<span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-ping"></span> Çalışan Uygulamalar (${activeRunningApps.length})` : runningTasks.length ? `<i data-lucide="loader-2" class="w-4 h-4 text-purple-400 animate-spin"></i> Canlı İşlem (${runningTasks.length})` : `<i data-lucide="check-circle-2" class="w-4 h-4 text-emerald-400"></i> Son İşlemler`}
        </span>
        <span class="text-[9px] font-mono text-purple-300 bg-purple-500/10 px-2 py-0.5 rounded-full border border-purple-500/20">Canlı Takip</span>
      </div>

      <div class="space-y-1.5 max-h-56 overflow-y-auto pr-1">
        ${activeRunningApps.map(app => `
          <div class="p-2.5 rounded-xl bg-emerald-950/40 border border-emerald-500/40 text-[11px] space-y-1.5">
            <div class="flex items-center justify-between font-bold text-emerald-200">
              <span class="truncate flex items-center gap-1.5" title="${escapeHtml(app.name)}">
                <span class="w-2 h-2 rounded-full bg-emerald-400"></span> ${escapeHtml(app.name)}
              </span>
              ${app.port ? `<span class="text-[10px] font-mono text-cyan-300 bg-cyan-900/60 px-1.5 py-0.5 rounded border border-cyan-500/30 shrink-0">Port: ${escapeHtml(app.port)}</span>` : '<span class="text-[9px] font-mono text-emerald-400 bg-emerald-500/20 px-1.5 py-0.5 rounded">Çalışıyor</span>'}
            </div>
            <div class="flex items-center justify-between gap-2 pt-0.5 border-t border-emerald-800/40">
              <span class="text-[10px] text-gray-400 font-mono truncate" title="${escapeHtml(app.command)}">${escapeHtml(app.command)}</span>
              <button onclick="killRunningApp('${app.name}', this)" class="px-2 py-0.5 rounded-lg text-[10px] font-bold bg-red-600 hover:bg-red-500 text-white cursor-pointer shrink-0 flex items-center gap-1">
                <i data-lucide="square" class="w-3 h-3 fill-current"></i> Durdur
              </button>
            </div>
          </div>
        `).join('')}

        ${runningTasks.map(t => `
          <div class="p-2.5 rounded-xl bg-purple-950/50 border border-purple-500/30 text-[11px] space-y-1.5">
            <div class="flex items-center justify-between font-semibold text-purple-200">
              <span class="truncate" title="${escapeHtml(t.name)}">${escapeHtml(t.name)}</span>
              <span class="text-[9px] uppercase px-1.5 py-0.2 rounded bg-purple-500/20 text-purple-300 border border-purple-500/30 animate-pulse">İşleniyor</span>
            </div>
            <p class="text-[10px] text-gray-400 leading-tight">${escapeHtml(t.message)}</p>
            <div class="flex items-center justify-end gap-1.5 pt-1 border-t border-purple-800/40">
              <button onclick="stopLiveTask('${t.id}', this)" class="px-2 py-0.5 rounded-lg text-[10px] font-bold bg-red-600 hover:bg-red-500 text-white cursor-pointer flex items-center gap-1">
                <i data-lucide="square" class="w-3 h-3 fill-current"></i> Durdur
              </button>
              <button onclick="deleteLiveTask('${t.id}', this)" title="Listeden sil" class="p-1 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-500/10 cursor-pointer">
                <i data-lucide="trash-2" class="w-3.5 h-3.5"></i>
              </button>
            </div>
          </div>
        `).join('')}

        ${recentCompleted.map(t => `
          <div class="p-2.5 rounded-xl ${t.status === 'completed' ? 'bg-emerald-950/40 border border-emerald-500/30' : t.status === 'cancelled' ? 'bg-gray-800/60 border border-gray-600/40' : 'bg-red-950/40 border border-red-500/30'} text-[11px] space-y-1">
            <div class="flex items-center justify-between font-semibold ${t.status === 'completed' ? 'text-emerald-200' : t.status === 'cancelled' ? 'text-gray-300' : 'text-red-200'}">
              <span class="truncate" title="${escapeHtml(t.name)}">${escapeHtml(t.name)}</span>
              <div class="flex items-center gap-1 shrink-0">
                <span class="text-[9px] uppercase px-1.5 py-0.2 rounded ${t.status === 'completed' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : t.status === 'cancelled' ? 'bg-gray-600/30 text-gray-300 border border-gray-500/30' : 'bg-red-500/20 text-red-300 border border-red-500/30'}">${t.status === 'completed' ? 'Tamamlandı' : t.status === 'cancelled' ? 'Durduruldu' : 'Hata'}</span>
                <button onclick="deleteLiveTask('${t.id}', this)" title="Listeden sil" class="p-0.5 rounded text-gray-500 hover:text-red-400 hover:bg-red-500/10 cursor-pointer">
                  <i data-lucide="x" class="w-3 h-3"></i>
                </button>
              </div>
            </div>
            <p class="text-[10px] text-gray-400 leading-tight truncate">${escapeHtml(t.message)}</p>
          </div>
        `).join('')}
      </div>
    </div>
  `;

  if (window.lucide) lucide.createIcons();
}

async function checkAllRepoUpdates(btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Kontrol Ediliyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/repos/check-all`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Kontrol hatası');
    
    if (data.behindCount > 0) {
      showToast(`📢 ${data.behindCount} repo için yeni güncelleme mevcut!`);
    } else {
      showToast(`✅ Tüm repolar güncel (${data.results ? data.results.length : 0} repo denetlendi).`);
    }

    if (document.getElementById('library-scan-body')) loadLibraryScan();
    if (document.getElementById('tracked-repos-body')) renderTrackedRepos();
  } catch (e) {
    showToast('⚠️ Güncelleme kontrolü hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function installSystemTool(id, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Kuruluyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/system/install`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Kurulum başlatılamadı');
    showToast(data.message || `⚡ Kurulum başlatıldı.`);

    setTimeout(() => {
      loadSystemHealth();
    }, 4000);
  } catch (e) {
    showToast('⚠️ Kurulum hatası: ' + e.message);
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function loadSystemHealth(btnEl) {
  const container = document.getElementById('system-health-grid');
  const summaryBadge = document.getElementById('runtime-summary-badge');
  if (!container) return;

  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Taranıyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/system/health`);
    if (!res.ok) throw new Error('Backend yanıt vermedi');
    const data = await res.json();

    if (summaryBadge) {
      if (data.missing === 0) {
        summaryBadge.textContent = `🟢 ${data.installed}/${data.total} Motor Hazır`;
        summaryBadge.className = 'text-[10px] font-mono font-bold px-2.5 py-1 rounded-full bg-emerald-500/15 text-emerald-400 border border-emerald-500/30';
      } else {
        summaryBadge.textContent = `⚠️ ${data.installed}/${data.total} Kurulu (${data.missing} Eksik)`;
        summaryBadge.className = 'text-[10px] font-mono font-bold px-2.5 py-1 rounded-full bg-amber-500/15 text-amber-400 border border-amber-500/30';
      }
    }

    const runtimes = data.runtimes || [];
    container.innerHTML = runtimes.map(rt => `
      <div class="p-3.5 rounded-xl bg-gray-950/60 border ${rt.installed ? 'border-gray-800 hover:border-emerald-500/40' : 'border-amber-500/30 bg-amber-950/10'} space-y-2 flex flex-col justify-between">
        <div class="space-y-1">
          <div class="flex items-center justify-between gap-1">
            <h4 class="font-bold text-xs ${rt.installed ? 'text-gray-100' : 'text-amber-200'} truncate" title="${escapeHtml(rt.name)}">${escapeHtml(rt.name)}</h4>
            <span class="text-[9px] font-semibold px-2 py-0.5 rounded uppercase shrink-0 ${rt.installed ? 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30' : 'bg-amber-500/15 text-amber-400 border-amber-500/30'}">
              ${rt.installed ? '✓ Kurulu' : '⚠️ Eksik'}
            </span>
          </div>
          <p class="text-[11px] text-gray-400 leading-snug">${escapeHtml(rt.desc)}</p>
        </div>

        <div class="pt-2 border-t border-gray-800/60 text-[10px] font-mono space-y-1.5">
          ${rt.installed ? `
            <div class="text-emerald-400 truncate" title="${escapeHtml(rt.version)}">
              ${escapeHtml(rt.version)}
            </div>
          ` : `
            <div class="space-y-1.5">
              <button onclick="installSystemTool('${rt.id}', this)" class="w-full py-1.5 px-3 rounded-xl text-xs font-bold bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-500 hover:to-orange-500 text-white shadow-lg cursor-pointer transition-all active:scale-95 flex items-center justify-center gap-1.5">
                <i data-lucide="zap" class="w-3.5 h-3.5 fill-current"></i> ⚡ Tek Tıkla Kur
              </button>
              <div class="text-[10px] text-amber-300/80 truncate">
                💡 Kurulum: <code class="bg-gray-900 px-1 py-0.5 rounded text-amber-200">${escapeHtml(rt.installHint)}</code>
              </div>
            </div>
          `}
        </div>
      </div>
    `).join('');

    if (window.lucide) lucide.createIcons();
  } catch (e) {
    container.innerHTML = `<p class="text-xs text-red-400 col-span-full text-center py-6">Sistem teşhisine ulaşılamıyor. ai-toolkit.exe çalışıyor mu?</p>`;
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

// ---------- Session Store (Oturum Koruma / Persistence Across F5) ----------
const SessionStore = {
  get(key, defaultVal) {
    try {
      const val = sessionStorage.getItem(`aitoolkit_session_${key}`);
      return val !== null ? JSON.parse(val) : defaultVal;
    } catch (e) {
      return defaultVal;
    }
  },
  set(key, val) {
    try {
      sessionStorage.setItem(`aitoolkit_session_${key}`, JSON.stringify(val));
    } catch (e) {}
  }
};

let trackedReposPollTimer = null;
function initSettingsPage() {
  loadSystemHealth();
  renderTrackedRepos();
  clearInterval(trackedReposPollTimer);
  trackedReposPollTimer = setInterval(renderTrackedRepos, 5 * 60 * 1000);
}

// ---------- WyvDev Agentic Loop Engine UI Handlers ----------

async function refreshLoopHeartbeat(btnEl) {
  const startTime = Date.now();
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-4 h-4 animate-spin"></i> Sorgulanıyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/loop/heartbeat`);
    const latency = Date.now() - startTime;
    if (!res.ok) throw new Error('Heartbeat alınamadı');
    const data = await res.json();

    const statusVal = document.getElementById('loop-status-val');
    const appsVal = document.getElementById('loop-active-apps-val');
    const idesVal = document.getElementById('loop-ides-val');
    const latencyVal = document.getElementById('loop-latency-val');

    if (statusVal) statusVal.innerText = '🟢 Aktif';
    if (appsVal) appsVal.innerText = `${data.activeAppsCount || 0} Uygulama`;
    if (idesVal) idesVal.innerText = `${data.idePathsCount || 0} Konfigürasyon`;
    if (latencyVal) latencyVal.innerText = `${latency} ms`;

    showToast('🌀 WyvDev Agentic Loop Engine Heartbeat güncellendi.');
  } catch (e) {
    showToast('⚠️ Loop Heartbeat hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function loadLoopEngineConfig() {
  try {
    const res = await fetch(`${API_BASE}/api/loop/config`);
    if (!res.ok) return;
    const cfg = await res.json();

    const healEl = document.getElementById('cfg-auto-heal-enabled');
    const killEl = document.getElementById('cfg-auto-kill-port');
    const retriesEl = document.getElementById('cfg-max-retries');
    const strategyEl = document.getElementById('cfg-cache-strategy');

    if (healEl) healEl.checked = cfg.autoHealEnabled !== false;
    if (killEl) killEl.checked = cfg.autoKillPort !== false;
    if (retriesEl) retriesEl.value = cfg.maxRetries || 3;
    if (strategyEl) strategyEl.value = cfg.cacheStrategy || 'force';
  } catch (e) {}
}

async function saveLoopEngineConfig(btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-4 h-4 animate-spin"></i> Kaydediliyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const cfg = {
      autoHealEnabled: document.getElementById('cfg-auto-heal-enabled')?.checked ?? true,
      autoKillPort: document.getElementById('cfg-auto-kill-port')?.checked ?? true,
      maxRetries: parseInt(document.getElementById('cfg-max-retries')?.value || '3', 10),
      cacheStrategy: document.getElementById('cfg-cache-strategy')?.value || 'force'
    };

    const res = await fetch(`${API_BASE}/api/loop/config`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(cfg)
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Kaydedilemedi');
    showToast(data.message || '🌀 Loop Engine ayarları başarıyla kaydedildi.');
  } catch (e) {
    showToast('⚠️ Ayar kaydetme hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function runLoopVerification(btnEl) {
  const repoName = document.getElementById('verify-repo-select')?.value || '';
  const cmd = document.getElementById('verify-cmd-input')?.value || '';
  const outputEl = document.getElementById('verify-output-log');
  const statusLabel = document.getElementById('verify-status-label');
  const exitCodeLabel = document.getElementById('verify-exit-code-label');

  if (!cmd.trim()) {
    showToast('⚠️ Lütfen bir doğrulama komutu girin.');
    return;
  }

  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-4 h-4 animate-spin"></i> Çalıştırılıyor...';
    if (window.lucide) lucide.createIcons();
  }

  if (outputEl) outputEl.textContent = 'Komut çalıştırılıyor, lütfen bekleyin...';
  if (statusLabel) statusLabel.textContent = 'Çalışıyor...';

  try {
    const res = await fetch(`${API_BASE}/api/loop/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: repoName, command: cmd })
    });
    const data = await res.json();
    if (statusLabel) {
      statusLabel.textContent = data.passed ? '🟢 BAŞARILI (Passed)' : '🔴 BAŞARISIZ (Failed)';
      statusLabel.className = data.passed ? 'font-bold text-emerald-400 font-mono' : 'font-bold text-red-400 font-mono';
    }
    if (exitCodeLabel) exitCodeLabel.textContent = `Exit Code: ${data.exitCode ?? -1}`;
    if (outputEl) outputEl.textContent = data.output || '(Çıktı üretilmedi)';
    showToast(data.passed ? '✅ Doğrulama testi başarılı!' : '⚠️ Doğrulama testi başarısız oldu.');
  } catch (e) {
    if (statusLabel) statusLabel.textContent = '⚠️ Hata';
    if (outputEl) outputEl.textContent = 'Hata: ' + e.message;
    showToast('⚠️ Çalıştırma hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function triggerManualAutoHeal(btnEl) {
  const repoName = document.getElementById('auto-heal-repo-select')?.value;
  if (!repoName) {
    showToast('⚠️ Lütfen iyileştirilecek bir repo seçin.');
    return;
  }

  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-4 h-4 animate-spin"></i> İyileştiriliyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/loop/auto-heal`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: repoName, reason: 'manual_simulator' })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'İyileştirme başarısız');
    showToast(data.message || `🌀 '${repoName}' başarıyla iyileştirildi.`);
    fetchLoopTelemetry();
  } catch (e) {
    showToast('⚠️ İyileştirme hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}

async function fetchLoopTelemetry() {
  const container = document.getElementById('loop-telemetry-output');
  if (!container) return;
  try {
    const res = await fetch(`${API_BASE}/api/activity`);
    if (!res.ok) return;
    const list = await res.json() || [];
    container.textContent = list.map(item => `[${item.timestamp}] ${item.type.toUpperCase()}: ${item.message}`).join('\n') || 'Henüz telemetri verisi yok.';
  } catch (e) {}
}

async function initLoopEnginePage() {
  refreshLoopHeartbeat();
  loadLoopEngineConfig();
  fetchLoopTelemetry();

  try {
    const res = await fetch(`${API_BASE}/api/repos/scan`);
    if (res.ok) {
      const repos = await res.json() || [];
      const verifySelect = document.getElementById('verify-repo-select');
      const healSelect = document.getElementById('auto-heal-repo-select');

      const optionsHtml = repos.map(r => `<option value="${escapeHtml(r.name)}">${escapeHtml(r.name)} (${r.repoTypeLabel || 'Repo'})</option>`).join('');

      if (verifySelect) verifySelect.innerHTML = `<option value="">(Proje Seçilmedi - Ana Dizin)</option>${optionsHtml}`;
      if (healSelect) healSelect.innerHTML = repos.length ? optionsHtml : '<option value="">Hiç repo bulunamadı</option>';
    }
  } catch (e) {}
}

document.addEventListener('DOMContentLoaded', () => {
  // i18n — dil tercihini uygula
  updateLanguageUI();

  // Safe render calls — index.html'e özgü fonksiyonlar diğer sayfalarda yoktur
  try { if (typeof renderMcps === 'function') renderMcps(); } catch (e) {}
  try { if (typeof renderSkills === 'function') renderSkills(); } catch (e) {}
  try { if (typeof renderIdePaths === 'function') renderIdePaths(); } catch (e) {}

  // Restore Yerel Kütüphane filter & sort session state
  const typeFilterEl = document.getElementById('repo-type-filter');
  const sortByEl = document.getElementById('repo-sort-by');
  if (typeFilterEl) typeFilterEl.value = SessionStore.get('repo_type_filter', 'all');
  if (sortByEl) sortByEl.value = SessionStore.get('repo_sort_by', 'name');

  try { setLibraryViewMode(libraryViewMode === 'grid' ? 'grid' : 'table'); } catch (e) {}

  if (document.getElementById('gh-search-results')) {
    const cachedQuery = SessionStore.get('gh_query', '');
    const ghInput = document.getElementById('gh-search-input');
    if (ghInput && cachedQuery) ghInput.value = cachedQuery;
    try { performGithubSearch(cachedQuery); } catch (e) {}
  }

  if (document.getElementById('library-scan-body')) {
    try { loadLibraryScan(); } catch (e) {}
  } else {
    try { updateRepoCountBadge(); } catch (e) {}
  }
  try { if (document.getElementById('loop-status-val')) initLoopEnginePage(); } catch (e) {}
  if (document.getElementById('activity-log-output')) {
    try { loadActivityLog(); } catch (e) {}
    setInterval(() => { try { loadActivityLog(); } catch (e) {} }, 30000);
  }
  try { if (document.getElementById('tracked-repos-body')) initSettingsPage(); } catch (e) {}
  try { if (document.getElementById('marketplace-grid')) initMarketplacePage(); } catch (e) {}

  if (window.lucide) lucide.createIcons();

  // Restore scroll position
  const savedY = SessionStore.get('scroll_y', 0);
  if (savedY > 0) window.scrollTo(0, savedY);

  window.addEventListener('scroll', () => {
    SessionStore.set('scroll_y', window.scrollY);
  });

  // Backend bağlantısı — HER sayfada çalışır
  hydrateFromBackend();
  try { pollLiveTasks(); } catch (e) {}

  setInterval(hydrateFromBackend, 30000);
  try { setInterval(pollLiveTasks, 2000); } catch (e) {}
  window.addEventListener('focus', hydrateFromBackend);
  try { window.addEventListener('focus', pollLiveTasks); } catch (e) {}
});

// ---------- Marketplace Logic ----------

let allMarketplaceItems = [];

async function initMarketplacePage() {
  await loadMarketplaceCatalog();
}

async function loadMarketplaceCatalog() {
  const container = document.getElementById('marketplace-grid');
  if (!container) return;

  try {
    const res = await fetch(`${API_BASE}/api/marketplace/catalog`);
    if (!res.ok) throw new Error('Marketplace verisi alınamadı');
    allMarketplaceItems = await res.json();
    filterMarketplaceItems();
  } catch (e) {
    container.innerHTML = `<div class="col-span-full p-6 text-center text-red-400 text-xs">⚠️ Marketplace yüklenemedi: ${e.message}</div>`;
  }
}

function filterMarketplaceItems() {
  const container = document.getElementById('marketplace-grid');
  if (!container) return;

  const searchInput = document.getElementById('marketplace-search-input');
  const typeSelect = document.getElementById('marketplace-type-filter');

  const query = searchInput ? searchInput.value.toLowerCase().trim() : '';
  const typeFilter = typeSelect ? typeSelect.value : 'all';

  const filtered = allMarketplaceItems.filter(item => {
    const matchType = typeFilter === 'all' || item.type === typeFilter;
    const matchQuery = !query || item.name.toLowerCase().includes(query) || item.repo.toLowerCase().includes(query) || item.desc.toLowerCase().includes(query);
    return matchType && matchQuery;
  });

  renderMarketplaceCatalog(filtered);
}

function renderMarketplaceCatalog(items) {
  const container = document.getElementById('marketplace-grid');
  if (!container) return;

  if (items.length === 0) {
    container.innerHTML = '<div class="col-span-full py-12 text-center text-gray-400 text-xs">Aradığınız kriterlere uygun Marketplace paketi bulunamadı.</div>';
    return;
  }

  let html = '';
  items.forEach(item => {
    const isMcp = item.type === 'mcp';
    const typeBadge = isMcp ? 'bg-cyan-500/20 text-cyan-300 border-cyan-500/30' : 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30';
    const iconName = item.icon || (isMcp ? 'server' : 'sparkles');

    html += `
      <div class="p-4 rounded-2xl bg-gray-900/40 border border-gray-800 hover:border-amber-500/40 transition-all space-y-3 flex flex-col justify-between group">
        <div class="space-y-2.5">
          <div class="flex items-start justify-between gap-2">
            <div class="flex items-center gap-2.5">
              <div class="w-9 h-9 rounded-xl ${isMcp ? 'bg-cyan-500/20 text-cyan-300' : 'bg-amber-500/20 text-amber-300'} border border-gray-800 flex items-center justify-center font-bold text-sm">
                <i data-lucide="${iconName}" class="w-4 h-4"></i>
              </div>
              <div>
                <h3 class="font-bold text-xs text-gray-100 group-hover:text-amber-300 transition-colors">${item.name}</h3>
                <p class="text-[10px] text-gray-400 font-mono">${item.repo}</p>
              </div>
            </div>
            <span class="text-[9px] font-bold px-2 py-0.5 rounded-full border ${typeBadge}">${item.badge || item.type}</span>
          </div>

          <p class="text-xs text-gray-400 leading-relaxed">${item.desc}</p>
        </div>

        <div class="pt-2 border-t border-gray-800/80 flex items-center justify-between gap-2">
          <span class="text-[10px] text-gray-400 flex items-center gap-1 font-semibold">
            <i data-lucide="star" class="w-3 h-3 text-amber-400"></i> ${item.stars || '⭐ Popüler'}
          </span>
          <button onclick="installMarketplaceItem('${item.repo}', '${item.name}', this)" class="px-3 py-1.5 rounded-xl text-xs font-bold bg-amber-500/20 hover:bg-amber-500 text-amber-300 hover:text-gray-950 border border-amber-500/30 transition-all cursor-pointer flex items-center gap-1.5 active:scale-95 shadow-lg">
            <i data-lucide="download" class="w-3.5 h-3.5"></i> 1-Tıkla Yükle & Aktar
          </button>
        </div>
      </div>
    `;
  });

  container.innerHTML = html;
  if (window.lucide) lucide.createIcons();
}

async function installMarketplaceItem(repo, name, btnEl) {
  const originalLabel = btnEl ? btnEl.innerHTML : null;
  if (btnEl) {
    btnEl.disabled = true;
    btnEl.innerHTML = '<i data-lucide="loader-2" class="w-3.5 h-3.5 animate-spin"></i> Kuruluyor...';
    if (window.lucide) lucide.createIcons();
  }

  try {
    const res = await fetch(`${API_BASE}/api/marketplace/install`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ repo, name })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Marketplace kurulumu başlatılamadı');

    showToast(data.message || `🚀 '${name}' indiriliyor ve senkronize ediliyor...`);
    if (typeof pollLiveTasks === 'function') pollLiveTasks();
  } catch (e) {
    showToast('⚠️ Marketplace kurulum hatası: ' + e.message);
  } finally {
    if (btnEl) {
      btnEl.disabled = false;
      btnEl.innerHTML = originalLabel;
      if (window.lucide) lucide.createIcons();
    }
  }
}


