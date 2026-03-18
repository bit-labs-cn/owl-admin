package jwt

import (
	"time"

	"bit-labs.cn/owl-admin/app/model"
	"github.com/golang-jwt/jwt/v5"
	jsoniter "github.com/json-iterator/go"
)

type JWTOptions struct {
	SigningKey string `json:"signing-key"`
	Expire     int    `json:"expire"`
	Issuer     string `json:"issuer"`
}

type JWTService struct {
	opt JWTOptions
}

func NewJWTService(opt JWTOptions) *JWTService {
	return &JWTService{opt: opt}
}

type UserClaims struct {
	model.User
	jwt.Claims
}

func (i *JWTService) GenerateToken(u *model.User) (string, error) {
	expire := time.Now().Add(time.Second * time.Duration(i.opt.Expire))
	claims := UserClaims{
		User: *u,
		Claims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			Issuer:    i.opt.Issuer,
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte(i.opt.SigningKey))

	return token, err
}

func (i *JWTService) ParseToken(token string) (u *model.User, err error) {
	claim, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(i.opt.SigningKey), nil
	})
	if err != nil || claim == nil {
		return nil, err
	}
	c := claim.Claims.(jwt.MapClaims)

	toString, err := jsoniter.MarshalToString(c)
	if err != nil {
		return nil, err
	}
	u = &model.User{}
	err = jsoniter.Unmarshal([]byte(toString), u)
	if err != nil {
		return nil, err
	}
	return u, err
}
