# GINFMT

[中文说明](Readme.CN.md)

[![Build Status](https://travis-ci.org/sleagon/ginfmt.svg?branch=master)](https://travis-ci.org/sleagon/ginfmt)  [![Go Report Card](https://goreportcard.com/badge/github.com/sleagon/ginfmt)](https://goreportcard.com/report/github.com/sleagon/ginfmt)  [![GoDoc](https://godoc.org/github.com/sleagon/ginfmt?status.svg)](https://godoc.org/github.com/sleagon/ginfmt)  [![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)


Go version >=1.13

```bash
go get github.com/sleagon/ginfmt
```

## Usage

ginfmt is a simple toolkit to format response of gin server.

Notice:

The `ErrGen` is a simple function in version 1.0.3, which is complicated to validate whether a error is wrapped from a
known ErrGen. So we just redesigned the ErrGen to a struct, this struct will help us to do the magic quite easily.

Upgrade:

Just replace all `XXXErr()` with `XXXErr.Gen`.

### Example

```GO
// default logger(logrus) and default translator (echo)
ginfmt.Init(nil, nil)
BadRequest := errfmt.Register(http.StatusNotFound, 10010, "record not found")


// normal response
ginfmt.Data(c, "foo")
{
	"code": 0,
	"messsage": "ok",
	"data": "foo"
}
// abnormal response
ginfmt.Error(c, BadRequest.Gen())
{
	"code": 10010,
	"message": "record not found",
	"data": nil
}
// abnormal response with payload
ginfmt.DataError(c, "foo", BadRequest.Gen())
{
	"code": 10010,
	"message": "record not found",
    "data": "foo"
}
```

## I18N

Most of time, error message should be translated to perticular language. You need define a translator like this:

```GO
func DemoTrans(ctx context.Context, locale string, key string) string {
	demoMap := map[string]map[string]string{
		"zh": map[string]string{
			"foo": "这是一个foo信息",
		},
		"en-US": map[string]string{
			"foo": "This is foo message",
		},
	}
	return demoMap[locale][key]
}

func TestI18n(t *testing.T) {
	Init(nil, DemoTrans)
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo")
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Request.Header.Set("locale", "zh")
	})
	r.Use(MW())
	r.GET("/ginfmt", func(c *gin.Context) {
		DataError(c, "bar", FooError.Gen())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ginfmt", nil)
	r.ServeHTTP(w, req)
	resp := new(Resp3)
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, resp.Code, FooError.Gen().Code())
	assert.Equal(t, resp.Message, FooError.Gen().Message(context.TODO(), "zh"))
	assert.Equal(t, "bar", resp.Data)
}
```

ginfmt will read "locale" from query/header/cookie/gin.Context，you may need set this value first.

## Logger

All response handled by ginfmt will be logged according to level of error, by default, error whose http status code is 
less than 500 will be recorded as information, other errors will be recorded as error.

All error returned to users should be pre defined before used.

```GO
    // default level
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message")
	// custom level
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message", errfmt.LevelError)
``` 

## Wrapped error

You may need to add extra log info to error, thanks to `errors.Unwrap` and `fmt.Errorf("%w, dome extra info", err)`
introduced in go 1.13, we can easily do this.
```GO
func TestWrappedError(t *testing.T) {
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message")
	r := gin.Default()
	r.Use(MW())
	r.GET("/ginfmt", func(c *gin.Context) {
		// YOUR ROUTER CODE
		err := fmt.Errorf("%w, extra info: test info", FooError.Gen())
		Error(c, err)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ginfmt", nil)
	r.ServeHTTP(w, req)
	resp := new(Resp2)
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, resp.Code, FooError.Gen().Code())
	assert.Equal(t, resp.Message, FooError.Gen().Message(context.TODO(), ""))
	assert.Equal(t, 0, resp.Data)
}
```

## [New] IS

You may familiar with `github.com/pkg/errors` or `errors` package after 1.17 which provide `errors.Is` method to judge
whether the error is wrapped from another error. In order to check whether a error is generated from a known `ErrGen`,
we add a new method Is to realize that.

```go
func TestErrorIs(t *testing.T) {
	infoErr := Register(http.StatusOK, 20001, "%v is a invalid name")
	err := infoErr.Gen("foo")
	nerr := fmt.Errorf("%w balabalababa info", err)
	assert.Equal(t, true, infoErr.Is(nerr), "")
}
```

## [New] NewHandlerFunc
Some people may not familiar with the chained handlers of gin, we provided another choice `NewHandlerFunc`

```go
	r.GET("/wrapped_handler", ginfmt.NewHandlerFunc(func(c *gin.Context) (interface{}, error) {
		if time.Now().Unix()%10 == 1 {
			return []int{1, 2, 3}, nil
		}
		return nil, BadRequest.Gen()
	}))
```

## A runnable example

Here is a runnable example.

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sleagon/ginfmt"
	"github.com/sleagon/ginfmt/errfmt"
)

var (
	BadRequest = errfmt.Register(http.StatusBadRequest, 10001, "Params is invalid")
)

func main() {
	ginfmt.Init(nil, nil)
	r := gin.Default()
	r.Use(ginfmt.MW())
	r.GET("/bad", func(c *gin.Context) {
		ginfmt.Error(c, BadRequest.Gen())
	})
	r.GET("/ping", func(c *gin.Context) {
		ginfmt.Data(c, "pong")
	})
	r.GET("/bad_payload", func(c *gin.Context) {
		// do sth
		err := fmt.Errorf("this is not a valid phone num %w", BadRequest.Gen())
		ginfmt.DataError(c, gin.H{"phone": "invalid", "email": "valid"}, err)
	})
	r.GET("/wrapped_handler", ginfmt.NewHandlerFunc(func(c *gin.Context) (interface{}, error) {
		if time.Now().Unix()%10 == 1 {
			return []int{1, 2, 3}, nil
		}
		return nil, BadRequest.Gen()
	}))

	r.Run() // listen and serve on 0.0.0.0:8080
}
```

