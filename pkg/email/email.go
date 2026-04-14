package email

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"gopkg.in/gomail.v2"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

var config Config

// InitEmail 初始化邮件配置
func InitEmail(c Config) {
	config = c
	fmt.Printf("Email Init - Host: %s, Port: %d, User: %s\n", config.Host, config.Port, config.Username)
}

// GenerateCode 生成6位随机验证码
func GenerateCode() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := ""
	for i := 0; i < 6; i++ {
		code += fmt.Sprintf("%d", rng.Intn(10))
	}
	return code
}

// SendVerificationCode 发送验证码
func SendVerificationCode(toEmail, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.Username)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "【AIWallHub】邮箱验证码")

	body := fmt.Sprintf(`
        <html>
        <body>
            <h2>欢迎使用 AIWallHub</h2>
            <p>您的验证码是：<strong style="color:red;font-size:24px;">%s</strong></p>
            <p>验证码5分钟内有效，请勿泄露。</p>
        </body>
        </html>
    `, code)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}
