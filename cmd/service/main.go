package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	// 项目首页接口
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"project": "AIWallHub",
			"desc":    "AI 壁纸分享平台",
			"status":  "running",
		})
	})

	// 启动服务，端口 8080
	r.Run(":8080")
}
