// API服务器
package main

import (
	"AIWallHub/config"
	"AIWallHub/internal/handler"
	"AIWallHub/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	r := gin.Default()

	//公开路由，不需要登录
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	//需要登录的路由组
	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.GET("/user", handler.GetUsers)          //用户列表
		authorized.GET("/user/:id", handler.GetUser)       //单个用户
		authorized.PUT("/user/:id", handler.UpdateUser)    //更新用户
		authorized.DELETE("/user/:id", handler.DeleteUser) //删除用户
	}

	r.Run(":8080")

}
