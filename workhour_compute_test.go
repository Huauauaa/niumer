package main

import (
	"testing"
	"time"
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

func TestParseWorkHourWindowsFromShiftNameZh_flex(t *testing.T) {
	zh := "China/Flex,Work:08:00-17:30,Rest 12:00-13:30/17:30-18:00,Core :09:00-17:30,Card: 05:00-04:59"
	w, ok := parseWorkHourWindowsFromShiftNameZh(zh)
	if !ok || len(w) != 4 {
		t.Fatalf("want 4 windows ok=true, got ok=%v len=%d", ok, len(w))
	}
	want := []workHourTimeWindow{
		{0, 5 * time.Hour},
		{8 * time.Hour, 12 * time.Hour},
		{13*time.Hour + 30*time.Minute, 17*time.Hour + 30*time.Minute},
		{18 * time.Hour, 24 * time.Hour},
	}
	for i := range want {
		if w[i].start != want[i].start || w[i].end != want[i].end {
			t.Fatalf("window[%d] want %v-%v, got %v-%v", i, want[i].start, want[i].end, w[i].start, w[i].end)
		}
	}
}

func TestEffectiveWorkHoursForRecord_flexShiftWindows(t *testing.T) {
	zh := "China/Flex,Work:08:00-17:30,Rest 12:00-13:30/17:30-18:00,Core :09:00-17:30"
	wins, ok := parseWorkHourWindowsFromShiftNameZh(zh)
	if !ok {
		t.Fatal("parse failed")
	}
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "09:00",
		LateClockInTime:  "21:00",
	}
	// 日班 3h+4h；晚间 18–21 计 3h → 10h
	if g := effectiveWorkHoursForRecordWithWindows(r, wins); g != 10 {
		t.Fatalf("want 10, got %v", g)
	}
}

func TestEffectiveWorkHoursForRecord_flexShiftWindowsEarlyMorning(t *testing.T) {
	zh := "China/Flex,Work:08:00-17:30,Rest 12:00-13:30/17:30-18:00"
	wins, ok := parseWorkHourWindowsFromShiftNameZh(zh)
	if !ok {
		t.Fatal("parse failed")
	}
	r := AttendanceRecord{
		AttendanceDate:   "2026-04-21",
		EarlyClockInTime: "02:00",
		LateClockInTime:  "04:00",
	}
	if g := effectiveWorkHoursForRecordWithWindows(r, wins); g != 2 {
		t.Fatalf("want 2 (凌晨 02–04 在 [0,05) 内), got %v", g)
	}
}
