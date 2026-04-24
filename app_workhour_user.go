package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// GetWorkHourUserProfile 从本地 SQLite 读取已缓存的租户 / 用户信息（含 user-info 的 data JSON）。
func (a *App) GetWorkHourUserProfile() (WorkHourUserProfileView, error) {
	if a == nil {
		return WorkHourUserProfileView{}, errors.New("app is nil")
	}
	return a.readWorkHourUserProfileView()
}

// RefreshWorkHourUserInfo 重新从 tenant + user-info 拉取并写入 SQLite 与进程内状态，再返回同 GetWorkHourUserProfile 的结构。
func (a *App) RefreshWorkHourUserInfo() (WorkHourUserProfileView, error) {
	if a == nil {
		return WorkHourUserProfileView{}, errors.New("app is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	select {
	case <-a.workHourBootstrapDone:
	case <-ctx.Done():
		return WorkHourUserProfileView{}, errors.New("等待首次会话初始化超时")
	}

	a.muWorkHourAuth.RLock()
	hasCookies := len(a.workHourCookies) > 0
	bootErr := a.workHourBootstrapErr
	a.muWorkHourAuth.RUnlock()
	if !hasCookies {
		if bootErr != nil {
			return WorkHourUserProfileView{}, fmt.Errorf("无有效 Cookie: %w", bootErr)
		}
		if err := a.refreshWorkHourCookiesFromBrowser(ctx); err != nil {
			return WorkHourUserProfileView{}, fmt.Errorf("重取登录 Cookie: %w", err)
		}
	}

	client := workHourHTTPClient()
	uid, err := a.fetchUserAccountForSession(ctx, client)
	if err != nil {
		return WorkHourUserProfileView{}, err
	}
	uinfo, dataJSON, err := a.fetchWorkHourUserInfoForSession(ctx, client, uid)
	if err != nil {
		return WorkHourUserProfileView{}, err
	}
	a.setWorkHourShiftZh(uinfo.ShiftNameZh)
	if upErr := a.upsertWorkHourUserProfile(uid, uinfo.HrID, uinfo.ShiftNameZh, dataJSON); upErr != nil {
		return WorkHourUserProfileView{}, upErr
	}
	a.muWorkHourAuth.Lock()
	a.workHourHrID = uinfo.HrID
	a.workHourUserAccount = strings.TrimSpace(uid)
	a.muWorkHourAuth.Unlock()
	return a.readWorkHourUserProfileView()
}
