package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/picassio/docker-magento-multiple-php/ui/server/exec"
)

type TableInfo struct {
	Name   string `json:"name"`
	Engine string `json:"engine"`
	Rows   string `json:"rows"`
	Size   string `json:"size"`
}

type ColumnInfo struct {
	Field   string `json:"field"`
	Type    string `json:"type"`
	Null    string `json:"null"`
	Key     string `json:"key"`
	Default string `json:"default"`
	Extra   string `json:"extra"`
}

type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
	Error   string          `json:"error,omitempty"`
}

func dbExec(service, query string) (*exec.Result, error) {
	return exec.DockerCompose("exec", "-T", "-e", "MYSQL_PWD=root",
		service, "mysql", "-u", "root", "-N", "-e", query)
}

func dbExecDB(service, database, query string) (*exec.Result, error) {
	return exec.DockerCompose("exec", "-T", "-e", "MYSQL_PWD=root",
		service, "mysql", "-u", "root", "-N", database, "-e", query)
}

// GET /api/dbmanager/tables?db=magento&service=mysql
func ListTables(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	service := r.URL.Query().Get("service")
	if db == "" || service == "" {
		jsonError(w, "db and service required", 400)
		return
	}

	res, err := dbExecDB(service, db,
		"SELECT table_name, engine, table_rows, ROUND((data_length+index_length)/1024/1024,2) "+
			"FROM information_schema.tables WHERE table_schema='"+db+"' ORDER BY table_name")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var tables []TableInfo
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		t := TableInfo{Name: fields[0]}
		if len(fields) > 1 {
			t.Engine = fields[1]
		}
		if len(fields) > 2 {
			t.Rows = fields[2]
		}
		if len(fields) > 3 {
			t.Size = fields[3] + " MB"
		}
		tables = append(tables, t)
	}

	if tables == nil {
		tables = make([]TableInfo, 0)
	}
	jsonOK(w, tables)
}

// GET /api/dbmanager/columns?db=magento&service=mysql&table=catalog_product_entity
func DescribeTable(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	service := r.URL.Query().Get("service")
	table := r.URL.Query().Get("table")
	if db == "" || service == "" || table == "" {
		jsonError(w, "db, service, and table required", 400)
		return
	}

	// Sanitize table name
	if strings.ContainsAny(table, "';\"\\") {
		jsonError(w, "invalid table name", 400)
		return
	}

	res, err := dbExecDB(service, db, "DESCRIBE `"+table+"`")
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var columns []ColumnInfo
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		c := ColumnInfo{Field: fields[0]}
		if len(fields) > 1 {
			c.Type = fields[1]
		}
		if len(fields) > 2 {
			c.Null = fields[2]
		}
		if len(fields) > 3 {
			c.Key = fields[3]
		}
		if len(fields) > 4 {
			c.Default = fields[4]
		}
		if len(fields) > 5 {
			c.Extra = fields[5]
		}
		columns = append(columns, c)
	}

	if columns == nil {
		columns = make([]ColumnInfo, 0)
	}
	jsonOK(w, columns)
}

// GET /api/dbmanager/data?db=magento&service=mysql&table=catalog_product_entity&page=1&limit=50
func TableData(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	service := r.URL.Query().Get("service")
	table := r.URL.Query().Get("table")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if db == "" || service == "" || table == "" {
		jsonError(w, "db, service, and table required", 400)
		return
	}
	if strings.ContainsAny(table, "';\"\\") {
		jsonError(w, "invalid table name", 400)
		return
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Get total count
	countRes, _ := dbExecDB(service, db, fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table))
	total := 0
	if countRes != nil {
		fmt.Sscanf(strings.TrimSpace(countRes.Stdout), "%d", &total)
	}

	// Get column names
	colRes, _ := dbExecDB(service, db,
		fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_schema='%s' AND table_name='%s' ORDER BY ordinal_position", db, table))
	var columns []string
	if colRes != nil {
		for _, line := range strings.Split(colRes.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				columns = append(columns, line)
			}
		}
	}

	// Get data
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", table, limit, offset)
	res, err := dbExecDB(service, db, query)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var rows [][]interface{}
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		row := make([]interface{}, len(fields))
		for i, f := range fields {
			if f == "NULL" {
				row[i] = nil
			} else {
				row[i] = f
			}
		}
		rows = append(rows, row)
	}

	if rows == nil {
		rows = make([][]interface{}, 0)
	}
	if columns == nil {
		columns = make([]string, 0)
	}

	jsonOK(w, map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"pages":   (total + limit - 1) / limit,
	})
}

// POST /api/dbmanager/query — run arbitrary SQL query
func RunQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DB      string `json:"db"`
		Service string `json:"service"`
		Query   string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}
	if req.DB == "" || req.Service == "" || req.Query == "" {
		jsonError(w, "db, service, and query required", 400)
		return
	}

	// Execute with column headers
	res, err := exec.DockerCompose("exec", "-T", "-e", "MYSQL_PWD=root",
		req.Service, "mysql", "-u", "root", req.DB, "-e", req.Query)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	if res.ExitCode != 0 {
		jsonOK(w, QueryResult{Error: res.Stderr})
		return
	}

	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	if len(lines) == 0 {
		jsonOK(w, QueryResult{Columns: []string{}, Rows: [][]interface{}{}, Count: 0})
		return
	}

	// First line = column headers
	columns := strings.Split(lines[0], "\t")
	var rows [][]interface{}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		row := make([]interface{}, len(fields))
		for i, f := range fields {
			if f == "NULL" {
				row[i] = nil
			} else {
				row[i] = f
			}
		}
		rows = append(rows, row)
	}
	if rows == nil {
		rows = make([][]interface{}, 0)
	}

	jsonOK(w, QueryResult{Columns: columns, Rows: rows, Count: len(rows)})
}
