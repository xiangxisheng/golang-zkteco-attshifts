package license

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"time"
)

const Secret = "FBTech2025-License-Key"

type Status int

const (
	Ok Status = iota
	Missing
	Invalid
	Expired
)

type License struct {
    Expiry    string `json:"expiry"`
    Message   string `json:"message"`
    Signature string `json:"signature"`
    Title     string `json:"title"`
    Footer    string `json:"footer"`
    Missing   string `json:"missing"`
    Invalid   string `json:"invalid"`
}

func resolvePath() string {
	wd, _ := os.Getwd()
	p1 := filepath.Join(wd, "license.json")
	if st, err := os.Stat(p1); err == nil && !st.IsDir() {
		return p1
	}
	exe, _ := os.Executable()
	p2 := filepath.Join(filepath.Dir(exe), "license.json")
	return p2
}

func verify(lic License) bool {
	payload := lic.Expiry + "|" + lic.Message + "|" + Secret
	sum := crc32.ChecksumIEEE([]byte(payload))
	expect := fmt.Sprintf("%08x", sum)
	return expect == lic.Signature
}

func Check() (Status, string) {
    path := resolvePath()
    b, err := os.ReadFile(path)
    if err != nil {
        return Missing, "未授权，请运行授权工具生成许可文件"
    }
    if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
        b = b[3:]
    }
    var lic License
    if err := json.Unmarshal(b, &lic); err != nil {
        return Invalid, "授权文件格式错误"
    }
    if !verify(lic) {
        return Invalid, "授权文件校验失败"
    }
    if lic.Expiry == "" {
        return Invalid, "授权文件缺少过期日期"
    }
    exp, err := time.Parse("2006-01-02", lic.Expiry)
    if err != nil {
        return Invalid, "过期日期格式错误，应为YYYY-MM-DD"
    }
    if time.Now().After(exp.Add(24 * time.Hour)) {
        msg := lic.Message
        if msg == "" {
            msg = "授权已过期，请联系管理员"
        }
        return Expired, msg
    }
    return Ok, ""
}

func Load() (License, error) {
    path := resolvePath()
    b, err := os.ReadFile(path)
    if err != nil {
        return License{}, err
    }
    if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
        b = b[3:]
    }
    var lic License
    if err := json.Unmarshal(b, &lic); err != nil {
        return License{}, err
    }
    return lic, nil
}
