package ginfmt

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sleagon/ginfmt/error"
)

// parse a valid error from gin.Context
func parseError(c *gin.Context) *error.Error {
	if len(c.Errors) < 1 {
		return error.NilError()
	}
	for _, err := range c.Errors {
		if e, ok := err.Err.(*error.Error); ok && e != nil {
			return e
		}
	}
	return error.UnknownError()
}

// get locale from query/header/cookie
func getLocale(c *gin.Context) string {

	if locale := c.Query("locale"); locale != "" {
		return locale
	}

	if locale := c.GetHeader("locale"); locale != "" {
		return locale
	}
	if locale, err := c.Cookie("locale"); err == nil && locale != "" {
		return locale
	}
	return ""
}

// Resp data template
type Resp struct {
	Code    int
	Message string
	Data    interface{}
}

// MW core middleware to format the response
func MW() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		ctx := context.TODO()
		err := parseError(c)
		locale := getLocale(c)
		resp := Resp{
			Code:    err.Code(),
			Message: err.Message(ctx, locale),
			Data:    c.Value(respKey),
		}
		c.JSON(err.HttpStatus(), resp)
	}
}
