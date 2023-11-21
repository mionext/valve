package socket

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

var ErrOutOfBounds = errors.New("read out of bounds")

type PacketBuilder struct {
	bytes.Buffer
}

func (p *PacketBuilder) WriteCString(s string) {
	p.Write([]byte(s))
	p.WriteByte(0)
}

func (p *PacketBuilder) WriteBytes(bytes []byte) {
	p.Write(bytes)
}

type PacketReader struct {
	buffer []byte
	pos    int
}

func NewPacketReader(bytes []byte) *PacketReader {
	return &PacketReader{buffer: bytes, pos: 0}
}

// CanRead returns true if there are at least size bytes left to read
func (p *PacketReader) CanRead(size int) bool {
	return len(p.buffer)-p.pos >= size
}

// Split returns the remaining bytes
func (p *PacketReader) Split(count int) []byte {
	if !p.CanRead(count) {
		return nil
	}

	data := p.buffer[p.pos : p.pos+count]
	p.pos += count

	return data
}

// Pos returns the current position
func (p *PacketReader) Pos() int {
	return p.pos
}

// ReadIPv4 reads an IPv4 address
func (p *PacketReader) ReadIPv4() (net.IP, error) {
	if !p.CanRead(4) {
		return nil, ErrOutOfBounds
	}

	ip := net.IP(p.buffer[p.pos : p.pos+net.IPv4len])
	p.pos += net.IPv4len

	return ip, nil
}

// ReadPort reads a port
func (p *PacketReader) ReadPort() (uint16, error) {
	if !p.CanRead(2) {
		return 0, ErrOutOfBounds
	}

	port := binary.BigEndian.Uint16(p.buffer[p.pos : p.pos+2])
	p.pos += 2

	return port, nil
}

// ReadUint8 reads a uint8
func (p *PacketReader) ReadUint8() uint8 {
	value := p.buffer[p.pos]
	p.pos++

	return value
}

// ReadUint16 reads a uint16
func (p *PacketReader) ReadUint16() uint16 {
	value := binary.LittleEndian.Uint16(p.buffer[p.pos : p.pos+2])
	p.pos += 2

	return value
}

// ReadUint32 reads a uint32
func (p *PacketReader) ReadUint32() uint32 {
	value := binary.LittleEndian.Uint32(p.buffer[p.pos : p.pos+4])
	p.pos += 4

	return value
}

// ReadInt32 reads an int32
func (p *PacketReader) ReadInt32() int32 {
	return int32(p.ReadUint32())
}

// ReadUint64 reads a uint64
func (p *PacketReader) ReadUint64() uint64 {
	value := binary.LittleEndian.Uint64(p.buffer[p.pos : p.pos+8])
	p.pos += 8

	return value
}

// TryReadString reads a string
func (p *PacketReader) TryReadString() (string, error) {
	start := p.pos
	for p.pos < len(p.buffer) {
		if p.buffer[p.pos] == 0 {
			p.pos++
			return string(p.buffer[start : p.pos-1]), nil
		}

		p.pos++
	}

	return "", ErrOutOfBounds
}

// ReadString reads a string
func (p *PacketReader) ReadString() string {
	start := p.pos
	for {
		// Note: it's intended that we panic for strings that are not null
		// terminated.
		if p.buffer[p.pos] == 0 {
			p.pos++
			break
		}
		p.pos++
	}

	return string(p.buffer[start : p.pos-1])
}

// More returns true if there are more bytes to read
func (p *PacketReader) More() bool {
	return p.pos < len(p.buffer)
}
