package main

import (
	"net/http"
	"time"

	"bvtc/client"
	"bvtc/config"
	"bvtc/log"
	"bvtc/route"
	"bvtc/tool/spew"
)

func main() {
	log.InitLogger(config.GetConfig().Log.Path, config.GetConfig().Log.Level)
	spew.InitSpew()
	// 程序启动时初始化网易云接口
	if err := client.InitNetcloudCli(); err != nil {
		panic("网易云接口初始化失败: " + err.Error())
	}
	if err := client.InitBiliCli(); err != nil {
		panic("哔哩哔哩接口初始化失败：" + err.Error())
	}

	newRouter := route.NewRouter()
	s := &http.Server{
		Addr:           ":8080",
		Handler:        newRouter,
		ReadTimeout:    300 * time.Second,
		WriteTimeout:   300 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Logger.Error("server error", log.Any("serverError", err))
	}
}
