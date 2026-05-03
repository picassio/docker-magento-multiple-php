package handlers

import (
	"encoding/json"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
	"nhooyr.io/websocket"
)

type Extension struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // "module" or "zend"
}

type PHPExtensions struct {
	Service    string      `json:"service"`
	Extensions []Extension `json:"extensions"`
}

// ListExtensions returns enabled extensions for a PHP service
func ListExtensions(c echo.Context) error {
	service := c.Param("service")
	if service == "" {
		service = "php83"
	}
	res, _ := exec.DockerCompose("exec", "-T", service, "php", "-m")
	if res == nil || res.ExitCode != 0 {
		errMsg := ""
		if res != nil { errMsg = res.Stderr }
		return c.JSON(500, map[string]string{"error": "Failed to list extensions: " + errMsg})
	}

	extensions := []Extension{}
	section := "module"
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" { continue }
		if line == "[PHP Modules]" { section = "module"; continue }
		if line == "[Zend Modules]" { section = "zend"; continue }
		extensions = append(extensions, Extension{Name: line, Type: section})
	}
	return ok(c, PHPExtensions{Service: service, Extensions: extensions})
}

// ListAllExtensions returns extensions for all running PHP services
func ListAllExtensions(c echo.Context) error {
	res, _ := exec.DockerCompose("ps", "--format", "{{.Service}}", "--status", "running")
	if res == nil {
		return ok(c, []PHPExtensions{})
	}

	var results []PHPExtensions
	for _, svc := range strings.Split(res.Stdout, "\n") {
		svc = strings.TrimSpace(svc)
		if !strings.HasPrefix(svc, "php") { continue }
		if svc == "phpmyadmin" { continue }
		phpRes, _ := exec.DockerCompose("exec", "-T", svc, "php", "-m")
		if phpRes == nil || phpRes.ExitCode != 0 { continue }

		extensions := []Extension{}
		section := "module"
		for _, line := range strings.Split(phpRes.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if line == "" { continue }
			if line == "[PHP Modules]" { section = "module"; continue }
			if line == "[Zend Modules]" { section = "zend"; continue }
			extensions = append(extensions, Extension{Name: line, Type: section})
		}
		results = append(results, PHPExtensions{Service: svc, Extensions: extensions})
	}
	return ok(c, results)
}

// InstallExtension installs extensions on a PHP service
func InstallExtension(c echo.Context) error {
	var req struct {
		Service    string   `json:"service"`
		Extensions []string `json:"extensions"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}
	if len(req.Extensions) == 0 {
		return c.JSON(400, map[string]string{"error": "No extensions specified"})
	}
	if req.Service == "" {
		req.Service = "php83"
	}

	args := append([]string{"ext", "install"}, req.Extensions...)
	args = append(args, "--php="+req.Service)
	res, _ := exec.Mage(args...)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	status := "installed"
	if res != nil && res.ExitCode != 0 { status = "error" }
	return ok(c, map[string]string{"status": status, "output": out})
}

// EnableExtension enables an extension on a PHP service
func EnableExtension(c echo.Context) error {
	var req struct {
		Service   string `json:"service"`
		Extension string `json:"extension"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}
	if req.Extension == "" {
		return c.JSON(400, map[string]string{"error": "No extension specified"})
	}
	if req.Service == "" {
		req.Service = "php83"
	}

	res, _ := exec.Mage("ext", "enable", req.Extension, "--php="+req.Service)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "enabled", "output": out})
}

// DisableExtension disables an extension on a PHP service
func DisableExtension(c echo.Context) error {
	var req struct {
		Service   string `json:"service"`
		Extension string `json:"extension"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}
	if req.Extension == "" {
		return c.JSON(400, map[string]string{"error": "No extension specified"})
	}
	if req.Service == "" {
		req.Service = "php83"
	}

	res, _ := exec.Mage("ext", "disable", req.Extension, "--php="+req.Service)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "disabled", "output": out})
}

// InstallExtensionWS installs extensions via WebSocket for live output
func InstallExtensionWS(c echo.Context) error {
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil { return err }
	defer conn.Close(websocket.StatusNormalClosure, "done")

	ctx := c.Request().Context()
	_, msg, err := conn.Read(ctx)
	if err != nil { return err }

	var req struct {
		Service    string   `json:"service"`
		Extensions []string `json:"extensions"`
	}
	if err := json.Unmarshal(msg, &req); err != nil { return err }

	if req.Service == "" { req.Service = "php83" }
	args := append([]string{"ext", "install"}, req.Extensions...)
	args = append(args, "--php="+req.Service)
	return exec.StreamToWS(ctx, conn, exec.RootDir+"/bin/mage", args...)
}
