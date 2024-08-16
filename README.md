## Vermouth

Vermonth 是一个基于 Gin 的增强工具，提供了一系列增强工具来帮助 Gin 的开发。

<img src="./img/banner2.jpg" alt="vermonth" width="400" />

### 控制器增强

使用Tag进行书写路由和请求参数。

```go
type TestController struct {
    _ any `path:"/test"`
	TestMethod func(a int, b int) any `method:"GET" path:"/test" params:"a,b"`
}

// 定义控制器
func TestMethod(a int, b int) any {
	return "Hello, Gin! " + strconv.Itoa(a) + strconv.Itoa(b)
}

func NewTestController() *TestController {
	return &TestController{
        TestMethod: TestMethod,
    }
}

// 注册控制器
r := gin.Default()
vermonth.RegisterController(r, "/api", NewTestController())

// 访问 /api/test

```

#### 参数注入

vermonth会自动将请求参数注入到控制器方法中，无需再通过gin获取，只要书写和Tag中相同的参数名即可。

```go
type TestController struct {
	TestMethod func(a int, b int) any `method:"GET" path:"/test" params:"a,b"`
	TestMethod2 func(req *Request) any `method:"GET" path:"/test" params:"req"`
}

type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

func TestMethod(a int, b int) any {
	return gin.H{
		"message": "Hello, Gin!" + strconv.Itoa(a+b),
	}
}

func TestMethod2(req *Request) any {
	return gin.H{
		"message": "Hello, Gin!" + strconv.Itoa(req.A+req.B),
	}
}
```

### 切面

vermonth支持AOP，可以通过正则表达式来匹配方法，并执行相应的AOP函数。

```go
// 控制器定义的时候，可以用 _ 为控制器附加名字，如果不附加，则控制器自动使用控制器类型名作为名字
type TestController struct {
    _ any `path:"/api" name:"test"`
	TestMethod func(a int, b int) any `method:"GET" path:"/test" params:"a,b"`
}

// 注册切面
// 第二个参数为切面优先级，越大的切面会越后面调用
// 例如同时有0和1两个切面，则调用顺序为 0 -> 1
vermonth.RegisterAop("*.test*", 0, func(aopContext *vermouth.AopContext) {
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