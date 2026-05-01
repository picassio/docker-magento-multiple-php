package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// GET /api/files?path=sources/mysite.com&project=mysite.com
func ListFiles(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		relPath = "sources"
	}

	// Security: prevent directory traversal
	if strings.Contains(relPath, "..") {
		jsonError(w, "invalid path", 400)
		return
	}

	absPath := filepath.Join(exec.RootDir, relPath)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		jsonError(w, "cannot read directory: "+err.Error(), 404)
		return
	}

	var files []FileEntry
	for _, e := range entries {
		info, _ := e.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		files = append(files, FileEntry{
			Name:  e.Name(),
			Path:  filepath.Join(relPath, e.Name()),
			IsDir: e.IsDir(),
			Size:  size,
		})
	}

	// Sort: dirs first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	jsonOK(w, files)
}

// GET /api/files/read?path=sources/mysite.com/app/etc/env.php
func ReadFile(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" || strings.Contains(relPath, "..") {
		jsonError(w, "invalid path", 400)
		return
	}

	absPath := filepath.Join(exec.RootDir, relPath)
	info, err := os.Stat(absPath)
	if err != nil {
		jsonError(w, "file not found", 404)
		return
	}

	// Limit file size to 2MB for viewing
	if info.Size() > 2*1024*1024 {
		jsonError(w, "file too large (>2MB)", 413)
		return
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]interface{}{
		"path":    relPath,
		"name":    filepath.Base(relPath),
		"content": string(data),
		"size":    info.Size(),
	})
}

// POST /api/files/write — save file content
func WriteFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}
	if req.Path == "" || strings.Contains(req.Path, "..") {
		jsonError(w, "invalid path", 400)
		return
	}

	absPath := filepath.Join(exec.RootDir, req.Path)

	// Ensure parent dir exists
	os.MkdirAll(filepath.Dir(absPath), 0755)

	if err := os.WriteFile(absPath, []byte(req.Content), 0644); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]string{"status": "saved", "path": req.Path})
}

// GET /api/files/logs?project=mysite.com — list log files for a project
func ListLogs(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")

	var logPaths []string

	if project != "" {
		// Project-specific logs
		srcDir := filepath.Join(exec.RootDir, "sources", project)

		// Magento logs
		logPaths = append(logPaths,
			filepath.Join(srcDir, "var", "log"),
			filepath.Join(srcDir, "var", "report"),
		)
		// Laravel logs
		logPaths = append(logPaths,
			filepath.Join(srcDir, "storage", "logs"),
		)
		// WordPress logs
		logPaths = append(logPaths,
			filepath.Join(srcDir, "wp-content", "debug.log"),
		)
	}

	// Nginx logs
	logPaths = append(logPaths, filepath.Join(exec.RootDir, "logs", "nginx"))

	var logs []FileEntry
	for _, lp := range logPaths {
		info, err := os.Stat(lp)
		if err != nil {
			continue
		}
		if info.IsDir() {
			entries, _ := os.ReadDir(lp)
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				eInfo, _ := e.Info()
				sz := int64(0)
				if eInfo != nil {
					sz = eInfo.Size()
				}
				rel, _ := filepath.Rel(exec.RootDir, filepath.Join(lp, e.Name()))
				logs = append(logs, FileEntry{
					Name: e.Name(),
					Path: rel,
					Size: sz,
				})
			}
		} else {
			rel, _ := filepath.Rel(exec.RootDir, lp)
			logs = append(logs, FileEntry{
				Name: info.Name(),
				Path: rel,
				Size: info.Size(),
			})
		}
	}

	jsonOK(w, logs)
}

// GET /api/files/tail?path=sources/shop.test/var/log/system.log&lines=100
func TailFile(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	lines := r.URL.Query().Get("lines")
	if relPath == "" || strings.Contains(relPath, "..") {
		jsonError(w, "invalid path", 400)
		return
	}
	if lines == "" {
		lines = "100"
	}

	absPath := filepath.Join(exec.RootDir, relPath)
	res, err := exec.Run("tail", "-n", lines, absPath)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]string{
		"path":    relPath,
		"lines":   lines,
		"content": res.Stdout,
	})
}

// GET /api/files/tail/ws?path=... — WebSocket live tail
func TailFileWS(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" || strings.Contains(relPath, "..") {
		http.Error(w, "invalid path", 400)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	absPath := filepath.Join(exec.RootDir, relPath)
	exec.StreamToWS(ws, "tail", "-f", "-n", "50", absPath)
}

// GET /api/files/download?path=...
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" || strings.Contains(relPath, "..") {
		jsonError(w, "invalid path", 400)
		return
	}
	absPath := filepath.Join(exec.RootDir, relPath)
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		jsonError(w, "file not found", 404)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(relPath))
	w.Header().Set("Content-Type", "application/octet-stream")
	f, _ := os.Open(absPath)
	defer f.Close()
	io.Copy(w, f)
}
