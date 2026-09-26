package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/atluixx/gftp/internal/protocol"
)

func main() {
	outputDir := flag.String(
		"output",
		".",
		"directory where the received file will be saved",
	)

	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Waiting for connection...")

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connected!")

	header := make([]byte, 17)
	var file *os.File

	for {
		_, err := io.ReadFull(conn, header)
		if err != nil {
			panic(err)
		}

		p, err := protocol.DecodeHeader(header)
		if err != nil {
			panic(err)
		}

		switch p.Type {

		case protocol.PacketFileInfo:
			payload := make([]byte, p.PayloadSize)

			_, err := io.ReadFull(conn, payload)
			if err != nil {
				panic(err)
			}

			fi, err := protocol.DecodeFileInfo(payload)
			if err != nil {
				panic(err)
			}

			fmt.Println("Receiving:", fi.Name)
			fmt.Println("File size:", fi.FileSize)

			outputPath := filepath.Join(*outputDir, fi.Name)

			file, err = os.Create(outputPath)
			if err != nil {
				panic(err)
			}

			fmt.Println("Saving to:", outputPath)

		case protocol.PacketChunk:
			payload := make([]byte, p.PayloadSize)

			_, err := io.ReadFull(conn, payload)
			if err != nil {
				panic(err)
			}

			if file == nil {
				panic("received chunk before FileInfo")
			}

			_, err = file.Write(payload)
			if err != nil {
				panic(err)
			}

		case protocol.PacketFileEnd:
			fmt.Println("Transfer finished!")

			if file != nil {
				if err := file.Close(); err != nil {
					panic(err)
				}
			}

			return

		default:
			panic("unknown packet type")
		}
	}
}
