// API服务器
package main

import (
	"AIWallHub/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.Run(":8080")
}
