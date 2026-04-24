package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// refreshWorkHourCookiesFromBrowser 用 chromedp 重登并缓存 Cookie（与启动阶段一致，供 403 重试等调用）。
func (a *App) refreshWorkHourCookiesFromBrowser(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("app is nil")
	}
	a.muWorkHourCookieCrawl.Lock()
	defer a.muWorkHourCookieCrawl.Unlock()
	cookies, err := getCookiesViaChromedp(ctx)
	if err != nil {
		return err
	}
	a.muWorkHourAuth.Lock()
	a.workHourCookies = cloneCookieMap(cookies)
	a.muWorkHourAuth.Unlock()
	return nil
}

// workHourPostJSON 使用当前缓存 Cookie 发 POST；若 HTTP 403 则重抓 Cookie 后重试一次（其它接口可共用）。
func (a *App) workHourPostJSON(ctx context.Context, client *http.Client, u string, body any) ([]byte, int, error) {
	ck := a.workHourCookiesForHTTP()
	b, code, err := postJSON(ctx, client, u, body, ck)
	if err != nil {
		return b, code, err
	}
	if code != 403 {
		return b, code, nil
	}
	if rerr := a.refreshWorkHourCookiesFromBrowser(ctx); rerr != nil {
		return b, code, fmt.Errorf("HTTP 403 后重取 Cookie: %w", rerr)
	}
	ck2 := a.workHourCookiesForHTTP()
	return postJSON(ctx, client, u, body, ck2)
}

// workHourGet 使用当前缓存 Cookie 发 GET；若 HTTP 403 则重抓 Cookie 后重试一次。
func (a *App) workHourGet(ctx context.Context, client *http.Client, u string) ([]byte, int, error) {
	ck := a.workHourCookiesForHTTP()
	b, code, err := getWithCookies(ctx, client, u, ck)
	if err != nil {
		return b, code, err
	}
	if code != 403 {
		return b, code, nil
	}
	if rerr := a.refreshWorkHourCookiesFromBrowser(ctx); rerr != nil {
		return b, code, fmt.Errorf("HTTP 403 后重取 Cookie: %w", rerr)
	}
	ck2 := a.workHourCookiesForHTTP()
	return getWithCookies(ctx, client, u, ck2)
}

// loadWorkHourUserFromDBIntoMemory 在 SQLite 已有 userAccount/ hrId/ 班次 时，仅写回内存，不请求网络。
func (a *App) loadWorkHourUserFromDBIntoMemory() error {
	if a == nil {
		return fmt.Errorf("app is nil")
	}
	p, err := a.readWorkHourUserProfileView()
	if err != nil {
		return err
	}
	if p.HrID == 0 || strings.TrimSpace(p.UserAccount) == "" {
		return fmt.Errorf("SQLite 中无有效用户概要")
	}
	a.setWorkHourShiftZh(p.ShiftNameZh)
	a.muWorkHourAuth.Lock()
	a.workHourHrID = p.HrID
	a.workHourUserAccount = p.UserAccount
	a.muWorkHourAuth.Unlock()
	return nil
}
