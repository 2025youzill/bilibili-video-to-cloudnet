package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"bvtc/ai/aititle"
	"bvtc/ai/providers"
	"bvtc/client"
	"bvtc/config"
	"bvtc/response"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/semaphore"
)

// SuggestTitleBatchStream 基于 SSE 的批量流式返回
func SuggestTitleBatchStream(c *gin.Context) {
	bvidsParam := c.Query("bvids")
	if bvidsParam == "" {
		c.JSON(http.StatusBadRequest, response.FailMsg("invalid request: bvids required"))
		return
	}

	// 基础 SSE 头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 部分反向代理需要

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, response.FailMsg("stream unsupported"))
		return
	}

	// 解析 bvids
	var bvids []string
	for _, s := range strings.Split(bvidsParam, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			bvids = append(bvids, s)
		}
	}

	writeEvent := func(event string, payload any) {
		if event != "" {
			fmt.Fprintf(c.Writer, "event: %s\n", event)
		}
		if payload != nil {
			b, _ := json.Marshal(payload)
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(b))
		} else {
			fmt.Fprint(c.Writer, "data: {}\n\n")
		}
		flusher.Flush()
	}

	// 为该 SSE 响应单独放宽写超时
	if rc := http.NewResponseController(c.Writer); rc != nil {
		// 仅写超时放宽到 10 分钟
		_ = rc.SetWriteDeadline(time.Now().Add(10 * time.Minute))
	}

	writeEvent("open", map[string]any{"message": "stream started"})

	cli, err := client.GetBiliClient()
	if err != nil {
		writeEvent("error", map[string]any{"message": "client init fail"})
		return
	}

	AiCfg := config.GetConfig().Ai
	baseURL := AiCfg.BaseURL
	model := AiCfg.Model
	timeout := AiCfg.Timeout
	maxTitleLength := AiCfg.MaxTitleLength

	var p aititle.Provider
	p = providers.NewOllamaProvider(baseURL, model, timeout)

	s := aititle.NewService(p, aititle.ServerConfig{
		Model:          model,
		Timeout:        timeout,
		CacheTTL:       AiCfg.CacheTTL,
		MaxTitleLength: maxTitleLength,
	})

	// 逐个视频独立处理，限制并发数
	type item struct {
		bvid   string
		title  string
		desc   string
		errMsg string
	}

	items := make([]item, 0, len(bvids))
	for _, bvid := range bvids {
		it := item{bvid: bvid}
		videoinfo, vErr := cli.GetVideoInfo(bilibili.VideoParam{Bvid: bvid})
		if vErr != nil {
			it.errMsg = "get video info fail"
		} else {
			it.title = ReplaceQuotes(videoinfo.Title)
			it.desc = ReplaceQuotes(videoinfo.Desc)
		}
		items = append(items, it)
	}

	// 先发占位进度，避免前端长时间无反馈
	for _, it := range items {
		writeEvent("progress", struct {
			Bvid           string `json:"bvid"`
			SuggestedTitle string `json:"suggestedTitle,omitempty"`
			Error          string `json:"error,omitempty"`
		}{Bvid: it.bvid})
	}

	// 结果通道，由主 goroutine 串行写 SSE，避免并发写
	type result struct {
		bvid   string
		title  string
		errMsg string
	}
	results := make(chan result, len(items))

	// 并发控制（使用加权信号量）
	sem := semaphore.NewWeighted(AiCfg.Concurrency)
	var wg sync.WaitGroup

	for _, it := range items {
		it := it
		wg.Add(1)
		go func() {
			defer wg.Done()
			if it.errMsg != "" {
				results <- result{bvid: it.bvid, errMsg: it.errMsg}
				return
			}
			// 避免受 HTTP 连接影响
			if err := sem.Acquire(context.Background(), 1); err != nil {
				results <- result{bvid: it.bvid, errMsg: "semaphore acquire failed: " + err.Error()}
				return
			}
			defer sem.Release(1)

			// 构造提示
			var builder strings.Builder
			builder.WriteString("标题：")
			builder.WriteString(it.title)
			builder.WriteString("\n简介：")
			builder.WriteString(it.desc)
			question := builder.String()
			//log.Logger.Info("AI SuggestTitle request", log.String("question", question))

			// 由 Service 内部超时控制
			suggestedText, callErr := s.Suggest(context.Background(), question)
			if callErr != nil {
				results <- result{bvid: it.bvid, errMsg: "AI 调用失败: " + callErr.Error()}
				return
			}
			if strings.TrimSpace(suggestedText) == "" {
				results <- result{bvid: it.bvid, errMsg: "AI 返回为空"}
				return
			}
			results <- result{bvid: it.bvid, title: suggestedText}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// 心跳，避免代理/中间层超时断流
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	remaining := len(items)
	for remaining > 0 {
		select {
		case r, ok := <-results:
			if !ok {
				remaining = 0
				break
			}
			writeEvent("progress", struct {
				Bvid           string `json:"bvid"`
				SuggestedTitle string `json:"suggestedTitle,omitempty"`
				Error          string `json:"error,omitempty"`
			}{
				Bvid:           r.bvid,
				SuggestedTitle: r.title,
				Error:          r.errMsg,
			})
			remaining--
		case <-ticker.C:
			writeEvent("ping", map[string]any{"t": time.Now().Unix()})
		case <-c.Request.Context().Done():
			return
		}
	}

	writeEvent("done", map[string]any{"message": "completed"})
}

// ReplaceQuotes 将不应该的字符全部替换为单引号
func ReplaceQuotes(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "“", "'")
	s = strings.ReplaceAll(s, "”", "'")
	return s
}
