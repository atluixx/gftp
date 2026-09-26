package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/atluixx/gftp/internal/protocol"
	"github.com/atluixx/gftp/internal/transfer"
)

const address = ":8080"

func main() {
	receive := flag.Bool("receive", false, "receive a file instead of sending one")
	filePath := flag.String("file", "", "path to the file to send")
	outputDir := flag.String("output", ".", "directory where the received file will be saved")
	flag.Parse()

	var err error
	if *receive {
		err = receiveFile(*outputDir)
	} else {
		if *filePath == "" {
			flag.Usage()
			os.Exit(2)
		}
		err = sendFile(*filePath)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "gftp:", err)
		os.Exit(1)
	}
}

func sendFile(filePath string) error {
	conn, err := net.Dial("tcp", "localhost"+address)
	if err != nil {
		return err
	}
	defer conn.Close()

	return transfer.EncodeFile(filePath, conn)
}

func receiveFile(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Waiting for connection...")
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Println("Connected!")

	return decodeFile(conn, outputDir)
}

func decodeFile(conn net.Conn, outputDir string) error {
	header := make([]byte, 17)
	var file *os.File
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()

	for {
		if _, err := io.ReadFull(conn, header); err != nil {
			return err
		}

		packet, err := protocol.DecodeHeader(header)
		if err != nil {
			return err
		}

		payload := make([]byte, packet.PayloadSize)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return err
		}

		switch packet.Type {
		case protocol.PacketFileInfo:
			fileInfo, err := protocol.DecodeFileInfo(payload)
			if err != nil {
				return err
			}

			fmt.Println("Receiving:", fileInfo.Name)
			fmt.Println("File size:", fileInfo.FileSize)

			outputPath := filepath.Join(outputDir, filepath.Base(fileInfo.Name))
			file, err = os.Create(outputPath)
			if err != nil {
				return err
			}
			fmt.Println("Saving to:", outputPath)

		case protocol.PacketChunk:
			if file == nil {
				return errors.New("received chunk before file info")
			}
			if _, err := file.Write(payload); err != nil {
				return err
			}

		case protocol.PacketFileEnd:
			if file != nil {
				if err := file.Close(); err != nil {
					return err
				}
				file = nil
			}
			fmt.Println("Transfer finished!")
			return nil

		default:
			return fmt.Errorf("unknown packet type: %d", packet.Type)
		}
	}
}
