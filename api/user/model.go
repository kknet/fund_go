package user

import (
	"github.com/go-redis/redis/v8"
	"log"
	"time"
	"xorm.io/xorm"
)

var (
	userDB  = connectUserDB()
	redisDB = connectRedisDB()
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
	Username string `xorm:"username" validate:"required"`
	Password string `xorm:"password" validate:"required"`
}

// 注册表单
// omitempty 空时忽略
type registerForm struct {
	Username string `xorm:"username" validate:"required"`
	Password string `xorm:"password" validate:"required"`
	// Phone    string `xorm:"phone" validate:"omitempty,len=11"`
	// Email    string `xorm:"email" validate:"omitempty,email"`
	Created time.Time `xorm:"created"`
}

// 用户信息
type userInfo struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Points   int    `json:"points"`
}

// 建表
func init() {
	err := userDB.Sync2(new(user))
	if err != nil {
		log.Println("建表失败！", err)
	}
}

// 连接用户数据库
func connectUserDB() *xorm.Engine {
	connStr := "postgres://postgres:123456@127.0.0.1:5432/user?sslmode=disable"
	db, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

// 连接redis
func connectRedisDB() *redis.Client {
	//连接服务器
	c := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	return c
}
