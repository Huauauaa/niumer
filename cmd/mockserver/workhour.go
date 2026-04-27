package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func registerWorkHourRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /id", handleTenant)
	mux.HandleFunc("POST /user-info", handleUserInfo)
	mux.HandleFunc("POST /work-hour", handleWorkHour)
}

func handleTenant(w http.ResponseWriter, r *http.Request) {
	// Stable fake account: first rune is skipped as employeeQuery in fetchWorkHourUserInfo.
	const userAccount = "M10001"
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"tenant": map[string]string{
				"userAccount": userAccount,
			},
		},
	})
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	eq, _ := body["employeeQuery"].(string)
	hrID := int64(900001)
	if eq != "" {
		if n, err := strconv.ParseInt(strings.TrimSpace(eq), 10, 64); err == nil && n > 0 {
			hrID = n
		} else {
			hrID = int64(len(eq))*1000 + 42
		}
	}
	hrIdStr := strconv.FormatInt(hrID, 10)
	// 与线上一致：hrId 可为 string；全量 data 供 SQLite userInfoJson 与联调阅读。
	// 敏感位用 ****** 占位，hrId 必为可解析数字串（niumer 会转成 int64）。
	shiftNameZh := "China/Flex,Work:08:00-17:30,Rest 12:00-13:30/17:30-18:00,Core :09:00-17:30,Card: 05:00-04:59"
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      200,
		"messageCode": "attendance.response.ok",
		"messageText": "Success",
		"ok":          true,
		"data": map[string]any{
			"hrId":             hrIdStr,
			"attendanceScheme": "CNEX",
			"departmentDTO": map[string]any{
				"departmentCode":        "******",
				"departmentEnglishName": "Cloud Infrastructure Platform Dept",
				"departmentChineseName": "云基础设施平台部",
				"orgList": []any{
					"******", "******", "******",
				},
				"level":         nil,
				"employeeCount": nil,
				"isConvolution": true,
			},
			"shiftInformationDTO": map[string]any{
				"shiftId":                    15,
				"shiftCode":                  "FLXF",
				"shiftName":                  "班次编码:FLXF,服务起止时间:08:00-17:30,必须在岗时间:09:00-17:30,取卡时间:05:00-04:59",
				"shiftNameZh":                shiftNameZh,
				"shiftNameEn":                "China/Flex,Work:08:00-17:30,Rest 12:00-13:30/17:30-18:00,Core :09:00-17:30,Card: 05:00-04:59",
				"shiftCountryCode":           "CN",
				"shiftDuration":              8,
				"standardStartTime":          "08:00",
				"standardEndTime":            "17:30",
				"workCoreTimeStart":          "09:00",
				"workCoreTimeEnd":            "17:30",
				"shiftEffectiveTimeStart":    "05:00",
				"shiftEffectiveTimeEnd":      "04:59",
				"restStartTime":              "12:00",
				"restEndTime":                "13:30",
				"restStartTimeTwo":           nil,
				"restEndTimeTwo":             nil,
				"freeCard":                   nil,
				"scheduling":                 false,
				"shiftType":                  "弹性班次",
				"shiftTypeCode":              "1",
				"flexible":                   nil,
				"workCalendarName":           "中国大陆假期日历",
				"brushCardTypeId":            2,
				"brushCardTypeName":          "固定办公打卡",
				"brushCardTypeCode":          "No",
				"holidayCalenderId":          "55",
				"workCalendarCode":           "HW1_Chinese Mainland Holiday Calendar",
				"workTimeEffectiveStartDate": "2026-04-01",
				"workTimeEffectiveEndDate":   "2199-12-31",
				"shiftBreakTime":             nil,
			},
			"isTopmanager":          false,
			"isAppointManager":      false,
			"isManagerPilotDept":    false,
			"supplierCode":          "******",
			"supplier":              "北京外企德科人力资源服务上海有限公司",
			"personWorkType":        "2040",
			"personWorkTypeName":    "在场人力服务外包",
			"personWorkSubType":     "204003",
			"personWorkSubTypeName": "研发OD在场",
		},
	})
}

func handleWorkHour(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	hrID := jsonNumberToInt64(body["hrId"])
	if hrID == 0 {
		hrID = jsonNumberToInt64(body["hrId"])
	}
	today := time.Now().Format("2006-01-02")
	rec := map[string]any{
		"id":                       1,
		"creationDate":             today + " 09:00:00",
		"createdBy":                "mock",
		"lastUpdateDate":           today + " 09:00:00",
		"lastUpdatedBy":            "mock",
		"originalId":               "mock-orig-1",
		"hrId":                     hrID,
		"dataSource":               "MOCK",
		"clockInReason":            "",
		"attendanceDate":           today,
		"clockInDate":              today,
		"clockInTime":              "09:00",
		"dayId":                    today,
		"clockingInSequenceNumber": 1,
		"earlyClockInTime":         "09:00",
		"lateClockInTime":          "21:00",
		"clockInType":              "NORMAL",
		"earlyClockInType":         "",
		"lateClockInType":          "",
		"attendanceStatus":         "Present",
		"minuteNumber":             "480",
		"hourNumber":               "8",
		"attendProcessId":          "",
		"workDay":                  "1",
		"attendanceStatusCode":     "OK",
		"earlyClockInReason":       "",
		"lateClockInReason":        "",
		"earlyClockTag":            "",
		"lateClockTag":             "",
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"records": []map[string]any{rec},
		},
	})
}

func jsonNumberToInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case json.Number:
		i, _ := x.Int64()
		return i
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}
