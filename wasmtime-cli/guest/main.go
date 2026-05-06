//go:build wasip1

package main

func main() {}

//go:wasmexport add
func add(a, b int32) int32 {
	return a + b
}

//go:wasmexport mul
func mul(a, b int32) int32 {
	return a * b
}
