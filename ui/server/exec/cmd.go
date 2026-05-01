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

	"nhooyr.io/websocket"
)

type Result struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

var RootDir string

func Run(name string, args ...string) (*Result, error) {
	return RunCtx(context.Background(), name, args...)
}

func RunCtx(ctx context.Context, name string, args ...string) (*Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = RootDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			exitCode = e.ExitCode()
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

func Mage(args ...string) (*Result, error) {
	return Run(RootDir+"/bin/mage", args...)
}

func MageTimeout(timeout time.Duration, args ...string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunCtx(ctx, RootDir+"/bin/mage", args...)
}

func DockerCompose(args ...string) (*Result, error) {
	return Run("docker", append([]string{"compose"}, args...)...)
}

func StreamToWS(ctx context.Context, conn *websocket.Conn, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
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
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64*1024), 64*1024)
		for sc.Scan() {
			msg, _ := json.Marshal(map[string]string{"stream": stream, "line": sc.Text()})
			conn.Write(ctx, websocket.MessageText, msg)
		}
	}
	wg.Add(2)
	go send(stdout, "stdout")
	go send(stderr, "stderr")
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

func StripAnsi(s string) string {
	var b strings.Builder
	esc := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			esc = true
			continue
		}
		if esc {
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				esc = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
