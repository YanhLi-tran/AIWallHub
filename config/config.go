package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var JWTSecret []byte

// LoadEnv加载.env文件
func LoadEnv() {
	// 获取当前可执行文件所在目录
	dir, err := os.Getwd()
	if err != nil {
		log.Println("获取当前目录失败:", err)
		return
	}

	// 向上查找包含go.mod的目录（项目根目录）
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break //找到项目根目录
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Println("未找到 go.mod，使用当前目录")
			break
		}
		dir = parent
	}

	// 加载.env文件
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
