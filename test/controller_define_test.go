package test

import(
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/llyb120/vermouth"
)

type TestController struct {
	_           any                                                   `path:"/api" name:"test"`
	TestMethod  func(a int, b int, params map[string]interface{}) any `method:"GET" path:"/test" params:"a,b,params"`
	TestMethod2 func(req *Request) any                                `method:"POST" path:"/test2" params:"req"`

	TestError func() any `method:"GET" path:"/test3"`

	// 事务
	TestTransaction func(tx *sql.Tx, c *gin.Context) any `method:"GET" path:"/test4" transaction:"true"`

	// 公共参数注入
	TestParams func(token string) any `method:"GET" path:"/test5" params:"token"`
}

func NewTestController() *TestController {
	return &TestController{
		TestMethod:      DoTestMethod,
		TestMethod2:     DoTestMethod2,
		TestError:       DoTestError,
		TestTransaction: DoTestTransaction,
		TestParams:      DoTestParams,
	}
}

func DoTestTransaction(tx *sql.Tx, c *gin.Context) any {
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

func DoTestParams(token string) any {
	return gin.H{
		"token": token,
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