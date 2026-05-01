package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type Database struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	Size    string `json:"size,omitempty"`
	Tables  int    `json:"tables,omitempty"`
}

func listDBs(service string) []Database {
	env := "MYSQL_PWD=root"
	res, _ := exec.DockerCompose("exec", "-T", "-e", env, service, "mysql", "-u", "root", "-N", "-e",
		"SELECT table_schema, ROUND(SUM(data_length+index_length)/1024/1024,1), COUNT(*) FROM information_schema.tables WHERE table_schema NOT IN ('information_schema','mysql','performance_schema','sys') GROUP BY table_schema")
	if res == nil || strings.TrimSpace(res.Stdout) == "" {
		res2, _ := exec.DockerCompose("exec", "-T", "-e", env, service, "mysql", "-u", "root", "-N", "-e", "SHOW DATABASES")
		if res2 == nil { return nil }
		var dbs []Database
		for _, l := range strings.Split(res2.Stdout, "\n") {
			n := strings.TrimSpace(l)
			if n == "" || n == "information_schema" || n == "mysql" || n == "performance_schema" || n == "sys" { continue }
			dbs = append(dbs, Database{Name: n, Service: service})
		}
		return dbs
	}
	var dbs []Database
	for _, l := range strings.Split(res.Stdout, "\n") {
		f := strings.Fields(strings.TrimSpace(l))
		if len(f) < 3 { continue }
		t := 0; fmt.Sscanf(f[2], "%d", &t)
		dbs = append(dbs, Database{Name: f[0], Service: service, Size: f[1] + " MB", Tables: t})
	}
	return dbs
}

func ListDatabases(c echo.Context) error {
	all := make([]Database, 0)
	for _, svc := range []string{"mysql", "mysql80", "mariadb"} {
		res, _ := exec.DockerCompose("ps", "--services", "--filter", "status=running")
		if res == nil || !strings.Contains(res.Stdout, svc) { continue }
		if dbs := listDBs(svc); dbs != nil { all = append(all, dbs...) }
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return ok(c, all)
}

func CreateDatabase(c echo.Context) error {
	var req struct{ Name, Service string }
	c.Bind(&req)
	if req.Name == "" { return fail(c, 400, "name required") }
	if req.Service == "" { req.Service = "mysql" }
	res, _ := exec.Run(exec.RootDir+"/scripts/database", "create", "--database-name="+req.Name, "--db-service="+req.Service)
	if res != nil && res.ExitCode != 0 { return fail(c, 400, exec.StripAnsi(res.Stderr+"\n"+res.Stdout)) }
	return ok(c, map[string]string{"status": "created", "name": req.Name})
}

func DropDatabase(c echo.Context) error {
	name := c.Param("name")
	svc := c.QueryParam("service")
	if svc == "" { svc = "mysql" }
	exec.DockerCompose("exec", "-T", "-e", "MYSQL_PWD=root", svc, "mysql", "-u", "root", "-e", "DROP DATABASE IF EXISTS `"+name+"`")
	return ok(c, map[string]string{"status": "dropped", "name": name})
}

func ExportDatabase(c echo.Context) error {
	name := c.Param("name")
	svc := c.QueryParam("service")
	if svc == "" { svc = "mysql" }
	exec.Run(exec.RootDir+"/scripts/database", "export", "--database-name="+name, "--db-service="+svc)
	dir := filepath.Join(exec.RootDir, "databases", "export")
	entries, _ := filepath.Glob(filepath.Join(dir, name+"-*.sql"))
	if len(entries) == 0 { return ok(c, map[string]string{"status": "exported"}) }
	sort.Strings(entries)
	f := filepath.Base(entries[len(entries)-1])
	return ok(c, map[string]string{"status": "exported", "file": f, "download": "/api/databases/" + name + "/download?file=" + f})
}

func DownloadDatabase(c echo.Context) error {
	file := c.QueryParam("file")
	if file == "" || strings.Contains(file, "..") { return fail(c, 400, "invalid file") }
	path := filepath.Join(exec.RootDir, "databases", "export", file)
	if _, err := os.Stat(path); err != nil { return fail(c, 404, "not found") }
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+file)
	f, _ := os.Open(path)
	defer f.Close()
	io.Copy(c.Response(), f)
	return nil
}

func ImportDatabase(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil { return fail(c, 400, "file required") }
	target := c.FormValue("target")
	svc := c.FormValue("service")
	if target == "" { return fail(c, 400, "target required") }
	if svc == "" { svc = "mysql" }
	dir := filepath.Join(exec.RootDir, "databases", "import")
	os.MkdirAll(dir, 0755)
	src, _ := file.Open()
	defer src.Close()
	dst, _ := os.Create(filepath.Join(dir, file.Filename))
	io.Copy(dst, src)
	dst.Close()
	res, _ := exec.Run(exec.RootDir+"/scripts/database", "import", "--source="+file.Filename, "--target="+target, "--db-service="+svc)
	os.Remove(filepath.Join(dir, file.Filename))
	out := ""
	if res != nil { out = exec.StripAnsi(res.Stdout) }
	return c.JSON(http.StatusOK, map[string]string{"status": "imported", "output": out})
}
