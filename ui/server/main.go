package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
	"github.com/picassio/docker-magento-multiple-php/ui/server/handlers"
)

//go:embed web
var embeddedWeb embed.FS

func main() {
	port := flag.Int("port", 8888, "HTTP port")
	root := flag.String("root", "", "Project root directory")
	flag.Parse()

	if *root == "" {
		exe, _ := os.Executable()
		for d := filepath.Dir(exe); d != "/"; d = filepath.Dir(d) {
			if _, err := os.Stat(filepath.Join(d, "docker-compose.yml")); err == nil {
				*root = d
				break
			}
		}
		if *root == "" {
			cwd, _ := os.Getwd()
			if _, err := os.Stat(filepath.Join(cwd, "docker-compose.yml")); err == nil {
				*root = cwd
			} else {
				log.Fatal("Cannot find project root. Use --root=<path>")
			}
		}
	}
	exec.RootDir = *root
	log.Printf("Project root: %s", exec.RootDir)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	// ── API routes ──────────────────────────────────────────────────────
	api := e.Group("/api")

	// Projects
	api.GET("/projects", handlers.ListProjects)
	api.POST("/projects", handlers.AddProject)
	api.DELETE("/projects/:domain", handlers.RemoveProject)
	api.PATCH("/projects/:domain", handlers.UpdateProject)
	api.POST("/projects/:domain/enable", handlers.EnableProject)
	api.POST("/projects/:domain/disable", handlers.DisableProject)

	// Services
	api.GET("/services", handlers.ListServices)
	api.POST("/services/up", handlers.ServicesUp)
	api.POST("/services/down", handlers.ServicesDown)
	api.POST("/services/stop", handlers.ServicesStop)
	api.POST("/services/:name/restart", handlers.RestartService)

	// Databases
	api.GET("/databases", handlers.ListDatabases)
	api.POST("/databases", handlers.CreateDatabase)
	api.DELETE("/databases/:name", handlers.DropDatabase)
	api.POST("/databases/:name/export", handlers.ExportDatabase)
	api.GET("/databases/:name/download", handlers.DownloadDatabase)
	api.POST("/databases/import", handlers.ImportDatabase)

	// Images + Build (WebSocket)
	api.GET("/images", handlers.ListImages)
	api.GET("/images/build/status", handlers.BuildStatus)
	api.POST("/images/build", handlers.BuildImages)
	api.GET("/images/build/ws", handlers.BuildImagesWS)
	api.GET("/images/build/reconnect/ws", handlers.BuildReconnectWS)

	// PHP Extensions
	api.GET("/extensions", handlers.ListAllExtensions)
	api.GET("/extensions/:service", handlers.ListExtensions)
	api.POST("/extensions/install", handlers.InstallExtension)
	api.POST("/extensions/enable", handlers.EnableExtension)
	api.POST("/extensions/disable", handlers.DisableExtension)
	api.GET("/extensions/install/ws", handlers.InstallExtensionWS)

	// Logs
	api.GET("/logs/:service", handlers.GetLogs)
	api.GET("/logs/:service/ws", handlers.StreamLogsWS)

	// Files
	api.GET("/files", handlers.ListFiles)
	api.GET("/files/read", handlers.ReadFile)
	api.POST("/files/write", handlers.WriteFile)
	api.GET("/files/logs", handlers.ListLogs)
	api.GET("/files/tail", handlers.TailFile)
	api.GET("/files/tail/ws", handlers.TailFileWS)
	api.GET("/files/download", handlers.DownloadFile)

	// DB Manager
	api.GET("/dbmanager/tables", handlers.ListTables)
	api.GET("/dbmanager/columns", handlers.DescribeTable)
	api.GET("/dbmanager/data", handlers.TableData)
	api.POST("/dbmanager/query", handlers.RunQuery)

	// System
	api.GET("/doctor", handlers.Doctor)
	api.POST("/doctor/fix", handlers.DoctorFix)
	api.GET("/env", handlers.GetEnv)
	api.PATCH("/env", handlers.UpdateEnv)
	api.GET("/xdebug/:php", handlers.XdebugStatus)
	api.POST("/xdebug/:php/:action", handlers.XdebugToggle)

	// Debug tools (phpMyAdmin + Redis Commander)
	api.POST("/debug/start", handlers.DebugStart)
	api.POST("/debug/stop", handlers.DebugStop)

	// Commands
	api.POST("/exec", handlers.ExecCommand)
	api.POST("/ssl/:domain", handlers.EnableSSL)
	api.POST("/varnish/:domain/:action", handlers.VarnishToggle)
	api.POST("/install", handlers.Install)

	// Terminal (WebSocket PTY)
	api.GET("/terminal/ws", handlers.TerminalWS)

	// Reverse proxies for embedded tools (strip CSP/X-Frame-Options for iframe)
	proxyTool := func(path, host string) {
		p := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: host})
		origDir := p.Director
		p.Director = func(r *http.Request) {
			origDir(r)
			r.Host = host
		}
		p.ModifyResponse = func(resp *http.Response) error {
			resp.Header.Del("Content-Security-Policy")
			resp.Header.Del("X-Frame-Options")
			return nil
		}
		e.Any(path+"/*", echo.WrapHandler(http.StripPrefix(path, p)))
	}
	proxyTool("/mailpit", "localhost:8025")
	proxyTool("/phpmyadmin", "localhost:8080")
	proxyTool("/redis-commander", "localhost:8081")

	// ── Static files (embedded frontend) ────────────────────────────────
	webFS, _ := fs.Sub(embeddedWeb, "web")
	e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(webFS))))

	// ── Graceful shutdown ───────────────────────────────────────────────
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Mage UI → http://localhost%s", addr)
	e.Logger.Fatal(e.Start(addr))
}
