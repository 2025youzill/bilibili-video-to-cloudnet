package bilibili

import (
	"bvtc/banked/client"
	"bvtc/banked/log"
	"bvtc/banked/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetVideoListReq struct {
	Avid string `json:"avid,omitempty" request:"query,omitempty"`
	Bvid string `json:"bvid,omitempty" request:"query,omitempty"`
}

func GetVideoList(ctx *gin.Context) {
	var req GetVideoListReq
	ctx.ShouldBindJSON(&req)
	if req.Avid == "" && req.Bvid == "" {
		ctx.JSON(http.StatusBadRequest, response.FailMsg("avid or bvid is required"))
		return
	}
	cli, err := client.GetBiliClient()
	if err != nil {
		log.Logger.Error("client init fail", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, "client init fail")
		return
	}
	cli = _
}
