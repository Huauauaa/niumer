package main

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// workHourTimeWindow is one half-open [start, end) slice of a calendar day in loc for counting.
type workHourTimeWindow struct {
	start, end time.Duration
}

// defaultWorkHourTimeWindows 与历史硬编码一致（未拿到 shiftNameZh 或解析失败时使用）。
func defaultWorkHourTimeWindows() []workHourTimeWindow {
	return []workHourTimeWindow{
		{8 * time.Hour, 12 * time.Hour},
		{13*time.Hour + 30*time.Minute, 17*time.Hour + 30*time.Minute},
		{18 * time.Hour, 24 * time.Hour},
	}
}

var (
	workFromShiftRe    = regexp.MustCompile(`(?i)Work:\s*(\d{1,2}:\d{2})\s*-\s*(\d{1,2}:\d{2})`)
	restSegFromShiftRe = regexp.MustCompile(`(?i),Rest\s*([^,]+)`)
	restPairRe         = regexp.MustCompile(`(\d{1,2}:\d{2})\s*-\s*(\d{1,2}:\d{2})`)
	cardFromShiftRe    = regexp.MustCompile(`(?i),Card:\s*(\d{1,2}:\d{2})\s*-\s*(\d{1,2}:\d{2})`)
)

func hhmmToDayDuration(s string) (time.Duration, bool) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 24 || m < 0 || m > 59 {
		return 0, false
	}
	if h == 24 && m != 0 {
		return 0, false
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute, true
}

// parseWorkHourWindowsFromShiftNameZh 从 shiftNameZh 解析有效计时段：Work 与 Rest 相减得日班窗；弹性（Flex）或含 Card 时
// 另加晚间 [max(18:00,Work 结束), 24:00) 与凌晨 [00:00, 05:00)（与跨天 Card 如 05:00–04:59 的考勤覆盖一致）；Core 不参与计算。
func parseWorkHourWindowsFromShiftNameZh(shift string) ([]workHourTimeWindow, bool) {
	shift = strings.TrimSpace(shift)
	if shift == "" {
		return nil, false
	}
	sm := workFromShiftRe.FindStringSubmatch(shift)
	if len(sm) < 3 {
		return nil, false
	}
	ws, ok1 := hhmmToDayDuration(sm[1])
	we, ok2 := hhmmToDayDuration(sm[2])
	if !ok1 || !ok2 || we <= ws {
		return nil, false
	}
	work := workHourTimeWindow{start: ws, end: we}

	var rests []workHourTimeWindow
	if m := restSegFromShiftRe.FindStringSubmatch(shift); len(m) >= 2 {
		seg := strings.TrimSpace(m[1])
		for _, chunk := range strings.Split(seg, "/") {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}
			pm := restPairRe.FindStringSubmatch(chunk)
			if len(pm) < 3 {
				continue
			}
			rs, okA := hhmmToDayDuration(pm[1])
			re, okB := hhmmToDayDuration(pm[2])
			if !okA || !okB || re <= rs {
				continue
			}
			rests = append(rests, workHourTimeWindow{start: rs, end: re})
		}
	}
	out := subtractRestsFromWork(work, rests)
	if len(out) == 0 {
		return nil, false
	}
	out = appendFlexEveningAndEarlyWindows(shift, work, out)
	out = mergeWorkHourWindows(out)
	return out, true
}

// appendFlexEveningAndEarlyWindows 为弹性/打卡跨天班次补上默认「晚间 + 凌晨」计时段（与日班 Work 窗不交叠）。
func appendFlexEveningAndEarlyWindows(shift string, work workHourTimeWindow, wins []workHourTimeWindow) []workHourTimeWindow {
	low := strings.ToLower(shift)
	if !strings.Contains(low, "flex") && !strings.Contains(low, "card:") {
		return wins
	}
	evStart := maxDur(18*time.Hour, work.end)
	if evStart < 24*time.Hour {
		wins = append(wins, workHourTimeWindow{start: evStart, end: 24 * time.Hour})
	}
	overnightCard := false
	if m := cardFromShiftRe.FindStringSubmatch(shift); len(m) >= 3 {
		cs, okA := hhmmToDayDuration(m[1])
		ce, okB := hhmmToDayDuration(m[2])
		if okA && okB && ce < cs {
			overnightCard = true
			wins = append(wins, workHourTimeWindow{start: 0, end: 5 * time.Hour})
		}
	}
	if !overnightCard && strings.Contains(low, "flex") {
		wins = append(wins, workHourTimeWindow{start: 0, end: 5 * time.Hour})
	}
	return wins
}

func mergeWorkHourWindows(w []workHourTimeWindow) []workHourTimeWindow {
	if len(w) == 0 {
		return w
	}
	sort.Slice(w, func(i, j int) bool {
		if w[i].start == w[j].start {
			return w[i].end < w[j].end
		}
		return w[i].start < w[j].start
	})
	out := []workHourTimeWindow{w[0]}
	for _, cur := range w[1:] {
		last := &out[len(out)-1]
		if cur.start <= last.end {
			if cur.end > last.end {
				last.end = cur.end
			}
		} else {
			out = append(out, cur)
		}
	}
	return out
}

func subtractRestsFromWork(work workHourTimeWindow, rests []workHourTimeWindow) []workHourTimeWindow {
	cur := []workHourTimeWindow{work}
	for _, r := range rests {
		rs := maxDur(r.start, work.start)
		re := minDur(r.end, work.end)
		if re <= rs {
			continue
		}
		rClip := workHourTimeWindow{start: rs, end: re}
		var next []workHourTimeWindow
		for _, c := range cur {
			next = append(next, subtractOneWindow(c, rClip)...)
		}
		cur = next
		if len(cur) == 0 {
			return nil
		}
	}
	return cur
}

func subtractOneWindow(c, r workHourTimeWindow) []workHourTimeWindow {
	if r.end <= c.start || r.start >= c.end {
		return []workHourTimeWindow{c}
	}
	var out []workHourTimeWindow
	if r.start > c.start {
		end := minDur(r.start, c.end)
		if end > c.start {
			out = append(out, workHourTimeWindow{start: c.start, end: end})
		}
	}
	if r.end < c.end {
		st := maxDur(r.end, c.start)
		if c.end > st {
			out = append(out, workHourTimeWindow{start: st, end: c.end})
		}
	}
	return out
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

// effectiveWorkHoursForRecordWithWindows 计算 [early,late] 与各有效时段并集（小时）。
func effectiveWorkHoursForRecordWithWindows(r AttendanceRecord, windows []workHourTimeWindow) float64 {
	if strings.TrimSpace(r.EarlyClockInTime) == "" || strings.TrimSpace(r.LateClockInTime) == "" {
		return 0
	}
	if len(windows) == 0 {
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
	for _, w := range windows {
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

// effectiveWorkHoursForRecord 使用内置默认三段窗口（单测与未刷新班次时）。
func effectiveWorkHoursForRecord(r AttendanceRecord) float64 {
	return effectiveWorkHoursForRecordWithWindows(r, defaultWorkHourTimeWindows())
}
