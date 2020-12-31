package ginfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sleagon/ginfmt/errfmt"
	"github.com/stretchr/testify/assert"
)

type Resp1 struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func TestData(t *testing.T) {
	r := gin.Default()
	r.Use(MW())
	r.GET("/ginfmt", func(c *gin.Context) {
		Data(c, "foo")
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ginfmt", nil)
	r.ServeHTTP(w, req)
	resp := new(Resp1)
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "ok", resp.Message)
	assert.Equal(t, "foo", resp.Data)
}

type Resp2 struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    int    `json:"data"`
}

func TestError(t *testing.T) {
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message")
	r := gin.Default()
	r.Use(MW())
	r.GET("/ginfmt", func(c *gin.Context) {
		Error(c, FooError.Gen())
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

func TestWrappedError(t *testing.T) {
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message")
	r := gin.Default()
	r.Use(MW())
	r.GET("/ginfmt", func(c *gin.Context) {
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

type Resp3 struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func TestDataError(t *testing.T) {
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo message")
	r := gin.Default()
	r.Use(MW())
	r.GET("/ginfmt", func(c *gin.Context) {
		DataError(c, "foo", FooError.Gen())
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ginfmt", nil)
	r.ServeHTTP(w, req)
	resp := new(Resp3)
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, resp.Code, FooError.Gen().Code())
	assert.Equal(t, resp.Message, FooError.Gen().Message(context.TODO(), ""))
	assert.Equal(t, "foo", resp.Data)
}

func DemoTrans(ctx context.Context, locale string, key string) string {
	demoMap := map[string]map[string]string{
		"zh": {
			"foo": "这是一个foo信息",
		},
		"en-US": {
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

type Resp4 struct {
	Code    int
	Message string
	Data    Data4
}

type Data4 struct {
	Name string
	Age  int
}

func TestHandler(t *testing.T) {

	Init(nil, DemoTrans)
	FooError := errfmt.Register(http.StatusNotFound, 10010, "foo")

	myhandler := func(c *gin.Context) (interface{}, error) {
		return Data4{"foo", 12}, FooError.Gen()
	}

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Request.Header.Set("locale", "zh")
	})
	r.Use(MW())
	r.GET("/ginfmt", NewHandlerFunc(myhandler))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ginfmt", nil)
	r.ServeHTTP(w, req)
	resp := new(Resp4)
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), resp))
	assert.Equal(t, resp.Code, FooError.Gen().Code())
	assert.Equal(t, resp.Message, FooError.Gen().Message(context.TODO(), "zh"))
	assert.Equal(t, "foo", resp.Data.Name)
	assert.Equal(t, 12, resp.Data.Age)
}
