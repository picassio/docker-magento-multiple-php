# Docker Magento — Web UI Plan

## Goal

Single-page web dashboard to manage the entire Docker Magento stack — replaces CLI for day-to-day operations.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Browser (http://localhost:9090)                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │  React SPA (Vite)                                 │  │
│  │  ┌──────┬──────┬──────┬──────┬──────┬──────────┐  │  │
│  │  │Dash  │Proj  │ DB   │Build │Logs  │Settings  │  │  │
│  │  └──────┴──────┴──────┴──────┴──────┴──────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
│                          ↕ REST API                     │
│  ┌───────────────────────────────────────────────────┐  │
│  │  API Server (Node.js / bin/mage wrapper)          │  │
│  │  - Reads projects.json                            │  │
│  │  - Shells out to bin/mage + docker compose        │  │
│  │  - WebSocket for live logs                        │  │
│  │  Port: 9090                                       │  │
│  └───────────────────────────────────────────────────┘  │
│                          ↕                              │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Docker Compose stack                              │  │
│  │  projects.json · .env · conf/                      │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**API server** is a thin wrapper — no business logic duplication. Every action calls `bin/mage` or `docker compose` under the hood.

## Tech Stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Frontend | React + Vite + Tailwind | Fast dev, tiny bundle, dark theme |
| API | Node.js (Express) | Already in PHP containers, simple process spawning |
| Real-time | WebSocket | Live logs, container status updates |
| State | `projects.json` + `docker compose ps` | No extra DB needed |

## Pages & Features

### 1. Dashboard (`/`)

Overview of the entire stack at a glance.

```
┌─────────────────────────────────────────────────────────┐
│  Docker Magento Stack                        [▶ Start]  │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Services          ┃  Projects (3)                      │
│  ● nginx     :80   ┃  ● shop.test    php83  mysql  ✓   │
│  ● php83     :9001 ┃  ● admin.test   php84  mariadb ✓  │
│  ● mysql     :3306 ┃  ○ legacy.test  php74  mysql80 ✗  │
│  ● redis     :6379 ┃                                    │
│  ● opensearch:9200 ┃  Quick Actions:                    │
│  ● mailpit   :8025 ┃  [+ Add Project] [↻ Restart]      │
│  ○ mariadb   :3308 ┃  [🔧 Doctor] [📦 Build Images]    │
│  ○ php74     :9001 ┃                                    │
│                     ┃  System Health:                    │
│  [▶ Start All]      ┃  ✓ vm.max_map_count  ✓ THP       │
│  [■ Stop All]       ┃  ✓ Docker logs       ✓ Disk 62%  │
│                     ┃                                    │
└─────────────────────────────────────────────────────────┘
```

- **Service cards**: green/red dot, name, port, image version, memory usage
- **Project list**: domain, PHP, DB, search, enabled/disabled toggle
- **System health**: from `bin/mage doctor` output
- **Quick actions**: start/stop all, add project, doctor fix

### 2. Projects (`/projects`)

Full project management CRUD.

```
┌──────────────────────────────────────────────────────────┐
│  Projects                              [+ Add Project]   │
├──────────┬──────┬──────────┬──────────┬────────┬────────┤
│ Domain   │ Type │ PHP      │ DB       │ Search │ Status │
├──────────┼──────┼──────────┼──────────┼────────┼────────┤
│shop.test │ M2   │ php83 [▾]│ mysql [▾]│ OS 2   │ ● on   │
│admin.test│ M2   │ php84 [▾]│mariadb[▾]│ ES 8   │ ● on   │
│blog.test │ WP   │ php83 [▾]│ mysql [▾]│ none   │ ○ off  │
│app.test  │ LV   │ php83 [▾]│ mysql [▾]│ none   │ ● on   │
└──────────┴──────┴──────────┴──────────┴────────┴────────┘
```

- **Inline editing**: click PHP/DB/search dropdowns to switch instantly
- **Add project modal**: domain, app type, PHP, DB, search, create vhost + DB checkboxes
- **Project detail panel**: info, vhost status, DB name, SSL status, xdebug toggle
- **Type badges**: M2 (magento2), M1 (magento1), WP (wordpress), LV (laravel), — (default)

### 3. Database (`/database`)

```
┌────────────────────────────────────────────────────────┐
│  Databases                                             │
├──────────┬──────────┬──────────┬───────┬───────────────┤
│ Name     │ Service  │ Size     │ Tables│ Actions       │
├──────────┼──────────┼──────────┼───────┼───────────────┤
│shop_test │ mysql    │ 245 MB   │ 412   │ [⬇ Export] [✕]│
│admin_test│ mariadb  │ 180 MB   │ 412   │ [⬇ Export] [✕]│
│blog_test │ mysql    │ 12 MB    │ 42    │ [⬇ Export] [✕]│
└──────────┴──────────┴──────────┴───────┴───────────────┘
│                                                        │
│  [+ Create Database]  [⬆ Import SQL]                   │
└────────────────────────────────────────────────────────┘
```

- **Create**: name + service dropdown
- **Import**: drag & drop SQL file upload → import into selected DB
- **Export**: one-click download .sql.gz
- **Drop**: with confirmation modal
- **Size info**: table count, data size

### 4. Build (`/build`)

```
┌──────────────────────────────────────────────────────────┐
│  PHP Images                                              │
├────────┬────────────┬──────────┬─────────┬───────────────┤
│ Version│ Base       │ Status   │ Size    │ Action        │
├────────┼────────────┼──────────┼─────────┼───────────────┤
│ php84  │ ubuntu:24  │ ✓ built  │ 892 MB  │ [↻ Rebuild]  │
│ php83  │ ubuntu:24  │ ✓ built  │ 878 MB  │ [↻ Rebuild]  │
│ php82  │ ubuntu:24  │ ✓ built  │ 865 MB  │ [↻ Rebuild]  │
│ php81  │ ubuntu:24  │ ✓ built  │ 851 MB  │ [↻ Rebuild]  │
│ php74  │ ubuntu:focal│ ○ not built│  —   │ [▶ Build]    │
│ php73  │ ubuntu:focal│ ○ not built│  —   │ [▶ Build]    │
└────────┴────────────┴──────────┴─────────┴───────────────┘
│  [▶ Build All]  [▶ Build Missing]                        │
│                                                          │
│  Build Log:                                              │
│  ┌──────────────────────────────────────────────────┐    │
│  │ Step 4/12: Installing PHP extensions...           │    │
│  │ Step 5/12: Installing Composer...                 │    │
│  │ ██████████████░░░░░░ 58%                         │    │
│  └──────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────┘
```

- **Image list**: version, base image, built status, size
- **Build with live log streaming** (WebSocket)
- **Rebuild**: force rebuild specific version

### 5. Logs (`/logs`)

```
┌──────────────────────────────────────────────────────────┐
│  Logs                                                    │
│  Service: [nginx ▾]  Lines: [100 ▾]  [▶ Follow]  [⬇]   │
├──────────────────────────────────────────────────────────┤
│  2026-04-30 10:23:45 shop.test GET /checkout 200 0.234s  │
│  2026-04-30 10:23:46 shop.test POST /rest/V1/.. 201 1.2s│
│  2026-04-30 10:23:47 shop.test GET /static/.. 304 0.01s │
│  ...                                                     │
└──────────────────────────────────────────────────────────┘
```

- **Service selector**: nginx, php83, mysql, redis, opensearch, etc.
- **Follow mode**: live tail via WebSocket
- **Download**: full log file
- **Search/filter**: grep through logs

### 6. Settings (`/settings`)

- **Environment**: edit `.env` values (versions, ports, credentials)
- **Doctor**: run system checks, one-click fix
- **Xdebug**: toggle per PHP version
- **SSL**: enable/disable per domain
- **Varnish**: toggle per domain

## API Endpoints

### Projects
```
GET    /api/projects                    → list all projects
POST   /api/projects                    → add project {domain, app, php, db, search}
DELETE /api/projects/:domain            → remove project
PATCH  /api/projects/:domain            → update field {field, value}
POST   /api/projects/:domain/enable     → enable project
POST   /api/projects/:domain/disable    → disable project
```

### Services
```
GET    /api/services                    → running containers + status
POST   /api/services/up                 → bin/mage up (smart start)
POST   /api/services/down               → bin/mage down
POST   /api/services/:name/restart      → restart specific service
```

### Database
```
GET    /api/databases                   → list all DBs with sizes
POST   /api/databases                   → create {name, service}
DELETE /api/databases/:name             → drop DB
POST   /api/databases/:name/export      → export → returns download URL
POST   /api/databases/import            → upload SQL file + import
```

### Build
```
GET    /api/images                      → list PHP images + build status
POST   /api/images/build                → build {versions: ["php83","php84"]}
WS     /api/images/build/stream         → live build output
```

### Logs
```
GET    /api/logs/:service?lines=100     → last N lines
WS     /api/logs/:service/stream        → live log streaming
```

### System
```
GET    /api/doctor                      → system health checks
POST   /api/doctor/fix                  → auto-fix issues
GET    /api/env                         → current .env values
PATCH  /api/env                         → update .env values
POST   /api/xdebug/:php/:action        → toggle xdebug (on/off)
GET    /api/xdebug/:php                 → xdebug status
```

## File Structure

```
ui/
├── server/
│   ├── index.js              ← Express + WebSocket server
│   ├── routes/
│   │   ├── projects.js       ← projects.json CRUD
│   │   ├── services.js       ← docker compose wrapper
│   │   ├── databases.js      ← scripts/database wrapper
│   │   ├── images.js         ← build commands
│   │   ├── logs.js           ← docker compose logs
│   │   └── system.js         ← doctor, env, xdebug
│   └── lib/
│       ├── exec.js           ← child_process wrapper with streaming
│       └── mage.js           ← bin/mage command builder
├── client/
│   ├── src/
│   │   ├── App.jsx
│   │   ├── pages/
│   │   │   ├── Dashboard.jsx
│   │   │   ├── Projects.jsx
│   │   │   ├── Database.jsx
│   │   │   ├── Build.jsx
│   │   │   ├── Logs.jsx
│   │   │   └── Settings.jsx
│   │   ├── components/
│   │   │   ├── ServiceCard.jsx
│   │   │   ├── ProjectRow.jsx
│   │   │   ├── AddProjectModal.jsx
│   │   │   ├── LogViewer.jsx
│   │   │   ├── BuildProgress.jsx
│   │   │   └── DoctorPanel.jsx
│   │   └── hooks/
│   │       ├── useWebSocket.js
│   │       └── useApi.js
│   ├── index.html
│   ├── vite.config.js
│   └── tailwind.config.js
├── package.json
└── Dockerfile              ← optional: run UI in a container
```

## Implementation Phases

### Phase 1: API Server + Dashboard (MVP)
- Express server with all API routes
- `exec.js` wrapper that shells out to `bin/mage`
- Dashboard page: service status, project list, quick actions
- WebSocket for live container status

### Phase 2: Project Management
- Full CRUD UI for projects
- Inline PHP/DB/search switching with dropdowns
- Add project modal with app type selection
- Vhost auto-creation on add

### Phase 3: Database + Build
- Database page with import/export
- SQL file drag & drop upload
- Build page with live log streaming
- Image status and size display

### Phase 4: Logs + Settings
- Log viewer with service selector + follow mode
- Settings page: .env editor, doctor, xdebug toggles
- SSL management per domain

## Key Principles

1. **No logic duplication** — API server calls `bin/mage`, never reimplements
2. **Real-time** — WebSocket for logs, build output, container status
3. **Dark theme** — matches terminal aesthetic
4. **Mobile-friendly** — responsive layout for tablet/phone
5. **Zero config** — `bin/mage ui` starts everything
6. **Offline-first** — works without internet (local Docker stack)
