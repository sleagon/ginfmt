package ginfmt

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sleagon/ginfmt/errfmt"
)

// parse a valid error from gin.Context
func parseError(c *gin.Context) *errfmt.Error {
	if len(c.Errors) < 1 {
		return errfmt.NilError()
	}
	for _, err := range c.Errors {
		if e, ok := err.Err.(*errfmt.Error); ok && e != nil {
			return e
		}
	}
	return errfmt.UnknownError()
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
		locale := getLocale(c)
		err := parseError(c)
		resp := Resp{
			Code:    err.Code(),
			Message: err.Message(ctx, locale),
			Data:    c.Value(respKey),
		}
		log := logger.Infof
		switch err.Level() {
		case errfmt.LevelDebug:
			log = logger.Debugf
		case errfmt.LevelInfo:
			log = logger.Infof
		case errfmt.LevelWarn:
			log = logger.Warnf
		case errfmt.LevelError:
			log = logger.Errorf
		default:
		}
		log("request recorded code = %d, message = %s, http status = %d", resp.Code, resp.Message, err.HttpStatus())
		c.JSON(err.HttpStatus(), resp)
	}
}
