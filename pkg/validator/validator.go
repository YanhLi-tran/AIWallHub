package validator

import (
	"mime/multipart"
	"strings"
)

// 不同场景的限制
var Limits = map[string]struct {
	MaxSize int64
	Exts    map[string]bool
}{
	//上传头像
	"avatar": {
		MaxSize: 2 << 20, // 2 MB
		Exts: map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		},
	},
	//上传帖子图片
	"post_image": {
		MaxSize: 10 << 20, // 10 MB
		Exts: map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		},
	},
	//上传壁纸
	"wallpaper": {
		MaxSize: 20 << 20, // 20 MB
		Exts: map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		},
	},
	//上传表情包
	"gif": {
		MaxSize: 5 << 20, // 5 MB
		Exts: map[string]bool{
			".gif": true,
		},
	},
}

// ValidateImage 验证图片
func ValidateImage(file *multipart.FileHeader, scene string) (bool, string) {
	limit, ok := Limits[scene]
	if !ok {
		return false, "无效的场景"
	}

	// 检查文件大小
	if file.Size > limit.MaxSize {
		return false, "图片大小超过限制"
	}

	// 检查文件类型
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return false, "只支持图片文件"
	}

	// 检查扩展名
	ext := strings.ToLower(file.Filename[strings.LastIndex(file.Filename, "."):])
	if !limit.Exts[ext] {
		return false, "不支持的图片格式"
	}

	return true, ""
}

// 视频场景限制
var VideoLimits = map[string]struct {
	MaxSize int64
	Exts    map[string]bool
}{
	"post_video": {
		MaxSize: 100 << 20, // 100 MB
		Exts: map[string]bool{
			".mp4":  true,
			".mov":  true,
			".avi":  true,
			".mkv":  true,
			".webm": true,
		},
	},
}

// ValidateVideo 验证视频
func ValidateVideo(file *multipart.FileHeader, scene string) (bool, string) {
	limit, ok := VideoLimits[scene]
	if !ok {
		return false, "无效的场景"
	}

	if file.Size > limit.MaxSize {
		return false, "视频大小超过限制"
	}

	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "video/") {
		return false, "只支持视频文件"
	}

	ext := strings.ToLower(file.Filename[strings.LastIndex(file.Filename, "."):])
	if !limit.Exts[ext] {
		return false, "不支持的视频格式"
	}

	return true, ""
}
