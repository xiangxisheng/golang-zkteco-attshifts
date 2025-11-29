package db

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

	_ "github.com/microsoft/go-mssqldb"
)

var conn *sql.DB
var initErr error

type Config struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// Init 自动使用当前目录下 config.json，不存在则使用 exe 同目录
func Init(configPath string) error {
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        exe, _ := os.Executable()
        configPath = filepath.Join(filepath.Dir(exe), "config.json")
    }

    data, err := os.ReadFile(configPath)
    if err != nil {
        initErr = fmt.Errorf("读取配置失败: %w", err)
        return initErr
    }

	var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        initErr = fmt.Errorf("解析配置失败: %w", err)
        return initErr
    }

	connStr := fmt.Sprintf(
		"server=%s;port=%d;user id=%s;password=%s;database=%s;encrypt=disable",
		cfg.Server, cfg.Port, cfg.User, cfg.Password, cfg.Database,
	)

    db, err := sql.Open("sqlserver", connStr)
    if err != nil {
        initErr = err
        return initErr
    }

    if err := db.Ping(); err != nil {
        initErr = err
        return initErr
    }

    conn = db
    fmt.Println("数据库连接成功")
    initErr = nil
    return nil
}

// Get 返回 *sql.DB
func Get() *sql.DB {
    return conn
}

func IsReady() bool { return conn != nil && initErr == nil }
func InitError() error { return initErr }

func Close() {
	if conn != nil {
		conn.Close()
	}
}
