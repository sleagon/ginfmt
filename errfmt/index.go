package errfmt

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type Level int

var (
	UnknownError = Register(http.StatusInternalServerError, 1, "unknown error")
	NilError     = Register(http.StatusOK, 0, "ok")
)

const (
	LevelError Level = 4
	LevelWarn  Level = 5
	LevelInfo  Level = 6
	LevelDebug Level = 7
)

type Translator func(ctx context.Context, locate string, key string) string

var (
	translator = EchoTrans
)

func Init(trans Translator) {
	if trans != nil {
		translator = trans
	}
}

func EchoTrans(ctx context.Context, locate string, key string) string {
	return key
}

type Error struct {
	level      Level
	code       int
	httpStatus int
	message    string
	args       []interface{}
}

func (e *Error) Level() Level {
	return e.level
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d|%d|%s|%v", e.code, e.httpStatus, e.message, e.args)
}

func (e Error) Message(ctx context.Context, locate string) string {
	// avoid unnecessary translation
	if e.code == 0 {
		return e.message
	}
	msgFmt := translator(ctx, locate, e.message)
	if len(e.args) < 1 {
		return msgFmt
	}
	return fmt.Sprintf(msgFmt, e.args...)
}

func (e Error) HttpStatus() int {
	return e.httpStatus
}

func (e Error) Code() int {
	return e.code
}

type ErrGen func(args ...interface{}) *Error

// Register add a new error generator
func Register(httpStatus int, code int, message string, opts ...interface{}) ErrGen {
	level := LevelError
	// set info level if http status code is less than 500
	if httpStatus < http.StatusInternalServerError {
		level = LevelInfo
	}
	for _, opt := range opts {
		switch v := opt.(type) {
		case Level:
			level = v
		default:
			log.Panicf("Invalid opt %v", opt)
		}
	}
	return func(args ...interface{}) *Error {
		return &Error{
			level:      level,
			code:       code,
			httpStatus: httpStatus,
			message:    message,
			args:       args,
		}
	}
}
