package test

import (
	"github.com/gin-gonic/gin"
	"github.com/llyb120/vermouth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestDoTestMethod(t *testing.T) {
	// 创建一个新的Gin引擎
	r := gin.Default()

	vermouth.RegisterControllers(r, NewTestController())

	// 创建一个HTTP请求
	req, _ := http.NewRequest("GET", "/api/test?a=1&b=2", nil)

	// 创建一个响应记录器
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查响应体
	expected := `{"message":"Hello, Gin!3"}`
	assert.JSONEq(t, expected, w.Body.String())

	// do test method2
	req2, _ := http.NewRequest("POST", "/api/test2", strings.NewReader(`{"a":1,"b":2}`))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	expected2 := `{"message":"Hello, Gin!3"}`
	assert.JSONEq(t, expected2, w2.Body.String())
}

func TestGin(t *testing.T) {
	r := gin.Default()
	vermouth.RegisterControllers(r, NewTestController())

	r.Run(":8080")
}
