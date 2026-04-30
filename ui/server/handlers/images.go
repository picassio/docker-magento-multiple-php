package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type PHPImage struct {
	Version string `json:"version"`
	Image   string `json:"image"`
	Built   bool   `json:"built"`
	Size    string `json:"size"`
	ID      string `json:"id,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// GET /api/images
func ListImages(w http.ResponseWriter, r *http.Request) {
	// Get all PHP versions from compose
	res, _ := exec.DockerCompose("config", "--services")
	var phpVersions []string
	if res != nil {
		for _, svc := range strings.Split(res.Stdout, "\n") {
			svc = strings.TrimSpace(svc)
			if strings.HasPrefix(svc, "php") {
				phpVersions = append(phpVersions, svc)
			}
		}
	}

	// Also check legacy
	res2, _ := exec.Run("docker", "compose", "-f", exec.RootDir+"/docker-compose.yml",
		"-f", exec.RootDir+"/compose/legacy.yml", "config", "--services")
	if res2 != nil {
		for _, svc := range strings.Split(res2.Stdout, "\n") {
			svc = strings.TrimSpace(svc)
			if strings.HasPrefix(svc, "php") {
				found := false
				for _, v := range phpVersions {
					if v == svc {
						found = true
						break
					}
				}
				if !found {
					phpVersions = append(phpVersions, svc)
				}
			}
		}
	}

	// Check which are built
	var images []PHPImage
	for _, ver := range phpVersions {
		img := PHPImage{Version: ver, Image: "magento-" + ver}
		// Check if image exists
		res, _ := exec.Run("docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.ID}}",
			"--filter", "reference=*"+ver+"*")
		if res != nil && res.Stdout != "" {
			for _, line := range strings.Split(res.Stdout, "\n") {
				if strings.Contains(line, ver) {
					parts := strings.Split(line, "\t")
					img.Built = true
					if len(parts) > 0 {
						img.Image = parts[0]
					}
					if len(parts) > 1 {
						img.Size = parts[1]
					}
					if len(parts) > 2 {
						img.ID = parts[2]
					}
					break
				}
			}
		}
		images = append(images, img)
	}

	jsonOK(w, images)
}

// POST /api/images/build
func BuildImages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Versions []string `json:"versions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}

	args := append([]string{"build"}, req.Versions...)
	res, err := exec.Mage(args...)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": "built",
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}

// GET /api/images/build/ws — WebSocket: live build output
func BuildImagesWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// Read versions from first message
	_, msg, err := ws.ReadMessage()
	if err != nil {
		return
	}
	var req struct {
		Versions []string `json:"versions"`
	}
	json.Unmarshal(msg, &req)

	args := append([]string{"build"}, req.Versions...)
	exec.StreamToWS(ws, exec.RootDir+"/bin/mage", args...)
}
