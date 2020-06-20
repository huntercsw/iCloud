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
	var s1, s2, s3 = "test", "yinuo", "abc"
	fmt.Println(path.Join(s1, s2, s3))
}


