package sample

import "fmt"

// Blaer is a thing that can do bla.
//
// flow:
//     in (int)-> [bar1] -> [bar2] -> out
type Blaer int

// NewBlaer creates a new *Blaer with the given increment.
func NewBlaer(inc int) *Blaer {
	b := Blaer(inc)
	return &b
}

// DoBla is the input port of the DoBla operation.
func (b *Blaer) DoBla(j int) int {
	j = bar1(int(*b), j)
	j = bar2(int(*b), j)
	return j
}

func bar1(b, j int) int {
	fmt.Println("b:", b, "j1:", j)
	return b + j + 1
}

func bar2(b, j int) int {
	fmt.Println("b:", b, "j2:", j)
	return b + j + 2
}
