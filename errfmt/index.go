package errfmt

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"
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
	ErrGen
	args  []interface{}
	file  string
	line  int
	stack string
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s:%d]%d|%d|%s|%v\nstack: %s",
		e.file, e.line, e.code, e.httpStatus, e.message, e.args, e.stack)
}

type ErrGen struct {
	level      Level
	code       int
	httpStatus int
	message    string
}

func (e *ErrGen) Level() Level {
	return e.level
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

func (e ErrGen) HttpStatus() int {
	return e.httpStatus
}

func (e ErrGen) Code() int {
	return e.code
}

func (e ErrGen) Gen(args ...interface{}) *Error {
	err := Error{
		ErrGen: e,
		args:   args,
	}
	_, err.file, err.line, _ = runtime.Caller(1)
	trace := make([]byte, 1<<10)
	if l := runtime.Stack(trace, false); l > 0 {
		err.stack = string(trace[:l])
	}
	return &err
}

func (e ErrGen) Is(err error) bool {
	rerr, ok := errors.Unwrap(err).(*Error)
	if !ok || rerr == nil {
		return false
	}
	return e.code == rerr.code
}

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
	return ErrGen{
		level:      level,
		code:       code,
		httpStatus: httpStatus,
		message:    message,
	}
}
