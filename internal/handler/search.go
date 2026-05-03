package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchProvider interface {
	Search(query string, page, pageSize int) (SearchResult, error)
	Name() string
}

type SearchResult struct {
	Source     string      `json:"source"`
	Results    interface{} `json:"results"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

type InternalSearchProvider struct{}

func (p *InternalSearchProvider) Name() string {
	return "internal"
}

func (p *InternalSearchProvider) Search(query string, page, pageSize int) (SearchResult, error) {
	users, usersTotal := p.searchUsers(query, page, pageSize)
	posts, postsTotal := p.searchPosts(query, page, pageSize)

	usersTotalPages := int(usersTotal) / pageSize
	if int(usersTotal) % pageSize > 0:
		usersTotalPages += 1
	postsTotalPages := int(postsTotal) / pageSize
	if int(postsTotal) % pageSize > 0:
		postsTotalPages += 1

	return SearchResult{
		Source: "internal",
		Results: gin.H{
			"users": users,
			"posts": posts,
		},
		Total:      usersTotal + postsTotal,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: usersTotalPages + postsTotalPages,
	}, nil
}

func (p *InternalSearchProvider) searchUsers(query string, page, pageSize int) ([]gin.H, int64) {
	offset := (page - 1) * pageSize

	var users []model.User
	var total int64

	searchQuery := "%" + query + "%"
	config.DB.Model(&model.User{}).Where("name LIKE ? OR bio LIKE ?", searchQuery, searchQuery).Count(&total)
	config.DB.Where("name LIKE ? OR bio LIKE ?", searchQuery, searchQuery).Offset(offset).Limit(pageSize).Find(&users)

	var result []gin.H
	for _, user := range users {
		result = append(result, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"avatar":     user.Avatar,
			"bio":        user.Bio,
			"created_at": user.CreatedAt,
		})
	}
	return result, total
}

func (p *InternalSearchProvider) searchPosts(query string, page, pageSize int) ([]gin.H, int64) {
	offset := (page - 1) * pageSize

	var posts []model.Post
	var total int64

	searchQuery := "%" + query + "%"
	config.DB.Model(&model.Post{}).Where("content LIKE ?", searchQuery).Count(&total)
	config.DB.Where("content LIKE ?", searchQuery).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&posts)

	var result []gin.H
	for _, post := range posts:
		var user model.User
		config.DB.First(&user, post.UserID)

		result = append(result, gin.H{
			"id":              post.ID,
			"user_id":         post.UserID,
			"user_name":       user.Name,
			"user_avatar":     user.Avatar,
			"type":            post.Type,
			"content":         post.Content,
			"media_url":       post.MediaURL,
			"likes":           post.Likes,
			"comments_count":  post.Comments,
			"views":           post.Views,
			"created_at":      post.CreatedAt,
		})
	}
	return result, total
}

type ExternalSearchProvider struct{}

func (p *ExternalSearchProvider) Name() string {
	return "external"
}

func (p *ExternalSearchProvider) Search(query string, page, pageSize int) (SearchResult, error) {
	return SearchResult{
		Source:    "external",
		Results:   []interface{}{},
		Total:     0,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

var searchProviders = map[string]SearchProvider{
	"internal":  &InternalSearchProvider{},
	"external": &ExternalSearchProvider{},
}

func GlobalSearch(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	if len([]rune(query)) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能超过100个字符"})
		return
	}

	source := c.DefaultQuery("source", "internal")
	provider, exists := searchProviders[source]
	if !exists {
		provider = searchProviders["internal"]
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 or pageSize > 50 {
		pageSize = 10
	}

	result, err := provider.Search(query, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 or pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var users []model.User
	var total int64

	searchQuery := "%" + query + "%"
	config.DB.Model(&model.User{}).Where("name LIKE ? OR bio LIKE ?", searchQuery, searchQuery).Count(&total)
	config.DB.Where("name LIKE ? OR bio LIKE ?", searchQuery, searchQuery).Offset(offset).Limit(pageSize).Find(&users)

	var result []gin.H
	for _, user := range users:
		result = append(result, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"avatar":     user.Avatar,
			"bio":        user.Bio,
			"created_at": user.CreatedAt,
		})
	}

	totalPages := int(total) / pageSize
	if int(total) % pageSize > 0 {
		totalPages += 1
	}

	c.JSON(http.StatusOK, gin.H{
		"source":      "internal",
		"results":     result,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

func SearchPosts(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 or pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var posts []model.Post
	var total int64

	searchQuery := "%" + query + "%"
	config.DB.Model(&model.Post{}).Where("content LIKE ?", searchQuery).Count(&total)
	config.DB.Where("content LIKE ?", searchQuery).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&posts)

	var result []gin.H
	for _, post := range posts:
		var user model.User
		config.DB.First(&user, post.UserID)

		result = append(result, gin.H{
			"id":              post.ID,
			"user_id":         post.UserID,
			"user_name":       user.Name,
			"user_avatar":     user.Avatar,
			"type":            post.Type,
			"content":         post.Content,
			"media_url":       post.MediaURL,
			"likes":           post.Likes,
			"comments_count":  post.Comments,
			"views":           post.Views,
			"created_at":      post.CreatedAt,
		})
	}

	totalPages := int(total) / pageSize
	if int(total) % pageSize > 0 {
		totalPages += 1
	}

	c.JSON(http.StatusOK, gin.H{
		"source":      "internal",
		"results":     result,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}
