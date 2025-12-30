package web

import (
	"html/template"
	"io"
	"net/http"
	"zkteco-attshifts/internal/license"
)

func LicenseGuard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, msg := license.Check()
		if status != license.Ok {
			lic, _ := license.Load()
			title := lic.Title
			if title == "" { title = "无法访问" }
			var detail string
			switch status {
			case license.Missing:
				if lic.Missing != "" { detail = lic.Missing } else { detail = msg }
			case license.Invalid:
				if lic.Invalid != "" { detail = lic.Invalid } else { detail = msg }
			case license.Expired:
				detail = msg
			default:
				detail = msg
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>"+template.HTMLEscapeString(title)+"</title><style>body{font-family:sans-serif;padding:24px}code{background:#f1f5f9;padding:4px 8px;border-radius:4px}</style></head><body>")
			io.WriteString(w, "<h1>"+template.HTMLEscapeString(title)+"</h1>")
			io.WriteString(w, "<p>"+template.HTMLEscapeString(detail)+"</p>")
			if lic.Footer != "" {
				io.WriteString(w, "<p>"+template.HTMLEscapeString(lic.Footer)+"</p>")
			} else {
				io.WriteString(w, "<p>请使用授权工具生成 <code>license.json</code> 并放置到程序目录。</p>")
			}
			io.WriteString(w, "</body></html>")
			return
		}
		next(w, r)
	}
}
