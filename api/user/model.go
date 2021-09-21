package user

import (
	"github.com/go-redis/redis/v8"
	"time"
	"xorm.io/xorm"
)

var (
	userDB  *xorm.Engine
	redisDB *redis.Client
)

// 用户表结构
type user struct {
	Id       int       `xorm:"int(10) pk not null autoincr"`
	Username string    `xorm:"varchar(32) unique not null"`
	Password string    `xorm:"varchar(32) not null"`
	Phone    string    `xorm:"char(11) unique"`
	Email    string    `xorm:"varchar(32) unique"`
	Points   int       `xorm:"not null default 0"`
	Created  time.Time `xorm:"created"`
}

// 登录表单
type loginForm struct {
	Id       int
	Username string `xorm:"username" validate:"required,min=4,max=10"`
	Password string `xorm:"password" validate:"required,min=6,max=16"`
}

// 注册表单
// omitempty 空时忽略
type registerForm struct {
	Username string      `xorm:"username" validate:"required,min=4,max=10"`
	Password string      `xorm:"password" validate:"required,min=6,max=16"`
	Phone    interface{} `xorm:"phone" validate:"omitempty,min=11,max=11"`
	Email    interface{} `xorm:"email" validate:"omitempty,email,min=2"`
	Created  time.Time   `xorm:"created"`
}

// 用户信息
type userInfo struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Points   int    `json:"points"`
}

// 初始化数据库
func init() {
	var err error

	// 连接UserDB
	connStr := "postgres://postgres:123456@127.0.0.1:5432/user?sslmode=disable"
	userDB, err = xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	// 建表
	err = userDB.Sync2(new(user))
	if err != nil {
		panic(err)
	}
	// 连接Redis
	redisDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
}
