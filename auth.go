package auth

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	AccessTokenExpireDuration = time.Minute * 30
	ValueContextKey           = "sess"

	cookieMaxAge         = 60 * 60 * 48
	cookieKeyAccessToken = "access_token"
	defaultSecretKey     = "secret"
	pathRoot             = "/"
)

var (
	ErrNotInvalidToken      = errors.New("token is not invalid")
	ErrNoAuthenticationData = errors.New("no authentication data")

	// use for generate and parse JWT
	secret = []byte(defaultSecretKey)
)

// Init initializes the auth package if you need.
func Init(secret string) error {
	SetJWTSecret(secret)
	return nil
}

// SetJWTSecret set the secret key for JWT.
func SetJWTSecret(s string) {
	secret = []byte(s)
}

type Session interface {
	ValueContextKey() string
	GenerateAccessToken() (string, error)
	ParseToken(ctx context.Context, token string) error
	IsParsed() bool
	SetIntoGinCtx(ctx *gin.Context) error
	GetFromGinCtx(ctx *gin.Context) (interface{}, error)
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
	return token.SignedString(secret)
}

func (s *DefaultSession) ParseToken(ctx context.Context, tokenStr string) error {
	var data interface{}
	if s.Ext != nil {
		vt := reflect.TypeOf(s.Ext)
		if vt.Kind() == reflect.Ptr {
			vt = vt.Elem()
		}

		if vt.Kind() == reflect.Struct {
			data = reflect.New(vt).Interface()
		}
	}

	token, err := jwt.ParseWithClaims(tokenStr, s, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil && !strings.HasPrefix(err.Error(), "token is expired by") {
		log.Print(ctx, "parse token err : %s", err)
		return err
	}

	if !token.Valid {
		return ErrNotInvalidToken
	}

	if _, ok := s.Ext.(map[string]interface{}); !ok {
		if reflect.TypeOf(s.Ext).Kind() == reflect.Struct {
			return nil
		}
		if reflect.TypeOf(s.Ext).Kind() == reflect.Ptr &&
			reflect.TypeOf(s.Ext).Elem().Kind() == reflect.Struct {
			s.Ext = reflect.ValueOf(s.Ext).Elem().Interface()
			return nil
		}
	}

	if data != nil {
		b, err := json.Marshal(s.Ext)
		if err != nil {
			return errors.Wrap(err, "json marshal parsed token data error")
		}

		if err = json.Unmarshal(b, data); err != nil {
			return errors.Wrap(err, "json unmarshal parsed token data to given struct error")
		}

		// Always let Ext field set struct instance
		if reflect.ValueOf(data).Type().Kind() == reflect.Ptr {
			data = reflect.ValueOf(data).Elem().Interface()
		}

		s.Ext = data
	}

	return nil
}

func (s *DefaultSession) SetIntoGinCtx(ctx *gin.Context) error {
	val := context.WithValue(context.Background(), ValueContextKey, s.Ext)
	ctx.Request = ctx.Request.WithContext(val)
	return nil
}

func (s DefaultSession) GetFromGinCtx(ctx *gin.Context) (interface{}, error) {
	sess := ctx.Request.Context().Value(ValueContextKey)
	if sess == nil {
		return nil, ErrNoAuthenticationData
	}
	return sess, nil
}

func (s DefaultSession) GetExtFromCtx(ctx context.Context) (interface{}, error) {
	return GetAuthenticationData(ctx, s.ValueContextKey())
}

func (s DefaultSession) ValueContextKey() string {
	return ValueContextKey
}

func (s DefaultSession) IsParsed() bool {
	if s.Ext != nil {
		return true
	}
	return false
}

func (s DefaultSession) GetToken(ctx *gin.Context) (string, error) {
	return GetAccessTokenFromCookie(ctx)
}

func GetAuthenticationData(ctx context.Context, key string) (interface{}, error) {
	sess := ctx.Value(key)
	if sess == nil {
		return nil, ErrNoAuthenticationData
	}
	return sess, nil
}

func GetAuthenticationDataFrom(ctx *gin.Context, key string) (interface{}, error) {
	return GetAuthenticationData(ctx.Request.Context(), key)
}

func SetAccessTokenToCookie(ctx *gin.Context, tokenStr string) {
	ctx.SetCookie(cookieKeyAccessToken, tokenStr, cookieMaxAge, pathRoot, ctx.Request.URL.Host, false, false)
}

func UnSetCookie(ctx *gin.Context) {
	ctx.SetCookie(cookieKeyAccessToken, "", -1, "/", ctx.Request.URL.Host, false, false)
}

func GetAccessTokenFromCookie(ctx *gin.Context) (string, error) {
	return ctx.Cookie(cookieKeyAccessToken)
}
