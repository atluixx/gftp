package main

import (
	"flag"
	"net"

	"github.com/atluixx/gftp/internal/transfer"
)

func main() {
	filePath := flag.String("file", "", "path to the file to send")

	flag.Parse()

	if *filePath == "" {
		flag.Usage()
		return
	}

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := transfer.EncodeFile(*filePath, conn); err != nil {
		panic(err)
	}
}
