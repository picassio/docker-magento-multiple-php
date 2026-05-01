package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

func GetLogs(c echo.Context) error {
	service := c.Param("service")
	lines := c.QueryParam("lines")
	if lines == "" { lines = "100" }
	res, _ := exec.DockerCompose("logs", "--no-color", "--tail="+lines, service)
	out := ""
	if res != nil { out = res.Stdout }
	return ok(c, map[string]string{"service": service, "lines": lines, "output": out})
}
