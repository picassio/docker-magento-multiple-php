package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type Project struct {
	Domain    string `json:"domain"`
	PHP       string `json:"php"`
	App       string `json:"app"`
	DBService string `json:"db_service"`
	DBName    string `json:"db_name"`
	Search    string `json:"search"`
	Enabled   bool   `json:"enabled"`
	Status    string `json:"status"`
}

func projectsFile() string { return filepath.Join(exec.RootDir, "projects.json") }

func readProjects() (map[string]Project, error) {
	data, err := os.ReadFile(projectsFile())
	if err != nil {
		if os.IsNotExist(err) { return map[string]Project{}, nil }
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil { return nil, err }
	projects := make(map[string]Project, len(raw))
	for domain, v := range raw {
		var p Project
		json.Unmarshal(v, &p)
		p.Domain = domain
		projects[domain] = p
	}
	return projects, nil
}

func writeProjects(projects map[string]Project) error {
	out := make(map[string]interface{}, len(projects))
	for d, p := range projects {
		out[d] = map[string]interface{}{
			"php": p.PHP, "app": p.App, "db_service": p.DBService,
			"db_name": p.DBName, "search": p.Search, "enabled": p.Enabled,
		}
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	return os.WriteFile(projectsFile(), data, 0644)
}

func ListProjects(c echo.Context) error {
	projects, err := readProjects()
	if err != nil { return fail(c, 500, err.Error()) }

	// Auto-discover projects from sources/ if projects.json is empty
	if len(projects) == 0 {
		srcDir := filepath.Join(exec.RootDir, "sources")
		entries, _ := os.ReadDir(srcDir)
		for _, e := range entries {
			if !e.IsDir() || e.Name() == ".keepme" { continue }
			domain := e.Name()
			app := detectAppType(filepath.Join(srcDir, domain))
			dbName := strings.ReplaceAll(strings.ReplaceAll(domain, ".", "_"), "-", "_")
			p := Project{
				Domain: domain, PHP: "php83", App: app,
				DBService: "mysql", DBName: dbName,
				Search: "none", Enabled: true,
			}
			if app == "magento2" { p.Search = "opensearch" }
			projects[domain] = p
		}
		if len(projects) > 0 {
			writeProjects(projects)
		}
	}

	// Get running services to compute project status
	running := getRunningServices()

	list := make([]Project, 0, len(projects))
	for _, p := range projects {
		p.Status = computeProjectStatus(p, running)
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Domain < list[j].Domain
	})
	return ok(c, list)
}

// detectAppType checks source dir for framework markers
func detectAppType(dir string) string {
	// Magento 2: app/etc/env.php or bin/magento
	if _, err := os.Stat(filepath.Join(dir, "bin", "magento")); err == nil { return "magento2" }
	if _, err := os.Stat(filepath.Join(dir, "app", "etc", "env.php")); err == nil { return "magento2" }
	// Laravel: artisan file
	if _, err := os.Stat(filepath.Join(dir, "artisan")); err == nil { return "laravel" }
	// WordPress: wp-config.php or wp-includes/
	if _, err := os.Stat(filepath.Join(dir, "wp-config.php")); err == nil { return "wordpress" }
	if _, err := os.Stat(filepath.Join(dir, "wp-includes")); err == nil { return "wordpress" }
	// Magento 1: app/Mage.php
	if _, err := os.Stat(filepath.Join(dir, "app", "Mage.php")); err == nil { return "magento1" }
	return "default"
}

func AddProject(c echo.Context) error {
	var p Project
	if err := c.Bind(&p); err != nil { return fail(c, 400, "invalid JSON") }
	if p.Domain == "" { return fail(c, 400, "domain required") }
	if p.PHP == "" { p.PHP = "php83" }
	if p.App == "" { p.App = "magento2" }
	if p.DBService == "" { p.DBService = "mysql" }
	if p.DBName == "" { p.DBName = strings.ReplaceAll(strings.ReplaceAll(p.Domain, ".", "_"), "-", "_") }
	if p.Search == "" { p.Search = "opensearch" }
	p.Enabled = true
	projects, err := readProjects()
	if err != nil { return fail(c, 500, err.Error()) }
	if _, exists := projects[p.Domain]; exists { return fail(c, 409, "project already exists") }
	projects[p.Domain] = p
	writeProjects(projects)
	os.MkdirAll(filepath.Join(exec.RootDir, "sources", p.Domain), 0755)
	rootDir := p.Domain
	if p.App == "laravel" { rootDir = p.Domain + "/public" }
	exec.Run(exec.RootDir+"/scripts/create-vhost", "--domain="+p.Domain, "--app="+p.App, "--root-dir="+rootDir, "--php-version="+p.PHP)
	exec.Run(exec.RootDir+"/scripts/database", "create", "--database-name="+p.DBName, "--db-service="+p.DBService)
	return ok(c, p)
}

func RemoveProject(c echo.Context) error {
	domain := c.Param("domain")
	projects, err := readProjects()
	if err != nil { return fail(c, 500, err.Error()) }
	if _, exists := projects[domain]; !exists { return fail(c, 404, "project not found") }
	os.Remove(filepath.Join(exec.RootDir, "conf", "nginx", "conf.d", domain+".conf"))
	delete(projects, domain)
	writeProjects(projects)
	return ok(c, map[string]string{"status": "removed"})
}

func UpdateProject(c echo.Context) error {
	domain := c.Param("domain")
	projects, err := readProjects()
	if err != nil { return fail(c, 500, err.Error()) }
	p, exists := projects[domain]
	if !exists { return fail(c, 404, "project not found") }
	var u map[string]interface{}
	c.Bind(&u)
	if v, ok := u["php"].(string); ok {
		p.PHP = v
		rootDir := p.Domain
		if p.App == "laravel" { rootDir = p.Domain + "/public" }
		exec.Run(exec.RootDir+"/scripts/create-vhost", "--domain="+p.Domain, "--app="+p.App, "--root-dir="+rootDir, "--php-version="+p.PHP)
	}
	if v, ok := u["app"].(string); ok { p.App = v }
	if v, ok := u["db_service"].(string); ok { p.DBService = v }
	if v, ok := u["db_name"].(string); ok { p.DBName = v }
	if v, ok := u["search"].(string); ok { p.Search = v }
	projects[domain] = p
	writeProjects(projects)
	return ok(c, p)
}

func EnableProject(c echo.Context) error  { return setEnabled(c, true) }
func DisableProject(c echo.Context) error { return setEnabled(c, false) }

// getRunningServices returns a set of currently running service names
func getRunningServices() map[string]bool {
	res, _ := exec.DockerCompose("ps", "--format", "{{.Service}}", "--status", "running")
	running := map[string]bool{}
	if res != nil {
		for _, s := range strings.Split(res.Stdout, "\n") {
			s = strings.TrimSpace(s)
			if s != "" { running[s] = true }
		}
	}
	return running
}

// computeProjectStatus returns "live", "partial", "stopped", or "disabled"
func computeProjectStatus(p Project, running map[string]bool) string {
	if !p.Enabled { return "disabled" }
	needed := projectServices(p)
	up := 0
	for _, s := range needed {
		if running[s] { up++ }
	}
	if up == len(needed) { return "live" }
	if up > 0 { return "partial" }
	return "stopped"
}

// projectServices returns docker compose services needed by a project
func projectServices(p Project) []string {
	svcs := []string{"nginx", p.PHP, p.DBService, "mailpit", "redis"}
	if p.Search != "" && p.Search != "none" {
		svcs = append(svcs, p.Search)
	}
	return svcs
}

// overridesForProject returns compose override file names needed
func overridesForProject(p Project) []string {
	var ov []string
	if len(p.PHP) >= 5 && p.PHP[:4] == "php7" {
		ov = append(ov, "legacy")
	}
	switch p.DBService {
	case "mariadb":
		ov = append(ov, "mariadb")
	case "mysql80":
		ov = append(ov, "mysql80")
	}
	switch p.Search {
	case "elasticsearch":
		ov = append(ov, "elasticsearch")
	case "elasticsearch7":
		ov = append(ov, "elasticsearch7")
	case "opensearch1":
		ov = append(ov, "opensearch1")
	}
	return ov
}

// buildProjectComposeArgs returns the docker compose args for a project
// Uses --project-directory with HOST path so volume mounts match existing
// containers (prevents unnecessary recreates). Uses container paths for -f
// since files must be readable from where the command runs.
func buildProjectComposeArgs(p Project) []string {
	hostDir := hostProjectDir()
	args := []string{"compose", "--project-directory", hostDir, "-f", exec.RootDir + "/docker-compose.yml"}
	for _, ov := range overridesForProject(p) {
		args = append(args, "-f", exec.RootDir+"/compose/"+ov+".yml")
	}
	return args
}

func StartProject(c echo.Context) error {
	domain := c.Param("domain")
	projects, err := readProjects()
	if err != nil { return fail(c, 500, err.Error()) }
	p, exists := projects[domain]
	if !exists { return fail(c, 404, "project not found") }

	if !p.Enabled {
		p.Enabled = true
		projects[domain] = p
		writeProjects(projects)
	}

	args := buildProjectComposeArgs(p)
	args = append(args, "up", "-d", "--no-build")
	args = append(args, projectServices(p)...)

	res, _ := exec.Run("docker", args...)
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	status := "started"
	if res != nil && res.ExitCode != 0 {
		status = "error"
		if strings.Contains(out, "no such image") || strings.Contains(out, "No such image") || strings.Contains(out, "pull access denied") {
			out = "Image not built yet. Go to Build page to build it first.\n\n" + out
		}
	}
	return ok(c, map[string]string{"status": status, "output": out})
}


func StopProject(c echo.Context) error {
	domain := c.Param("domain")
	projects, err := readProjects()
	if err != nil { return fail(c, 500, err.Error()) }
	p, exists := projects[domain]
	if !exists { return fail(c, 404, "project not found") }

	// Find services used ONLY by this project (not shared with others)
	myServices := projectServices(p)
	shared := map[string]bool{"nginx": true, "mailpit": true}
	for _, op := range projects {
		if op.Domain == domain || !op.Enabled { continue }
		for _, s := range projectServices(op) {
			shared[s] = true
		}
	}

	var toStop []string
	for _, s := range myServices {
		if !shared[s] {
			toStop = append(toStop, s)
		}
	}

	out := ""
	if len(toStop) > 0 {
		args := append([]string{"stop"}, toStop...)
		res, _ := exec.DockerCompose(args...)
		if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	} else {
		out = "All services shared with other projects — nothing to stop"
	}

	return ok(c, map[string]string{"status": "stopped", "output": out, "stopped": strings.Join(toStop, ", ")})
}

func setEnabled(c echo.Context, enabled bool) error {
	domain := c.Param("domain")
	projects, err := readProjects()
	if err != nil { return fail(c, 500, err.Error()) }
	p, exists := projects[domain]
	if !exists { return fail(c, 404, "project not found") }
	p.Enabled = enabled
	projects[domain] = p
	writeProjects(projects)
	return ok(c, p)
}
