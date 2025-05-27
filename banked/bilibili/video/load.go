package video

import (
	"banked/client"
	"banked/constant"
	"banked/log"
	"banked/response"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

type VideoStreamReq struct {
	Bvid string `json:"bvid,omitempty" request:"query,omitempty"` // 稿件 bvid。avid 与 bvid 任选一个
	Cid  int    `json:"cid"`                                      // 视频 cid
}

func LoadMP4(ctx *gin.Context) {
	var req VideoStreamReq
	ctx.ShouldBindJSON(&req)
	if req.Bvid == "" {
		log.Logger.Error("bvid is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("bvid is empty"))
		return
	}

	cli, err := client.GetBiliClient()
	if err != nil {
		log.Logger.Error("client init fail", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, "client init fail")
		return
	}

	videoinfo, err := cli.GetVideoInfo(bilibili.VideoParam{Bvid: req.Bvid})
	if err != nil {
		log.Logger.Error("get video info fail", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, "get video info fail")
		return
	}
	req.Cid = videoinfo.Cid

	stream, err := cli.GetVideoStream(bilibili.GetVideoStreamParam{Bvid: req.Bvid, Cid: req.Cid})
	if err != nil {
		fmt.Println("err : ", err)
		ctx.JSON(http.StatusInternalServerError, "fail")
		return
	}

	// 下载视频
	url := stream.Durl[0].Url
	filename := filepath.Join(constant.Filepath, fmt.Sprintf("%s.mp4", videoinfo.Title))
	defer os.Remove(filename)
	err = os.MkdirAll(constant.Filepath, 0755)
	if err != nil {
		log.Logger.Error("创建输出目录失败", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("创建输出目录失败"))
		return
	}
	referer := cli.Resty().Header.Get("Referer")
	useragent := cli.Resty().Header.Get("User-Agent")
	resp, err := resty.New().R().
		SetHeader("User-Agent", useragent).
		SetHeader("Referer", referer).
		SetOutput(filename). // 设置输出文件
		Get(url)
	if err != nil {
		log.Logger.Error("下载失败", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("下载失败"))
		return
	}
	if resp.StatusCode() != 200 {
		log.Logger.Error("请求失败", log.Any("statusCode : ", resp.StatusCode()))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("请求失败"))
		return
	}

	// 转化为mp3
	err = TranslateVideoToAudio(filename)
	if err != nil {
		log.Logger.Error("转化mp3失败", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("转化mp3失败"))
		return
	}

	ctx.JSON(http.StatusOK, "success")

}
