package auth

import "github.com/gin-gonic/gin"

const (
	cookieKeyAccessToken = "access_token"
	cookieMaxAge         = 60 * 60 * 48
	pathRoot             = "/"
)

func SetCookie(ctx *gin.Context, tokenStr string) {
	ctx.SetCookie(cookieKeyAccessToken, tokenStr, cookieMaxAge, pathRoot, ctx.Request.URL.Host, false, false)
}

func UnSetCookie(ctx *gin.Context) {
	ctx.SetCookie(cookieKeyAccessToken, "", -1, pathRoot, ctx.Request.URL.Host, false, false)
}

func GetTokenFromCookie(ctx *gin.Context) (string, error) {
	return ctx.Cookie(cookieKeyAccessToken)
}
