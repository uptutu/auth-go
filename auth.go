package auth

import (
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	AccessTokenExpireDuration = time.Hour * 24
	ValueContextKey           = "sess"

	defaultSecretKey = "secret"
)

var (
	ErrNotInvalidToken      = errors.New("token is not invalid")
	ErrNoAuthenticationData = errors.New("no authentication data")
	ErrTryToSetInvalidData  = errors.New("try to set invalid data")

	// use for generate and parse JWT
	_secret = []byte(defaultSecretKey)
)

// Init initializes the auth package if you need.
func Init(secret string) error {
	SetJWTSecret(secret)
	return nil
}

// SetJWTSecret set the secret key for JWT.
func SetJWTSecret(s string) {
	_secret = []byte(s)
}

type Session interface {
	GenerateAccessToken() (string, error)
	ParseToken(ctx context.Context, token string) error
	IsParsed() bool
	SetExtIntoGinContext(ctx *gin.Context) error
	GetExtFromGinContext(ctx *gin.Context) (interface{}, error)
	GetToken(ctx *gin.Context) (string, error)
}

// DefaultSession is the session data.
// Ext this struct to add your own data.
type DefaultSession struct {
	Ext interface{}
	jwt.StandardClaims
}

func NewSession(ext interface{}) Session {
	if s, ok := ext.(Session); ok {
		return s
	}

	return &DefaultSession{
		Ext: ext,
	}
}

func (s *DefaultSession) GenerateAccessToken() (string, error) {
	s.ExpiresAt = time.Now().Add(AccessTokenExpireDuration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, s)
	return token.SignedString(_secret)
}

// ParseToken parses the token and converts it to what type you want.
func (s *DefaultSession) ParseToken(ctx context.Context, tokenStr string) error {
	// 找到设置的扩展数据设置的结构体类型
	// 将用户设置的类型保存到 data 临时变量中
	var data interface{}
	if s.Ext != nil {
		vt := reflect.TypeOf(s.Ext)
		for vt.Kind() == reflect.Ptr {
			vt = vt.Elem()
		}

		if vt.Kind() == reflect.Struct {
			data = reflect.New(vt).Interface()
		}
	}

	// 解析 Token
	token, err := jwt.ParseWithClaims(tokenStr, s, func(token *jwt.Token) (interface{}, error) {
		return _secret, nil
	})
	if err != nil && !strings.HasPrefix(err.Error(), "token is expired by") {
		log.Print(ctx, "parse token err : %s", err)
		return err
	}

	if !token.Valid {
		return ErrNotInvalidToken
	}

	// 解析 Token 时拓展字段 interface 类型的数据会被解析到的 map[string]interface{} 覆盖
	// 如果解析出来的数据不是 map[string]interface{} 那么说明解析到的数据已经解析到了用户自定义的类型中
	if _, ok := s.Ext.(map[string]interface{}); !ok {
		if reflect.TypeOf(s.Ext).Kind() == reflect.Struct {
			return nil
		}
		if reflect.TypeOf(s.Ext).Kind() == reflect.Ptr &&
			reflect.TypeOf(s.Ext).Elem().Kind() == reflect.Struct && !reflect.ValueOf(s.Ext).IsNil() {
			s.Ext = reflect.ValueOf(s.Ext).Elem().Interface()
			return nil
		}
	}

	// 如果 Ext 拓展类型被取到
	if data != nil {
		b, err := json.Marshal(s.Ext)
		if err != nil {
			return errors.Wrap(err, "json marshal parsed token data error")
		}

		if err = json.Unmarshal(b, data); err != nil {
			return errors.Wrap(err, "json unmarshal parsed token data to given struct error")
		}

		// Always let Ext field set struct instance
		for reflect.ValueOf(data).Type().Kind() == reflect.Ptr {
			data = reflect.ValueOf(data).Elem().Interface()
		}

		s.Ext = data
	}

	return nil
}

func (s *DefaultSession) SetExtIntoGinContext(ctx *gin.Context) error {
	if reflect.ValueOf(s.Ext).IsNil() ||
		reflect.ValueOf(s.Ext).IsZero() {
		return ErrTryToSetInvalidData
	}
	ctx.Set(ValueContextKey, s.Ext)
	return nil
}

func (s DefaultSession) GetExtFromGinContext(ctx *gin.Context) (interface{}, error) {
	sess := ctx.Request.Context().Value(ValueContextKey)
	if sess == nil {
		return nil, ErrNoAuthenticationData
	}
	return sess, nil
}

func (s DefaultSession) GetToken(ctx *gin.Context) (string, error) {
	return GetTokenFromCookie(ctx)
}

func (s DefaultSession) IsParsed() bool {
	if s.Ext != nil {
		return true
	}
	return false
}

func GetAuthenticationDataFrom(ctx *gin.Context, key string) (interface{}, error) {
	val, ok := ctx.Get(key)
	if !ok {
		return nil, ErrNoAuthenticationData
	}
	return val, nil
}
