package video

import (
	"bvtc/constant"
	"bvtc/log"
	"bvtc/netcloud/upload"
	"bvtc/tool/ffmpeg"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type AudioReq struct {
	Title  string
	Artist string
}

func TranslateVideoToAudio(filename string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Logger.Error("获取当前目录失败", log.Any("err", err))
		return err
	}
	inputFile := filepath.Join(currentDir, filename)

	if _, err = os.Stat(inputFile); os.IsNotExist(err) {
		log.Logger.Error("输入文件不存在", log.Any("file", inputFile))
		return err
	}

	outputFile := strings.TrimSuffix(filename, ".mp4") + ".mp3"
	defer os.Remove(outputFile) // 确保最后删除临时文件

	ffmpegPath, err := ffmpeg.ExtractFFmpeg()
	if err != nil {
		log.Logger.Error("FFmpeg 初始化失败", log.Any("err", err))
		return err
	}
	defer os.Remove(ffmpegPath)

	// 执行转换
	if err := convertToMP3(ffmpegPath, inputFile, outputFile); err != nil {
		return err
	}

	err = upload.UploadToNetCloud(outputFile)
	if err != nil {
		log.Logger.Error("上传失败", log.Any("err", err))
		return err
	}

	return nil
}

func convertToMP3(ffmpegPath, inputFile, outputFile string) error {
	// 临时文件路径
	tmpOutput := filepath.Join(constant.Filepath, "temp_audio.mp3")
	defer os.Remove(tmpOutput) // 确保最后删除临时文件

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
	step1Cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	log.Logger.Info("开始提取纯音频", log.Any("input", inputFile))
	if err := step1Cmd.Run(); err != nil {
		log.Logger.Error("第一步提取音频失败",
			log.Any("err", err),
			log.Any("stderr", stderrStep1.String()),
		)
		return err
	}

	// 添加元数据
	coverArt := filepath.Join(constant.Filepath, "71ff5ea4-3c19-49ca-858b-fefc5022d84b.jpeg") // 文件封面（并非歌曲封面）
	step2Cmd := exec.Command(ffmpegPath,
		"-i", tmpOutput, // 音频文件
		"-i", coverArt, // 封面图片
		"-filter_complex", "[1:v]scale=960:960:force_original_aspect_ratio=decrease,pad=960:960:(ow-iw)/2:(oh-ih)/2[v]", // 调整尺寸并保持宽高比,尺寸不够用黑边补全(pad)
		"-map", "0:a", // 音频流
		"-map", "[v]", // 调整后的封面流
		"-c:a", "copy", // 直接复制流（无需重新编码）
		"-c:v", "mjpeg", // 重新编码封面为JPEG格式
		"-id3v2_version", "3", // 采用ID3V2.3版本，可以保存封面
		"-metadata", "title=探窗", // 标题
		"-metadata", "artist=兰音", // 歌手
		"-metadata", "album=", // 专辑(留空)
		"-metadata:s:v", "title=Cover", 
		"-metadata:s:v", "comment=Cover (Front)", 
		"-disposition:v", "attached_pic", 
		"-y",
		outputFile,
	)
	var stderrStep2 bytes.Buffer
	step2Cmd.Stderr = &stderrStep2
	step2Cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	log.Logger.Info("开始添加元数据", log.Any("output", outputFile))
	if err := step2Cmd.Run(); err != nil {
		log.Logger.Error("添加元数据失败",
			log.Any("err", err),
			log.Any("stderr", stderrStep2.String()),
		)
		return err
	}

	log.Logger.Info("转换成功", log.Any("output", outputFile))
	return nil
}
