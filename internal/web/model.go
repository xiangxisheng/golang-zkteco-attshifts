package web

import (
    "context"
    "net/http"
    "strconv"
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
    lastDay := firstDay.AddDate(0, 1, -1)
    dayCount := lastDay.Day()

    users, err := service.QueryUsersFiltered(ctx, deptIDPtr, q)
    if err != nil {
        return ReportModel{}, err
    }
    att, err := service.QueryAtt(ctx, firstDay, lastDay)
    if err != nil {
        return ReportModel{}, err
    }

    daily := make(map[int]map[int]DayValue)
    sum := make(map[int]SumValue)
    for _, row := range att {
        uid := row.UserID
        d := row.AttDate.Day()
        if daily[uid] == nil {
            daily[uid] = make(map[int]DayValue)
        }
        daily[uid][d] = DayValue{Work: formatFloat(row.Work), Over: formatFloat(row.Over)}
        s := sum[uid]
        if row.Required > 0 {
            s.PresentDays += row.Work / row.Required
        }
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
    for _, r2 := range leaves {
        val := extractFloat(r2.Symbol)
        s := sum[r2.UserID]
        s.LeaveHours += val
        switch r2.ExceptionID {
        case 1:
            s.E1Business += val
        case 2:
            s.E2Sick += val
        case 3:
            s.E3Personal += val
        case 4:
            s.E4Home += val
        case 5:
            s.E5Annual += val
        }
        sum[r2.UserID] = s
    }
    var days []int
    for i := 1; i <= dayCount; i++ {
        days = append(days, i)
    }
    return ReportModel{Year: y, Month: m, Days: days, Users: users, Daily: daily, Sum: sum, Show: show, Mode: mode}, nil
}
