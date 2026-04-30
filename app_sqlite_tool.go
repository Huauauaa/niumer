package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	_ "modernc.org/sqlite"
)

type SQLiteColumn struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	NotNull bool   `json:"notNull"`
	PK      bool   `json:"pk"`
	Default any    `json:"default"`
}

type SQLiteTableInfo struct {
	Table     string         `json:"table"`
	Columns   []SQLiteColumn `json:"columns"`
	PKColumns []string       `json:"pkColumns"`
	RowID     bool           `json:"rowId"`
}

type SQLiteQueryResult struct {
	Table     string           `json:"table"`
	Columns   []SQLiteColumn   `json:"columns"`
	PKColumns []string         `json:"pkColumns"`
	Rows      []map[string]any `json:"rows"`
	Total     int64            `json:"total"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

func quoteIdent(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("empty identifier")
	}
	if strings.ContainsRune(s, 0) {
		return "", errors.New("invalid identifier")
	}
	// SQLite accepts quoted identifiers with "" escaping.
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`, nil
}

func (a *App) ChooseSQLiteDBPath() (string, error) {
	if a.ctx == nil {
		return "", errors.New("app not ready")
	}
	opts := runtime.OpenDialogOptions{
		Title: "Choose SQLite database",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "SQLite database",
				Pattern:     "*.db;*.sqlite;*.sqlite3;*.sqlite3-wal;*.sqlite3-shm",
			},
			{DisplayName: "All files", Pattern: "*"},
		},
	}
	a.muSQLiteTool.Lock()
	cur := a.sqliteToolDBPath
	a.muSQLiteTool.Unlock()
	if cur != "" {
		dir := filepath.Dir(cur)
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			opts.DefaultDirectory = dir
		}
	}
	return runtime.OpenFileDialog(a.ctx, opts)
}

func (a *App) SQLiteToolGetDBPath() string {
	a.muSQLiteTool.Lock()
	defer a.muSQLiteTool.Unlock()
	return a.sqliteToolDBPath
}

func (a *App) SQLiteToolOpenDB(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("path is empty")
	}
	path = filepath.Clean(path)
	if _, err := os.Stat(path); err != nil {
		return err
	}

	a.muSQLiteTool.Lock()
	defer a.muSQLiteTool.Unlock()

	if a.sqliteToolDB != nil {
		_ = a.sqliteToolDB.Close()
		a.sqliteToolDB = nil
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Best-effort pragmas.
	_, _ = db.Exec(`PRAGMA foreign_keys = ON;`)
	_, _ = db.Exec(`PRAGMA busy_timeout = 3000;`)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return err
	}
	a.sqliteToolDB = db
	a.sqliteToolDBPath = path
	return nil
}

func (a *App) sqliteToolGetDB() (*sql.DB, error) {
	a.muSQLiteTool.Lock()
	defer a.muSQLiteTool.Unlock()
	if a.sqliteToolDB == nil {
		return nil, errors.New("SQLite DB not opened")
	}
	return a.sqliteToolDB, nil
}

func (a *App) SQLiteToolListTables() ([]string, error) {
	db, err := a.sqliteToolGetDB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, nil
}

func (a *App) SQLiteToolDescribeTable(table string) (SQLiteTableInfo, error) {
	db, err := a.sqliteToolGetDB()
	if err != nil {
		return SQLiteTableInfo{}, err
	}
	qTable, err := quoteIdent(table)
	if err != nil {
		return SQLiteTableInfo{}, err
	}
	// Validate table exists.
	var exists int
	if err := db.QueryRow(`SELECT 1 FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SQLiteTableInfo{}, fmt.Errorf("table not found: %s", table)
		}
		return SQLiteTableInfo{}, err
	}

	pragma := fmt.Sprintf("PRAGMA table_info(%s);", qTable)
	rows, err := db.Query(pragma)
	if err != nil {
		return SQLiteTableInfo{}, err
	}
	defer rows.Close()

	var cols []SQLiteColumn
	var pkCols []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt any
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return SQLiteTableInfo{}, err
		}
		cols = append(cols, SQLiteColumn{
			Name:    name,
			Type:    ctype,
			NotNull: notnull != 0,
			PK:      pk != 0,
			Default: dflt,
		})
		if pk != 0 {
			pkCols = append(pkCols, name)
		}
	}
	sort.SliceStable(cols, func(i, j int) bool { return cols[i].Name < cols[j].Name })

	info := SQLiteTableInfo{
		Table:     table,
		Columns:   cols,
		PKColumns: pkCols,
		RowID:     len(pkCols) == 0, // we will use rowid as key when PK absent
	}
	return info, nil
}

func (a *App) SQLiteToolQueryTable(table string, limit, offset int) (SQLiteQueryResult, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	db, err := a.sqliteToolGetDB()
	if err != nil {
		return SQLiteQueryResult{}, err
	}
	info, err := a.SQLiteToolDescribeTable(table)
	if err != nil {
		return SQLiteQueryResult{}, err
	}
	qTable, _ := quoteIdent(table)

	var selectCols []string
	var pkCols []string
	if len(info.PKColumns) > 0 {
		pkCols = append(pkCols, info.PKColumns...)
	} else {
		// rowid works for most tables; for WITHOUT ROWID tables, queries may fail.
		selectCols = append(selectCols, `rowid AS __rowid`)
		pkCols = []string{"__rowid"}
		info.Columns = append([]SQLiteColumn{{Name: "__rowid", Type: "INTEGER", NotNull: true, PK: true}}, info.Columns...)
	}
	for _, c := range info.Columns {
		if c.Name == "__rowid" {
			continue
		}
		qc, err := quoteIdent(c.Name)
		if err != nil {
			continue
		}
		selectCols = append(selectCols, qc)
	}
	if len(selectCols) == 0 {
		return SQLiteQueryResult{}, errors.New("no selectable columns")
	}

	orderBy := ""
	if len(pkCols) > 0 {
		var parts []string
		for _, k := range pkCols {
			if k == "__rowid" {
				parts = append(parts, "__rowid DESC")
				continue
			}
			qk, err := quoteIdent(k)
			if err != nil {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s DESC", qk))
		}
		if len(parts) > 0 {
			orderBy = " ORDER BY " + strings.Join(parts, ", ")
		}
	}
	query := fmt.Sprintf("SELECT %s FROM %s%s LIMIT ? OFFSET ?;", strings.Join(selectCols, ", "), qTable, orderBy)

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return SQLiteQueryResult{}, err
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return SQLiteQueryResult{}, err
	}

	var outRows []map[string]any
	for rows.Next() {
		dest := make([]any, len(colNames))
		ptrs := make([]any, len(colNames))
		for i := range dest {
			ptrs[i] = &dest[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return SQLiteQueryResult{}, err
		}
		m := make(map[string]any, len(colNames))
		for i, n := range colNames {
			v := dest[i]
			// Normalize []byte to string for JSON.
			if b, ok := v.([]byte); ok {
				m[n] = string(b)
			} else {
				m[n] = v
			}
		}
		outRows = append(outRows, m)
	}

	var total int64
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(1) FROM %s;", qTable)).Scan(&total); err != nil {
		total = 0
	}

	return SQLiteQueryResult{
		Table:     table,
		Columns:   info.Columns,
		PKColumns: pkCols,
		Rows:      outRows,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func (a *App) SQLiteToolInsertRow(table string, values map[string]any) error {
	db, err := a.sqliteToolGetDB()
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return errors.New("values empty")
	}
	qTable, err := quoteIdent(table)
	if err != nil {
		return err
	}
	var rawCols []string
	for k := range values {
		if strings.TrimSpace(k) == "" {
			continue
		}
		rawCols = append(rawCols, k)
	}
	sort.Strings(rawCols)

	var cols []string
	var placeholders []string
	var args []any
	for _, k := range rawCols {
		qk, err := quoteIdent(k)
		if err != nil {
			continue
		}
		cols = append(cols, qk)
		placeholders = append(placeholders, "?")
		args = append(args, values[k])
	}
	if len(cols) == 0 {
		return errors.New("no valid columns")
	}
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", qTable, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	_, err = db.Exec(stmt, args...)
	return err
}

func (a *App) SQLiteToolUpdateRow(table string, key map[string]any, values map[string]any) error {
	db, err := a.sqliteToolGetDB()
	if err != nil {
		return err
	}
	if len(key) == 0 {
		return errors.New("key empty")
	}
	if len(values) == 0 {
		return errors.New("values empty")
	}
	qTable, err := quoteIdent(table)
	if err != nil {
		return err
	}
	var setCols []string
	var args []any
	var setKeys []string
	for k := range values {
		qk, err := quoteIdent(k)
		if err != nil {
			continue
		}
		setCols = append(setCols, fmt.Sprintf("%s = ?", qk))
		setKeys = append(setKeys, k)
	}
	sort.Strings(setKeys)
	setCols = setCols[:0]
	args = args[:0]
	for _, k := range setKeys {
		qk, _ := quoteIdent(k)
		setCols = append(setCols, fmt.Sprintf("%s = ?", qk))
		args = append(args, values[k])
	}

	var where []string
	var keyKeys []string
	for k := range key {
		keyKeys = append(keyKeys, k)
	}
	sort.Strings(keyKeys)
	for _, k := range keyKeys {
		if k == "__rowid" {
			where = append(where, "rowid = ?")
			args = append(args, key[k])
			continue
		}
		qk, err := quoteIdent(k)
		if err != nil {
			continue
		}
		where = append(where, fmt.Sprintf("%s = ?", qk))
		args = append(args, key[k])
	}
	if len(setCols) == 0 || len(where) == 0 {
		return errors.New("no valid set/where columns")
	}
	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", qTable, strings.Join(setCols, ", "), strings.Join(where, " AND "))
	_, err = db.Exec(stmt, args...)
	return err
}

func (a *App) SQLiteToolDeleteRow(table string, key map[string]any) error {
	db, err := a.sqliteToolGetDB()
	if err != nil {
		return err
	}
	if len(key) == 0 {
		return errors.New("key empty")
	}
	qTable, err := quoteIdent(table)
	if err != nil {
		return err
	}
	var where []string
	var args []any
	var keyKeys []string
	for k := range key {
		keyKeys = append(keyKeys, k)
	}
	sort.Strings(keyKeys)
	for _, k := range keyKeys {
		if k == "__rowid" {
			where = append(where, "rowid = ?")
			args = append(args, key[k])
			continue
		}
		qk, err := quoteIdent(k)
		if err != nil {
			continue
		}
		where = append(where, fmt.Sprintf("%s = ?", qk))
		args = append(args, key[k])
	}
	if len(where) == 0 {
		return errors.New("no valid where")
	}
	stmt := fmt.Sprintf("DELETE FROM %s WHERE %s;", qTable, strings.Join(where, " AND "))
	_, err = db.Exec(stmt, args...)
	return err
}

