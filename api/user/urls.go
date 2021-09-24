package user

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
)

// Authorize 验证中间件
func Authorize(c *gin.Context) {
	token := c.GetHeader("token")
	claims, err := parseToken(token)

	if err != nil {
		c.JSON(http.StatusUnauthorized, bson.M{
			"status": false, "msg": "请先登录",
		})
		c.Abort()
	} else {
		// 用户id
		c.Set("id", claims.Id)
	}
}

// Register 用户注册
func Register(c *gin.Context) {
	data := &registerForm{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}
	phone, ok := c.GetPostForm("phone")
	if ok {
		data.Phone = phone
	}
	email, ok := c.GetPostForm("email")
	if ok {
		data.Email = email
	}

	err := register(data)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false, "msg": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true, "msg": "注册成功",
	})
}

// GetInfo 查看用户信息
func GetInfo(c *gin.Context) {
	id := c.GetInt("id")
	info, err := getInfo(id)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false, "msg": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true, "data": info,
	})
}

// UpdateInfo 更新信息
func UpdateInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": false, "msg": "该接口暂不可用",
	})
}

// Login 登录
func Login(c *gin.Context) {
	data := &loginForm{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}
	token, err := login(data)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false, "msg": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true, "msg": "登录成功", "token": token,
	})
}

// Logout 注销
func Logout(c *gin.Context) {
	id := c.GetInt("id")
	// 删除token
	err := redisDB.Del(ctx, strconv.Itoa(id)).Err()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false, "msg": "注销失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true, "msg": "注销成功",
	})
}
