package vermouth

import (
	"regexp"
	"slices"
	"strings"
	"github.com/gin-gonic/gin"
)

type aopItem struct {
	Expression *regexp.Regexp
	Fn         func(*AopContext)
	Order      int
	// controllerCaller interface{}
}

type AopContext struct {
	// 调用方法
	Fn func()
	// 参数表
	Arguments     []interface{}
	ArgumentNames []string
	// 返回值
	Result []interface{}
	// 上下文环境
	GinContext *gin.Context

	// 控制器信息
	ControllerInformation *ControllerDefinition
	// 方法信息
	MethodInformation *RequestMapping

}

func (aopContext *AopContext) Call() {
	aopContext.Fn()
}

func newAopContext(argumentsLength int) *AopContext {
	return &AopContext{
		Arguments:     make([]interface{}, argumentsLength),
		ArgumentNames: make([]string, argumentsLength),
	}
}

var aopItems []*aopItem = make([]*aopItem, 0)

func RegisterAop(exp string, order int, fn func(*AopContext)) {
	// 替换.为\.
	exp = strings.Replace(exp, ".", "\\.", -1)
	exp = strings.Replace(exp, "*", "(.+)", -1)
	exp = "^" + exp + "$"
	reg, err := regexp.Compile(exp)
	if err != nil {
		return
	}
	aopItems = append(aopItems, &aopItem{Expression: reg, Fn: fn, Order: order})
	slices.SortFunc(aopItems, func(a, b *aopItem) int {
		return b.Order - a.Order
	})
}

// func main(){
// 	RegisterAop("*.*", func (aopContext *AopContext)  {
// 		aopContext.Arguments[0] = reflect.ValueOf(1)
// 		aopContext.Fn()
// 	})
// }

// type ControllerContext struct {
// 	caller *Caller
// }

// func test(){
// 	RegisterAop("*.*", func (caller *Caller)  {
// 			caller.call()
// 	})
// }
