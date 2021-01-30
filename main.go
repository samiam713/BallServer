package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("Starting Server")

	l, err := net.Listen("tcp4", ":8001")

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Client!")
		go handleConnection(&c)
	}
}
