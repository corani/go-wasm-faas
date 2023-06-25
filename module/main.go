package main

import (
	"fmt"
	"os"
	"unsafe"
)

//go:wasmimport env log_i32
func log_i32(v uint32)

//go:wasmimport env log_string
func _log_string(offset, count uint32)

func strToPtr(s string) (uint32, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))

	return uint32(unsafePtr), uint32(len(buf))
}

func log_string(s string) {
	_log_string(strToPtr(s))
}

func main() {
	log_i32(42)

	fmt.Println("goenv environment:")

	for _, e := range os.Environ() {
		fmt.Println("- ", e)
	}

	log_string("Hello from module")
}
