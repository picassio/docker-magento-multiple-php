package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	Redis     string `json:"redis,omitempty"`
}

func projectsFile() string {
	return filepath.Join(exec.RootDir, "projects.json")
}

func readProjects() (map[string]Project, error) {
	data, err := os.ReadFile(projectsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]Project{}, nil
		}
		return nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	projects := make(map[string]Project, len(raw))
	for domain, v := range raw {
		var p Project
		if err := json.Unmarshal(v, &p); err != nil {
			continue
		}
		p.Domain = domain
		projects[domain] = p
	}
	return projects, nil
}

func writeProjects(projects map[string]Project) error {
	out := make(map[string]interface{}, len(projects))
	for domain, p := range projects {
		out[domain] = map[string]interface{}{
			"php":        p.PHP,
			"app":        p.App,
			"db_service": p.DBService,
			"db_name":    p.DBName,
			"search":     p.Search,
			"enabled":    p.Enabled,
		}
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(projectsFile(), data, 0644)
}

// GET /api/projects
func ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := readProjects()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	list := make([]Project, 0, len(projects))
	for _, p := range projects {
		list = append(list, p)
	}
	jsonOK(w, list)
}

// POST /api/projects
func AddProject(w http.ResponseWriter, r *http.Request) {
	var p Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}
	if p.Domain == "" {
		jsonError(w, "domain required", 400)
		return
	}
	if p.PHP == "" {
		p.PHP = "php83"
	}
	if p.App == "" {
		p.App = "magento2"
	}
	if p.DBService == "" {
		p.DBService = "mysql"
	}
	if p.DBName == "" {
		p.DBName = strings.ReplaceAll(strings.ReplaceAll(p.Domain, ".", "_"), "-", "_")
	}
	if p.Search == "" {
		p.Search = "opensearch"
	}
	p.Enabled = true

	projects, err := readProjects()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	if _, exists := projects[p.Domain]; exists {
		jsonError(w, "project already exists", 409)
		return
	}

	projects[p.Domain] = p
	if err := writeProjects(projects); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	// Create source dir
	srcDir := filepath.Join(exec.RootDir, "sources", p.Domain)
	os.MkdirAll(srcDir, 0755)

	// Create vhost
	rootDir := p.Domain
	if p.App == "laravel" {
		rootDir = p.Domain + "/public"
	}
	exec.Run(exec.RootDir+"/scripts/create-vhost",
		"--domain="+p.Domain, "--app="+p.App,
		"--root-dir="+rootDir, "--php-version="+p.PHP)

	// Create database
	exec.Run(exec.RootDir+"/scripts/database", "create",
		"--database-name="+p.DBName, "--db-service="+p.DBService)

	jsonOK(w, p)
}

// DELETE /api/projects/{domain}
func RemoveProject(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	projects, err := readProjects()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	if _, exists := projects[domain]; !exists {
		jsonError(w, "project not found", 404)
		return
	}

	// Remove vhost
	vhost := filepath.Join(exec.RootDir, "conf", "nginx", "conf.d", domain+".conf")
	os.Remove(vhost)

	delete(projects, domain)
	if err := writeProjects(projects); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]string{"status": "removed"})
}

// PATCH /api/projects/{domain}
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	projects, err := readProjects()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	p, exists := projects[domain]
	if !exists {
		jsonError(w, "project not found", 404)
		return
	}

	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}

	if v, ok := update["php"].(string); ok {
		p.PHP = v
		// Recreate vhost with new PHP
		rootDir := p.Domain
		if p.App == "laravel" {
			rootDir = p.Domain + "/public"
		}
		exec.Run(exec.RootDir+"/scripts/create-vhost",
			"--domain="+p.Domain, "--app="+p.App,
			"--root-dir="+rootDir, "--php-version="+p.PHP)
	}
	if v, ok := update["app"].(string); ok {
		p.App = v
	}
	if v, ok := update["db_service"].(string); ok {
		p.DBService = v
	}
	if v, ok := update["db_name"].(string); ok {
		p.DBName = v
	}
	if v, ok := update["search"].(string); ok {
		p.Search = v
	}

	projects[domain] = p
	if err := writeProjects(projects); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, p)
}

// POST /api/projects/{domain}/enable
func EnableProject(w http.ResponseWriter, r *http.Request) {
	setProjectEnabled(w, r, true)
}

// POST /api/projects/{domain}/disable
func DisableProject(w http.ResponseWriter, r *http.Request) {
	setProjectEnabled(w, r, false)
}

func setProjectEnabled(w http.ResponseWriter, r *http.Request, enabled bool) {
	domain := r.PathValue("domain")
	projects, err := readProjects()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	p, exists := projects[domain]
	if !exists {
		jsonError(w, "project not found", 404)
		return
	}
	p.Enabled = enabled
	projects[domain] = p
	if err := writeProjects(projects); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, p)
}
