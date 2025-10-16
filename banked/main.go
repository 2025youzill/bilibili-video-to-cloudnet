package main

import (
	"net/http"
	"time"

	"bvtc/ai"
	"bvtc/client"
	"bvtc/config"
	"bvtc/log"
	"bvtc/route"

	redis_pool "bvtc/tool/pool"
	"bvtc/tool/spew"
)

func main() {
	log.InitLogger(config.GetConfig().Log.Path, config.GetConfig().Log.Level)
	spew.InitSpew()
	// 程序启动时初始化网易云接口
	if _, _, err := client.MultiInitNetcloudCli(""); err != nil {
		panic("网易云接口初始化失败: " + err.Error())
	}
	if err := client.InitBiliCli(); err != nil {
		panic("哔哩哔哩接口初始化失败：" + err.Error())
	}

	// 初始化redis
	redis_pool.InitRedis()

	go ai.WarmupAITitle()
	newRouter := route.NewRouter()
	s := &http.Server{
		Addr:           ":8080",
		Handler:        newRouter,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Logger.Error("server error", log.Any("serverError", err))
	}
}
