package user

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("lucario secret key")

type Claims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	jwt.StandardClaims
}

// 生成token
func generateToken(username, password string) (string, error) {
	claims := Claims{
		username,
		password,
		jwt.StandardClaims{
			Issuer: "lucario.ltd",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)

	return token, err
}

// 验证token
func parseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}

// 用户注册
func register(user *RegisterForm) error {
	// 查找数据库
	_, err := userDB.Table("user").Insert(user)
	return err
}

// 用户登录
func login(form *LoginForm) (string, error) {
	// 查找数据库
	user := &LoginForm{}
	exist, err := userDB.Table("user").Where("username=?", form.Username).Get(user)

	if !exist {
		return "", errors.New("用户不存在")
	}
	if err != nil {
		return "", errors.New("服务器未知错误")
	}
	// 密码正确
	if form.Password == user.Password {
		token, err := generateToken(user.Username, user.Password)
		// 写入redis

		return token, err
	} else {
		return "", errors.New("密码错误")
	}

	//key, err := parseToken(token)
	//fmt.Println(key, err)
}
