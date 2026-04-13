package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var JWTSecret []byte

// LoadEnv 加载 .env 文件（从项目根目录）
func LoadEnv() {
	// 获取当前文件路径
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("获取当前文件路径失败")
		return
	}

	// 向上查找项目根目录（包含 go.mod 的目录）
	dir := filepath.Dir(filename) // config 目录
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break // 找到项目根目录
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Println("未找到 go.mod，使用当前目录")
			break
		}
		dir = parent
	}

	// 加载 .env 文件
	envPath := filepath.Join(dir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("未找到 .env 文件 (%s)，使用默认配置", envPath)
	} else {
		log.Printf("已加载配置文件: %s", envPath)
	}
}

func InitDB() {
	// 加载环境变量
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
