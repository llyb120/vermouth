package test

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/llyb120/vermouth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type TestController struct {
	_           any                                                   `path:"/api" name:"test"`
	TestMethod  func(a int, b int, params map[string]interface{}) any `method:"GET" path:"/test" params:"a,b,params"`
	TestMethod2 func(req *Request) any                                `method:"POST" path:"/test2" params:"req"`

	TestError func() any `method:"GET" path:"/test3"`

	// 事务
	TestTransaction func(tx *sql.Tx) any `method:"GET" path:"/test4" transaction:"true"`
}

func NewTestController() *TestController {
	return &TestController{
		TestMethod:      DoTestMethod,
		TestMethod2:     DoTestMethod2,
		TestError:       DoTestError,
		TestTransaction: DoTestTransaction,
	}
}

func DoTestTransaction(tx *sql.Tx) any {
	tx.Exec("INSERT INTO user (name, age) VALUES (?, ?)", "John", 20)
	panic("must be rollback")
	return nil
}

func DoTestError() any {
	err := vermouth.NewRuntimeError(400, "test error")
	panic(err)
}

func DoTestMethod(a int, b int, params map[string]interface{}) any {
	return gin.H{
		"message": "Hello, Gin!" + strconv.Itoa(a+b),
	}
}

type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

func DoTestMethod2(req *Request) any {
	return gin.H{
		"message": "Hello, Gin!" + strconv.Itoa(req.A+req.B),
	}
}

// 自定义异常处理类
// 定义一个结构体来表示自定义错误
type MyError struct {
	Message string
	Code    int
}

// 实现error接口的Error方法
func (e *MyError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

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
