package user

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator"
	"strconv"
	"time"
)

var (
	jwtSecret = []byte("lucario website secret")
	ctx       = context.Background()
)

type Claims struct {
	Id       int
	Username string
	Password string
	jwt.StandardClaims
}

// 生成token
func generateToken(form *loginForm) (string, error) {
	nowTime := time.Now()

	claims := Claims{
		form.Id,
		form.Username,
		form.Password,
		jwt.StandardClaims{
			IssuedAt:  nowTime.Unix(),
			ExpiresAt: nowTime.Add(time.Minute * 76).Unix(),
			Issuer:    "lucario.ltd",
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
		// 解密token
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {

			res, _ := redisDB.Get(ctx, strconv.Itoa(claims.Id)).Result()

			if res == token {
				return claims, nil
			} else {
				return nil, errors.New("token无效或已过期")
			}
		}
	}
	return nil, err
}

// 用户信息
func getInfo(token string) (*userInfo, error) {
	claims, err := parseToken(token)
	if err != nil {
		return nil, err
	}
	// 查找数据库
	info := &userInfo{}
	exist, err := userDB.Table("user").Where("id=?", claims.Id).Get(info)
	if exist && err == nil {
		return info, nil
	}

	return nil, err
}

// 用户注册
func register(user *registerForm) error {
	validate := validator.New()
	err := validate.Struct(user)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Field() + ":格式错误")
		}
	}
	// 插入
	_, err = userDB.Table("user").Insert(user)
	return err
}

// 用户登录
func login(form *loginForm) (string, error) {
	// 查找数据库
	info := &loginForm{}
	exist, err := userDB.Table("user").Where("username=?", form.Username).Get(info)

	if !exist {
		return "", errors.New("用户不存在")
	} else if err != nil {
		return "", errors.New("服务器未知错误")
	}

	// 密码正确
	if form.Password == info.Password {
		token, err := generateToken(info)

		// 将token写入redis
		redisDB.Set(ctx, strconv.Itoa(info.Id), token, 24*time.Hour)

		return token, err

	} else {
		return "", errors.New("密码错误")
	}
}

// 用户注销
func logout(token string) error {
	claims, err := parseToken(token)
	res, err := redisDB.Get(ctx, strconv.Itoa(claims.Id)).Result()

	// token不存在
	if err != nil {
		return nil
	}

	// 删除token
	if res == token {
		err = redisDB.Del(ctx, strconv.Itoa(claims.Id)).Err()
		return err
	}
	return nil
}
