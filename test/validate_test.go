package test

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/llyb120/vermouth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestValidate(t *testing.T) {
	r := gin.Default()

	vermouth.RegisterControllers(r, NewTestController())
	vermouth.RegisterAop("/**", 1, func(ctx *vermouth.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ve, ok := err.(*vermouth.ValidatorError); ok {
					ctx.AutoReturn = false
					errorResponse := map[string]interface{}{
						"success":  false,
						"messages": ve.ErrorMessages,
					}
					ctx.GinContext.JSON(http.StatusBadRequest, errorResponse)
					// ctx.GinContext.JSON(http.StatusBadRequest, ve.ErrorMessages)
					return
				}

				panic(err)
			}
		}()
		ctx.Call()
	})

	// 创建一个HTTP请求
	// 创建一个JSON请求体
	jsonData := `{"b":2}`
	req, _ := http.NewRequest("POST", "/api/test6", bytes.NewBuffer([]byte(jsonData)))

	// 创建一个响应记录器
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusBadRequest, w.Code)

	fmt.Println(w.Body.String())
	// 检查响应体
	//expected := `{"a":1,"b":2}`
	//assert.JSONEq(t, expected, w.Body.String())

	form := url.Values{}
	form.Add("a", "1")
	//form.Add("b", "2")

	req, _ = http.NewRequest("POST", "/api/test7", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 创建一个响应记录器
	w = httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusBadRequest, w.Code)

	fmt.Println(w.Body.String())
	// 检查响应体
	// expected = `{"a":1,"b":2}`
	// assert.JSONEq(t, expected, w.Body.String())
}
