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
	"bvtc/log"
	"bvtc/response"
	redis_pool "bvtc/tool/pool"
	"io"
	"net/http"

	"github.com/chaunsin/netease-cloud-music/api/weapi"
	"github.com/gin-gonic/gin"
)

func GetUserAvatar(ctx *gin.Context) {
	sessionId, err := ctx.Cookie("SessionId")
	if err != nil {
		log.Logger.Error("fail to get cookiefile", log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to get cookiefile"))
		return
	}
	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sessionId
	cookiefile, err := rdb.HGet(rtcx, key, "cookieFile").Result()
	if err != nil {
		log.Logger.Error("fail to get cookiefile", log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to get cookiefile"))
		return
	}
	api, _, err := client.MultiInitNetcloudCli(cookiefile)
	if err != nil {
		log.Logger.Error("fail to init netcloud client", log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to init netcloud client"))
		return
	}
	userinfo, err := api.GetUserInfo(ctx, &weapi.GetUserInfoReq{})
	if err != nil {
		log.Logger.Error("fail to get userinfo", log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to get userinfo"))
		return
	}

	// 直接获取图片数据并返回
	avatarUrl := userinfo.Profile.AvatarUrl
	if avatarUrl == "" {
		log.Logger.Error("avatar URL is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("avatar URL is empty"))
		return
	}

	// 发起HTTP请求获取图片
	resp, err := http.Get(avatarUrl)
	if err != nil {
		log.Logger.Error("fail to fetch avatar image", log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to fetch avatar image"))
		return
	}
	defer resp.Body.Close()

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Error("fail to read avatar image data", log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to read avatar image data"))
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", resp.Header.Get("Content-Type"))
	ctx.Header("Content-Length", resp.Header.Get("Content-Length"))
	ctx.Header("Cache-Control", "public, max-age=3600") // 缓存1小时

	// 直接返回图片数据
	ctx.Data(http.StatusOK, resp.Header.Get("Content-Type"), imageData)

}
