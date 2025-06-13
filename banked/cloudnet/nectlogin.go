package cloudnet

import (
	"bvtc/client"
	"bvtc/log"
	"bvtc/response"
	"context"
	"net/http"

	"github.com/chaunsin/netease-cloud-music/api/weapi"
	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Phone  string `json:"phone"`
	CtCode int64
}

func SendByPhone(ctx *gin.Context) {
	var req LoginReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Logger.Error("fail to bind json", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to bind json"))
		return
	}
	if req.Phone == "" {
		log.Logger.Error("phone number is empty", log.Any("req : ", req))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("phone number is empty"))
		return
	}
	req.CtCode = 86

	api, err := client.GetNetcloudApi()
	if err != nil {
		log.Logger.Error("client fail to init", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("client fail to init"))
		return
	}

	//发送验证码
	resp, err := api.SendSMS(ctx, &weapi.SendSMSReq{Cellphone: req.Phone, CtCode: req.CtCode})
	if err != nil {
		log.Logger.Error("email fail to send", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("email fail to send"))
		return
	}
	_ = resp
	log.Logger.Info("email success to send", log.Any("phone : ", req.Phone))
	ctx.JSON(http.StatusOK, response.SuccessMsg("email success to send"))
}

type VerifyReq struct {
	Phone    string `json:"phone"`
	Captcha  string `json:"captcha"`
	Remember bool   `json:"remember"`
	CtCode   int64
}

func VerifyCaptcha(ctx *gin.Context) {
	api, err := client.GetNetcloudApi()
	if err != nil {
		log.Logger.Error("client fail to init", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("client fail to init"))
		return
	}

	var req VerifyReq
	ctx.ShouldBindJSON(&req)
	if req.Phone == "" {
		log.Logger.Error("phone number is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("phone number is empty"))
		return
	}
	if req.Captcha == "" {
		log.Logger.Error("captcha is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("captcha is empty"))
		return
	}
	req.CtCode = 86

	//检验验证码
	resp, err := api.SMSVerify(ctx, &weapi.SMSVerifyReq{Cellphone: req.Phone, Captcha: req.Captcha, CtCode: req.CtCode})
	if err != nil {
		log.Logger.Error("captcha fail to verify", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("captcha fail to verify"))
		return
	}
	if resp.Code != 200 {
		log.Logger.Error("captcha fail to verify", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg(resp.Message))
		return
	}

	//接入网易云api
	loginResp, err := api.LoginCellphone(ctx, &weapi.LoginCellphoneReq{
		Phone:       req.Phone,
		Countrycode: req.CtCode,  //手机前缀（默认+86）
		Captcha:     req.Captcha, // 使用验证码登录
		Remember:    true,        // 记住登录状态
	})
	_ = loginResp
	if err != nil {
		log.Logger.Error("fail to netclogin", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("fail to netclogin"))
		return
	}

	//获取用户信息
	user, err := api.GetUserInfo(ctx, &weapi.GetUserInfoReq{})
	if err != nil {
		log.Logger.Error("获取用户信息失败", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("获取用户信息失败"))
		return
	}

	log.Logger.Info("user netclogin", log.Any("user : ", user))
	ctx.JSON(http.StatusOK, response.SuccessMsg(""))

}

func CheckCookie(ctx *gin.Context) {
	api, err := client.GetNetcloudApi()
	if err != nil {
		log.Logger.Error("client fail to init", log.Any("err : ", err))
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("client fail to init"))
		return
	}

	status := api.NeedLogin(context.Background())
	if status {
		log.Logger.Error("user need login")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("user need login"))
		return
	}

	log.Logger.Info("user already login")
	ctx.JSON(http.StatusOK, response.SuccessMsg("user already login"))
	
}
