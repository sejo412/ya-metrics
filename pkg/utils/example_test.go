package utils

import "fmt"

func ExampleHash() {
	data := []byte("hello world")
	key := "super-secret-key"

	out := Hash(data, key)
	fmt.Println(out)

	// Output:
	// YnAst5vO66TT9k/9glPD1E++GYcNg+CZ9kmK8mXIO2E=
}
