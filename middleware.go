package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const invalidTokenMsg = "invalid token"

func Identify(s Session, customFuncs ...func(s Session)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := ctx.Cookie(cookieKeyAccessToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, invalidTokenMsg)
			ctx.Abort()
			return
		}

		if err = s.ParseToken(ctx, tokenString); err == nil {
			s.SetIntoGinCtx(ctx)
			for _, f := range customFuncs {
				f(s)
			}
			ctx.Next()
			return
		}
		ctx.JSON(http.StatusUnauthorized, invalidTokenMsg)
		ctx.Abort()
		return

	}
}
