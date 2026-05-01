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

func ListServices(c echo.Context) error {
	res, _ := exec.DockerCompose("ps", "--format", "json", "-a")
	containers := make([]Container, 0)
	if res != nil {
		for _, line := range strings.Split(res.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line[0] != '{' { continue }
			var raw map[string]interface{}
			if json.Unmarshal([]byte(line), &raw) != nil { continue }
			containers = append(containers, Container{
				Name: str(raw, "Name"), Service: str(raw, "Service"),
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

func ServicesDown(c echo.Context) error {
	res, _ := exec.Mage("down")
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "stopped", "output": out})
}

func ServicesStop(c echo.Context) error {
	res, _ := exec.Mage("stop")
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "stopped", "output": out})
}

func RestartService(c echo.Context) error {
	name := c.Param("name")
	res, _ := exec.Mage("restart", name)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "restarted", "service": name, "output": out})
}
