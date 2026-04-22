package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"niumer/internal/config"
)

// 与 scripts/work-hour/get_work_hour.py 中 get_cookies 一致（无头 Chromium + 等待选择器）。
// 默认 URL / 选择器来自 configs/*.yaml（可用 WORK_HOUR_* 环境变量覆盖）。
func getCookiesViaChromedp(parentCtx context.Context) (map[string]string, error) {
	wh := config.GetWorkHour()
	loginURL := envOr("WORK_HOUR_LOGIN_URL", wh.LoginURL)
	sel := envOr("WORK_HOUR_WAIT_CSS", wh.WaitCSS)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	if p := strings.TrimSpace(os.Getenv("WORK_HOUR_CHROME_PATH")); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parentCtx, opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	if err := chromedp.Run(browserCtx, chromedp.Navigate(loginURL)); err != nil {
		return nil, fmt.Errorf("打开登录页: %w", err)
	}

	waitCtx, cancelWait := context.WithTimeout(browserCtx, 15*time.Second)
	defer cancelWait()
	// WaitReady avoids headless false negatives where the node is in the DOM but
	// WaitVisible never sees a stable box (common on minimal mock pages).
	if err := chromedp.Run(waitCtx, chromedp.WaitReady(sel, chromedp.ByQuery)); err != nil {
		return nil, fmt.Errorf("登录失败: 未在 15s 内就绪 %s（请开启免密登录）: %w", sel, err)
	}

	var cookies []*network.Cookie
	if err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = network.GetCookies().WithUrls([]string{loginURL}).Do(ctx)
		return err
	})); err != nil {
		return nil, fmt.Errorf("读取 Cookie: %w", err)
	}

	out := make(map[string]string)
	for _, c := range cookies {
		if c == nil {
			continue
		}
		out[c.Name] = c.Value
	}
	if len(out) == 0 {
		return nil, errors.New("浏览器未返回任何 Cookie")
	}
	return out, nil
}
