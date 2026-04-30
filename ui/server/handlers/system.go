package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

// GET /api/doctor
func Doctor(w http.ResponseWriter, r *http.Request) {
	res, err := exec.Run(exec.RootDir+"/scripts/doctor", "check")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	output := exec.StripAnsi(res.Stdout + "\n" + res.Stderr)

	// Parse into structured checks
	var checks []map[string]string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		check := map[string]string{"raw": line}
		if strings.Contains(line, "✔") || strings.Contains(line, "[OK]") || strings.Contains(line, "PASS") {
			check["status"] = "pass"
		} else if strings.Contains(line, "✖") || strings.Contains(line, "[FAIL]") || strings.Contains(line, "WARNING") {
			check["status"] = "fail"
		} else {
			check["status"] = "info"
		}
		checks = append(checks, check)
	}

	jsonOK(w, map[string]interface{}{
		"output": output,
		"checks": checks,
	})
}

// POST /api/doctor/fix
func DoctorFix(w http.ResponseWriter, r *http.Request) {
	res, err := exec.Run(exec.RootDir+"/scripts/doctor", "fix")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": "fixed",
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}

// GET /api/env
func GetEnv(w http.ResponseWriter, r *http.Request) {
	envFile := filepath.Join(exec.RootDir, ".env")
	f, err := os.Open(envFile)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	defer f.Close()

	var entries []map[string]string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			entries = append(entries, map[string]string{"type": "comment", "value": line})
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			entries = append(entries, map[string]string{
				"type":  "var",
				"key":   parts[0],
				"value": parts[1],
			})
		}
	}
	jsonOK(w, entries)
}

// PATCH /api/env
func UpdateEnv(w http.ResponseWriter, r *http.Request) {
	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}

	envFile := filepath.Join(exec.RootDir, ".env")
	data, err := os.ReadFile(envFile)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			if newVal, ok := updates[parts[0]]; ok {
				lines[i] = parts[0] + "=" + newVal
				delete(updates, parts[0])
			}
		}
	}

	// Append any new keys
	for k, v := range updates {
		lines = append(lines, k+"="+v)
	}

	if err := os.WriteFile(envFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"status": "updated"})
}

// GET /api/xdebug/{php}
func XdebugStatus(w http.ResponseWriter, r *http.Request) {
	php := r.PathValue("php")
	res, _ := exec.Mage("xdebug", "status", php)
	output := ""
	if res != nil {
		output = exec.StripAnsi(res.Stdout)
	}
	enabled := strings.Contains(strings.ToLower(output), "enabled") || strings.Contains(strings.ToLower(output), "active")
	jsonOK(w, map[string]interface{}{
		"php":     php,
		"enabled": enabled,
		"output":  output,
	})
}

// POST /api/xdebug/{php}/{action}
func XdebugToggle(w http.ResponseWriter, r *http.Request) {
	php := r.PathValue("php")
	action := r.PathValue("action")
	if action != "on" && action != "off" {
		jsonError(w, "action must be on or off", 400)
		return
	}
	res, err := exec.Mage("xdebug", action, php)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{
		"status": action,
		"php":    php,
		"output": exec.StripAnsi(res.Stdout + "\n" + res.Stderr),
	})
}
