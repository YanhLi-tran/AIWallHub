package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var JWTSecret []byte

func InitDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到.env文件，使用默认配置")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	if user == "" {
		user = "root"
	}
	if password == "" {
		password = "123456"
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "3306"
	}
	if dbname == "" {
		dbname = "wallpaper"
	}

	dsn := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname + "?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	log.Println("数据库连接成功")

	// 读取 JWT 密钥
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this"
	}
	JWTSecret = []byte(jwtSecret)

}
