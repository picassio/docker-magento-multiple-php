/* ==========================================================================
   Mage UI — Docker Magento Dashboard
   Vanilla JS, hash routing, REST + WebSocket
   ========================================================================== */

const API = '';  // same origin
let state = { services: [], projects: [], databases: [], images: [], env: [], doctor: null };
let pollTimer = null;
let logWS = null;
let buildWS = null;

// ── API helpers ──────────────────────────────────────────────────────────────
async function api(path, opts = {}) {
  const url = API + path;
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  });
  if (opts.raw) return res;
  return res.json();
}
const GET = (p) => api(p);
const POST = (p, body) => api(p, { method: 'POST', body: body ? JSON.stringify(body) : undefined });
const PATCH = (p, body) => api(p, { method: 'PATCH', body: JSON.stringify(body) });
const DELETE = (p) => api(p, { method: 'DELETE' });

// ── Toast notifications ──────────────────────────────────────────────────────
function toast(msg, type = 'info') {
  const el = document.createElement('div');
  el.className = `toast toast-${type}`;
  el.textContent = msg;
  document.getElementById('toasts').appendChild(el);
  setTimeout(() => el.remove(), 4000);
}

// ── Modal ────────────────────────────────────────────────────────────────────
function showModal(html) {
  document.getElementById('modal-root').innerHTML =
    `<div class="modal-overlay" onclick="if(event.target===this)closeModal()"><div class="modal">${html}</div></div>`;
}
function closeModal() { document.getElementById('modal-root').innerHTML = ''; }

// ── Theme ────────────────────────────────────────────────────────────────────
function toggleTheme() {
  const isDark = !document.documentElement.hasAttribute('data-theme');
  document.documentElement.setAttribute('data-theme', isDark ? 'light' : '');
  if (!isDark) document.documentElement.removeAttribute('data-theme');
  document.getElementById('theme-icon').textContent = isDark ? '☀️' : '🌙';
  document.getElementById('theme-label').textContent = isDark ? 'Light' : 'Dark';
  localStorage.setItem('theme', isDark ? 'light' : 'dark');
}
// Init theme
if (localStorage.getItem('theme') === 'light') toggleTheme();

// ── Router ───────────────────────────────────────────────────────────────────
const routes = {
  '/': renderDashboard,
  '/projects': renderProjects,
  '/db': renderDatabase,
  '/build': renderBuild,
  '/logs': renderLogs,
  '/settings': renderSettings,
};

function navigate() {
  const hash = location.hash.replace('#', '') || '/';
  const render = routes[hash] || routes['/'];

  // Update nav
  document.querySelectorAll('.nav-item').forEach(el => {
    el.classList.toggle('active', el.getAttribute('href') === '#' + hash);
  });

  // Stop polling/ws
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
  if (logWS) { logWS.close(); logWS = null; }
  if (buildWS) { buildWS.close(); buildWS = null; }

  render();
}
window.addEventListener('hashchange', navigate);

// ── Data loading ─────────────────────────────────────────────────────────────
async function loadServices() { state.services = await GET('/api/services').catch(() => []) || []; }
async function loadProjects() { state.projects = await GET('/api/projects').catch(() => []) || []; }
async function loadDatabases() { state.databases = await GET('/api/databases').catch(() => []) || []; }
async function loadImages() { state.images = await GET('/api/images').catch(() => []) || []; }

// ── Helpers ──────────────────────────────────────────────────────────────────
function $(sel) { return document.querySelector(sel); }
function h(tag, attrs = {}, ...children) {
  const el = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'onclick' || k === 'onchange' || k === 'oninput') el[k] = v;
    else if (k === 'html') el.innerHTML = v;
    else if (k === 'class') el.className = v;
    else el.setAttribute(k, v);
  }
  for (const c of children) {
    if (typeof c === 'string') el.appendChild(document.createTextNode(c));
    else if (c) el.appendChild(c);
  }
  return el;
}
const content = () => document.getElementById('content');

function serviceState(svc) {
  const s = (svc.state || svc.State || '').toLowerCase();
  if (s.includes('running') || s.includes('up')) return 'running';
  if (s.includes('exit') || s.includes('dead') || s.includes('stop')) return 'stopped';
  return 'other';
}

function appBadge(app) {
  const map = { magento2: ['M2','blue'], magento1: ['M1','orange'], wordpress: ['WP','green'], laravel: ['LV','red'], default: ['—','purple'] };
  const [label, color] = map[app] || ['?','purple'];
  return `<span class="badge badge-${color}">${label}</span>`;
}

function portFromPorts(ports) {
  if (!ports) return '';
  const m = ports.match(/:(\d+)->/);
  return m ? ':' + m[1] : '';
}

// ══════════════════════════════════════════════════════════════════════════════
// DASHBOARD
// ══════════════════════════════════════════════════════════════════════════════
async function renderDashboard() {
  content().innerHTML = '<div class="spinner" style="margin:40px auto;display:block;width:24px;height:24px"></div>';
  await Promise.all([loadServices(), loadProjects()]);

  const running = state.services.filter(s => serviceState(s) === 'running').length;
  const total = state.services.length;

  let html = `<div class="page-header"><h1>Dashboard</h1><div class="actions">
    <button class="btn btn-success" onclick="dashAction('up')">▶ Start All</button>
    <button class="btn btn-danger" onclick="dashAction('stop')">■ Stop</button>
    <button class="btn" onclick="dashAction('down')">⏏ Down</button>
  </div></div>`;

  // Service cards
  html += `<div class="card-header">Services <span class="badge badge-${running===total?'green':'orange'}">${running}/${total} running</span></div>`;
  html += '<div class="card-grid">';
  for (const svc of state.services) {
    const st = serviceState(svc);
    const port = portFromPorts(svc.ports);
    html += `<div class="card service-card">
      <div class="dot ${st}"></div>
      <div class="info">
        <div class="name">${svc.service || svc.name}${port ? ' <span style="color:var(--text2);font-weight:400">' + port + '</span>' : ''}</div>
        <div class="meta">${svc.status || svc.state || 'unknown'}</div>
      </div>
      <button class="btn-icon" title="Restart" onclick="restartService('${svc.service}')">↻</button>
    </div>`;
  }
  html += '</div>';

  // Projects
  html += `<div class="card-header" style="margin-top:24px">Projects <span class="badge badge-blue">${state.projects.length}</span></div>`;
  if (state.projects.length === 0) {
    html += '<div class="card empty"><div class="icon">📁</div><p>No projects registered</p><button class="btn btn-primary" onclick="location.hash=\'#/projects\'">Add Project</button></div>';
  } else {
    html += '<div class="card table-wrap"><table><thead><tr><th>Domain</th><th>Type</th><th>PHP</th><th>DB</th><th>Search</th><th>Status</th></tr></thead><tbody>';
    for (const p of state.projects) {
      html += `<tr><td><strong>${p.domain}</strong></td><td>${appBadge(p.app)}</td><td>${p.php}</td><td>${p.db_service}</td><td>${p.search}</td>
        <td><span class="badge ${p.enabled ? 'badge-green' : 'badge-red'}">${p.enabled ? 'on' : 'off'}</span></td></tr>`;
    }
    html += '</tbody></table></div>';
  }

  content().innerHTML = html;

  // Poll services every 5s
  pollTimer = setInterval(async () => {
    await loadServices();
    const cards = document.querySelectorAll('.service-card');
    state.services.forEach((svc, i) => {
      if (cards[i]) {
        const dot = cards[i].querySelector('.dot');
        if (dot) { dot.className = 'dot ' + serviceState(svc); }
        const meta = cards[i].querySelector('.meta');
        if (meta) meta.textContent = svc.status || svc.state || '';
      }
    });
  }, 5000);
}

async function dashAction(action) {
  toast(`Running ${action}...`, 'info');
  await POST(`/api/services/${action}`);
  toast(`${action} complete`, 'success');
  renderDashboard();
}

async function restartService(name) {
  toast(`Restarting ${name}...`, 'info');
  await POST(`/api/services/${name}/restart`);
  toast(`${name} restarted`, 'success');
  setTimeout(renderDashboard, 1000);
}

// ══════════════════════════════════════════════════════════════════════════════
// PROJECTS
// ══════════════════════════════════════════════════════════════════════════════
async function renderProjects() {
  content().innerHTML = '<div class="spinner" style="margin:40px auto;display:block;width:24px;height:24px"></div>';
  await loadProjects();

  let html = `<div class="page-header"><h1>Projects</h1><div class="actions">
    <button class="btn btn-primary" onclick="showAddProject()">+ Add Project</button>
  </div></div>`;

  if (state.projects.length === 0) {
    html += '<div class="card empty"><div class="icon">📁</div><p>No projects yet</p><button class="btn btn-primary" onclick="showAddProject()">Add your first project</button></div>';
  } else {
    html += '<div class="card table-wrap"><table><thead><tr><th>Domain</th><th>Type</th><th>PHP</th><th>DB</th><th>Search</th><th>Enabled</th><th></th></tr></thead><tbody>';
    const phpOpts = ['php70','php71','php72','php73','php74','php81','php82','php83','php84'];
    const dbOpts = ['mysql','mysql80','mariadb'];
    const searchOpts = ['opensearch','opensearch1','elasticsearch','elasticsearch7','none'];

    for (const p of state.projects) {
      const mkSel = (field, opts, val) => {
        const options = opts.map(o => `<option ${o===val?'selected':''}>${o}</option>`).join('');
        return `<select class="inline-select" onchange="updateProject('${p.domain}','${field}',this.value)">${options}</select>`;
      };
      html += `<tr>
        <td><strong>${p.domain}</strong></td>
        <td>${appBadge(p.app)}</td>
        <td>${mkSel('php', phpOpts, p.php)}</td>
        <td>${mkSel('db_service', dbOpts, p.db_service)}</td>
        <td>${mkSel('search', searchOpts, p.search)}</td>
        <td><label class="toggle"><input type="checkbox" ${p.enabled?'checked':''} onchange="toggleProject('${p.domain}',this.checked)"><span class="slider"></span></label></td>
        <td><button class="btn-icon" style="color:var(--red)" title="Remove" onclick="removeProject('${p.domain}')">✕</button></td>
      </tr>`;
    }
    html += '</tbody></table></div>';
  }
  content().innerHTML = html;
}

function showAddProject() {
  showModal(`<h2>Add Project</h2>
    <div class="form-group"><label>Domain</label><input id="ap-domain" placeholder="mysite.test" style="width:100%"></div>
    <div class="form-row">
      <div class="form-group"><label>App Type</label><select id="ap-app" style="width:100%">
        <option value="magento2">Magento 2</option><option value="magento1">Magento 1</option>
        <option value="laravel">Laravel</option><option value="wordpress">WordPress</option><option value="default">Default</option>
      </select></div>
      <div class="form-group"><label>PHP</label><select id="ap-php" style="width:100%">
        <option>php84</option><option selected>php83</option><option>php82</option><option>php81</option><option>php74</option>
      </select></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Database</label><select id="ap-db" style="width:100%">
        <option value="mysql">MySQL 8.4</option><option value="mysql80">MySQL 8.0</option><option value="mariadb">MariaDB 11.4</option>
      </select></div>
      <div class="form-group"><label>Search</label><select id="ap-search" style="width:100%">
        <option value="opensearch">OpenSearch 2</option><option value="elasticsearch">ES 8</option>
        <option value="elasticsearch7">ES 7</option><option value="none">None</option>
      </select></div>
    </div>
    <div class="modal-actions">
      <button class="btn" onclick="closeModal()">Cancel</button>
      <button class="btn btn-primary" onclick="addProject()">Add Project</button>
    </div>`);
}

async function addProject() {
  const domain = document.getElementById('ap-domain').value.trim();
  if (!domain) { toast('Domain required', 'error'); return; }
  const body = {
    domain,
    app: document.getElementById('ap-app').value,
    php: document.getElementById('ap-php').value,
    db_service: document.getElementById('ap-db').value,
    search: document.getElementById('ap-search').value,
  };
  const res = await POST('/api/projects', body);
  if (res.error) { toast(res.error, 'error'); return; }
  closeModal();
  toast(`Project ${domain} added`, 'success');
  renderProjects();
}

async function updateProject(domain, field, value) {
  await PATCH(`/api/projects/${domain}`, { [field]: value });
  toast(`${domain}: ${field} → ${value}`, 'success');
}

async function toggleProject(domain, enabled) {
  await POST(`/api/projects/${domain}/${enabled ? 'enable' : 'disable'}`);
  toast(`${domain} ${enabled ? 'enabled' : 'disabled'}`, 'success');
}

async function removeProject(domain) {
  if (!confirm(`Remove project "${domain}"? This deletes the vhost config.`)) return;
  await DELETE(`/api/projects/${domain}`);
  toast(`${domain} removed`, 'success');
  renderProjects();
}

// ══════════════════════════════════════════════════════════════════════════════
// DATABASE
// ══════════════════════════════════════════════════════════════════════════════
async function renderDatabase() {
  content().innerHTML = '<div class="spinner" style="margin:40px auto;display:block;width:24px;height:24px"></div>';
  await loadDatabases();

  let html = `<div class="page-header"><h1>Database</h1><div class="actions">
    <button class="btn btn-primary" onclick="showCreateDB()">+ Create</button>
    <button class="btn" onclick="showImportDB()">⬆ Import</button>
  </div></div>`;

  if (state.databases.length === 0) {
    html += '<div class="card empty"><div class="icon">🗄️</div><p>No databases found. Are DB services running?</p></div>';
  } else {
    html += '<div class="card table-wrap"><table><thead><tr><th>Name</th><th>Service</th><th>Size</th><th>Tables</th><th>Actions</th></tr></thead><tbody>';
    for (const db of state.databases) {
      html += `<tr><td><strong>${db.name}</strong></td><td><span class="badge badge-blue">${db.service}</span></td>
        <td>${db.size || '—'}</td><td>${db.tables || '—'}</td>
        <td>
          <button class="btn btn-sm" onclick="exportDB('${db.name}','${db.service}')">⬇ Export</button>
          <button class="btn btn-sm btn-danger" onclick="dropDB('${db.name}','${db.service}')">✕ Drop</button>
        </td></tr>`;
    }
    html += '</tbody></table></div>';
  }
  content().innerHTML = html;
}

function showCreateDB() {
  showModal(`<h2>Create Database</h2>
    <div class="form-group"><label>Name</label><input id="cdb-name" placeholder="my_database" style="width:100%"></div>
    <div class="form-group"><label>Service</label><select id="cdb-svc" style="width:100%">
      <option value="mysql">mysql</option><option value="mysql80">mysql80</option><option value="mariadb">mariadb</option>
    </select></div>
    <div class="modal-actions">
      <button class="btn" onclick="closeModal()">Cancel</button>
      <button class="btn btn-primary" onclick="createDB()">Create</button>
    </div>`);
}

async function createDB() {
  const name = document.getElementById('cdb-name').value.trim();
  if (!name) { toast('Name required', 'error'); return; }
  const res = await POST('/api/databases', { name, service: document.getElementById('cdb-svc').value });
  if (res.error) { toast(res.error, 'error'); return; }
  closeModal();
  toast(`Database ${name} created`, 'success');
  renderDatabase();
}

async function exportDB(name, service) {
  toast(`Exporting ${name}...`, 'info');
  const res = await POST(`/api/databases/${name}/export?service=${service}`);
  if (res.download) {
    const a = document.createElement('a');
    a.href = res.download;
    a.download = res.file;
    a.click();
    toast(`${name} exported`, 'success');
  } else {
    toast(res.error || 'Export failed', 'error');
  }
}

async function dropDB(name, service) {
  if (!confirm(`Drop database "${name}" on ${service}? This cannot be undone.`)) return;
  await DELETE(`/api/databases/${name}?service=${service}`);
  toast(`${name} dropped`, 'success');
  renderDatabase();
}

function showImportDB() {
  showModal(`<h2>Import Database</h2>
    <div class="form-group"><label>SQL File</label><input id="imp-file" type="file" accept=".sql,.sql.gz"></div>
    <div class="form-row">
      <div class="form-group"><label>Target Database</label><input id="imp-target" placeholder="my_database" style="width:100%"></div>
      <div class="form-group"><label>Service</label><select id="imp-svc" style="width:100%">
        <option value="mysql">mysql</option><option value="mysql80">mysql80</option><option value="mariadb">mariadb</option>
      </select></div>
    </div>
    <div class="modal-actions">
      <button class="btn" onclick="closeModal()">Cancel</button>
      <button class="btn btn-primary" onclick="importDB()">Import</button>
    </div>`);
}

async function importDB() {
  const fileInput = document.getElementById('imp-file');
  const target = document.getElementById('imp-target').value.trim();
  if (!fileInput.files[0] || !target) { toast('File and target required', 'error'); return; }
  const fd = new FormData();
  fd.append('file', fileInput.files[0]);
  fd.append('target', target);
  fd.append('service', document.getElementById('imp-svc').value);
  toast('Importing...', 'info');
  const res = await fetch('/api/databases/import', { method: 'POST', body: fd }).then(r => r.json());
  if (res.error) { toast(res.error, 'error'); return; }
  closeModal();
  toast(`Imported into ${target}`, 'success');
  renderDatabase();
}

// ══════════════════════════════════════════════════════════════════════════════
// BUILD
// ══════════════════════════════════════════════════════════════════════════════
async function renderBuild() {
  content().innerHTML = '<div class="spinner" style="margin:40px auto;display:block;width:24px;height:24px"></div>';
  await loadImages();

  let html = `<div class="page-header"><h1>PHP Images</h1><div class="actions">
    <button class="btn btn-primary" onclick="buildAll()">▶ Build All</button>
    <button class="btn" onclick="buildMissing()">Build Missing</button>
  </div></div>`;

  html += '<div class="card table-wrap"><table><thead><tr><th>Version</th><th>Image</th><th>Status</th><th>Size</th><th>Action</th></tr></thead><tbody>';
  for (const img of state.images) {
    html += `<tr><td><strong>${img.version}</strong></td><td style="font-family:var(--mono);font-size:12px">${img.image}</td>
      <td><span class="badge ${img.built?'badge-green':'badge-red'}">${img.built?'built':'not built'}</span></td>
      <td>${img.size || '—'}</td>
      <td><button class="btn btn-sm" onclick="buildOne('${img.version}')">${img.built ? '↻ Rebuild' : '▶ Build'}</button></td></tr>`;
  }
  html += '</tbody></table></div>';

  html += '<div class="card" style="margin-top:16px"><div class="card-header">Build Output</div><div class="log-viewer" id="build-log" style="min-height:200px"></div></div>';
  content().innerHTML = html;
}

function buildAll() { startBuild(state.images.map(i => i.version)); }
function buildMissing() { startBuild(state.images.filter(i => !i.built).map(i => i.version)); }
function buildOne(v) { startBuild([v]); }

function startBuild(versions) {
  if (versions.length === 0) { toast('Nothing to build', 'info'); return; }
  const logEl = document.getElementById('build-log');
  if (logEl) logEl.textContent = '';
  toast(`Building ${versions.join(', ')}...`, 'info');

  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  buildWS = new WebSocket(`${proto}//${location.host}/api/images/build/ws`);
  buildWS.onopen = () => buildWS.send(JSON.stringify({ versions }));
  buildWS.onmessage = (e) => {
    const data = JSON.parse(e.data);
    if (logEl) {
      logEl.textContent += data.line + '\n';
      logEl.scrollTop = logEl.scrollHeight;
    }
    if (data.stream === 'done') {
      toast(`Build complete (exit: ${data.exitCode})`, data.exitCode === 0 ? 'success' : 'error');
      loadImages().then(() => renderBuild());
    }
  };
}

// ══════════════════════════════════════════════════════════════════════════════
// LOGS
// ══════════════════════════════════════════════════════════════════════════════
async function renderLogs() {
  await loadServices();
  const svcs = state.services.map(s => s.service).filter(Boolean);

  let html = `<div class="page-header"><h1>Logs</h1></div>`;
  html += `<div style="display:flex;gap:12px;margin-bottom:16px;align-items:center">
    <select id="log-service" style="min-width:150px">${svcs.map(s => `<option>${s}</option>`).join('')}</select>
    <select id="log-lines"><option>50</option><option selected>100</option><option>500</option><option>1000</option></select>
    <button class="btn btn-primary" onclick="startLogStream()">▶ Follow</button>
    <button class="btn" onclick="stopLogStream()">■ Stop</button>
    <button class="btn" onclick="fetchLogs()">Fetch</button>
    <input id="log-filter" placeholder="Filter..." style="flex:1" oninput="filterLogs()">
  </div>`;
  html += '<div class="card"><div class="log-viewer" id="log-output" style="min-height:400px"></div></div>';
  content().innerHTML = html;
}

async function fetchLogs() {
  const svc = document.getElementById('log-service').value;
  const lines = document.getElementById('log-lines').value;
  const res = await GET(`/api/logs/${svc}?lines=${lines}`);
  const el = document.getElementById('log-output');
  if (el) { el.textContent = res.output || ''; el.scrollTop = el.scrollHeight; }
}

function startLogStream() {
  stopLogStream();
  const svc = document.getElementById('log-service').value;
  const lines = document.getElementById('log-lines').value;
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  logWS = new WebSocket(`${proto}//${location.host}/api/logs/${svc}/ws?lines=${lines}`);
  const el = document.getElementById('log-output');
  logWS.onmessage = (e) => {
    const data = JSON.parse(e.data);
    if (el) {
      el.textContent += (data.line || '') + '\n';
      el.scrollTop = el.scrollHeight;
    }
  };
  logWS.onclose = () => { if (el) el.textContent += '\n--- stream closed ---\n'; };
  toast(`Streaming ${svc} logs...`, 'info');
}

function stopLogStream() {
  if (logWS) { logWS.close(); logWS = null; }
}

function filterLogs() {
  const filter = document.getElementById('log-filter').value.toLowerCase();
  const el = document.getElementById('log-output');
  if (!el || !filter) return;
  // Simple highlight — just scroll to match
}

// ══════════════════════════════════════════════════════════════════════════════
// SETTINGS
// ══════════════════════════════════════════════════════════════════════════════
async function renderSettings() {
  content().innerHTML = '<div class="spinner" style="margin:40px auto;display:block;width:24px;height:24px"></div>';

  // Load all settings data
  const [envData, doctorData] = await Promise.all([
    GET('/api/env').catch(() => []),
    GET('/api/doctor').catch(() => ({ output: '', checks: [] })),
  ]);
  state.env = envData || [];
  state.doctor = doctorData;

  let html = `<div class="page-header"><h1>Settings</h1></div>`;

  // ── Doctor ──
  html += '<div class="card" style="margin-bottom:16px"><div class="card-header">System Health <button class="btn btn-sm btn-success" onclick="runDoctorFix()">🔧 Auto-fix</button></div>';
  if (state.doctor && state.doctor.checks) {
    for (const c of state.doctor.checks) {
      if (c.status === 'info' && !c.raw.includes(':')) continue;
      const icon = c.status === 'pass' ? '✔' : c.status === 'fail' ? '✖' : 'ℹ';
      html += `<div class="doctor-check ${c.status}"><span class="icon">${icon}</span><span>${c.raw}</span></div>`;
    }
  }
  html += '</div>';

  // ── Xdebug ──
  html += '<div class="card" style="margin-bottom:16px"><div class="card-header">Xdebug</div>';
  const phpVersions = ['php81','php82','php83','php84'];
  html += '<div style="display:flex;gap:16px;flex-wrap:wrap">';
  for (const php of phpVersions) {
    html += `<div style="display:flex;align-items:center;gap:8px;padding:8px 12px;background:var(--bg);border-radius:var(--radius-sm)">
      <strong>${php}</strong>
      <button class="btn btn-sm btn-success" onclick="xdebugToggle('${php}','on')">On</button>
      <button class="btn btn-sm" onclick="xdebugToggle('${php}','off')">Off</button>
    </div>`;
  }
  html += '</div></div>';

  // ── Environment ──
  html += '<div class="card"><div class="card-header">.env Configuration <button class="btn btn-sm btn-primary" onclick="saveEnv()">💾 Save</button></div>';
  html += '<div id="env-editor">';
  for (const entry of state.env) {
    if (entry.type === 'comment') {
      html += `<div class="env-row comment">${entry.value}</div>`;
    } else {
      html += `<div class="env-row"><span class="env-key">${entry.key}</span><span class="env-val"><input data-key="${entry.key}" value="${entry.value || ''}"></span></div>`;
    }
  }
  html += '</div></div>';

  content().innerHTML = html;
}

async function runDoctorFix() {
  toast('Running doctor fix...', 'info');
  await POST('/api/doctor/fix');
  toast('Doctor fix complete', 'success');
  renderSettings();
}

async function xdebugToggle(php, action) {
  toast(`Xdebug ${action} for ${php}...`, 'info');
  await POST(`/api/xdebug/${php}/${action}`);
  toast(`Xdebug ${action} for ${php}`, 'success');
}

async function saveEnv() {
  const updates = {};
  document.querySelectorAll('#env-editor input[data-key]').forEach(inp => {
    updates[inp.dataset.key] = inp.value;
  });
  await PATCH('/api/env', updates);
  toast('.env saved', 'success');
}

// ── Init ─────────────────────────────────────────────────────────────────────
navigate();
