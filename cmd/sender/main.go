package main

import (
	"net"

	"github.com/atluixx/gftp/internal/transfer"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	if err := transfer.EncodeFile("tests/file.txt", conn); err != nil {
		panic(err)
	}
}
