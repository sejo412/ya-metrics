package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello world")
	os.Exit(0) // want "os.Exit from main function main package"
}
