# GINFMT


[![Build Status](https://travis-ci.org/sleagon/ginfmt.svg?branch=master)](https://travis-ci.org/sleagon/ginfmt)  [![Go Report Card](https://goreportcard.com/badge/github.com/sleagon/ginfmt)](https://goreportcard.com/report/github.com/sleagon/ginfmt)  [![GoDoc](https://godoc.org/github.com/sleagon/ginfmt?status.svg)](https://godoc.org/github.com/sleagon/ginfmt)  [![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)


Go version >=1.13

```bash
go get github.com/sleagon/ginfmt
```

## Usage

ginfmt是一个简单的用来格式化gin服务器输入输出的工具，本身没有任何以来，非常简洁干净。

如果你对你gin服务器有下面的需求，非常适合使用ginfmt：

- 服务器的返回是格式化的，每个错误都有明确的code/message信息。
- 所有的错误都能明确的日志和合适的日志级别。
- 所有的错误都有清晰的错误栈信息，不需要手动打印日志。
- 具备I18N的能力，可以自动处理各种语言的错误文案。


### 例子

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

错误文案的多语言支持。

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

注意：ginfmt会按照下面的顺序读取"locale"字段：query/header/cookie/gin.Context，如果你的业务场景不符合这个标准，建议自行家个中间件提前设置好。

## Logger

ginfmt的所有错误都会按照错误级别设置错误日志的级别，默认地500以上的返回都会一error级别打印日志，低于500的返回的则打印为info级别。注意：所有的日志都应该是预先定义的，不应该采用动态的错误。

```GO
    // default level
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message")
	// custom level
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message", errfmt.LevelError)
``` 

## Wrapped error

有时候你可能希望在错误日志中加入一些上下文信息，类似pkg/errors这个包的能力，可以采用go 1.13以后的版本的标准做法，具体方式如下：

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

你可能比较熟悉`github.com/pkg/errors`提供的能力，为了方便大家使用，这里也提供Is方法，你可以通过生成以后的error的Is方法判断它是不是出自一个特定类型的error。注意：这里不应该直接使用`==`判断，因为很可能底层对这个方法做了Wrap操作，`==`无法区分出`Wrap`后的错误。

```go
func TestErrorIs(t *testing.T) {
	infoErr := Register(http.StatusOK, 20001, "%v is a invalid name")
	err := infoErr.Gen("foo")
	nerr := fmt.Errorf("%w balabalababa info", err)
	assert.Equal(t, true, infoErr.Is(nerr), "")
}
```

## [New] NewHandlerFunc
有些人可能不习惯ginfmt这种req/resp的形式，习惯一个handler有明确的response和error，一方面可以方便单测，另一方面保证逻辑更清晰。

```go
	r.GET("/wrapped_handler", ginfmt.NewHandlerFunc(func(c *gin.Context) (interface{}, error) {
		if time.Now().Unix()%10 == 1 {
			return []int{1, 2, 3}, nil
		}
		return nil, BadRequest.Gen()
	}))
```



## 一个现成的例子

大家可以直接用下面的例子实验。

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

