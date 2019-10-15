package ginfmt

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sleagon/ginfmt/errfmt"
)

var logger Logger = logrus.New()

const respKey = "$$X_GINFMT_RESP_KEY$$"
const stackKey = "$$X_GINFMT_STACK_KEY$$"

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
	setStack(c)
	body, ok := c.Get(respKey)
	if ok && body != nil {
		logger.Warnf("Response body has been written")
		return
	}
	c.Set(respKey, data)
}

// Error set response error
func Error(c *gin.Context, err error) {
	setStack(c)
	if err == nil {
		return
	}
	c.Error(err)
}

// DataError Set response data and error at the same time.
func DataError(c *gin.Context, data interface{}, err error) {
	setStack(c)
	if data != nil {
		Data(c, data)
	}
	if err != nil {
		Error(c, err)
	}
}

type stack struct {
	file       string
	line       int
	stacktrace string
}

// setStack set error stack
func setStack(c *gin.Context) {
	s := stack{}
	_, s.file, s.line, _ = runtime.Caller(2)
	trace := make([]byte, 1<<10)
	if l := runtime.Stack(trace, false); l > 0 {
		s.stacktrace = string(trace[:l])
		c.Set(stackKey, s)
	}
}
