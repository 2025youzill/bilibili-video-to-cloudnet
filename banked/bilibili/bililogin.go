// 暂时不用这部分,如果后续需要登录,可以参考这个文件
package bilibili

import (
	"net/http"
	"os"

	"bvtc/client"
	"bvtc/log"
	"bvtc/response"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
)

func BiliLogin(ctx *gin.Context) {
	cli, _ := client.GetBiliClient()
	qrCode, _ := cli.GetQRCode()
	buf, _ := qrCode.Encode()
	os.WriteFile("qrcode.png", buf, 0o644)
	result, err := cli.LoginWithQRCode(bilibili.LoginWithQRCodeParam{
		QrcodeKey: qrCode.QrcodeKey,
	})
	if err != nil || result.Code != 0 {
		log.Logger.Error("登录失败", log.Any("err", err))
		ctx.JSON(http.StatusOK, response.FailMsg("登录失败"))
		return
	}
	cookiesString := cli.GetCookiesString()

	// 确保data目录存在
	if err := os.MkdirAll("./bilibili/data", 0o755); err != nil {
		log.Logger.Error("创建data目录失败", log.Any("err", err))
		ctx.JSON(http.StatusOK, response.FailMsg("保存cookie失败"))
		return
	}

	// 保存cookie到文件
	if err := os.WriteFile("./bilibili/data/cookies.txt", []byte(cookiesString), 0o644); err != nil {
		log.Logger.Error("保存cookie失败", log.Any("err", err))
		ctx.JSON(http.StatusOK, response.FailMsg("保存cookie失败"))
		return
	}

	log.Logger.Info("登录成功", log.Any("cookies", cookiesString))
	ctx.JSON(http.StatusOK, response.SuccessMsg("登录成功"))
}

func BiliLoginWithCookie(ctx *gin.Context) {
	cli, _ := client.GetBiliClient()

	// 读取cookie文件
	cookiestring, err := os.ReadFile("./bilibili/data/cookies.txt")
	if err != nil {
		log.Logger.Error("读取cookie文件失败", log.Any("err", err))
		ctx.JSON(http.StatusOK, response.FailMsg("读取cookie失败"))
		return
	}
	// 设置cookie
	cli.SetCookiesString(string(cookiestring))

	log.Logger.Info("使用cookie登录成功", log.Any("cookies", string(cookiestring)))
	ctx.JSON(http.StatusOK, response.SuccessMsg("登录成功"))
}
