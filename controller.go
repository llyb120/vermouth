package vermouth

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strconv"
	"strings"
)

type requestMapping struct {
	Method      string
	Path        string
	Params      []*paramItem
	Transaction bool
}
type paramItem struct {
	ParamName string
	From      string
}
type controllerDefinition struct {
	Path        string
	Name        string
	Transaction bool
}

func RegisterControllers(r interface{}, controller ...interface{}) {
	// 初始化事务管理器
	// 只执行一次
	initTransactionManager()
	initValidator()
	// 注册控制器
	for _, controller := range controller {
		registerController(r, controller)
	}
}

func registerController(r interface{}, controller interface{}) {
	controllerDefinition := &controllerDefinition{}
	controllerType := reflect.TypeOf(controller)
	controllerValue := reflect.ValueOf(controller)
	// controllerMethods := controllerType.NumMethod()
	// 解析控制器字段
	controllerElemType := controllerType.Elem()
	controllerFields := controllerElemType.NumField()
	apiMap := make(map[string]*requestMapping)
	globalField, ok := controllerElemType.FieldByName("_")
	if ok {
		globalFieldTag := globalField.Tag.Get("path")
		if globalFieldTag != "" {
			controllerDefinition.Path = globalFieldTag
		}
		name := globalField.Tag.Get("name")
		if name != "" {
			controllerDefinition.Name = name
		}
	}
	// 如果没有_，默认生成一个
	// 如果没有起名，则默认用类名
	if controllerDefinition.Name == "" {
		controllerDefinition.Name = getStructName(controller)
	}
	for i := 0; i < controllerFields; i++ {
		field := controllerElemType.Field(i)
		if field.Name == "_" {
			continue
		}
		// 该字段必须是一个方法
		if field.Type.Kind() != reflect.Func {
			continue
		}
		// 获取字段上的标签
		tag := field.Tag.Get("method")
		if tag == "" {
			continue
		}
		path := field.Tag.Get("path")
		if path == "" {
			continue
		}
		api := &requestMapping{Method: tag, Path: path}
		params := field.Tag.Get("params")
		if params != "" {
			api.Params = []*paramItem{}
			for _, param := range strings.Split(params, ",") {
				// trim
				param = strings.TrimSpace(param)
				paramParts := strings.SplitN(param, "=", 2)
				if len(paramParts) == 2 {
					api.Params = append(api.Params, &paramItem{ParamName: paramParts[0], From: paramParts[1]})
				} else {
					// 默认GET请求从query中获取参数，POST请求从body中获取参数
					if api.Method == "GET" {
						api.Params = append(api.Params, &paramItem{ParamName: param, From: "query"})
					} else {
						api.Params = append(api.Params, &paramItem{ParamName: param, From: "json"})
					}
				}
			}
		} else {
			api.Params = []*paramItem{}
		}
		// 事务
		transaction := field.Tag.Get("transaction")
		apiMap[field.Name] = api

		// 计算完整的路径
		fullPath := controllerDefinition.Path
		if !strings.HasPrefix(api.Path, "/") {
			fullPath += "/"
		}
		fullPath += api.Path
		// 去除可能出现的双斜杠
		fullPath = strings.Replace(fullPath, "//", "/", -1)
		api.Path = fullPath

		controllerInformation := NewControllerInformation()
		controllerInformation.Path = fullPath
		controllerInformation.Transaction = transaction == "true"
		coverUrl := field.Tag.Get("cover_url")
		if coverUrl != "" {
			urlCoverCache.Store(coverUrl, fullPath)
		}

		// tag整理成map
		tagMap := make(map[string]string)
		for _, tag := range strings.Split(string(field.Tag), " ") {
			parts := strings.SplitN(tag, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], `"`)
				value := strings.Trim(parts[1], `"`)
				tagMap[key] = value
			}
		}
		controllerInformation.Attributes = tagMap
		//r可能为*gin.Engine或*gin.RouterGroup
		switch v := r.(type) {
		case *gin.Engine:
			v.Handle(api.Method, fullPath, generateApi(controllerDefinition, field.Name, api, controllerValue.Elem().FieldByName(field.Name), controllerInformation))
		case *gin.RouterGroup:
			v.Handle(api.Method, fullPath, generateApi(controllerDefinition, field.Name, api, controllerValue.Elem().FieldByName(field.Name), controllerInformation))
		default:
			panic("unsupported router type")
		}
	}

}

func generateApi(controllerDefinition *controllerDefinition, methodName string, api *requestMapping, method reflect.Value, controllerInformation *ControllerInformation) gin.HandlerFunc {
	// methodFullName := fmt.Sprintf("%s.%s", controllerDefinition.Name, methodName)
	// fmt.Println("current method name is ", methodFullName)
	return func(c *gin.Context) {
		// 根据params继续拼接参数
		// 获取method参数表
		methodType := method.Type()
		numIn := methodType.NumIn()
		// args := make([]reflect.Value, numIn)
		aopContext := newAopContext(numIn)
		// aopContext.ControllerInformation = controllerDefinition
		// aopContext.MethodInformation = api
		aopContext.GinContext = c
		aopContext.ControllerInformation = controllerInformation
		aopContext.Fn = func() {

			// 公共参数注入
			commonParams := make(map[string]interface{})
			for _, paramHandler := range paramHandlers {
				if paramHandler.expression.MatchString(controllerInformation.Path) {
					singleParams := paramHandler.paramsFunc(aopContext)
					for k, v := range singleParams {
						commonParams[k] = v
					}
				}
			}

			// 拼装参数
			for i := 0; i < numIn; i++ {
				// 字段提取顺序，优先从body中取，如果没有body，则从query中取
				methodParams := methodType.In(i)
				var pi *paramItem
				if len(api.Params) > i {
					pi = api.Params[i]
				}
				if pi == nil {
					if api.Method == "GET" {
						pi = &paramItem{ParamName: "args" + strconv.Itoa(i), From: "query"}
					} else {
						pi = &paramItem{ParamName: "args" + strconv.Itoa(i), From: "json"}
					}
				}
				// 如果有公共参数，则优先使用公共参数
				if value, ok := commonParams[pi.ParamName]; ok {
					aopContext.Arguments[i] = value
				} else {
					aopContext.Arguments[i] = extractParamFromContext(aopContext, methodParams, pi).Interface()
				}
				aopContext.ArgumentNames[i] = pi.ParamName
			}

			// 转换成reflect.Value
			reflectArguments := make([]reflect.Value, len(aopContext.Arguments))
			for i, arg := range aopContext.Arguments {
				reflectArguments[i] = reflect.ValueOf(arg)
			}
			res := method.Call(reflectArguments)
			aopContext.Result = make([]interface{}, len(res))
			for i, v := range res {
				aopContext.Result[i] = v.Interface()
			}
		}
		fn := aopContext.Fn

		// 检查是否有切面
		for _, aopItem := range aopItems {
			if !aopItem.Expression.MatchString(controllerInformation.Path) {
				continue
			}
			oldFn := fn
			aopItemCopy := aopItem // 创建aopItem的副本
			fn = func() {
				aopContext.Fn = oldFn
				aopItemCopy.Fn(aopContext) // 使用副本
			}
			//aopContext.Fn = func(){
			//	aopItem.Fn(aopContext)
			//}
		}

		//aopContext.Fn()
		fn()

		if aopContext.AutoReturn {
			res := aopContext.Result
			if len(res) > 0 {
				c.JSON(200, res[0])
			} else {
				c.JSON(200, nil)
			}
		}
	}
}

func getStringFromContext(c *gin.Context, key string) string {
	if strValue, ok := c.GetPostForm(key); ok {
		return strValue
	} else if strValue, ok := c.GetQuery(key); ok {
		return strValue
	}
	return ""
}

func extractParamFromContext(ctx *Context, methodParams reflect.Type, pi *paramItem) reflect.Value {
	c := ctx.GinContext
	switch methodParams.Kind() {
	case reflect.Ptr:
		// 有一些特殊值需要处理
		if methodParams.AssignableTo(reflect.TypeOf((*sql.Tx)(nil))) {
			tx, ok := c.Get("Vermouth:tx")
			if ok {
				return reflect.ValueOf(tx)
			}
		} else if methodParams.AssignableTo(reflect.TypeOf((*gin.Context)(nil))) {
			return reflect.ValueOf(c)
		}
		elemValue := extractParamFromContext(ctx, methodParams.Elem(), pi)
		ptrValue := reflect.New(methodParams.Elem())
		ptrValue.Elem().Set(elemValue)
		return ptrValue
		// return extractParamFromContext(c, methodParams.Elem(), paramName)
	case reflect.Map:
		newMapValue := reflect.MakeMap(methodParams)
		newMap := newMapValue.Interface()
		if err := c.ShouldBindJSON(&newMap); err == nil {
			return newMapValue
		}
		queryMap := make(map[string]string)
		if err := c.ShouldBindQuery(&queryMap); err == nil {
			for k, v := range queryMap {
				newMapValue.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
			}
		}
		return newMapValue
	case reflect.Struct:
		newStructPtrRef := reflect.New(methodParams)
		newStructRef := newStructPtrRef.Elem()
		var ve *ValidatorError
		if pi.From == "json" {
			if err := c.ShouldBindJSON(newStructPtrRef.Interface()); err != nil {
				// 如果绑定失败，则返回一个空的结构体
				if ers, ok := err.(validator.ValidationErrors); ok {
					ve = makeValidatorError(newStructPtrRef.Interface(), ers)
				}
			}
		} else if pi.From == "query" {
			if err := c.ShouldBindQuery(newStructPtrRef.Interface()); err != nil {
				if ers, ok := err.(validator.ValidationErrors); ok {
					ve = makeValidatorError(newStructPtrRef.Interface(), ers)
				}
			}
		} else if pi.From == "form" {
			if err := c.ShouldBind(newStructPtrRef.Interface()); err != nil {
				if ers, ok := err.(validator.ValidationErrors); ok {
					ve = makeValidatorError(newStructPtrRef.Interface(), ers)
				}
			}
		}

		// if err := c.ShouldBindQuery(newStructPtrRef.Interface()); err == nil {
		// 	return newStructRef
		// }
		// 如果可以调用Test系列方法
		// methodParams.NumMethod().for
		tp := newStructPtrRef.Type()
		for i := 0; i < tp.NumMethod(); i++ {
			method := tp.Method(i)
			if strings.HasPrefix(method.Name, "Test") {
				// 只允许有一个参数
				if method.Type.NumIn() == 2 {
					// 第一个参数必定为*Context
					if method.Type.In(1).AssignableTo(reflect.TypeOf((*Context)(nil))) {
						vals := method.Func.Call([]reflect.Value{newStructPtrRef, reflect.ValueOf(ctx)})
						if len(vals) > 0 {
							if err, ok := vals[0].Interface().(error); ok {
								if ve == nil {
									ve = &ValidatorError{}
								}
								ve.ErrorMessages = append(ve.ErrorMessages, err.Error())
							}
						}
					}
				} else if method.Type.NumIn() == 1 {
					vals := method.Func.Call([]reflect.Value{newStructPtrRef})
					if len(vals) > 0 {
						if err, ok := vals[0].Interface().(error); ok {
							if ve == nil {	
								ve = &ValidatorError{}
							}
							ve.ErrorMessages = append(ve.ErrorMessages, err.Error())
						}
					}
				}
			}
		}

		if ve != nil && len(ve.ErrorMessages) > 0 {
			panic(ve)
		}
		return newStructRef
	case reflect.String:
		return reflect.ValueOf(getStringFromContext(c, pi.ParamName))
	case reflect.Int:
		strValue := getStringFromContext(c, pi.ParamName)
		if strValue != "" {
			intValue, _ := strconv.Atoi(strValue)
			return reflect.ValueOf(intValue)
		}
		return reflect.ValueOf(0)
	case reflect.Int64:
		strValue := getStringFromContext(c, pi.ParamName)
		if strValue != "" {
			intValue, _ := strconv.ParseInt(strValue, 10, 64)
			return reflect.ValueOf(intValue)
		}
		return reflect.ValueOf(int64(0))
	case reflect.Slice:
		strValue := getStringFromContext(c, pi.ParamName)
		if strValue != "" {
			// 暂时先这样，后面再改
			// 使用逗号拆分字符串
			splitValues := strings.Split(strValue, ",")
			// 创建一个与methodParams类型相同的slice
			sliceValue := reflect.MakeSlice(methodParams, len(splitValues), len(splitValues))
			for i, v := range splitValues {
				sliceValue.Index(i).Set(reflect.ValueOf(v).Convert(methodParams.Elem()))
			}
			return sliceValue
		}
		return reflect.ValueOf([]string{})
	default:
		return reflect.ValueOf(nil)
	}
}

func getStructName(i interface{}) string {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return t.Name()
	}
	return ""
}
