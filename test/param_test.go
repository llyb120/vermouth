package test

import (
	"github.com/gin-gonic/gin"
	"github.com/llyb120/vermouth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParam(t *testing.T) {
	r := gin.Default()

	vermouth.RegisterControllers(r, NewTestController())
	vermouth.RegisterParamsFunc("/**", func() map[string]interface{} {
		return map[string]interface{}{
			"token": "123",
		}
	})

	// 创建一个HTTP请求
	req, _ := http.NewRequest("GET", "/api/test5", nil)

	// 创建一个响应记录器
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查响应体
	expected := `{"token":"123"}`
	assert.JSONEq(t, expected, w.Body.String())
}
