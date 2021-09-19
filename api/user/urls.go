package user

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Register 用户注册
func Register(c *gin.Context) {
	data := &RegisterForm{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
		//Phone: c.PostForm("phone"),
		//Email: c.PostForm("email"),
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

// Login 用户登录
func Login(c *gin.Context) {
	data := &LoginForm{
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
