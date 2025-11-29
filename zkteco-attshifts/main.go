package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "time"
    "zkteco-attshifts/internal/config"
    "zkteco-attshifts/internal/db"
    "zkteco-attshifts/internal/web"
)

func main() {
	cfgPath := "./config.json"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		exe, _ := os.Executable()
		cfgPath = filepath.Join(filepath.Dir(exe), "config.json")
	}

    cfg, err := config.Load(cfgPath)
    if err != nil {
        log.Println("读取配置失败:", err)
        // 使用默认配置继续启动，端口与wwwroot使用默认值
    }

    if err := db.Init(cfgPath); err != nil {
        log.Println("数据库初始化失败:", err)
    }
    defer db.Close()

    web.RegisterRoutes(cfg)

	port := cfg.HTTPPort
	if port == 0 {
		port = 8080
	}
    addr := fmt.Sprintf("http://127.0.0.1:%d", port)
    fmt.Printf("Server started: %s\n", addr)

    if runtime.GOOS == "windows" {
        go func() {
            time.Sleep(300 * time.Millisecond)
            openBrowser(addr)
        }()
    }

    if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
        log.Println("服务启动失败:", err)
    }
}

func openBrowser(url string) {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
    case "darwin":
        cmd = exec.Command("open", url)
    default:
        cmd = exec.Command("xdg-open", url)
    }
    _ = cmd.Start()
}
