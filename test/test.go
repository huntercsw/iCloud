package main

import (
	"fmt"
	"path"
)

type T struct {
	name string
	age int
	love bool
}

func main() {
	dir := "/usr/"
	f := "test"
	fmt.Println(path.Join(dir, f))
}


