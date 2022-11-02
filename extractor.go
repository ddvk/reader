package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/sirupsen/logrus"
)

func blah[K any, V any](k K, v V) {

}

type Lww[T any] struct {
	Value     T
	Timestamp CrdtId
	Flag      bool
}

type LwwBool struct {
	Value     bool
	Timestamp CrdtId
	Flag      bool
}
type LwwString struct {
	Value     string
	Timestamp CrdtId
	Flag      bool
}

type LwwCrdt struct {
	Value     CrdtId
	Timestamp CrdtId
	Flag      bool
}
type LwwByte struct {
	Value     byte
	Timestamp CrdtId
	Flag      bool
}
type LwwFloat struct {
	Value     float32
	Timestamp CrdtId
	Flag      bool
}

// func ExtractLwwAny[T any](control byte) (result Lww[T], found bool, err error) {
// 	result = Lww[T]{}
// 	if _, ok := any(result.Value).(string); ok {

// 		result.Value = "aa".(T)

// 	}
// 	return Lww[T]{}, false, nil
// }
func (decoder *Decoder) ExtractLwwByte(control byte) (result LwwByte, found bool, err error) {
	result = LwwByte{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	_, err = decoder.GetInt()
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
func (decoder *Decoder) ExtractLwwFloat(control byte) (result LwwFloat, found bool, err error) {
	result = LwwFloat{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	_, err = decoder.GetInt()
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

	result.Timestamp = timestamp
	result.Value = val

	return
}
func (decoder *Decoder) ExtractLwwCrdt(control byte) (result LwwCrdt, found bool, err error) {
	result = LwwCrdt{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	length, err := decoder.GetInt()
	if err != nil {
		return
	}
	logrus.Print("lwwcrd len: ", length)
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
func (decoder *Decoder) ExtractLwwBool(control byte) (result LwwBool, found bool, err error) {
	result = LwwBool{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}

	totalBytes, err := decoder.GetInt()
	if err != nil {
		return
	}
	logrus.Println("lwwbool, totalbytes:", totalBytes)
	if totalBytes == 0 {
		return
	}
	timeStamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}
	someBool, _, err := decoder.ExtractBool(2)
	if err != nil {
		return
	}
	result.Flag = someBool
	logrus.Println("lwwbool, ts", timeStamp, someBool)

	return
}

func (decoder *Decoder) checkTag(expectedIndex byte, tag TagType) (bool, error) {
	if decoder.hasPending {
		//the index doesnt match, continue
		if decoder.lastIndex != expectedIndex {
			return false, nil
		}

		if decoder.lastTag != tag {
			logrus.Errorf("lastTag != current,index:%d, have: %x, wants: %x", expectedIndex, decoder.lastTag, tag)
			return false, ErrTagMismatch
		}

		//the tag matches
		decoder.hasPending = false
		return true, nil
	}
	id, err := decoder.GetVarint()
	if err == io.EOF {
		logrus.Warn("EOF reading tag: ", tag)
		return false, nil
	}
	if err != nil {
		return false, err
	}

	index := byte(id >> 4 & 0xF)
	currentTag := TagType(id & 0xF)
	if index != expectedIndex {
		logrus.Warn("skipping index no match: ", index, tag)
		decoder.hasPending = true
		decoder.lastIndex = index
		decoder.lastTag = currentTag
		return false, nil
	}

	if currentTag != tag {
		return false, fmt.Errorf("tag mismatch tag: %x, expected:  %x", id, tag)
	}

	return true, nil
}

func (decoder *Decoder) ExtractLwwString(control byte) (result LwwString, found bool, err error) {
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}

	totalLength, err := decoder.GetInt()
	if err != nil {
		return
	}
	timestamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}

	log.Printf("got totalLength: %d timestamp: %x\n", totalLength, timestamp)
	if totalLength == 0 {
		return
	}

	//tagid
	_, err = decoder.checkTag(2, ItemTag)
	if err != nil {
		return
	}

	someInt, err := decoder.GetInt()
	if err != nil {
		return
	}
	log.Printf("got someInt??: %d \n", someInt)

	strLen, err := decoder.GetVarint()
	if err != nil {
		return
	}

	log.Printf("got strlen %d \n", strLen)
	isAscii, err := decoder.GetByte()
	if err != nil {
		return
	}
	if strLen == 0 {
		return
	}
	log.Printf("got isascii %d \n", isAscii)
	theStr, err := decoder.GetBytes(int(strLen))
	if err != nil {
		return
	}
	result.Value = string(theStr)
	result.Timestamp = timestamp
	log.Printf("got string %s \n", theStr)
	return
}

type TagType byte

const (
	ItemTag   TagType = 12
	CrdtTag   TagType = 15
	NumberTag TagType = 4
	BoolTag   TagType = 1
)

func (d *Decoder) ExtractInt(control byte) (result uint32, found bool, err error) {
	if found, err = d.checkTag(control, NumberTag); !found {
		return
	}
	result, err = d.GetInt()
	return
}
func (d *Decoder) ExtractBool(control byte) (result bool, found bool, err error) {
	if found, err = d.checkTag(control, BoolTag); !found {
		return
	}

	b, err := d.GetByte()
	result = b != 0
	return
}

func (d *Decoder) ExtractByte(control byte) (result byte, found bool, err error) {
	if found, err = d.checkTag(control, BoolTag); !found {
		return
	}
	result, err = d.GetByte()
	return
}
func (d *Decoder) ExtractFloat(control byte) (result float32, found bool, err error) {
	if found, err = d.checkTag(control, NumberTag); !found {
		return
	}
	err = binary.Read(d.r, binary.LittleEndian, &result)
	return
}
func (decoder *Decoder) ExtractCrdtId(index byte) (result CrdtId, found bool, err error) {
	if found, err = decoder.checkTag(index, CrdtTag); !found {
		return
	}
	short, err := decoder.GetVarint()
	if err != nil {
		return
	}
	part1 := short

	short, err = decoder.GetVarint()
	if err != nil {
		return
	}

	if short&0xFFFF0000 != 0 {
		return
	}
	part2 := uint64(short)
	result = CrdtId(uint64(part1)<<(16+32) | part2)
	return
}
