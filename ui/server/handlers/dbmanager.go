package handlers

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type TableInfo struct{ Name, Engine, Rows, Size string }
type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
	Error   string          `json:"error,omitempty"`
}

func dbq(svc, db, q string) (*exec.Result, error) {
	return exec.DockerCompose("exec", "-T", "-e", "MYSQL_PWD=root", svc, "mysql", "-u", "root", "-N", db, "-e", q)
}

func ListTables(c echo.Context) error {
	db, svc := c.QueryParam("db"), c.QueryParam("service")
	if db == "" || svc == "" { return fail(c, 400, "db and service required") }
	res, _ := dbq(svc, db, "SELECT table_name, engine, table_rows, ROUND((data_length+index_length)/1024/1024,2) FROM information_schema.tables WHERE table_schema='"+db+"' ORDER BY table_name")
	tables := make([]TableInfo, 0)
	if res != nil {
		for _, l := range strings.Split(res.Stdout, "\n") {
			f := strings.Split(strings.TrimSpace(l), "\t")
			if len(f) < 1 || f[0] == "" { continue }
			t := TableInfo{Name: f[0]}
			if len(f) > 1 { t.Engine = f[1] }
			if len(f) > 2 { t.Rows = f[2] }
			if len(f) > 3 { t.Size = f[3] + " MB" }
			tables = append(tables, t)
		}
	}
	return ok(c, tables)
}

func DescribeTable(c echo.Context) error {
	db, svc, table := c.QueryParam("db"), c.QueryParam("service"), c.QueryParam("table")
	if db == "" || svc == "" || table == "" { return fail(c, 400, "db, service, table required") }
	if strings.ContainsAny(table, "';\"\\") { return fail(c, 400, "invalid table name") }
	res, _ := dbq(svc, db, "DESCRIBE `"+table+"`")
	cols := make([]map[string]string, 0)
	if res != nil {
		for _, l := range strings.Split(res.Stdout, "\n") {
			f := strings.Split(strings.TrimSpace(l), "\t")
			if len(f) < 1 || f[0] == "" { continue }
			col := map[string]string{"field": f[0]}
			if len(f) > 1 { col["type"] = f[1] }
			if len(f) > 2 { col["null"] = f[2] }
			if len(f) > 3 { col["key"] = f[3] }
			cols = append(cols, col)
		}
	}
	return ok(c, cols)
}

func TableData(c echo.Context) error {
	db, svc, table := c.QueryParam("db"), c.QueryParam("service"), c.QueryParam("table")
	if db == "" || svc == "" || table == "" { return fail(c, 400, "db, service, table required") }
	if strings.ContainsAny(table, "';\"\\") { return fail(c, 400, "invalid table name") }
	page, limit := 1, 50
	fmt.Sscanf(c.QueryParam("page"), "%d", &page)
	fmt.Sscanf(c.QueryParam("limit"), "%d", &limit)
	if page < 1 { page = 1 }
	if limit < 1 || limit > 500 { limit = 50 }
	offset := (page - 1) * limit

	total := 0
	if r, _ := dbq(svc, db, fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)); r != nil {
		fmt.Sscanf(strings.TrimSpace(r.Stdout), "%d", &total)
	}

	// Column names
	columns := make([]string, 0)
	if r, _ := dbq(svc, db, fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_schema='%s' AND table_name='%s' ORDER BY ordinal_position", db, table)); r != nil {
		for _, l := range strings.Split(r.Stdout, "\n") {
			if l = strings.TrimSpace(l); l != "" { columns = append(columns, l) }
		}
	}

	// Data
	rows := make([][]interface{}, 0)
	if r, _ := dbq(svc, db, fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", table, limit, offset)); r != nil {
		for _, l := range strings.Split(r.Stdout, "\n") {
			if l = strings.TrimSpace(l); l == "" { continue }
			fields := strings.Split(l, "\t")
			row := make([]interface{}, len(fields))
			for i, f := range fields {
				if f == "NULL" { row[i] = nil } else { row[i] = f }
			}
			rows = append(rows, row)
		}
	}

	return ok(c, map[string]interface{}{
		"columns": columns, "rows": rows, "total": total,
		"page": page, "limit": limit, "pages": (total + limit - 1) / limit,
	})
}

func RunQuery(c echo.Context) error {
	var req struct{ DB, Service, Query string }
	c.Bind(&req)
	if req.DB == "" || req.Service == "" || req.Query == "" { return fail(c, 400, "db, service, query required") }
	// Use column headers
	res, _ := exec.DockerCompose("exec", "-T", "-e", "MYSQL_PWD=root", req.Service, "mysql", "-u", "root", req.DB, "-e", req.Query)
	if res == nil { return ok(c, QueryResult{Error: "failed to execute"}) }
	if res.ExitCode != 0 { return ok(c, QueryResult{Error: res.Stderr}) }
	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	if len(lines) == 0 { return ok(c, QueryResult{Columns: []string{}, Rows: [][]interface{}{}, Count: 0}) }
	columns := strings.Split(lines[0], "\t")
	rows := make([][]interface{}, 0)
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "" { continue }
		fields := strings.Split(l, "\t")
		row := make([]interface{}, len(fields))
		for i, f := range fields {
			if f == "NULL" { row[i] = nil } else { row[i] = f }
		}
		rows = append(rows, row)
	}
	return ok(c, QueryResult{Columns: columns, Rows: rows, Count: len(rows)})
}
