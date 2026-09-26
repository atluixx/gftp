package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
)

type FileInfo struct {
	Name     string
	FileSize int64
}

func DecodeFileInfo(data []byte) (FileInfo, error) {
	var fi FileInfo

	reader := bytes.NewReader(data)

	var nameLength uint16

	if err := binary.Read(reader, binary.BigEndian, &nameLength); err != nil {
		return fi, err
	}

	name := make([]byte, nameLength)

	if _, err := io.ReadFull(reader, name); err != nil {
		return fi, err
	}

	fi.Name = string(name)

	if err := binary.Read(reader, binary.BigEndian, &fi.FileSize); err != nil {
		return fi, err
	}

	return fi, nil
}

func EncodeFileInfo(fi FileInfo) ([]byte, error) {
	var buf bytes.Buffer

	nameBytes := []byte(fi.Name)

	if err := binary.Write(&buf, binary.BigEndian, uint16(len(nameBytes))); err != nil {
		return nil, err
	}

	if _, err := buf.Write(nameBytes); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, fi.FileSize); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
