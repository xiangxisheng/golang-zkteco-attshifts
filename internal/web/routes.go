package web

import (
    "context"
    "html/template"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "time"
    "zkteco-attshifts/internal/config"
    "zkteco-attshifts/internal/service"
)

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
    {{.TableHTML}}
    </main>
    </body>
    </html>
    `

    t, _ := template.New("html").Parse(tpl)

    now := time.Now()
    var years []int
    for i := now.Year()-1; i <= now.Year()+1; i++ {
        years = append(years, i)
    }
    var months []int
    for i := 1; i <= 12; i++ {
        months = append(months, i)
    }

    weekend, weekNames := computeWeekInfo(y, m)
    tableHTML := renderGridTableHTML(mModel, weekend, weekNames)

    obj := map[string]any{
        "Year":      y,
        "Month":     m,
        "Users":     users,
        "TableHTML": template.HTML(tableHTML),
        "Years":     years,
        "Months":    months,
        "SelYear":   map[int]bool{y: true},
        "SelMonth":  map[int]bool{m: true},
        "Weekend":   weekend,
        "WeekNames": weekNames,
        "Depts":     depts,
        "Dept": func() int { if deptIDPtr == nil { return 0 } ; return *deptIDPtr }(),
        "SelDept": func() map[int]bool { m := map[int]bool{} ; if deptIDPtr != nil { m[*deptIDPtr] = true } ; return m }(),
        "SelDept0": deptIDPtr == nil,
        "Query":    q,
        "Show":     mModel.Show,
        "SelCols": func() map[string]bool { m := map[string]bool{} ; for k,v := range mModel.Show { if v { m[k] = true } } ; return m }(),
        "ColOptions": func() []map[string]string { var opts []map[string]string ; for _, c := range allColumns() { opts = append(opts, map[string]string{"key": c.Key, "label": c.Title}) } ; return opts }(),
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
