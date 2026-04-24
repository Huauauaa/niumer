package main

import (
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// AttendanceRecord mirrors attendance_records table (JSON field names from API / SQLite).
type AttendanceRecord struct {
	ID                       int64  `json:"id"`
	CreationDate             string `json:"creationDate"`
	CreatedBy                string `json:"createdBy"`
	LastUpdateDate           string `json:"lastUpdateDate"`
	LastUpdatedBy            string `json:"lastUpdatedBy"`
	OriginalID               string `json:"originalId"`
	HrID                     int64  `json:"hrId"`
	DataSource               string `json:"dataSource"`
	ClockInReason            string `json:"clockInReason"`
	AttendanceDate           string `json:"attendanceDate"`
	ClockInDate              string `json:"clockInDate"`
	ClockInTime              string `json:"clockInTime"`
	DayID                    string `json:"dayId"`
	ClockingInSequenceNumber int64  `json:"clockingInSequenceNumber"`
	EarlyClockInTime         string `json:"earlyClockInTime"`
	LateClockInTime          string `json:"lateClockInTime"`
	ClockInType              string `json:"clockInType"`
	EarlyClockInType         string `json:"earlyClockInType"`
	LateClockInType          string `json:"lateClockInType"`
	AttendanceStatus         string `json:"attendanceStatus"`
	MinuteNumber             string `json:"minuteNumber"`
	HourNumber               string `json:"hourNumber"`
	AttendProcessID          string `json:"attendProcessId"`
	WorkDay                  string `json:"workDay"`
	AttendanceStatusCode     string `json:"attendanceStatusCode"`
	EarlyClockInReason       string `json:"earlyClockInReason"`
	LateClockInReason        string `json:"lateClockInReason"`
	EarlyClockTag            string `json:"earlyClockTag"`
	LateClockTag             string `json:"lateClockTag"`
	// EffectiveWorkHours 由应用层根据打卡与有效时段计算，不写入 SQLite。
	EffectiveWorkHours float64 `json:"effectiveWorkHours"`
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

const workHourUserProfileSchema = `
CREATE TABLE IF NOT EXISTS workhour_user_profile (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    userAccount TEXT,
    hrId INTEGER NOT NULL DEFAULT 0,
    shiftNameZh TEXT,
    userInfoJson TEXT,
    updatedAt TEXT
);
`

// WorkHourUserProfileView is persisted user metadata + /user-info data (JSON) for the Preferences UI.
type WorkHourUserProfileView struct {
	UserAccount  string `json:"userAccount"`
	HrID         int64  `json:"hrId"`
	ShiftNameZh  string `json:"shiftNameZh"`
	UpdatedAt    string `json:"updatedAt"`
	UserInfoJSON string `json:"userInfoJson"`
}

func migrateWorkHourUserProfileTable(db *sql.DB) {
	// Safe to re-run: duplicate column is ignored in drivers that error; we ignore any ALTER error.
	_, _ = db.Exec(`ALTER TABLE workhour_user_profile ADD COLUMN userInfoJson TEXT`)
}

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
	if _, err := db.Exec(workHourUserProfileSchema); err != nil {
		_ = db.Close()
		return nil, err
	}
	migrateWorkHourUserProfileTable(db)
	return db, nil
}

// upsertWorkHourUserProfile persists tenant + user-info fields (single logical row id=1).
// userInfoJSON is the raw JSON of /user-info `data` (or empty).
func (a *App) upsertWorkHourUserProfile(userAccount string, hrID int64, shiftNameZh, userInfoJSON string) error {
	db, err := a.openWorkHourDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(
		`INSERT OR REPLACE INTO workhour_user_profile (id, userAccount, hrId, shiftNameZh, userInfoJson, updatedAt) VALUES (1, ?, ?, ?, ?, ?)`,
		userAccount, hrID, shiftNameZh, nullString(userInfoJSON), time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func nullString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// hasWorkHourUserProfileInDB returns true when the logical profile row has enough data to use without calling user-info.
func (a *App) hasWorkHourUserProfileInDB() (bool, error) {
	if a == nil {
		return false, nil
	}
	db, err := a.openWorkHourDB()
	if err != nil {
		return false, err
	}
	defer db.Close()
	var ua sql.NullString
	var hrID sql.NullInt64
	err = db.QueryRow(
		`SELECT userAccount, hrId FROM workhour_user_profile WHERE id = 1`,
	).Scan(&ua, &hrID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if !ua.Valid || strings.TrimSpace(ua.String) == "" {
		return false, nil
	}
	if !hrID.Valid || hrID.Int64 == 0 {
		return false, nil
	}
	return true, nil
}

// readWorkHourUserProfileView loads id=1 row for UI or memory (best-effort).
func (a *App) readWorkHourUserProfileView() (WorkHourUserProfileView, error) {
	var out WorkHourUserProfileView
	if a == nil {
		return out, nil
	}
	db, err := a.openWorkHourDB()
	if err != nil {
		return out, err
	}
	defer db.Close()
	var ua, shift, at, uij sql.NullString
	var hrID sql.NullInt64
	err = db.QueryRow(
		`SELECT userAccount, hrId, shiftNameZh, updatedAt, userInfoJson FROM workhour_user_profile WHERE id = 1`,
	).Scan(&ua, &hrID, &shift, &at, &uij)
	if err == sql.ErrNoRows {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.UserAccount = ns(ua)
	out.HrID = ni(hrID)
	out.ShiftNameZh = ns(shift)
	out.UpdatedAt = ns(at)
	out.UserInfoJSON = ns(uij)
	return out, nil
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

// GetWorkHourRecords returns rows from attendance_records (newest dates first)，
// 展示层忽略上班或下班打卡时间为空的记录（仍保留在 SQLite 中）。
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
			id, hrID, seq                                  sql.NullInt64
			creationDate, createdBy                        sql.NullString
			lastUpdateDate, lastUpdatedBy                  sql.NullString
			originalID, dataSource, clockInReason          sql.NullString
			attendanceDate, clockInDate, clockInTime       sql.NullString
			dayID                                          sql.NullString
			earlyClockInTime, lateClockInTime              sql.NullString
			clockInType, earlyClockInType, lateClockInType sql.NullString
			attendanceStatus, minuteNumber, hourNumber     sql.NullString
			attendProcessID, workDay, attendanceStatusCode sql.NullString
			earlyClockInReason, lateClockInReason          sql.NullString
			earlyClockTag, lateClockTag                    sql.NullString
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
		rec := AttendanceRecord{
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
		}
		rec.EffectiveWorkHours = effectiveWorkHoursForRecordWithWindows(
			rec,
			a.workHourEffectiveWindows(),
		)
		if strings.TrimSpace(rec.EarlyClockInTime) == "" || strings.TrimSpace(rec.LateClockInTime) == "" {
			continue
		}
		out = append(out, rec)
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
