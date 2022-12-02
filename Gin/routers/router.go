package routers

import (
	"awesomeProject1/Gin/controller"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.POST("/logagent_etcdctl", controller.Log_etcdctl)
	r.GET("/logagent_etcdctl", controller.Log_etcdctlPage)

}
