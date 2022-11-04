package main

import (
	"encoding/binary"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

type Lww[T any] struct {
	Value     T
	Timestamp CrdtId
}

type LwwBool struct {
	Value     bool
	Timestamp CrdtId
}
type LwwString struct {
	Value     string
	Timestamp CrdtId
}

type LwwCrdt struct {
	Value     CrdtId
	Timestamp CrdtId
}
type LwwByte struct {
	Value     byte
	Timestamp CrdtId
}
type LwwFloat struct {
	Value     float32
	Timestamp CrdtId
}

// func ExtractLwwAny[T any](control byte) (result Lww[T], found bool, err error) {
// 	result = Lww[T]{}
// 	if _, ok := any(result.Value).(string); ok {

// 		result.Value = "aa".(T)

// 	}
// 	return Lww[T]{}, false, nil
// }
func (decoder *ChunkDecoder) ExtractLwwByte(control TagIndex) (result LwwByte, found bool, err error) {
	result = LwwByte{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	_, err = decoder.GetUInt32()
	if err != nil {
		return
	}
	timestamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}
	val, _, err := decoder.ExtractByte(2)
	if err != nil {
		return
	}
	result.Value = val
	result.Timestamp = timestamp
	return
}
func (decoder *ChunkDecoder) ExtractLwwFloat(control TagIndex) (result LwwFloat, found bool, err error) {
	result = LwwFloat{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	_, err = decoder.GetUInt32()
	if err != nil {
		return
	}
	timestamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}
	val, _, err := decoder.ExtractFloat(2)
	if err != nil {
		return
	}

	result.Value = val
	result.Timestamp = timestamp

	return
}
func (decoder *ChunkDecoder) ExtractLwwCrdt(control TagIndex) (result LwwCrdt, found bool, err error) {
	result = LwwCrdt{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	someInt, err := decoder.GetUInt32()
	if err != nil {
		return
	}
	pos := decoder.Pos()
	log.Printf("LwCrdt, Int?: %d, pos:%d, max:%d", someInt, pos, decoder.max)
	val, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}
	timeStamp, _, err := decoder.ExtractCrdtId(2)
	if err != nil {
		return
	}
	result.Value = val
	result.Timestamp = timeStamp
	return
}
func (decoder *ChunkDecoder) ExtractLwwBool(control TagIndex) (result LwwBool, found bool, err error) {
	result = LwwBool{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}

	someInt, err := decoder.GetUInt32()
	if err != nil {
		return
	}
	log.Info("lwwbool, someInt:", someInt)

	timeStamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}
	someBool, _, err := decoder.ExtractBool(2)
	if err != nil {
		return
	}
	result.Value = someBool
	result.Timestamp = timeStamp
	return
}

type tagInfo struct {
	TagIndex TagIndex
	TagId    TagId
}

//TODO: hack
const ignore = 0xFF

// checkTag reads a tag or a pending tag and advances if the index does not match
func (decoder *ChunkDecoder) checkTag(expectedIndex TagIndex, tag TagId) (bool, error) {
	if decoder.lastTag != nil {
		lastIndex := decoder.lastTag.TagIndex
		//the index doesnt match, continue
		if lastIndex != expectedIndex && expectedIndex != ignore {
			return false, nil
		}

		lastTag := decoder.lastTag.TagId
		if lastTag != tag {
			log.Errorf("lastTag != current,index:%d, have: %x, wants: %x", expectedIndex, lastTag, tag)
			return false, ErrTagMismatch
		}

		//the tag matches
		decoder.lastTag = nil
		return true, nil
	}
	id, err := decoder.GetVarUInt32()
	if err == io.EOF {
		//TODO: handle better
		// logrus.Warn("EOF reading tag: ", tag)
		return false, nil
	}
	if err != nil {
		return false, err
	}

	index := TagIndex(id >> 4)
	currentTag := TagId(id & 0xF)
	if index != expectedIndex && expectedIndex != ignore {
		log.Trace("skipping index no match: ", index, tag)
		decoder.lastTag = &tagInfo{
			TagIndex: index,
			TagId:    currentTag,
		}
		return false, nil
	}

	if currentTag != tag {
		return false, fmt.Errorf("tag mismatch tag: %x, expected:  %x", id, tag)
	}

	return true, nil
}

func (decoder *ChunkDecoder) ExtractLwwString(index TagIndex) (result LwwString, found bool, err error) {
	if found, err = decoder.checkTag(index, ItemTag); !found {
		return
	}

	elementLength, err := decoder.GetUInt32()
	if err != nil {
		return
	}
	pos := decoder.Pos()
	endPos := pos + int(elementLength)
	log.Printf("LwwString, Int?: %d, pos:%d, max:%d", elementLength, pos, decoder.max)
	timestamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}

	_, err = decoder.checkTag(2, ItemTag)
	if err != nil {
		return
	}

	someInt2, err := decoder.GetUInt32()
	if err != nil {
		return
	}
	log.Printf("Lww someInt2??: %d ", someInt2)

	strLen, err := decoder.GetVarUInt32()
	if err != nil {
		return
	}

	log.Printf("got strlen %d ", strLen)
	isAscii, err := decoder.GetByte()
	if err != nil {
		return
	}
	// if strLen == 0 {
	// 	return
	// }
	log.Printf("got isascii %d ", isAscii)
	theStr, err := decoder.GetBytes(int(strLen))
	if err != nil {
		return
	}
	result.Value = string(theStr)
	result.Timestamp = timestamp
	log.Printf("got string: '%s'", theStr)
	pos = decoder.Pos()
	if pos > endPos {
		err = fmt.Errorf("buffer overflow, pos: %d max: %d", pos, endPos)
		return
	}

	log.Printf("LwwStringEnd, pos:%d, max:%d", pos, decoder.max)
	return
}

type TagId byte
type TagIndex int16

const (
	CrdtTag   TagId = 0xf
	ItemTag   TagId = 0xc
	NumberTag TagId = 4
	BoolTag   TagId = 1
)

func (d *ChunkDecoder) ExtractInt(index TagIndex) (result uint32, found bool, err error) {
	if found, err = d.checkTag(index, NumberTag); !found {
		return
	}
	result, err = d.GetUInt32()
	return
}
func (d *ChunkDecoder) ExtractBool(index TagIndex) (result bool, found bool, err error) {
	if found, err = d.checkTag(index, BoolTag); !found {
		return
	}

	b, err := d.GetByte()
	result = b != 0
	return
}

func (d *ChunkDecoder) ExtractByte(index TagIndex) (result byte, found bool, err error) {
	if found, err = d.checkTag(index, BoolTag); !found {
		return
	}
	result, err = d.GetByte()
	return
}
func (d *ChunkDecoder) ExtractFloat(index TagIndex) (result float32, found bool, err error) {
	if found, err = d.checkTag(index, NumberTag); !found {
		return
	}
	err = binary.Read(d, binary.LittleEndian, &result)
	return
}
func (decoder *ChunkDecoder) ExtractCrdtId(index TagIndex) (result CrdtId, found bool, err error) {
	if found, err = decoder.checkTag(index, CrdtTag); !found {
		return
	}
	short, err := decoder.GetVarUInt32()
	if err != nil {
		log.Error("can't get short1")
		return
	}
	part1 := short

	short2, err := decoder.GetVarInt64()
	if err != nil {
		log.Error("can't get short2")
		return
	}

	if short2&0xFFFF0000 != 0 {
		log.Warnf("short1: %x, short2 > fmask true, %x", short, short2)
		return
	}
	part2 := uint64(short)
	result = CrdtId(uint64(part1)<<(16+32) | part2)
	return
}

func (d *ChunkDecoder) ExtractBob() (bob []byte, err error) {
	bobLength := d.max - d.position
	if bobLength > 0 {
		bob, err = d.GetBytes(bobLength)
		if err != nil {
			return
		}
	}
	return
}

type Chunk struct {
	Size   int32
	Header Header
}

func ExtractChunk(reader io.Reader, index int) (ch Chunk, err error) {
	var size int32
	err = binary.Read(reader, binary.LittleEndian, &size)
	if err != nil {
		return
	}
	buffer := make([]byte, 4)
	// var tag uint32
	_, err = io.ReadFull(reader, buffer[:4])
	if err != nil {
		return
	}

	// b1 := tag >> 0x10 & 0xF
	// b2 := tag >> 0x8 & 0xF
	header := Header{
		TagID: buffer[3],
		B1:    buffer[2],
		B2:    buffer[1],
	}
	if err != nil {
		return
	}

	ch = Chunk{
		Size:   size,
		Header: header,
	}
	return
}
