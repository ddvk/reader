package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Header struct {
	TagID byte
	B1    byte //middle
	B2    byte //last
}

func sw(x int) int {
	if x == 4 {
		return 2
	}
	if x > 9 {
		return 0
	}
	return 1
}

func parserUUID(buffer []byte) error {
	logrus.Println("UUID map")
	log.Println(hex.EncodeToString(buffer))
	decoder := NewDecoder(bytes.NewReader(buffer))
	count, err := decoder.GetByte()
	if err != nil {
		return err
	}
	logrus.Println("count uuids: ", count)
	uva, err := decoder.GetBytes(5)
	if err != nil {
		return err
	}
	logrus.Println("unknown: ", hex.EncodeToString(uva))
	uuidLen, err := decoder.GetByte()
	if err != nil {
		return err
	}
	uuidBytes, err := decoder.GetBytes(int(uuidLen))
	if err != nil {
		return err
	}
	myUuid, err := uuid.FromBytes(uuidBytes)
	if err != nil {
		return err
	}
	id, err := decoder.GetShort()
	if err != nil {
		return err
	}
	log.Println("uid: ", myUuid.String(), " index: ", id)
	return nil
}

var ErrTagMismatch = errors.New("tag mismatch")
var ErrIndexMismatch = errors.New("index mismatch")

func parseMigrationInfo(buffer []byte) error {
	decoder := NewDecoder(bytes.NewReader(buffer))
	//migrationId?
	///device?
	crdtId, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return err
	}
	fmt.Printf(">> Tree CrdtId id: %x \n", crdtId)
	return nil
}

type CrdtId uint64

type TreeNodeInfo struct {
	curVersion uint8
	minVersion uint8
}

func readCrdt() {

}

func readSceneNode(buffer []byte) error {
	decoder := NewDecoder(bytes.NewReader(buffer))
	crdt, _, err := decoder.ExtractCrdtId(1)
	if err != nil {
		return err
	}
	lww, _, err := decoder.ExtractLwwString(2)
	if err != nil {
		return err
	}

	lwb, _, err := decoder.ExtractLwwBool(3)
	if err != nil {
		return err
	}
	log.Print("bool")

	crdt2, found, err := decoder.ExtractCrdtId(4)
	if err != nil {
		//read 5,6
		//char
		//float
		return fmt.Errorf("no crdtid %v", err)
	}
	if found {

	} else {
		var lwc LwwCrdt
		var lwb LwwByte
		var lwfx LwwFloat
		var lwfy LwwFloat
		lwc, _, err = decoder.ExtractLwwCrdt(7)
		if err != nil {
			return fmt.Errorf("no lww, %v", err)
		}
		log.Print(lwc)
		lwb, _, err = decoder.ExtractLwwByte(8)
		if err != nil {
			return err
		}
		log.Print(lwb)

		lwfx, _, err = decoder.ExtractLwwFloat(9)
		if err != nil {
			return err
		}

		lwfy, _, err = decoder.ExtractLwwFloat(10)
		if err != nil {
			return err
		}
		logrus.Print("float x:", lwfx.Value, " y:", lwfy.Value)

	}
	log.Print("Crd2:", crdt2, lww, lwb, crdt)

	//crdtid
	//crdtid
	//bool
	//TreeNode::Info
	return nil
}

type PageInfo struct {
	Loads     uint32
	Merges    uint32
	TextChars uint32
	TextLinex uint32
	Bob       uint
}

func (p PageInfo) String() string {
	return fmt.Sprintf("L: %d, M:%d Tc: %d, TL:%d", p.Loads, p.Merges, p.TextChars, p.TextLinex)

}

func parsePageStuff(buffer []byte) (err error) {
	decoder := NewDecoder(bytes.NewReader(buffer))
	pageInfo := PageInfo{}
	pageInfo.Loads, _, err = decoder.ExtractInt(1)
	if err != nil {
		return err
	}
	pageInfo.Merges, _, err = decoder.ExtractInt(2)
	if err != nil {
		return err
	}
	pageInfo.TextChars, _, err = decoder.ExtractInt(3)
	if err != nil {
		return err
	}
	pageInfo.TextLinex, _, err = decoder.ExtractInt(4)
	if err != nil {
		return err
	}
	//todo: load the reset in bob

	log.Println("Loads: ", pageInfo)
	return nil
}

//readscene: crdt control = 2
func interpret(header Header, buffer []byte) error {
	tag := header.TagID

	switch tag {
	case 0:
		return parseMigrationInfo(buffer)
	case 9:
		return parserUUID(buffer)
	case 10:
		return parsePageStuff(buffer)
	case 2:
		log.Printf("tag: %d\tb1: %d\tb2: %d\n", tag, header.B1, header.B2)
		log.Printf("> %s\n", hex.EncodeToString(buffer))
		return readSceneNode(buffer)

	}
	return nil
}

type Chunk struct {
	Size   int32
	Header Header
}

func ExtractChunk(reader io.Reader, index int) (*Chunk, error) {
	var size int32
	err := binary.Read(reader, binary.LittleEndian, &size)
	log.Printf("- length: %d (0x%x), pos: 0x%x\n", size, size, index)
	if err != nil {
		return nil, err
	}
	index += 4
	buffer := make([]byte, 4)
	// var tag uint32
	_, err = io.ReadFull(reader, buffer[:4])
	if err != nil {
		return nil, err
	}

	// b1 := tag >> 0x10 & 0xF
	// b2 := tag >> 0x8 & 0xF
	header := Header{
		TagID: buffer[3],
		B1:    buffer[2],
		B2:    buffer[1],
	}
	if err != nil {
		return nil, err
	}

	result := &Chunk{
		Size:   size,
		Header: header,
	}
	return result, nil
}

func parse(file io.ReadSeekCloser) (err error) {
	headerLength := 0x2b
	buffer := make([]byte, headerLength)
	_, err = io.ReadFull(file, buffer)
	if err != nil {
		return err
	}
	pos64, err1 := file.Seek(0, io.SeekCurrent)
	if err1 != nil {
		return err1
	}
	pos := int(pos64)
	var chunk *Chunk
	for {
		chunk, err = ExtractChunk(file, pos)
		if err != nil {
			return err
		}

		buffer = make([]byte, chunk.Size)
		_, err = io.ReadFull(file, buffer)
		if err != nil {
			return
		}
		err = interpret(chunk.Header, buffer)
		if err != nil {
			return err
		}
	}

	return
}

func _main() error {
	if len(os.Args) < 2 {
		log.Print("missing file")
		return nil
	}
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return parse(file)
}

func main() {
	prefixed := &prefixed.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	}
	logrus.SetFormatter(prefixed)
	err := _main()
	if err != nil {
		log.Fatal(err)
	}
}
