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

// User 用户表model
type User struct {
	Id       int       `xorm:"int(10) pk not null autoincr"`
	Username string    `xorm:"varchar(32) unique not null"`
	Password string    `xorm:"varchar(32) not null"`
	Phone    string    `xorm:"char(11) unique"`
	Email    string    `xorm:"varchar(32) unique"`
	Points   int       `xorm:"not null default 0"`
	Created  time.Time `xorm:"created"`
}

// LoginForm 登录表单
type LoginForm struct {
	Username string `xorm:"username"`
	Password string `xorm:"password"`
}

// RegisterForm 注册表单
type RegisterForm struct {
	Username string `xorm:"username"`
	Password string `xorm:"password"`
	//Phone    string `xorm:"phone"`
	//Email    string `xorm:"email"`
	Created time.Time `xorm:"created"`
}

// UserInfo 用户信息
type UserInfo struct {
	Id       int
	Username string
	Phone    string
	Email    string
	Points   int
}

// 尝试建表
func init() {
	err := userDB.Sync2(new(User))
	if err != nil {
		log.Println(err)
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
