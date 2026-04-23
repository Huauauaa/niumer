package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"niumer/internal/config"
)

// 默认 URL 来自 configs/*.yaml；可用 WORK_HOUR_* 环境变量覆盖。

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func workHourHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr, Timeout: 3 * time.Minute}
}

func addCookies(req *http.Request, cookies map[string]string) {
	for name, val := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: val})
	}
}

func postJSON(ctx context.Context, client *http.Client, url string, body any, cookies map[string]string) ([]byte, int, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	addCookies(req, cookies)
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return b, resp.StatusCode, nil
}

type tenantResp struct {
	Data *struct {
		Tenant *struct {
			UserAccount string `json:"userAccount"`
		} `json:"tenant"`
	} `json:"data"`
}

// userInfoResp matches POST /user-info (replaces legacy /hr-id envelope).
type userInfoResp struct {
	Status      int    `json:"status"`
	MessageCode string `json:"messageCode"`
	MessageText string `json:"messageText"`
	OK          bool   `json:"ok"`
	Data        *struct {
		HrID                interface{} `json:"hrId"` // number or string from upstream
		ShiftInformationDTO *struct {
			ShiftNameZh string `json:"shiftNameZh"`
		} `json:"shiftInformationDTO"`
	} `json:"data"`
}

type workHourUserInfo struct {
	HrID        int64
	ShiftNameZh string
}

func fetchUserAccount(ctx context.Context, client *http.Client, cookies map[string]string) (string, error) {
	url := envOr("WORK_HOUR_TENANT_URL", config.GetWorkHour().TenantURL)
	body := map[string]string{
		"rentId":     "HuaWei",
		"groupId":    "servicetimeflow",
		"tenantId":   "",
		"url":        "hr.huawei.com",
		"parentCode": "servicetimeflow",
	}
	raw, code, err := postJSON(ctx, client, url, body, cookies)
	if err != nil {
		return "", err
	}
	if code < 200 || code >= 300 {
		return "", fmt.Errorf("tenant 接口 HTTP %d: %s", code, truncate(string(raw), 500))
	}
	var tr tenantResp
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", err
	}
	if tr.Data == nil || tr.Data.Tenant == nil {
		return "", errors.New("tenant 响应缺少 data.tenant")
	}
	return tr.Data.Tenant.UserAccount, nil
}

func fetchWorkHourUserInfo(ctx context.Context, client *http.Client, cookies map[string]string, userAccount string) (workHourUserInfo, error) {
	var out workHourUserInfo
	if len(userAccount) < 2 {
		return out, errors.New("userAccount 过短")
	}
	url := envOr("WORK_HOUR_USER_INFO_URL", config.GetWorkHour().UserInfoURL)
	body := map[string]string{
		"employeeQuery": userAccount[1:],
		"queryDate":     time.Now().Format("2006-01-02"),
		"locale":        "zh",
		"platform":      "PC",
	}
	raw, code, err := postJSON(ctx, client, url, body, cookies)
	if err != nil {
		return out, err
	}
	if code < 200 || code >= 300 {
		return out, fmt.Errorf("user-info 接口 HTTP %d: %s", code, truncate(string(raw), 500))
	}
	var env userInfoResp
	if err := json.Unmarshal(raw, &env); err != nil {
		return out, err
	}
	if env.Status != 200 || !env.OK {
		return out, fmt.Errorf(
			"user-info 业务失败: status=%d ok=%v message=%s (%s)",
			env.Status, env.OK, env.MessageText, env.MessageCode,
		)
	}
	if env.Data == nil {
		return out, errors.New("user-info 响应缺少 data")
	}
	id := numToInt64(env.Data.HrID)
	if id == 0 {
		return out, errors.New("user-info 响应 data.hrId 无效或为空")
	}
	out.HrID = id
	if env.Data.ShiftInformationDTO != nil {
		out.ShiftNameZh = strings.TrimSpace(env.Data.ShiftInformationDTO.ShiftNameZh)
	}
	return out, nil
}

func fetchWorkHourPayload(ctx context.Context, client *http.Client, cookies map[string]string, hrID int64) (json.RawMessage, error) {
	url := envOr("WORK_HOUR_API_URL", config.GetWorkHour().APIURL)
	body := map[string]any{
		"hrId":     hrID,
		"locale":   "zh",
		"platform": "PC",
	}
	raw, code, err := postJSON(ctx, client, url, body, cookies)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("work-hour 接口 HTTP %d: %s", code, truncate(string(raw), 500))
	}
	return json.RawMessage(raw), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// extractRecords 将接口 JSON 规范为记录列表（与已删除的 sync_work_hour.py 一致）。
func extractRecords(payload json.RawMessage) ([]map[string]interface{}, error) {
	var top interface{}
	if err := json.Unmarshal(payload, &top); err != nil {
		return nil, err
	}
	switch v := top.(type) {
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out, nil
	case map[string]interface{}:
		if data, ok := v["data"]; ok {
			switch d := data.(type) {
			case []interface{}:
				out := make([]map[string]interface{}, 0, len(d))
				for _, item := range d {
					if m, ok := item.(map[string]interface{}); ok {
						out = append(out, m)
					}
				}
				return out, nil
			case map[string]interface{}:
				for _, key := range []string{"records", "list", "rows", "attendanceRecords", "items", "content"} {
					if inner, ok := d[key]; ok {
						if arr, ok := inner.([]interface{}); ok {
							out := make([]map[string]interface{}, 0, len(arr))
							for _, item := range arr {
								if m, ok := item.(map[string]interface{}); ok {
									out = append(out, m)
								}
							}
							return out, nil
						}
					}
				}
				if len(d) > 0 {
					return []map[string]interface{}{d}, nil
				}
			}
		}
		for _, key := range []string{"records", "list", "rows"} {
			if inner, ok := v[key]; ok {
				if arr, ok := inner.([]interface{}); ok {
					out := make([]map[string]interface{}, 0, len(arr))
					for _, item := range arr {
						if m, ok := item.(map[string]interface{}); ok {
							out = append(out, m)
						}
					}
					return out, nil
				}
			}
		}
	}
	return nil, nil
}

func recordFromMap(m map[string]interface{}) AttendanceRecord {
	return AttendanceRecord{
		ID:                       numToInt64(m["id"]),
		CreationDate:             strVal(m["creationDate"]),
		CreatedBy:                strVal(m["createdBy"]),
		LastUpdateDate:           strVal(m["lastUpdateDate"]),
		LastUpdatedBy:            strVal(m["lastUpdatedBy"]),
		OriginalID:               strVal(m["originalId"]),
		HrID:                     numToInt64(m["hrId"]),
		DataSource:               strVal(m["dataSource"]),
		ClockInReason:            strVal(m["clockInReason"]),
		AttendanceDate:           strVal(m["attendanceDate"]),
		ClockInDate:              strVal(m["clockInDate"]),
		ClockInTime:              strVal(m["clockInTime"]),
		DayID:                    strVal(m["dayId"]),
		ClockingInSequenceNumber: numToInt64(m["clockingInSequenceNumber"]),
		EarlyClockInTime:         strVal(m["earlyClockInTime"]),
		LateClockInTime:          strVal(m["lateClockInTime"]),
		ClockInType:              strVal(m["clockInType"]),
		EarlyClockInType:         strVal(m["earlyClockInType"]),
		LateClockInType:          strVal(m["lateClockInType"]),
		AttendanceStatus:         strVal(m["attendanceStatus"]),
		MinuteNumber:             strVal(m["minuteNumber"]),
		HourNumber:               strVal(m["hourNumber"]),
		AttendProcessID:          strVal(m["attendProcessId"]),
		WorkDay:                  strVal(m["workDay"]),
		AttendanceStatusCode:     strVal(m["attendanceStatusCode"]),
		EarlyClockInReason:       strVal(m["earlyClockInReason"]),
		LateClockInReason:        strVal(m["lateClockInReason"]),
		EarlyClockTag:            strVal(m["earlyClockTag"]),
		LateClockTag:             strVal(m["lateClockTag"]),
	}
}

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return fmt.Sprintf("%.0f", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(x)
	}
}

func numToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	case int:
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

// RefreshWorkHourData 拉取远程考勤 JSON、写入 SQLite，再查询返回列表。
func (a *App) RefreshWorkHourData() ([]AttendanceRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cookies, err := getCookiesViaChromedp(ctx)
	if err != nil {
		return nil, err
	}
	client := workHourHTTPClient()

	uid, err := fetchUserAccount(ctx, client, cookies)
	if err != nil {
		return nil, fmt.Errorf("获取 userAccount: %w", err)
	}
	uinfo, err := fetchWorkHourUserInfo(ctx, client, cookies, uid)
	if err != nil {
		return nil, fmt.Errorf("获取 user-info: %w", err)
	}
	a.setWorkHourShiftZh(uinfo.ShiftNameZh)
	raw, err := fetchWorkHourPayload(ctx, client, cookies, uinfo.HrID)
	if err != nil {
		return nil, fmt.Errorf("获取考勤数据: %w", err)
	}

	maps, err := extractRecords(raw)
	if err != nil {
		return nil, fmt.Errorf("解析考勤 JSON: %w", err)
	}
	if len(maps) == 0 {
		return a.GetWorkHourRecords()
	}
	recs := make([]AttendanceRecord, 0, len(maps))
	for _, m := range maps {
		recs = append(recs, recordFromMap(m))
	}
	if err := a.upsertAttendanceRecords(recs); err != nil {
		return nil, fmt.Errorf("写入数据库: %w", err)
	}
	return a.GetWorkHourRecords()
}
