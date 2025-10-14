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
// 输入原始视频标题，返回更像歌曲名的简短建议与置信度，供业务层决策是否采纳。
type Suggester interface {
	Suggest(ctx context.Context, originalTitle string) (suggestion string, err error)
}

// Provider 抽象“大模型提供方”，根据提示词返回纯文本。
type Provider interface {
	CompleteText(ctx context.Context, prompt string) (string, error)
}

// Config 为服务配置项，包含模型名、超时、缓存 TTL、最小置信度与最大标题长度等。
type Config struct {
	Model          string
	Timeout        time.Duration
	CacheTTL       time.Duration
	MaxTitleLength int
}

// Service 实现 Suggester，提供：LLM 调用、超时控制、内存缓存。
type Service struct {
	provider Provider
	cfg      Config

	mu    sync.Mutex
	cache map[string]cachedItem
}

type cachedItem struct {
	suggestion string
	expireAt   time.Time
}

// NewService 创建 Service，并设置默认值（MinConfidence/MaxTitleLength）。
func NewService(p Provider, cfg Config) *Service {
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
	orig = ReplaceQuotes(orig)

	if suggested, ok := s.getCache(orig); ok {
		log.Logger.Info("AITitle cache hit",
			log.String("original", truncateForLog(orig, 120)),
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
			log.String("promptPreview", truncateForLog(prompt, 200)),
			log.String("error", err.Error()),
			log.Float32("elapsedMs", float32(time.Since(start).Milliseconds())),
		)
		return "", err
	}

	// 输出结果
	suggested := strings.TrimSpace(text)

	if suggested == "" {
		log.Logger.Warn("AITitle low confidence or empty suggestion",
			log.String("original", truncateForLog(orig, 120)),
			log.String("suggested", truncateForLog(suggested, 120)),
		)
		return "", errors.New("empty suggestion from AI model")
	}
	s.setCache(orig, suggested)
	log.Logger.Info("AITitle success",
		log.String("original", truncateForLog(orig, 120)),
		log.String("suggested", truncateForLog(suggested, 120)),
		log.Float32("elapsedMs", float32(time.Since(start).Milliseconds())),
	)
	return suggested, nil
}

// buildPrompt：提问模板
func (s *Service) buildPrompt(question string) string {
	return fmt.Sprintf(`
下面给了一首歌曲的视频的标题和简介，根据这些内容，提取里面的歌名，只用你提取的结果，其他什么都不要
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

// normalizeQuotes 将双引号替换为单引号
func ReplaceQuotes(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "“", "'")
	s = strings.ReplaceAll(s, "”", "'")
	return s
}

// truncateForLog 截断长文本，避免日志爆量
func truncateForLog(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}
