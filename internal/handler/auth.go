package handler

import (
	"AIWallHub/internal/model"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// 注册接口(用邮箱方式)
func Register(c *gin.Context) {
	var json struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数格式错误",
		})
		return
	}

	//校验用户名长度
	usernamelen := utf8.RuneCountInString(json.Username)
	if usernamelen < 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户名需要3-20个字符,当前用户名过短",
		})
		return
	} else if usernamelen > 20 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户名需要3-20个字符,当前用户名过长",
		})
		return
	}

	//校验密码长度
	if len(json.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "密码至少需要6个字符",
		})
		return
	}

	//校验密码格式
	passwordRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !passwordRegex.MatchString(json.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "密码只能包含大小写字母、数字和下划线_",
		})
		return
	}

	//校验邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(json.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "邮箱格式不正确",
		})
		return
	}

	//检查邮箱是否已被注册
	if _, exists := model.UserByEmail[json.Email]; exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "该邮箱已被注册",
		})
		return
	}

	//创建新用户
	user := model.User{
		ID:       model.NextID,
		Name:     json.Username,
		Email:    json.Email,
		Password: json.Password,
	}

	model.Users[user.ID] = user
	model.UserByEmail[user.Email] = user
	model.NextID++

	c.JSON(http.StatusOK, gin.H{
		"message": "您已成功注册，请记好密码",
		"user_id": user.ID,
	})

}

/*-----------------------------------------------------------------*/

// 登录接口(用邮箱方式)
func Login(c *gin.Context) {
	var json struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//解析JSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数格式错误",
		})
		return
	}

	//登录信息检查
	user, exists := model.UserByEmail[json.Email]
	if !exists || user.Password != json.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "邮箱或密码错误",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "登陆成功",
		"user_id":  user.ID,
		"username": user.Name,
	})
}
