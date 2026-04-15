// Copyright (c) 2025 Youzill
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bvtc/ai"
	"bvtc/client"
	"bvtc/config"
	"bvtc/log"
	"bvtc/route"

	redis_pool "bvtc/tool/pool"
	"bvtc/tool/socket"
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
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8081"
	}
	s := &http.Server{
		Addr:           ":" + appPort,
		Handler:        newRouter,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.ListenAndServe()
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Error("server error", log.Any("serverError", err))
		}
	case <-sigCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Logger.Error("http server shutdown failed", log.Any("err", err))
		}
		if err := socket.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			log.Logger.Error("websocket manager shutdown failed", log.Any("err", err))
		}

		if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Error("server error", log.Any("serverError", err))
		}
	}

	if err := sigCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		log.Logger.Error("server error", log.Any("serverError", err))
	}
}
