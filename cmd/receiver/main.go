package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/atluixx/gftp/internal/protocol"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	header := make([]byte, 16)

	file, err := os.Create("tests/output.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for {
		_, err := io.ReadFull(conn, header)

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		h, err := protocol.DecodeHeader(header)
		if err != nil {
			panic(err)
		}
		fmt.Println("expected payload:", h.PayloadSize)

		payload := make([]byte, h.PayloadSize)
		_, err = io.ReadFull(conn, payload)
		if err != nil {
			panic(err)
		}

		_, err = file.Write(payload)
		if err != nil {
			panic(err)
		}
	}
}
