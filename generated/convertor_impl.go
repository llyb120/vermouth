package generated

import (
	"github.com/llyb120/vermouth/support"
)

// support.MyStruct => support.MyStruct2
func Convertor0(src *support.MyStruct) *support.MyStruct2 {
	dest := &support.MyStruct2{}
	dest.Name = src.Name
	dest.Tp = support.MyStruct3{}
	dest.Tp.Name = src.Tp.Name
	if src.Tp2 != nil {
		dest.Tp2 = &support.MyStruct3{}
		dest.Tp2.Name = src.Tp2.Name
	}
	return dest
}
