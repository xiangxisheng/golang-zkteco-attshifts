package web

import (
    "context"
    "encoding/csv"
    "fmt"
    "html/template"
    "net/http"
    "os"
    "path/filepath"
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
	LateMins    float64
	EarlyMins   float64
	LeaveHours  float64
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
func format0f(f float64) string {
	return strconv.FormatFloat(f, 'f', 0, 64)
}
func extractFloat(s string) float64 {
	buf := []rune{}
	started := false
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			buf = append(buf, r)
			started = true
		} else if started {
			break
		}
	}
	if len(buf) == 0 {
		return 0
	}
	v, err := strconv.ParseFloat(string(buf), 64)
	if err != nil {
		return 0
	}
	return v
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
			"OverDays":    format0f(v.OverDays),
			"LateMins":    format0f(v.LateMins),
			"EarlyMins":   format0f(v.EarlyMins),
			"LeaveHours":  format0f(v.LeaveHours),
		}
	}
	return out
}

func RegisterRoutes() {
    http.HandleFunc("/", handlerIndex)
    http.HandleFunc("/download", handlerDownload)
    http.HandleFunc("/download.xls", handlerDownloadXLS)
    exe, _ := os.Executable()
    staticDir := filepath.Join(filepath.Dir(exe), "wwwroot", "static")
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
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

	deptParam := r.URL.Query().Get("dept")
	var deptIDPtr *int
	if deptParam != "" {
		if dv, err := strconv.Atoi(deptParam); err == nil && dv > 0 {
			deptIDPtr = &dv
		}
	}
	q := r.URL.Query().Get("q")

	users, err := service.QueryUsersFiltered(ctx, deptIDPtr, q)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	depts, _ := service.QueryDepartments(ctx)

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
		}
		if row.Over > 0 {
			s.OverDays += 1
		}
		s.OverHours += row.Over
		s.LateMins += row.Late
		s.EarlyMins += row.Early
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
      <h1>考勤报表{{.Year}}-{{.Month}}</h1>
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
        <label>部门</label>
        <select name="dept">
          <option value="0" {{if $.SelDept0}}selected{{end}}>全部</option>
          {{range .Depts}}
          <option value="{{.DeptID}}" {{if index $.SelDept .DeptID}}selected{{end}}>{{.DeptName}}</option>
          {{end}}
        </select>
        <label>搜索</label>
        <input type="text" name="q" value="{{.Query}}" placeholder="工号/姓名" />
        <button type="submit">切换</button>
      </form>
      <button id="open-dl" class="download">下载</button>
      <div id="dl-modal" class="modal hidden">
        <div class="modal-content">
          <h2>导出报表</h2>
          <form id="dl-form" method="get" action="/download">
            <input type="hidden" name="year" value="{{.Year}}" />
            <input type="hidden" name="month" value="{{.Month}}" />
            <input type="hidden" name="dept" value="{{.Dept}}" />
            <input type="hidden" name="q" value="{{.Query}}" />
            <label>格式</label>
            <select name="fmt">
              <option value="csv" selected>CSV</option>
              <option value="xls">Excel</option>
            </select>
            <div class="modal-actions">
              <button type="submit" class="primary">开始下载</button>
              <button type="button" id="close-dl">取消</button>
            </div>
          </form>
        </div>
      </div>
    </header>
    <main>
    <div id="loading" class="modal hidden"><div class="modal-content"><span>处理中...</span></div></div>
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
    <th class="sum-col"><span>迟到</span><br><span>分钟</span></th>
    <th class="sum-col"><span>早退</span><br><span>分钟</span></th>
    <th class="sum-col"><span>请假</span><br><span>小时</span></th>
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
    <td class="sum-col">{{.LateMins}}</td>
    <td class="sum-col">{{.EarlyMins}}</td>
    <td class="sum-col">{{.LeaveHours}}</td>
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
	for i := now.Year() - 1; i <= now.Year()+1; i++ {
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

	leavesIdx, _ := service.QueryLeaveSymbols(ctx, firstDay, lastDay)
	leaveSumIdx := map[int]float64{}
	for _, r := range leavesIdx {
		val := extractFloat(r.Symbol)
		leaveSumIdx[r.UserID] += val
	}
	for uid, v := range leaveSumIdx {
		s := sum[uid]
		s.LeaveHours += v
		sum[uid] = s
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
		"Depts":     depts,
		"Dept": func() int {
			if deptIDPtr == nil {
				return 0
			}
			return *deptIDPtr
		}(),
		"SelDept": func() map[int]bool {
			m := map[int]bool{}
			if deptIDPtr != nil {
				m[*deptIDPtr] = true
			}
			return m
		}(),
		"SelDept0": deptIDPtr == nil,
		"Query":    q,
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

	deptParam := r.URL.Query().Get("dept")
	var deptIDPtr *int
	if deptParam != "" {
		if dv, err := strconv.Atoi(deptParam); err == nil && dv > 0 {
			deptIDPtr = &dv
		}
	}
	q := r.URL.Query().Get("q")

	users, _ := service.QueryUsersFiltered(ctx, deptIDPtr, q)
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

	w.Write([]byte("\xEF\xBB\xBF"))
	cw := csv.NewWriter(w)
	defer cw.Flush()

	row := []string{"工号", "姓名", "部门"}
	for i := 1; i <= dayCount; i++ {
		row = append(row, fmt.Sprintf("%d号上班", i))
		row = append(row, fmt.Sprintf("%d号加班", i))
	}
	row = append(row, "出勤天数", "加班小时", "加班天数", "迟到分钟", "早退分钟", "请假小时")
	cw.Write(row)

	sum2 := make(map[int]SumValue)
	for _, row := range att {
		s := sum2[row.UserID]
		if row.Required > 0 {
			s.PresentDays += row.Work / row.Required
		}
		if row.Over > 0 {
			s.OverDays += 1
		}
		s.OverHours += row.Over
		s.LateMins += row.Late
		s.EarlyMins += row.Early
		sum2[row.UserID] = s
	}

	leaves2, _ := service.QueryLeaveSymbols(ctx, firstDay, lastDay)
	leaveSum2 := map[int]float64{}
	for _, r := range leaves2 {
		val := extractFloat(r.Symbol)
		leaveSum2[r.UserID] += val
	}
	for uid, v := range leaveSum2 {
		s := sum2[uid]
		s.LeaveHours += v
		sum2[uid] = s
	}

	for _, u := range users {
		r := []string{u.Badge, u.Name, u.DeptName}
		for i := 1; i <= dayCount; i++ {
			v := data[u.UserID][i]
			r = append(r, v.Work)
			r = append(r, v.Over)
		}
		s := sum2[u.UserID]
		r = append(r, format2f(s.PresentDays), formatFloat(s.OverHours), format0f(s.OverDays), format0f(s.LateMins), format0f(s.EarlyMins), format0f(s.LeaveHours))
		cw.Write(r)
	}
}

func handlerDownloadXLS(w http.ResponseWriter, r *http.Request) {
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
	deptParam := r.URL.Query().Get("dept")
	var deptIDPtr *int
	if deptParam != "" {
		if dv, err := strconv.Atoi(deptParam); err == nil && dv > 0 {
			deptIDPtr = &dv
		}
	}
	q := r.URL.Query().Get("q")

	firstDay := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)
	dayCount := lastDay.Day()

	users, _ := service.QueryUsersFiltered(ctx, deptIDPtr, q)
	att, _ := service.QueryAtt(ctx, firstDay, lastDay)

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
		}
		if row.Over > 0 {
			s.OverDays += 1
		}
		s.OverHours += row.Over
		s.LateMins += row.Late
		s.EarlyMins += row.Early
		sum[uid] = s
	}

	leavesX, _ := service.QueryLeaveSymbols(ctx, firstDay, lastDay)
	leaveSumX := map[int]float64{}
	for _, r := range leavesX {
		val := extractFloat(r.Symbol)
		leaveSumX[r.UserID] += val
	}
	for uid, v := range leaveSumX {
		s := sum[uid]
		s.LeaveHours += v
		sum[uid] = s
	}

	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Header().Set("Content-Disposition", "attachment; filename=att.xls")

	w.Write([]byte("\xEF\xBB\xBF"))
	fmt.Fprint(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>att</title></head><body>")
	fmt.Fprint(w, "<table border=1>")
	fmt.Fprint(w, "<tr><th>工号</th><th>姓名</th><th>部门</th>")
	for i := 1; i <= dayCount; i++ {
		fmt.Fprintf(w, "<th>%d号上班</th>", i)
		fmt.Fprintf(w, "<th>%d号加班</th>", i)
	}
	fmt.Fprint(w, "<th>出勤天数</th><th>加班小时</th><th>加班天数</th><th>迟到分钟</th><th>早退分钟</th><th>请假小时</th></tr>")

	for _, u := range users {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td>", u.Badge, u.Name, u.DeptName)
		for i := 1; i <= dayCount; i++ {
			v := data[u.UserID][i]
			fmt.Fprintf(w, "<td>%s</td>", v.Work)
			fmt.Fprintf(w, "<td>%s</td>", v.Over)
		}
		s := sum[u.UserID]
		fmt.Fprintf(w, "<td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", format2f(s.PresentDays), formatFloat(s.OverHours), format0f(s.OverDays), format0f(s.LateMins), format0f(s.EarlyMins), format0f(s.LeaveHours))
	}
	fmt.Fprint(w, "</table></body></html>")
}
