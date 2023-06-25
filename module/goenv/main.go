package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
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
	values, err := url.ParseQuery(os.Getenv("http_query"))
	if err != nil {
		fmt.Println(err)

		return
	}

	if values.Has("a") {
		v, err := strconv.Atoi(values.Get("a"))
		if err != nil {
			fmt.Println(err)

			return
		}

		log_i32(uint32(v))
	} else {
		log_i32(42)
	}

	fmt.Println("goenv environment:")

	for _, e := range os.Environ() {
		fmt.Println("- ", e)
	}

	log_string(fmt.Sprintf("Hello from %v", os.Getenv("remote_addr")))
}
