package handlers

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
	"nhooyr.io/websocket"
)

type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

func ListFiles(c echo.Context) error {
	rel := c.QueryParam("path")
	if rel == "" { rel = "sources" }
	if strings.Contains(rel, "..") { return fail(c, 400, "invalid path") }
	abs := filepath.Join(exec.RootDir, rel)
	entries, err := os.ReadDir(abs)
	if err != nil { return fail(c, 404, "cannot read directory") }
	files := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		sz := int64(0)
		if info != nil { sz = info.Size() }
		files = append(files, FileEntry{Name: e.Name(), Path: filepath.Join(rel, e.Name()), IsDir: e.IsDir(), Size: sz})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir { return files[i].IsDir }
		return files[i].Name < files[j].Name
	})
	return ok(c, files)
}

func ReadFile(c echo.Context) error {
	rel := c.QueryParam("path")
	if rel == "" || strings.Contains(rel, "..") { return fail(c, 400, "invalid path") }
	abs := filepath.Join(exec.RootDir, rel)
	info, err := os.Stat(abs)
	if err != nil { return fail(c, 404, "not found") }
	if info.Size() > 2*1024*1024 { return fail(c, 413, "file too large (>2MB)") }
	data, _ := os.ReadFile(abs)
	return ok(c, map[string]interface{}{"path": rel, "name": filepath.Base(rel), "content": string(data), "size": info.Size()})
}

func WriteFile(c echo.Context) error {
	var req struct{ Path, Content string }
	c.Bind(&req)
	if req.Path == "" || strings.Contains(req.Path, "..") { return fail(c, 400, "invalid path") }
	abs := filepath.Join(exec.RootDir, req.Path)
	os.MkdirAll(filepath.Dir(abs), 0755)
	os.WriteFile(abs, []byte(req.Content), 0644)
	return ok(c, map[string]string{"status": "saved", "path": req.Path})
}

func ListLogs(c echo.Context) error {
	project := c.QueryParam("project")
	var paths []string
	if project != "" {
		src := filepath.Join(exec.RootDir, "sources", project)
		paths = append(paths, filepath.Join(src, "var", "log"), filepath.Join(src, "var", "report"),
			filepath.Join(src, "storage", "logs"), filepath.Join(src, "wp-content", "debug.log"))
	}
	paths = append(paths, filepath.Join(exec.RootDir, "logs", "nginx"))
	logs := make([]FileEntry, 0)
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil { continue }
		if info.IsDir() {
			entries, _ := os.ReadDir(p)
			for _, e := range entries {
				if e.IsDir() { continue }
				ei, _ := e.Info()
				sz := int64(0)
				if ei != nil { sz = ei.Size() }
				rel, _ := filepath.Rel(exec.RootDir, filepath.Join(p, e.Name()))
				logs = append(logs, FileEntry{Name: e.Name(), Path: rel, Size: sz})
			}
		} else {
			rel, _ := filepath.Rel(exec.RootDir, p)
			logs = append(logs, FileEntry{Name: info.Name(), Path: rel, Size: info.Size()})
		}
	}
	return ok(c, logs)
}

func TailFile(c echo.Context) error {
	rel := c.QueryParam("path")
	lines := c.QueryParam("lines")
	if rel == "" || strings.Contains(rel, "..") { return fail(c, 400, "invalid path") }
	if lines == "" { lines = "100" }
	res, _ := exec.Run("tail", "-n", lines, filepath.Join(exec.RootDir, rel))
	out := ""
	if res != nil { out = res.Stdout }
	return ok(c, map[string]string{"path": rel, "content": out})
}

func TailFileWS(c echo.Context) error {
	rel := c.QueryParam("path")
	if rel == "" || strings.Contains(rel, "..") { return fail(c, 400, "invalid path") }
	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil { return err }
	defer conn.Close(websocket.StatusNormalClosure, "done")
	return exec.StreamToWS(context.Background(), conn, "tail", "-f", "-n", "50", filepath.Join(exec.RootDir, rel))
}

func DownloadFile(c echo.Context) error {
	rel := c.QueryParam("path")
	if rel == "" || strings.Contains(rel, "..") { return fail(c, 400, "invalid path") }
	abs := filepath.Join(exec.RootDir, rel)
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() { return fail(c, 404, "not found") }
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(rel))
	f, _ := os.Open(abs)
	defer f.Close()
	io.Copy(c.Response(), f)
	return nil
}
