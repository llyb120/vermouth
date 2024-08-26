package test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/llyb120/vermouth"
)

type User struct {
	Name string
	Age  int
}

func TestReflect(t *testing.T) {
	user := &User{
		Name: "test",
		Age:  18,
	}
	vermouth.SetField(user, "Name", "test2")
	fmt.Println(user)

	// 直接设置值的基准测试
	t.Run("Direct Set", func(t *testing.T) {
		user := &User{Name: "test", Age: 18}
		for i := 0; i < 1000000; i++ {
			user.Name = "newName"
		}
	})

	// 通过反射设置值的基准测试
	t.Run("Reflect Set", func(t *testing.T) {
		user := &User{Name: "test", Age: 18}
		for i := 0; i < 1000000; i++ {
			vermouth.SetField(user, "Name", "newName")
		}
	})

	t.Run("Cache Test", func(t *testing.T) {
		t1 := reflect.TypeOf(user)
		t2 := reflect.TypeOf(user)
		t.Errorf("t1 == t2: %v", t1 == t2)
	})

	// 比较结果
	t.Run("Benchmark Comparison", func(t *testing.T) {
		directResult := testing.Benchmark(func(b *testing.B) {
			user := &User{Name: "test", Age: 18}
			for i := 0; i < b.N; i++ {
				user.Name = "newName"
			}
		})

		reflectResult := testing.Benchmark(func(b *testing.B) {
			user := &User{Name: "test", Age: 18}
			info := vermouth.GetTypeInfo(reflect.TypeOf(*user))
			fieldInfo, _ := info.Fields["Name"]
			fieldPtr := fieldInfo.GetPointer(user)

			// 使用 unsafe 包直接操作内存
			// fieldPtr := unsafe.Pointer(uintptr(unsafe.Pointer(user)) + fieldInfo.Offset)
			for i := 0; i < b.N; i++ {
				vermouth.SetFieldByPtr(fieldPtr, "newName")
				// fieldInfo.Set(user, "newName")
				// *(*string)(fieldPtr) = "newName"
			}
		})

		reflectResult2 := testing.Benchmark(func(b *testing.B) {
			user := &User{Name: "test", Age: 18}
			v := reflect.ValueOf(user).Elem()
			for i := 0; i < b.N; i++ {
				v.FieldByName("Name").SetString("newName")
			}
		})

		fmt.Printf("直接设置: %.3f ns/op\n", float64(directResult.T.Nanoseconds())/float64(directResult.N))
		fmt.Printf("vermouth/reflect设置: %.3f ns/op\n", float64(reflectResult.T.Nanoseconds())/float64(reflectResult.N))
		fmt.Printf("普通反射设置: %.3f ns/op\n", float64(reflectResult2.T.Nanoseconds())/float64(reflectResult2.N))
		倍数 := float64(reflectResult.N) / float64(reflectResult2.N)
		fmt.Printf("比普通反射快了 %.2f 倍\n", 倍数)
		fmt.Printf("性能比例: %.2f%%\n", float64(reflectResult.N)/float64(directResult.N)*100)
	})
}
