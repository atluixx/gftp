package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Packet struct {
	Index       uint32
	FileSize    int64
	PayloadSize uint32
	Payload     []byte
}

type PacketHeader struct {
	Index       uint32
	FileSize    int64
	PayloadSize uint32
}

func EncodePacket(p Packet) ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, p.Index); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, p.FileSize); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, p.PayloadSize); err != nil {
		return nil, err
	}

	if _, err := buf.Write(p.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodePacket(data []byte) (Packet, error) {
	var p Packet
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &p.Index); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.FileSize); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.PayloadSize); err != nil {
		return p, err
	}

	p.Payload = make([]byte, p.PayloadSize)

	if _, err := io.ReadFull(reader, p.Payload); err != nil {
		return p, err
	}

	return p, nil
}

func DecodeHeader(data []byte) (PacketHeader, error) {
	var ph PacketHeader
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &ph.Index); err != nil {
		return ph, err
	}

	if err := binary.Read(reader, binary.BigEndian, &ph.FileSize); err != nil {
		return ph, err
	}
	if err := binary.Read(reader, binary.BigEndian, &ph.PayloadSize); err != nil {
		return ph, err
	}

	return ph, nil
}
