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
