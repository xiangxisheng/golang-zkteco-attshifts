package web

import "zkteco-attshifts/internal/service"

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

type HeaderDef struct {
    Title string
    Style string
    Width int
}
