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
    // compute sticky left offsets for identity columns
    lefts := []int{}
    sum := 0
    defs := identityHeaderDefs()
    for _, d := range defs {
        lefts = append(lefts, sum)
        sum += d.Width
    }
    for i, d := range defs {
        style := fmt.Sprintf("min-width:%dpx;width:%dpx;position:sticky;left:%dpx;top:0;z-index:3;background:#f1f5f9", d.Width, d.Width, lefts[i])
        fmt.Fprintf(&b, "<th rowspan=\"2\" style=\"%s\">%s</th>", style, html.EscapeString(d.Title))
    }
    for _, day := range m.Days {
        wk := ""
        if weekend[day] {
            wk = "weekend"
        }
        fmt.Fprintf(&b, "<th class=\"%s\" colspan=\"2\" style=\"position:sticky;top:0;z-index:2;background:#f1f5f9\">%d<br><span class=\"wk\">%s</span></th>", wk, day, weekNames[day])
    }
    otherCols, overtimeCols, leaveCols := groupSumColumns(m)
    for _, c := range otherCols {
        fmt.Fprintf(&b, "<th class=\"sum-col\" rowspan=\"2\" style=\"position:sticky;top:0;z-index:2;background:#f1f5f9\">%s</th>", html.EscapeString(c.Title))
    }
    if len(overtimeCols) > 0 {
        fmt.Fprintf(&b, "<th class=\"sum-col\" colspan=\"%d\" style=\"position:sticky;top:0;z-index:2;background:#f1f5f9\">加班</th>", len(overtimeCols))
    }
    if len(leaveCols) > 0 {
        fmt.Fprintf(&b, "<th class=\"sum-col\" colspan=\"%d\" style=\"position:sticky;top:0;z-index:2;background:#f1f5f9\">请假</th>", len(leaveCols))
    }
    b.WriteString("</tr>\n")

    b.WriteString("<tr align=\"center\">\n")
    for range m.Days {
        b.WriteString("<th style=\"position:sticky;top:30px;z-index:2;background:#f1f5f9\">上</th><th style=\"position:sticky;top:30px;z-index:2;background:#f1f5f9\">加</th>")
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
        // sticky left identity cells
        for i, d := range defs {
            var val string
            if i == 0 { val = u.Badge } else if i == 1 { val = u.Name } else { val = u.DeptName }
            style := fmt.Sprintf("min-width:%dpx;width:%dpx;position:sticky;left:%dpx;z-index:1;background:#fff", d.Width, d.Width, lefts[i])
            fmt.Fprintf(&b, "<td style=\"%s\">%s</td>", style, html.EscapeString(val))
        }
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
