package main

import (
	"awesomeProject1/Gin/routers"
	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default() // 创建路由

	routers.InitRouter(r)

	r.Run("127.0.0.1:3000") //监听端口

}
