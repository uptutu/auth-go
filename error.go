package auth

import "github.com/gin-gonic/gin"

func NewGinPrivateError(err error) *gin.Error {
	return &gin.Error{
		Type: gin.ErrorTypePrivate,
		Err:  err,
	}
}
