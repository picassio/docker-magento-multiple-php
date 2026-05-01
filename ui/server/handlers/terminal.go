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

// GET /api/terminal/ws — WebSocket PTY terminal
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

	// Start bash shell in the project root
	cmd := osexec.CommandContext(ctx, "bash", "--login")
	cmd.Dir = exec.RootDir
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"PS1=\\[\\033[1;32m\\]mage-ui\\[\\033[0m\\]:\\[\\033[1;34m\\]\\w\\[\\033[0m\\]$ ",
	)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	var wg sync.WaitGroup

	// PTY → WebSocket (output)
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

	// WebSocket → PTY (input)
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
				// Check for resize messages: {"type":"resize","cols":80,"rows":24}
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
			// Regular input
			io.WriteString(ptmx, string(msg))
		}
	}()

	wg.Wait()
	cmd.Wait()
	return nil
}
