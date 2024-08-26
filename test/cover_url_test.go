package test

import (
	"github.com/gin-gonic/gin"
	"github.com/llyb120/vermouth"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestCoverUrl(t *testing.T) {
	go func() {
		// 创建一个新的Gin引擎
		r := gin.Default()
		r.Use(vermouth.CoverUrlMiddleware("d:/temp/log"))
		vermouth.RegisterControllers(r, NewTestController())
		r.Run(":8080")
	}()

	// 创建一个HTTP请求
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/test?a=1&b=2", nil)
	// 直接发送请求并获取响应
	client := &http.Client{}    // {{ edit_1 }}
	resp, err := client.Do(req) // {{ edit_2 }}
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 检查响应体
	expected := `{"message":"Hello, Gin!3"}`
	assert.JSONEq(t, expected, string(body))

	time.Sleep(5 * time.Second)
}
