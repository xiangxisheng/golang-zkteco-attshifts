package web

import (
    "fmt"
    "html"
    "strings"
    "time"
)

func computeWeekInfo(year int, month int) (map[int]bool, map[int]string) {
    firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
    lastDay := firstDay.AddDate(0, 1, -1)
    dayCount := lastDay.Day()
    weekend := map[int]bool{}
    weekNames := map[int]string{}
    names := []string{"日", "一", "二", "三", "四", "五", "六"}
    for i := 1; i <= dayCount; i++ {
        wd := firstDay.AddDate(0, 0, i-1).Weekday()
        if wd == time.Saturday || wd == time.Sunday {
            weekend[i] = true
        }
        weekNames[i] = names[int(wd)]
    }
    return weekend, weekNames
}

func renderGridTableHTML(m ReportModel, weekend map[int]bool, weekNames map[int]string) string {
    var b strings.Builder
    b.WriteString("<table class=\"grid\">\n")
    b.WriteString("<tr align=\"center\">\n")
    for _, d := range identityHeaderDefs() {
        if d.Style != "" {
            fmt.Fprintf(&b, "<th rowspan=\"2\" style=\"%s\">%s</th>", html.EscapeString(d.Style), html.EscapeString(d.Title))
        } else {
            fmt.Fprintf(&b, "<th rowspan=\"2\">%s</th>", html.EscapeString(d.Title))
        }
    }
    for _, day := range m.Days {
        wk := ""
        if weekend[day] {
            wk = "weekend"
        }
        fmt.Fprintf(&b, "<th class=\"%s\" colspan=\"2\">%d<br><span class=\"wk\">%s</span></th>", wk, day, weekNames[day])
    }
    otherCols, overtimeCols, leaveCols := groupSumColumns(m)
    for _, c := range otherCols {
        fmt.Fprintf(&b, "<th class=\"sum-col\" rowspan=\"2\">%s</th>", html.EscapeString(c.Title))
    }
    if len(overtimeCols) > 0 {
        fmt.Fprintf(&b, "<th class=\"sum-col\" colspan=\"%d\">加班</th>", len(overtimeCols))
    }
    if len(leaveCols) > 0 {
        fmt.Fprintf(&b, "<th class=\"sum-col\" colspan=\"%d\">请假</th>", len(leaveCols))
    }
    b.WriteString("</tr>\n")

    b.WriteString("<tr align=\"center\">\n")
    for range m.Days {
        b.WriteString("<th>上</th><th>加</th>")
    }
    for _, c := range overtimeCols {
        fmt.Fprintf(&b, "<th class=\"sum-col\">%s</th>", html.EscapeString(c.Title))
    }
    for _, c := range leaveCols {
        fmt.Fprintf(&b, "<th class=\"sum-col\">%s</th>", html.EscapeString(c.Title))
    }
    b.WriteString("</tr>\n")

    for _, u := range m.Users {
        b.WriteString("<tr align=\"center\">")
        fmt.Fprintf(&b, "<td>%s</td>", html.EscapeString(u.Badge))
        fmt.Fprintf(&b, "<td>%s</td>", html.EscapeString(u.Name))
        fmt.Fprintf(&b, "<td>%s</td>", html.EscapeString(u.DeptName))
        for _, d := range m.Days {
            v := m.Daily[u.UserID][d]
            wk := ""
            if weekend[d] {
                wk = "weekend"
            }
            clsW := "empty"
            if v.Work != "" {
                clsW = "hasval"
            }
            clsO := "empty"
            if v.Over != "" {
                clsO = "hasval"
            }
            fmt.Fprintf(&b, "<td class=\"work %s %s\" width=\"24\">%s</td>", wk, clsW, html.EscapeString(v.Work))
            fmt.Fprintf(&b, "<td class=\"over %s %s\" width=\"24\">%s</td>", wk, clsO, html.EscapeString(v.Over))
        }
        s := m.Sum[u.UserID]
        for _, c := range otherCols {
            fmt.Fprintf(&b, "<td class=\"sum-col\">%s</td>", html.EscapeString(c.Value(s)))
        }
        for _, c := range overtimeCols {
            fmt.Fprintf(&b, "<td class=\"sum-col\">%s</td>", html.EscapeString(c.Value(s)))
        }
        for _, c := range leaveCols {
            fmt.Fprintf(&b, "<td class=\"sum-col\">%s</td>", html.EscapeString(c.Value(s)))
        }
        b.WriteString("</tr>\n")
    }

    b.WriteString("</table>")
    return b.String()
}
