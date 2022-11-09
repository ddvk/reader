package v6

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type reader interface {
	io.ByteReader
	io.Reader
}

// BinaryDeserializer
type BinaryDeserializer struct {
	_r       reader
	position int
	max      int
}

// NewDecoder returns a chunk decoder with the specified limit
func NewDecoder(r io.Reader, limit int) *BinaryDeserializer {
	decoder := &BinaryDeserializer{
		_r:  bufio.NewReader(io.LimitReader(r, int64(limit))),
		max: limit,
	}
	return decoder
}
func NewDeserializer(buffer []byte) *BinaryDeserializer {
	limit := len(buffer)
	reader := bytes.NewBuffer(buffer)
	decoder := &BinaryDeserializer{
		_r:  reader,
		max: limit,
	}
	return decoder
}

// Pos current position in the stream
func (d *BinaryDeserializer) Pos() int {
	return d.position
}

func (d *BinaryDeserializer) Limit() int {
	return d.max
}

func (d *BinaryDeserializer) Read(b []byte) (n int, err error) {
	n, err = d._r.Read(b)
	d.position += n
	return
}
func (d *BinaryDeserializer) ReadByte() (b byte, err error) {
	b, err = d._r.ReadByte()
	if err != nil {
		return b, err
	}
	d.position += 1
	return
}

func (d *BinaryDeserializer) GetBytes(size int) (result []byte, err error) {
	result = make([]byte, size)
	_, err = io.ReadFull(d, result)
	return
}
func (d *BinaryDeserializer) GetShort() (result uint16, err error) {
	err = binary.Read(d, binary.LittleEndian, &result)
	return
}

func (d *BinaryDeserializer) GetFloat32() (result float32, err error) {
	err = binary.Read(d, binary.LittleEndian, &result)
	return
}

func (d *BinaryDeserializer) GetFloat64() (result float64, err error) {
	err = binary.Read(d, binary.LittleEndian, &result)
	return
}

func (d *BinaryDeserializer) GetVarUInt32() (result uint32, err error) {
	val, err := binary.ReadUvarint(d)
	if val > math.MaxUint32 {
		return 0, fmt.Errorf("uint32 exceeded, %x, position: %d", val, d.position)
	}
	return uint32(val), err
}

func (d *BinaryDeserializer) GetVarUInt64() (uint64, error) {
	return binary.ReadUvarint(d)
}
func (d *BinaryDeserializer) GetFixedUInt32() (val uint32, err error) {
	err = binary.Read(d, binary.LittleEndian, &val)
	return
}

func (d *BinaryDeserializer) GetFixedInt32() (val int32, err error) {
	err = binary.Read(d, binary.LittleEndian, &val)
	return
}
