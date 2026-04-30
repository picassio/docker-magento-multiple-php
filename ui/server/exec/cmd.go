package exec

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Result holds command output
type Result struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

// RootDir is the project root (set from main.go)
var RootDir string

// Run executes a command and returns combined output
func Run(name string, args ...string) (*Result, error) {
	return RunCtx(context.Background(), name, args...)
}

// RunCtx executes a command with context
func RunCtx(ctx context.Context, name string, args ...string) (*Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = RootDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("exec %s: %w", name, err)
		}
	}

	return &Result{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: exitCode,
	}, nil
}

// Mage runs bin/mage with the given args
func Mage(args ...string) (*Result, error) {
	return Run(RootDir+"/bin/mage", args...)
}

// MageTimeout runs bin/mage with a timeout
func MageTimeout(timeout time.Duration, args ...string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunCtx(ctx, RootDir+"/bin/mage", args...)
}

// DockerCompose runs docker compose in the project root
func DockerCompose(args ...string) (*Result, error) {
	allArgs := append([]string{"compose"}, args...)
	return Run("docker", allArgs...)
}

// StreamToWS runs a command and streams stdout/stderr to a WebSocket
func StreamToWS(ws *websocket.Conn, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = RootDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	send := func(r io.Reader, stream string) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			msg, _ := json.Marshal(map[string]string{
				"stream": stream,
				"line":   scanner.Text(),
			})
			ws.WriteMessage(websocket.TextMessage, msg)
		}
	}

	wg.Add(2)
	go send(stdout, "stdout")
	go send(stderr, "stderr")
	wg.Wait()

	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	done, _ := json.Marshal(map[string]interface{}{
		"stream":   "done",
		"exitCode": exitCode,
	})
	ws.WriteMessage(websocket.TextMessage, done)
	return nil
}

// StripAnsi removes ANSI escape codes from a string
func StripAnsi(s string) string {
	var result strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				inEsc = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
