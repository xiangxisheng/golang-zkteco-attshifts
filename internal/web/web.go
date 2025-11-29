package web

import (
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zkteco-attshifts/internal/config"
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
	NormalOT    float64
	WeekendOT   float64
	HolidayOT   float64
	E1Business  float64
	E2Sick      float64
	E3Personal  float64
	E4Home      float64
	E5Annual    float64
}

type ReportModel struct {
	Year  int
	Month int
	Days  []int
	Users []service.UserInfo
	Daily map[int]map[int]DayValue
	Sum   map[int]SumValue
	Show  map[string]bool
	Mode  string
}

type Column struct {
	Key      string
	Title    string
	SumField string
	Value    func(SumValue) string
	Default  bool
}

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

func identityHeaders() []string               { return []string{"工号", "姓名", "部门"} }
func identityRow(u service.UserInfo) []string { return []string{u.Badge, u.Name, u.DeptName} }
func dailyHeaderTitles(m ReportModel) []string {
	titles := []string{}
	if m.Mode == "all" || m.Mode == "work" || m.Mode == "over" {
		for _, d := range m.Days {
			if m.Mode == "all" || m.Mode == "work" {
				titles = append(titles, fmt.Sprintf("%d号上班", d))
			}
			if m.Mode == "all" || m.Mode == "over" {
				titles = append(titles, fmt.Sprintf("%d号加班", d))
			}
		}
	}
	return titles
}
func dailyRowValues(m ReportModel, userID int) []string {
	vals := []string{}
	if m.Mode == "all" || m.Mode == "work" || m.Mode == "over" {
		for _, d := range m.Days {
			v := m.Daily[userID][d]
			if m.Mode == "all" || m.Mode == "work" {
				vals = append(vals, v.Work)
			}
			if m.Mode == "all" || m.Mode == "over" {
				vals = append(vals, v.Over)
			}
		}
	}
	return vals
}

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
		r := append([]string{}, identityRow(u)...)
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
		for _, v := range identityRow(u) {
			fmt.Fprintf(w, "<td>%s</td>", v)
		}
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
	fmt.Fprint(w, "<table>")
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
		for _, v := range identityRow(u) {
			fmt.Fprintf(w, "<td>%s</td>", v)
		}
		for _, v := range dailyRowValues(m, u.UserID) {
			fmt.Fprintf(w, "<td>%s</td>", v)
		}
		s := m.Sum[u.UserID]
		for _, c := range visibleColumns(m) {
			fmt.Fprintf(w, "<td>%s</td>", c.Value(s))
		}
		fmt.Fprint(w, "</tr>")
	}
	io.WriteString(w, "</table></body></html>")
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
	if f == 0 {
		return "0"
	}
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}
func formatPresent(f float64) string {
	if f == 0 {
		return "0"
	}
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if i := strings.IndexByte(s, '.'); i >= 0 {
		if len(s)-i-1 > 2 {
			return fmt.Sprintf("%d", int64(f))
		}
	}
	return s
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
			"PresentDays": formatPresent(v.PresentDays),
			"OverHours":   formatFloat(v.OverHours),
			"OverDays":    format0f(v.OverDays),
			"LateMins":    format0f(v.LateMins),
			"EarlyMins":   format0f(v.EarlyMins),
			"LeaveHours":  format2f(v.LeaveHours),
			"NormalOT":    formatFloat(v.NormalOT),
			"WeekendOT":   formatFloat(v.WeekendOT),
			"HolidayOT":   formatFloat(v.HolidayOT),
			"E1":          format2f(v.E1Business),
			"E2":          format2f(v.E2Sick),
			"E3":          format2f(v.E3Personal),
			"E4":          format2f(v.E4Home),
			"E5":          format2f(v.E5Annual),
		}
	}
	return out
}

func resolveWWWRoot(cfg config.Config) string {
	base := cfg.WWWRoot
	if base == "" {
		base = "wwwroot"
	}
	if filepath.IsAbs(base) {
		return base
	}
	wd, _ := os.Getwd()
	p1 := filepath.Join(wd, base)
	if st, err := os.Stat(p1); err == nil && st.IsDir() {
		return p1
	}
	exe, _ := os.Executable()
	p2 := filepath.Join(filepath.Dir(exe), base)
	return p2
}

func RegisterRoutes(cfg config.Config) {
	root := resolveWWWRoot(cfg)
	fsRoot := http.FileServer(http.Dir(root))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handlerIndex(w, r)
			return
		}
		fsRoot.ServeHTTP(w, r)
	})
	http.HandleFunc("/download", handlerDownload)
	http.HandleFunc("/download.xls", handlerDownloadXLS)
	http.HandleFunc("/download.html", handlerDownloadHTML)
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	mModel, err := buildModel(ctx, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	y := mModel.Year
	m := mModel.Month
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

	users := mModel.Users
	depts, _ := service.QueryDepartments(ctx)

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
        <button type="button" id="open-cols">列选择</button>
      </form>
      <div id="cols-modal" class="modal hidden">
        <div class="modal-content">
          <h2>选择要显示的列</h2>
          <form id="cols-form" method="get" action="/">
            <input type="hidden" name="year" value="{{.Year}}" />
            <input type="hidden" name="month" value="{{.Month}}" />
            <input type="hidden" name="dept" value="{{.Dept}}" />
            <input type="hidden" name="q" value="{{.Query}}" />
            <div class="col-picker">
              {{range .ColOptions}}
              <label><input type="checkbox" name="cols" value="{{.key}}" {{if index $.SelCols .key}}checked{{end}}>{{.label}}</label>
              {{end}}
            </div>
            <div class="modal-actions">
              <button type="submit" class="primary">应用</button>
              <button type="button" id="close-cols">取消</button>
            </div>
          </form>
        </div>
      </div>
      <button id="open-dl" class="download">下载</button>
      <div id="dl-modal" class="modal hidden">
        <div class="modal-content">
          <h2>导出报表</h2>
          <form id="dl-form" method="get" action="/download">
            <input type="hidden" name="year" value="{{.Year}}" />
            <input type="hidden" name="month" value="{{.Month}}" />
            <input type="hidden" name="dept" value="{{.Dept}}" />
            <input type="hidden" name="q" value="{{.Query}}" />
            {{range $k,$v := .SelCols}}{{if $v}}<input type="hidden" name="cols" value="{{$k}}" />{{end}}{{end}}
            <label>格式</label>
            <div class="col-picker">
              <label><input type="radio" name="fmt" value="csv" checked>CSV</label>
              <label><input type="radio" name="fmt" value="xls">Excel</label>
              <label><input type="radio" name="fmt" value="html">HTML</label>
            </div>
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
    {{range .IdentityHeaders}}
    <th>{{.}}</th>
    {{end}}
    {{range .Days}}
    <th class="{{if index $.Weekend .}}weekend{{end}}">{{.}}<br><span class="wk">{{index $.WeekNames .}}</span><br>上</th>
    <th class="{{if index $.Weekend .}}weekend{{end}}">{{.}}<br><span class="wk">{{index $.WeekNames .}}</span><br>加</th>
    {{end}}
    {{range .SumHeaderTitles}}
    <th class="sum-col">{{.}}</th>
    {{end}}
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
    {{ $sum := index $.SumStr (printf "%d" $uid) }}
    {{range $.SumValueOrder}}
    <td class="sum-col">{{index $sum .}}</td>
    {{end}}
    </tr>
    {{end}}

    </table>
    </main>
    </body>
    </html>
    `

	t, _ := template.New("html").Parse(tpl)

	now := time.Now()
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

	show := mModel.Show
	data := mModel.Daily
	sum := mModel.Sum

	obj := map[string]any{
		"Year":      y,
		"Month":     m,
		"Days":      mModel.Days,
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
		"Show":     show,
		"SelCols": func() map[string]bool {
			m := map[string]bool{}
			for k, v := range show {
				if v {
					m[k] = true
				}
			}
			return m
		}(),
		"IdentityHeaders": identityHeaders(),
		"SumHeaderTitles": func() []string {
			titles := []string{}
			for _, c := range allColumns() {
				if show[c.Key] {
					titles = append(titles, c.Title)
				}
			}
			return titles
		}(),
		"SumValueOrder": func() []string {
			order := []string{}
			for _, c := range allColumns() {
				if show[c.Key] {
					order = append(order, c.SumField)
				}
			}
			return order
		}(),
		"ColOptions": func() []map[string]string {
			var opts []map[string]string
			for _, c := range allColumns() {
				opts = append(opts, map[string]string{"key": c.Key, "label": c.Title})
			}
			return opts
		}(),
	}

	t.Execute(w, obj)
}

func handlerDownload(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	mModel, _ := buildModel(ctx, r)
	renderCSVModel(w, mModel)
}

func handlerDownloadXLS(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	mModel, _ := buildModel(ctx, r)
	renderXLSModel(w, mModel)
}

func handlerDownloadHTML(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	mModel, _ := buildModel(ctx, r)
	renderHTMLModel(w, mModel)
}
