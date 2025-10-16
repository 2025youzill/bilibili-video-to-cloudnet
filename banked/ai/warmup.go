package ai

import (
	"context"
	"time"

	"bvtc/ai/providers"
	"bvtc/config"
	"bvtc/log"
)

// WarmupAITitle 进程启动后预热一次模型，避免首次调用超时
func WarmupAITitle() {
	AiCfg := config.GetConfig().Ai
	p := providers.NewOllamaProvider(AiCfg.BaseURL, AiCfg.Model, AiCfg.Timeout)

	warmupTimeout := 60 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), warmupTimeout)
	defer cancel()

	// 轻提示：不关心输出内容，只为触发模型加载
	if _, err := p.CompleteText(ctx, "你好"); err != nil {
		log.Logger.Warn("AI warmup failed",
			log.String("baseURL", AiCfg.BaseURL),
			log.String("model", AiCfg.Model),
			log.Int("timeoutSeconds", int(warmupTimeout/time.Second)),
			log.String("error", err.Error()),
		)
		return
	}
	log.Logger.Info("AI warmup success",
		log.String("baseURL", AiCfg.BaseURL),
		log.String("model", AiCfg.Model),
	)
}
