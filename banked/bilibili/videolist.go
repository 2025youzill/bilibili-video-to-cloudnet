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

package bilibili

import (
	"net/http"
	"strings"

	"bvtc/client"
	"bvtc/log"
	"bvtc/response"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
)

type GetVideoListReq struct {
	Avid int    `form:"avid,omitempty"`
	Bvid string `form:"bvid,omitempty"`
}

type GetVideoListResp struct {
	Author    string      `json:"author"`               // 作者名字
	ListTitle *string     `json:"list_title,omitempty"` // 合集标题
	VideoList []videoList `json:"video_list,omitempty"` // 合集列表
}

type videoList struct {
	Bvid  string `json:"bvid"`  // 视频bvid
	Title string `json:"title"` // 视频标题
	Url   string `json:"url"`
}

func GetVideoList(ctx *gin.Context) {
	var req GetVideoListReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		log.Logger.Error("bind query fail", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("bind query fail"))
		return
	}

	if req.Avid == 0 && req.Bvid == "" {
		log.Logger.Error("avid or bvid is required")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("avid or bvid is required"))
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
		errMsg := err.Error()
		if strings.Contains(errMsg, "错误码: -400") {
			ctx.JSON(http.StatusBadRequest, response.FailMsg(errMsg))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.FailMsg(errMsg))
		}
		return
	}
	var videolist []videoList
	videolist = append(videolist, videoList{
		Bvid:  videoinfo.Bvid,
		Title: videoinfo.Title,
		Url:   "https://www.bilibili.com/video/" + videoinfo.Bvid,
	})
	seasonid := videoinfo.SeasonId
	mid := videoinfo.Owner.Mid

	var resp GetVideoListResp
	if seasonid != 0 {
		for pagenum := 1; pagenum <= 5; pagenum++ {
			listinfo, err := cli.GetVideoCollectionInfo(bilibili.GetVideoCollectionInfoParam{Mid: mid, SeasonId: seasonid, PageNum: pagenum, PageSize: 100})
			if err != nil {
				log.Logger.Error("get video collection info fail", log.Any("err : ", err))
				ctx.JSON(http.StatusInternalServerError, "get video collection info fail")
				return
			}
			for i := range listinfo.Archives {
				if listinfo.Archives[i].Bvid != req.Bvid {
					videolist = append(videolist, videoList{
						Bvid:  listinfo.Archives[i].Bvid,
						Title: listinfo.Archives[i].Title,
						Url:   "https://www.bilibili.com/video/" + listinfo.Archives[i].Bvid,
					})
				}
			}
			resp.ListTitle = &listinfo.Meta.Name
		}
	}

	resp.Author = videoinfo.Owner.Name
	resp.VideoList = videolist

	ctx.JSON(http.StatusOK, response.SuccessMsg(resp))
}
