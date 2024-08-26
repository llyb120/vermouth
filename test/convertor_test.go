package test

import (
	"fmt"
	"github.com/llyb120/vermouth/support"
	"github.com/modern-go/reflect2"
	"testing"

	"github.com/llyb120/vermouth"
	// "github.com/modern-go/reflect2"
)

// given package is github.com/your/awesome-package

// will return the type
// however, if the type has not been used
// it will be eliminated by compiler, so we can not get it in runtime
func TestConvertor(t *testing.T) {
	// var a MyStruct
	// a.Name = "zhangsan"
	// fmt.Println(a)
	// // res := reflect2.TypeByName("test.MyStruct")
	// res := reflect2.TypeOf(&a)
	// fmt.Println(res.Type1().Elem().NumField())
	// res.Set(&a, &MyStruct{Name: "lisi"})
	// fmt.Println(res)

	vermouth.RegisterConvertor(support.MyStruct{}, support.MyStruct2{})

	vermouth.GenerateConvertors("generated", "../generated")
}

func TestConvertorImpl(t *testing.T) {
	var a support.MyStruct
	res := reflect2.TypeOf(a)
	fmt.Println(res)

	//src := generated.MyStruct{Name: "zhangsan"}
	//dest := generated.ConvertSourceToDestination(&src)
	//fmt.Println(dest)
}

func TestDynamicConvert(t *testing.T) {
	var a support.MyStruct
	var b support.MyStruct2
	vermouth.Convert(&a, &b)
	fmt.Println(b)
}
