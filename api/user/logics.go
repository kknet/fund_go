package user

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
	"xorm.io/xorm"
)

var (
	jwtSecret = []byte("lucario secret key")
	userDB    = connectDB()
)

type Claims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	jwt.StandardClaims
}

type User struct {
	Username string `xorm:"username"`
	Password string `xorm:"password"`
	Phone    string `xorm:"phone"`
	Email    string `xorm:"email"`
	Points   int    `xorm:"points"`
}

func connectDB() *xorm.Engine {
	connStr := "postgres://postgres:123456@127.0.0.1:5432/user?sslmode=disable"
	db, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

// 生成token
func generateToken(username, password string) (string, error) {
	// 查找数据库
	res, err := userDB.Table("user").QueryString()
	fmt.Println(res, err)

	nowTime := time.Now()

	claims := Claims{
		username,
		password,
		jwt.StandardClaims{
			IssuedAt: nowTime.Unix(),
			Issuer:   "lucario.ltd",
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

// Register 用户注册
func Register() {
	username := "lucario"
	password := "n3dgu5fccv"
	token, err := generateToken(username, password)
	fmt.Println(token, err)

	key, err := parseToken(token)
	fmt.Println(key, err)
}

// Login 用户登录
func Login() {

}
