package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type PacketType uint8

const (
	PacketFileInfo PacketType = 1
	PacketChunk    PacketType = 2
	PacketFileEnd  PacketType = 3
	PacketAck      PacketType = 4
	PacketAuth     PacketType = 5
	PacketError    PacketType = 6
)

const HeaderSize = 17
const MaxPayloadSize = 16 << 20

type Packet struct {
	Type        PacketType
	Index       uint32
	FileSize    int64
	PayloadSize uint32
	Payload     []byte
}

func EncodePacket(p Packet) ([]byte, error) {
	if uint64(len(p.Payload)) > uint64(MaxPayloadSize) || p.PayloadSize != uint32(len(p.Payload)) {
		return nil, fmt.Errorf("invalid packet payload size")
	}
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, p.Type); err != nil {
		return nil, err
	}

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

	if err := binary.Read(reader, binary.BigEndian, &p.Type); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.Index); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.FileSize); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.PayloadSize); err != nil {
		return p, err
	}
	if p.PayloadSize > MaxPayloadSize {
		return p, fmt.Errorf("packet payload too large: %d", p.PayloadSize)
	}

	p.Payload = make([]byte, p.PayloadSize)

	if _, err := io.ReadFull(reader, p.Payload); err != nil {
		return p, err
	}

	return p, nil
}

func DecodeHeader(data []byte) (Packet, error) {
	var p Packet
	if len(data) != HeaderSize {
		return p, fmt.Errorf("invalid packet header size: %d", len(data))
	}
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &p.Type); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.Index); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.FileSize); err != nil {
		return p, err
	}

	if err := binary.Read(reader, binary.BigEndian, &p.PayloadSize); err != nil {
		return p, err
	}
	if p.PayloadSize > MaxPayloadSize {
		return p, fmt.Errorf("packet payload too large: %d", p.PayloadSize)
	}

	return p, nil
}
