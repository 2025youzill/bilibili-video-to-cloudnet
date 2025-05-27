package ffmpeg

import (
	_ "embed"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"
)

//go:embed ffmpeg.exe
var ffmpegBin []byte


// 生成随机字符串
func randomString(n int) string {
	// 1. 创建独立随机源（避免全局锁，并发安全）
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>32)))

	// 2. 定义字符集
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)

	// 3. 生成随机字符
	for i := range b {
		b[i] = letters[rng.IntN(len(letters))]
	}
	return string(b)
}

// 提取 FFmpeg 到临时文件
func ExtractFFmpeg() (string, string, error) {
	// 生成唯一的临时文件名（避免多用户冲突）
	ffmpegTmpExe := filepath.Join(os.TempDir(), "ffmpeg_"+randomString(8)+".exe")
	ffprobeTmpExe := filepath.Join(os.TempDir(), "ffprobe_"+randomString(8)+".exe")
	// 写入二进制数据
	if err := os.WriteFile(ffmpegTmpExe, ffmpegBin, 0755); err != nil {
		return "", "", fmt.Errorf("写入临时文件失败: %v", err)
	}
	return ffmpegTmpExe, ffprobeTmpExe, nil
}
