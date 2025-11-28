package web

import (
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"zkteco-attshifts/internal/service"
)

type DayValue struct {
	Work string
	Over string
}

func formatFloat(f float64) string {
	if f == 0 {
		return ""
	}
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

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

func RegisterRoutes() {
	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/download", handlerDownload)
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)
	dayCount := lastDay.Day()

	users, err := service.QueryUsers(ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	att, err := service.QueryAtt(ctx, firstDay, lastDay)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := make(map[int]map[int]DayValue)
	for _, row := range att {
		uid := row.UserID
		day := row.AttDate.Day()
		if data[uid] == nil {
			data[uid] = make(map[int]DayValue)
		}
		data[uid][day] = DayValue{
			Work: formatFloat(row.Work),
			Over: formatFloat(row.Over),
		}
	}

	tpl := `
    <!DOCTYPE html>
    <html>
    <head>
    <meta charset="utf-8">
    <title>考勤报表</title>
    <style>
    table { border-collapse: collapse; }
    td,th { border:1px solid #555; padding: 1px; font-size: 12px; }
    </style>
    </head>
    <body>
    <h2>{{.Year}} 年 {{.Month}} 月考勤</h2>
    <a href="/download">下载 Excel</a>
    <br><br>
    <table>
    <tr align="center">
    <th width="60">工号</th>
    <th width="80">姓名</th>
    <th width="60">部门</th>
    {{range .Days}}
    <th>{{.}}<br>上</th>
    <th>{{.}}<br>加</th>
    {{end}}
    </tr>

    {{range .Users}}
    <tr align="center">
    <td>{{.Badge}}</td>
    <td>{{.Name}}</td>
    <td>{{.DeptName}}</td>
    {{ $uid := .UserID }}

    {{range $.Days}}
    {{with index $.Data (printf "%d" $uid) (printf "%d" .) }}
    <td width="20">{{.Work}}</td>
    <td width="20">{{.Over}}</td>
    {{end}}
    {{end}}
    </tr>
    {{end}}

    </table>
    </body>
    </html>
    `

	t, _ := template.New("html").Parse(tpl)

	var days []int
	for i := 1; i <= dayCount; i++ {
		days = append(days, i)
	}

	obj := map[string]any{
		"Year":  now.Year(),
		"Month": int(now.Month()),
		"Days":  days,
		"Users": users,
		"Data":  wrapData(data),
	}

	t.Execute(w, obj)
}

func handlerDownload(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)
	dayCount := lastDay.Day()

	users, _ := service.QueryUsers(ctx)
	att, _ := service.QueryAtt(ctx, firstDay, lastDay)

	data := make(map[int]map[int]DayValue)
	for _, row := range att {
		uid := row.UserID
		day := row.AttDate.Day()
		if data[uid] == nil {
			data[uid] = make(map[int]DayValue)
		}
		data[uid][day] = DayValue{
			Work: formatFloat(row.Work),
			Over: formatFloat(row.Over),
		}
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=att.csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	row := []string{"部门", "工号", "姓名"}
	for i := 1; i <= dayCount; i++ {
		row = append(row, fmt.Sprintf("%d号上班/加班", i))
	}
	cw.Write(row)

	for _, u := range users {
		r := []string{u.DeptName, u.Badge, u.Name}
		for i := 1; i <= dayCount; i++ {
			v := data[u.UserID][i]
			r = append(r, v.Work+"/"+v.Over)
		}
		cw.Write(r)
	}
}
