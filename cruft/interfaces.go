package main

import "fmt"

type TypeA struct {
	F1 int
}

type TypeB struct {
	F1 int
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

func main() {
	var items []Common = []Common{&TypeA{F1: 7}, &TypeB{F1: 8}}
	for i, item := range items {
		fmt.Printf("item[%d]: %d\n", i, item.GetF1())
	}
}
