package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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

// GET /api/services
func ListServices(w http.ResponseWriter, r *http.Request) {
	res, err := exec.DockerCompose("ps", "--format", "json", "-a")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	containers := make([]Container, 0)
	// docker compose ps --format json outputs one JSON object per line
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		c := Container{
			Name:    getString(raw, "Name"),
			Service: getString(raw, "Service"),
			Status:  getString(raw, "Status"),
			State:   getString(raw, "State"),
			Ports:   getString(raw, "Ports"),
			Image:   getString(raw, "Image"),
		}
		containers = append(containers, c)
	}

	jsonOK(w, containers)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// POST /api/services/up
func ServicesUp(w http.ResponseWriter, r *http.Request) {
	res, err := exec.Mage("up")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": "started",
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}

// POST /api/services/down
func ServicesDown(w http.ResponseWriter, r *http.Request) {
	res, err := exec.Mage("down")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": "stopped",
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}

// POST /api/services/stop
func ServicesStop(w http.ResponseWriter, r *http.Request) {
	res, err := exec.Mage("stop")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": "stopped",
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}

// POST /api/services/{name}/restart
func RestartService(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	res, err := exec.Mage("restart", name)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status":  "restarted",
		"service": name,
		"output":  exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}
