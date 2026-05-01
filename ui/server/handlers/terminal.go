package handlers

import (
	"context"
	"encoding/json"
	"io"
	"os"
	osexec "os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
	"nhooyr.io/websocket"
)

// GET /api/terminal/ws?project=shop.test — WebSocket PTY terminal
// If project is set, opens shell inside the project's PHP container
// Otherwise opens host bash in the project root
func TerminalWS(c echo.Context) error {
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	project := c.QueryParam("project")

	var cmd *osexec.Cmd

	if project != "" {
		// Read project's PHP version from projects.json
		projects, _ := readProjects()
		p, exists := projects[project]
		php := "php83"
		if exists {
			php = p.PHP
		}

		// Get compose project name for docker exec
		composeProjName := os.Getenv("COMPOSE_PROJECT_NAME")
		containerName := composeProjName + "-" + php + "-1"
		if composeProjName == "" {
			containerName = "docker-magento-multiple-php-" + php + "-1"
		}

		// Open bash inside the PHP container, cd to project dir
		cmd = osexec.CommandContext(ctx,
			"docker", "exec", "-it",
			"-e", "TERM=xterm-256color",
			"-u", "nginx",
			"-w", "/home/public_html/"+project,
			containerName,
			"bash", "--login",
		)
	} else {
		// Host shell in project root
		cmd = osexec.CommandContext(ctx, "bash", "--login")
		cmd.Dir = exec.RootDir
		cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		// Fallback: try without -it flag for docker exec
		if project != "" {
			projects, _ := readProjects()
			p, exists := projects[project]
			php := "php83"
			if exists {
				php = p.PHP
			}
			composeProjName := os.Getenv("COMPOSE_PROJECT_NAME")
			containerName := composeProjName + "-" + php + "-1"
			if composeProjName == "" {
				containerName = "docker-magento-multiple-php-" + php + "-1"
			}
			cmd = osexec.CommandContext(ctx,
				"docker", "exec", "-i",
				"-e", "TERM=xterm-256color",
				"-u", "nginx",
				"-w", "/home/public_html/"+project,
				containerName,
				"bash",
			)
			ptmx, err = pty.Start(cmd)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer ptmx.Close()

	var wg sync.WaitGroup

	// PTY → WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				conn.Write(ctx, websocket.MessageBinary, buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// WebSocket → PTY
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		for {
			typ, msg, err := conn.Read(ctx)
			if err != nil {
				break
			}
			if typ == websocket.MessageText {
				var resize struct {
					Type string `json:"type"`
					Cols int    `json:"cols"`
					Rows int    `json:"rows"`
				}
				if json.Unmarshal(msg, &resize) == nil && resize.Type == "resize" {
					pty.Setsize(ptmx, &pty.Winsize{Cols: uint16(resize.Cols), Rows: uint16(resize.Rows)})
					continue
				}
			}
			io.WriteString(ptmx, string(msg))
		}
	}()

	wg.Wait()
	cmd.Wait()
	return nil
}
