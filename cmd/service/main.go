// API服务器
package main

import (
	"AIWallHub/config"
	"AIWallHub/internal/handler"
	"AIWallHub/internal/middleware"
	"AIWallHub/pkg/email"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	config.InitDB()
	config.LoadSMTPConfig()

	// 初始化邮件服务
	email.InitEmail(email.Config{
		Host:     config.SMTPConfig.Host,
		Port:     config.SMTPConfig.Port,
		Username: config.SMTPConfig.Username,
		Password: config.SMTPConfig.Password,
	})

	r := gin.Default()
	r.Static("/uploads", "./uploads")

	//公开路由，不需要登录
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)
	r.POST("/send-code", handler.SendVerifyCode)
	r.GET("/posts", handler.GetPosts)
	r.GET("/post/:id", handler.GetPost)

	//需要登录的路由组
	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		//用户管理
		authorized.GET("/user", handler.GetUsers)                //用户列表
		authorized.GET("/user/:id", handler.GetUser)             //单个用户
		authorized.PUT("/user/:id", handler.UpdateUser)          //更新用户
		authorized.DELETE("/user/:id", handler.DeleteUser)       //删除用户
		authorized.PUT("/user/password", handler.UpdatePassword) // 修改密码
		authorized.PUT("/user/email", handler.UpdateEmail)       // 修改邮箱

		//动态管理
		authorized.POST("/post", handler.CreatePost)
		authorized.DELETE("/post/:id", handler.DeletePost)
		authorized.GET("/user/:id/posts", handler.GetUserPosts)
	}

	r.Run(":8080")

}
