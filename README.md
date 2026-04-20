AIWallHub - 个人内容空间
基于 Go + Gin 开发的高性能个人内容空间后端服务，支持文字 / 图片 / 视频动态发布、社交互动、私信聊天等完整功能，开箱即用。
✨ 项目特性
完整用户体系：注册、登录、JWT 身份认证、邮箱验证
内容发布：支持纯文本、多图、视频动态
社交功能：点赞、收藏、评论、关注、互相关注检测
私信系统：用户一对一聊天，支持对话列表
隐私保护：用户可配置隐私权限，控制数据可见性
完善的文件存储、分页接口、错误处理
🛠️ 技术栈
表格
技术	说明
开发语言	Go
Web 框架	Gin
ORM 框架	GORM
数据库	MySQL
缓存	Redis
身份认证	JWT
文件存储	本地存储
邮件服务	SMTP
🚀 快速启动
1. 配置环境变量
在项目根目录创建 .env 文件，并填写配置信息：
env
# 数据库配置
DB_USER=root
DB_PASSWORD=123456
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=wallpaper

# JWT 密钥
JWT_SECRET=your-secret-key

# 邮箱验证码配置
SMTP_HOST=smtp.qq.com
SMTP_PORT=587
SMTP_USER=your-email@qq.com
SMTP_PASSWORD=your-auth-code
2. 安装项目依赖
bash
运行
go mod tidy
3. 启动服务
bash
运行
go run cmd/service/main.go
服务默认运行地址：http://localhost:8080
📖 API 文档
基础信息
Base URL：http://localhost:8080
认证方式：JWT Token
请求头格式：Authorization: Bearer <token>
🔓 公开路由（无需登录）
表格
接口	说明
POST /register	用户注册
POST /login	用户登录
POST /send-code	发送邮箱验证码
GET /posts	动态广场列表
GET /post/:id	动态详情
GET /post/:id/comments	动态评论列表
1. 用户注册
请求方式：POST
请求体：
json
{
    "username": "张三",
    "password": "abc123",
    "confirm_password": "abc123",
    "email": "user@example.com",
    "code": "123456"
}
响应：
json
{
    "message": "注册成功",
    "user_id": 1
}
2. 用户登录
请求方式：POST
请求体：
json
{
    "email": "user@example.com",
    "password": "abc123"
}
响应：
json
{
    "message": "登录成功",
    "user_id": 1,
    "username": "张三",
    "token": "eyJhbGciOiJIUzI1NiIs..."
}
🔒 需要认证路由（需登录）
所有接口需在请求头携带有效 JWT Token
用户管理
表格
方法	URL	说明
GET	/user	用户列表（分页）
GET	/user/:id	获取用户信息
PUT	/user/:id	更新用户信息
DELETE	/user/:id	删除用户
PUT	/user/password	修改密码
PUT	/user/email	修改邮箱
DELETE	/user/account	注销账号
动态管理
表格
方法	URL	说明
POST	/post	发布动态
DELETE	/post/:id	删除动态
GET	/user/:id/posts	获取用户动态列表
PUT	/post/:id	更新动态隐私设置
点赞 / 收藏
表格
方法	URL	说明
POST	/post/:id/like	点赞
DELETE	/post/:id/like	取消点赞
GET	/post/:id/likes	点赞用户列表
POST	/post/:id/favorite	收藏
DELETE	/post/:id/favorite	取消收藏
评论 / 关注
表格
方法	URL	说明
POST	/post/:id/comment	发表评论
DELETE	/comment/:id	删除评论
POST	/user/:id/follow	关注用户
DELETE	/user/:id/follow	取消关注
GET	/user/:id/followers	粉丝列表
GET	/user/:id/following	关注列表
私信聊天
表格
方法	URL	说明
POST	/user/:id/message	发送私信
GET	/user/:id/messages	获取聊天记录
GET	/messages/conversations	对话列表
⚠️ 错误响应
表格
HTTP 状态码	说明
400	请求参数错误
401	未登录 / Token 无效
403	无权限访问
404	资源不存在
500	服务器内部错误
通用错误格式：
json
{
    "error": "错误信息描述"
}
隐私限制（403）：
json
{
    "error": "该用户未公开关注/粉丝列表"
}
📂 项目结构
plaintext
AIWallHub/
├── cmd/
│   └── service/
│       └── main.go         # 项目入口
├── config/                 # 配置文件
├── internal/
│   ├── handler/            # HTTP 路由处理器
│   ├── service/            # 业务逻辑层
│   ├── model/              # 数据模型
│   └── middleware/         # 中间件（JWT、日志等）
├── pkg/                    # 公共工具包
├── uploads/                # 上传文件存储目录
├── .env                    # 环境变量
├── go.mod
└── README.md
📌 功能清单
 用户注册 / 登录 / JWT 认证
 邮箱验证码发送
 用户信息管理、密码修改、账号注销
 文字 / 图片 / 视频动态发布
 动态列表、详情、分页
 点赞 / 取消点赞
 收藏 / 取消收藏
 评论发布与删除
 关注 / 粉丝 / 互相关注
 一对一私信、对话列表
 隐私权限控制