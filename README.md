# AIWallHub - 个人内容空间

基于 Go + Gin 的个人内容空间后端服务，支持文字、图片、视频动态发布，点赞、收藏、评论、关注、私信等社交功能。

## 技术栈

- **框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **缓存**: Redis
- **认证**: JWT
- **文件存储**: 本地存储

## 快速开始

### 1.配置环境变量

创建`.env`文件

```go
DB_USER=root
DB_PASSWORD=123456
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=wallpaper
JWT_SECRET=your-secret-key
SMTP_HOST=smtp.qq.com
SMTP_PORT=587
SMTP_USER=your-email@qq.com
SMTP_PASSWORD=your-auth-code
```

### 2. 安装依赖

    go mod tidy

### 3.运行服务

     go run cmd/service/main.go

服务默认运行在 http://localhost:8080

## API文档
- **Base URL:** http://localhost:8080
- **认证方式:** JWT Token（需要登录的接口在 Header 中携带 Authorization: Bearer <token>）

## 公开路由（不需要认证）
### 1.用户注册
POST /register

**请求参数:**

| 参数  | 类型  | 必填  |  说明 |
| ------------ | ------------ | ------------ | ------------ |
| username  |  string |  是 |  用户名，3-20字符 |
| password  |  string | 是  | 密码，至少6位  |
| confirm_password  |  string |  是 | 确认密码  |
| email  | string  |  是 | 邮箱  |
| code  |  string | 是  | 验证码  |

**请求示例：**

```json
{
    "username": "张三",
    "password": "abc123",
    "confirm_password": "abc123",
    "email": "user@example.com",
    "code": "123456"
}
```

**响应示例:**

```json
{
    "message": "注册成功",
    "user_id": 1
}
```

### 2. 用户登录

POST /login

**请求参数:**

| 参数  | 类型  |  必填 | 说明  |
| ------------ | ------------ | ------------ | ------------ |
| email  | string  |  是 | 邮箱  |
| password  |  string |  是 |  密码 |

**请求示例：**

```json
{
    "email": "user@example.com",
    "password": "abc123"
}
```

**响应示例:**

```json
{
    "message": "登录成功",
    "user_id": 1,
    "username": "张三",
    "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 3. 发送验证码

POST /send-code

**请求参数:**

|  参数 | 类型  | 必填  |  说明 |
| ------------ | ------------ | ------------ | ------------ |
| email  |  string |  是 | 邮箱  |

**请求示例：**

```json
{
    "email": "user@example.com"
}
```

**响应示例:**

```json
{
    "message": "验证码已发送"
}
```

### 4. 动态列表（广场）

GET /posts

**请求参数:**

| 参数  | 类型  | 必填  |  默认值 |  说明 |
| ------------ | ------------ | ------------ | ------------ | ------------ |
|  page |  int |  否 | 1  |  页码 |
|  page_size |  int |  否 |  10 | 每页条数  |

**响应示例:**

```json
{
    "total": 10,
    "page": 1,
    "page_size": 10,
    "list": [
        {
            "id": 1,
            "user": { "id": 2, "name": "张三" },
            "type": "text",
            "content": "今天心情很好！",
            "media_urls": null,
            "video_url": "",
            "likes": 5,
            "comments_count": 2,
            "views": 100,
            "created_at": "2026-04-20T10:00:00Z"
        }
    ]
}
```

### 5.动态详情

GET /post/:id

**响应示例:**
```json
{
    "id": 1,
    "user": { "id": 2, "name": "张三" },
    "type": "image",
    "content": "海边日落",
    "media_urls": ["/uploads/1.jpg", "/uploads/2.jpg"],
    "video_url": "",
    "likes": 10,
    "comments_count": 3,
    "views": 128,
    "created_at": "2026-04-20T10:00:00Z"
}
```

### 6.评论列表

GET /post/:id/comments

**请求参数:**

| 参数  | 类型  | 必填  |  默认值 | 说明  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| page  | int  | 否  | 1  | 页码  |
|  page_size | int  |  否 | 10  | 每页条数  |

**响应示例:**
```json
{
    "total": 5,
    "page": 1,
    "page_size": 10,
    "list": [
        {
            "id": 1,
            "user": { "id": 3, "name": "李四" },
            "content": "写得太好了！",
            "created_at": "2026-04-20T10:00:00Z"
        }
    ]
}
```

### 需要认证的路由

**Header:** Authorization: Bearer <token>

**用户管理**

|  方法 | URL  | 说明  |
| ------------ | ------------ | ------------ |
|  GET | DELETE  |  用户列表 |
|  GET |  /user/:id | 获取用户信息  |
| PUT  | /user/:id  |  更新用户信息 |
|  DELETE |  /user/:id | 删除用户  |
|  PUT | /user/password  | 修改密码  |
| PUT  |/user/email   |  修改邮箱 |
|  DELETE |  /user/account |  注销账号 |

**动态管理**

|  方法 | URL  |  说明 |
| ------------ | ------------ | ------------ |
| POST  | /post  | 发布动态  |
| DELETE  | /post/:id  | 删除动态  |
|  GET | /user/:id/posts  | 用户的动态列表  |
|  PUT |  /post/:id | 更新动态设置  |



**点赞**

| 方法  |  URL | 说明  |
| ------------ | ------------ | ------------ |
| POST  |  /post/:id/like |  点赞 |
| DELETE  |  /post/:id/like | 取消点赞  |
| GET  | /post/:id/likes  | 动态的点赞用户列表  |
| GET  |/user/:id/likes   | 用户的点赞列表  |

**收藏**

| 方法  |  URL |  说明 |
| ------------ | ------------ | ------------ |
| POST  | /post/:id/favorite  | 收藏  |
| DELETE  | /post/:id/favorite  | 取消收藏  |
| GET  | /post/:id/favorites  |  动态的收藏用户列表 |
| GET  | /user/:id/favorites  |  用户的收藏列表 |

**评论**

| 方法  | URL  | 说明  |
| ------------ | ------------ | ------------ |
| POST  | /post/:id/comment  | 发表评论  |
| DELETE  |  /comment/:id | 删除评论  |



