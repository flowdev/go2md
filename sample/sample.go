package sample

import "fmt"

//go:generate ../go2md -package
////go:generate ../go2md -file "$GOFILE"

// Comment for all types.
type (
	// t1 is an int.
	t1 int
	// t2 is a string.
	t2 string
	// t3 is a float64.
	t3 float64
)

// Bla is a simple filter.
//
// flow:
//    in (int)-> [foo1] -> [foo2] -> out
func Bla(i int) int {
	i = foo1(i)
	i = foo2(i)
	return i
}

func foo1(i int) int {
	fmt.Println("i1:", i)
	return i + 1
}

func foo2(i int) int {
	fmt.Println("i2:", i)
	return i + 2
}
