package service

import (
	"context"
	"time"
	"zkteco-attshifts/internal/db"
)

type UserInfo struct {
	UserID   int
	Badge    string
	Name     string
	DeptName string
}

type AttRow struct {
	UserID  int
	AttDate time.Time
	Work    float64
	Over    float64
}

func QueryUsers(ctx context.Context) ([]UserInfo, error) {
	sqlStr := `
    SELECT u.userid, u.badgenumber, u.name, ISNULL(d.deptname,'')
    FROM userinfo u
    LEFT JOIN departments d ON u.defaultdeptid=d.deptid
    WHERE u.[deltag]=0
    ORDER BY d.deptid, u.badgenumber
    `

	rows, err := db.Get().QueryContext(ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []UserInfo{}
	for rows.Next() {
		var u UserInfo
		rows.Scan(&u.UserID, &u.Badge, &u.Name, &u.DeptName)
		list = append(list, u)
	}
	return list, nil
}

func QueryAtt(ctx context.Context, start, end time.Time) ([]AttRow, error) {
	sqlStr := `
    SELECT userid, attdate,
        SUM(realworkday) AS work,
        SUM(overtime) AS [over]
    FROM attshifts
    WHERE attdate BETWEEN @p1 AND @p2
      AND realworkday IS NOT NULL
    GROUP BY userid, attdate
    `

	rows, err := db.Get().QueryContext(ctx, sqlStr, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []AttRow{}
	for rows.Next() {
		var a AttRow
		rows.Scan(&a.UserID, &a.AttDate, &a.Work, &a.Over)
		list = append(list, a)
	}
	return list, nil
}
