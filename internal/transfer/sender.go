package transfer

import (
	"bytes"
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

	// Send file information first.
	fileInfo := protocol.FileInfo{
		Name:     info.Name(),
		FileSize: info.Size(),
	}

	if err := SendFileInfo(fileInfo, conn); err != nil {
		return err
	}

	// Send file chunks.
	b := make([]byte, ChunkSize)
	var i uint32

	for {
		n, err := file.Read(b)

		if n > 0 {
			chunk := protocol.Chunk{
				Index:    i,
				FileSize: info.Size(),
				Data:     b[:n],
			}

			if err := SendChunk(chunk, conn); err != nil {
				return err
			}

			i++
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	// Tell the receiver that the transfer is finished.
	if err := SendFileEnd(conn); err != nil {
		return err
	}

	return nil
}

func EncodeChunk(c protocol.Chunk) ([]byte, error) {
	p := protocol.Packet{
		Type:        protocol.PacketChunk,
		Index:       c.Index,
		FileSize:    c.FileSize,
		PayloadSize: uint32(len(c.Data)),
		Payload:     c.Data,
	}

	return protocol.EncodePacket(p)
}

func SendChunk(c protocol.Chunk, conn net.Conn) error {
	data, err := EncodeChunk(c)
	if err != nil {
		return err
	}

	_, err = io.Copy(conn, bytes.NewReader(data))
	return err
}

func EncodeFileInfo(fi protocol.FileInfo) ([]byte, error) {
	payload, err := protocol.EncodeFileInfo(fi)
	if err != nil {
		return nil, err
	}

	p := protocol.Packet{
		Type:        protocol.PacketFileInfo,
		Index:       0,
		FileSize:    fi.FileSize,
		PayloadSize: uint32(len(payload)),
		Payload:     payload,
	}

	return protocol.EncodePacket(p)
}

func SendFileInfo(fi protocol.FileInfo, conn net.Conn) error {
	data, err := EncodeFileInfo(fi)
	if err != nil {
		return err
	}

	_, err = io.Copy(conn, bytes.NewReader(data))
	return err
}

func SendFileEnd(conn net.Conn) error {
	p := protocol.Packet{
		Type:        protocol.PacketFileEnd,
		Index:       0,
		FileSize:    0,
		PayloadSize: 0,
		Payload:     nil,
	}

	data, err := protocol.EncodePacket(p)
	if err != nil {
		return err
	}

	_, err = io.Copy(conn, bytes.NewReader(data))
	return err
}
