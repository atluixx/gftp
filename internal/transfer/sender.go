package transfer

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/atluixx/gftp/internal/protocol"
)

const ChunkSize = 1024

func EncodeFile(path string, conn net.Conn) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	b := make([]byte, ChunkSize)
	i := 0
	fileSize := info.Size()

	for {
		n, err := file.Read(b)

		if n > 0 {
			c := protocol.Chunk{
				Index:    uint32(i),
				FileSize: fileSize,
				Data:     b[:n],
			}

			if err := SendChunk(c, conn); err != nil {
				return err
			}

			i += 1
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

	}

	return nil
}

func EncodeChunk(c protocol.Chunk) ([]byte, error) {
	p := protocol.Packet{
		Index:       c.Index,
		FileSize:    c.FileSize,
		PayloadSize: uint32(len(c.Data)),
		Payload:     c.Data,
	}

	encoded, err := protocol.EncodePacket(p)
	if err != nil {
		return nil, err
	}

	fmt.Println("packet size:", len(encoded))
	fmt.Println("payload size:", len(p.Payload))

	return encoded, nil
}

func SendChunk(c protocol.Chunk, conn net.Conn) error {
	data, err := EncodeChunk(c)
	if err != nil {
		return err
	}

	_, err = io.Copy(conn, bytes.NewReader(data))
	return err
}
