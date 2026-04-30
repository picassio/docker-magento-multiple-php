package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

// POST /api/exec — run arbitrary bin/mage command (shell, composer, magento, artisan, wp)
// Body: {"command": "composer", "args": ["mysite.com", "install"]}
func ExecCommand(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}

	allowed := map[string]bool{
		"shell": true, "composer": true, "magento": true, "artisan": true, "wp": true,
		"ssl": true, "varnish": true, "install": true, "install-laravel": true, "install-wp": true,
		"setup": true, "vhost": true,
	}
	if !allowed[req.Command] {
		jsonError(w, "command not allowed: "+req.Command, 403)
		return
	}

	args := append([]string{req.Command}, req.Args...)
	res, err := exec.MageTimeout(5*time.Minute, args...)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]interface{}{
		"command":  req.Command,
		"args":     req.Args,
		"stdout":   exec.StripAnsi(res.Stdout),
		"stderr":   exec.StripAnsi(res.Stderr),
		"exitCode": res.ExitCode,
	})
}

// GET /api/exec/ws — WebSocket: run bin/mage command with live streaming
func ExecCommandWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	_, msg, err := ws.ReadMessage()
	if err != nil {
		return
	}
	var req struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	json.Unmarshal(msg, &req)

	allowed := map[string]bool{
		"shell": true, "composer": true, "magento": true, "artisan": true, "wp": true,
		"ssl": true, "varnish": true, "install": true, "install-laravel": true, "install-wp": true,
		"setup": true, "vhost": true, "build": true,
	}
	if !allowed[req.Command] {
		ws.WriteMessage(websocket.TextMessage, []byte(`{"stream":"error","line":"command not allowed"}`))
		return
	}

	args := append([]string{req.Command}, req.Args...)
	exec.StreamToWS(ws, exec.RootDir+"/bin/mage", args...)
}

// POST /api/ssl/{domain} — enable SSL
func EnableSSL(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	res, err := exec.MageTimeout(60*time.Second, "ssl", domain)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": "ok",
		"domain": domain,
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}

// POST /api/varnish/{domain}/{action}
func VarnishToggle(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	action := r.PathValue("action")
	if action != "on" && action != "off" && action != "status" {
		jsonError(w, "action must be on, off, or status", 400)
		return
	}
	res, err := exec.MageTimeout(30*time.Second, "varnish", action, domain)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": action,
		"domain": domain,
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}

// POST /api/install — install Magento/Laravel/WordPress
// Body: {"type": "magento", "version": "2.4.8", "edition": "community", "domain": "shop.test", "php": "php83"}
//   or: {"type": "laravel", "domain": "app.test", "php": "php83"}
//   or: {"type": "wordpress", "domain": "blog.test", "php": "php83"}
func Install(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string `json:"type"`
		Version string `json:"version"`
		Edition string `json:"edition"`
		Domain  string `json:"domain"`
		PHP     string `json:"php"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}
	if req.Domain == "" {
		jsonError(w, "domain required", 400)
		return
	}

	var args []string
	switch req.Type {
	case "magento":
		if req.Version == "" || req.Edition == "" {
			jsonError(w, "version and edition required for Magento", 400)
			return
		}
		args = []string{"install", req.Version, req.Edition, req.Domain}
		if req.PHP != "" {
			args = append(args, req.PHP)
		}
	case "laravel":
		args = []string{"install-laravel", req.Domain}
		if req.PHP != "" {
			args = append(args, req.PHP)
		}
	case "wordpress":
		args = []string{"install-wp", req.Domain}
		if req.PHP != "" {
			args = append(args, req.PHP)
		}
	default:
		jsonError(w, "type must be magento, laravel, or wordpress", 400)
		return
	}

	res, err := exec.MageTimeout(10*time.Minute, args...)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]interface{}{
		"status":   "installed",
		"type":     req.Type,
		"domain":   req.Domain,
		"stdout":   exec.StripAnsi(res.Stdout),
		"stderr":   exec.StripAnsi(res.Stderr),
		"exitCode": res.ExitCode,
	})
}
