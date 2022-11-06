package v6

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ElementTag byte
type TagIndex int16

const (
	CrdtTag   ElementTag = 0xf
	ItemTag   ElementTag = 0xc
	DoubleTag ElementTag = 8
	NumberTag ElementTag = 4
	BoolTag   ElementTag = 1
)

type Lww[T any] struct {
	Value     T
	Timestamp CrdtId
}

type Extractor struct {
	d       *BinaryDeserializer
	lastTag *tagInfo
	buffer  []byte
}

func NewExtractor(d *BinaryDeserializer) *Extractor {
	return &Extractor{
		d: d,
	}
}

// func ExtractLwwAny[T any](decoder *ChunkDecoder, index TagIndex) (result Lww[T], found bool, err error) {
// 	if found, err = decoder.checkTag(index, ItemTag); !found {
// 		return
// 	}
// 	_, err = decoder.GetUInt32()
// 	if err != nil {
// 		return
// 	}
// 	timestamp, _, err := decoder.ExtractCrdtId(1)
// 	if err != nil {
// 		return
// 	}
// 	val, _, err := decoder.ExtractByte(2)
// 	if err != nil {
// 		return
// 	}
// 	result.Value = val
// 	result.Timestamp = timestamp
// 	return
// 	binary.Read(nil, binary.LittleEndian, result.Value)
// }
func (decoder *Extractor) ExtractLwwByte(control TagIndex) (result Lww[byte], found bool, err error) {
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	_, err = decoder.d.GetUInt32()
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
func (decoder *Extractor) ExtractLwwFloat(control TagIndex) (result Lww[float32], found bool, err error) {
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	_, err = decoder.d.GetUInt32()
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
func (decoder *Extractor) ExtractLwwCrdt(control TagIndex) (result Lww[CrdtId], found bool, err error) {
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}
	someInt, err := decoder.d.GetUInt32()
	if err != nil {
		return
	}
	pos := decoder.d.Pos()
	log.Printf("LwCrdt, Int?: %d, pos:%d, max:%d", someInt, pos, decoder.d.max)
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
func (decoder *Extractor) ExtractLwwBool(control TagIndex) (result Lww[bool], found bool, err error) {
	result = Lww[bool]{}
	if found, err = decoder.checkTag(control, ItemTag); !found {
		return
	}

	someInt, err := decoder.d.GetUInt32()
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
	TagId    ElementTag
}

//TODO: hack
const ignoreTagIndex = -1

// checkTag reads a tag or a pending tag and advances if the index does not match
func (e *Extractor) checkTag(expectedIndex TagIndex, tag ElementTag) (bool, error) {
	if e.lastTag != nil {
		lastIndex := e.lastTag.TagIndex
		//the index doesnt match, continue
		if lastIndex != expectedIndex && expectedIndex != ignoreTagIndex {
			return false, nil
		}

		lastTag := e.lastTag.TagId
		if lastTag != tag {
			log.Errorf("lastTag != current,index:%d, have: %x, wants: %x", expectedIndex, lastTag, tag)
			return false, ErrTagMismatch
		}

		//the tag matches
		e.lastTag = nil
		return true, nil
	}
	id, err := e.d.GetVarUInt32()
	if err == io.EOF {
		//TODO: no more tags in the stream
		//logrus.Warn("EOF reading tag: ", tag)
		return false, nil
	}
	log.Tracef("got %x", id)
	if err != nil {
		return false, err
	}

	index := TagIndex(id >> 4)
	currentTag := ElementTag(id & 0xF)
	if index != expectedIndex && expectedIndex != ignoreTagIndex {
		log.Tracef("skipping index: %x at pos:%d , wants: %x%x", index, e.d.Pos(), expectedIndex, tag)
		e.Debug()
		e.lastTag = &tagInfo{
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

func (decoder *Extractor) ExtractLwwString(index TagIndex) (result Lww[string], found bool, err error) {
	if found, err = decoder.checkTag(index, ItemTag); !found {
		return
	}

	elementLength, err := decoder.d.GetUInt32()
	if err != nil {
		return
	}
	pos := decoder.d.Pos()
	endPos := pos + int(elementLength)
	log.Printf("LwwString, Int?: %d, pos:%d, max:%d", elementLength, pos, decoder.d.max)
	timestamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}

	_, err = decoder.checkTag(2, ItemTag)
	if err != nil {
		return
	}

	someInt2, err := decoder.d.GetUInt32()
	if err != nil {
		return
	}
	log.Printf("Lww someInt2??: %d ", someInt2)

	strLen, err := decoder.d.GetVarUInt32()
	if err != nil {
		return
	}

	log.Printf("got strlen %d ", strLen)
	isAscii, err := decoder.d.GetByte()
	if err != nil {
		return
	}
	// if strLen == 0 {
	// 	return
	// }
	log.Printf("got isascii %d ", isAscii)
	theStr, err := decoder.d.GetBytes(int(strLen))
	if err != nil {
		return
	}
	result.Value = string(theStr)
	result.Timestamp = timestamp
	log.Printf("got string: '%s'", theStr)
	pos = decoder.d.Pos()
	if pos > endPos {
		err = fmt.Errorf("buffer overflow, pos: %d max: %d", pos, endPos)
		return
	}

	log.Printf("LwwStringEnd, pos:%d, max:%d", pos, decoder.d.max)
	return
}

func (e *Extractor) ExtractInt(index TagIndex) (result uint32, found bool, err error) {
	if found, err = e.checkTag(index, NumberTag); !found {
		return
	}
	result, err = e.d.GetUInt32()
	return
}
func (e *Extractor) ExtractDouble(index TagIndex) (result float64, found bool, err error) {
	if found, err = e.checkTag(index, DoubleTag); !found {
		return
	}
	err = binary.Read(e.d, binary.LittleEndian, &result)
	return
}
func (e *Extractor) ExtractBool(index TagIndex) (result bool, found bool, err error) {
	if found, err = e.checkTag(index, BoolTag); !found {
		return
	}

	b, err := e.d.GetByte()
	result = b != 0
	return
}

func (e *Extractor) ExtractByte(index TagIndex) (result byte, found bool, err error) {
	if found, err = e.checkTag(index, BoolTag); !found {
		return
	}
	result, err = e.d.GetByte()
	return
}
func (e *Extractor) ExtractFloat(index TagIndex) (result float32, found bool, err error) {
	if found, err = e.checkTag(index, NumberTag); !found {
		return
	}
	err = binary.Read(e.d, binary.LittleEndian, &result)
	return
}
func (e *Extractor) ExtractCrdtId(index TagIndex) (result CrdtId, found bool, err error) {
	if found, err = e.checkTag(index, CrdtTag); !found {
		return
	}
	short, err := e.d.GetVarUInt32()
	if err != nil {
		log.Error("can't get short1")
		return
	}
	part1 := short

	short2, err := e.d.GetVarInt64()
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
func (e *Extractor) ExtractInfo(index TagIndex) (result TreeItemInfo, found bool, err error) {
	if found, err = e.checkTag(index, ItemTag); !found {
		return
	}
	nodeLength, err := e.d.GetUInt32()
	nodeEnd := nodeLength + uint32(e.d.Pos())
	if err != nil {
		return
	}
	result.ParentId, _, err = e.ExtractCrdtId(1)
	if err != nil {
		return
	}

	if found, err = e.checkTag(2, ItemTag); !found {
		return
	}
	item, err := e.d.GetUInt32()
	if err != nil {
		return
	}
	result.Value.CurrentVersion = byte(item >> 4)
	result.Value.MinVersion = byte(0xFFFF0000 & item)

	result.Bob, err = e.ExtractBobUntil(int(nodeEnd))
	return
}

func (e *Extractor) ExtractBob() (bob []byte, err error) {
	return e.ExtractBobUntil(e.d.max)
}

func (e *Extractor) ExtractBobUntil(max int) (bob []byte, err error) {
	if e.lastTag != nil {
		log.Error("pending last tag, fix this")
		err = errors.New("pending last tag")
		return

	}
	pos := e.d.position
	bobLength := max - pos
	if bobLength > 0 {
		log.Tracef("Extracting bob with length:%d (%d,%d)", bobLength, pos, max)
		bob, err = e.d.GetBytes(bobLength)
		if err != nil {
			return
		}
	}
	return
}

func (e *Extractor) ExtractLine(lineItem SceneItemBase) (err error) {
	item, ok := lineItem.(*LineItem)
	if !ok {
		return errors.New("not a LineItem")
	}

	tool, _, err := e.ExtractInt(1)
	if err != nil {
		return err
	}
	line := &item.Line.Value
	line.Tool = byte(tool)

	color, _, err := e.ExtractInt(2)
	if err != nil {
		return err
	}
	line.Color = byte(color)

	thickness, _, err := e.ExtractDouble(3)
	if err != nil {
		return err
	}
	line.ThicknessScale = thickness

	line.StartingLength, _, err = e.ExtractFloat(4)
	if err != nil {
		return err
	}
	found, err := e.checkTag(5, ItemTag)
	if err != nil {
		return
	}
	if !found {
		log.Trace("only bob")
		item.Bob, err = e.ExtractBob()
		return
	}

	var length uint32
	length, err = e.d.GetUInt32()
	if err != nil {
		return
	}
	nPoints := int(length / 0x18)
	log.Infof("Length: 0x%x, Points %d", length, nPoints)

	for i := 0; i < nPoints; i++ {
		point := &PenPoint{}
		point.X, err = e.d.GetFloat32()
		if err != nil {
			return
		}
		point.Y, err = e.d.GetFloat32()
		if err != nil {
			return
		}
		var tmp float32
		tmp, err = e.d.GetFloat32()
		if err != nil {
			return
		}
		point.Speed = int16(math.Round(float64(tmp) * 4))

		tmp, err = e.d.GetFloat32()
		if err != nil {
			return
		}
		point.Direction = byte(math.Round(float64(255 * tmp / 6.2831855)))

		tmp, err = e.d.GetFloat32()
		if err != nil {
			return
		}
		point.Width = int16(math.Round(float64(tmp) * 4))

		tmp, err = e.d.GetFloat32()
		if err != nil {
			return
		}
		point.Pressure = byte(255 * tmp)

		log.Trace(point)
		line.AddPoint(point)
	}
	return
}
func (e *Extractor) ExtractSceneItem(index TagIndex, sceneItem SceneItemBase) (found bool, err error) {
	if found, err = e.checkTag(index, ItemTag); !found {
		return
	}
	i, err := e.d.GetUInt32()
	if err != nil {
		return
	}
	log.Infof("sceneItem: %x", i)
	// sceneItem.Type, _, err = e.ExtractByte()
	if err != nil {
		return
	}

	b, err := e.d.GetByte()
	if err != nil {
		return
	}
	sceneType := SceneType(b)
	switch sceneType {
	case LineType:
		err = e.ExtractLine(sceneItem)
		if err != nil {
			return
		}
		log.Trace(sceneItem)
	case GlyphRangeType:
		log.Info("TODO: glyphs")
	}

	return
}

func (e *Extractor) ExtractUUIDPair() (u uuid.UUID, index AuthorId, err error) {
	_, err = e.checkTag(ignoreTagIndex, ItemTag)
	if err != nil {
		return
	}

	var elementLength uint32
	elementLength, err = e.d.GetUInt32()
	if err != nil {
		return
	}
	log.Info("got length: ", elementLength)
	var uuidLen uint32
	uuidLen, err = e.d.GetVarUInt32()
	if err != nil {
		return
	}
	if uuidLen != 16 {
		err = fmt.Errorf("uuid length != 16")
		return
	}
	var buffer []byte
	buffer, err = e.d.GetBytes(int(uuidLen))
	if err != nil {
		return
	}

	u, err = uuid.FromBytes(buffer)
	if err != nil {
		return
	}
	idx, err := e.d.GetShort()
	if err != nil {
		return
	}
	index = AuthorId(idx)
	return

}
func (e *Extractor) Debug() {
	pos := e.d.position
	fmt.Println(hex.EncodeToString(e.buffer))
	padding := ""
	if pos > 0 {
		padding = strings.Repeat("  ", pos-1)
	}
	fmt.Printf("%s ^  pos: %d\n", padding, pos)
}
