package vermouth

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

var convertors = map[reflect.Type][]reflect.Type{}

type Type struct {
	Name   string
	Fields []*reflect.Type
}

type convertorDefinition struct {
	PackageName     string
	SourceType      *reflect.Type
	DestinationType *reflect.Type
	Imports         map[string]struct{}
}

func Convert(src interface{}, dst interface{}) {
	// 提取对应的转换器
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)
	srcInfo := GetTypeInfo(srcType)
	dstInfo := GetTypeInfo(dstType)

	for _, fieldInfo := range srcInfo.Fields {
		dstFieldInfo, ok := dstInfo.Fields[fieldInfo.Name]
		if !ok {
			continue
		}
		// 先取目标的值
		dstFieldValue, err := dstFieldInfo.Get(dst)
		if err != nil {
			continue
		}
		srcFieldValue, err := fieldInfo.Get(src)
		if err != nil {
			continue
		}

		if srcFieldValue == nil {
			dstFieldValue = nil
			dstFieldInfo.Set(dst, dstFieldValue)
			continue
		}

		if fieldInfo.realType == reflect.Struct && dstFieldInfo.realType == reflect.Map {
			dstFieldValue = reflect.New(dstFieldInfo.Type).Interface()
			structToMap(srcFieldValue, dstFieldValue, fieldInfo, dstFieldInfo)
		}

		// 如果是结构体，则递归转换
		if fieldInfo.realType == reflect.Struct && dstFieldInfo.realType == reflect.Struct {
			// 如果src是nil，则目标也设置为nil
			if srcFieldValue == nil {
				dstFieldValue = nil
				dstFieldInfo.Set(dst, dstFieldValue)
				continue
			} else {
				dstFieldValue = reflect.New(dstFieldInfo.Type).Interface()
				dstFieldInfo.Set(dst, dstFieldValue)
				Convert(srcFieldValue, dstFieldValue)
			}
		} else {
			fieldValue, err := fieldInfo.Get(src)
			if err != nil {
				continue
			}
			fieldInfo.Set(dst, fieldValue)
		}
	}

}

// 结构体 -> Map
func structToMap(src interface{}, dst interface{}, srcFieldInfo *FieldInfo, dstFieldInfo *FieldInfo) {
	// srcFieldinfo要不是一个结构体，要不也是一个map，否则无法转换
	// srcType := GetTypeInfo(srcFieldInfo.Type)
	// for _, fieldInfo := range srcType.Fields {
	// 	dstFieldValue, err := dstFieldInfo.Get(dst)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	srcFieldValue, err := fieldInfo.Get(src)
	// 	if err != nil {
	// 		continue
	// 	}
	// }
	// dstType := GetTypeInfo(dstFieldInfo.Type).Type
	// srcValue := reflect.TypeOf(src)
	// dstValue := reflect.TypeOf(dst)

}

func RegisterConvertor(src interface{}, dst interface{}) {
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	// var convertor *convertorDefinition = &convertorDefinition{
	// 	SourceType: &srcType,
	// 	DestinationType: &dstType,
	// 	Imports: map[string]struct{}{
	// 		srcType.PkgPath(): {},
	// 		dstType.PkgPath(): {},
	// 	},
	// }
	convertors[srcType] = append(convertors[srcType], dstType)

	// imports去个重

}

func GenerateConvertors(packageName, path string) error {
	// 确保path以斜杠结尾
	if path[len(path)-1] != '/' {
		path += "/"
	}
	// 没有目录则创建
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	os.MkdirAll(path, os.ModePerm)
	f, err := os.Create(fmt.Sprintf("%sconvertor_impl.go", path))
	if err != nil {
		return err
	}
	defer f.Close()

	var builder strings.Builder
	var imports = make(map[string]struct{})

	var i = 0
	for srcType, dstTypes := range convertors {
		// 分析类型
		for _, dstType := range dstTypes {
			// imports[srcType.PkgPath()] = struct{}{}
			// dstType.PackageName = packageName
			renderCode(&builder, srcType, dstType, i, imports)
			i++
			// fileName := filepath.Join(path, fmt.Sprintf("%s_to_%s.go", srcType.Name(), dstType.Name()))
			// os.Create(fileName)
		}
	}

	var headBuilder strings.Builder
	headBuilder.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	headBuilder.WriteString("import (\n")
	for importPath := range imports {
		headBuilder.WriteString(fmt.Sprintf("\t\"%s\"\n", importPath))
	}
	headBuilder.WriteString(")\n\n")
	f.WriteString(headBuilder.String())
	f.WriteString(builder.String())

	return nil
}

func renderCode(builder *strings.Builder, srcType, dstType reflect.Type, index int, imports map[string]struct{}) error {
	// var builder strings.Builder
	// builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	// builder.WriteString("import (\n")
	// for importPath := range convertor.Imports {
	// 	builder.WriteString(fmt.Sprintf("\t\"%s\"\n", importPath))
	// }
	// builder.WriteString(")\n\n")
	builder.WriteString(fmt.Sprintf("// %v => %v\n", srcType, dstType))
	builder.WriteString(fmt.Sprintf("func Convertor%d(src *%v) *%v {\n", index, srcType, dstType))
	// 实现方法体
	builder.WriteString(fmt.Sprintf("	dest := &%v{}\n", dstType))

	generateFieldAssignments(builder, srcType, dstType, "src", "dest", imports)
	// for i := 0; i < srcType.NumField(); i++ {
	// 	srcField := srcType.Field(i)
	// 	if dstField, ok := dstType.FieldByName(srcField.Name); ok {
	// 		if dstField.Type == srcField.Type {
	// 			builder.WriteString(fmt.Sprintf("	dest.%s = src.%s\n", dstField.Name, srcField.Name))
	// 		} else {
	// 			// 处理类型不一致的情况
	// 			if dstField.Type.Kind() == reflect.String && srcField.Type.Kind() == reflect.Int {
	// 				builder.WriteString(fmt.Sprintf("	dest.%s = fmt.Sprintf(\"%%d\", src.%s)\n", dstField.Name, srcField.Name))
	// 			}
	// 			// 其他类型转换可以在这里添加
	// 		}
	// 	}
	// }

	// builder.WriteString(fmt.Sprintf("	dest.Name = src.Name\n"))
	// builder.WriteString(fmt.Sprintf("	dest.Age = src.Age\n"))
	// builder.WriteString(fmt.Sprintf("	dest.Email = src.Email\n"))
	builder.WriteString("	return dest\n")
	builder.WriteString("}\n")

	// t := template.Must(template.New("convert").Parse(tmpl))
	// err = t.Execute(f, convertor)
	// if err != nil {
	// 	return err
	// }
	//.WriteString(builder.String())

	return nil
}

func generateFieldAssignments(builder *strings.Builder, srcType, dstType reflect.Type, srcVar, dstVar string, imports map[string]struct{}) {
	imports[srcType.PkgPath()] = struct{}{}
	imports[dstType.PkgPath()] = struct{}{}
	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		// 只处理大写开头的字段
		if !isExported(srcField.Name) {
			continue
		}
		if dstField, ok := dstType.FieldByName(srcField.Name); ok {
			if dstField.Type == srcField.Type {
				builder.WriteString(fmt.Sprintf("	%s.%s = %s.%s\n", dstVar, dstField.Name, srcVar, srcField.Name))
			} else if dstField.Type.Kind() == reflect.Ptr && srcField.Type.Kind() == reflect.Ptr {
				// 处理指针字段
				builder.WriteString(fmt.Sprintf("	if %s.%s != nil {\n", srcVar, srcField.Name))
				builder.WriteString(fmt.Sprintf("		%s.%s = &%v{}\n", dstVar, dstField.Name, dstField.Type.Elem()))
				generateFieldAssignments(builder, srcField.Type.Elem(), dstField.Type.Elem(), fmt.Sprintf("%s.%s", srcVar, srcField.Name), fmt.Sprintf("%s.%s", dstVar, dstField.Name), imports)
				builder.WriteString("	}\n")
			} else if dstField.Type.Kind() == reflect.Struct && srcField.Type.Kind() == reflect.Struct {
				// 递归处理子结构
				builder.WriteString(fmt.Sprintf("	%s.%s = %v{}\n", dstVar, dstField.Name, dstField.Type))
				generateFieldAssignments(builder, srcField.Type, dstField.Type, fmt.Sprintf("%s.%s", srcVar, srcField.Name), fmt.Sprintf("%s.%s", dstVar, dstField.Name), imports)
			} else {
				// 处理类型不一致的情况
				if dstField.Type.Kind() == reflect.String && srcField.Type.Kind() == reflect.Int {
					builder.WriteString(fmt.Sprintf("	%s.%s = fmt.Sprintf(\"%%d\", %s.%s)\n", dstVar, dstField.Name, srcVar, srcField.Name))
				}
				// 其他类型转换可以在这里添加
			}
		}
	}
}

func isExported(name string) bool {
	return strings.ToUpper(name[:1]) == name[:1]
}
