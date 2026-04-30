package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

// GET /api/logs/{service}
func GetLogs(w http.ResponseWriter, r *http.Request) {
	service := r.PathValue("service")
	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "100"
	}

	res, err := exec.DockerCompose("logs", "--no-color", "--tail="+lines, service)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"service": service,
		"lines":   lines,
		"output":  res.Stdout,
	})
}

// GET /api/logs/{service}/ws — WebSocket: live tail
func StreamLogs(w http.ResponseWriter, r *http.Request) {
	service := r.PathValue("service")
	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "50"
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// Send initial lines, then follow
	n, _ := strconv.Atoi(lines)
	if n > 0 {
		res, _ := exec.DockerCompose("logs", "--no-color", "--tail="+lines, service)
		if res != nil {
			msg := []byte(`{"stream":"stdout","line":"` + escapeJSON(res.Stdout) + `"}`)
			ws.WriteMessage(websocket.TextMessage, msg)
		}
	}

	// Stream follow
	exec.StreamToWS(ws, "docker", "compose", "logs", "--no-color", "--follow", "--tail=0", service)
}

func escapeJSON(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			result = append(result, '\\', '"')
		case '\\':
			result = append(result, '\\', '\\')
		case '\n':
			result = append(result, '\\', 'n')
		case '\r':
			result = append(result, '\\', 'r')
		case '\t':
			result = append(result, '\\', 't')
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}
