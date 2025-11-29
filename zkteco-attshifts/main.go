package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
		panic(err)
	}

	db.Init(cfgPath)
	defer db.Close()

    web.RegisterRoutes(cfg)

	port := cfg.HTTPPort
	if port == 0 {
		port = 8080
	}
	fmt.Printf("Server started: http://127.0.0.1:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
