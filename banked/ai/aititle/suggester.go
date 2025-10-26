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

package aititle

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"bvtc/log"
)

// Suggester 为“标题建议器”的接口：
type Suggester interface {
	Suggest(ctx context.Context, originalTitle string) (suggestion string, err error)
}

// Provider 抽象“大模型提供方”，根据提示词返回纯文本。
type Provider interface {
	CompleteText(ctx context.Context, prompt string) (string, error)
}

// ServerConfig 服务配置项
type ServerConfig struct {
	Model          string
	Timeout        time.Duration
	CacheTTL       time.Duration
	MaxTitleLength int
}

// Service 实现 Suggester，提供：LLM 调用、超时控制、内存缓存。
type Service struct {
	provider Provider
	cfg      ServerConfig

	mu    sync.Mutex
	cache map[string]cachedItem
}

type cachedItem struct {
	suggestion string
	expireAt   time.Time
}

// NewService 初始化 Service。
func NewService(p Provider, cfg ServerConfig) *Service {
	cfg.MaxTitleLength = 200

	return &Service{
		provider: p,
		cfg:      cfg,
		cache:    make(map[string]cachedItem),
	}
}

// 纯文本模式，不再依赖 JSON 结构化返回

// Suggest：核心流程
// 1) 预处理与缓存命中
// 2) 构造提示词并调用 LLM（带超时）
// 3) 读取模型文本输出
// 4) 写入缓存并返回
func (s *Service) Suggest(ctx context.Context, question string) (string, error) {
	orig := strings.TrimSpace(question)
	if orig == "" {
		return "", errors.New("empty original title")
	}

	if suggested, ok := s.getCache(orig); ok {
		log.Logger.Info("AITitle cache hit",
			log.String("original", orig),
		)
		return suggested, nil
	}

	prompt := s.buildPrompt(orig)
	ctx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()

	start := time.Now()
	text, err := s.provider.CompleteText(ctx, prompt)
	if err != nil {
		log.Logger.Warn("AITitle provider error",
			log.String("model", s.cfg.Model),
			log.Int("timeoutSeconds", int(s.cfg.Timeout/time.Second)),
			log.String("promptPreview", prompt),
			log.String("error", err.Error()),
			log.Float32("elapsedMs", float32(time.Since(start).Milliseconds())),
		)
		return "", err
	}

	// 输出结果
	suggested := strings.TrimSpace(text)

	if suggested == "" {
		log.Logger.Warn("AITitle low confidence or empty suggestion",
			log.String("original", orig),
			log.String("suggested", suggested),
		)
		return "", errors.New("empty suggestion from AI model")
	}
	s.setCache(orig, suggested)
	log.Logger.Info("AITitle success",
		log.String("original", orig),
		log.String("suggested", suggested),
		log.Float32("elapsedMs", float32(time.Since(start).Milliseconds())),
	)
	return suggested, nil
}

// buildPrompt：提问模板
func (s *Service) buildPrompt(question string) string {
	return fmt.Sprintf(`下面给了一首歌曲视频的标题和简介，根据这些内容，分析这个视频的歌名是什么
有以下限制：
1.返回它的歌名，并且只用返回最最最主要的一首
2.只用你提取的结果，其他什么都不要
现在的输入：%s`, question)
}

// getCache：同一原标题在 TTL 内不重复请求 LLM。
func (s *Service) getCache(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.cache[key]
	if !ok || time.Now().After(item.expireAt) {
		if ok {
			delete(s.cache, key)
		}
		return "", false
	}
	return item.suggestion, true
}

// setCache：写入简易内存缓存。
func (s *Service) setCache(key, suggestion string) {
	s.mu.Lock()
	s.cache[key] = cachedItem{
		suggestion: suggestion,
		expireAt:   time.Now().Add(s.cfg.CacheTTL),
	}
	s.mu.Unlock()
}
