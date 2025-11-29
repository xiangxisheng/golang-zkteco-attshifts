package web

import (
    "encoding/csv"
    "fmt"
    "io"
    "net/http"
    "time"
)

func renderCSVModel(w http.ResponseWriter, m ReportModel) {
    w.Header().Set("Content-Type", "text/csv")
    ts := time.Now().Format("20060102_150405")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=att_%s.csv", ts))
    w.Write([]byte("\xEF\xBB\xBF"))
    cw := csv.NewWriter(w)
    defer cw.Flush()

    row := append([]string{}, identityHeaders()...)
    row = append(row, dailyHeaderTitles(m)...)
    for _, c := range orderedVisibleColumns(m) {
        row = append(row, c.Title)
    }
    cw.Write(row)

    for _, u := range m.Users {
        r := append([]string{}, []string{u.Badge, u.Name, u.DeptName}...)
        r = append(r, dailyRowValues(m, u.UserID)...)
        s := m.Sum[u.UserID]
        for _, c := range orderedVisibleColumns(m) {
            r = append(r, c.Value(s))
        }
        cw.Write(r)
    }
}

func renderXLSModel(w http.ResponseWriter, m ReportModel) {
    w.Header().Set("Content-Type", "application/vnd.ms-excel")
    ts := time.Now().Format("20060102_150405")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=att_%s.xls", ts))
    w.Write([]byte("\xEF\xBB\xBF"))
    fmt.Fprint(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>att</title></head><body>")
    fmt.Fprint(w, "<table border=1>")

    // header row 1: identity (rowspan=2), per-day (colspan=2), grouped sum (others rowspan, then overtime/leave colspan)
    fmt.Fprint(w, "<tr>")
    for _, h := range identityHeaders() {
        fmt.Fprintf(w, "<th rowspan=\"2\">%s</th>", h)
    }
    _, weekNames := computeWeekInfo(m.Year, m.Month)
    for _, d := range m.Days {
        fmt.Fprintf(w, "<th colspan=\"2\">%d<br><span class=\"wk\">%s</span></th>", d, weekNames[d])
    }
    otherCols, overtimeCols, leaveCols := orderedVisibleColumns(m)[:0], orderedVisibleColumns(m)[:0], orderedVisibleColumns(m)[:0]
    // recompute groups once
    otherCols, overtimeCols, leaveCols = groupSumColumns(m)
    for _, c := range otherCols {
        fmt.Fprintf(w, "<th rowspan=\"2\">%s</th>", c.Title)
    }
    if len(overtimeCols) > 0 {
        fmt.Fprintf(w, "<th colspan=\"%d\">加班</th>", len(overtimeCols))
    }
    if len(leaveCols) > 0 {
        fmt.Fprintf(w, "<th colspan=\"%d\">请假</th>", len(leaveCols))
    }
    fmt.Fprint(w, "</tr>")

    // header row 2: per-day subheaders and grouped sum subheaders
    fmt.Fprint(w, "<tr>")
    for range m.Days {
        fmt.Fprint(w, "<th>上</th><th>加</th>")
    }
    for _, c := range overtimeCols {
        fmt.Fprintf(w, "<th>%s</th>", c.Title)
    }
    for _, c := range leaveCols {
        fmt.Fprintf(w, "<th>%s</th>", c.Title)
    }
    fmt.Fprint(w, "</tr>")

    // data rows
    for _, u := range m.Users {
        fmt.Fprint(w, "<tr>")
        fmt.Fprintf(w, "<td>%s</td>", u.Badge)
        fmt.Fprintf(w, "<td>%s</td>", u.Name)
        fmt.Fprintf(w, "<td>%s</td>", u.DeptName)
        for _, d := range m.Days {
            v := m.Daily[u.UserID][d]
            fmt.Fprintf(w, "<td>%s</td>", v.Work)
            fmt.Fprintf(w, "<td>%s</td>", v.Over)
        }
        s := m.Sum[u.UserID]
        for _, c := range otherCols {
            fmt.Fprintf(w, "<td>%s</td>", c.Value(s))
        }
        for _, c := range overtimeCols {
            fmt.Fprintf(w, "<td>%s</td>", c.Value(s))
        }
        for _, c := range leaveCols {
            fmt.Fprintf(w, "<td>%s</td>", c.Value(s))
        }
        fmt.Fprint(w, "</tr>")
    }
    fmt.Fprint(w, "</table></body></html>")
}

func renderHTMLModel(w http.ResponseWriter, m ReportModel) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    ts := time.Now().Format("20060102_150405")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=att_%s.html", ts))
    io.WriteString(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>att</title><style>table{border-collapse:collapse}td,th{border:1px solid #999;padding:4px;font-size:12px}th{background:#f1f5f9}tr:nth-child(even){background:#f9fafb}td{text-align:center}</style></head><body>")
    weekend, weekNames := computeWeekInfo(m.Year, m.Month)
    io.WriteString(w, renderGridTableHTML(m, weekend, weekNames))
    io.WriteString(w, "</body></html>")
}
