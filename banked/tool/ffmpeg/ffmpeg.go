package ffmpeg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"bvtc/log"
	"bvtc/tool/randomstring"
)

// 获取操作系统特定的FFmpeg二进制数据
func getFFmpegBin() ([]byte, error) {
	// 根据操作系统确定文件名
	filename := "ffmpeg"
	if runtime.GOOS == "windows" {
		filename = "ffmpeg.exe"
	}
	if runtime.GOOS == "linux" {
		filename = "ffmpeg"
	}
	// 构建文件路径（相对于当前工作目录）
	filePath := filepath.Join("tool", "ffmpeg", filename)

	// 读取文件内容
	binData, err := os.ReadFile(filePath)
	if err != nil {
		log.Logger.Error("读取FFmpeg二进制文件失败:", log.Any("err", err))
		return nil, fmt.Errorf("读取FFmpeg二进制文件失败: %v", err)
	}

	return binData, nil
}

// 提取 FFmpeg 到临时文件
func ExtractFFmpeg() (string, error) {
	// 获取二进制数据
	binData, err := getFFmpegBin()
	if err != nil {
		return "", err
	}

	// 生成唯一的临时文件名（避免多用户冲突）
	filename := "ffmpeg"
	if runtime.GOOS == "windows" {
		filename = "ffmpeg_" + randomstring.GenerateRandomString(8) + ".exe"
	} else {
		filename = "ffmpeg_" + randomstring.GenerateRandomString(8)
	}
	ffmpegTmpExe := filepath.Join("tool", "ffmpeg", filename)

	// 写入二进制数据
	if err := os.WriteFile(ffmpegTmpExe, binData, 0o755); err != nil {
		log.Logger.Error("写入临时文件失败:", log.Any("err", err))
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}
	return ffmpegTmpExe, nil
}
