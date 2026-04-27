package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// App is the Wails-bound application object.
type App struct {
	ctx context.Context

	muWorkHourShift sync.RWMutex
	// workHourShiftZh is the last data.shiftInformationDTO.shiftNameZh from POST /user-info.
	workHourShiftZh string
	// workHourEffWindows derived from shiftNameZh (Work / Rest); nil = use hardcoded default in compute.
	workHourEffWindows []workHourTimeWindow

	// 1）启动：chromedp 取 Cookie 全局缓存；2）若 SQLite 已有 userAccount+hrId 则仅从库恢复用户态，否则 tenant + /user-info 后落库；
	// 3）各接口用 workHourPostJSON/workHourGet，遇 403 则重取 Cookie 再发一次。PR 列表 RefreshPullRequestList、总数 GetPullRequestListTotal；GET 与考勤同 Cookie。工时 POST 见 workHourPostJSON。
	muWorkHourAuth        sync.RWMutex
	muWorkHourCookieCrawl sync.Mutex // 串行 chromedp 取 Cookie，避免与 403 重登入并发
	workHourCookies       map[string]string
	workHourHrID          int64
	workHourUserAccount   string
	workHourBootstrapErr  error
	workHourBootstrapDone chan struct{} // closed after first bootstrap attempt (success or fail)

	muTerminal   sync.Mutex
	terminalFile *os.File  // PTY master (unix only)
	terminalCmd  *exec.Cmd // shell process
}

func NewApp() *App {
	return &App{
		workHourBootstrapDone: make(chan struct{}),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// On first launch, persist the default blog directory so Markdown always has an
	// explicit on-disk location and Preferences shows the resolved path.
	if err := a.ensureDefaultBlogWorkDirInConfig(); err != nil {
		log.Printf("niumer: default blog directory: %v", err)
	}
	if err := a.ensureDefaultJsonFormatterWorkDirInConfig(); err != nil {
		log.Printf("niumer: default JSON formatter directory: %v", err)
	}
	go func() {
		defer close(a.workHourBootstrapDone)
		bctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
		defer cancel()
		if err := a.bootstrapWorkHourSession(bctx); err != nil {
			log.Printf("niumer: work hour bootstrap (login / tenant / user-info): %v", err)
		}
	}()
}
