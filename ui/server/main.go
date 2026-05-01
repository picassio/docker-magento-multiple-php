package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
	"github.com/picassio/docker-magento-multiple-php/ui/server/handlers"
)

//go:embed web
var embeddedWeb embed.FS

func main() {
	port := flag.Int("port", 8888, "HTTP port")
	root := flag.String("root", "", "Project root directory")
	flag.Parse()

	// Resolve root
	if *root == "" {
		// Auto-detect: go up from binary location to find docker-compose.yml
		exe, _ := os.Executable()
		dir := filepath.Dir(exe)
		for d := dir; d != "/"; d = filepath.Dir(d) {
			if _, err := os.Stat(filepath.Join(d, "docker-compose.yml")); err == nil {
				*root = d
				break
			}
		}
		if *root == "" {
			// Try cwd
			cwd, _ := os.Getwd()
			if _, err := os.Stat(filepath.Join(cwd, "docker-compose.yml")); err == nil {
				*root = cwd
			} else {
				log.Fatal("Cannot find project root (docker-compose.yml). Use --root=<path>")
			}
		}
	}

	exec.RootDir = *root
	log.Printf("Project root: %s", exec.RootDir)

	mux := http.NewServeMux()

	// ── API routes ──────────────────────────────────────────────────────
	// Projects
	mux.HandleFunc("GET /api/projects", handlers.ListProjects)
	mux.HandleFunc("POST /api/projects", handlers.AddProject)
	mux.HandleFunc("DELETE /api/projects/{domain}", handlers.RemoveProject)
	mux.HandleFunc("PATCH /api/projects/{domain}", handlers.UpdateProject)
	mux.HandleFunc("POST /api/projects/{domain}/enable", handlers.EnableProject)
	mux.HandleFunc("POST /api/projects/{domain}/disable", handlers.DisableProject)

	// Services
	mux.HandleFunc("GET /api/services", handlers.ListServices)
	mux.HandleFunc("POST /api/services/up", handlers.ServicesUp)
	mux.HandleFunc("POST /api/services/down", handlers.ServicesDown)
	mux.HandleFunc("POST /api/services/stop", handlers.ServicesStop)
	mux.HandleFunc("POST /api/services/{name}/restart", handlers.RestartService)

	// Databases
	mux.HandleFunc("GET /api/databases", handlers.ListDatabases)
	mux.HandleFunc("POST /api/databases", handlers.CreateDatabase)
	mux.HandleFunc("DELETE /api/databases/{name}", handlers.DropDatabase)
	mux.HandleFunc("POST /api/databases/{name}/export", handlers.ExportDatabase)
	mux.HandleFunc("GET /api/databases/{name}/download", handlers.DownloadDatabase)
	mux.HandleFunc("POST /api/databases/import", handlers.ImportDatabase)

	// Images
	mux.HandleFunc("GET /api/images", handlers.ListImages)
	mux.HandleFunc("POST /api/images/build", handlers.BuildImages)
	mux.HandleFunc("GET /api/images/build/ws", handlers.BuildImagesWS)

	// Logs
	mux.HandleFunc("GET /api/logs/{service}", handlers.GetLogs)
	mux.HandleFunc("GET /api/logs/{service}/ws", handlers.StreamLogs)

	// System
	mux.HandleFunc("GET /api/doctor", handlers.Doctor)
	mux.HandleFunc("POST /api/doctor/fix", handlers.DoctorFix)
	mux.HandleFunc("GET /api/env", handlers.GetEnv)
	mux.HandleFunc("PATCH /api/env", handlers.UpdateEnv)
	mux.HandleFunc("GET /api/xdebug/{php}", handlers.XdebugStatus)
	mux.HandleFunc("POST /api/xdebug/{php}/{action}", handlers.XdebugToggle)

	// Files
	mux.HandleFunc("GET /api/files", handlers.ListFiles)
	mux.HandleFunc("GET /api/files/read", handlers.ReadFile)
	mux.HandleFunc("POST /api/files/write", handlers.WriteFile)
	mux.HandleFunc("GET /api/files/logs", handlers.ListLogs)
	mux.HandleFunc("GET /api/files/tail", handlers.TailFile)
	mux.HandleFunc("GET /api/files/tail/ws", handlers.TailFileWS)
	mux.HandleFunc("GET /api/files/download", handlers.DownloadFile)

	// DB Manager
	mux.HandleFunc("GET /api/dbmanager/tables", handlers.ListTables)
	mux.HandleFunc("GET /api/dbmanager/columns", handlers.DescribeTable)
	mux.HandleFunc("GET /api/dbmanager/data", handlers.TableData)
	mux.HandleFunc("POST /api/dbmanager/query", handlers.RunQuery)

	// Commands (shell, composer, magento, artisan, wp, ssl, varnish, install)
	mux.HandleFunc("POST /api/exec", handlers.ExecCommand)
	mux.HandleFunc("GET /api/exec/ws", handlers.ExecCommandWS)
	mux.HandleFunc("POST /api/ssl/{domain}", handlers.EnableSSL)
	mux.HandleFunc("POST /api/varnish/{domain}/{action}", handlers.VarnishToggle)
	mux.HandleFunc("POST /api/install", handlers.Install)

	// ── Static files (embedded frontend) ────────────────────────────────
	webFS, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /", http.FileServer(http.FS(webFS)))

	// ── CORS middleware (dev) ───────────────────────────────────────────
	handler := corsMiddleware(mux)

	// ── Start ───────────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Mage UI starting on http://localhost%s", addr)

	// Graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
