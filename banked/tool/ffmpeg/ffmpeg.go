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

package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"bvtc/log"
	"bvtc/tool/randomstring"
)

// 检查系统是否安装了ffmpeg
func checkSystemFFmpeg() (string, bool) {
	// 尝试在PATH中查找ffmpeg
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path, true
	}
	return "", false
}

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
	// 首先尝试使用内嵌的ffmpeg
	binData, err := getFFmpegBin()
	if err == nil {
		log.Logger.Info("使用内嵌的FFmpeg")

		// 生成唯一的临时文件名（避免多用户冲突）
		filename := "ffmpeg"
		if runtime.GOOS == "windows" {
			filename = "ffmpeg_" + randomstring.GenerateRandomString(8) + ".exe"
		} else {
			filename = "ffmpeg_" + randomstring.GenerateRandomString(8)
		}
		ffmpegTmpExe := filepath.Join("file", filename)

		// 写入二进制数据
		if err := os.WriteFile(ffmpegTmpExe, binData, 0o755); err != nil {
			log.Logger.Error("写入临时文件失败:", log.Any("err", err))
			return "", fmt.Errorf("写入临时文件失败: %v", err)
		}
		return ffmpegTmpExe, nil
	}

	// 如果内嵌版本不存在，尝试使用系统版本
	if systemPath, exists := checkSystemFFmpeg(); exists {
		log.Logger.Info("内嵌FFmpeg不存在，使用系统安装的FFmpeg", log.Any("path", systemPath))
		return systemPath, nil
	}

	log.Logger.Error("未找到可用的FFmpeg")
	return "", fmt.Errorf("未找到可用的FFmpeg，请确保已正确安装")
}
