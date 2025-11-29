package web

func allColumns() []Column {
    return []Column{
        {Key: "present", Title: "出勤天数", SumField: "PresentDays", Value: func(s SumValue) string { return formatPresent(s.PresentDays) }, Default: true},
        {Key: "overhours", Title: "加班小时", SumField: "OverHours", Value: func(s SumValue) string { return formatFloat(s.OverHours) }, Default: false},
        {Key: "overdays", Title: "加班天数", SumField: "OverDays", Value: func(s SumValue) string { return format0f(s.OverDays) }, Default: true},
        {Key: "normalot", Title: "普通加班", SumField: "NormalOT", Value: func(s SumValue) string { return formatFloat(s.NormalOT) }, Default: true},
        {Key: "weekendot", Title: "周末加班", SumField: "WeekendOT", Value: func(s SumValue) string { return formatFloat(s.WeekendOT) }, Default: true},
        {Key: "holidayot", Title: "节日加班", SumField: "HolidayOT", Value: func(s SumValue) string { return formatFloat(s.HolidayOT) }, Default: true},
        {Key: "latemins", Title: "迟到分钟", SumField: "LateMins", Value: func(s SumValue) string { return format0f(s.LateMins) }, Default: true},
        {Key: "earlymins", Title: "早退分钟", SumField: "EarlyMins", Value: func(s SumValue) string { return format0f(s.EarlyMins) }, Default: true},
        {Key: "leavehours", Title: "请假小时", SumField: "LeaveHours", Value: func(s SumValue) string { return format2f(s.LeaveHours) }, Default: true},
        {Key: "e1", Title: "公出", SumField: "E1", Value: func(s SumValue) string { return format2f(s.E1Business) }, Default: true},
        {Key: "e2", Title: "病假", SumField: "E2", Value: func(s SumValue) string { return format2f(s.E2Sick) }, Default: true},
        {Key: "e3", Title: "事假", SumField: "E3", Value: func(s SumValue) string { return format2f(s.E3Personal) }, Default: true},
        {Key: "e4", Title: "探亲", SumField: "E4", Value: func(s SumValue) string { return format2f(s.E4Home) }, Default: true},
        {Key: "e5", Title: "年假", SumField: "E5", Value: func(s SumValue) string { return format2f(s.E5Annual) }, Default: true},
    }
}

func visibleColumns(m ReportModel) []Column {
    cols := []Column{}
    for _, c := range allColumns() {
        if m.Show[c.Key] {
            cols = append(cols, c)
        }
    }
    return cols
}
