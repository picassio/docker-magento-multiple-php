package handlers

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

func Doctor(c echo.Context) error {
	res, _ := exec.Run(exec.RootDir+"/scripts/doctor", "check")
	output := ""
	if res != nil { output = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	var checks []map[string]string
	for _, l := range strings.Split(output, "\n") {
		l = strings.TrimSpace(l)
		if l == "" { continue }
		ch := map[string]string{"raw": l, "status": "info"}
		if strings.Contains(l, "✔") || strings.Contains(l, "[OK]") { ch["status"] = "pass" }
		if strings.Contains(l, "✖") || strings.Contains(l, "[FAIL]") || strings.Contains(l, "WARNING") { ch["status"] = "fail" }
		checks = append(checks, ch)
	}
	return ok(c, map[string]interface{}{"output": output, "checks": checks})
}

func DoctorFix(c echo.Context) error {
	res, _ := exec.Run(exec.RootDir+"/scripts/doctor", "fix")
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "fixed", "output": out})
}

func GetEnv(c echo.Context) error {
	f, err := os.Open(filepath.Join(exec.RootDir, ".env"))
	if err != nil { return fail(c, 500, err.Error()) }
	defer f.Close()
	var entries []map[string]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			entries = append(entries, map[string]string{"type": "comment", "value": line})
		} else if parts := strings.SplitN(t, "=", 2); len(parts) == 2 {
			entries = append(entries, map[string]string{"type": "var", "key": parts[0], "value": parts[1]})
		}
	}
	return ok(c, entries)
}

func UpdateEnv(c echo.Context) error {
	var updates map[string]string
	c.Bind(&updates)
	envFile := filepath.Join(exec.RootDir, ".env")
	data, _ := os.ReadFile(envFile)
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "#") { continue }
		if parts := strings.SplitN(t, "=", 2); len(parts) == 2 {
			if v, ok := updates[parts[0]]; ok {
				lines[i] = parts[0] + "=" + v
				delete(updates, parts[0])
			}
		}
	}
	for k, v := range updates { lines = append(lines, k+"="+v) }
	os.WriteFile(envFile, []byte(strings.Join(lines, "\n")), 0644)
	return ok(c, map[string]string{"status": "updated"})
}

func XdebugStatus(c echo.Context) error {
	php := c.Param("php")
	res, _ := exec.Mage("xdebug", "status", php)
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout) }
	enabled := strings.Contains(strings.ToLower(out), "enabled")
	return ok(c, map[string]interface{}{"php": php, "enabled": enabled, "output": out})
}

func XdebugToggle(c echo.Context) error {
	php, action := c.Param("php"), c.Param("action")
	if action != "on" && action != "off" { return fail(c, 400, "action must be on or off") }
	res, _ := exec.Mage("xdebug", action, php)
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": action, "php": php, "output": out})
}

func ExecCommand(c echo.Context) error {
	var req struct{ Command string; Args []string }
	c.Bind(&req)
	allowed := map[string]bool{"shell":true,"composer":true,"magento":true,"artisan":true,"wp":true,"ssl":true,"varnish":true,"install":true,"install-laravel":true,"install-wp":true,"setup":true,"vhost":true}
	if !allowed[req.Command] { return fail(c, 403, "command not allowed") }
	res, _ := exec.MageTimeout(5*time.Minute, append([]string{req.Command}, req.Args...)...)
	out, stderr := "", ""
	exitCode := 0
	if res != nil { out = exec.StripNoise(res.Stdout); stderr = exec.StripNoise(res.Stderr); exitCode = res.ExitCode }
	return ok(c, map[string]interface{}{"command": req.Command, "stdout": out, "stderr": stderr, "exitCode": exitCode})
}


// POST /api/debug/start — start phpMyAdmin + Redis Commander
// Uses --project-directory with host path so volume mounts resolve on the host,
// but -f flags use container paths (where files are accessible).
func DebugStart(c echo.Context) error {
	hostDir := exec.HostProjectDir()
	res, _ := exec.Run("docker", "compose",
		"--project-directory", hostDir,
		"-f", exec.RootDir+"/docker-compose.yml",
		"-f", exec.RootDir+"/compose/debug.yml",
		"up", "-d", "phpmyadmin", "redis-commander")
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "started", "output": out})
}

// POST /api/debug/stop
func DebugStop(c echo.Context) error {
	hostDir := exec.HostProjectDir()
	exec.Run("docker", "compose",
		"--project-directory", hostDir,
		"-f", exec.RootDir+"/docker-compose.yml",
		"-f", exec.RootDir+"/compose/debug.yml",
		"rm", "-sf", "phpmyadmin", "redis-commander")
	return ok(c, map[string]string{"status": "stopped"})
}

// POST /api/dashboards/start — start OpenSearch Dashboards
func DashboardsStart(c echo.Context) error {
	hostDir := exec.HostProjectDir()
	res, _ := exec.Run("docker", "compose",
		"--project-directory", hostDir,
		"-f", exec.RootDir+"/docker-compose.yml",
		"-f", exec.RootDir+"/compose/dashboards.yml",
		"up", "-d", "opensearch-dashboards")
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "started", "output": out})
}

// POST /api/dashboards/stop
func DashboardsStop(c echo.Context) error {
	hostDir := exec.HostProjectDir()
	exec.Run("docker", "compose",
		"--project-directory", hostDir,
		"-f", exec.RootDir+"/docker-compose.yml",
		"-f", exec.RootDir+"/compose/dashboards.yml",
		"rm", "-sf", "opensearch-dashboards")
	return ok(c, map[string]string{"status": "stopped"})
}

func EnableSSL(c echo.Context) error {
	domain := c.Param("domain")
	res, _ := exec.MageTimeout(60*time.Second, "ssl", domain)
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": "ok", "domain": domain, "output": out})
}

func VarnishToggle(c echo.Context) error {
	domain, action := c.Param("domain"), c.Param("action")
	res, _ := exec.MageTimeout(30*time.Second, "varnish", action, domain)
	out := ""
	if res != nil { out = exec.StripNoise(res.Stdout + "\n" + res.Stderr) }
	return ok(c, map[string]string{"status": action, "domain": domain, "output": out})
}

func Install(c echo.Context) error {
	var req struct{ Type, Version, Edition, Domain, PHP string }
	c.Bind(&req)
	if req.Domain == "" { return fail(c, 400, "domain required") }
	var args []string
	switch req.Type {
	case "magento":
		if req.Version == "" || req.Edition == "" { return fail(c, 400, "version and edition required") }
		args = []string{"install", req.Version, req.Edition, req.Domain}
		if req.PHP != "" { args = append(args, req.PHP) }
	case "laravel":
		args = []string{"install-laravel", req.Domain}
		if req.PHP != "" { args = append(args, req.PHP) }
	case "wordpress":
		args = []string{"install-wp", req.Domain}
		if req.PHP != "" { args = append(args, req.PHP) }
	default:
		return fail(c, 400, "type must be magento, laravel, or wordpress")
	}
	res, _ := exec.MageTimeout(10*time.Minute, args...)
	out, stderr := "", ""
	exitCode := 0
	if res != nil { out = exec.StripNoise(res.Stdout); stderr = exec.StripNoise(res.Stderr); exitCode = res.ExitCode }
	return ok(c, map[string]interface{}{"status": "installed", "type": req.Type, "domain": req.Domain, "stdout": out, "stderr": stderr, "exitCode": exitCode})
}
