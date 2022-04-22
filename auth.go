package auth

import (
	"context"
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

// Session is the session data.
// Ext this struct to add your own data.
type Session struct {
	Ext interface{}
	jwt.StandardClaims
}

func NewSession(ext interface{}) *Session {
	return &Session{
		Ext: ext,
	}
}

func (s *Session) GenerateAccessToken() (string, error) {
	s.ExpiresAt = time.Now().Add(AccessTokenExpireDuration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, s)
	return token.SignedString(secret)
}

func (s *Session) ParseToken(ctx context.Context, tokenStr string) error {
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
		vt := reflect.TypeOf(data)
		vv := reflect.ValueOf(data)
		if vt.Kind() == reflect.Ptr {
			vt = vt.Elem()
			vv = vv.Elem()
		}

		for i := 0; i < vt.NumField(); i++ {
			if !vv.Field(i).CanSet() {
				continue
			}
			f := vt.Field(i)
			k := f.Tag.Get("json")
			if k == "" {
				k = f.Name
			}
			m, ok := s.Ext.(map[string]interface{})
			if !ok {
				continue
			}
			if _, ok = m[k]; !ok {
				continue
			}

			dataType := reflect.TypeOf(m[k])
			structType := vv.Field(i).Type()
			if structType == dataType {
				vv.Field(i).Set(reflect.ValueOf(m[k]))
				continue
			}
			if dataType.ConvertibleTo(structType) {
				vv.Field(i).Set(reflect.ValueOf(m[k]).Convert(structType))
			}

		}

		if reflect.TypeOf(data).Kind() == reflect.Ptr {
			data = reflect.ValueOf(data).Elem().Interface()
		}
		s.Ext = data
	}

	return nil
}

func (s *Session) SetIntoGinCtx(ctx *gin.Context) {
	val := context.WithValue(context.Background(), ValueContextKey, s.Ext)
	ctx.Request = ctx.Request.WithContext(val)
}

func (s Session) GetFromGinCtx(ctx *gin.Context) (interface{}, error) {
	sess := ctx.Request.Context().Value(ValueContextKey)
	if sess == nil {
		return nil, ErrNoAuthenticationData
	}
	return sess, nil
}

func (s Session) GetExtFromCtx(ctx context.Context) (interface{}, error) {
	return GetAuthenticationData(ctx)
}

func GetAuthenticationData(ctx context.Context) (interface{}, error) {
	sess := ctx.Value(ValueContextKey)
	if sess == nil {
		return nil, ErrNoAuthenticationData
	}
	return sess, nil
}

func GetAuthenticationDataFrom(ctx *gin.Context) (interface{}, error) {
	return GetAuthenticationData(ctx.Request.Context())
}

func SetCookie(ctx *gin.Context, tokenStr string) {
	ctx.SetCookie(cookieKeyAccessToken, tokenStr, cookieMaxAge, pathRoot, ctx.Request.URL.Host, false, false)
}

func UnSetCookie(ctx *gin.Context) {
	ctx.SetCookie(cookieKeyAccessToken, "", -1, "/", ctx.Request.URL.Host, false, false)
}
