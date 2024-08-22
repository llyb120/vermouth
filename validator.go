package vermouth

import (
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10" // 引入验证库
	"reflect"
	"strings"
	"sync"
)

var translator ut.Translator

var messageCache = &cacheMap{}

type cacheMap struct {
	m  sync.Map
	mu sync.Mutex
}

// 设置缓存，如果不存在则使用提供的函数生成值并添加
func (c *cacheMap) SetIfAbsent(key string, valueFunc func() interface{}) (actual interface{}, loaded bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 先检查是否存在
	if actual, loaded = c.m.Load(key); !loaded {
		// 如果不存在，调用生成函数并存储值
		actual = valueFunc()
		c.m.Store(key, actual)
	}
	return
}

// 获取缓存
func (c *cacheMap) Get(key string) (interface{}, bool) {
	return c.m.Load(key)
}

// 删除缓存
func (c *cacheMap) Delete(key string) {
	c.m.Delete(key)
}

func initValidator() {
	// zhLocale := zh.New()
	// uni := ut.New(zhLocale, zhLocale)
	// translator, _ = uni.GetTranslator("zh")

	// if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
	// 	// 注册自定义校验规则
	// 	v.RegisterValidation("customToken", func(fl validator.FieldLevel) bool {
	// 		return fl.Field().String() == "customToken"
	// 	})
	// 	// 注册翻译器
	// 	v.RegisterTranslation("required", translator, func(ut ut.Translator) error {
	// 		return ut.Add("required", "{0} 必填", true) // 使用标签中的字段名称
	// 	}, func(ut ut.Translator, fe validator.FieldError) string {
	// 		t, _ := ut.T("required", fe.Field())
	// 		return t
	// 	})
	// 	v.RegisterTranslation("customToken", translator, func(ut ut.Translator) error {
	// 		return ut.Add("customToken", "{0} 校验失败", true) // 自定义错误信息
	// 	}, func(ut ut.Translator, fe validator.FieldError) string {
	// 		t, _ := ut.T("customToken", fe.Field())
	// 		return t
	// 	})
	// }
}

type ValidatorError struct {
	ErrorMessages []string `json:messages`
}

func (e *ValidatorError) Error() string {
	return strings.Join(e.ErrorMessages, "; ")
}

func makeValidatorError(u interface{}, ers validator.ValidationErrors) *ValidatorError {
	ve := &ValidatorError{}
	for _, validationErr := range ers {
		ns := validationErr.Namespace() // 获取字段的命名空间
		// ns去掉第一个.之前的内容
		ns = ns[strings.Index(ns, ".")+1:]
		_, errorInfo := getFieldAndTag(reflect.TypeOf(u), ns, "message")
		if errorInfo != "" {
			mp, _ := messageCache.SetIfAbsent(errorInfo, func() interface{} {
				// 解析
				mp := parseMessage(validationErr.Tag(), errorInfo)
				return mp
			})
			if mp != nil {
				if cache, ok := mp.(map[string]string); ok {
					if errorInfo, ok = cache[validationErr.Tag()]; ok {
						ve.ErrorMessages = append(ve.ErrorMessages, errorInfo)
					}
				}
			}
			// ve.ErrorMessages = append(ve.ErrorMessages, errorInfo)
		}
		// if field != nil {
		// 	return ns + ":" + errorInfo // 返回错误
		// } else {
		// 	return "缺失reg_error_info"
		// }
	}
	return ve
}

// 获取字段和标签的递归函数
func getFieldAndTag(t reflect.Type, ns, tagName string) (reflect.StructField, string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return reflect.StructField{}, ""
	}

	parts := strings.Split(ns, ".")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == parts[0] {
			if len(parts) == 1 {
				return field, field.Tag.Get(tagName)
			}
			if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
				return getFieldAndTag(field.Type, strings.Join(parts[1:], "."), tagName)
			}
		}
	}
	return reflect.StructField{}, ""
}

// 解析 message 字符串
func parseMessage(tag, message string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(message, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) != "" {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else if len(kv) == 1 { // 处理只有一句话的情况
			result[tag] = strings.TrimSpace(part) // 将这句话作为tag的错误信息
		}
	}
	return result
}
