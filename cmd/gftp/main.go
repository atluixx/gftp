package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/atluixx/gftp/internal/protocol"
	"github.com/atluixx/gftp/internal/transfer"
)

const address = ":8080"

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
		fmt.Fprintln(flag.CommandLine.Output(), "  gftp send -file <filename>")
		fmt.Fprintln(flag.CommandLine.Output(), "  gftp receive [-output <directory/[filename]>]")
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "send":
		err = runSend(os.Args[2:])
	case "receive":
		err = runReceive(os.Args[2:])
	default:
		flag.Usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "gftp:", err)
		os.Exit(1)
	}
}

func runSend(args []string) error {
	flags := flag.NewFlagSet("send", flag.ContinueOnError)
	filePath := flags.String("file", "", "path to the file to send")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *filePath == "" {
		flags.Usage()
		return errors.New("-file is required")
	}
	return sendFile(*filePath)
}

func runReceive(args []string) error {
	flags := flag.NewFlagSet("receive", flag.ContinueOnError)
	outputPath := flags.String("output", ".", "output directory, optionally followed by a filename")
	if err := flags.Parse(args); err != nil {
		return err
	}
	return receiveFile(*outputPath)
}

func sendFile(filePath string) error {
	conn, err := net.Dial("tcp", "localhost"+address)
	if err != nil {
		return err
	}
	defer conn.Close()

	return transfer.EncodeFile(filePath, conn)
}

func receiveFile(outputPath string) error {
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

	return decodeFile(conn, outputPath)
}

func decodeFile(conn net.Conn, requestedOutput string) error {
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

			outputPath, err := resolveOutputPath(requestedOutput, fileInfo.Name)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return err
			}
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

func resolveOutputPath(requestedOutput, receivedName string) (string, error) {
	if requestedOutput == "" {
		return "", errors.New("output path cannot be empty")
	}

	isDirectory := strings.HasSuffix(requestedOutput, string(os.PathSeparator))
	if info, err := os.Stat(requestedOutput); err == nil {
		isDirectory = info.IsDir()
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if isDirectory {
		return filepath.Join(requestedOutput, filepath.Base(receivedName)), nil
	}
	return requestedOutput, nil
}
