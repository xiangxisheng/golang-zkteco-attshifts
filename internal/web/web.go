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
type SumValue struct {
	PresentDays float64
	OverHours   float64
	OverDays    float64
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
func format2f(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
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
			"PresentDays": format2f(v.PresentDays),
			"OverHours":   formatFloat(v.OverHours),
			"OverDays":    format2f(v.OverDays),
		}
	}
	return out
}

func RegisterRoutes() {
	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/download", handlerDownload)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("D:\\data\\code\\golang\\golang-zkteco-attshifts\\zkteco-attshifts\\wwwroot\\static"))))
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

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

	firstDay := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Local)
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
	sum := make(map[int]SumValue)
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
		s := sum[uid]
		if row.Required > 0 {
			s.PresentDays += row.Work / row.Required
			s.OverDays += row.Over / row.Required
		}
		s.OverHours += row.Over
		sum[uid] = s
	}

	tpl := `
    <!DOCTYPE html>
    <html>
    <head>
    <meta charset="utf-8">
    <title>考勤报表{{.Year}}-{{.Month}} - 富邦科技</title>
    <link rel="stylesheet" href="/static/main.css">
    <script src="/static/app.js" defer></script>
    </head>
    <body>
    <header class="topbar">
      <h1>考勤报表{{.Year}}-{{.Month}} - 富邦科技</h1>
      <form id="ym-form" method="get" class="ym-picker">
        <label>年份</label>
        <select name="year">
          {{range .Years}}
          <option value="{{.}}" {{if index $.SelYear .}}selected{{end}}>{{.}}</option>
          {{end}}
        </select>
        <label>月份</label>
        <select name="month">
          {{range .Months}}
          <option value="{{.}}" {{if index $.SelMonth .}}selected{{end}}>{{.}}</option>
          {{end}}
        </select>
        <button type="submit">切换</button>
      </form>
      <a class="download" href="/download?year={{.Year}}&month={{.Month}}">下载 CSV</a>
    </header>
    <main>
    <table class="grid">
    <tr align="center">
    <th style="min-width: 60px; width: 60px;">工号</th>
    <th style="min-width: 90px; width: 90px;">姓名</th>
    <th style="min-width: 90px; width: 90px;">部门</th>
    {{range .Days}}
    <th class="{{if index $.Weekend .}}weekend{{end}}">{{.}}<br><span class="wk">{{index $.WeekNames .}}</span><br>上</th>
    <th class="{{if index $.Weekend .}}weekend{{end}}">{{.}}<br><span class="wk">{{index $.WeekNames .}}</span><br>加</th>
    {{end}}
    <th class="sum-col"><span>出勤</span><br><span>天数</span></th>
    <th class="sum-col"><span>加班</span><br><span>小时</span></th>
    <th class="sum-col"><span>加班</span><br><span>天数</span></th>
    </tr>

    {{range .Users}}
    <tr align="center">
    <td>{{.Badge}}</td>
    <td>{{.Name}}</td>
    <td>{{.DeptName}}</td>
    {{ $uid := .UserID }}

    {{range $.Days}}
    {{ $d := . }}
    {{with index $.Data (printf "%d" $uid) (printf "%d" $d) }}
    <td class="work {{if index $.Weekend $d}}weekend{{end}} {{if .Work}}hasval{{else}}empty{{end}}" width="24">{{.Work}}</td>
    <td class="over {{if index $.Weekend $d}}weekend{{end}} {{if .Over}}hasval{{else}}empty{{end}}" width="24">{{.Over}}</td>
    {{end}}
    {{end}}
    {{with index $.SumStr (printf "%d" $uid) }}
    <td class="sum-col">{{.PresentDays}}</td>
    <td class="sum-col">{{.OverHours}}</td>
    <td class="sum-col">{{.OverDays}}</td>
    {{end}}
    </tr>
    {{end}}

    </table>
    </main>
    </body>
    </html>
    `

	t, _ := template.New("html").Parse(tpl)

	var days []int
	for i := 1; i <= dayCount; i++ {
		days = append(days, i)
	}
	var years []int
	for i := now.Year() - 5; i <= now.Year()+1; i++ {
		years = append(years, i)
	}
	var months []int
	for i := 1; i <= 12; i++ {
		months = append(months, i)
	}

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

	obj := map[string]any{
		"Year":      y,
		"Month":     m,
		"Days":      days,
		"Users":     users,
		"Data":      wrapData(data),
		"Sum":       wrapSum(sum),
		"SumStr":    wrapSumStr(sum),
		"Years":     years,
		"Months":    months,
		"SelYear":   map[int]bool{y: true},
		"SelMonth":  map[int]bool{m: true},
		"Weekend":   weekend,
		"WeekNames": weekNames,
	}

	t.Execute(w, obj)
}

func handlerDownload(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

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

	firstDay := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Local)
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
	row = append(row, "出勤天数", "加班小时", "加班天数")
	cw.Write(row)

	sum2 := make(map[int]SumValue)
	for _, row := range att {
		s := sum2[row.UserID]
		if row.Required > 0 {
			s.PresentDays += row.Work / row.Required
			s.OverDays += row.Over / row.Required
		}
		s.OverHours += row.Over
		sum2[row.UserID] = s
	}

	for _, u := range users {
		r := []string{u.DeptName, u.Badge, u.Name}
		for i := 1; i <= dayCount; i++ {
			v := data[u.UserID][i]
			r = append(r, v.Work+"/"+v.Over)
		}
		s := sum2[u.UserID]
		r = append(r, format2f(s.PresentDays), formatFloat(s.OverHours), format2f(s.OverDays))
		cw.Write(r)
	}
}
