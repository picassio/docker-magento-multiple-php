/* ==========================================================================
   Mage UI — Preact + HTM + Ace + xterm.js
   CDN-loaded, no build step
   ========================================================================== */

const { h, render, html, useState, useEffect, useRef, useCallback } = window._preact;

// ── API helpers ──────────────────────────────────────────────────────────────
const API = '';
async function api(path, opts = {}) {
  const res = await fetch(API + path, { headers: { 'Content-Type': 'application/json', ...opts.headers }, ...opts });
  if (opts.raw) return res;
  return res.json();
}
const GET = p => api(p);
const POST = (p, body) => api(p, { method: 'POST', body: body ? JSON.stringify(body) : undefined });
const PATCH = (p, body) => api(p, { method: 'PATCH', body: JSON.stringify(body) });
const DELETE = p => api(p, { method: 'DELETE' });

// ── Toast ────────────────────────────────────────────────────────────────────
function toast(msg, type = 'info') {
  const el = document.createElement('div');
  el.className = `toast toast-${type}`;
  el.textContent = msg;
  document.getElementById('toasts').appendChild(el);
  setTimeout(() => el.remove(), 4000);
}

// ── Theme ────────────────────────────────────────────────────────────────────
function getTheme() { return localStorage.getItem('theme') || 'dark'; }
function setTheme(t) { document.documentElement.setAttribute('data-theme', t === 'light' ? 'light' : ''); localStorage.setItem('theme', t); }
setTheme(getTheme());

// ── Escape HTML ──────────────────────────────────────────────────────────────
const esc = s => s ? s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;') : '';
const fmtSize = b => b < 1024 ? b+' B' : b < 1048576 ? (b/1024).toFixed(1)+' KB' : (b/1048576).toFixed(1)+' MB';
const portOf = ports => { const m = (ports||'').match(/:(\d+)->/); return m ? ':'+m[1] : ''; };
const svcState = s => { const st = (s.state||s.State||'').toLowerCase(); return st.includes('running') ? 'running' : st.includes('exit')||st.includes('stop') ? 'stopped' : 'other'; };
const appBadge = a => ({magento2:['M2','blue'],magento1:['M1','orange'],wordpress:['WP','green'],laravel:['LV','red']}[a]||['—','purple']);

// ══════════════════════════════════════════════════════════════════════════════
// COMPONENTS
// ══════════════════════════════════════════════════════════════════════════════

// ── Sidebar ──────────────────────────────────────────────────────────────────
function Sidebar({ page, setPage }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const theme = getTheme();
  const nav = p => { setPage(p); setMenuOpen(false); };
  const I = (d) => html`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" dangerouslySetInnerHTML=${{__html:d}}/>`;
  const items = [
    ['/', 'Dashboard', I('<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/>')],
    ['/projects', 'Projects', I('<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>')],
    ['/db', 'Database', I('<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/><path d="M3 12c0 1.66 4 3 9 3s9-1.34 9-3"/>')],
    ['/build', 'Build', I('<path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>')],
    ['/logs', 'Logs', I('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/>')],
    ['/files', 'Files', I('<path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><path d="M13 2v7h7"/>')],
    ['/sql', 'SQL', I('<rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8"/><path d="M12 17v4"/>')],
    ['/terminal', 'Terminal', I('<polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>')],
    ['/settings', 'Settings', I('<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9c.2.65.77 1.09 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>')],
  ];
  const logoSvg = html`<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>`;
  return html`<div>
    <button class="hamburger ${menuOpen ? 'open' : ''}" onClick=${() => setMenuOpen(!menuOpen)}><span/><span/><span/></button>
    <div class="mobile-overlay ${menuOpen ? 'show' : ''}" onClick=${() => setMenuOpen(false)}/>
    <nav class="sidebar ${menuOpen ? 'open' : ''}">
    <div class="logo" onClick=${() => nav('/')}>${logoSvg} <span>Mage UI</span></div>
    ${items.map(([path, label, icon]) => html`
      <a class="nav-item ${page === path ? 'active' : ''}" onClick=${e => { e.preventDefault(); nav(path); }} href="#">
        ${icon} <span>${label}</span>
      </a>
    `)}
    <div class="sidebar-footer">
      <div class="theme-toggle" onClick=${() => { setMenuOpen(false); setTheme(theme === 'dark' ? 'light' : 'dark'); location.reload(); }}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><${() => theme === 'dark' ? html`<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>` : html`<circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>`}/></svg>
        <span>${theme === 'dark' ? 'Dark' : 'Light'}</span>
      </div>
    </div>
  </nav>
  </div>`;
}

// ── Modal ────────────────────────────────────────────────────────────────────
function Modal({ show, onClose, title, children }) {
  if (!show) return null;
  return html`<div class="modal-overlay" onClick=${e => e.target === e.currentTarget && onClose()}>
    <div class="modal"><h2>${title}</h2>${children}</div>
  </div>`;
}

// ── Dashboard ────────────────────────────────────────────────────────────────
function Dashboard() {
  const [services, setServices] = useState([]);
  const [projects, setProjects] = useState([]);
  const load = async () => { setServices(await GET('/api/services') || []); setProjects(await GET('/api/projects') || []); };
  useEffect(() => { load(); const t = setInterval(load, 5000); return () => clearInterval(t); }, []);
  const running = services.filter(s => svcState(s) === 'running').length;
  return html`<div>
    <div class="page-header"><h1>Dashboard</h1><div class="actions">
      <button class="btn btn-success" onClick=${async () => { toast('Starting...'); await POST('/api/services/up'); load(); }}>▶ Start</button>
      <button class="btn btn-danger" onClick=${async () => { await POST('/api/services/stop'); load(); }}>■ Stop</button>
      <button class="btn" onClick=${async () => { await POST('/api/services/down'); load(); }}>⏏ Down</button>
    </div></div>
    <div class="card-header">Services <span class="badge badge-${running===services.length?'green':'orange'}">${running}/${services.length}</span></div>
    <div class="card-grid">${services.map(s => {
      const st = svcState(s);
      return html`<div class="card service-card">
        <div class="dot ${st}"></div>
        <div class="info"><div class="name">${s.service||s.name}${portOf(s.ports) ? html` <span style="color:var(--text2)">${portOf(s.ports)}</span>` : ''}</div><div class="meta">${s.status||s.state}</div></div>
        <button class="btn-icon" onClick=${async () => { toast('Restarting '+s.service); await POST('/api/services/'+(s.service)+'/restart'); load(); }}>↻</button>
      </div>`;
    })}</div>
    <div class="card-header" style="margin-top:24px">Projects <span class="badge badge-blue">${projects.length}</span></div>
    ${projects.length === 0 ? html`<div class="card empty"><div class="icon">📁</div><p>No projects</p></div>` :
      html`<div class="card table-wrap"><table><thead><tr><th>Domain</th><th>Type</th><th>PHP</th><th>DB</th><th>Search</th><th>Status</th></tr></thead><tbody>
        ${projects.map(p => { const [label,color] = appBadge(p.app); return html`<tr><td><b>${p.domain}</b></td><td><span class="badge badge-${color}">${label}</span></td><td>${p.php}</td><td>${p.db_service}</td><td>${p.search}</td><td><span class="badge ${p.enabled?'badge-green':'badge-red'}">${p.enabled?'on':'off'}</span></td></tr>`; })}
      </tbody></table></div>`}
  </div>`;
}

// ── Projects ─────────────────────────────────────────────────────────────────
function Projects() {
  const [projects, setProjects] = useState([]);
  const [showAdd, setShowAdd] = useState(false);
  const load = async () => setProjects(await GET('/api/projects') || []);
  useEffect(() => { load(); }, []);
  const phpOpts = ['php70','php71','php72','php73','php74','php81','php82','php83','php84'];
  const dbOpts = ['mysql','mysql80','mariadb'];
  const searchOpts = ['opensearch','opensearch1','elasticsearch','elasticsearch7','none'];
  return html`<div>
    <div class="page-header"><h1>Projects</h1><div class="actions"><button class="btn btn-primary" onClick=${()=>setShowAdd(true)}>+ Add Project</button></div></div>
    ${projects.length === 0 ? html`<div class="card empty"><div class="icon">📁</div><p>No projects yet</p><button class="btn btn-primary" onClick=${()=>setShowAdd(true)}>Add your first project</button></div>` :
      html`<div class="card table-wrap"><table><thead><tr><th>Domain</th><th>Type</th><th>PHP</th><th>DB</th><th>Search</th><th>Enabled</th><th></th></tr></thead><tbody>
        ${projects.map(p => { const [label,color] = appBadge(p.app); return html`<tr>
          <td><b>${p.domain}</b></td><td><span class="badge badge-${color}">${label}</span></td>
          <td><select class="inline-select" value=${p.php} onChange=${e=>{PATCH('/api/projects/'+p.domain,{php:e.target.value});toast(p.domain+': PHP → '+e.target.value,'success');load();}}>${phpOpts.map(o=>html`<option selected=${o===p.php}>${o}</option>`)}</select></td>
          <td><select class="inline-select" value=${p.db_service} onChange=${e=>{PATCH('/api/projects/'+p.domain,{db_service:e.target.value});toast(p.domain+': DB → '+e.target.value,'success');}}>${dbOpts.map(o=>html`<option selected=${o===p.db_service}>${o}</option>`)}</select></td>
          <td><select class="inline-select" value=${p.search} onChange=${e=>{PATCH('/api/projects/'+p.domain,{search:e.target.value});toast(p.domain+': Search → '+e.target.value,'success');}}>${searchOpts.map(o=>html`<option selected=${o===p.search}>${o}</option>`)}</select></td>
          <td><label class="toggle"><input type="checkbox" checked=${p.enabled} onChange=${e=>{POST('/api/projects/'+p.domain+'/'+(e.target.checked?'enable':'disable'));toast(p.domain+' '+(e.target.checked?'enabled':'disabled'),'success');}}/><span class="slider"></span></label></td>
          <td style="white-space:nowrap"><button class="btn-icon" title="SSL" onClick=${()=>{toast('SSL for '+p.domain);POST('/api/ssl/'+p.domain);}}>🔒</button><button class="btn-icon" style="color:var(--red)" title="Remove" onClick=${async()=>{if(confirm('Remove '+p.domain+'?')){await DELETE('/api/projects/'+p.domain);toast(p.domain+' removed','success');load();}}}>✕</button></td>
        </tr>`; })}
      </tbody></table></div>`}
    <${AddProjectModal} show=${showAdd} onClose=${()=>{setShowAdd(false);load();}} />
  </div>`;
}

function AddProjectModal({ show, onClose }) {
  const [form, setForm] = useState({ domain:'', app:'magento2', php:'php83', db_service:'mysql', search:'opensearch' });
  if (!show) return null;
  const submit = async () => {
    if (!form.domain) { toast('Domain required','error'); return; }
    const res = await POST('/api/projects', form);
    if (res.error) { toast(res.error,'error'); return; }
    toast(form.domain+' added','success'); onClose();
  };
  return html`<${Modal} show=${show} onClose=${onClose} title="Add Project">
    <div class="form-group"><label>Domain</label><input value=${form.domain} onInput=${e=>setForm({...form,domain:e.target.value})} placeholder="mysite.test" style="width:100%"/></div>
    <div class="form-row">
      <div class="form-group"><label>Type</label><select value=${form.app} onChange=${e=>setForm({...form,app:e.target.value})} style="width:100%"><option value="magento2">Magento 2</option><option value="laravel">Laravel</option><option value="wordpress">WordPress</option><option value="default">Default</option></select></div>
      <div class="form-group"><label>PHP</label><select value=${form.php} onChange=${e=>setForm({...form,php:e.target.value})} style="width:100%"><option>php84</option><option>php83</option><option>php82</option><option>php81</option></select></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>DB</label><select value=${form.db_service} onChange=${e=>setForm({...form,db_service:e.target.value})} style="width:100%"><option value="mysql">MySQL 8.4</option><option value="mysql80">MySQL 8.0</option><option value="mariadb">MariaDB</option></select></div>
      <div class="form-group"><label>Search</label><select value=${form.search} onChange=${e=>setForm({...form,search:e.target.value})} style="width:100%"><option value="opensearch">OpenSearch</option><option value="none">None</option></select></div>
    </div>
    <div class="modal-actions"><button class="btn" onClick=${onClose}>Cancel</button><button class="btn btn-primary" onClick=${submit}>Add</button></div>
  <//>`;
}

// ── Database ─────────────────────────────────────────────────────────────────
function DatabasePage() {
  const [dbs, setDbs] = useState([]);
  const load = async () => setDbs(await GET('/api/databases') || []);
  useEffect(() => { load(); }, []);
  return html`<div>
    <div class="page-header"><h1>Database</h1><div class="actions">
      <button class="btn btn-primary" onClick=${()=>{ const n=prompt('Database name:'); if(n){POST('/api/databases',{name:n,service:'mysql'}).then(()=>{toast(n+' created','success');load();})}}}>+ Create</button>
    </div></div>
    ${dbs.length===0 ? html`<div class="card empty"><div class="icon">🗄️</div><p>No databases found</p></div>` :
      html`<div class="card table-wrap"><table><thead><tr><th>Name</th><th>Service</th><th>Size</th><th>Tables</th><th>Actions</th></tr></thead><tbody>
        ${dbs.map(d => html`<tr><td><b>${d.name}</b></td><td><span class="badge badge-blue">${d.service}</span></td><td>${d.size||'—'}</td><td>${d.tables||'—'}</td>
          <td><button class="btn btn-sm" onClick=${async()=>{toast('Exporting...');const r=await POST('/api/databases/'+d.name+'/export?service='+d.service);if(r.download){const a=document.createElement('a');a.href=r.download;a.click();toast('Exported','success');}}}>⬇ Export</button> <button class="btn btn-sm btn-danger" onClick=${async()=>{if(confirm('Drop '+d.name+'?')){await DELETE('/api/databases/'+d.name+'?service='+d.service);toast(d.name+' dropped','success');load();}}}>✕</button></td>
        </tr>`)}
      </tbody></table></div>`}
  </div>`;
}

// ── Build ────────────────────────────────────────────────────────────────────
function BuildPage() {
  const [images, setImages] = useState([]);
  const [log, setLog] = useState('');
  const load = async () => setImages(await GET('/api/images') || []);
  useEffect(() => { load(); }, []);
  const build = (versions) => {
    setLog(''); toast('Building '+versions.join(', ')+'...');
    const ws = new WebSocket(`${location.protocol==='https:'?'wss:':'ws:'}//${location.host}/api/images/build/ws`);
    ws.onopen = () => ws.send(JSON.stringify({ versions }));
    ws.onmessage = e => { const d = JSON.parse(e.data); setLog(l => l + (d.line||'') + '\n'); if (d.stream==='done') { toast('Build done','success'); load(); } };
  };
  return html`<div>
    <div class="page-header"><h1>PHP Images</h1><div class="actions">
      <button class="btn btn-primary" onClick=${()=>build(images.map(i=>i.version))}>▶ Build All</button>
      <button class="btn" onClick=${()=>build(images.filter(i=>!i.built).map(i=>i.version))}>Build Missing</button>
    </div></div>
    <div class="card table-wrap"><table><thead><tr><th>Version</th><th>Image</th><th>Status</th><th>Size</th><th></th></tr></thead><tbody>
      ${images.map(i => html`<tr><td><b>${i.version}</b></td><td style="font-family:var(--mono);font-size:12px">${i.image}</td><td><span class="badge ${i.built?'badge-green':'badge-red'}">${i.built?'built':'—'}</span></td><td>${i.size||'—'}</td><td><button class="btn btn-sm" onClick=${()=>build([i.version])}>${i.built?'↻ Rebuild':'▶ Build'}</button></td></tr>`)}
    </tbody></table></div>
    ${log && html`<div class="card" style="margin-top:16px"><div class="card-header">Build Output</div><pre class="log-viewer">${log}</pre></div>`}
  </div>`;
}

// ── Logs ──────────────────────────────────────────────────────────────────────
function LogsPage() {
  const [services, setServices] = useState([]);
  const [svc, setSvc] = useState('');
  const [output, setOutput] = useState('Loading...');
  const wsRef = useRef(null);
  const logRef = useRef(null);

  // Load services, then auto-fetch first service logs
  useEffect(() => {
    GET('/api/services').then(s => {
      const list = (s||[]).map(x=>x.service).filter(Boolean);
      setServices(list);
      if (list.length) {
        setSvc(list[0]);
        GET('/api/logs/'+list[0]+'?lines=200').then(r => setOutput(r.output||'No logs yet'));
      } else {
        setOutput('No running services');
      }
    });
  }, []);

  const fetch_ = async () => {
    if (!svc) return;
    setOutput('Loading...');
    const r = await GET('/api/logs/'+svc+'?lines=200');
    setOutput(r.output||'No logs');
    if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight;
  };
  const follow = () => {
    if (wsRef.current) wsRef.current.close();
    setOutput('');
    const ws = new WebSocket(`${location.protocol==='https:'?'wss:':'ws:'}//${location.host}/api/logs/${svc}/ws?lines=100`);
    ws.onmessage = e => { const d = JSON.parse(e.data); setOutput(o => o + (d.line||'') + '\n'); if(logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight; };
    wsRef.current = ws;
  };
  useEffect(() => () => { if (wsRef.current) wsRef.current.close(); }, []);
  return html`<div>
    <div class="page-header"><h1>Logs</h1></div>
    <div style="display:flex;gap:12px;margin-bottom:16px;align-items:center">
      <select value=${svc} onChange=${e=>{setSvc(e.target.value); GET('/api/logs/'+e.target.value+'?lines=200').then(r=>setOutput(r.output||'No logs'));}} style="min-width:150px">${services.map(s=>html`<option>${s}</option>`)}</select>
      <button class="btn btn-primary" onClick=${follow}>▶ Follow</button>
      <button class="btn" onClick=${()=>{if(wsRef.current)wsRef.current.close();}}>■ Stop</button>
      <button class="btn" onClick=${fetch_}>Fetch</button>
    </div>
    <div class="card"><pre ref=${logRef} class="log-viewer" style="min-height:400px">${output}</pre></div>
  </div>`;
}

// ── Files ────────────────────────────────────────────────────────────────────
function FilesPage() {
  const [path, setPath] = useState('sources');
  const [files, setFiles] = useState([]);
  const [file, setFile] = useState(null);
  const [projects, setProjects] = useState([]);
  const editorRef = useRef(null);
  const aceRef = useRef(null);

  useEffect(() => { GET('/api/projects').then(p => setProjects(p||[])); }, []);
  useEffect(() => { GET('/api/files?path='+encodeURIComponent(path)).then(f => setFiles(f||[])); }, [path]);

  const viewFile = async (p) => {
    const r = await GET('/api/files/read?path='+encodeURIComponent(p));
    if (r.error) { toast(r.error,'error'); return; }
    setFile(r);
    // Init Ace editor after render
    setTimeout(() => {
      const el = document.getElementById('ace-editor');
      if (el && window.ace) {
        if (aceRef.current) aceRef.current.destroy();
        const editor = ace.edit(el);
        const ext = p.split('.').pop().toLowerCase();
        const modeMap = {php:'php',js:'javascript',css:'css',json:'json',xml:'xml',html:'html',sql:'sql',yml:'yaml',yaml:'yaml',sh:'sh',md:'markdown',env:'text',conf:'text'};
        editor.setTheme('ace/theme/' + (getTheme()==='dark'?'monokai':'chrome'));
        editor.session.setMode('ace/mode/' + (modeMap[ext]||'text'));
        editor.setOptions({ fontSize: 13, showPrintMargin: false, wrap: true });
        editor.setValue(r.content, -1);
        aceRef.current = editor;
      }
    }, 100);
  };

  const saveFile = async () => {
    if (!aceRef.current || !file) return;
    await POST('/api/files/write', { path: file.path, content: aceRef.current.getValue() });
    toast('Saved','success');
  };

  const breadcrumbs = path.split('/');
  const icon = name => { const e = name.split('.').pop().toLowerCase(); const colors = {php:'#777BB3',js:'#F0DB4F',css:'#264de4',json:'#5B9BD5',xml:'#E44D26',sql:'#336791',log:'#6B7280',env:'#EAB308',yml:'#CB171E',yaml:'#CB171E',conf:'#94A3B8',sh:'#4EAA25',md:'#083FA1',html:'#E44D26',txt:'#6B7280'}; return html`<span style="color:${colors[e]||'var(--text3)'};font-weight:700;font-size:10px;width:20px;display:inline-block;text-align:center;font-family:var(--mono)">${(e||'?').toUpperCase().slice(0,3)}</span>`; };

  return html`<div>
    <div class="page-header"><h1>Files</h1></div>
    <div style="display:flex;gap:12px;margin-bottom:16px;align-items:center">
      <select onChange=${e=>setPath(e.target.value)} style="min-width:180px">
        <option value="sources">All Projects</option>
        ${projects.map(p=>html`<option value=${'sources/'+p.domain}>${p.domain}</option>`)}
      </select>
      <div style="font-family:var(--mono);font-size:13px;color:var(--text2)">
        ${breadcrumbs.map((p,i)=>html`<a href="#" style="color:var(--accent)" onClick=${e=>{e.preventDefault();setPath(breadcrumbs.slice(0,i+1).join('/'))}}>${p}</a>${i<breadcrumbs.length-1?' / ':''}`)}
      </div>
    </div>
    <div class="split-layout" style="display:flex;gap:16px">
      <div class="card" style="width:320px;min-height:400px;overflow-y:auto;max-height:70vh;flex-shrink:0">
        ${path!=='sources' && html`<div style="padding:8px 12px;border-bottom:1px solid var(--border);cursor:pointer" onClick=${()=>setPath(path.split('/').slice(0,-1).join('/')||'sources')}>← ..</div>`}
        ${files.map(f => html`<div style="padding:8px 12px;border-bottom:1px solid var(--border);cursor:pointer;display:flex;justify-content:space-between" onClick=${()=>f.isDir ? setPath(f.path) : viewFile(f.path)}>
          <span>${f.isDir ? html`<svg viewBox="0 0 24 24" width="16" height="16" fill="var(--accent)" stroke="none" style="vertical-align:-2px;margin-right:4px"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>` : icon(f.name)} ${f.name}</span><span style="color:var(--text2);font-size:11px">${f.isDir?'':fmtSize(f.size)}</span>
        </div>`)}
      </div>
      <div class="card" style="flex:1;min-height:400px;position:relative">
        ${file ? html`<div>
          <div style="display:flex;justify-content:space-between;align-items:center;padding:8px 12px;border-bottom:1px solid var(--border);background:var(--bg3)">
            <span style="font-family:var(--mono);font-size:12px;color:var(--text2)">${file.path}</span>
            <div style="display:flex;gap:6px"><button class="btn btn-sm btn-primary" onClick=${saveFile}>💾 Save</button><a class="btn btn-sm" href=${'/api/files/download?path='+encodeURIComponent(file.path)} target="_blank">⬇</a></div>
          </div>
          <div id="ace-editor" style="width:100%;height:calc(70vh - 60px)"></div>
        </div>` : html`<div class="empty"><div class="icon">📄</div><p>Select a file</p></div>`}
      </div>
    </div>
  </div>`;
}

// ── SQL Manager ──────────────────────────────────────────────────────────────
function SQLPage() {
  const [dbs, setDbs] = useState([]);
  const [db, setDb] = useState('');
  const [svc, setSvc] = useState('mysql');
  const [tables, setTables] = useState([]);
  const [query, setQuery] = useState('');
  const [result, setResult] = useState(null);
  const [page, setPage] = useState(1);

  useEffect(() => { GET('/api/databases').then(d => { setDbs(d||[]); if(d&&d.length) { setDb(d[0].name); setSvc(d[0].service); } }); }, []);
  useEffect(() => { if(db && svc) GET('/api/dbmanager/tables?db='+db+'&service='+svc).then(t=>setTables(t||[])); }, [db, svc]);

  const selectTable = async (t) => {
    setQuery('SELECT * FROM `'+t+'` LIMIT 50;');
    const r = await GET('/api/dbmanager/data?db='+db+'&service='+svc+'&table='+t+'&page=1&limit=50');
    setResult(r); setPage(1);
  };

  const loadPage = async (t, p) => {
    const r = await GET('/api/dbmanager/data?db='+db+'&service='+svc+'&table='+t+'&page='+p+'&limit=50');
    setResult(r); setPage(p);
  };

  const runQuery = async () => {
    if(!query.trim()) return;
    const r = await POST('/api/dbmanager/query', { db, service: svc, query: query.trim() });
    setResult(r);
  };

  return html`<div>
    <div class="page-header"><h1>SQL Manager</h1></div>
    <div style="display:flex;gap:12px;margin-bottom:16px;align-items:center">
      <select value=${db} onChange=${e=>{ const opt=e.target.options[e.target.selectedIndex]; setDb(e.target.value); setSvc(opt.dataset.svc||'mysql'); }} style="min-width:200px">
        ${dbs.map(d=>html`<option value=${d.name} data-svc=${d.service}>${d.name} (${d.service})</option>`)}
      </select>
      <button class="btn btn-primary" onClick=${runQuery}>▶ Run Query</button>
    </div>
    <div class="card" style="margin-bottom:16px">
      <textarea value=${query} onInput=${e=>setQuery(e.target.value)} spellcheck="false" placeholder="SELECT * FROM ... LIMIT 50;" style="width:100%;min-height:80px;background:var(--bg);color:var(--text);border:none;padding:12px;font-family:var(--mono);font-size:13px;resize:vertical;outline:none"></textarea>
    </div>
    <div class="split-layout" style="display:flex;gap:16px">
      <div class="card" style="width:280px;min-height:300px;overflow-y:auto;max-height:60vh;flex-shrink:0">
        <div style="padding:8px 12px;border-bottom:1px solid var(--border);font-weight:600;font-size:12px;color:var(--text2)">${tables.length} tables</div>
        ${tables.map(t=>html`<div style="padding:6px 12px;border-bottom:1px solid var(--border);cursor:pointer;font-size:13px;display:flex;justify-content:space-between" onClick=${()=>selectTable(t.name)}>
          <span>🗃️ ${t.name}</span><span style="color:var(--text2);font-size:11px">${t.rows} rows</span>
        </div>`)}
      </div>
      <div class="card" style="flex:1;min-height:300px;overflow:auto;max-height:60vh">
        ${result ? html`<div>
          ${result.error ? html`<div style="padding:16px;color:var(--red);font-family:var(--mono)">${result.error}</div>` :
            html`<div class="table-wrap" style="max-height:calc(60vh - 60px);overflow:auto"><table><thead><tr>${(result.columns||[]).map(c=>html`<th>${c}</th>`)}</tr></thead><tbody>
              ${(result.rows||[]).map(r=>html`<tr>${r.map(cell=>html`<td style="font-family:var(--mono);font-size:12px;max-width:300px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${cell===null?html`<span style="color:var(--text2);font-style:italic">NULL</span>`:String(cell).substring(0,200)}</td>`)}</tr>`)}
            </tbody></table></div>`}
          ${result.total > 0 && html`<div style="padding:8px 12px;font-size:12px;color:var(--text2);border-top:1px solid var(--border)">
            ${result.total} rows, page ${result.page||page}/${result.pages||1}
            ${(result.page||page)>1 && html` <button class="btn btn-sm" onClick=${()=>loadPage(query.match(/`(\w+)`/)?.[1],page-1)}>← Prev</button>`}
            ${(result.page||page)<(result.pages||1) && html` <button class="btn btn-sm" onClick=${()=>loadPage(query.match(/`(\w+)`/)?.[1],page+1)}>Next →</button>`}
          </div>`}
          ${result.count >= 0 && !result.total && html`<div style="padding:8px 12px;font-size:12px;color:var(--text2);border-top:1px solid var(--border)">${result.count} row(s)</div>`}
        </div>` : html`<div class="empty"><div class="icon">🗃️</div><p>Select a table or run a query</p></div>`}
      </div>
    </div>
  </div>`;
}

// ── Terminal ─────────────────────────────────────────────────────────────────
function TerminalPage() {
  const termRef = useRef(null);
  const wsRef = useRef(null);

  useEffect(() => {
    if (!window.Terminal) { toast('xterm.js not loaded','error'); return; }
    const term = new window.Terminal({ theme: getTheme()==='dark' ? { background: '#0d1117' } : { background: '#ffffff', foreground: '#1f2328' }, fontFamily: "'SF Mono','Fira Code',monospace", fontSize: 14, cursorBlink: true });
    const fitAddon = new window.FitAddon.FitAddon();
    term.loadAddon(fitAddon);
    term.open(termRef.current);
    fitAddon.fit();

    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${proto}//${location.host}/api/terminal/ws`);
    ws.binaryType = 'arraybuffer';
    ws.onopen = () => {
      // Send resize
      ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
    };
    ws.onmessage = e => {
      if (e.data instanceof ArrayBuffer) term.write(new Uint8Array(e.data));
      else term.write(e.data);
    };
    ws.onclose = () => term.write('\r\n\x1b[31m--- disconnected ---\x1b[0m\r\n');
    term.onData(data => ws.send(data));
    term.onResize(({ cols, rows }) => ws.send(JSON.stringify({ type: 'resize', cols, rows })));

    const resizeObs = new ResizeObserver(() => fitAddon.fit());
    resizeObs.observe(termRef.current);

    wsRef.current = ws;
    return () => { ws.close(); term.dispose(); resizeObs.disconnect(); };
  }, []);

  return html`<div>
    <div class="page-header"><h1>Terminal</h1></div>
    <div class="card" style="padding:8px"><div ref=${termRef} style="height:calc(80vh - 100px)"></div></div>
  </div>`;
}

// ── Settings ─────────────────────────────────────────────────────────────────
function SettingsPage() {
  const [doctor, setDoctor] = useState(null);
  const [env, setEnv] = useState([]);
  const [projects, setProjects] = useState([]);
  useEffect(() => { Promise.all([GET('/api/doctor'),GET('/api/env'),GET('/api/projects')]).then(([d,e,p])=>{setDoctor(d);setEnv(e||[]);setProjects(p||[]);}); }, []);

  const saveEnv = async () => {
    const updates = {};
    document.querySelectorAll('[data-envkey]').forEach(el => { updates[el.dataset.envkey] = el.value; });
    await PATCH('/api/env', updates);
    toast('.env saved','success');
  };

  return html`<div>
    <div class="page-header"><h1>Settings</h1><div class="actions">
      <button class="btn btn-primary" onClick=${()=>{ /* install modal handled below */ }}>🚀 Install</button>
    </div></div>

    <div class="card" style="margin-bottom:16px"><div class="card-header">System Health <button class="btn btn-sm btn-success" onClick=${async()=>{toast('Fixing...');await POST('/api/doctor/fix');const d=await GET('/api/doctor');setDoctor(d);toast('Done','success');}}>🔧 Fix</button></div>
      ${doctor && doctor.checks && doctor.checks.map(ch => ch.status!=='info'||ch.raw.includes(':') ? html`<div class="doctor-check ${ch.status}"><span class="icon">${ch.status==='pass'?'✔':ch.status==='fail'?'✖':'ℹ'}</span><span>${ch.raw}</span></div>` : null)}
    </div>

    <div class="card" style="margin-bottom:16px"><div class="card-header">Xdebug</div>
      <div style="display:flex;gap:16px;flex-wrap:wrap;padding:8px">
        ${['php81','php82','php83','php84'].map(p=>html`<div style="display:flex;align-items:center;gap:8px;padding:8px 12px;background:var(--bg);border-radius:var(--radius-sm)">
          <b>${p}</b> <button class="btn btn-sm btn-success" onClick=${()=>{toast('Xdebug on '+p);POST('/api/xdebug/'+p+'/on');}}>On</button><button class="btn btn-sm" onClick=${()=>{toast('Xdebug off '+p);POST('/api/xdebug/'+p+'/off');}}>Off</button>
        </div>`)}
      </div>
    </div>

    <div class="card"><div class="card-header">.env <button class="btn btn-sm btn-primary" onClick=${saveEnv}>💾 Save</button></div>
      ${env.map(e => e.type==='comment' ? html`<div class="env-row comment">${e.value}</div>` :
        html`<div class="env-row"><span class="env-key">${e.key}</span><span class="env-val"><input data-envkey=${e.key} value=${e.value}/></span></div>`)}
    </div>
  </div>`;
}

// ══════════════════════════════════════════════════════════════════════════════
// APP
// ══════════════════════════════════════════════════════════════════════════════
function App() {
  const [page, setPage] = useState(location.hash.replace('#','') || '/');
  useEffect(() => { const h = () => setPage(location.hash.replace('#','') || '/'); window.addEventListener('hashchange', h); return () => window.removeEventListener('hashchange', h); }, []);
  const nav = p => { location.hash = p; setPage(p); };
  const pages = { '/': Dashboard, '/projects': Projects, '/db': DatabasePage, '/build': BuildPage, '/logs': LogsPage, '/files': FilesPage, '/sql': SQLPage, '/terminal': TerminalPage, '/settings': SettingsPage };
  const Page = pages[page] || Dashboard;
  return html`<div class="app"><${Sidebar} page=${page} setPage=${nav} /><main class="main"><${Page} /></main></div>`;
}

render(html`<${App} />`, document.getElementById('app'));
