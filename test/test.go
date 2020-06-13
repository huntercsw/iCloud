package main

import "fmt"

type T struct {
	name string
	age int
	love bool
}

func main() {
	t := new(T)
	test(t)
	fmt.Println(t)
}

func test(t *T) {
	if t.name == "" {
		t.name = "yinuo"
	}
}
