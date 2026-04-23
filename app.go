package main

import (
	"context"
	"log"
	"sync"
)

// App is the Wails-bound application object.
type App struct {
	ctx context.Context

	muWorkHourShift sync.RWMutex
	// workHourShiftZh is the last data.shiftInformationDTO.shiftNameZh from POST /user-info.
	workHourShiftZh string
	// workHourEffWindows derived from shiftNameZh (Work / Rest); nil = use hardcoded default in compute.
	workHourEffWindows []workHourTimeWindow
}

func NewApp() *App {
	return &App{}
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
}
