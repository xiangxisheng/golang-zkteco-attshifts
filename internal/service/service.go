package service

import (
    "context"
    "fmt"
    "time"
    "zkteco-attshifts/internal/db"
)

type UserInfo struct {
	UserID   int
	Badge    string
	Name     string
	DeptName string
	DeptID   int
}

type AttRow struct {
	UserID   int
	AttDate  time.Time
	Work     float64
	Over     float64
	Required float64
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

type Department struct {
	DeptID   int
	DeptName string
}

func QueryDepartments(ctx context.Context) ([]Department, error) {
	sqlStr := `
    SELECT deptid, ISNULL(deptname,'')
    FROM departments
    ORDER BY deptid
    `

	rows, err := db.Get().QueryContext(ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Department{}
	for rows.Next() {
		var d Department
		rows.Scan(&d.DeptID, &d.DeptName)
		list = append(list, d)
	}
	return list, nil
}

func QueryUsersFiltered(ctx context.Context, deptID *int, q string) ([]UserInfo, error) {
    sqlStr := `
    SELECT u.userid, u.badgenumber, u.name, ISNULL(d.deptname,''), u.defaultdeptid
    FROM userinfo u
    LEFT JOIN departments d ON u.defaultdeptid=d.deptid
    WHERE u.[deltag]=0
    `
    args := []any{}
    if deptID != nil {
        sqlStr += fmt.Sprintf(" AND u.defaultdeptid=@p%d", len(args)+1)
        args = append(args, *deptID)
    }
    if q != "" {
        sqlStr += fmt.Sprintf(" AND (u.badgenumber LIKE @p%d OR u.name LIKE @p%d)", len(args)+1, len(args)+1)
        args = append(args, "%"+q+"%")
    }
    sqlStr += ` ORDER BY d.deptid, u.badgenumber`

    rows, err := db.Get().QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []UserInfo{}
	for rows.Next() {
		var u UserInfo
		rows.Scan(&u.UserID, &u.Badge, &u.Name, &u.DeptName, &u.DeptID)
		list = append(list, u)
	}
	return list, nil
}

func QueryAtt(ctx context.Context, start, end time.Time) ([]AttRow, error) {
	sqlStr := `
    SELECT userid, attdate,
        SUM(realworkday) AS work,
        SUM(overtime) AS [over],
        SUM(workday) AS required
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
		rows.Scan(&a.UserID, &a.AttDate, &a.Work, &a.Over, &a.Required)
		list = append(list, a)
	}
	return list, nil
}
