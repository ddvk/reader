package main

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

// NewDecoder returns a chunk decoder with the specified limit
func NewDecoder(r io.Reader, limit int) *ChunkDecoder {
	decoder := &ChunkDecoder{
		r:   bufio.NewReader(io.LimitReader(r, int64(limit))),
		max: limit,
	}
	return decoder
}
func NewBytesDecoder(buffer []byte) *ChunkDecoder {
	limit := len(buffer)
	reader := bytes.NewBuffer(buffer)
	decoder := &ChunkDecoder{
		r:   reader,
		max: limit,
	}
	return decoder
}

// ChunkDecoder
type ChunkDecoder struct {
	r        reader
	lastTag  *tagInfo
	position int
	max      int
}

// Pos current position
func (d *ChunkDecoder) Pos() int {
	return d.position
}

func (d *ChunkDecoder) Read(b []byte) (n int, err error) {
	n, err = d.r.Read(b)
	d.position += n
	return
}
func (d *ChunkDecoder) ReadByte() (b byte, err error) {
	b, err = d.r.ReadByte()
	if err != nil {
		return b, err
	}
	d.position += 1
	return
}

func (d *ChunkDecoder) GetByte() (byte, error) {
	return d.ReadByte()
}

func (d *ChunkDecoder) GetBytes(size int) (result []byte, err error) {
	result = make([]byte, size)
	_, err = io.ReadFull(d, result)
	return
}
func (d *ChunkDecoder) GetShort() (result uint16, err error) {
	err = binary.Read(d.r, binary.LittleEndian, &result)
	return
}

func (d *ChunkDecoder) GetVarUInt32() (result uint32, err error) {
	val, err := binary.ReadUvarint(d)
	if val > math.MaxUint32 {
		return 0, fmt.Errorf("uint32 exceeded, %x, position: %d", val, d.position)
	}
	return uint32(val), err
}

func (d *ChunkDecoder) GetVarInt64() (uint64, error) {
	return binary.ReadUvarint(d)
}
func (d *ChunkDecoder) GetUInt32() (val uint32, err error) {
	err = binary.Read(d, binary.LittleEndian, &val)
	return
}

func DebugBuffer(buffer []byte, pos, max int) {
	fmt.Println(hex.EncodeToString(buffer))
	padding := ""
	if pos > 0 {
		padding = strings.Repeat("  ", pos-1)
	}
	fmt.Printf("%s ^  pos: %d (max: %d)\n", padding, pos, max)
}
