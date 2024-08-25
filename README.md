## Vermouth

Vermonth 是一个基于 Gin 的增强工具，提供了一系列增强工具来帮助 Gin 的开发。

<img src="./img/banner2.jpg" alt="vermonth" width="400" />

### 控制器增强

使用Tag进行书写路由和请求参数。

```go
type TestController struct {
	// 定义该控制器的总路径
    _ interface{} `path:"/api"`

	// 方法
	TestMethod func(a int, b int) interface{} `method:"GET" path:"/test" params:"a,b"`
}

// 例如
// 访问 /api/test 则调用 TestMethod 方法


// 定义控制器
func TestMethod(a int, b int) interface{} {
	return "Hello, Gin! " + strconv.Itoa(a) + strconv.Itoa(b)
}

func NewTestController() *TestController {
	return &TestController{
        TestMethod: TestMethod,
    }
}

// 注册控制器
r := gin.Default()
vermonth.RegisterController(r, NewTestController())

// 访问 /api/test

```

#### 参数注入

- vermonth会自动将请求参数注入到控制器方法中，无需再通过gin获取，只要书写和Tag中相同的参数名即可。
- 参数获取遵循gin的规范，你仍可以使用gin的全部功能。

```go
type TestController struct {
	TestMethod func(a int, b int) interface{} `method:"GET" path:"/test" params:"a,b"`
	TestMethod2 func(req *Request) interface{} `method:"GET" path:"/test" params:"req"`
}

type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

func TestMethod(a int, b int) interface{} {
	return gin.H{
		"message": "Hello, Gin!" + strconv.Itoa(a+b),
	}
}

func TestMethod2(req *Request) interface{} {
	return gin.H{
		"message": "Hello, Gin!" + strconv.Itoa(req.A+req.B),
	}
}
```

#### 公共参数注入
- 当多个控制器需要使用相同的参数时，可以通过公共参数注入来实现。
- 例如获得当前登录的用户
```go
// 公共参数注入
RegisterParamsFunc("/**", func() map[string]interface{} {
	return map[string]interface{}{
		"token": "123",
	}
})


func DoTestParams(token string) interface{} {
	return gin.H{   
		"token": token,
	}
}	

```

#### 参数校验
- 遵循gin的规范，通过`binding:"xxx"`来校验参数
- 对校验进行增强，在定义了`binding:"xxx"`的参数时，可以再使用message=来定义校验失败时的返回信息。

```go
type TestParams struct {
	Name string `json:"name" binding:"required" message:"required=姓名不能为空"` // 如果只有一个校验的话，required=可以不写
	Age int `json:"age" binding:",gt=18" message:"年龄必须大于18"` // 如果只有一个校验的话，required=可以不写
}

type TestController struct {
	TestParams func(params *TestParams) interface{} `method:"GET" path:"/test" params:"params=query" ` //params=query 表示参数从query中获取（包括了表单方式），params=body 表示参数从json中获取，可以不写，默认情况下，POST请求会从body中获取参数，GET请求会从query中获取参数
}
```

#### 自定义校验
- 可以依然沿用gin的校验方式，注册自定义校验器。
- 除此之外，vermonth还提供了自定义校验器，用于一些复杂的校验。

```go
type TestParams struct {
	Name string `json:"name" binding:"required" message:"required=姓名不能为空"` // 如果只有一个校验的话，required=可以不写
	Age int `json:"age" binding:",gt=18" message:"年龄必须大于18"` // 如果只有一个校验的话，required=可以不写
}

// 该结构体内所有以Test开头的方法，都会被认为是自定义校验方法，vermouth会自动调用
func(t *TestParams) TestA() error {
	// 例如某些数据在插入时，需要检查数据库中是否有同名数据
	count := db.QueryRow("SELECT * FROM user WHERE name = ?", t.Name)
	if count > 0 {
		return errors.New("姓名不能重复")
	}
	return nil
}

// 自定义校验方法，可以传入ctx参数，ctx参数中包含了当前请求的所有信息
func(t *TestParams) TestB(ctx *vermouth.Context) error {
	if t.Age < 18 {
		return errors.New("年龄必须大于18")
	}
	return nil
}
```



#### 自定义参数解析
- 待开发


### 切面

vermouth支持AOP，可以通过正则表达式来匹配方法，并执行相应的AOP函数。

```go
// 控制器定义的时候，可以用 _ 为控制器附加名字，如果不附加，则控制器自动使用控制器类型名作为名字
type TestController struct {
    _ interface{} `path:"/api" `
	TestMethod func(a int, b int) interface{} `method:"GET" path:"/test" params:"a,b"`
}

// 注册切面
// 第二个参数为切面优先级，越大的切面会越后面调用
// 例如同时有0和1两个切面，则调用顺序为 0 -> 1
vermonth.RegisterAop("/**", 0, func(aopContext *vermouth.Context) {
	fmt.Println("aop called")

	// 在控制器启动前，你可以随意修改参数
	aopContext.Arguments[0] = 2

	// 调用方法
	aopContext.Call()

	// 修改返回值，例如你可以定义所有接口的通用返回
	aopContext.Result[0] = map[string]interface{}{
		"success": true,
		"data": aopContext.Result[0],
	}
})
```

#### 全局错误处理
- 利用切面，可以轻松完成全局错误的捕获和处理，并返回统一的错误结构。

```go
// 自定义异常处理类
// 定义一个结构体来表示自定义错误
type MyError struct {
	Message string
	Code    int
}

func NewMyError(code int, message string) *MyError {
	return &MyError{
		Message: message,
		Code:    code,
	}
}

// 实现error接口的Error方法
func (e *MyError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// 注册全局错误处理器
vermouth.RegisterAop("/**", 0, func(aopContext *vermouth.Context) {
	defer func() {
		if err := recover(); err != nil {
			// 判断是否是自定义错误
			if myErr, ok := err.(*MyError); ok {
				aopContext.GinContext.JSON(myErr.Code, myErr.Message)
				return
			}
			// 不是我的异常，抛回给中间件处理
			panic(err)
		}
	}()
	aopContext.Call()
})

// 控制器中抛出异常
func DoTestError() interface{} {
	err := NewMyError(400, "test error")
	if true {
		panic(err)
	}
	// 正常的业务逻辑
	return "ok"
}
```

#### 事务
- 利用切面，你可以轻松管理事务。
- 只需要在控制器定义上添加```transaction:"true"``即可。
```go
type TestController struct {
    _ interface{} `path:"/api" `

	// 事务
	TestTransaction func(tx *sql.Tx) interface{} `method:"GET" path:"/test4" params:"tx" transaction:"true"`
}

func DoTestTransaction(tx *sql.Tx) interface{} {
	tx.Exec("INSERT INTO user (name, age) VALUES (?, ?)", "John", 20)
	// do something...
	if true {
		// 当需要回滚事务的时候，只要抛出异常即可
		panic("xxx")
	}
	return nil
}
```


#### 自定义增强
- 你可以通过自定义增强来实现更多的功能，例如日志、缓存、权限控制等。

```go
vermouth.RegisterAop("/**", 0, func(aopContext *vermouth.Context) {
	// 获取控制器中的自定义属性
	logConfig,ok := aopContext.ControllerInformation.Attributes["log"]
	if ok {
		// do something...
		fmt.Println("logConfig:", logConfig)
	}
	aopContext.Call()
})
```

### 协程上下文
- 协程上下文，可以让你在协程中获取当前的上下文信息。
- 例如，你可以通过协程上下文来获取当前的请求信息，或者在协程中传递一些上下文信息。
- 多适用于在controller和service调用中传递上下文信息。

```go
tl := vermouth.NewThreadLocal()

func a(){
	tl.Set("test")
}

func b(){
	s := tl.Get()
	fmt.Println(s) // test
}

func main(){
	a()
	b()
}

```

- 你可以在子协程中，通过tl.Go(func(){})来创建子协程，子协程会继承父协程的上下文。
```go
tl.Go(func(){
	fmt.Println(tl.Get()) // test
})
```	

### 转换器
- 利用converter，可以轻松将一个结构体转换为另一个结构体。
- 使用go generate自动生成转换器代码，避免了调用反射的开销。


