package web

import "fmt"

func identityHeaderDefs() []HeaderDef {
	return []HeaderDef{
		{Title: "工号", Style: "min-width: 60px; width: 60px;"},
		{Title: "姓名", Style: "min-width: 100px; width: 100px;"},
		{Title: "部门", Style: "min-width: 70px; width: 70px;"},
	}
}

func identityHeaders() []string {
	defs := identityHeaderDefs()
	out := make([]string, 0, len(defs))
	for _, d := range defs {
		out = append(out, d.Title)
	}
	return out
}

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
