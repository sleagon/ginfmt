# GINFMT


[![Build Status](https://travis-ci.org/sleagon/ginfmt.svg?branch=master)](https://travis-ci.org/sleagon/ginfmt)  [![Go Report Card](https://goreportcard.com/badge/github.com/sleagon/ginfmt)](https://goreportcard.com/report/github.com/sleagon/ginfmt)  [![GoDoc](https://godoc.org/github.com/sleagon/ginfmt?status.svg)](https://godoc.org/github.com/sleagon/ginfmt)  [![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)


Go version >=1.13

```bash
go get github.com/sleagon/ginfmt
```

## Usage

ginfmt is a simple toolkit to format response of gin server.

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
ginfmt.Error(c, BadRequest())
{
	"code": 10010,
	"message": "record not found",
	"data": nil
}
// abnormal response with payload
ginfmt.DataError(c, "foo", BadRequest())
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
		DataError(c, "bar", FooError())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ginfmt", nil)
	r.ServeHTTP(w, req)
	resp := new(Resp3)
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, resp.Code, FooError().Code())
	assert.Equal(t, resp.Message, FooError().Message(context.TODO(), "zh"))
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
		err := fmt.Errorf("%w, extra info: test info", FooError())
		Error(c, err)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ginfmt", nil)
	r.ServeHTTP(w, req)
	resp := new(Resp2)
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, resp.Code, FooError().Code())
	assert.Equal(t, resp.Message, FooError().Message(context.TODO(), ""))
	assert.Equal(t, 0, resp.Data)
}
```

## A runnable example

Here is a runnable example.

```go
package main

import (
	"fmt"
	"net/http"

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
		ginfmt.Error(c, BadRequest())
	})
	r.GET("/ping", func(c *gin.Context) {
		ginfmt.Data(c, "pong")
	})
	r.GET("/bad_payload", func(c *gin.Context) {
		// do sth
		err := fmt.Errorf("this is not a valid phone num %w", BadRequest())
		ginfmt.DataError(c, gin.H{"phone": "invalid", "email": "valid"}, err)
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
```

