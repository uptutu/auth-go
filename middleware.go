package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const invalidTokenMsg = "invalid token"

// Identify is a middleware that checks for a valid token
// s is the Session implementation with u what you need
// opts is the option functions for you create your advanced handler logic.
func Identify(s Session, opts ...OptionFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := ctx.Cookie(cookieKeyAccessToken)
		if err != nil {
			ctx.Errors = append(ctx.Errors, NewGinPrivateError(err))
			ctx.JSON(http.StatusUnauthorized, invalidTokenMsg)
			ctx.Abort()
			return
		}

		if err = s.ParseToken(ctx, tokenString); err != nil {
			ctx.Errors = append(ctx.Errors, NewGinPrivateError(err))
			ctx.JSON(http.StatusUnauthorized, invalidTokenMsg)
			ctx.Abort()
			return
		}
		if err = s.SetIntoGinCtx(ctx); err != nil {
			ctx.Errors = append(ctx.Errors, NewGinPrivateError(err))
			ctx.JSON(http.StatusUnauthorized, invalidTokenMsg)
			ctx.Abort()
			return
		}

		// custom functions before next
		for _, f := range opts {
			f(ctx, s)
		}
		ctx.Next()
		return

	}
}

func IdentifyAndLogger(s Session, opts ...OptionFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		gin.Logger(),
		Identify(s, opts...),
	}
}
