package main

import (
	"fmt"
	"strconv"
)

func main() {
	s := "1.5"
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(f)
	}

}
