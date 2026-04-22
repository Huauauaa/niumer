package main

import (
    "database/sql"
    "errors"
    "fmt"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/google/uuid"
)

// CustomReminder is persisted in reminder.db (custom_reminders).
type CustomReminder struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Date string `json:"date"` // YYYY-MM-DD
}

const reminderSchema = `
CREATE TABLE IF NOT EXISTS custom_reminders (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    date TEXT NOT NULL,
    created_at INTEGER NOT NULL
);
`

var reminderDateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (a *App) openReminderDB() (*sql.DB, error) {
    p, err := a.resolvedReminderDBPath()
    if err != nil {
        return nil, err
    }
    dsn := "file:" + filepath.ToSlash(p) + "?_pragma=busy_timeout(5000)"
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(1)
    if _, err := db.Exec(reminderSchema); err != nil {
        _ = db.Close()
        return nil, err
    }
    return db, nil
}

// ListCustomReminders returns all reminders ordered by date then creation time.
func (a *App) ListCustomReminders() ([]CustomReminder, error) {
    db, err := a.openReminderDB()
    if err != nil {
        return nil, err
    }
    defer db.Close()

    q := `SELECT id, name, date FROM custom_reminders ORDER BY date ASC, created_at ASC, id ASC`
    rows, err := db.Query(q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []CustomReminder
    for rows.Next() {
        var r CustomReminder
        if err := rows.Scan(&r.ID, &r.Name, &r.Date); err != nil {
            return nil, err
        }
        out = append(out, r)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    if out == nil {
        out = []CustomReminder{}
    }
    return out, nil
}

// AddCustomReminder inserts a row and returns the new id.
func (a *App) AddCustomReminder(name, date string) (string, error) {
    name = strings.TrimSpace(name)
    if name == "" {
        return "", errors.New("name is empty")
    }
    date = strings.TrimSpace(date)
    if !reminderDateRe.MatchString(date) {
        return "", errors.New("date must be YYYY-MM-DD")
    }
    db, err := a.openReminderDB()
    if err != nil {
        return "", err
    }
    defer db.Close()

    id := uuid.New().String()
    _, err = db.Exec(
        `INSERT INTO custom_reminders (id, name, date, created_at) VALUES (?, ?, ?, ?)`,
        id, name, date, time.Now().UnixMilli(),
    )
    if err != nil {
        return "", err
    }
    return id, nil
}

// UpdateCustomReminder updates name and date for an existing id.
func (a *App) UpdateCustomReminder(id, name, date string) error {
    id = strings.TrimSpace(id)
    if id == "" {
        return errors.New("id is empty")
    }
    name = strings.TrimSpace(name)
    if name == "" {
        return errors.New("name is empty")
    }
    date = strings.TrimSpace(date)
    if !reminderDateRe.MatchString(date) {
        return errors.New("date must be YYYY-MM-DD")
    }
    db, err := a.openReminderDB()
    if err != nil {
        return err
    }
    defer db.Close()

    res, err := db.Exec(
        `UPDATE custom_reminders SET name = ?, date = ? WHERE id = ?`,
        name, date, id,
    )
    if err != nil {
        return err
    }
    n, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if n == 0 {
        return errors.New("reminder not found")
    }
    return nil
}

// DeleteCustomReminder removes one row: by id when it matches, otherwise by name+date
// (covers bridges where the client id string does not exactly match the DB row).
func (a *App) DeleteCustomReminder(id string, name string, date string) error {
    fmt.Println("DeleteCustomReminder", id, name, date)
    id = strings.TrimSpace(id)
    name = strings.TrimSpace(name)
    date = strings.TrimSpace(date)

    db, err := a.openReminderDB()
    if err != nil {
        return err
    }
    defer db.Close()

    if id != "" {
        res, errExec := db.Exec(`DELETE FROM custom_reminders WHERE id = ?`, id)
        if errExec != nil {
            return errExec
        }
        n, errRA := res.RowsAffected()
        if errRA != nil {
            return errRA
        }
        if n > 0 {
            return nil
        }
    }

    if name != "" && reminderDateRe.MatchString(date) {
        _, err = db.Exec(`DELETE FROM custom_reminders WHERE name = ? AND date = ?`, name, date)
        return err
    }

    if id != "" {
        return errors.New("reminder not found")
    }
    return errors.New("cannot delete: missing id and valid name/date")
}
