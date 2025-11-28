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

type Config struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// Init 自动使用当前目录下 config.json，不存在则使用 exe 同目录
func Init(configPath string) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		exe, _ := os.Executable()
		configPath = filepath.Join(filepath.Dir(exe), "config.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic("读取配置失败: " + err.Error())
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic("解析配置失败: " + err.Error())
	}

	connStr := fmt.Sprintf(
		"server=%s;port=%d;user id=%s;password=%s;database=%s;encrypt=disable",
		cfg.Server, cfg.Port, cfg.User, cfg.Password, cfg.Database,
	)

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	conn = db
	fmt.Println("数据库连接成功")
}

// Get 返回 *sql.DB
func Get() *sql.DB {
	return conn
}

func Close() {
	if conn != nil {
		conn.Close()
	}
}
