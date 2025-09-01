/*
Package client provides API clients for interacting with external services.
It includes initialization and access methods for Netease Cloud Music and Bilibili API clients.
*/
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bvtc/config"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/chaunsin/netease-cloud-music/api"
	"github.com/chaunsin/netease-cloud-music/api/weapi"
	"github.com/chaunsin/netease-cloud-music/pkg/cookie"
	"github.com/chaunsin/netease-cloud-music/pkg/log"
)

var (
	netcApi *weapi.Api
	netcCli *api.Client
	biliCli *bilibili.Client
)

// 单机下暴露初始化方法给主程序
func InitNetcloudCli(cookieFile string) error {
	log.Default = log.New(&log.Config{
		Level:  "info",
		Stdout: true,
	})

	// 获取配置信息
	cfg := config.GetConfig()
	var userCookieFile string
	if cookieFile == "" {
		userCookieFile = filepath.Join(append(strings.Split(cfg.Api.Cookie.Filepath, "/"), "cookie.json")...)
	} else {
		userCookieFile = filepath.Join(append(strings.Split(cfg.Api.Cookie.Filepath, "/"), cookieFile)...)
	}
	// 检查 cookie 文件是否存在，如果不存在则创建
	if _, err := os.Stat(userCookieFile); os.IsNotExist(err) {
		// 创建文件目录
		dir := filepath.Dir(userCookieFile)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Println("创建 cookie 目录失败,err:", err)
			return err
		}
		// 创建空的 cookie 文件
		file, err := os.Create(userCookieFile)
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
			Filepath: userCookieFile,
			Interval: cfg.Api.Cookie.Interval,
		},
	}

	// 初始化客户端
	netcCli = api.New(&netcApiCfg)
	netcApi = weapi.New(netcCli)
	return nil
}

// 单机下获取已初始化的客户端
func GetNetcloudCli() (*api.Client, error) {
	if netcCli == nil {
		return nil, errors.New("client not initialized")
	}
	return netcCli, nil
}

// 单机下获取已初始化的api接口
func GetNetcloudApi() (*weapi.Api, error) {
	if netcApi == nil {
		return nil, errors.New("client not initialized")
	}
	return netcApi, nil
}

// 多用户部署下重置cloudnet客户端信息
func MultiInitNetcloudCli(cookieFile string) (*weapi.Api, *api.Client, error) {
	log.Default = log.New(&log.Config{
		Level:  "info",
		Stdout: true,
	})
	var netcCli *api.Client
	var netcApi *weapi.Api
	// 获取配置信息
	cfg := config.GetConfig()
	var userCookieFile string
	if cookieFile == "" {
		userCookieFile = filepath.Join(append(strings.Split(cfg.Api.Cookie.Filepath, "/"), "cookie.json")...)
	} else {
		userCookieFile = filepath.Join(append(strings.Split(cfg.Api.Cookie.Filepath, "/"), cookieFile)...)
	}
	// 检查 cookie 文件是否存在，如果不存在则创建
	if _, err := os.Stat(userCookieFile); os.IsNotExist(err) {
		// 创建文件目录
		dir := filepath.Dir(userCookieFile)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Println("创建 cookie 目录失败,err:", err)
			return netcApi, netcCli, err
		}
		// 创建空的 cookie 文件
		file, err := os.Create(userCookieFile)
		if err != nil {
			fmt.Println("创建 cookie 文件失败,err:", err)
			return netcApi, netcCli, err
		}
		defer file.Close()
		// 写入空的 JSON 对象
		if err := json.NewEncoder(file).Encode(map[string]interface{}{}); err != nil {
			fmt.Println("写入 cookie 文件失败,err:", err)
			return netcApi, netcCli, err
		}
	}

	// 将配置信息转换为 api.Config 结构体
	netcApiCfg := api.Config{
		Debug:   cfg.Api.Debug,
		Timeout: cfg.Api.Timeout,
		Retry:   cfg.Api.Retry,
		Cookie: cookie.Config{
			Options:  nil,
			Filepath: userCookieFile,
			Interval: cfg.Api.Cookie.Interval,
		},
	}

	// 初始化客户端
	netcCli = api.New(&netcApiCfg)
	netcApi = weapi.New(netcCli)
	return netcApi, netcCli, nil
}

func MultiGetNetcloudApi(cookieFile string) (*weapi.Api, error) {
	netcApi, _, err := MultiInitNetcloudCli(cookieFile)
	if err != nil {
		return nil, err
	}
	return netcApi, nil
}

func MultiGetNetcloudCli(cookieFile string) (*api.Client, error) {
	_, netcCli, err := MultiInitNetcloudCli(cookieFile)
	if err != nil {
		return nil, err
	}
	return netcCli, nil
}

// 单机下游客访问bilibili客户端
func InitBiliCli() error {
	// 获取配置信息
	biliCli = bilibili.NewAnonymousClient()
	if biliCli == nil {
		return errors.New("failed to create client")
	}
	return nil
}

// 登录bilibili客户端
func InitBiliLoginCli() error {
	biliCli = bilibili.New()
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

// 多用户部署下重置bilibili客户端信息
func MultiInitBiliCli() (*bilibili.Client, error) {
	// 获取配置信息
	var multiBiliCli *bilibili.Client
	multiBiliCli = bilibili.NewAnonymousClient()
	if multiBiliCli == nil {
		return nil, errors.New("failed to create client")
	}
	return multiBiliCli, nil
}
