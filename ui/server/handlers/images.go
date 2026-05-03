package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	appexec "github.com/picassio/docker-magento-multiple-php/ui/server/exec"
	"nhooyr.io/websocket"
)

type PHPImage struct {
	Version string `json:"version"`
	Image   string `json:"image"`
	Built   bool   `json:"built"`
	Size    string `json:"size"`
	ID      string `json:"id,omitempty"`
}

// ── Build state tracking ─────────────────────────────────────────────────────
type buildState struct {
	mu       sync.Mutex
	active   bool
	target   string
	started  time.Time
	log      []string
	maxLines int
	subs     map[chan string]struct{}
}

var build = &buildState{maxLines: 5000, subs: make(map[chan string]struct{})}

func (b *buildState) start(target string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.active = true
	b.target = target
	b.started = time.Now()
	b.log = nil
}

func (b *buildState) appendLine(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.log) < b.maxLines {
		b.log = append(b.log, line)
	}
	for ch := range b.subs {
		select {
		case ch <- line:
		default:
		}
	}
}

func (b *buildState) finish() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.active = false
	for ch := range b.subs {
		close(ch)
		delete(b.subs, ch)
	}
}

func (b *buildState) subscribe() ([]string, chan string, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	history := make([]string, len(b.log))
	copy(history, b.log)
	ch := make(chan string, 100)
	b.subs[ch] = struct{}{}
	unsub := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.subs, ch)
	}
	return history, ch, unsub
}

func (b *buildState) status() (bool, string, time.Time, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active, b.target, b.started, len(b.log)
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func ListImages(c echo.Context) error {
	seen := map[string]bool{}
	var versions []string
	for _, args := range [][]string{
		{"config", "--services"},
		{"-f", appexec.RootDir + "/docker-compose.yml", "-f", appexec.RootDir + "/compose/legacy.yml", "config", "--services"},
	} {
		res, _ := appexec.Run("docker", append([]string{"compose"}, args...)...)
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
		res, _ := appexec.Run("docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.ID}}", "--filter", "reference=*"+v+"*")
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
	// Sort by PHP version
	phpOrder := map[string]int{
		"php70": 0, "php71": 1, "php72": 2, "php73": 3, "php74": 4,
		"php81": 5, "php82": 6, "php83": 7, "php84": 8, "php85": 9,
	}
	sort.SliceStable(images, func(i, j int) bool {
		oi, oj := 99, 99
		if v, ok := phpOrder[images[i].Version]; ok { oi = v }
		if v, ok := phpOrder[images[j].Version]; ok { oj = v }
		return oi < oj
	})
	return ok(c, images)
}

// BuildStatus returns whether a build is currently running
func BuildStatus(c echo.Context) error {
	active, target, started, lines := build.status()
	return ok(c, map[string]interface{}{
		"active":  active,
		"target":  target,
		"started": started,
		"lines":   lines,
	})
}

func BuildImages(c echo.Context) error {
	var req struct{ Versions []string `json:"versions"` }
	c.Bind(&req)
	res, _ := appexec.Mage(append([]string{"build"}, req.Versions...)...)
	out := ""
	if res != nil { out = appexec.StripAnsi(res.Stdout + "\n" + res.Stderr) }
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

	target := strings.Join(req.Versions, ", ")
	build.start(target)
	defer build.finish()

	// Run the build and stream to both the WebSocket and the build log
	cmd := exec.CommandContext(ctx, appexec.RootDir+"/bin/mage", args...)
	cmd.Dir = appexec.RootDir
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	send := func(r *bufio.Scanner, stream string) {
		defer wg.Done()
		for r.Scan() {
			line := r.Text()
			build.appendLine(line)
			msg, _ := json.Marshal(map[string]string{"stream": stream, "line": line})
			conn.Write(ctx, websocket.MessageText, msg)
		}
	}
	wg.Add(2)
	go send(bufio.NewScanner(stdout), "stdout")
	go send(bufio.NewScanner(stderr), "stderr")
	wg.Wait()

	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			exitCode = e.ExitCode()
		}
	}
	done, _ := json.Marshal(map[string]interface{}{"stream": "done", "exitCode": exitCode})
	conn.Write(ctx, websocket.MessageText, done)
	return nil
}

// BuildReconnectWS allows reconnecting to an in-progress build
func BuildReconnectWS(c echo.Context) error {
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil { return err }
	defer conn.Close(websocket.StatusNormalClosure, "done")
	ctx := c.Request().Context()

	active, _, _, _ := build.status()
	if !active {
		msg, _ := json.Marshal(map[string]string{"stream": "done", "line": "No active build"})
		conn.Write(ctx, websocket.MessageText, msg)
		return nil
	}

	// Send history then subscribe to new lines
	history, ch, unsub := build.subscribe()
	defer unsub()

	for _, line := range history {
		msg, _ := json.Marshal(map[string]string{"stream": "stdout", "line": line})
		conn.Write(ctx, websocket.MessageText, msg)
	}

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				// Build finished
				msg, _ := json.Marshal(map[string]string{"stream": "done", "line": ""})
				conn.Write(ctx, websocket.MessageText, msg)
				return nil
			}
			msg, _ := json.Marshal(map[string]string{"stream": "stdout", "line": line})
			conn.Write(ctx, websocket.MessageText, msg)
		case <-ctx.Done():
			return nil
		}
	}
}

func StreamLogsWS(c echo.Context) error {
	service := c.Param("service")
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil { return err }
	defer conn.Close(websocket.StatusNormalClosure, "done")
	ctx := context.Background()
	lines := c.QueryParam("lines")
	if lines == "" { lines = "50" }
	return appexec.StreamToWS(ctx, conn, "docker", "compose", "logs", "--no-color", "--follow", "--tail="+lines, service)
}
