// API服务器
package main

import (
	handler "AIWallHub/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)
	r.Run(":8080")
}
