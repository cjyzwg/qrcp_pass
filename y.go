package main

import (
	"fmt"

	"github.com/hashicorp/go-getter"
)

func main() {
	dst := "/Users/cj/Desktop/"
	src := "http://test.hexiefamily.xin/static/static.tar.gz"
	err := getter.GetFile(dst, src)
	if err != nil {
		fmt.Println("err is :", err)
	}
	fmt.Println("get file")
}
