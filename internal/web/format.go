package web

import (
    "fmt"
    "strconv"
    "strings"
)

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

func format0f(f float64) string { return strconv.FormatFloat(f, 'f', 0, 64) }

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
