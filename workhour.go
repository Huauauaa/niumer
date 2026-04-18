package main

import (
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// AttendanceRecord mirrors attendance_records table (JSON field names from API / SQLite).
type AttendanceRecord struct {
	ID                        int64  `json:"id"`
	CreationDate              string `json:"creationDate"`
	CreatedBy                 string `json:"createdBy"`
	LastUpdateDate            string `json:"lastUpdateDate"`
	LastUpdatedBy             string `json:"lastUpdatedBy"`
	OriginalID                string `json:"originalId"`
	HrID                      int64  `json:"hrId"`
	DataSource                string `json:"dataSource"`
	ClockInReason             string `json:"clockInReason"`
	AttendanceDate            string `json:"attendanceDate"`
	ClockInDate               string `json:"clockInDate"`
	ClockInTime               string `json:"clockInTime"`
	DayID                     string `json:"dayId"`
	ClockingInSequenceNumber  int64  `json:"clockingInSequenceNumber"`
	EarlyClockInTime          string `json:"earlyClockInTime"`
	LateClockInTime           string `json:"lateClockInTime"`
	ClockInType               string `json:"clockInType"`
	EarlyClockInType          string `json:"earlyClockInType"`
	LateClockInType           string `json:"lateClockInType"`
	AttendanceStatus          string `json:"attendanceStatus"`
	MinuteNumber              string `json:"minuteNumber"`
	HourNumber                string `json:"hourNumber"`
	AttendProcessID           string `json:"attendProcessId"`
	WorkDay                   string `json:"workDay"`
	AttendanceStatusCode      string `json:"attendanceStatusCode"`
	EarlyClockInReason        string `json:"earlyClockInReason"`
	LateClockInReason         string `json:"lateClockInReason"`
	EarlyClockTag             string `json:"earlyClockTag"`
	LateClockTag              string `json:"lateClockTag"`
}

const workHourSchema = `
CREATE TABLE IF NOT EXISTS attendance_records (
    id INTEGER PRIMARY KEY,
    creationDate TEXT,
    createdBy TEXT,
    lastUpdateDate TEXT,
    lastUpdatedBy TEXT,
    originalId TEXT,
    hrId INTEGER,
    dataSource TEXT,
    clockInReason TEXT,
    attendanceDate TEXT,
    clockInDate TEXT,
    clockInTime TEXT,
    dayId TEXT,
    clockingInSequenceNumber INTEGER,
    earlyClockInTime TEXT,
    lateClockInTime TEXT,
    clockInType TEXT,
    earlyClockInType TEXT,
    lateClockInType TEXT,
    attendanceStatus TEXT,
    minuteNumber TEXT,
    hourNumber TEXT,
    attendProcessId TEXT,
    workDay TEXT,
    attendanceStatusCode TEXT,
    earlyClockInReason TEXT,
    lateClockInReason TEXT,
    earlyClockTag TEXT,
    lateClockTag TEXT
);
`

func (a *App) openWorkHourDB() (*sql.DB, error) {
	p, err := a.resolvedWorkHourDBPath()
	if err != nil {
		return nil, err
	}
	dsn := "file:" + filepath.ToSlash(p) + "?_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(workHourSchema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func ns(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func ni(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}
	return 0
}

// GetWorkHourRecords returns all rows from attendance_records (newest dates first).
func (a *App) GetWorkHourRecords() ([]AttendanceRecord, error) {
	db, err := a.openWorkHourDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := `SELECT id, creationDate, createdBy, lastUpdateDate, lastUpdatedBy, originalId, hrId, dataSource,
		clockInReason, attendanceDate, clockInDate, clockInTime, dayId, clockingInSequenceNumber,
		earlyClockInTime, lateClockInTime, clockInType, earlyClockInType, lateClockInType,
		attendanceStatus, minuteNumber, hourNumber, attendProcessId, workDay, attendanceStatusCode,
		earlyClockInReason, lateClockInReason, earlyClockTag, lateClockTag
		FROM attendance_records
		ORDER BY COALESCE(attendanceDate, '') DESC, id DESC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AttendanceRecord
	for rows.Next() {
		var (
			id, hrID, seq                        sql.NullInt64
			creationDate, createdBy              sql.NullString
			lastUpdateDate, lastUpdatedBy        sql.NullString
			originalID, dataSource, clockInReason sql.NullString
			attendanceDate, clockInDate, clockInTime sql.NullString
			dayID                                sql.NullString
			earlyClockInTime, lateClockInTime    sql.NullString
			clockInType, earlyClockInType, lateClockInType sql.NullString
			attendanceStatus, minuteNumber, hourNumber sql.NullString
			attendProcessID, workDay, attendanceStatusCode sql.NullString
			earlyClockInReason, lateClockInReason sql.NullString
			earlyClockTag, lateClockTag          sql.NullString
		)
		if err := rows.Scan(
			&id, &creationDate, &createdBy, &lastUpdateDate, &lastUpdatedBy, &originalID, &hrID, &dataSource,
			&clockInReason, &attendanceDate, &clockInDate, &clockInTime, &dayID, &seq,
			&earlyClockInTime, &lateClockInTime, &clockInType, &earlyClockInType, &lateClockInType,
			&attendanceStatus, &minuteNumber, &hourNumber, &attendProcessID, &workDay, &attendanceStatusCode,
			&earlyClockInReason, &lateClockInReason, &earlyClockTag, &lateClockTag,
		); err != nil {
			return nil, err
		}
		out = append(out, AttendanceRecord{
			ID:                       ni(id),
			CreationDate:             ns(creationDate),
			CreatedBy:                ns(createdBy),
			LastUpdateDate:           ns(lastUpdateDate),
			LastUpdatedBy:            ns(lastUpdatedBy),
			OriginalID:               ns(originalID),
			HrID:                     ni(hrID),
			DataSource:               ns(dataSource),
			ClockInReason:            ns(clockInReason),
			AttendanceDate:           ns(attendanceDate),
			ClockInDate:              ns(clockInDate),
			ClockInTime:              ns(clockInTime),
			DayID:                    ns(dayID),
			ClockingInSequenceNumber: ni(seq),
			EarlyClockInTime:         ns(earlyClockInTime),
			LateClockInTime:          ns(lateClockInTime),
			ClockInType:              ns(clockInType),
			EarlyClockInType:         ns(earlyClockInType),
			LateClockInType:          ns(lateClockInType),
			AttendanceStatus:         ns(attendanceStatus),
			MinuteNumber:             ns(minuteNumber),
			HourNumber:               ns(hourNumber),
			AttendProcessID:          ns(attendProcessID),
			WorkDay:                  ns(workDay),
			AttendanceStatusCode:     ns(attendanceStatusCode),
			EarlyClockInReason:       ns(earlyClockInReason),
			LateClockInReason:        ns(lateClockInReason),
			EarlyClockTag:            ns(earlyClockTag),
			LateClockTag:             ns(lateClockTag),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []AttendanceRecord{}
	}
	return out, nil
}

// upsertAttendanceRecords writes or replaces rows in attendance_records.
func (a *App) upsertAttendanceRecords(records []AttendanceRecord) error {
	if len(records) == 0 {
		return nil
	}
	db, err := a.openWorkHourDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `INSERT OR REPLACE INTO attendance_records (
		id, creationDate, createdBy, lastUpdateDate, lastUpdatedBy, originalId, hrId, dataSource,
		clockInReason, attendanceDate, clockInDate, clockInTime, dayId, clockingInSequenceNumber,
		earlyClockInTime, lateClockInTime, clockInType, earlyClockInType, lateClockInType,
		attendanceStatus, minuteNumber, hourNumber, attendProcessId, workDay, attendanceStatusCode,
		earlyClockInReason, lateClockInReason, earlyClockTag, lateClockTag
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, r := range records {
		_, err := tx.Exec(q,
			r.ID, r.CreationDate, r.CreatedBy, r.LastUpdateDate, r.LastUpdatedBy, r.OriginalID, r.HrID, r.DataSource,
			r.ClockInReason, r.AttendanceDate, r.ClockInDate, r.ClockInTime, r.DayID, r.ClockingInSequenceNumber,
			r.EarlyClockInTime, r.LateClockInTime, r.ClockInType, r.EarlyClockInType, r.LateClockInType,
			r.AttendanceStatus, r.MinuteNumber, r.HourNumber, r.AttendProcessID, r.WorkDay, r.AttendanceStatusCode,
			r.EarlyClockInReason, r.LateClockInReason, r.EarlyClockTag, r.LateClockTag,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
