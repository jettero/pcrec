package main

import "fmt"

type TypeA struct {
	F1 int
}

type TypeB struct {
	F1 int
	F2 int
}

type Common interface {
	GetF1() int
}

func (p *TypeA) GetF1() int {
	return p.F1
}

func (p *TypeB) GetF1() int {
	return p.F1
}

func (p *TypeB) GetF2() int {
	return p.F2
}

func main() {
	var items []Common = []Common{&TypeA{F1: 7}, &TypeB{F1: 8, F2: 9}}
	for i, item := range items {
		switch typed := item.(type) {
		case *TypeB:
			fmt.Printf("item[%d]: %d … %d\n", i, item.GetF1(), typed.GetF2())
		default:
			fmt.Printf("item[%d]: %d\n", i, item.GetF1())
		}
	}
}
