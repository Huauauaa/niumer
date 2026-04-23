package main

import (
	"context"
	"log"
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

	// 考勤：启动后 chromedp 写入 Cookie，再 tenant + user-info；进入工时页仅用 Cookie 调 workhour_url。
	muWorkHourAuth       sync.RWMutex
	workHourCookies      map[string]string
	workHourHrID         int64
	workHourUserAccount  string
	workHourBootstrapErr error
	workHourBootstrapDone chan struct{} // closed after first bootstrap attempt (success or fail)
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
