package support

type MyStruct struct {
	Name string
	Age  int
	Tp   MyStruct4
	Tp2  *MyStruct4
}

type MyStruct2 struct {
	Name string
	Tp   MyStruct3
	Tp2  *MyStruct3
}

type MyStruct3 struct {
	Name string
}

type MyStruct4 struct {
	Name string
}
