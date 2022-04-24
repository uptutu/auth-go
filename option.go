package auth

import "github.com/gin-gonic/gin"

// OptionFunc function will be called after the token has been parsed
// The Session passed in when called is an implementation that has already been parsed, so you can get the parsed data
type OptionFunc func(*gin.Context, Session)
