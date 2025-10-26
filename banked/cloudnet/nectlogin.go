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

package cloudnet

import (
	"bvtc/client"
	"bvtc/config"
	"bvtc/log"
	"bvtc/response"
	redis_pool "bvtc/tool/pool"
	"bvtc/tool/session"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chaunsin/netease-cloud-music/api/weapi"
	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Phone  string `json:"phone"`
	CtCode int64
}

func SendByPhone(ctx *gin.Context) {
	var req LoginReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Logger.Error("fail to bind json", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to bind json"))
		return
	}
	if req.Phone == "" {
		log.Logger.Error("phone number is empty", log.Any("req : ", req))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("phone number is empty"))
		return
	}
	req.CtCode = 86

	cookieFile := filepath.Join(session.GenerateSessionID(32) + ".json")

	api, err := client.MultiGetNetcloudApi(cookieFile)
	if err != nil {
		log.Logger.Error("client fail to init", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("client fail to init"))
		return
	}

	//发送验证码
	resp, err := api.SendSMS(ctx, &weapi.SendSMSReq{Cellphone: req.Phone, CtCode: req.CtCode})
	if err != nil {
		log.Logger.Error("email fail to send", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("email fail to send"))
		return
	}
	_ = resp

	sid := session.GenerateSessionID(16)
	spew.Dump("sid : ", sid)
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("SessionId", sid, 60*10, "/", "", false, true)
	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sid
	rdb.HSet(rtcx, key, map[string]interface{}{
		"cookieFile": cookieFile,
		"createdAt":  time.Now().Format(time.RFC3339),
		"isValid":    "true",
	})
	// 设置 Redis 会话键 10 分钟过期
	rdb.Expire(rtcx, key, 10*time.Minute)

	log.Logger.Info("email success to send", log.Any("phone : ", req.Phone))
	ctx.JSON(http.StatusOK, response.SuccessMsg("email success to send"))
}

type VerifyReq struct {
	Phone    string `json:"phone"`
	Captcha  string `json:"captcha"`
	Remember bool   `json:"remember"`
	CtCode   int64
}

func VerifyCaptcha(ctx *gin.Context) {
	sid, err := ctx.Cookie("SessionId")
	if err != nil {
		log.Logger.Error("fail to get sessionId", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to get sessionId"))
		return
	}
	spew.Dump("sid : ", sid)
	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sid
	cookieFile, _ := rdb.HGet(rtcx, key, "cookieFile").Result()
	if cookieFile == "" {
		log.Logger.Error("cookie file not found", log.String("sid", sid))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("cookie file not found"))
		return
	}

	api, err := client.MultiGetNetcloudApi(cookieFile)
	if err != nil {
		log.Logger.Error("client fail to init", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("client fail to init"))
		return
	}

	var req VerifyReq
	ctx.ShouldBindJSON(&req)
	if req.Phone == "" {
		log.Logger.Error("phone number is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("phone number is empty"))
		return
	}
	if req.Captcha == "" {
		log.Logger.Error("captcha is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("captcha is empty"))
		return
	}
	req.CtCode = 86

	//检验验证码
	resp, err := api.SMSVerify(ctx, &weapi.SMSVerifyReq{Cellphone: req.Phone, Captcha: req.Captcha, CtCode: req.CtCode})
	if err != nil {
		log.Logger.Error("captcha fail to verify", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("captcha fail to verify"))
		return
	}
	if resp.Code != 200 {
		log.Logger.Error("captcha fail to verify", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg(resp.Message))
		return
	}

	//接入网易云api
	loginResp, err := api.LoginCellphone(ctx, &weapi.LoginCellphoneReq{
		Phone:       req.Phone,
		Countrycode: req.CtCode,  //手机前缀（默认+86）
		Captcha:     req.Captcha, // 使用验证码登录
		Remember:    true,        // 记住登录状态
	})
	_ = loginResp
	if err != nil {
		log.Logger.Error("fail to netclogin", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("fail to netclogin"))
		return
	}

	//获取用户信息
	user, err := api.GetUserInfo(ctx, &weapi.GetUserInfoReq{})
	if err != nil {
		log.Logger.Error("获取用户信息失败", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("获取用户信息失败"))
		return
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("SessionId", sid, 60*60*24*7, "/", "", false, true)

	log.Logger.Info("user netclogin", log.Any("user : ", user))

	// 登录成功后，将 Redis 中的会话有效期延长至 7 天，并刷新标记
	rdb.HSet(rtcx, key, map[string]interface{}{
		"createdAt": time.Now().Format(time.RFC3339),
		"isValid":   "true",
	})
	rdb.Expire(rtcx, key, 7*24*time.Hour)
	ctx.JSON(http.StatusOK, response.SuccessMsg(""))

}

func CheckCookie(ctx *gin.Context) {
	// 先从 cookie 读取 SessionId，再到 Redis 查询对应的 cookie 文件名
	sid, err := ctx.Cookie("SessionId")
	if err != nil {
		log.Logger.Error("fail to get sessionId", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to get sessionId"))
		return
	}

	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sid
	cookieFile, rerr := rdb.HGet(rtcx, key, "cookieFile").Result()
	if rerr != nil || cookieFile == "" {
		log.Logger.Error("session not found or expired", log.Any("err : ", rerr), log.String("cookieFile", cookieFile))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("session not found or expired"))
		return
	}

	// 检查cookie文件是否存在且有内容
	cfg := config.GetConfig()
	cookieFilePath := filepath.Join(append(strings.Split(cfg.Api.Cookie.Filepath, "/"), cookieFile)...)
	if _, err := os.ReadFile(cookieFilePath); err != nil {
		log.Logger.Error("Cookie file not found or cannot be read", log.String("filePath", cookieFilePath), log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("Cookie file not found"))
		return
	} else {
		log.Logger.Info("Cookie file content")
	}

	api, _, err := client.MultiInitNetcloudCli(cookieFile)
	if err != nil {
		log.Logger.Error("client fail to init", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("client fail to init"))
		return
	}

	status := api.NeedLogin(context.Background())

	if status {
		log.Logger.Error("user need login")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("user need login"))
		return
	}
	log.Logger.Info("user already login")
	ctx.JSON(http.StatusOK, response.SuccessMsg("user already login"))

}

func DeleteCookie(ctx *gin.Context) {
	sid, err := ctx.Cookie("SessionId")
	if err != nil {
		log.Logger.Error("fail to get sessionId", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to get sessionId"))
		return
	}
	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sid
	// 读取 cookie 文件名
	cookieFile, _ := rdb.HGet(rtcx, key, "cookieFile").Result()

	// 删除磁盘上的 cookie 文件
	if cookieFile != "" {
		cfg := config.GetConfig()
		// 生成与创建时相同的路径：<cfg.Api.Cookie.Filepath>/<cookieFile>
		filePath := filepath.Join(append(strings.Split(cfg.Api.Cookie.Filepath, "/"), cookieFile)...)
		if err := os.Remove(filePath); err != nil {
			if !os.IsNotExist(err) {
				log.Logger.Error("failed to remove cookie file", log.Any("file", filePath), log.Any("err", err))
			}
		}
	}

	// 删除 Redis 中的会话数据
	rdb.Del(rtcx, key)

	// 清除浏览器端 SessionId Cookie（保持与设置时同样的属性）
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("SessionId", "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, response.SuccessMsg("cookie deleted"))
}
