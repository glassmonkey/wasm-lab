//go:build wasip1

package main

import "unsafe"

func main() {}

// keepAlive はホストから ptr で参照されるバッファを Go GC から守る。
// wasip1 の Go GC は非移動 (non-moving) なので、ここに入れている間は
// linear memory 上のアドレスが固定される。
var keepAlive = map[uint32][]byte{}

//go:wasmexport alloc
func alloc(size uint32) uint32 {
	if size == 0 {
		return 0
	}
	buf := make([]byte, size)
	ptr := uint32(uintptr(unsafe.Pointer(&buf[0])))
	keepAlive[ptr] = buf
	return ptr
}

//go:wasmexport free
func free(ptr uint32) {
	delete(keepAlive, ptr)
}

// greet は (out_ptr<<32) | out_len を packed i64 で返す。
// Go の //go:wasmexport は戻り値 1 つまでなので、ptr と len を 32bit ずつに詰める。
//
//go:wasmexport greet
func greet(inPtr, inLen uint32) uint64 {
	name := unsafe.String((*byte)(unsafe.Pointer(uintptr(inPtr))), inLen)
	result := []byte("Hello, " + name + "!")
	ptr := uint32(uintptr(unsafe.Pointer(&result[0])))
	keepAlive[ptr] = result
	return uint64(ptr)<<32 | uint64(len(result))
}
