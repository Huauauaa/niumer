package main

import (
	"runtime"
	"runtime/debug"
)

// AboutView is shown in the About dialog (Cursor-style build metadata).
type AboutView struct {
	AppName      string `json:"appName"`
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	BuildTime    string `json:"buildTime"`
	GoVersion    string `json:"goVersion"`
	WailsVersion string `json:"wailsVersion"`
	OSArch       string `json:"osArch"`
}

// GetAboutInfo returns version and build metadata from the Go binary (debug.ReadBuildInfo).
func (a *App) GetAboutInfo() AboutView {
	out := AboutView{
		AppName:      "niumer",
		Version:      "0.0.0",
		WailsVersion: "v2.12.0",
		OSArch:       runtime.GOOS + " " + runtime.GOARCH,
		GoVersion:    runtime.Version(),
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			out.Version = v
		}
		if bi.GoVersion != "" {
			out.GoVersion = bi.GoVersion
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				out.Commit = s.Value
			case "vcs.time":
				out.BuildTime = s.Value
			}
		}
	}
	return out
}
