package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"AIWallHub/pkg/cache"
	"AIWallHub/pkg/crypto"
	"AIWallHub/pkg/jwt"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// 注册接口(用邮箱方式)
func Register(c *gin.Context) {
	var json struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		Email           string `json:"email"`
		Code            string `json:"code"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数格式错误",
		})
		return
	}

	// 校验两次密码是否一致
	if json.Password != json.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "两次输入的密码不一致",
		})
		return
	}

	//校验用户名长度
	usernamelen := utf8.RuneCountInString(json.Username)
	if usernamelen < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户名需要1-20个字符,当前用户名过短",
		})
		return
	} else if usernamelen > 20 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户名需要1-20个字符,当前用户名过长",
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
	//临时注释
	// // 验证验证码
	// savedCode, ok := cache.Get(json.Email)
	// if !ok || savedCode != json.Code {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"error": "验证码错误或已过期",
	// 	})
	// 	return
	// }

	//检查邮箱是否已被注册
	var existingUser model.User
	if err := config.DB.Where("email=?", json.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "该邮箱已被注册",
		})
		return
	}

	//加密密码
	hashedPassword, err := crypto.HashPassword(json.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	//创建新用户
	user := model.User{
		Name:     json.Username,
		Email:    json.Email,
		Password: hashedPassword,
	}
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "注册失败",
		})
		return
	}

	// 删除验证码
	cache.Delete(json.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
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
	var user model.User
	if err := config.DB.Where("email=?", json.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "邮箱或密码错误",
		})
		return
	}

	// 检查用户是否已注销
	if user.Status == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "账号已注销",
		})
		return
	}

	if !crypto.CheckPassword(json.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "邮箱或密码错误",
		})
		return
	}

	// 生成 JWT token
	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "登陆成功",
		"user_id":  user.ID,
		"username": user.Name,
		"token":    token,
	})
}
