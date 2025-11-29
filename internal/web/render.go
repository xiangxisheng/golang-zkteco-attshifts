package web

import (
    "encoding/csv"
    "fmt"
    "io"
    "net/http"
)

func renderCSVModel(w http.ResponseWriter, m ReportModel) {
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename=att.csv")
    w.Write([]byte("\xEF\xBB\xBF"))
    cw := csv.NewWriter(w)
    defer cw.Flush()

    row := append([]string{}, identityHeaders()...)
    row = append(row, dailyHeaderTitles(m)...)
    for _, c := range visibleColumns(m) {
        row = append(row, c.Title)
    }
    cw.Write(row)

    for _, u := range m.Users {
        r := append([]string{}, []string{u.Badge, u.Name, u.DeptName}...)
        r = append(r, dailyRowValues(m, u.UserID)...)
        s := m.Sum[u.UserID]
        for _, c := range visibleColumns(m) {
            r = append(r, c.Value(s))
        }
        cw.Write(r)
    }
}

func renderXLSModel(w http.ResponseWriter, m ReportModel) {
    w.Header().Set("Content-Type", "application/vnd.ms-excel")
    w.Header().Set("Content-Disposition", "attachment; filename=att.xls")
    w.Write([]byte("\xEF\xBB\xBF"))
    fmt.Fprint(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>att</title></head><body>")
    fmt.Fprint(w, "<table border=1>")
    fmt.Fprint(w, "<tr>")
    for _, h := range identityHeaders() {
        fmt.Fprintf(w, "<th>%s</th>", h)
    }
    for _, h := range dailyHeaderTitles(m) {
        fmt.Fprintf(w, "<th>%s</th>", h)
    }
    for _, c := range visibleColumns(m) {
        fmt.Fprintf(w, "<th>%s</th>", c.Title)
    }
    fmt.Fprint(w, "</tr>")

    for _, u := range m.Users {
        fmt.Fprint(w, "<tr>")
        fmt.Fprintf(w, "<td>%s</td>", u.Badge)
        fmt.Fprintf(w, "<td>%s</td>", u.Name)
        fmt.Fprintf(w, "<td>%s</td>", u.DeptName)
        for _, v := range dailyRowValues(m, u.UserID) {
            fmt.Fprintf(w, "<td>%s</td>", v)
        }
        s := m.Sum[u.UserID]
        for _, c := range visibleColumns(m) {
            fmt.Fprintf(w, "<td>%s</td>", c.Value(s))
        }
        fmt.Fprint(w, "</tr>")
    }
    fmt.Fprint(w, "</table></body></html>")
}

func renderHTMLModel(w http.ResponseWriter, m ReportModel) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Header().Set("Content-Disposition", "attachment; filename=att.html")
    io.WriteString(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>att</title><style>table{border-collapse:collapse}td,th{border:1px solid #999;padding:4px;font-size:12px}th{background:#f1f5f9}tr:nth-child(even){background:#f9fafb}td{text-align:center}</style></head><body>")
    weekend, weekNames := computeWeekInfo(m.Year, m.Month)
    io.WriteString(w, renderGridTableHTML(m, weekend, weekNames))
    io.WriteString(w, "</body></html>")
}

