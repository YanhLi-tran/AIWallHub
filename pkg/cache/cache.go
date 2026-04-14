package cache

import (
	"sync"
	"time"
)

var (
	store = make(map[string]cacheItem)
	mu    sync.RWMutex
)

type cacheItem struct {
	code     string
	expireAt time.Time
}

// Set 存储验证码，5分钟过期
func Set(email, code string) {
	mu.Lock()
	defer mu.Unlock()
	store[email] = cacheItem{
		code:     code,
		expireAt: time.Now().Add(5 * time.Minute),
	}
}

// Get 获取验证码
func Get(email string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	item, exists := store[email]
	if !exists || time.Now().After(item.expireAt) {
		return "", false
	}
	return item.code, true
}

// Delete 删除验证码
func Delete(email string) {
	mu.Lock()
	defer mu.Unlock()
	delete(store, email)
}
