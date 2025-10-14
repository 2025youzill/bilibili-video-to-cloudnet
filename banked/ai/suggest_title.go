package ai

import (
	"fmt"
	"net/http"
	"time"

	"bvtc/ai/aititle"
	"bvtc/ai/providers"
	"bvtc/client"
	"bvtc/config"
	"bvtc/log"
	"bvtc/response"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
)

type suggestTitleReq struct {
	Bvid string `json:"bvid"`
}

type suggestTitleResp struct {
	SuggestedTitle string `json:"suggestedTitle"`
}

// SuggestTitle 提供基于 AI 的歌曲标题建议
func SuggestTitle(c *gin.Context) {
	var req suggestTitleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailMsg("invalid request: bvid required"))
		return
	}

	cli, err := client.GetBiliClient()
	if err != nil {
		log.Logger.Error("client init fail", log.Any("err", err))
		return
	}
	videoinfo, err := cli.GetVideoInfo(bilibili.VideoParam{Bvid: req.Bvid})
	if err != nil {
		log.Logger.Error("get video info fail", log.Any("err", err))
		c.JSON(http.StatusInternalServerError, response.FailMsg("get video info fail"))
		return
	}
	originalTitle := videoinfo.Title
	originalDesc := videoinfo.Desc
	// log.Logger.Info("AI SuggestTitle request", log.Any("req", req))
	AiCfg := config.GetConfig().Ai

	apiProvider := AiCfg.Provider
	baseURL := AiCfg.BaseURL
	model := AiCfg.Model
	timeout := AiCfg.Timeout
	maxTitleLength := AiCfg.MaxTitleLength

	var p aititle.Provider
	p = providers.NewOllamaProvider(baseURL, model, timeout)

	s := aititle.NewService(p, aititle.Config{
		Model:          model,
		Timeout:        timeout,
		CacheTTL:       AiCfg.CacheTTL,
		MaxTitleLength: maxTitleLength,
	})

	question := fmt.Sprintf("标题：%s\n简介：%s\n", originalTitle, originalDesc)
	suggested, err := s.Suggest(c.Request.Context(), question)
	if err != nil {
		// 打印下游错误详情
		log.Logger.Warn("AI SuggestTitle failed",
			log.String("provider", apiProvider),
			log.String("baseURL", baseURL),
			log.String("model", model),
			log.Int("timeoutSeconds", int(timeout/time.Second)),
			log.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, response.FailMsg("AI 调用失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessMsg(suggestTitleResp{SuggestedTitle: suggested}))
}
