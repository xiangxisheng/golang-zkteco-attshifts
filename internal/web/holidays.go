package web

import (
	"context"
	"zkteco-attshifts/internal/service"
	"time"
)

var holidaySet = map[string]bool{}

func setHolidays(ctx context.Context, start, end time.Time) {
	holidaySet = map[string]bool{}
	rows, err := service.QueryHolidays(ctx, start, end)
	if err != nil {
		return
	}
	for _, h := range rows {
		days := h.Duration
		if days <= 0 {
			days = 1
		}
		for i := 0; i < days; i++ {
			d := h.StartTime.AddDate(0, 0, i)
			if d.Before(start) || d.After(end) {
				continue
			}
			holidaySet[d.Format("2006-01-02")] = true
		}
	}
}

func isHoliday(t time.Time) bool {
	return holidaySet[t.Format("2006-01-02")]
}

