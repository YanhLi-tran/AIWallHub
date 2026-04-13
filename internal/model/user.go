package model

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// 后续续改为数据库存储用户信息，当前为内存存储
var (
	Users       = make(map[int]User)    //用ID存储用户信息
	UserByEmail = make(map[string]User) //用Email存储用户信息
	NextID      = 1                     //下一个用户的ID
)
