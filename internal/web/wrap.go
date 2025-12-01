package web

import "strconv"

func wrapData(data map[int]map[int]DayValue) map[string]map[string]DayValue {
    out := make(map[string]map[string]DayValue)
    for uid, m := range data {
        uidStr := strconv.Itoa(uid)
        out[uidStr] = make(map[string]DayValue)
        for day, v := range m {
            out[uidStr][strconv.Itoa(day)] = v
        }
    }
    return out
}

func wrapSum(data map[int]SumValue) map[string]SumValue {
    out := make(map[string]SumValue)
    for uid, v := range data {
        out[strconv.Itoa(uid)] = v
    }
    return out
}

func wrapSumStr(data map[int]SumValue) map[string]map[string]string {
    out := make(map[string]map[string]string)
    for uid, v := range data {
        out[strconv.Itoa(uid)] = map[string]string{
            "PresentDays": formatPresent(v.PresentDays),
            "OverHours":   formatFloat(v.OverHours),
            "OverDays":    format0f(v.OverDays),
            "LateMins":    format0f(v.LateMins),
            "EarlyMins":   format0f(v.EarlyMins),
            "LeaveHours":  formatPresent(v.LeaveHours),
            "LeaveHoursH": formatFloat(v.LeaveHoursH),
            "NormalOT":    formatFloat(v.NormalOT),
            "WeekendOT":   formatFloat(v.WeekendOT),
            "HolidayOT":   formatFloat(v.HolidayOT),
            "E1":          formatPresent(v.E1Business),
            "E2":          formatPresent(v.E2Sick),
            "E3":          formatPresent(v.E3Personal),
            "E4":          formatPresent(v.E4Home),
            "E5":          formatPresent(v.E5Annual),
        }
    }
    return out
}
