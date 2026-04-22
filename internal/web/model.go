package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zkteco-attshifts/internal/service"
)

func parseShowFrom(r *http.Request) map[string]bool {
	cols := r.URL.Query()["cols"]
	show := map[string]bool{}
	for _, c := range allColumns() {
		show[c.Key] = c.Default
	}
	if len(cols) > 0 {
		for k := range show {
			show[k] = false
		}
		for _, c := range cols {
			show[c] = true
		}
	}
	return show
}

func parseModeFrom(r *http.Request) string {
	s := r.URL.Query().Get("mode")
	if s == "" {
		s = "all"
	}
	return s
}

func buildModel(ctx context.Context, r *http.Request) (ReportModel, error) {
	now := time.Now()
	y := now.Year()
	m := int(now.Month())
	if v := r.URL.Query().Get("year"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			y = iv
		}
	}
	if v := r.URL.Query().Get("month"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil && iv >= 1 && iv <= 12 {
			m = iv
		}
	}
	deptParam := r.URL.Query().Get("dept")
	var deptIDPtr *int
	if deptParam != "" {
		if dv, err := strconv.Atoi(deptParam); err == nil && dv > 0 {
			deptIDPtr = &dv
		}
	}
	q := r.URL.Query().Get("q")
	show := parseShowFrom(r)
	mode := parseModeFrom(r)

	firstDay := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	dayCount := firstDay.AddDate(0, 1, -1).Day()

	setHolidays(ctx, firstDay, lastDay)

	users, err := service.QueryUsersFiltered(ctx, deptIDPtr, q)
	if err != nil {
		return ReportModel{}, err
	}
	att, err := service.QueryAtt(ctx, firstDay, lastDay)
	if err != nil {
		return ReportModel{}, err
	}

	daily := make(map[int]map[int]DayValue)
	reqPerDay := make(map[int]map[int]float64)
	sum := make(map[int]SumValue)
	for _, row := range att {
		uid := row.UserID
		d := row.AttDate.Day()
		if daily[uid] == nil {
			daily[uid] = make(map[int]DayValue)
		}
		if reqPerDay[uid] == nil {
			reqPerDay[uid] = make(map[int]float64)
		}
		workStr := formatFloat(row.Work)
		req := row.Required
		wd := row.AttDate.Weekday()
		isWeekend := wd == time.Saturday || wd == time.Sunday
		isH := isHoliday(row.AttDate)

		s := sum[uid]
		if !isWeekend && !isH && req > 0 {
			s.PresentDays += row.Work / req
			missing := req - row.Work
			if missing > 0.001 { // small epsilon
				s.AbsentDays += missing / req
				if row.Work == 0 {
					workStr = "旷"
				} else {
					workStr += "旷"
				}
			}
		}

		daily[uid][d] = DayValue{Work: workStr, Over: formatFloat(row.Over)}
		reqPerDay[uid][d] = req

		if row.Over > 0 {
			s.OverDays += 1
		}
		s.OverHours += row.Over
		s.LateMins += row.Late
		s.EarlyMins += row.Early
		s.NormalOT += row.NormalOT
		s.WeekendOT += row.WeekendOT
		s.HolidayOT += row.HolidayOT
		sum[uid] = s
	}

	leaves, _ := service.QueryLeaveSymbols(ctx, firstDay, lastDay)
	exceptionSymbols := map[int]string{
		1: "检",
		2: "病",
		3: "事",
		4: "产",
		5: "年",
	}

	for _, r2 := range leaves {
		val := extractFloat(r2.Symbol)
		uid := r2.UserID
		d := r2.AttDate.Day()
		s := sum[uid]
		days := 0.0
		req := 0.0
		if v, ok := reqPerDay[uid][d]; ok && v > 0 {
			req = v
			days = val / v
		} else if r2.Required > 0 {
			req = r2.Required
			days = val / r2.Required
		}

		// Update Display String
		dv := daily[uid][d]
		sym := exceptionSymbols[r2.ExceptionID]
		if sym == "" {
			sym = "假"
		}
		if strings.Contains(dv.Work, "旷") {
			// Replace "旷" or "X旷" with leave symbol
			if dv.Work == "旷" {
				dv.Work = sym
			} else {
				dv.Work = strings.Replace(dv.Work, "旷", sym, 1)
			}
		} else if dv.Work == "" {
			dv.Work = sym
		} else {
			dv.Work += sym
		}
		daily[uid][d] = dv

		// Update Sums
		s.LeaveHours += days
		s.LeaveHoursH += val
		wd := r2.AttDate.Weekday()
		isWeekend := wd == time.Saturday || wd == time.Sunday
		if !isWeekend && !isHoliday(r2.AttDate) && req > 0 {
			s.AbsentDays -= days
			if s.AbsentDays < 0 {
				s.AbsentDays = 0
			}
		}

		switch r2.ExceptionID {
		case 1:
			s.E1Business += days
		case 2:
			s.E2Sick += days
		case 3:
			s.E3Personal += days
		case 4:
			s.E4Home += days
		case 5:
			s.E5Annual += days
		}
		sum[uid] = s
	}
	var days []int
	for i := 1; i <= dayCount; i++ {
		days = append(days, i)
	}
	return ReportModel{Year: y, Month: m, Days: days, Users: users, Daily: daily, Sum: sum, Show: show, Mode: mode}, nil
}
