package handlers

import (
	"encoding/json"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type Container struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	Status  string `json:"status"`
	State   string `json:"state"`
	Ports   string `json:"ports"`
	Image   string `json:"image"`
}

// Services that should never appear in the UI or be controlled by it
var hiddenServices = map[string]bool{"ui": true, "mage-ui": true}

func ListServices(c echo.Context) error {
	res, _ := exec.DockerCompose("ps", "--format", "json", "-a")
	containers := make([]Container, 0)
	if res != nil {
		for _, line := range strings.Split(res.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line[0] != '{' { continue }
			var raw map[string]interface{}
			if json.Unmarshal([]byte(line), &raw) != nil { continue }
			svc := str(raw, "Service")
			if hiddenServices[svc] { continue }
			containers = append(containers, Container{
				Name: str(raw, "Name"), Service: svc,
				Status: str(raw, "Status"), State: str(raw, "State"),
				Ports: str(raw, "Ports"), Image: str(raw, "Image"),
			})
		}
	}
	return ok(c, containers)
}

func str(m map[string]interface{}, k string) string {
	if v, ok := m[k].(string); ok { return v }
	return ""
}

func ServicesUp(c echo.Context) error {
	res, _ := exec.Mage("up")
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "started", "output": out})
}

// getNonUIServices returns all running service names except the UI
func getNonUIServices() []string {
	res, _ := exec.DockerCompose("ps", "--format", "{{.Service}}", "--status", "running")
	if res == nil { return nil }
	var svcs []string
	for _, s := range strings.Split(res.Stdout, "\n") {
		s = strings.TrimSpace(s)
		if s != "" && !hiddenServices[s] {
			svcs = append(svcs, s)
		}
	}
	return svcs
}

func ServicesDown(c echo.Context) error {
	// Stop only non-UI services to avoid killing ourselves
	svcs := getNonUIServices()
	if len(svcs) == 0 {
		return ok(c, map[string]string{"status": "stopped", "output": "No services to stop"})
	}
	args := append([]string{"stop"}, svcs...)
	res, _ := exec.DockerCompose(args...)
	// Then remove the stopped containers
	rmArgs := append([]string{"rm", "-f"}, svcs...)
	exec.DockerCompose(rmArgs...)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "stopped", "output": out})
}

func ServicesStop(c echo.Context) error {
	// Stop only non-UI services
	svcs := getNonUIServices()
	if len(svcs) == 0 {
		return ok(c, map[string]string{"status": "stopped", "output": "No services to stop"})
	}
	args := append([]string{"stop"}, svcs...)
	res, _ := exec.DockerCompose(args...)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "stopped", "output": out})
}

func RestartService(c echo.Context) error {
	name := c.Param("name")
	if hiddenServices[name] {
		return fail(c, 400, "Cannot restart the UI service from the UI")
	}
	res, _ := exec.Mage("restart", name)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "restarted", "service": name, "output": out})
}
