package test

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/llyb120/vermouth"
)

type TestController struct {
	_           interface{}                                                   `path:"/api" name:"test"`
	TestMethod  func(a int, b int, params map[string]interface{}) interface{} `method:"GET" path:"/test" params:"a,b,params"`
	TestMethod2 func(req *Request) interface{}                                `method:"POST" path:"/test2" params:"req"`

	TestError func() interface{} `method:"GET" path:"/test3"`

	// 事务
	TestTransaction func(tx *sql.Tx, c *gin.Context) interface{} `method:"GET" path:"/test4" transaction:"true"`

	// 公共参数注入
	TestParams func(token string) interface{} `method:"GET" path:"/test5" params:"token"`

	TestParamValidate  func(p *TestParamValidate) interface{} `method:"POST" path:"/test6" params:"p=json"`
	TestParamValidate2 func(p *TestParamValidate) interface{} `method:"POST" path:"/test7" params:"p=form"`
}

func NewTestController() *TestController {
	return &TestController{
		TestMethod:         DoTestMethod,
		TestMethod2:        DoTestMethod2,
		TestError:          DoTestError,
		TestTransaction:    DoTestTransaction,
		TestParams:         DoTestParams,
		TestParamValidate:  DoTestParamValidate,
		TestParamValidate2: DoTestParamValidate,
	}
}

func DoTestTransaction(tx *sql.Tx, c *gin.Context) interface{} {
	tx.Exec("INSERT INTO user (name, age) VALUES (?, ?)", "John", 20)
	panic("must be rollback")
	return nil
}

func DoTestError() interface{} {
	err := vermouth.NewRuntimeError(400, "test error")
	panic(err)
}

func DoTestMethod(a int, b int, params map[string]interface{}) interface{} {
	return gin.H{
		"message": "Hello, Gin!" + strconv.Itoa(a+b),
	}
}

func DoTestParams(token string) interface{} {
	return gin.H{
		"token": token,
	}
}

type TestParamValidate struct {
	A string                  `json:"a" binding:"required" message:"required=a字段必填"`
	B string                  `json:"b" binding:"required" message:"b字段必填"`
	C *TestParamValidateChild `json:"c"`
}

func (p *TestParamValidate) Test() error {
	fmt.Println("test called")
	return errors.New("test error")
}
func (p *TestParamValidate) Test2(ctx *vermouth.Context) error {
	fmt.Println("test2 called")
	return nil
}

type TestParamValidateChild struct {
	A int `json:"a" validate:"required"`
}

func DoTestParamValidate(p *TestParamValidate) interface{} {
	return nil
}

type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

func DoTestMethod2(req *Request) interface{} {
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
