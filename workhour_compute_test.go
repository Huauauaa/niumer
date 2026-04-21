package main

import (
	"testing"
)

func TestEffectiveWorkHoursForRecord_missingClock(t *testing.T) {
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "09:00",
		LateClockInTime:  "",
	}
	if g := effectiveWorkHoursForRecord(r); g != 0 {
		t.Fatalf("want 0 missing late, got %v", g)
	}
	r.EarlyClockInTime = ""
	r.LateClockInTime = "18:00"
	if g := effectiveWorkHoursForRecord(r); g != 0 {
		t.Fatalf("want 0 missing early, got %v", g)
	}
}

func TestEffectiveWorkHoursForRecord_fullDayInWindows(t *testing.T) {
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "08:00",
		LateClockInTime:  "24:00",
	}
	// 4 + 4 + 6 = 14h（24:00 为时段右端点，与左闭右开一致：到 24:00 整点为止）
	if g := effectiveWorkHoursForRecord(r); g != 14 {
		t.Fatalf("want 14, got %v", g)
	}
}

func TestEffectiveWorkHoursForRecord_partialOverlap(t *testing.T) {
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "11:00",
		LateClockInTime:  "14:00",
	}
	// 11-12: 1h; 13:30-14: 0.5h
	if g := effectiveWorkHoursForRecord(r); g != 1.5 {
		t.Fatalf("want 1.5, got %v", g)
	}
}

func TestEffectiveWorkHoursForRecord_outsideWindows(t *testing.T) {
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "12:05",
		LateClockInTime:  "13:00",
	}
	if g := effectiveWorkHoursForRecord(r); g != 0 {
		t.Fatalf("want 0, got %v", g)
	}
}

func TestEffectiveWorkHoursForRecord_eveningOnly(t *testing.T) {
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "19:00",
		LateClockInTime:  "22:00",
	}
	if g := effectiveWorkHoursForRecord(r); g != 3 {
		t.Fatalf("want 3, got %v", g)
	}
}

func TestEffectiveWorkHoursForRecord_typicalDay0921(t *testing.T) {
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "09:00",
		LateClockInTime:  "21:00",
	}
	// 9–12: 3h; 13:30–17:30: 4h; 18–21: 3h
	if g := effectiveWorkHoursForRecord(r); g != 10 {
		t.Fatalf("want 10, got %v", g)
	}
}
