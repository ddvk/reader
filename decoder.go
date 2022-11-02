package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

func NewDecoder(r *bytes.Reader) *Decoder {
	return &Decoder{
		r:          r,
		underlying: r,
	}
}

type Decoder struct {
	r          *bytes.Reader
	underlying io.Reader
	hasPending bool
	lastIndex  byte
	lastTag    TagType
	position   int
}

func (d *Decoder) Info() int {
	pos, _ := d.r.Seek(0, io.SeekCurrent)
	return int(pos)
}

func (d *Decoder) GetByte() (byte, error) {
	var b byte
	err := binary.Read(d.r, binary.LittleEndian, &b)
	return b, err
}

func (d *Decoder) GetBytes(size int) ([]byte, error) {
	buffer := make([]byte, size)
	_, err := io.ReadFull(d.r, buffer)
	return buffer, err
}
func (d *Decoder) GetShort() (uint16, error) {
	var b uint16
	err := binary.Read(d.r, binary.LittleEndian, &b)
	return b, err
}
func (d *Decoder) GetVarint() (uint32, error) {
	b, err := binary.ReadUvarint(d.r)
	return uint32(b), err
}

func (d *Decoder) GetInt() (uint32, error) {
	var b uint32
	err := binary.Read(d.r, binary.LittleEndian, &b)
	return b, err
}
