package vermouth

import (
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

type TypeCache struct {
	mu    sync.RWMutex
	types map[reflect.Type]*TypeInfo
}

type TypeInfo struct {
	Fields  map[string]*FieldInfo
	Methods map[string]*MethodInfo
}

type FieldInfo struct {
	*reflect.StructField
	Offset uintptr
}

type MethodInfo struct {
	*reflect.Method
	Offset uintptr
}

var cache = &TypeCache{
	types: make(map[reflect.Type]*TypeInfo),
}

func GetTypeInfo(t reflect.Type) *TypeInfo {
	cache.mu.RLock()
	info, exists := cache.types[t]
	cache.mu.RUnlock()
	if exists {
		return info
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Double-check to avoid race condition
	if info, exists = cache.types[t]; exists {
		return info
	}

	info = &TypeInfo{
		Fields:  getFields(t),
		Methods: getMethods(t),
	}
	cache.types[t] = info
	return info
}

func getFields(t reflect.Type) map[string]*FieldInfo {
	fields := make(map[string]*FieldInfo)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields[field.Name] = &FieldInfo{
			StructField: &field,
			Offset:      field.Offset,
		}
	}
	return fields
}

func getMethods(t reflect.Type) map[string]*MethodInfo {
	methods := make(map[string]*MethodInfo)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methods[method.Name] = &MethodInfo{
			Method: &method,
			Offset: uintptr(method.Index),
		}
	}
	return methods
}

func (t *FieldInfo) GetPointer(obj interface{}) unsafe.Pointer {
	return unsafe.Pointer(uintptr(reflect.ValueOf(obj).Pointer()) + t.Offset)
}

func (t *FieldInfo) Set(obj interface{}, value interface{}) error {
	// v := reflect.ValueOf(obj)
	// 使用 unsafe 包直接操作内存
	// fieldPtr := t.unsafe.Pointer(uintptr(reflect.ValueOf(obj).Pointer()) + t.Offset)
	// val := reflect.ValueOf(value)

	SetFieldByPtr(t.GetPointer(obj), value)
	// 类型检查
	// if t.Type != val.Type() {
	// 	return errors.New("提供的值类型与字段类型不匹配")
	// }

	// 根据字段类型进行赋值
	// switch value.(type) {
	// case int:
	// 	*(*int)(fieldPtr) = value.(int) //int(val.Int())
	// case string:
	// 	*(*string)(fieldPtr) = value.(string)
	// case bool:
	// 	*(*bool)(fieldPtr) = value.(bool)
	// // 可以根据需要添加更多类型
	// default:
	// 	// 对于其他类型，使用反射设置值
	// 	reflect.NewAt(t.Type, fieldPtr).Elem().Set(reflect.ValueOf(value))
	// }

	return nil
}

func (t *FieldInfo) Get(obj interface{}) (interface{}, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, errors.New("obj必须是结构体指针")
	}
	v = v.Elem()

	// t := v.Type()
	// info := GetTypeInfo(t)
	// fieldInfo, exists := info.Fields[name]
	// if !exists {
	// 	return nil, errors.New("字段不存在: " + name)
	// }

	// 使用 unsafe 包直接操作内存
	fieldPtr := unsafe.Pointer(uintptr(unsafe.Pointer(v.UnsafeAddr())) + t.Offset)

	// 根据字段类型获取值
	switch t.Type.Kind() {
	case reflect.Int:
		return *(*int)(fieldPtr), nil
	case reflect.String:
		return *(*string)(fieldPtr), nil
	case reflect.Bool:
		return *(*bool)(fieldPtr), nil
	// 可以根据需要添加更多类型
	default:
		// 对于其他类型，使用反射获取值
		return reflect.NewAt(t.Type, fieldPtr).Elem().Interface(), nil
	}
}

func SetFieldByPtr(ptr unsafe.Pointer, value interface{}) {
	switch value.(type) {
	case int:
		*(*int)(ptr) = value.(int) //int(val.Int())
	case string:
		*(*string)(ptr) = value.(string)
	case bool:
		*(*bool)(ptr) = value.(bool)
	}
	// // 可以根据需要添加更多类型
	// default:
	// 	// 对于其他类型，使用反射设置值
	// 	reflect.NewAt(t.Type, fieldPtr).Elem().Set(reflect.ValueOf(value))
	// }	
}

// 优化 SetField 函数，使用缓存和指针操作
func SetField(obj interface{}, name string, value interface{}) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("obj必须是结构体指针")
	}
	v = v.Elem()

	t := v.Type()
	info := GetTypeInfo(t)
	fieldInfo, exists := info.Fields[name]
	if !exists {
		return errors.New("字段不存在: " + name)
	}
	return fieldInfo.Set(obj, value)
}

// 优化 GetField 函数，使用缓存和指针操作
func GetField(obj interface{}, name string) (interface{}, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, errors.New("obj必须是结构体指针")
	}
	v = v.Elem()

	t := v.Type()
	info := GetTypeInfo(t)
	fieldInfo, exists := info.Fields[name]
	if !exists {
		return nil, errors.New("字段不存在: " + name)
	}
	return fieldInfo.Get(obj)
}
