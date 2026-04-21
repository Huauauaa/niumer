package main

import (
	"math"
	"strings"
	"time"
)

// 有效工时窗口（与考勤日同一日历日，左闭右开，避免相邻窗口重复计分钟）。
var effectiveWorkHourWindows = []struct {
	start, end time.Duration
}{
	{8 * time.Hour, 12 * time.Hour},
	{13*time.Hour + 30*time.Minute, 17*time.Hour + 30*time.Minute},
	{18 * time.Hour, 24 * time.Hour},
}

func maxDur(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// anchorDate 用于解析仅含时刻的字符串；优先 attendanceDate，其次 clockInDate。
func anchorDateForRecord(r AttendanceRecord) string {
	for _, d := range []string{r.AttendanceDate, r.ClockInDate} {
		d = strings.TrimSpace(d)
		if len(d) >= 10 {
			return d[:10]
		}
	}
	return time.Now().Format("2006-01-02")
}

// parseClockPair 将 early / late 解析为同一 loc 下的两个时刻；支持整段日期时间或仅时刻（与 anchor 组合）。
func parseClockPair(anchorDate, earlyStr, lateStr string, loc *time.Location) (early, late time.Time, ok bool) {
	early, okE := parseOneClock(anchorDate, earlyStr, loc)
	late, okL := parseOneClock(anchorDate, lateStr, loc)
	if !okE || !okL {
		return time.Time{}, time.Time{}, false
	}
	return early, late, true
}

func parseOneClock(anchorDate, s string, loc *time.Location) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	fullLayouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
	}
	for _, layout := range fullLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, true
		}
	}
	if len(anchorDate) < 10 {
		return time.Time{}, false
	}
	datePart := anchorDate[:10]
	if s == "24:00" || s == "24:00:00" {
		day0, err := time.ParseInLocation("2006-01-02", datePart, loc)
		if err != nil {
			return time.Time{}, false
		}
		return day0.Add(24 * time.Hour), true
	}
	for _, dl := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(dl, datePart+" "+s, loc); err == nil {
			return t, true
		}
	}
	for _, tl := range []string{"15:04:05", "15:04"} {
		clock, err := time.ParseInLocation(tl, s, loc)
		if err != nil {
			continue
		}
		y, m, d := mustDateParts(datePart, loc)
		hh, mm, ss := clock.Clock()
		return time.Date(y, m, d, hh, mm, ss, 0, loc), true
	}
	return time.Time{}, false
}

func mustDateParts(datePart string, loc *time.Location) (y int, m time.Month, d int) {
	t, err := time.ParseInLocation("2006-01-02", datePart, loc)
	if err != nil {
		now := time.Now().In(loc)
		return now.Date()
	}
	return t.Date()
}

// effectiveWorkHoursForRecord 计算当天有效工时（小时）。earlyClockInTime 或 lateClockInTime
// 任一缺失或无法解析则为 0；否则为 [early, late] 与有效时段并集的交集时长。
func effectiveWorkHoursForRecord(r AttendanceRecord) float64 {
	if strings.TrimSpace(r.EarlyClockInTime) == "" || strings.TrimSpace(r.LateClockInTime) == "" {
		return 0
	}
	loc := time.Local
	anchor := anchorDateForRecord(r)
	early, late, ok := parseClockPair(anchor, r.EarlyClockInTime, r.LateClockInTime, loc)
	if !ok {
		return 0
	}
	if !late.After(early) {
		return 0
	}
	day := time.Date(early.Year(), early.Month(), early.Day(), 0, 0, 0, 0, loc)
	segStart := early.Sub(day)
	segEnd := late.Sub(day)
	var total time.Duration
	for _, w := range effectiveWorkHourWindows {
		s := maxDur(segStart, w.start)
		e := minDur(segEnd, w.end)
		if e > s {
			total += e - s
		}
	}
	h := total.Hours()
	if h < 0 {
		return 0
	}
	return math.Round(h*100) / 100
}
