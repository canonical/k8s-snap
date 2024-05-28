package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println(net.SplitHostPort("[fd94::/64]:3080"))
}
