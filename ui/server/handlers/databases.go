package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type Database struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	Size    string `json:"size,omitempty"`
	Tables  int    `json:"tables,omitempty"`
}

func listDBsForService(service string) ([]Database, error) {
	mysqlPwdEnv := "MYSQL_PWD=root"
	res, err := exec.DockerCompose("exec", "-T", "-e", mysqlPwdEnv, service, "mysql", "-u", "root", "-N", "-e",
		"SELECT table_schema, ROUND(SUM(data_length+index_length)/1024/1024,1), COUNT(*) FROM information_schema.tables WHERE table_schema NOT IN ('information_schema','mysql','performance_schema','sys') GROUP BY table_schema")
	if err != nil || res == nil || res.ExitCode != 0 || strings.TrimSpace(res.Stdout) == "" {
		// Fallback: just list names
		res2, _ := exec.DockerCompose("exec", "-T", "-e", mysqlPwdEnv, service, "mysql", "-u", "root", "-N", "-e", "SHOW DATABASES")
		if res2 == nil {
			return nil, fmt.Errorf("cannot connect to %s", service)
		}
		var dbs []Database
		for _, line := range strings.Split(res2.Stdout, "\n") {
			name := strings.TrimSpace(line)
			if name == "" || name == "information_schema" || name == "mysql" || name == "performance_schema" || name == "sys" {
				continue
			}
			dbs = append(dbs, Database{Name: name, Service: service})
		}
		return dbs, nil
	}

	var dbs []Database
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		tables := 0
		fmt.Sscanf(fields[2], "%d", &tables)
		dbs = append(dbs, Database{
			Name:    fields[0],
			Service: service,
			Size:    fields[1] + " MB",
			Tables:  tables,
		})
	}
	return dbs, nil
}

// GET /api/databases
func ListDatabases(w http.ResponseWriter, r *http.Request) {
	services := []string{"mysql", "mysql80", "mariadb"}
	all := make([]Database, 0)

	for _, svc := range services {
		// Check if running
		res, _ := exec.DockerCompose("ps", "--services", "--filter", "status=running")
		if res == nil || !strings.Contains(res.Stdout, svc) {
			continue
		}
		dbs, err := listDBsForService(svc)
		if err != nil {
			continue
		}
		all = append(all, dbs...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	jsonOK(w, all)
}

// POST /api/databases
func CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Service string `json:"service"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}
	if req.Name == "" {
		jsonError(w, "name required", 400)
		return
	}
	if req.Service == "" {
		req.Service = "mysql"
	}

	res, err := exec.Run(exec.RootDir+"/scripts/database", "create",
		"--database-name="+req.Name, "--db-service="+req.Service)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	if res.ExitCode != 0 {
		jsonError(w, exec.StripAnsi(res.Stderr+"\n"+res.Stdout), 400)
		return
	}
	jsonOK(w, map[string]string{"status": "created", "name": req.Name, "service": req.Service})
}

// DELETE /api/databases/{name}
func DropDatabase(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	service := r.URL.Query().Get("service")
	if service == "" {
		service = "mysql"
	}

	mysqlPwdEnv := "MYSQL_PWD=root"
	res, err := exec.DockerCompose("exec", "-T", "-e", mysqlPwdEnv, service, "mysql", "-u", "root", "-e",
		"DROP DATABASE IF EXISTS `"+name+"`")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	if res.ExitCode != 0 {
		jsonError(w, exec.StripAnsi(res.Stderr), 400)
		return
	}
	jsonOK(w, map[string]string{"status": "dropped", "name": name})
}

// POST /api/databases/{name}/export
func ExportDatabase(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	service := r.URL.Query().Get("service")
	if service == "" {
		service = "mysql"
	}

	res, err := exec.Run(exec.RootDir+"/scripts/database", "export",
		"--database-name="+name, "--db-service="+service)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	// Find the exported file
	exportDir := filepath.Join(exec.RootDir, "databases", "export")
	entries, _ := filepath.Glob(filepath.Join(exportDir, name+"-*.sql"))
	if len(entries) == 0 {
		jsonOK(w, map[string]string{"status": "exported", "output": exec.StripAnsi(res.Stdout)})
		return
	}
	sort.Strings(entries)
	latest := entries[len(entries)-1]
	jsonOK(w, map[string]string{
		"status":   "exported",
		"file":     filepath.Base(latest),
		"download": "/api/databases/" + name + "/download?file=" + filepath.Base(latest),
	})
}

// GET /api/databases/{name}/download
func DownloadDatabase(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" || strings.Contains(file, "..") {
		jsonError(w, "invalid file", 400)
		return
	}
	path := filepath.Join(exec.RootDir, "databases", "export", file)
	if _, err := os.Stat(path); err != nil {
		jsonError(w, "file not found", 404)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+file)
	w.Header().Set("Content-Type", "application/sql")
	f, _ := os.Open(path)
	defer f.Close()
	io.Copy(w, f)
}

// POST /api/databases/import
func ImportDatabase(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(500 << 20) // 500MB max
	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "file required", 400)
		return
	}
	defer file.Close()

	target := r.FormValue("target")
	service := r.FormValue("service")
	if target == "" {
		jsonError(w, "target database required", 400)
		return
	}
	if service == "" {
		service = "mysql"
	}

	// Save uploaded file
	importDir := filepath.Join(exec.RootDir, "databases", "import")
	os.MkdirAll(importDir, 0755)
	destPath := filepath.Join(importDir, header.Filename)
	dest, err := os.Create(destPath)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	io.Copy(dest, file)
	dest.Close()

	// Import
	res, err := exec.Run(exec.RootDir+"/scripts/database", "import",
		"--source="+header.Filename, "--target="+target, "--db-service="+service)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	// Cleanup
	os.Remove(destPath)

	jsonOK(w, map[string]string{
		"status": "imported",
		"target": target,
		"output": exec.StripAnsi(res.Stdout),
	})
}
