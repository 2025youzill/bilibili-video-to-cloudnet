package bilibili

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bvtc/cloudnet"
	"bvtc/constant"
	"bvtc/log"
	"bvtc/tool/ffmpeg"
	"bvtc/tool/randomstring"
)

type AudioReq struct {
	Filename string
	Artist   string
	Title    string
	CoverArt string
}

func TranslateVideoToAudio(req AudioReq, splaylist bool, pid int64) error {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Logger.Error("获取当前目录失败", log.Any("err", err))
		return errors.New("获取当前目录失败")
	}
	inputFile := filepath.Join(currentDir, req.Filename)

	if _, err = os.Stat(inputFile); os.IsNotExist(err) {
		log.Logger.Error("输入文件不存在", log.Any("file", inputFile))
		return errors.New("输入文件不存在")
	}

	outputFile := strings.TrimSuffix(req.Filename, ".mp4") + ".mp3"
	defer os.Remove(outputFile) // 确保最后删除临时文件

	ffmpegPath, err := ffmpeg.ExtractFFmpeg()
	if err != nil {
		log.Logger.Error("FFmpeg 初始化失败", log.Any("err", err))
		return errors.New("FFmpeg 初始化失败")
	}
	defer os.Remove(ffmpegPath)

	// 执行转换
	if err := convertToMP3(ffmpegPath, inputFile, outputFile, req); err != nil {
		return errors.New("转换失败")
	}

	err = cloudnet.UploadToNetCloud(outputFile, splaylist, pid)
	if err != nil {
		log.Logger.Error("上传失败", log.Any("err", err))
		return err
	}

	return nil
}

// 这个封面有时候能用有时候不能？不知道是什么逻辑
func convertToMP3(ffmpegPath, inputFile, outputFile string, req AudioReq) error {
	// 检查封面文件是否存在
	if _, err := os.Stat(req.CoverArt); os.IsNotExist(err) {
		log.Logger.Error("封面文件不存在", log.Any("file", req.CoverArt))
		return err
	}

	// 生成暂时文件存储纯音频数据，防止并行时瞎缝
	tmpOutput := filepath.Join(constant.Filepath, fmt.Sprintf("%s.mp3", randomstring.GenerateRandomString(16)))
	defer os.Remove(tmpOutput) // 恢复临时文件清理

	// 生成无元数据的纯音频
	step1Cmd := exec.Command(ffmpegPath,
		"-i", inputFile,
		"-vn",                 // 禁用视频流
		"-map_metadata", "-1", // 清除所有元数据
		"-acodec", "libmp3lame", // 	音频编码(LAME3.101(bate3)) ? 网易云MP3保存部分此处为乱码,flac用	libFLAC 1.3.2 (2017-01-01)
		"-b:a", "320k", //	比特率
		"-y", // 覆盖输出文件
		tmpOutput,
	)
	var stderrStep1 bytes.Buffer
	step1Cmd.Stderr = &stderrStep1

	log.Logger.Info("开始提取纯音频", log.Any("input", inputFile), log.Any("tmpOutput", tmpOutput))
	if err := step1Cmd.Run(); err != nil {
		log.Logger.Error("第一步提取音频失败",
			log.Any("err", err),
			log.Any("stderr", stderrStep1.String()),
			log.Any("inputFile", inputFile),
			log.Any("tmpOutput", tmpOutput))
		return fmt.Errorf("提取音频失败: %v, 错误输出: %s", err, stderrStep1.String())
	}

	// 添加元数据
	step2Cmd := exec.Command(ffmpegPath,
		"-i", tmpOutput, // 音频文件
		"-i", req.CoverArt, // 封面图片
		"-filter_complex", "[1:v]scale=960:960:force_original_aspect_ratio=decrease,pad=960:960:(ow-iw)/2:(oh-ih)/2[v]", // 调整尺寸并保持宽高比,尺寸不够用黑边补全(pad)
		"-map", "0:a", // 音频流
		"-map", "[v]", // 调整后的封面流
		"-c:a", "copy", // 直接复制流（无需重新编码）
		"-c:v", "mjpeg", // 重新编码封面为JPEG格式
		"-id3v2_version", "3", // 采用ID3V2.3版本
		"-metadata", "title="+req.Title, // 标题
		"-metadata", "artist="+req.Artist, // 歌手
		"-metadata", "album=", // 专辑(留空)
		"-metadata:s:v", "title=Cover",
		"-metadata:s:v", "comment=Cover (Front)",
		"-disposition:v", "attached_pic",
		"-y",
		outputFile,
	)
	var stderrStep2 bytes.Buffer
	step2Cmd.Stderr = &stderrStep2

	log.Logger.Info("开始添加元数据", log.Any("output", outputFile))
	if err := step2Cmd.Run(); err != nil {
		log.Logger.Error("添加元数据失败", log.Any("err", err))
		return err
	}

	log.Logger.Info("转换成功", log.Any("output", outputFile))
	return nil
}
