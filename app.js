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
  return DEFAULT_MCP_SERVERS;
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
  return DEFAULT_IDE_PATHS;
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
const API_BASE = window.location.protocol === 'file:' ? 'http://127.0.0.1:47651' : '';

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

async function hydrateFromBackend() {
  try {
    const res = await fetch(`${API_BASE}/api/state`);
    if (!res.ok) throw new Error('bad status');
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

    renderMcps();
    renderSkills();
    renderIdePaths();
    if (document.getElementById('library-scan-body')) loadLibraryScan();

    if (Array.isArray(data.prunedRepos) && data.prunedRepos.length) {
      showToast(`🧹 ${data.prunedRepos.length} öğe kaldırıldı — yerel repo/ klasörü silinmişti: ${data.prunedRepos.join(', ')}`);
    }

    if (window.lucide) lucide.createIcons();
  } catch (e) {
    setBackendStatus(false);
  }
}

let pushDebounceTimer = null;
function pushStateToBackend() {
  clearTimeout(pushDebounceTimer);
  pushDebounceTimer = setTimeout(async () => {
    try {
      const bundle = {
        mcpServers: activeMcpServers,
        recommendedRepos: activeRecommendedRepos,
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
  if (!container) return;

  const filtered = selectedCategory === 'All' ? activeMcpServers : activeMcpServers.filter(s => s.category === selectedCategory);

  container.innerHTML = filtered.map(s => {
    const fields = s.env ? Object.keys(s.env) : (s.headers ? Object.keys(s.headers) : []);
    const hasKeys = fields.length > 0;
    const badgeColor = s.badge === 'Resmi' ? 'bg-cyan-500/15 text-cyan-400 border-cyan-500/30' : (s.badge === 'Self-Host' ? 'bg-indigo-500/15 text-indigo-300 border-indigo-500/30' : 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30');

    return `
      <div class="glass-card flex flex-col justify-between space-y-3">
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
  document.getElementById('install-result-title').textContent = ok ? `✅ '${repo}' kuruldu` : `⚠️ '${repo}' kurulamadı`;
  document.getElementById('install-result-command').textContent = command || '';
  document.getElementById('install-result-output').textContent = [output, error].filter(Boolean).join('\n\n') || '(çıktı yok)';
  modal.classList.remove('hidden');
}

function closeInstallResultModal() {
  const modal = document.getElementById('install-result-modal');
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
  if (confirm('Tüm MCP sunucularını, IDE yollarını ve önerilen repoları varsayılan fabrika ayarlarına sıfırlamak istiyor musunuz?')) {
    activeMcpServers = DEFAULT_MCP_SERVERS;
    activeRecommendedRepos = DEFAULT_RECOMMENDED_REPOS;
    activeIdePaths = DEFAULT_IDE_PATHS;
    saveMcpServers(activeMcpServers);
    saveRecommendedRepos(activeRecommendedRepos);
    saveIdePaths(activeIdePaths);
    renderMcps();
    renderSkills();
    renderIdePaths();
    showToast('Tüm ayarlar varsayılan ayarlara sıfırlandı.');
  }
}

// Initial Render
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
      if (!res.ok) throw new Error('backend reddetti');
      setBackendStatus(true);
      showToast('✅ İçe aktarma tamamlandı — git yolu içeren skiller yerelde yoksa otomatik klonlanacak.');
      renderTrackedRepos();
    } catch (e) {
      showToast('⚠️ İçe aktarma hatası: ' + e.message);
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
  } catch (e) {
    setBackendStatus(false);
    tbody.innerHTML = `<tr><td colspan="3" class="text-xs text-red-400 text-center py-6">Backend'e ulaşılamıyor. ai-toolkit.exe çalışıyor mu?</td></tr>`;
  }
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
  const sortBy = document.getElementById('repo-sort-by')?.value || 'name';

  SessionStore.set('repo_type_filter', typeFilter);
  SessionStore.set('repo_sort_by', sortBy);

  // 1. Filter
  let filtered = [...lastScanResults];
  if (typeFilter !== 'all') {
    filtered = filtered.filter(e => e.repoType === typeFilter);
  }

  // 2. Sort
  filtered.sort((a, b) => {
    if (sortBy === 'name') return a.name.localeCompare(b.name);
    if (sortBy === 'running') return (b.isRunning ? 1 : 0) - (a.isRunning ? 1 : 0);
    if (sortBy === 'installed') return (b.isInstalled ? 1 : 0) - (a.isInstalled ? 1 : 0);
    if (sortBy === 'type') return (a.repoType || '').localeCompare(b.repoType || '');
    return 0;
  });

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
    docker: '🐳 Docker'
  };

  const TYPE_BADGE_STYLE = {
    service: 'bg-amber-500/15 text-amber-300 border-amber-500/30',
    mcp: 'bg-cyan-500/15 text-cyan-300 border-cyan-500/30',
    skill: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
    library: 'bg-purple-500/15 text-purple-300 border-purple-500/30',
    other: 'bg-gray-700/40 text-gray-400 border-gray-600/40'
  };

  // Render Table View
  if (tbody) {
    tbody.innerHTML = filtered.map((entry, i) => {
      const runtimes = entry.runtimes || [];
      const typeStyle = TYPE_BADGE_STYLE[entry.repoType] || TYPE_BADGE_STYLE.other;
      const typeBadge = `<span class="text-[10px] font-bold px-2 py-0.5 rounded border ${typeStyle}">${escapeHtml(entry.repoTypeLabel || '📁 Diğer')}</span>`;

      const badges = [
        typeBadge,
        entry.gitStatus ? statusToBadge(entry.gitStatus) : '',
        ...runtimes.map(rt => `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-gray-700/40 text-gray-300 border border-gray-600/40">${RUNTIME_BADGE[rt] || rt}</span>`)
      ].filter(Boolean).join(' ');

      const alreadySkill = activeRecommendedRepos.some(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));
      const alreadyMcp = activeMcpServers.some(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));

      const installs = [];
      if (entry.isRunning) {
        const portTxt = entry.runningPort ? `Port: ${entry.runningPort}` : 'Aktif';
        installs.push(`<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 animate-pulse">🟢 Çalışıyor (${portTxt})</span>`);
      } else if (entry.hasStartError) {
        installs.push(`<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-red-500/15 text-red-400 border border-red-500/30" title="${escapeHtml(entry.startErrorMsg || '')}">⚠️ Çalıştırma Hatası</span>`);
      } else if (entry.isInstalled) {
        installs.push(`<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">✓ Bağımlılıklar Yüklü (%100)</span>`);
      } else {
        installs.push(`<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">⚠️ Yükleme Gerekli</span>`);
        if (runtimes.includes('node')) installs.push(`<button onclick="runRepoAction('${entry.name}', 'npm-install', this)" class="px-2 py-0.5 rounded-lg text-[10px] font-semibold bg-gray-800 hover:bg-gray-700 text-gray-200 cursor-pointer">Kur (npm)</button>`);
        if (runtimes.includes('python')) installs.push(`<button onclick="runRepoAction('${entry.name}', 'pip-install', this)" class="px-2 py-0.5 rounded-lg text-[10px] font-semibold bg-gray-800 hover:bg-gray-700 text-gray-200 cursor-pointer">Kur (pip)</button>`);
        if (runtimes.includes('rust')) installs.push(`<button onclick="runRepoAction('${entry.name}', 'cargo-build', this)" class="px-2 py-0.5 rounded-lg text-[10px] font-semibold bg-gray-800 hover:bg-gray-700 text-gray-200 cursor-pointer">Derle (cargo)</button>`);
        if (runtimes.includes('docker')) installs.push(`<button onclick="runRepoAction('${entry.name}', 'docker-build', this)" class="px-2 py-0.5 rounded-lg text-[10px] font-semibold bg-blue-900/60 hover:bg-blue-800 text-blue-200 cursor-pointer">Docker Build</button>`);
      }

      const actions = [];
      if (entry.isRunning) {
        actions.push(`<button onclick="killRunningApp('${entry.name}', this)" class="px-3 py-1 rounded-xl text-[11px] font-bold bg-red-600 hover:bg-red-500 text-white shadow-lg cursor-pointer transition-all active:scale-95 flex items-center gap-1.5"><i data-lucide="square" class="w-3.5 h-3.5 fill-current"></i> 🛑 Durdur</button>`);
      } else if (entry.hasStartError) {
        actions.push(`<button onclick="repairRepoProject('${entry.name}', ${JSON.stringify(runtimes).replace(/"/g, '&quot;')}, this)" title="Bağımlılıkları ve ortamı zorla yenile/onar" class="px-3 py-1 rounded-xl text-[11px] font-bold bg-gradient-to-r from-amber-600 to-red-600 hover:from-amber-500 hover:to-red-500 text-white shadow-lg cursor-pointer transition-all active:scale-95 flex items-center gap-1.5"><i data-lucide="wrench" class="w-3.5 h-3.5 fill-current"></i> 🔧 Onar (Repair)</button>`);
      } else if (entry.isInstalled || entry.startCommand) {
        const startCmd = entry.startCommand || 'npm start';
        actions.push(`<button onclick="startRepoProject('${entry.name}', null, this)" title="Çalıştır: ${escapeHtml(startCmd)}" class="px-3 py-1 rounded-xl text-[11px] font-bold bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white shadow-lg cursor-pointer transition-all active:scale-95 flex items-center gap-1.5"><i data-lucide="play" class="w-3.5 h-3.5 fill-current"></i> 🚀 Başlat</button>`);
      }

      actions.push(alreadySkill ? `<span class="text-[10px] text-gray-500">Skill ekli</span>` : `<button onclick="addSkillFromScan(${i})" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-emerald-600 hover:bg-emerald-500 text-white cursor-pointer">Skill Ekle</button>`);
      actions.push(alreadyMcp ? `<span class="text-[10px] text-gray-500">MCP ekli</span>` : `<button onclick="addMcpFromScan(${i})" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-cyan-600 hover:bg-cyan-500 text-white cursor-pointer">MCP Yapılandır</button>`);
      if (entry.repo) actions.push(`<button onclick="pullRepo('${entry.name}')" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-gray-800 hover:bg-gray-700 text-gray-300 cursor-pointer">Güncelle (pull)</button>`);
      actions.push(`<button onclick="deleteRepoFolder('${entry.name}')" class="px-2 py-1 rounded-lg text-[11px] font-semibold bg-red-600/80 hover:bg-red-500 text-white cursor-pointer">Klasörü Sil</button>`);

      return `
        <tr class="border-t border-gray-800/60 hover:bg-gray-900/40">
          <td class="py-2.5 px-3 text-xs font-mono text-gray-200 align-middle">
            <div class="font-bold text-gray-100">${escapeHtml(entry.name)}</div>
            <div class="text-[10px] text-gray-500 truncate">${escapeHtml(entry.packageName || entry.path)}</div>
          </td>
          <td class="py-2.5 px-3 space-x-1 space-y-1 align-middle">${badges}</td>
          <td class="py-2.5 px-3 space-x-1 space-y-1 align-middle">${installs.join('')}</td>
          <td class="py-2.5 px-3 text-right space-x-2 space-y-1 align-middle">${actions.join('')}</td>
        </tr>
      `;
    }).join('');
  }

  // Render Grid View
  if (gridContainer) {
    gridContainer.innerHTML = filtered.map((entry, i) => {
      const runtimes = entry.runtimes || [];
      const typeStyle = TYPE_BADGE_STYLE[entry.repoType] || TYPE_BADGE_STYLE.other;

      const alreadySkill = activeRecommendedRepos.some(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));
      const alreadyMcp = activeMcpServers.some(s => s.id === entry.name || (entry.repo && s.repo === entry.repo));

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
              ${entry.isRunning ? `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 animate-pulse">🟢 Çalışıyor (Port: ${entry.runningPort || 'Aktif'})</span>` : entry.isInstalled ? `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">✓ Bağımlılıklar Yüklü</span>` : `<span class="text-[10px] font-semibold px-2 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">⚠️ Yükleme Gerekli</span>`}
              ${runtimes.map(rt => `<span class="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-gray-800 text-gray-300 border border-gray-700">${RUNTIME_BADGE[rt] || rt}</span>`).join('')}
            </div>
          </div>

          <div class="pt-2 border-t border-gray-800/80 space-y-2">
            <div class="flex items-center gap-1.5 flex-wrap">
              ${entry.isRunning ? `<button onclick="killRunningApp('${entry.name}', this)" class="flex-1 py-1.5 px-3 rounded-xl text-xs font-bold bg-red-600 hover:bg-red-500 text-white flex items-center justify-center gap-1 cursor-pointer"><i data-lucide="square" class="w-3.5 h-3.5 fill-current"></i> 🛑 Durdur</button>` : entry.hasStartError ? `<button onclick="repairRepoProject('${entry.name}', ${JSON.stringify(runtimes).replace(/"/g, '&quot;')}, this)" class="flex-1 py-1.5 px-3 rounded-xl text-xs font-bold bg-gradient-to-r from-amber-600 to-red-600 hover:from-amber-500 text-white flex items-center justify-center gap-1 cursor-pointer"><i data-lucide="wrench" class="w-3.5 h-3.5 fill-current"></i> 🔧 Onar</button>` : (entry.isInstalled || entry.startCommand) ? `<button onclick="startRepoProject('${entry.name}', null, this)" class="flex-1 py-1.5 px-3 rounded-xl text-xs font-bold bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 text-white flex items-center justify-center gap-1 cursor-pointer"><i data-lucide="play" class="w-3.5 h-3.5 fill-current"></i> 🚀 Başlat</button>` : ''}
              ${!entry.isInstalled && runtimes.includes('node') ? `<button onclick="runRepoAction('${entry.name}', 'npm-install', this)" class="py-1 px-2 rounded-lg text-[10px] font-semibold bg-gray-800 hover:bg-gray-700 text-gray-200 cursor-pointer">Kur (npm)</button>` : ''}
              ${!entry.isInstalled && runtimes.includes('python') ? `<button onclick="runRepoAction('${entry.name}', 'pip-install', this)" class="py-1 px-2 rounded-lg text-[10px] font-semibold bg-gray-800 hover:bg-gray-700 text-gray-200 cursor-pointer">Kur (pip)</button>` : ''}
            </div>

            <div class="flex items-center justify-between text-[11px] pt-1">
              ${alreadySkill ? `<span class="text-[10px] text-gray-500">Skill ekli</span>` : `<button onclick="addSkillFromScan(${i})" class="text-emerald-400 hover:underline cursor-pointer font-semibold">+ Skill Ekle</button>`}
              ${alreadyMcp ? `<span class="text-[10px] text-gray-500">MCP ekli</span>` : `<button onclick="addMcpFromScan(${i})" class="text-cyan-400 hover:underline cursor-pointer font-semibold">+ MCP Yapılandır</button>`}
              <button onclick="deleteRepoFolder('${entry.name}')" class="text-red-400 hover:underline cursor-pointer">Sil</button>
            </div>
          </div>
        </div>
      `;
    }).join('');
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
  const target = entry.packageName || entry.repo || entry.name;
  activeMcpServers.push({
    id, name: entry.name, type: 'stdio', command: 'npx', args: ['-y', target], repo: entry.repo || undefined,
    desc: `${entry.name} MCP sunucusu (repo/${entry.name} içinden tespit edildi — komutu kontrol edin).`,
    category: 'Yerel Kütüphane', badge: 'Kütüphane', icon: 'folder-git-2', iconColor: 'text-indigo-400', auth: false
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
          <div class="p-2.5 rounded-xl bg-purple-950/50 border border-purple-500/30 text-[11px] space-y-1">
            <div class="flex items-center justify-between font-semibold text-purple-200">
              <span class="truncate" title="${escapeHtml(t.name)}">${escapeHtml(t.name)}</span>
              <span class="text-[9px] uppercase px-1.5 py-0.2 rounded bg-purple-500/20 text-purple-300 border border-purple-500/30 animate-pulse">İşleniyor</span>
            </div>
            <p class="text-[10px] text-gray-400 leading-tight">${escapeHtml(t.message)}</p>
          </div>
        `).join('')}

        ${recentCompleted.map(t => `
          <div class="p-2.5 rounded-xl ${t.status === 'completed' ? 'bg-emerald-950/40 border border-emerald-500/30' : 'bg-red-950/40 border border-red-500/30'} text-[11px] space-y-1">
            <div class="flex items-center justify-between font-semibold ${t.status === 'completed' ? 'text-emerald-200' : 'text-red-200'}">
              <span class="truncate" title="${escapeHtml(t.name)}">${escapeHtml(t.name)}</span>
              <span class="text-[9px] uppercase px-1.5 py-0.2 rounded ${t.status === 'completed' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : 'bg-red-500/20 text-red-300 border border-red-500/30'}">${t.status === 'completed' ? 'Tamamlandı' : 'Hata'}</span>
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
  renderMcps();
  renderSkills();
  renderIdePaths();

  // Restore Yerel Kütüphane filter & sort session state
  const typeFilterEl = document.getElementById('repo-type-filter');
  const sortByEl = document.getElementById('repo-sort-by');
  if (typeFilterEl) typeFilterEl.value = SessionStore.get('repo_type_filter', 'all');
  if (sortByEl) sortByEl.value = SessionStore.get('repo_sort_by', 'name');

  const modeBtn = libraryViewMode === 'grid' ? 'grid' : 'table';
  setLibraryViewMode(modeBtn);

  if (document.getElementById('gh-search-results')) {
    const cachedQuery = SessionStore.get('gh_query', '');
    const ghInput = document.getElementById('gh-search-input');
    if (ghInput && cachedQuery) ghInput.value = cachedQuery;
    performGithubSearch(cachedQuery);
  }

  if (document.getElementById('library-scan-body')) loadLibraryScan();
  if (document.getElementById('loop-status-val')) initLoopEnginePage();
  if (document.getElementById('activity-log-output')) {
    loadActivityLog();
    setInterval(loadActivityLog, 30000);
  }
  if (document.getElementById('tracked-repos-body')) initSettingsPage();
  if (window.lucide) lucide.createIcons();

  // Restore scroll position
  const savedY = SessionStore.get('scroll_y', 0);
  if (savedY > 0) window.scrollTo(0, savedY);

  window.addEventListener('scroll', () => {
    SessionStore.set('scroll_y', window.scrollY);
  });

  hydrateFromBackend();
  pollLiveTasks();

  // Canlı takip ve otomatik senkronizasyon
  setInterval(hydrateFromBackend, 30000);
  setInterval(pollLiveTasks, 2000);
  window.addEventListener('focus', hydrateFromBackend);
  window.addEventListener('focus', pollLiveTasks);
});
