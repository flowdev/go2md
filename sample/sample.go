package sample

import "fmt"

//go:generate ./go2md

// Comment for all types.
type (
	// Tint1 is an int.
	Tint1 int
	// t2 is a string.
	t2 string
	// t3 is a float64.
	t3 float64
)

// Bla is a simple filter.
//
// flow:
//     in (Tint1)-> [foo1] -> [BlaSome] -> [foo2] -> out
// Some additional bla, bla.
func Bla(i Tint1) Tint1 {
	i = foo1(i)
	i = BlaSome(i)
	return foo2(i)
}

func foo1(i Tint1) Tint1 {
	fmt.Println("i1:", i)
	return i + 1
}

func foo2(i Tint1) Tint1 {
	fmt.Println("i2:", i)
	return i + 2
}

// BlaSome is a simple filter.
//
// flow:
//     in (Tint1)-> [foo3] (TBlaer)-> [DoBla] -> out
// Some additional ...
func BlaSome(i Tint1) Tint1 {
	i = foo3(i)
	doBla := NewBlaer(4)
	return Tint1(doBla.DoBla(TBlaer(i)))
}

func foo3(i Tint1) Tint1 {
	fmt.Println("i3:", i)
	return i + 3
}
