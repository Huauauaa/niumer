package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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

func getWithCookies(ctx context.Context, client *http.Client, u string, cookies map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
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

// userInfoEnvelope: Data 用 RawMessage 保留 /user-info 的完整 `data`（写入 SQLite 与 UI），
// 避免窄 struct Unmarshal 丢字段导致「从服务器刷新」后仍只见 hrId+shiftNameZh 片段。
type userInfoEnvelope struct {
	Status      int             `json:"status"`
	MessageCode string          `json:"messageCode"`
	MessageText string          `json:"messageText"`
	OK          bool            `json:"ok"`
	Data        json.RawMessage `json:"data"`
}

type workHourUserInfo struct {
	HrID        int64
	ShiftNameZh string
}

func (a *App) fetchUserAccountForSession(ctx context.Context, client *http.Client) (string, error) {
	url := envOr("WORK_HOUR_TENANT_URL", config.GetWorkHour().TenantURL)
	body := map[string]string{
		"rentId":     "HuaWei",
		"groupId":    "servicetimeflow",
		"tenantId":   "",
		"url":        "hr.huawei.com",
		"parentCode": "servicetimeflow",
	}
	raw, code, err := a.workHourPostJSON(ctx, client, url, body)
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

// userInfoDataJSON 为 /user-info 的 data 段 JSON，写入 SQLite 供「用户信息」与离线记忆。
func (a *App) fetchWorkHourUserInfoForSession(ctx context.Context, client *http.Client, userAccount string) (workHourUserInfo, string, error) {
	var out workHourUserInfo
	var outJSON string
	if len(userAccount) < 2 {
		return out, outJSON, errors.New("userAccount 过短")
	}
	url := envOr("WORK_HOUR_USER_INFO_URL", config.GetWorkHour().UserInfoURL)
	body := map[string]string{
		"employeeQuery": userAccount[1:],
		"queryDate":     time.Now().Format("2006-01-02"),
		"locale":        "zh",
		"platform":      "PC",
	}
	raw, code, err := a.workHourPostJSON(ctx, client, url, body)
	if err != nil {
		return out, outJSON, err
	}
	if code < 200 || code >= 300 {
		return out, outJSON, fmt.Errorf("user-info 接口 HTTP %d: %s", code, truncate(string(raw), 500))
	}
	var env userInfoEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return out, outJSON, err
	}
	if env.Status != 200 || !env.OK {
		return out, outJSON, fmt.Errorf(
			"user-info 业务失败: status=%d ok=%v message=%s (%s)",
			env.Status, env.OK, env.MessageText, env.MessageCode,
		)
	}
	if len(bytes.TrimSpace(env.Data)) == 0 {
		return out, outJSON, errors.New("user-info 响应缺少 data")
	}
	outJSON = string(env.Data)

	// 仅从 data 中解出业务必要字段
	var dataWire struct {
		HrID                interface{} `json:"hrId"`
		ShiftInformationDTO *struct {
			ShiftNameZh string `json:"shiftNameZh"`
		} `json:"shiftInformationDTO"`
	}
	if err := json.Unmarshal(env.Data, &dataWire); err != nil {
		return out, outJSON, fmt.Errorf("解析 user-info data: %w", err)
	}
	id := numToInt64(dataWire.HrID)
	if id == 0 {
		return out, outJSON, errors.New("user-info 响应 data.hrId 无效或为空")
	}
	out.HrID = id
	if dataWire.ShiftInformationDTO != nil {
		out.ShiftNameZh = strings.TrimSpace(dataWire.ShiftInformationDTO.ShiftNameZh)
	}
	return out, outJSON, nil
}

func (a *App) fetchWorkHourPayloadForSession(ctx context.Context, client *http.Client, hrID int64) (json.RawMessage, error) {
	url := envOr("WORK_HOUR_WORKHOUR_URL", config.GetWorkHour().WorkHourURL)
	body := map[string]any{
		"hrId":     hrID,
		"locale":   "zh",
		"platform": "PC",
	}
	raw, code, err := a.workHourPostJSON(ctx, client, url, body)
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

func cloneCookieMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (a *App) workHourCookiesForHTTP() map[string]string {
	if a == nil {
		return nil
	}
	a.muWorkHourAuth.RLock()
	defer a.muWorkHourAuth.RUnlock()
	return cloneCookieMap(a.workHourCookies)
}

// bootstrapWorkHourSession：1）无头登录取 Cookie 到全局 2）若 SQLite 已缓存用户，则只从库恢复内存 3）否则再请求 tenant + user-info 并写入 SQLite 与内存。
func (a *App) bootstrapWorkHourSession(ctx context.Context) (err error) {
	if a == nil {
		return errors.New("app is nil")
	}
	defer func() {
		a.muWorkHourAuth.Lock()
		a.workHourBootstrapErr = err
		a.muWorkHourAuth.Unlock()
	}()

	if err = a.refreshWorkHourCookiesFromBrowser(ctx); err != nil {
		return err
	}

	if ok, qerr := a.hasWorkHourUserProfileInDB(); qerr == nil && ok {
		if loadErr := a.loadWorkHourUserFromDBIntoMemory(); loadErr == nil {
			return nil
		} else {
			log.Printf("niumer: load user profile from sqlite: %v, will re-fetch from network", loadErr)
		}
	} else if qerr != nil {
		log.Printf("niumer: check workhour_user_profile: %v", qerr)
	}

	client := workHourHTTPClient()
	uid, nerr := a.fetchUserAccountForSession(ctx, client)
	if nerr != nil {
		return fmt.Errorf("获取 userAccount: %w", nerr)
	}
	uinfo, dataJSON, nerr2 := a.fetchWorkHourUserInfoForSession(ctx, client, uid)
	if nerr2 != nil {
		return fmt.Errorf("获取 user-info: %w", nerr2)
	}
	a.setWorkHourShiftZh(uinfo.ShiftNameZh)
	if upErr := a.upsertWorkHourUserProfile(uid, uinfo.HrID, uinfo.ShiftNameZh, dataJSON); upErr != nil {
		log.Printf("niumer: persist workhour_user_profile: %v", upErr)
	}
	a.muWorkHourAuth.Lock()
	a.workHourHrID = uinfo.HrID
	a.workHourUserAccount = uid
	a.muWorkHourAuth.Unlock()
	return nil
}

// ensureWorkHourUserAndHRID 确保内存中已有 hrId / 用户账号；优先从 SQLite 恢复，否则在已有 Cookie 时请求 tenant + user-info 并落库。
func (a *App) ensureWorkHourUserAndHRID(ctx context.Context) error {
	if a == nil {
		return errors.New("app is nil")
	}
	a.muWorkHourAuth.RLock()
	hrID := a.workHourHrID
	ua := strings.TrimSpace(a.workHourUserAccount)
	bootErr := a.workHourBootstrapErr
	cookieCount := len(a.workHourCookies)
	a.muWorkHourAuth.RUnlock()
	if hrID != 0 && ua != "" {
		return nil
	}
	if ok, _ := a.hasWorkHourUserProfileInDB(); ok {
		if loadErr := a.loadWorkHourUserFromDBIntoMemory(); loadErr == nil {
			return nil
		}
	}
	if cookieCount == 0 {
		if bootErr != nil {
			return fmt.Errorf("无有效考勤 Cookie: %w", bootErr)
		}
		return errors.New("无有效考勤 Cookie（请确认登录页与 WORK_HOUR_WAIT_CSS）")
	}
	client := workHourHTTPClient()
	uid, err := a.fetchUserAccountForSession(ctx, client)
	if err != nil {
		return fmt.Errorf("重新获取 userAccount: %w", err)
	}
	uinfo, dataJSON, err := a.fetchWorkHourUserInfoForSession(ctx, client, uid)
	if err != nil {
		return fmt.Errorf("重新获取 user-info: %w", err)
	}
	a.setWorkHourShiftZh(uinfo.ShiftNameZh)
	if upErr := a.upsertWorkHourUserProfile(uid, uinfo.HrID, uinfo.ShiftNameZh, dataJSON); upErr != nil {
		log.Printf("niumer: persist workhour_user_profile: %v", upErr)
	}
	a.muWorkHourAuth.Lock()
	a.workHourHrID = uinfo.HrID
	a.workHourUserAccount = uid
	a.muWorkHourAuth.Unlock()
	return nil
}

// RefreshWorkHourData 使用启动阶段缓存的 Cookie，仅请求 workhour_url 拉取考勤 JSON、写入 SQLite，再查询返回列表。
func (a *App) RefreshWorkHourData() ([]AttendanceRecord, error) {
	if a == nil {
		return nil, errors.New("app is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	select {
	case <-a.workHourBootstrapDone:
	case <-ctx.Done():
		return nil, errors.New("等待考勤登录阶段结束超时")
	}

	if err := a.ensureWorkHourUserAndHRID(ctx); err != nil {
		return nil, err
	}

	a.muWorkHourAuth.RLock()
	hrID := a.workHourHrID
	a.muWorkHourAuth.RUnlock()
	if hrID == 0 {
		return nil, errors.New("缺少 hrId，无法拉取考勤明细")
	}

	client := workHourHTTPClient()
	raw, err := a.fetchWorkHourPayloadForSession(ctx, client, hrID)
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
