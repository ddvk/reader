package v6

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strings"
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

// Pos current position
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

func (d *BinaryDeserializer) GetByte() (byte, error) {
	return d.ReadByte()
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

//TODO: LEB128
func (d *BinaryDeserializer) GetVarUInt32() (result uint32, err error) {
	val, err := binary.ReadUvarint(d)
	if val > math.MaxUint32 {
		return 0, fmt.Errorf("uint32 exceeded, %x, position: %d", val, d.position)
	}
	return uint32(val), err
}

func (d *BinaryDeserializer) GetVarInt64() (uint64, error) {
	return binary.ReadUvarint(d)
}
func (d *BinaryDeserializer) GetUInt32() (val uint32, err error) {
	err = binary.Read(d, binary.LittleEndian, &val)
	return
}

func DebugBuffer(buffer []byte, pos, max int) {
	fmt.Println(hex.EncodeToString(buffer))
	padding := ""
	if pos > 0 {
		padding = strings.Repeat("  ", pos-1)
	}
	fmt.Printf("%s ^  pos: %d (max: %d x%x)\n", padding, pos, max, max)
}
