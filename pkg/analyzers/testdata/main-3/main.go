package main

import (
	"fmt"
	"os"
)

func main() {
	_ = "preved"
	fmt.Println("hello world")
}

func nonmain() {
	fmt.Println("hello world")
}

func anothernonmain() {
	_ = "preved"
	fmt.Println("hello world")
	os.Exit(0)
}
