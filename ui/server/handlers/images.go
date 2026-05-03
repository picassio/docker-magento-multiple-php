package handlers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
	"nhooyr.io/websocket"
)

type PHPImage struct {
	Version string `json:"version"`
	Image   string `json:"image"`
	Built   bool   `json:"built"`
	Size    string `json:"size"`
	ID      string `json:"id,omitempty"`
}

func ListImages(c echo.Context) error {
	// Collect PHP services from core + legacy compose
	seen := map[string]bool{}
	var versions []string
	for _, args := range [][]string{
		{"config", "--services"},
		{"-f", exec.RootDir + "/docker-compose.yml", "-f", exec.RootDir + "/compose/legacy.yml", "config", "--services"},
	} {
		res, _ := exec.Run("docker", append([]string{"compose"}, args...)...)
		if res == nil { continue }
		for _, s := range strings.Split(res.Stdout, "\n") {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, "php") && !seen[s] {
				seen[s] = true
				versions = append(versions, s)
			}
		}
	}
	images := make([]PHPImage, 0, len(versions))
	for _, v := range versions {
		img := PHPImage{Version: v, Image: "magento-" + v}
		res, _ := exec.Run("docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.ID}}", "--filter", "reference=*"+v+"*")
		if res != nil {
			for _, l := range strings.Split(res.Stdout, "\n") {
				if strings.Contains(l, v) {
					p := strings.Split(l, "\t")
					img.Built = true
					if len(p) > 0 { img.Image = p[0] }
					if len(p) > 1 { img.Size = p[1] }
					if len(p) > 2 { img.ID = p[2] }
					break
				}
			}
		}
		images = append(images, img)
	}
	return ok(c, images)
}

func BuildImages(c echo.Context) error {
	var req struct{ Versions []string `json:"versions"` }
	c.Bind(&req)
	res, _ := exec.Mage(append([]string{"build"}, req.Versions...)...)
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "built", "output": out})
}

func BuildImagesWS(c echo.Context) error {
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil { return err }
	defer conn.Close(websocket.StatusNormalClosure, "done")
	ctx := c.Request().Context()
	_, msg, err := conn.Read(ctx)
	if err != nil { return err }
	var req struct {
		Versions   []string `json:"versions"`
		Extensions string   `json:"extensions"`
	}
	json.Unmarshal(msg, &req)
	args := []string{"build"}
	if req.Extensions != "" {
		args = append(args, "--ext="+req.Extensions)
	}
	args = append(args, req.Versions...)
	return exec.StreamToWS(ctx, conn, exec.RootDir+"/bin/mage", args...)
}

func StreamLogsWS(c echo.Context) error {
	service := c.Param("service")
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil { return err }
	defer conn.Close(websocket.StatusNormalClosure, "done")
	ctx := context.Background()
	// Send initial + follow
	lines := c.QueryParam("lines")
	if lines == "" { lines = "50" }
	return exec.StreamToWS(ctx, conn, "docker", "compose", "logs", "--no-color", "--follow", "--tail="+lines, service)
}
