package ginfmt

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sleagon/ginfmt/errfmt"
)

var logger Logger = logrus.New()

const respKey = "$$X_GINFMT_RESP_KEY$$"

// Handler ginfmt handler
// the interface{} will be set to data field in response body
// the error will be transformed to some specific response body
type Handler func(c *gin.Context) (interface{}, error)

type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

// Init init ginfmt with log and i18n translator
func Init(log Logger, trans errfmt.Translator) {
	if log != nil {
		logger = log
	}
	if trans != nil {
		errfmt.Init(trans)
	}
}

// Data set data field of response
func Data(c *gin.Context, data interface{}) {
	body, ok := c.Get(respKey)
	if ok && body != nil {
		logger.Warnf("Response body has been written")
		return
	}
	c.Set(respKey, data)
}

// Error set response error
func Error(c *gin.Context, err error) {
	if err == nil {
		return
	}
	c.Error(err) // nolint: errcheck
}

// DataError Set response data and error at the same time.
func DataError(c *gin.Context, data interface{}, err error) {
	if data != nil {
		Data(c, data)
	}
	if err != nil {
		Error(c, err)
	}
}

// NewHandlderFunc build a gin#HandlerFunc based on Handler
func NewHandlerFunc(h Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		d, e := h(c)
		DataError(c, d, e)
	}
}
