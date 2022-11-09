package v6

import (
	"bytes"
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
	CrdtTag ElementTag = 0xf
	Length4 ElementTag = 0xc
	Byte8   ElementTag = 8
	Byte4   ElementTag = 4
	Byte2   ElementTag = 2
	Byte1   ElementTag = 1
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

// func NewExtractor(d *BinaryDeserializer) *Extractor {
// 	return &Extractor{
// 		d: d,
// 	}
// }
func NewDebugExtractor(reader io.Reader, max int) (extractor *Extractor, err error) {
	buffer := make([]byte, max)
	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return
	}
	breader := bytes.NewReader(buffer)
	decoder := NewDecoder(breader, max)
	extractor = &Extractor{
		d:      decoder,
		buffer: buffer,
	}
	return
}

func NewExtractor(reader io.Reader, max int) (extractor *Extractor, err error) {
	decoder := NewDecoder(reader, max)
	extractor = &Extractor{
		d: decoder,
	}
	return
}
func (e *Extractor) Discard() (n int64, err error) {
	if e.d.Pos() == e.d.max {
		return
	}
	return io.Copy(io.Discard, e.d)
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
	if found, err = decoder.checkTag(control, Length4); !found {
		return
	}
	_, err = decoder.d.GetFixedUInt32()
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
	if found, err = decoder.checkTag(control, Length4); !found {
		return
	}
	_, err = decoder.d.GetFixedUInt32()
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
func (e *Extractor) ExtractLwwCrdt(control TagIndex) (result Lww[CrdtId], found bool, err error) {
	elementLength, found, err := e.ExtractUInt(control)
	if err != nil || !found {
		return
	}
	pos := e.d.Pos()
	log.Printf("ElementLength: %d, pos:%d, max:%d", elementLength, pos, e.d.max)
	val, _, err := e.ExtractCrdtId(1)
	if err != nil {
		return
	}
	timeStamp, _, err := e.ExtractCrdtId(2)
	if err != nil {
		return
	}
	result.Value = val
	result.Timestamp = timeStamp
	return
}
func (decoder *Extractor) ExtractLwwBool(control TagIndex) (result Lww[bool], found bool, err error) {
	result = Lww[bool]{}
	if found, err = decoder.checkTag(control, Length4); !found {
		return
	}

	//node size
	_, err = decoder.d.GetFixedUInt32()
	if err != nil {
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
	log.Trace("consumingTag: %x at pos: %x", id, e.d.Pos())
	if err != nil {
		return false, err
	}

	index := TagIndex(id >> 4)
	currentTag := ElementTag(id & 0xF)
	if index != expectedIndex && expectedIndex != ignoreTagIndex {
		log.Tracef("skipping index: %x at pos:%d , wants: %x%x", index, e.d.Pos(), expectedIndex, tag)
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
	if found, err = decoder.checkTag(index, Length4); !found {
		return
	}

	elementLength, err := decoder.d.GetFixedUInt32()
	if err != nil {
		return
	}
	pos := decoder.d.Pos()
	endPos := pos + int(elementLength)
	timestamp, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}

	_, _, err = decoder.ExtractUInt(2)
	if err != nil {
		return
	}

	stringLength, err := decoder.d.GetVarUInt32()
	if err != nil {
		return
	}
	log.Trace("LwwString: got strlen %d ", stringLength)

	isAscii, err := decoder.d.ReadByte()
	if err != nil {
		return
	}
	// if strLen == 0 {
	// 	return
	// }
	log.Trace("got isascii %d ", isAscii)
	strBytes, err := decoder.d.GetBytes(int(stringLength))
	if err != nil {
		return
	}
	result.Value = string(strBytes)
	result.Timestamp = timestamp
	log.Trace("got string: '%s'", strBytes)
	pos = decoder.d.Pos()
	if pos > endPos {
		err = fmt.Errorf("buffer overflow, pos: %d max: %d", pos, endPos)
		return
	}

	log.Trace("LwwStringEnd, pos:%d, max:%d", pos, decoder.d.max)
	return
}

func (e *Extractor) ExtractUInt(index TagIndex) (result uint32, found bool, err error) {
	if found, err = e.checkTag(index, Length4); !found {
		return
	}
	result, err = e.d.GetFixedUInt32()
	return
}
func (e *Extractor) ExtractInt(index TagIndex) (result int, found bool, err error) {
	if found, err = e.checkTag(index, Byte4); !found {
		return
	}
	intResult, err := e.d.GetFixedInt32()
	result = int(intResult)
	return
}
func (e *Extractor) ExtractShort(index TagIndex) (result uint16, found bool, err error) {
	if found, err = e.checkTag(index, Byte2); !found {
		return
	}
	result, err = e.d.GetShort()
	return
}
func (e *Extractor) ExtractDouble(index TagIndex) (result float64, found bool, err error) {
	if found, err = e.checkTag(index, Byte8); !found {
		return
	}
	err = binary.Read(e.d, binary.LittleEndian, &result)
	return
}
func (e *Extractor) ExtractBool(index TagIndex) (result bool, found bool, err error) {
	if found, err = e.checkTag(index, Byte1); !found {
		return
	}

	b, err := e.d.ReadByte()
	result = b != 0
	return
}

func (e *Extractor) ExtractByte(index TagIndex) (result byte, found bool, err error) {
	if found, err = e.checkTag(index, Byte1); !found {
		return
	}
	result, err = e.d.ReadByte()
	return
}
func (e *Extractor) ExtractFloat(index TagIndex) (result float32, found bool, err error) {
	if found, err = e.checkTag(index, Byte4); !found {
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
	if short > binary.MaxVarintLen16 {
		err = errors.New("part1 > short max")
		return
	}
	part1 := short

	short2, err := e.d.GetVarUInt64()
	if err != nil {
		return
	}

	if short2&0xFFFF0000 != 0 {
		log.Warnf("short1: %x, short2 > fmask true, %x", short, short2)
	}
	part2 := uint64(short2)
	result = CrdtId(uint64(part1)<<48 | part2)
	return
}
func (e *Extractor) ExtractInfo(index TagIndex) (result TreeItemInfo, found bool, err error) {
	nodeLength, found, err := e.ExtractUInt(index)
	if err != nil || !found {
		return
	}
	nodeEnd := nodeLength + uint32(e.d.Pos())
	if err != nil {
		return
	}
	result.ParentId, _, err = e.ExtractCrdtId(1)
	if err != nil {
		return
	}

	_, found, err = e.ExtractUInt(2)
	if err != nil || !found {
		return
	}
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
		log.Warnf("Extracting bob with length:%d (%d,%d)", bobLength, pos, max)
		bob, err = e.d.GetBytes(bobLength)
		if err != nil {
			return
		}
	}
	return
}
func (e *Extractor) ExtractPoint() (point *PenPoint, err error) {
	point = &PenPoint{}

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
	point.Direction = byte(math.Round(float64(255 * tmp / (math.Pi * 2))))

	tmp, err = e.d.GetFloat32()
	if err != nil {
		return
	}
	point.Width = int16(math.Round(float64(tmp) * 4))

	tmp, err = e.d.GetFloat32()
	if err != nil {
		return
	}
	tmpInt := int(math.Round(float64(tmp * 255)))
	point.Pressure = byte(tmpInt)
	return
}

func (e *Extractor) ExtractLine() (item *LineItem, err error) {
	item = &LineItem{
		SceneItem: SceneItem{
			Type: LineType,
		},
	}

	tool, _, err := e.ExtractInt(1)
	if err != nil {
		return
	}
	line := &item.Line.Value
	line.Tool = byte(tool)

	color, _, err := e.ExtractInt(2)
	if err != nil {
		return
	}
	line.Color = byte(color)

	line.ThicknessScale, _, err = e.ExtractDouble(3)
	if err != nil {
		return
	}

	line.StartingLength, _, err = e.ExtractFloat(4)
	if err != nil {
		return
	}
	length, found, err := e.ExtractUInt(5)
	if err != nil {
		return
	}
	if !found {
		item.Bob, err = e.ExtractBob()
		return
	}

	nPoints := int(length / PenPointSize)
	if length%PenPointSize != 0 {
		log.Error("point size mismatch")
	}

	var point *PenPoint
	for i := 0; i < nPoints; i++ {
		point, err = e.ExtractPoint()
		if err != nil {
			return
		}

		log.Trace(point)
		line.AddPoint(point)
	}
	item.Line.Timestamp, _, err = e.ExtractCrdtId(6)

	return
}
func (e *Extractor) ExtractSceneItem(index TagIndex) (sceneItem SceneBaseItem, err error) {
	length, found, err := e.ExtractUInt(index)
	if err != nil || !found {
		return
	}
	elementEnd := int(length) + e.d.Pos()
	log.Infof("sceneItem length: %d", length)
	if err != nil {
		return
	}

	sct, err := e.d.ReadByte()
	if err != nil {
		return
	}
	sceneType := SceneType(sct)
	switch sceneType {
	case GroupType:
		sceneItem = new(GroupItem)
	case LineType:
		sceneItem, err = e.ExtractLine()
	case GlyphRangeType:
		sceneItem = new(GlyphRange)
	case TextType:
		sceneItem = new(SceneTextItem)
	default:
		sceneItem = &SceneItem{
			Type: sceneType,
		}
	}

	if err != nil {
		return
	}
	sceneItem.Item().Id, _, err = e.ExtractCrdtId(2)
	if err != nil {
		return
	}
	sceneItem.Item().Bob, err = e.ExtractBobUntil(elementEnd)
	if err != nil {
		return
	}

	log.Trace(sceneItem)

	return
}

func (e *Extractor) ExtractUUIDPair() (u uuid.UUID, index AuthorId, err error) {
	_, err = e.checkTag(ignoreTagIndex, Length4)
	if err != nil {
		return
	}

	elementLength, err := e.d.GetFixedUInt32()
	if err != nil {
		return
	}
	log.Trace("extractuuidpair: got elementlength: ", elementLength)

	uuidLen, err := e.d.GetVarUInt32()
	if err != nil {
		return
	}
	if uuidLen != 16 {
		err = fmt.Errorf("extractuuid: uuid length != 16")
		return
	}

	buffer, err := e.d.GetBytes(int(uuidLen))
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
	fmt.Println(strings.ToUpper(hex.EncodeToString(e.buffer)))
	padding := ""
	if pos > 0 {
		padding = strings.Repeat("  ", pos-1)
	}
	fmt.Printf("%s ^  pos: %d\n", padding, pos)
}

func (e *Extractor) ExtractTextItem() (textItem Item[TextItem], err error) {
	length, _, err := e.ExtractUInt(0)
	if err != nil {
		return
	}
	// for {
	textItem.Id, _, err = e.ExtractCrdtId(2)
	if err != nil {
		return
	}
	textItem.Left, _, err = e.ExtractCrdtId(3)
	if err != nil {
		return
	}
	textItem.Right, _, err = e.ExtractCrdtId(4)
	if err != nil {
		return
	}
	textItem.DeletedLength, _, err = e.ExtractInt(5)
	if err != nil {
		return
	}

	var found bool
	elementLength2, found, err := e.ExtractUInt(6)
	if err != nil {
		return
	}
	log.Trace(found, length, elementLength2)

	var strLength uint32
	strLength, err = e.d.GetVarUInt32()
	if err != nil {
		return
	}
	_, err = e.d.ReadByte()
	if err != nil {
		return
	}
	var strBytes []byte
	strBytes, err = e.d.GetBytes(int(strLength))
	if err != nil {
		return
	}
	textItem.Value.Text = string(strBytes)
	log.Debug(textItem.Value.Text)

	textItem.Value.Format, _, err = e.ExtractUInt(2)
	if err != nil {
		return
	}
	return
}
