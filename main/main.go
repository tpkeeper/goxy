package main

import (
	"fmt"
	"goxy/goxy"
)

func main() {
	err := goxy.StartProxy()
	if err != nil {
		fmt.Print(err)
	}
}
