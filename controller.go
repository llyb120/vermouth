package vermouth

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"reflect"
	"strconv"
	"strings"
)

type requestMapping struct {
	Method      string
	Path        string
	Params      []string
	Transaction bool
}
type controllerDefinition struct {
	Path        string
	Name        string
	Transaction bool
}

func RegisterControllers(r *gin.Engine, controller ...any) {
	// 初始化事务管理器
	// 只执行一次
	initTransactionManager()
	// 注册控制器
	for _, controller := range controller {
		registerController(r, controller)
	}
}

func registerController(r *gin.Engine, controller any) {
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
			api.Params = strings.Split(params, ",")
		} else {
			api.Params = []string{}
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
		r.Handle(api.Method, fullPath, generateApi(controllerDefinition, field.Name, api, controllerValue.Elem().FieldByName(field.Name), controllerInformation))
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
					singleParams := paramHandler.paramsFunc()
					for k, v := range singleParams {
						commonParams[k] = v
					}
				}
			}

			// 拼装参数
			for i := 0; i < numIn; i++ {
				// 字段提取顺序，优先从body中取，如果没有body，则从query中取
				methodParams := methodType.In(i)
				var paramName string
				if len(api.Params) > i {
					paramName = api.Params[i]
				}
				if paramName == "" {
					paramName = "args" + strconv.Itoa(i)
				}
				// 如果有公共参数，则优先使用公共参数
				if value, ok := commonParams[paramName]; ok {
					aopContext.Arguments[i] = value
				} else {
					aopContext.Arguments[i] = extractParamFromContext(c, methodParams, paramName).Interface()
				}
				aopContext.ArgumentNames[i] = paramName
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
			fn = func() {
				aopContext.Fn = oldFn
				aopItem.Fn(aopContext)
			}
			//aopContext.Fn = func(){
			//	aopItem.Fn(aopContext)
			//}
		}

		//aopContext.Fn()
		fn()

		res := aopContext.Result

		if len(res) > 0 {
			c.JSON(200, res[0])
		} else {
			c.JSON(200, nil)
		}

		// // 执行切面
		// fn(c)
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

func extractParamFromContext(c *gin.Context, methodParams reflect.Type, paramName string) reflect.Value {
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
		elemValue := extractParamFromContext(c, methodParams.Elem(), paramName)
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
		if err := c.ShouldBindJSON(newStructPtrRef.Interface()); err == nil {
			return newStructRef
		}
		if err := c.ShouldBindQuery(newStructPtrRef.Interface()); err == nil {
			return newStructRef
		}
		return newStructRef
	case reflect.String:
		return reflect.ValueOf(getStringFromContext(c, paramName))
	case reflect.Int:
		strValue := getStringFromContext(c, paramName)
		if strValue != "" {
			intValue, _ := strconv.Atoi(strValue)
			return reflect.ValueOf(intValue)
		}
		return reflect.ValueOf(0)
	case reflect.Int64:
		strValue := getStringFromContext(c, paramName)
		if strValue != "" {
			intValue, _ := strconv.ParseInt(strValue, 10, 64)
			return reflect.ValueOf(intValue)
		}
		return reflect.ValueOf(int64(0))

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
