package bilibili

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"bvtc/client"
	"bvtc/config"
	"bvtc/constant"
	"bvtc/log"
	"bvtc/response"
	"bvtc/tool/randomstring"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/semaphore"
)

type VideoStreamReq struct {
	Bvid      []string `json:"bvid"`          // 稿件 bvid
	Splaylist bool     `json:"splaylist"`     // 是否上传到歌单
	Pid       int64    `json:"pid,omitempty"` // 歌单 id
}

type VideoStreamResp struct {
	Success []string `json:"success,omitempty"` // 成功处理的视频bvid
	Failed  []failed `json:"failed,omitempty"`  // 失败处理的视频bvid以及错误信息
}

type failed struct {
	Title string `json:"title,omitempty"` // 视频标题
	Error string `json:"error,omitempty"` // 错误信息
}

type result struct {
	Title string
	Err   error
}

func LoadMP4(ctx *gin.Context) {
	var req VideoStreamReq
	ctx.ShouldBindJSON(&req)
	if len(req.Bvid) == 0 || req.Bvid[0] == "" {
		log.Logger.Error("bvid is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("bvid is empty"))
		return
	}

	if req.Splaylist && req.Pid == 0 {
		log.Logger.Error("pid is required when splaylist is true")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("pid is required when splaylist is true"))
		return
	}

	cli, err := client.GetBiliClient()
	if err != nil {
		log.Logger.Error("client init fail", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, "client init fail")
		return
	}

	var wg sync.WaitGroup

	// 创建信号量，限制最大并发数
	sem := semaphore.NewWeighted(config.GetConfig().Music.Concurrency)
	resultChan := make(chan result, len(req.Bvid))

	for i, bvid := range req.Bvid {
		wg.Add(1)
		// 获取信号量
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Logger.Error("获取信号量失败", log.Any("err", err))
			continue
		}

		go func(index int, bvid string) {
			defer wg.Done()
			defer sem.Release(1) // 释放信号量

			videoinfo, err := cli.GetVideoInfo(bilibili.VideoParam{Bvid: bvid})
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("get video info fail: %v", err)}
				return
			}
			cid := videoinfo.Cid

			stream, err := cli.GetVideoStream(bilibili.GetVideoStreamParam{Bvid: bvid, Cid: cid})
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("get video stream fail: %v", err)}
				return
			}
			// 处理特殊字符
			title := sanitizeFilename(videoinfo.Title)
			// 下载视频
			url := stream.Durl[0].Url
			filename := filepath.Join(constant.Filepath, fmt.Sprintf("%s.mp4", title))
			defer os.Remove(filename)
			err = os.MkdirAll(constant.Filepath, 0o755)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("创建输出目录失败: %v", err)}
				return
			}

			referer := cli.Resty().Header.Get("Referer")
			useragent := cli.Resty().Header.Get("User-Agent")
			resp, err := resty.New().R().
				SetHeader("User-Agent", useragent).
				SetHeader("Referer", referer).
				SetOutput(filename).
				Get(url)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("下载失败: %v", err)}
				return
			}
			if resp.StatusCode() != 200 {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("请求失败: status code %d", resp.StatusCode())}
				return
			}

			// 获取视频信息
			var audioreq AudioReq
			audioreq.Filename = filename
			audioreq.Artist = videoinfo.Owner.Name
			audioreq.Title = title

			// 获取up头像当作封面
			mid := videoinfo.Owner.Mid
			artistinfo, err := cli.GetUserCard(bilibili.GetUserCardParam{Mid: mid})
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("获取用户空间详情失败: %v", err)}
				return
			}
			coverurl := artistinfo.Card.Face
			coverfilename := filepath.Join(constant.Filepath, fmt.Sprintf("%s.jpeg", randomstring.GenerateRandomString(16)))
			defer os.Remove(coverfilename)
			coverresp, err := resty.New().R().
				SetOutput(coverfilename).
				Get(coverurl)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("下载封面失败: %v", err)}
				return
			}
			if coverresp.StatusCode() != 200 {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("请求封面失败: status code %d", coverresp.StatusCode())}
				return
			}
			audioreq.CoverArt = coverfilename

			// 转化为mp3
			err = TranslateVideoToAudio(audioreq, req.Splaylist, req.Pid)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("上传失败: %v", err)}
				return
			}

			// 处理成功
			resultChan <- result{Title: videoinfo.Title, Err: nil}
		}(i, bvid)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集处理结果
	resp := VideoStreamResp{
		Success: make([]string, 0),
		Failed:  make([]failed, 0),
	}

	for result := range resultChan {
		if result.Err != nil {
			resp.Failed = append(resp.Failed, failed{
				Title: result.Title,
				Error: result.Err.Error(),
			})
		} else {
			resp.Success = append(resp.Success, result.Title)
		}
	}

	if len(resp.Failed) > 0 {
		ctx.JSON(http.StatusOK, response.SuccessMsg(resp))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessMsg(resp))
	}
}

func sanitizeFilename(filename string) string {
	// 替换所有可能造成问题的字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}
