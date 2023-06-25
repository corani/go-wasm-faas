package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

func getInt(values url.Values, key string) int {
	if !values.Has(key) {
		return 0
	}

	v, _ := strconv.Atoi(values.Get(key))

	return v
}

func main() {
	values, err := url.ParseQuery(os.Getenv("http_query"))
	if err != nil {
		fmt.Println(err)

		return
	}

	a := getInt(values, "a")
	b := getInt(values, "b")

	fmt.Print(a + b)
}
