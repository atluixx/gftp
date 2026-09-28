// Package transfer implements gftp's reliable application protocol.
package transfer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/atluixx/gftp/internal/protocol"
)

const ChunkSize = 32 * 1024

type Progress func(name string, sent, total int64)

func writePacket(conn net.Conn, p protocol.Packet) error {
	b, err := protocol.EncodePacket(p)
	if err != nil {
		return err
	}
	for len(b) > 0 {
		n, err := conn.Write(b)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}
func ReadPacket(conn net.Conn) (protocol.Packet, error) {
	h := make([]byte, protocol.HeaderSize)
	if _, err := io.ReadFull(conn, h); err != nil {
		return protocol.Packet{}, err
	}
	p, err := protocol.DecodeHeader(h)
	if err != nil {
		return p, err
	}
	p.Payload = make([]byte, p.PayloadSize)
	_, err = io.ReadFull(conn, p.Payload)
	return p, err
}
func WritePacket(conn net.Conn, p protocol.Packet) error { return writePacket(conn, p) }

func Authenticate(conn net.Conn, token string) error {
	if token == "" {
		return nil
	}
	if err := writePacket(conn, protocol.Packet{Type: protocol.PacketAuth, PayloadSize: uint32(len(token)), Payload: []byte(token)}); err != nil {
		return err
	}
	p, err := ReadPacket(conn)
	if err != nil {
		return err
	}
	if p.Type == protocol.PacketError {
		return fmt.Errorf("authentication rejected: %s", p.Payload)
	}
	if p.Type != protocol.PacketAck {
		return fmt.Errorf("expected authentication acknowledgement")
	}
	return nil
}

func EncodeFile(path string, conn net.Conn) error {
	return EncodeFileWithProgress(path, filepath.Base(path), conn, nil)
}
func EncodeFileWithProgress(path, name string, conn net.Conn, progress Progress) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}
	if err := SendFileInfo(protocol.FileInfo{Name: filepath.ToSlash(name), FileSize: info.Size()}, conn); err != nil {
		return err
	}
	ack, err := ReadPacket(conn)
	if err != nil {
		return err
	}
	if ack.Type == protocol.PacketError {
		return fmt.Errorf("receiver rejected %s: %s", name, ack.Payload)
	}
	if ack.Type != protocol.PacketAck || len(ack.Payload) != 8 {
		return fmt.Errorf("expected resume acknowledgement")
	}
	offset := int64(binary.BigEndian.Uint64(ack.Payload))
	if offset < 0 || offset > info.Size() {
		return fmt.Errorf("invalid resume offset %d", offset)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return err
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	b := make([]byte, ChunkSize)
	index, sent := uint32(offset/ChunkSize), offset
	for {
		n, er := f.Read(b)
		if n > 0 {
			if err := SendChunk(protocol.Chunk{Index: index, FileSize: info.Size(), Data: b[:n]}, conn); err != nil {
				return err
			}
			index++
			sent += int64(n)
			if progress != nil {
				progress(name, sent, info.Size())
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			return er
		}
	}
	if err := SendFileEndChecksum(conn, hash.Sum(nil)); err != nil {
		return err
	}
	ack, err = ReadPacket(conn)
	if err != nil {
		return err
	}
	if ack.Type == protocol.PacketError {
		return fmt.Errorf("transfer rejected: %s", ack.Payload)
	}
	if ack.Type != protocol.PacketAck {
		return fmt.Errorf("expected transfer acknowledgement")
	}
	return nil
}
func EncodePaths(paths []string, conn net.Conn, progress Progress) error {
	for _, source := range paths {
		info, err := os.Stat(source)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if err := EncodeFileWithProgress(source, filepath.Base(source), conn, progress); err != nil {
				return err
			}
			continue
		}
		root := filepath.Dir(source)
		err = filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			return EncodeFileWithProgress(path, rel, conn, progress)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func EncodeChunk(c protocol.Chunk) ([]byte, error) {
	return protocol.EncodePacket(protocol.Packet{Type: protocol.PacketChunk, Index: c.Index, FileSize: c.FileSize, PayloadSize: uint32(len(c.Data)), Payload: c.Data})
}
func SendChunk(c protocol.Chunk, conn net.Conn) error {
	return writePacket(conn, protocol.Packet{Type: protocol.PacketChunk, Index: c.Index, FileSize: c.FileSize, PayloadSize: uint32(len(c.Data)), Payload: c.Data})
}
func EncodeFileInfo(fi protocol.FileInfo) ([]byte, error) {
	p, e := protocol.EncodeFileInfo(fi)
	if e != nil {
		return nil, e
	}
	return protocol.EncodePacket(protocol.Packet{Type: protocol.PacketFileInfo, FileSize: fi.FileSize, PayloadSize: uint32(len(p)), Payload: p})
}
func SendFileInfo(fi protocol.FileInfo, conn net.Conn) error {
	p, e := protocol.EncodeFileInfo(fi)
	if e != nil {
		return e
	}
	return writePacket(conn, protocol.Packet{Type: protocol.PacketFileInfo, FileSize: fi.FileSize, PayloadSize: uint32(len(p)), Payload: p})
}
func SendFileEnd(conn net.Conn) error { return SendFileEndChecksum(conn, nil) }
func SendFileEndChecksum(conn net.Conn, sum []byte) error {
	return writePacket(conn, protocol.Packet{Type: protocol.PacketFileEnd, PayloadSize: uint32(len(sum)), Payload: sum})
}
