package main

// GetWorkHourShiftSchedule returns the last shiftNameZh from a successful /user-info call
// during startup bootstrap or RefreshWorkHourData recovery (empty until then).
func (a *App) GetWorkHourShiftSchedule() string {
	if a == nil {
		return ""
	}
	a.muWorkHourShift.RLock()
	defer a.muWorkHourShift.RUnlock()
	return a.workHourShiftZh
}

func (a *App) setWorkHourShiftZh(s string) {
	if a == nil {
		return
	}
	a.muWorkHourShift.Lock()
	defer a.muWorkHourShift.Unlock()
	a.workHourShiftZh = s
	if wins, ok := parseWorkHourWindowsFromShiftNameZh(s); ok && len(wins) > 0 {
		a.workHourEffWindows = wins
	} else {
		a.workHourEffWindows = nil
	}
}

// workHourEffectiveWindows returns counting windows for effective hours (shiftNameZh or default).
func (a *App) workHourEffectiveWindows() []workHourTimeWindow {
	if a == nil {
		return defaultWorkHourTimeWindows()
	}
	a.muWorkHourShift.RLock()
	defer a.muWorkHourShift.RUnlock()
	if len(a.workHourEffWindows) > 0 {
		out := make([]workHourTimeWindow, len(a.workHourEffWindows))
		copy(out, a.workHourEffWindows)
		return out
	}
	return defaultWorkHourTimeWindows()
}
