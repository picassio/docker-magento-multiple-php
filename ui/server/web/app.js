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
    ['/services', 'Services', I('<circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>')],
    ['/projects', 'Projects', I('<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>')],
    ['/db', 'Database', I('<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/><path d="M3 12c0 1.66 4 3 9 3s9-1.34 9-3"/>')],
    ['/build', 'Build', I('<path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>')],
    ['/extensions', 'Extensions', I('<path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>')],
    ['/logs', 'Logs', I('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/>')],
    ['/files', 'Files', I('<path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><path d="M13 2v7h7"/>')],
    ['/sql', 'SQL', I('<rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8"/><path d="M12 17v4"/>')],
    ['/mail', 'Mail', I('<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>')],
    ['/search', 'Search', I('<circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>')],
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
      <a class="nav-item ${page === path.split('?')[0] ? 'active' : ''}" onClick=${e => { e.preventDefault(); nav(path); }} href="#">
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
    <div class="page-header"><h1>Dashboard</h1></div>
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
      html`<div class="card table-wrap"><table><thead><tr><th>Domain</th><th>Status</th><th>Type</th><th>PHP</th><th>DB</th><th>Search</th></tr></thead><tbody>
        ${projects.map(p => { const [label,color] = appBadge(p.app); const stColor = {live:'green',partial:'orange',stopped:'red',disabled:'purple'}[p.status]||'red'; return html`<tr><td><b>${p.domain}</b></td><td><span class="badge badge-${stColor}">${p.status||'unknown'}</span></td><td><span class="badge badge-${color}">${label}</span></td><td>${p.php}</td><td>${p.db_service}</td><td>${p.search}</td></tr>`; })}
      </tbody></table></div>`}
  </div>`;
}

// ── Services ─────────────────────────────────────────────────────────────────
function ServicesPage() {
  const [services, setServices] = useState([]);
  const [busy, setBusy] = useState({});
  const [actionLog, setActionLog] = useState('');
  const [actionTarget, setActionTarget] = useState('');
  const load = async () => setServices(await GET('/api/services/all') || []);
  useEffect(() => { load(); const t = setInterval(load, 4000); return () => clearInterval(t); }, []);

  const act = async (svc, action) => {
    setBusy(b => ({...b, [svc]: action}));
    setActionTarget(svc);
    setActionLog(l => l + `\n\u2501\u2501\u2501 ${action.toUpperCase()} ${svc} \u2501\u2501\u2501\n`);
    const r = await POST('/api/services/'+svc+'/'+action);
    setActionLog(l => l + (r.output || 'Done') + '\n');
    if (r.status === 'error' && ((r.output||'').includes('No such image') || (r.output||'').includes('not built yet'))) {
      toast(`Image for ${svc} not built yet. Go to Build page first.`, 'error');
    } else {
      toast(`${svc} ${action}ed`, 'success');
    }
    await load();
    setBusy(b => { const n = {...b}; delete n[svc]; return n; });
  };

  const actAll = async (action) => {
    setActionTarget('all services');
    setActionLog(l => l + `\n\u2501\u2501\u2501 ${action.toUpperCase()} ALL \u2501\u2501\u2501\n`);
    const r = await POST('/api/services/'+action);
    setActionLog(l => l + (r.output || 'Done') + '\n');
    toast(`All services ${action}ed`, 'success');
    load();
  };

  const running = services.filter(s => (s.state||'').toLowerCase().includes('running')).length;
  const stopped = services.length - running;

  return html`<div>
    <div class="page-header"><h1>Services</h1><div class="actions">
      <button class="btn btn-success" onClick=${() => actAll('up')}>\u25b6 Start All</button>
      <button class="btn btn-danger" onClick=${() => actAll('stop')}>\u25a0 Stop All</button>
    </div></div>

    <div style="display:flex;gap:12px;margin-bottom:16px">
      <div class="badge badge-green" style="padding:6px 12px">${running} running</div>
      <div class="badge badge-red" style="padding:6px 12px">${stopped} stopped</div>
    </div>

    <div class="card table-wrap"><table><thead><tr><th>Service</th><th>State</th><th>Status</th><th>Ports</th><th style="text-align:right">Actions</th></tr></thead><tbody>
      ${services.map(s => {
        const isRunning = (s.state||'').toLowerCase().includes('running');
        const isBusy = !!busy[s.service];
        return html`<tr>
          <td><b>${s.service}</b></td>
          <td><span class="badge ${isRunning?'badge-green':'badge-red'}">${isRunning?'running':'stopped'}</span></td>
          <td style="font-size:12px;color:var(--text2)">${s.status||'\u2014'}</td>
          <td style="font-family:var(--mono);font-size:12px">${s.ports||'\u2014'}</td>
          <td style="text-align:right;white-space:nowrap">
            ${isRunning ? html`
              <button class="btn btn-sm btn-danger" onClick=${()=>act(s.service,'stop')} disabled=${isBusy}>${isBusy?'...':'\u25a0 Stop'}</button>
              <button class="btn btn-sm" onClick=${()=>act(s.service,'restart')} disabled=${isBusy}>${isBusy?'...':'\u21bb Restart'}</button>
            ` : html`
              <button class="btn btn-sm btn-success" onClick=${()=>act(s.service,'start')} disabled=${isBusy}>${isBusy?'...':'\u25b6 Start'}</button>
            `}
          </td>
        </tr>`;
      })}
    </tbody></table></div>
    ${actionLog && html`<div class="card" style="margin-top:16px"><div class="card-header">Output \u2014 ${actionTarget}</div><pre class="log-viewer" style="max-height:300px;overflow-y:auto">${actionLog}</pre></div>`}
  </div>`;
}

// ── Projects ─────────────────────────────────────────────────────────────────
function Projects() {
  const [projects, setProjects] = useState([]);
  const [showAdd, setShowAdd] = useState(false);
  const [cmdProject, setCmdProject] = useState(null);
  const [actionLog, setActionLog] = useState('');
  const [actionTarget, setActionTarget] = useState('');
  const [acting, setActing] = useState('');
  const logRef = useRef(null);
  const load = async () => setProjects(await GET('/api/projects') || []);
  useEffect(() => { load(); const t = setInterval(load, 5000); return () => clearInterval(t); }, []);
  useEffect(() => { if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight; }, [actionLog]);
  const phpOpts = ['php70','php71','php72','php73','php74','php81','php82','php83','php84','php85'];
  const dbOpts = ['mysql','mysql80','mariadb'];
  const searchOpts = ['opensearch','opensearch1','elasticsearch','elasticsearch7','none'];
  const redisOpts = ['redis','redis6','none'];
  const rabbitmqOpts = ['none','rabbitmq'];

  const projectAction = async (domain, action) => {
    setActing(domain+':'+action); setActionTarget(domain);
    const header = `\n━━━ ${action.toUpperCase()} ${domain} ━━━\n`;
    setActionLog(l => l + header);
    const r = await POST('/api/projects/'+domain+'/'+action);
    setActionLog(l => l + (r.output || 'Done') + '\n');
    if (r.status === 'error' && ((r.output||'').includes('No such image') || (r.output||'').includes('not built yet'))) {
      toast('Image not built yet. Go to Build page first.', 'error');
    } else {
      toast(domain+' '+r.status, r.status === 'error' ? 'error' : 'success');
    }
    setActing('');
    load();
  };

  return html`<div>
    <div class="page-header"><h1>Projects</h1><div class="actions"><button class="btn btn-primary" onClick=${()=>setShowAdd(true)}>+ Add Project</button></div></div>
    ${projects.length === 0 ? html`<div class="card empty"><div class="icon">📁</div><p>No projects yet</p><button class="btn btn-primary" onClick=${()=>setShowAdd(true)}>Add your first project</button></div>` :
      html`<div class="card table-wrap"><table><thead><tr><th>Domain</th><th>Status</th><th>Type</th><th>PHP</th><th>DB</th><th>Search</th><th>Redis</th><th>RMQ</th><th>Enabled</th><th></th></tr></thead><tbody>
        ${projects.map(p => { const [label,color] = appBadge(p.app); const isBusy = acting.startsWith(p.domain+':'); const stColor = {live:'green',partial:'orange',stopped:'red',disabled:'purple'}[p.status]||'red'; return html`<tr>
          <td><b>${p.domain}</b></td><td><span class="badge badge-${stColor}">${p.status||'unknown'}</span></td><td><span class="badge badge-${color}">${label}</span></td>
          <td><select class="inline-select" value=${p.php} onChange=${e=>{PATCH('/api/projects/'+p.domain,{php:e.target.value});toast(p.domain+': PHP → '+e.target.value,'success');load();}}>${phpOpts.map(o=>html`<option selected=${o===p.php}>${o}</option>`)}</select></td>
          <td><select class="inline-select" value=${p.db_service} onChange=${e=>{PATCH('/api/projects/'+p.domain,{db_service:e.target.value});toast(p.domain+': DB → '+e.target.value,'success');}}>${dbOpts.map(o=>html`<option selected=${o===p.db_service}>${o}</option>`)}</select></td>
          <td><select class="inline-select" value=${p.search} onChange=${e=>{PATCH('/api/projects/'+p.domain,{search:e.target.value});toast(p.domain+': Search → '+e.target.value,'success');}}>${searchOpts.map(o=>html`<option selected=${o===p.search}>${o}</option>`)}</select></td>
          <td><select class="inline-select" value=${p.redis||'redis'} onChange=${e=>{PATCH('/api/projects/'+p.domain,{redis:e.target.value});toast(p.domain+': Redis → '+e.target.value,'success');}}>${redisOpts.map(o=>html`<option selected=${o===(p.redis||'redis')}>${o}</option>`)}</select></td>
          <td><select class="inline-select" value=${p.rabbitmq||'none'} onChange=${e=>{PATCH('/api/projects/'+p.domain,{rabbitmq:e.target.value});toast(p.domain+': RMQ → '+e.target.value,'success');}}>${rabbitmqOpts.map(o=>html`<option selected=${o===(p.rabbitmq||'none')}>${o}</option>`)}</select></td>
          <td><label class="toggle"><input type="checkbox" checked=${p.enabled} onChange=${e=>{POST('/api/projects/'+p.domain+'/'+(e.target.checked?'enable':'disable'));toast(p.domain+' '+(e.target.checked?'enabled':'disabled'),'success');}}/><span class="slider"></span></label></td>
          <td style="white-space:nowrap"><button class="btn btn-sm btn-success" title="Start project services" onClick=${()=>projectAction(p.domain,'start')} disabled=${isBusy}>${isBusy && acting.endsWith(':start') ? '⏳' : '▶ Start'}</button> <button class="btn btn-sm btn-danger" title="Stop project services" onClick=${()=>projectAction(p.domain,'stop')} disabled=${isBusy}>${isBusy && acting.endsWith(':stop') ? '⏳' : '■ Stop'}</button> <button class="btn btn-sm" title="Run command" onClick=${()=>setCmdProject(p)}>⊞ Run</button> <button class="btn-icon" title="SSL" onClick=${async()=>{setActionTarget(p.domain);setActionLog(l=>l+'\n━━━ SSL '+p.domain+' ━━━\n');const r=await POST('/api/ssl/'+p.domain);setActionLog(l=>l+(r.output||'Done')+'\n');toast(r.output&&r.output.includes('Missing')?'mkcert not installed':'SSL done',r.output&&r.output.includes('Missing')?'error':'success');}}>🔒</button><button class="btn-icon" style="color:var(--red)" title="Remove" onClick=${async()=>{if(confirm('Remove '+p.domain+'?')){await DELETE('/api/projects/'+p.domain);toast(p.domain+' removed','success');load();}}}>✕</button></td>
        </tr>`; })}
      </tbody></table></div>`}
    ${(actionLog || acting) && html`<div class="card" style="margin-top:16px"><div class="card-header">Output — ${actionTarget} ${acting ? html`<span class="badge badge-blue" style="margin-left:8px">LIVE</span>` : ''}</div><pre class="log-viewer" ref=${logRef} style="max-height:400px;overflow-y:auto">${actionLog || 'Waiting for output...'}</pre></div>`}
    <${AddProjectModal} show=${showAdd} onClose=${()=>{setShowAdd(false);load();}} />
    <${RunCommandModal} show=${!!cmdProject} onClose=${()=>setCmdProject(null)} project=${cmdProject} />
  </div>`;
}

// ── Run Command Modal ────────────────────────────────────────────────────────
function RunCommandModal({ show, onClose, project }) {
  const [cmd, setCmd] = useState('');
  const [output, setOutput] = useState('');
  const [running, setRunning] = useState(false);
  if (!show || !project) return null;

  const shortcuts = [];
  if (project.app === 'magento2') shortcuts.push(
    { label: 'cache:flush', cmd: 'magento', args: 'cache:flush' },
    { label: 'setup:upgrade', cmd: 'magento', args: 'setup:upgrade' },
    { label: 'di:compile', cmd: 'magento', args: 'setup:di:compile' },
    { label: 'deploy:static', cmd: 'magento', args: 'setup:static-content:deploy -f' },
    { label: 'reindex', cmd: 'magento', args: 'indexer:reindex' },
    { label: 'composer install', cmd: 'composer', args: 'install' },
  );
  if (project.app === 'laravel') shortcuts.push(
    { label: 'migrate', cmd: 'artisan', args: 'migrate' },
    { label: 'cache:clear', cmd: 'artisan', args: 'cache:clear' },
    { label: 'config:cache', cmd: 'artisan', args: 'config:cache' },
    { label: 'route:list', cmd: 'artisan', args: 'route:list' },
    { label: 'queue:work', cmd: 'artisan', args: 'queue:work --once' },
    { label: 'composer install', cmd: 'composer', args: 'install' },
  );
  if (project.app === 'wordpress') shortcuts.push(
    { label: 'plugin list', cmd: 'wp', args: 'plugin list' },
    { label: 'theme list', cmd: 'wp', args: 'theme list' },
    { label: 'cache flush', cmd: 'wp', args: 'cache flush' },
    { label: 'db check', cmd: 'wp', args: 'db check' },
    { label: 'core update', cmd: 'wp', args: 'core update' },
  );
  // Always available — open terminal navigates to Terminal page with project pre-selected
  const openTerminal = () => { onClose(); location.hash = '#/terminal?project=' + encodeURIComponent(project.domain); };

  const run = async (command, argsStr) => {
    const args = [project.domain, ...(argsStr ? argsStr.split(' ') : [])].filter(Boolean);
    setRunning(true);
    setOutput('Running: bin/mage ' + command + ' ' + args.join(' ') + '\n\n');
    try {
      const r = await POST('/api/exec', { command, args });
      setOutput(prev => prev + (r.stdout || '') + (r.stderr ? '\n' + r.stderr : '') + '\n\nExit code: ' + (r.exitCode || 0));
    } catch (e) {
      setOutput(prev => prev + '\nError: ' + e.message);
    }
    setRunning(false);
  };

  const runCustom = () => {
    if (!cmd.trim()) return;
    const parts = cmd.trim().split(' ');
    const command = parts[0];
    const args = parts.slice(1).join(' ');
    run(command, project.domain + (args ? ' ' + args : ''));
  };

  return html`<${Modal} show=${show} onClose=${onClose} title="Run Command — ${project.domain}">
    <div style="margin-bottom:14px">
      <label>Quick Commands</label>
      <div style="display:flex;gap:6px;flex-wrap:wrap">
        ${shortcuts.map(s => html`<button class="btn btn-sm" disabled=${running} onClick=${() => run(s.cmd, s.args)}>${s.label}</button>`)}
        <button class="btn btn-sm btn-success" onClick=${openTerminal}>▶ Open Terminal</button>
      </div>
    </div>
    <div style="margin-bottom:14px">
      <label>Custom Command</label>
      <div style="display:flex;gap:8px">
        <input value=${cmd} onInput=${e => setCmd(e.target.value)} onKeyDown=${e => e.key === 'Enter' && runCustom()} placeholder="composer require package/name" style="flex:1"/>
        <button class="btn btn-primary btn-sm" disabled=${running} onClick=${runCustom}>${running ? html`<span class="spinner"/>` : 'Run'}</button>
      </div>
      <div style="font-size:11px;color:var(--text3);margin-top:4px">Available: composer, magento, artisan, wp, shell</div>
    </div>
    ${output && html`<pre class="log-viewer" style="max-height:300px;margin-top:8px">${output}</pre>`}
    <div class="modal-actions"><button class="btn" onClick=${onClose}>Close</button></div>
  <//>`;
}

function AddProjectModal({ show, onClose }) {
  const [form, setForm] = useState({ domain:'', app:'magento2', php:'php83', db_service:'mysql', search:'opensearch', redis:'redis', rabbitmq:'none' });
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
      <div class="form-group"><label>PHP</label><select value=${form.php} onChange=${e=>setForm({...form,php:e.target.value})} style="width:100%"><option>php85</option><option>php84</option><option>php83</option><option>php82</option><option>php81</option></select></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>DB</label><select value=${form.db_service} onChange=${e=>setForm({...form,db_service:e.target.value})} style="width:100%"><option value="mysql">MySQL 8.4</option><option value="mysql80">MySQL 8.0</option><option value="mariadb">MariaDB</option></select></div>
      <div class="form-group"><label>Search</label><select value=${form.search} onChange=${e=>setForm({...form,search:e.target.value})} style="width:100%"><option value="opensearch">OpenSearch 2.x</option><option value="opensearch1">OpenSearch 1.3</option><option value="elasticsearch">ES 8.x</option><option value="elasticsearch7">ES 7.x</option><option value="none">None</option></select></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Redis</label><select value=${form.redis} onChange=${e=>setForm({...form,redis:e.target.value})} style="width:100%"><option value="redis">Redis 7.4</option><option value="redis6">Redis 6.2</option><option value="none">None</option></select></div>
      <div class="form-group"><label>RabbitMQ</label><select value=${form.rabbitmq} onChange=${e=>setForm({...form,rabbitmq:e.target.value})} style="width:100%"><option value="none">None</option><option value="rabbitmq">RabbitMQ</option></select></div>
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
  const [building, setBuilding] = useState(false);
  const [buildTarget, setBuildTarget] = useState('');
  const [extensions, setExtensions] = useState('');
  const logRef = useRef(null);
  const load = async () => setImages(await GET('/api/images') || []);
  useEffect(() => { load(); checkActiveBuild(); }, []);
  useEffect(() => { if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight; }, [log]);
  const checkActiveBuild = async () => {
    const status = await GET('/api/images/build/status');
    if (status && status.active) {
      setBuilding(true); setBuildTarget(status.target); setLog('');
      const ws = new WebSocket(`${location.protocol==='https:'?'wss:':'ws:'}//${location.host}/api/images/build/reconnect/ws`);
      ws.onmessage = e => { const d = JSON.parse(e.data); setLog(l => l + (d.line||'') + '\n'); if (d.stream==='done') { setBuilding(false); toast('Build done','success'); load(); } };
      ws.onerror = () => setBuilding(false);
      ws.onclose = () => setBuilding(false);
    }
  };
  const build = (versions) => {
    setLog(''); setBuilding(true); setBuildTarget(versions.join(', '));
    const ws = new WebSocket(`${location.protocol==='https:'?'wss:':'ws:'}//${location.host}/api/images/build/ws`);
    ws.onopen = () => ws.send(JSON.stringify({ versions, extensions: extensions.trim() }));
    ws.onmessage = e => { const d = JSON.parse(e.data); setLog(l => l + (d.line||'') + '\n'); if (d.stream==='done') { setBuilding(false); toast('Build done','success'); load(); } };
    ws.onerror = () => { setBuilding(false); toast('Build connection error','error'); };
    ws.onclose = () => { setBuilding(false); };
  };
  return html`<div>
    <div class="page-header"><h1>PHP Images</h1><div class="actions">
      <button class="btn btn-primary" onClick=${()=>build(images.map(i=>i.version))} disabled=${building}>${building ? '\u23f3 Building...' : '\u25b6 Build All'}</button>
      <button class="btn" onClick=${()=>build(images.filter(i=>!i.built).map(i=>i.version))} disabled=${building}>Build Missing</button>
    </div></div>
    <div class="card" style="margin-bottom:16px;padding:16px">
      <div style="display:flex;gap:12px;align-items:center;flex-wrap:wrap">
        <label style="font-weight:600;white-space:nowrap">Bake extensions into image:</label>
        <input class="input" style="flex:1;min-width:200px" placeholder="e.g. redis imagick newrelic (space-separated, optional)"
          value=${extensions} onChange=${e => setExtensions(e.target.value)} />
      </div>
      <div style="margin-top:8px;font-size:12px;opacity:0.7">Extensions listed here are compiled into the image and persist across container restarts.</div>
    </div>
    ${building && html`<div class="card" style="margin-bottom:16px;padding:16px;border-left:3px solid var(--primary)">
      <div style="display:flex;align-items:center;gap:10px">
        <span class="spinner"></span>
        <span><b>Building ${buildTarget}</b> \u2014 streaming output below...</span>
      </div>
    </div>`}
    <div class="card table-wrap"><table><thead><tr><th>Version</th><th>Image</th><th>Status</th><th>Size</th><th></th></tr></thead><tbody>
      ${images.map(i => html`<tr><td><b>${i.version}</b></td><td style="font-family:var(--mono);font-size:12px">${i.image}</td><td><span class="badge ${i.built?'badge-green':'badge-red'}">${i.built?'built':'\u2014'}</span></td><td>${i.size||'\u2014'}</td><td><button class="btn btn-sm" onClick=${()=>build([i.version])} disabled=${building}>${i.built?'\u21bb Rebuild':'\u25b6 Build'}</button></td></tr>`)}
    </tbody></table></div>
    ${(log || building) && html`<div class="card" style="margin-top:16px"><div class="card-header">Build Output ${building ? html`<span class="badge badge-blue" style="margin-left:8px">LIVE</span>` : ''}</div><pre class="log-viewer" ref=${logRef} style="max-height:500px;overflow-y:auto">${log || 'Waiting for output...'}</pre></div>`}
  </div>`;
}

// ── Extensions ────────────────────────────────────────────────────────────────
function ExtensionsPage() {
  const [allExts, setAllExts] = useState([]);
  const [selected, setSelected] = useState('');
  const [installing, setInstalling] = useState(false);
  const [log, setLog] = useState('');
  const [newExt, setNewExt] = useState('');

  const load = async () => {
    const data = await GET('/api/extensions') || [];
    setAllExts(data);
    if (!selected && data.length > 0) setSelected(data[0].service);
  };
  useEffect(() => { load(); }, []);

  const current = allExts.find(e => e.service === selected);
  const extensions = current ? current.extensions : [];

  const installExt = () => {
    const exts = newExt.trim().split(/[\s,]+/).filter(Boolean);
    if (exts.length === 0) return;
    setInstalling(true); setLog('');
    toast('Installing ' + exts.join(', ') + ' on ' + selected + '...');
    const ws = new WebSocket(`${location.protocol==='https:'?'wss:':'ws:'}//${location.host}/api/extensions/install/ws`);
    ws.onopen = () => ws.send(JSON.stringify({ service: selected, extensions: exts }));
    ws.onmessage = e => {
      const d = JSON.parse(e.data);
      setLog(l => l + (d.line||'') + '\n');
      if (d.stream === 'done') { toast('Install complete', 'success'); setInstalling(false); setNewExt(''); load(); }
    };
    ws.onerror = () => { setInstalling(false); toast('WebSocket error', 'error'); };
  };

  const toggleExt = async (ext, enable) => {
    const endpoint = enable ? '/api/extensions/enable' : '/api/extensions/disable';
    await POST(endpoint, { service: selected, extension: ext.name });
    toast(`${ext.name} ${enable ? 'enabled' : 'disabled'} on ${selected}`, 'success');
    load();
  };

  return html`<div>
    <div class="page-header"><h1>PHP Extensions</h1></div>

    <div class="card" style="margin-bottom:16px;padding:16px">
      <div style="display:flex;gap:12px;align-items:center;flex-wrap:wrap">
        <select class="input" value=${selected} onChange=${e => setSelected(e.target.value)} style="width:auto">
          ${allExts.map(s => html`<option value=${s.service}>${s.service}</option>`)}
        </select>
        <input class="input" style="flex:1;min-width:200px" placeholder="Extension names (space-separated, e.g. redis imagick mongodb)"
          value=${newExt} onChange=${e => setNewExt(e.target.value)}
          onKeyDown=${e => e.key === 'Enter' && installExt()} />
        <button class="btn btn-primary" onClick=${installExt} disabled=${installing}>
          ${installing ? '\u23f3 Installing...' : '\u2795 Install'}
        </button>
      </div>
    </div>

    <div class="card table-wrap">
      <div class="card-header">Enabled on ${selected} (${extensions.length} extensions)</div>
      <table><thead><tr><th>Extension</th><th>Type</th><th></th></tr></thead><tbody>
        ${extensions.map(ext => html`<tr>
          <td><b>${ext.name}</b></td>
          <td><span class="badge ${ext.type==='zend'?'badge-orange':'badge-blue'}">${ext.type}</span></td>
          <td><button class="btn btn-sm btn-danger" onClick=${()=>toggleExt(ext, false)} title="Disable">\u2716</button></td>
        </tr>`)}
      </tbody></table>
    </div>

    ${log && html`<div class="card" style="margin-top:16px"><div class="card-header">Install Output</div><pre class="log-viewer">${log}</pre></div>`}
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
  const [tab, setTab] = useState('phpmyadmin');
  const [dbs, setDbs] = useState([]);
  const [db, setDb] = useState('');
  const [svc, setSvc] = useState('mysql');
  const [tables, setTables] = useState([]);
  const [query, setQuery] = useState('');
  const [result, setResult] = useState(null);
  const [toolsUp, setToolsUp] = useState(false);

  useEffect(() => {
    GET('/api/databases').then(d => { setDbs(d||[]); if(d&&d.length){setDb(d[0].name);setSvc(d[0].service);} });
    // Check if debug tools running
    GET('/api/services').then(s => { setToolsUp((s||[]).some(x => x.service === 'phpmyadmin')); });
  }, []);
  useEffect(() => { if(db&&svc&&tab==='query') GET('/api/dbmanager/tables?db='+db+'&service='+svc).then(t=>setTables(t||[])); }, [db,svc,tab]);

  const startTools = async () => {
    toast('Starting phpMyAdmin + Redis Commander...');
    await POST('/api/debug/start');
    setToolsUp(true);
    toast('Tools started','success');
  };

  const runQuery = async () => {
    if(!query.trim()) return;
    setResult(null);
    const r = await POST('/api/dbmanager/query', {db, service:svc, query:query.trim()});
    setResult(r);
  };

  const pmaUrl = '/phpmyadmin/';
  const redisUrl = '/redis-commander/';

  return html`<div>
    <div class="page-header"><h1>SQL & Data</h1><div class="actions">
      ${!toolsUp && html`<button class="btn btn-success" onClick=${startTools}>Start phpMyAdmin + Redis</button>`}
    </div></div>

    <div style="display:flex;gap:0;margin-bottom:16px">
      ${[['phpmyadmin','phpMyAdmin'],['redis','Redis'],['query','Quick Query']].map(([id,label],i,a)=>
        html`<button class="btn ${tab===id?'btn-primary':''}" style="border-radius:${i===0?'var(--radius-sm) 0 0 var(--radius-sm)':i===a.length-1?'0 var(--radius-sm) var(--radius-sm) 0':'0'};margin-left:${i>0?'-1px':'0'}" onClick=${()=>setTab(id)}>${label}</button>`
      )}
    </div>

    ${tab === 'phpmyadmin' && html`<div class="card" style="overflow:hidden">
      ${toolsUp ? html`<iframe src=${pmaUrl} style="width:100%;height:calc(80vh - 160px);border:none"/>` :
        html`<div class="empty" style="padding:40px"><p>phpMyAdmin is not running</p><button class="btn btn-primary" onClick=${startTools}>Start Debug Tools</button></div>`}
      <div style="padding:8px 14px;border-top:1px solid var(--border);font-size:12px;color:var(--text3);display:flex;justify-content:space-between;align-items:center">
        <span>phpMyAdmin — browse, query, manage all databases</span>
        <a href=${pmaUrl} target="_blank" class="btn btn-sm">Open in new tab</a>
      </div>
    </div>`}

    ${tab === 'redis' && html`<div class="card" style="overflow:hidden">
      ${toolsUp ? html`<iframe src=${redisUrl} style="width:100%;height:calc(80vh - 160px);border:none"/>` :
        html`<div class="empty" style="padding:40px"><p>Redis Commander is not running</p><button class="btn btn-primary" onClick=${startTools}>Start Debug Tools</button></div>`}
      <div style="padding:8px 14px;border-top:1px solid var(--border);font-size:12px;color:var(--text3);display:flex;justify-content:space-between;align-items:center">
        <span>Redis Commander — keys, TTL, memory usage</span>
        <a href=${redisUrl} target="_blank" class="btn btn-sm">Open in new tab</a>
      </div>
    </div>`}

    ${tab === 'query' && html`<div>
      <div style="display:flex;gap:10px;margin-bottom:10px;align-items:center">
        <select value=${db} onChange=${e=>{const o=e.target.options[e.target.selectedIndex];setDb(e.target.value);setSvc(o.dataset.svc||'mysql');}} style="min-width:180px">
          ${dbs.map(d=>html`<option value=${d.name} data-svc=${d.service}>${d.name} (${d.service})</option>`)}
        </select>
        <button class="btn btn-primary btn-sm" onClick=${runQuery}>▶ Run</button>
      </div>
      <div class="card" style="margin-bottom:10px">
        <textarea value=${query} onInput=${e=>setQuery(e.target.value)} onKeyDown=${e=>{if(e.ctrlKey&&e.key==='Enter')runQuery();}} spellcheck="false" placeholder="SELECT * FROM ... LIMIT 50;  (Ctrl+Enter to run)" style="width:100%;min-height:60px;background:var(--bg);color:var(--text);border:none;padding:10px;font-family:var(--mono);font-size:13px;resize:vertical;outline:none"/>
      </div>
      <div class="split-layout">
        <div class="card" style="width:220px;min-height:200px;overflow-y:auto;max-height:45vh;flex-shrink:0">
          <div style="padding:6px 10px;border-bottom:1px solid var(--border);font-size:11px;color:var(--text3);font-weight:600">${tables.length} TABLES</div>
          ${tables.map(t=>html`<div style="padding:4px 10px;border-bottom:1px solid var(--border);cursor:pointer;font-size:11px;font-family:var(--mono)" onClick=${()=>{setQuery('SELECT * FROM \`'+t.name+'\` LIMIT 50;');}}>${t.name}</div>`)}
        </div>
        <div class="card" style="flex:1;min-height:200px;overflow:auto;max-height:45vh">
          ${result ? html`<div>
            ${result.error ? html`<pre style="padding:10px;color:var(--red);font-family:var(--mono);font-size:12px">${result.error}</pre>` :
              html`<div class="table-wrap"><table><thead><tr>${(result.columns||[]).map(c=>html`<th>${c}</th>`)}</tr></thead><tbody>
                ${(result.rows||[]).map(r=>html`<tr>${r.map(cell=>html`<td style="font-family:var(--mono);font-size:11px;max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${cell===null?html`<i style="color:var(--text3)">NULL</i>`:String(cell).substring(0,100)}</td>`)}</tr>`)}\n              </tbody></table></div>`}
            <div style="padding:6px 10px;font-size:11px;color:var(--text3);border-top:1px solid var(--border)">${result.count||0} row(s)</div>
          </div>` : html`<div class="empty" style="padding:20px"><p style="font-size:12px">Run a query</p></div>`}
        </div>
      </div>
    </div>`}
  </div>`;
}

// ── Mail (Mailpit) ──────────────────────────────────────────────────────────
function MailPage() {
  const mailUrl = '/mailpit/';
  const directUrl = location.protocol+'//'+location.hostname+':8025';
  return html`<div>
    <div class="page-header"><h1>Mail</h1><div class="actions">
      <a href=${directUrl} target="_blank" class="btn btn-sm">Open Mailpit ↗</a>
    </div></div>
    <div class="card" style="overflow:hidden">
      <iframe src=${mailUrl} style="width:100%;height:calc(85vh - 100px);border:none"/>
    </div>
    <div style="padding:8px 0;font-size:12px;color:var(--text3)">SMTP: mailpit:1025 — All emails sent from PHP are captured here</div>
  </div>`;
}

// ── OpenSearch Dashboards ────────────────────────────────────────────────────
function SearchPage() {
  const [tab, setTab] = useState('dashboards');
  const [services, setServices] = useState({});
  const [starting, setStarting] = useState('');
  const [log, setLog] = useState('');
  const checkStatus = () => GET('/api/services').then(s => {
    if (!s) return;
    const m = {};
    s.forEach(x => m[x.service] = (x.state||'').includes('running'));
    setServices(m);
  });
  useEffect(() => { checkStatus(); const t = setInterval(checkStatus, 5000); return () => clearInterval(t); }, []);

  const startTool = async (api, label) => {
    setStarting(label); setLog(l => l + '\n━━━ START ' + label + ' ━━━\n');
    const r = await POST('/api/' + api + '/start');
    setLog(l => l + (r.output || 'Done') + '\n');
    setStarting('');
    toast(r.status === 'started' ? label + ' started' : (r.output||'Error'), r.status === 'started' ? 'success' : 'error');
    checkStatus();
  };
  const stopTool = async (api, label) => {
    await POST('/api/' + api + '/stop');
    toast(label + ' stopped', 'success');
    checkStatus();
  };

  const engines = ['opensearch','opensearch1','elasticsearch','elasticsearch7'];
  const dashboards = ['opensearch-dashboards','kibana','kibana7'];

  const tabs = [
    { id: 'dashboards', label: 'OpenSearch Dashboards', url: '/opensearch-dashboards/app/dev_tools', svc: 'opensearch-dashboards', api: 'dashboards', needs: 'OpenSearch' },
    { id: 'kibana', label: 'Kibana 8.x', url: '/kibana/app/dev_tools', svc: 'kibana', api: 'kibana', needs: 'Elasticsearch 8.x' },
    { id: 'kibana7', label: 'Kibana 7.x', url: '/kibana7/app/dev_tools', svc: 'kibana7', api: 'kibana', needs: 'Elasticsearch 7.x' },
  ];
  const active = tabs.find(t => t.id === tab) || tabs[0];
  const isUp = services[active.svc];

  return html`<div>
    <div class="page-header"><h1>Search</h1><div class="actions">
      ${isUp && html`<a href=${active.url} target="_blank" class="btn btn-sm">Open in new tab ↗</a>`}
    </div></div>

    <div style="display:flex;gap:10px;margin-bottom:12px;flex-wrap:wrap">
      ${engines.filter(e => services[e] !== undefined).map(e => html`<div class="badge badge-${services[e]?'green':'red'}" style="padding:6px 12px">${e} ${services[e]?'● running':'○ stopped'}</div>`)}
    </div>

    <div style="display:flex;gap:0;margin-bottom:16px">
      ${tabs.map((t,i,a) => html`<button class="btn ${tab===t.id?'btn-primary':''}" style="border-radius:${i===0?'var(--radius-sm) 0 0 var(--radius-sm)':i===a.length-1?'0 var(--radius-sm) var(--radius-sm) 0':'0'};margin-left:${i>0?'-1px':'0'}" onClick=${()=>setTab(t.id)}>${t.label}</button>`)}
    </div>

    ${isUp ? html`<div class="card" style="overflow:hidden">
      <iframe src=${active.url} style="width:100%;height:calc(80vh - 200px);border:none"/>
      <div style="padding:8px 14px;border-top:1px solid var(--border);font-size:12px;display:flex;justify-content:space-between;align-items:center">
        <span>${active.label} — connected to ${active.needs}</span>
        <button class="btn btn-sm btn-danger" onClick=${() => stopTool(active.api, active.label)}>■ Stop</button>
      </div>
    </div>`
    : html`<div class="card empty" style="padding:40px;text-align:center">
      <p>${active.label} is not running</p>
      <p style="font-size:13px;color:var(--text3);margin-top:8px">Requires ${active.needs} to be running first. Start it from the Services page.</p>
      <button class="btn btn-primary" style="margin-top:16px" onClick=${() => startTool(active.api, active.label)} disabled=${!!starting}>${starting ? '⏳ ' + starting + '...' : '▶ Start ' + active.label}</button>
    </div>`}

    ${log && html`<div class="card" style="margin-top:16px"><div class="card-header">Output</div><pre class="log-viewer" style="max-height:200px;overflow-y:auto">${log}</pre></div>`}
  </div>`;
}
// ── Terminal ─────────────────────────────────────────────────────────────────
function TerminalPage() {
  const [projects, setProjects] = useState([]);
  const [target, setTarget] = useState('');
  const termRef = useRef(null);
  const wsRef = useRef(null);
  const termInst = useRef(null);
  const fitRef = useRef(null);

  // Read project from URL hash: #/terminal?project=shop.test
  useEffect(() => {
    GET('/api/projects').then(p => {
      setProjects(p||[]);
      const params = new URLSearchParams(location.hash.split('?')[1] || '');
      const proj = params.get('project') || '';
      if (proj) { setTarget(proj); connect(proj); }
    });
  }, []);

  const connect = useCallback((proj) => {
    // Cleanup previous
    if (wsRef.current) wsRef.current.close();
    if (termInst.current) termInst.current.dispose();

    if (!window.Terminal || !termRef.current) return;

    const term = new window.Terminal({
      theme: getTheme()==='dark' ? { background: '#0f172a', foreground: '#f1f5f9' } : { background: '#ffffff', foreground: '#1f2328' },
      fontFamily: "'Fira Code','SF Mono',monospace", fontSize: 13, cursorBlink: true,
    });
    const fitAddon = new window.FitAddon.FitAddon();
    term.loadAddon(fitAddon);
    term.open(termRef.current);
    fitAddon.fit();
    termInst.current = term;
    fitRef.current = fitAddon;

    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = proj ? `${proto}//${location.host}/api/terminal/ws?project=${encodeURIComponent(proj)}` : `${proto}//${location.host}/api/terminal/ws`;
    const ws = new WebSocket(url);
    ws.binaryType = 'arraybuffer';
    ws.onopen = () => {
      ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
      term.write(proj ? `\x1b[32mConnected to ${proj} container\x1b[0m\r\n` : '\x1b[32mConnected to host\x1b[0m\r\n');
    };
    ws.onmessage = e => {
      if (e.data instanceof ArrayBuffer) term.write(new Uint8Array(e.data));
      else term.write(e.data);
    };
    ws.onclose = () => term.write('\r\n\x1b[31m--- disconnected ---\x1b[0m\r\n');
    term.onData(data => ws.send(data));
    term.onResize(({ cols, rows }) => { if (ws.readyState === 1) ws.send(JSON.stringify({ type: 'resize', cols, rows })); });
    wsRef.current = ws;
  }, []);

  // Cleanup + resize observer
  useEffect(() => {
    if (!target) connect('');
    return () => { if (wsRef.current) wsRef.current.close(); if (termInst.current) termInst.current.dispose(); };
  }, []);

  const switchTarget = (proj) => {
    setTarget(proj);
    connect(proj);
  };

  return html`<div>
    <div class="page-header"><h1>Terminal</h1></div>
    <div style="display:flex;gap:10px;margin-bottom:12px;align-items:center">
      <select value=${target} onChange=${e => switchTarget(e.target.value)} style="min-width:200px">
        <option value="">Host (project root)</option>
        ${projects.map(p => html`<option value=${p.domain}>${p.domain} (${p.php})</option>`)}
      </select>
      <button class="btn btn-sm" onClick=${() => connect(target)}>Reconnect</button>
      <span style="font-size:12px;color:var(--text3)">${target ? 'Inside PHP container → /home/public_html/'+target : 'Host shell → project root'}</span>
    </div>
    <div class="card" style="padding:8px"><div ref=${termRef} style="height:calc(80vh - 130px)"></div></div>
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
        ${['php81','php82','php83','php84','php85'].map(p=>html`<div style="display:flex;align-items:center;gap:8px;padding:8px 12px;background:var(--bg);border-radius:var(--radius-sm)">
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
  const getPage = () => (location.hash.replace('#','').split('?')[0]) || '/';
  const [page, setPage] = useState(getPage());
  useEffect(() => { const h = () => setPage(getPage()); window.addEventListener('hashchange', h); return () => window.removeEventListener('hashchange', h); }, []);
  const nav = p => { location.hash = p; setPage(p.split('?')[0]); };
  const pages = { '/': Dashboard, '/services': ServicesPage, '/projects': Projects, '/db': DatabasePage, '/build': BuildPage, '/extensions': ExtensionsPage, '/logs': LogsPage, '/files': FilesPage, '/sql': SQLPage, '/mail': MailPage, '/search': SearchPage, '/terminal': TerminalPage, '/settings': SettingsPage };
  const Page = pages[page] || Dashboard;
  return html`<div class="app"><${Sidebar} page=${page} setPage=${nav} /><main class="main"><${Page} /></main></div>`;
}

render(html`<${App} />`, document.getElementById('app'));
