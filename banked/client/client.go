package client

import (
	"bvtc/banked/config"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/chaunsin/netease-cloud-music/api"
	"github.com/chaunsin/netease-cloud-music/api/weapi"
	"github.com/chaunsin/netease-cloud-music/pkg/cookie"
	"github.com/chaunsin/netease-cloud-music/pkg/log"
)

var (
	netcApi  *weapi.Api
	netcCli  *api.Client
	biliCli  *bilibili.Client
	initOnce sync.Once
	initErr  error
)

// 暴露初始化方法给主程序
func InitNetcloudCli() error {
	log.Default = log.New(&log.Config{
		Level:  "info",
		Stdout: true,
	})

	// 获取配置信息
	cfg := config.GetConfig()

	// 检查 cookie 文件是否存在，如果不存在则创建
	cookieFile := cfg.Api.Cookie.Filepath
	if _, err := os.Stat(cookieFile); os.IsNotExist(err) {
		// 创建文件目录
		dir := filepath.Dir(cookieFile)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Println("创建 cookie 目录失败,err:", err)
			return err
		}
		// 创建空的 cookie 文件
		file, err := os.Create(cookieFile)
		if err != nil {
			fmt.Println("创建 cookie 文件失败,err:", err)
			return err
		}
		defer file.Close()
		// 写入空的 JSON 对象
		if err := json.NewEncoder(file).Encode(map[string]interface{}{}); err != nil {
			fmt.Println("写入 cookie 文件失败,err:", err)
			return err
		}
	}

	// 将配置信息转换为 api.Config 结构体
	netcApiCfg := api.Config{
		Debug:   cfg.Api.Debug,
		Timeout: cfg.Api.Timeout,
		Retry:   cfg.Api.Retry,
		Cookie: cookie.Config{
			Options:  nil,
			Filepath: cfg.Api.Cookie.Filepath,
			Interval: cfg.Api.Cookie.Interval,
		},
	}

	// 初始化客户端
	netcCli = api.New(&netcApiCfg)
	netcApi = weapi.New(netcCli)
	return nil
}

// 获取已初始化的客户端
func GetNetcloudCli() (*api.Client, error) {
	if netcCli == nil {
		return nil, errors.New("client not initialized")
	}
	return netcCli, nil
}

// 获取已初始化的api接口
func GetNetcloudApi() (*weapi.Api, error) {
	if netcApi == nil {
		return nil, errors.New("client not initialized")
	}
	return netcApi, nil
}

// 暴露初始化方法给主程序
func InitBiliCli() error {
	// 获取配置信息
	biliCli = bilibili.NewAnonymousClient()
	if biliCli == nil {
		return errors.New("failed to create client")
	}
	return nil
}

// 获取已初始化的客户端
func GetBiliClient() (*bilibili.Client, error) {
	if biliCli == nil {
		return nil, errors.New("client not initialized")
	}
	return biliCli, nil
}
