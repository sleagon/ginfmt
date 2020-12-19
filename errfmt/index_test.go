package errfmt

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEchoTrans(t *testing.T) {
	msg := EchoTrans(context.Background(), "zh", "foo")
	assert.Equal(t, msg, "foo")
}

func TestErrorGen(t *testing.T) {
	errNotFound := Register(http.StatusNotFound, 1009, "this is a test message", LevelError)
	err := errNotFound.Gen()
	_, file, line, _ := runtime.Caller(0)
	bl := strings.HasPrefix(err.Error(), fmt.Sprintf("[%s:%d]1009|404|this is a test message|[]", file, line-1))
	assert.True(t, bl)
}

func TestErrorWithArgs(t *testing.T) {
	infoErr := Register(http.StatusOK, 20001, "%v is a invalid name")
	err := infoErr.Gen("foo")
	assert.Equal(t, "foo is a invalid name", err.Message(context.TODO(), ""))
}

func TestErrorIs(t *testing.T) {
	infoErr := Register(http.StatusOK, 20001, "%v is a invalid name")
	err := infoErr.Gen("foo")
	nerr := fmt.Errorf("%w balabalababa info", err)
	assert.Equal(t, true, infoErr.Is(nerr), "")
}
