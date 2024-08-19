package vermouth

import (
	"regexp"
	"strings"
)

type paramHandler struct {
	expression *regexp.Regexp
	paramsFunc func() map[string]interface{}
	// params     map[string]interface{}
}

var (
	paramHandlers = []*paramHandler{}
)

// func RegisterParams(exp string, params map[string]interface{}) {
// 	RegisterParamsFunc(exp, func(aopContext *AopContext) map[string]interface{} {
// 		return params
// 	})
// }

func RegisterParamsFunc(exp string, paramsFunc func() map[string]interface{}) {
	// 替换.为\.
	exp = strings.Replace(exp, "**", "[^/]{0,}", -1)
	exp = strings.Replace(exp, "*", "(.+)", -1)
	exp = "^" + exp + "$"
	reg, err := regexp.Compile(exp)
	if err != nil {
		return
	}
	paramHandlers = append(paramHandlers, &paramHandler{
		expression: reg,
		paramsFunc: paramsFunc,
	})
}

