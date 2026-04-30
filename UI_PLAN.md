# Docker Magento — Web UI Plan

## Goal

Single-page web dashboard to manage the entire Docker Magento stack — replaces CLI for day-to-day operations.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Browser (http://localhost:8888)                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Vanilla HTML + JS + CSS (embedded in Go binary)  │  │
│  │  Hash routing: #/ #/projects #/db #/build #/logs  │  │
│  │  Light + dark theme toggle                        │  │
│  └───────────────────────────────────────────────────┘  │
│                  ↕ REST + WebSocket                     │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Go binary (mage-ui)                              │  │
│  │  - Shells out to bin/mage + docker compose        │  │
│  │  - Reads/writes projects.json, .env               │  │
│  │  - WebSocket for logs + build streaming            │  │
│  │  - Frontend embedded via go:embed                  │  │
│  │  - Port 8888                                       │  │
│  └───────────────────────────────────────────────────┘  │
│                          ↕                              │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Docker Compose stack                              │  │
│  │  projects.json · .env · conf/                      │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Decisions (from grill session)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Audience | Solo dev, localhost, no auth | Dev tool, same trust as phpMyAdmin |
| Frontend | Vanilla HTML + JS + CSS | Zero deps, zero build step, like vm-provisioner |
| API Server | Go binary | Single file, fast, native WebSocket, cross-platform |
| Embedding | `go:embed` — HTML/JS/CSS compiled into binary | One binary, zero file deps |
| Distribution | Auto-download on `bin/mage ui` from GitHub releases | No Go required for users |
| Runs on | Host process (`bin/mage ui`) | Direct access to projects.json, scripts, Docker |
| Real-time | Polling (3s) for status, WebSocket for logs/builds | Simple where possible, streaming where needed |
| Routing | Hash routing (`#/projects`) | Shareable URLs, zero-dep, vanilla JS |
| Theme | Light + dark toggle | Accessible |
| Terminal | None | GUI is the point; users have their own terminal |
| Scope | All features, no phasing | Ship complete |
| Build order | Server-first — all Go endpoints, then HTML | Solid API foundation |
| Port | 8888 | Matches vm-provisioner |

## Pages

### 1. Dashboard (`#/`)

Stack overview at a glance.

- **Service cards**: name, status (green/red dot), port, image version, memory
- **Project list**: domain, type badge (M2/WP/LV), PHP, DB, search, enabled toggle
- **System health**: from `bin/mage doctor` — sysctl, THP, Docker logs, disk
- **Quick actions**: Start All, Stop All, Add Project, Doctor Fix

### 2. Projects (`#/projects`)

Full project CRUD with inline editing.

- **Table**: domain, type, PHP (dropdown), DB (dropdown), search (dropdown), status toggle
- **Add project modal**: domain, app type, PHP, DB, search, auto-create vhost + DB
- **Project detail panel**: vhost status, DB name, SSL, source dir path
- **Type badges**: M2 (magento2), M1 (magento1), WP (wordpress), LV (laravel), — (default)
- **Inline switching**: click PHP/DB/search → dropdown → instant switch via API

### 3. Database (`#/db`)

- **Table**: name, service, size, table count, actions (export/drop)
- **Create**: name + service dropdown
- **Import**: file upload → import into selected DB
- **Export**: one-click download .sql.gz
- **Drop**: confirmation modal

### 4. Build (`#/build`)

- **Image list**: PHP version, base image, built status, size
- **Build actions**: Build All, Build Missing, Rebuild specific
- **Live progress**: WebSocket stream of docker build output
- **Progress bar**: parsed from build output

### 5. Logs (`#/logs`)

- **Service selector**: dropdown of running containers
- **Follow mode**: live tail via WebSocket
- **Line count**: 100/500/1000
- **Search**: filter log lines
- **Download**: full log file

### 6. Settings (`#/settings`)

- **.env editor**: key/value table with save
- **Doctor panel**: check results + one-click fix
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
POST   /api/projects/:domain/enable     → enable project + create vhost
POST   /api/projects/:domain/disable    → disable project + remove vhost
```

### Services
```
GET    /api/services                    → container list + status + ports + memory
POST   /api/services/up                 → bin/mage up (smart start)
POST   /api/services/down               → bin/mage down
POST   /api/services/stop               → bin/mage stop
POST   /api/services/:name/restart      → restart specific service
```

### Database
```
GET    /api/databases                   → list all DBs with sizes
POST   /api/databases                   → create {name, service}
DELETE /api/databases/:name             → drop DB
POST   /api/databases/:name/export      → export → returns download link
POST   /api/databases/import            → multipart upload SQL + import
```

### Build
```
GET    /api/images                      → list PHP images + build status + size
POST   /api/images/build                → build {versions: [...]}
GET    /api/images/build/ws             → WebSocket: live build output
```

### Logs
```
GET    /api/logs/:service?lines=100     → last N lines
GET    /api/logs/:service/ws            → WebSocket: live tail
```

### System
```
GET    /api/doctor                      → system health checks (JSON)
POST   /api/doctor/fix                  → auto-fix issues
GET    /api/env                         → current .env key/values
PATCH  /api/env                         → update .env values
GET    /api/xdebug/:php                 → xdebug status
POST   /api/xdebug/:php/:action        → toggle (on/off)
```

## File Structure

```
ui/
├── server/
│   ├── main.go               ← entry point, HTTP server, router
│   ├── handlers/
│   │   ├── projects.go       ← projects.json CRUD
│   │   ├── services.go       ← docker compose wrapper
│   │   ├── databases.go      ← scripts/database wrapper
│   │   ├── images.go         ← build commands + WebSocket stream
│   │   ├── logs.go           ← docker compose logs + WebSocket
│   │   └── system.go         ← doctor, env, xdebug
│   ├── exec/
│   │   ├── cmd.go            ← child_process wrapper (run bin/mage)
│   │   └── stream.go         ← streaming exec → WebSocket
│   ├── go.mod
│   └── go.sum
├── web/
│   ├── index.html            ← single HTML file, all pages
│   ├── app.js                ← hash router + API calls + DOM updates
│   ├── style.css             ← dark/light theme, responsive
│   └── assets/               ← favicon, icons (optional)
└── Makefile                  ← build targets: go build, cross-compile
```

## Go Server Design

### Thin wrapper — no business logic

Every handler calls `bin/mage` or `docker compose` via `exec.Cmd()`:

```go
// handlers/projects.go
func ListProjects(w http.ResponseWriter, r *http.Request) {
    out, _ := exec.Run("bin/mage", "project", "list", "--json")
    w.Header().Set("Content-Type", "application/json")
    w.Write(out)
}

func SwitchPHP(w http.ResponseWriter, r *http.Request) {
    domain := chi.URLParam(r, "domain")
    php := r.FormValue("php")
    exec.Run("bin/mage", "project", "switch-php", domain, php)
}
```

### WebSocket streaming

```go
// exec/stream.go
func StreamCmd(ws *websocket.Conn, name string, args ...string) {
    cmd := exec.Command(name, args...)
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()
    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        ws.WriteMessage(websocket.TextMessage, scanner.Bytes())
    }
    cmd.Wait()
}
```

### Frontend embedding

```go
//go:embed web/*
var webFS embed.FS

func main() {
    // Serve embedded frontend
    http.Handle("/", http.FileServer(http.FS(webFS)))
    // API routes
    http.HandleFunc("/api/projects", handlers.ListProjects)
    // ...
}
```

## Distribution

1. **GitHub Actions** builds Go binary for linux-amd64, linux-arm64, darwin-amd64, darwin-arm64
2. **GitHub Releases** hosts the binaries
3. **`bin/mage ui`** checks for `ui/mage-ui` binary:
   - If exists → run it
   - If not → download latest from GitHub releases → run it
4. Binary opens `http://localhost:8888` in default browser

## bin/mage ui command

```bash
cmd_ui() {
    local ui_bin="${MAGE_ROOT}/ui/mage-ui"
    local release_url="https://github.com/picassio/docker-magento-multiple-php/releases/latest/download"

    if [[ ! -x "$ui_bin" ]]; then
        _arrow "Downloading UI server..."
        local os=$(uname -s | tr '[:upper:]' '[:lower:]')
        local arch=$(uname -m); [[ "$arch" == "x86_64" ]] && arch="amd64"
        mkdir -p "$(dirname "$ui_bin")"
        curl -fsSL "${release_url}/mage-ui-${os}-${arch}" -o "$ui_bin"
        chmod +x "$ui_bin"
        _success "UI server downloaded"
    fi

    _arrow "Starting UI at http://localhost:8888"
    "$ui_bin" --root="$MAGE_ROOT" --port=8888 &
    local pid=$!

    # Open browser
    if command -v xdg-open &>/dev/null; then xdg-open "http://localhost:8888"
    elif command -v open &>/dev/null; then open "http://localhost:8888"
    fi

    wait $pid
}
```

## Build Order

1. **Go server** — all API endpoints, test with curl
2. **HTML/CSS** — layout, theme toggle, navigation
3. **JS** — hash router, API client, DOM rendering
4. **WebSocket** — logs + build streaming
5. **Polish** — error handling, loading states, mobile responsive
6. **CI** — GitHub Actions cross-compile + release
