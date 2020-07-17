// scan.go

package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	url := "./upload"
	newurl := strings.Replace(url, "/", "\\", -1)
	fmt.Println(newurl)
	Open(newurl)
	// Open(".\\upload\\")
}

// Open 目录
func Open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
