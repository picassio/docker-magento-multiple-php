package handlers

import (
	"encoding/json"
	"sort"
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
	sort.SliceStable(containers, func(i, j int) bool {
		oi, oj := 100, 100
		if v, ok := svcOrder[containers[i].Service]; ok { oi = v }
		if v, ok := svcOrder[containers[j].Service]; ok { oj = v }
		return oi < oj
	})
	return ok(c, containers)
}

func str(m map[string]interface{}, k string) string {
	if v, ok := m[k].(string); ok { return v }
	return ""
}

// ListAllServices returns all defined services (running or not) from compose config
func ListAllServices(c echo.Context) error {
	// Get all defined services
	res, _ := exec.DockerCompose("config", "--services")
	if res == nil {
		return ok(c, []map[string]string{})
	}

	// Get running services
	runningRes, _ := exec.DockerCompose("ps", "--format", "{{.Service}}\t{{.State}}\t{{.Status}}\t{{.Ports}}", "-a")
	running := map[string]map[string]string{}
	if runningRes != nil {
		for _, line := range strings.Split(runningRes.Stdout, "\n") {
			parts := strings.SplitN(strings.TrimSpace(line), "\t", 4)
			if len(parts) >= 2 {
				running[parts[0]] = map[string]string{"state": parts[1]}
				if len(parts) >= 3 { running[parts[0]]["status"] = parts[2] }
				if len(parts) >= 4 { running[parts[0]]["ports"] = parts[3] }
			}
		}
	}

	var services []map[string]string
	for _, svc := range strings.Split(res.Stdout, "\n") {
		svc = strings.TrimSpace(svc)
		if svc == "" || hiddenServices[svc] { continue }
		entry := map[string]string{"service": svc, "state": "stopped", "status": "", "ports": ""}
		if r, ok := running[svc]; ok {
			entry["state"] = r["state"]
			entry["status"] = r["status"]
			entry["ports"] = r["ports"]
		}
		services = append(services, entry)
	}
	sort.SliceStable(services, func(i, j int) bool {
		oi, oj := 100, 100
		if v, ok := svcOrder[services[i]["service"]]; ok { oi = v }
		if v, ok := svcOrder[services[j]["service"]]; ok { oj = v }
		return oi < oj
	})
	return ok(c, services)
}

// svcOrder defines stable display order for services
var svcOrder = map[string]int{
		"nginx": 1,
		"php70": 10, "php71": 11, "php72": 12, "php73": 13, "php74": 14,
		"php81": 15, "php82": 16, "php83": 17, "php84": 18, "php85": 19,
		"mysql": 30, "mysql80": 31, "mariadb": 32,
		"redis": 40, "redis6": 41,
		"opensearch": 50, "opensearch1": 51,
		"elasticsearch": 52, "elasticsearch7": 53,
		"rabbitmq": 60,
		"mailpit": 70, "varnish": 80,
}

func ServicesUp(c echo.Context) error {
	res, _ := exec.DockerCompose("up", "-d", "--no-build")
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	status := "started"
	if res != nil && res.ExitCode != 0 { status = "error" }
	return ok(c, map[string]string{"status": status, "output": out})
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

func StartService(c echo.Context) error {
	name := c.Param("name")
	if hiddenServices[name] {
		return fail(c, 400, "Cannot control the UI service from the UI")
	}
	// Use --no-build to prevent hanging on missing images
	res, _ := exec.DockerCompose("up", "-d", "--no-build", name)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	status := "started"
	if res != nil && res.ExitCode != 0 {
		status = "error"
		if strings.Contains(out, "no such image") || strings.Contains(out, "No such image") || strings.Contains(out, "pull access denied") {
			out = "Image for " + name + " not built yet. Go to Build page to build it first.\n\n" + out
		}
	}
	return ok(c, map[string]string{"status": status, "service": name, "output": out})
}

func StopService(c echo.Context) error {
	name := c.Param("name")
	if hiddenServices[name] {
		return fail(c, 400, "Cannot control the UI service from the UI")
	}
	res, _ := exec.DockerCompose("stop", name)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "stopped", "service": name, "output": out})
}

func RestartService(c echo.Context) error {
	name := c.Param("name")
	if hiddenServices[name] {
		return fail(c, 400, "Cannot control the UI service from the UI")
	}
	res, _ := exec.Mage("restart", name)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "restarted", "service": name, "output": out})
}
