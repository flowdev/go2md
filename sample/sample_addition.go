package sample

import "fmt"

type TBlaer int

// Blaer is a thing that can do bla.
type Blaer int

// NewBlaer creates a new *Blaer with the given increment.
func NewBlaer(inc int) *Blaer {
	b := Blaer(inc)
	return &b
}

// DoBla is the input port of the DoBla operation.
//
// flow:
//     in (TBlaer)-> [bar1] -> [bar2] -> out
func (b *Blaer) DoBla(j TBlaer) TBlaer {
	j = bar1(TBlaer(*b), j)
	j = bar2(TBlaer(*b), j)
	return j
}

func bar1(b, j TBlaer) TBlaer {
	fmt.Println("b:", b, "j1:", j)
	return b + j + 1
}

func bar2(b, j TBlaer) TBlaer {
	fmt.Println("b:", b, "j2:", j)
	return b + j + 2
}
