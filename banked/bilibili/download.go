package bilibili

import (
	"bvtc/client"
	"bvtc/constant"
	"bvtc/log"
	"bvtc/response"
	"bvtc/tool/randomstring"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func DownloadVideo(ctx *gin.Context) {
	cli, err := client.GetBiliClient()
	if err != nil {
		log.Logger.Error("client init fail", log.Any("err", err))
		return
	}

	bvid := ctx.Query("bvid")

	videoinfo, err := cli.GetVideoInfo(bilibili.VideoParam{Bvid: bvid})
	if err != nil {
		log.Logger.Error("get video info fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("get video info fail"))
		return
	}

	cid := videoinfo.Cid
	stream, err := cli.GetVideoStream(bilibili.GetVideoStreamParam{Bvid: bvid, Cid: cid})
	if err != nil {
		log.Logger.Error("get video stream fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("get video stream fail"))
		return
	}

	title := videoinfo.Title

	url := stream.Durl[0].Url
	filename := filepath.Join(constant.Filepath, fmt.Sprintf("%s.mp4", title))
	defer os.Remove(filename)

	err = os.MkdirAll(constant.Filepath, 0o755)
	if err != nil {
		log.Logger.Error("create output directory fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("create output directory fail"))
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
		log.Logger.Error("download fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("download fail"))
		return
	}
	if resp.StatusCode() != 200 {
		log.Logger.Error("request fail", log.Any("status code", resp.StatusCode()))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("request fail"))
		return
	}

	var audioreq AudioReq
	audioreq.Filename = filename
	audioreq.Artist = videoinfo.Owner.Name
	audioreq.Title = title

	mid := videoinfo.Owner.Mid
	artistinfo, err := cli.GetUserCard(bilibili.GetUserCardParam{Mid: mid})
	if err != nil {
		log.Logger.Error("get user card fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("get user card fail"))
		return
	}
	coverurl := artistinfo.Card.Face
	coverfilename := filepath.Join(constant.Filepath, fmt.Sprintf("%s.jpeg", randomstring.GenerateRandomString(16)))
	defer os.Remove(coverfilename)
	coverresp, err := resty.New().R().
		SetOutput(coverfilename).
		Get(coverurl)
	if err != nil {
		log.Logger.Error("download cover fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("download cover fail"))
		return
	}
	if coverresp.StatusCode() != 200 {
		log.Logger.Error("request cover fail", log.Any("status code", coverresp.StatusCode()))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("request cover fail"))
		return
	}
	audioreq.CoverArt = coverfilename

	err = TranslateVideoToAudio(audioreq, false, 0, "")
	if err != nil {
		log.Logger.Error("translate video to audio fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("translate video to audio fail"))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessMsg("download success"))
}

func GetVideoDesc(ctx *gin.Context) {
	cli, err := client.GetBiliClient()
	if err != nil {
		log.Logger.Error("client init fail", log.Any("err", err))
		return
	}
	bvid := ctx.Query("bvid")
	desc, err := cli.GetVideoDesc(bilibili.VideoParam{Bvid: bvid})
	if err != nil {
		log.Logger.Error("get video desc fail", log.Any("err", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("get video desc fail"))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessMsg(desc))
}
