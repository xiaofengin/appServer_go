package routers

import (
	"AppServer/controller"
	"AppServer/setting"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	if setting.Conf.Release {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.POST("/app/chat/user/register", controller.Register)
	r.POST("/app/chat/user/login", controller.Login)
	r.GET("/token/rtc/channel/:channel/agorauid/:agorauid", controller.RtcToken)
	return r
}
