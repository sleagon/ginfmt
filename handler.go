package ginfmt

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger Logger = new(logrus.Entry)

const respKey = "$$X_RESP_KEY$$"

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

func Init(log Logger) {
	if log != nil {
		logger = log
	}
}

func Data(c *gin.Context, data interface{}) {
	body, ok := c.Get(respKey)
	if ok && body != nil {
		logger.Warnf("Response body has been written")
		return
	}
	c.Set(respKey, data)
}

func Error(c *gin.Context, err error) {
	if err == nil {
		return
	}
	c.Error(err)
}

func DataError(c *gin.Context, data interface{}, err error) {
	if data != nil {
		Data(c, data)
	}
	if err != nil {
		Error(c, err)
	}
}
