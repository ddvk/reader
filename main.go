package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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

type Index uint16
type UUIDMap struct {
	UUID2Index map[uuid.UUID]Index
	Index2UUID map[Index]uuid.UUID
}

func NewMap() UUIDMap {
	return UUIDMap{
		UUID2Index: make(map[uuid.UUID]Index),
		Index2UUID: make(map[Index]uuid.UUID),
	}
}

func (um *UUIDMap) Entries() int {
	return len(um.Index2UUID)
}

func (um *UUIDMap) Add(u uuid.UUID, index Index) {
	if um.Index2UUID == nil {
		um.Index2UUID = make(map[Index]uuid.UUID)
	}
	if um.UUID2Index == nil {
		um.UUID2Index = make(map[uuid.UUID]Index)
	}
	um.Index2UUID[index] = u
	um.UUID2Index[u] = index
}

func parserUUID(decoder *ChunkDecoder) (uuidMap UUIDMap, err error) {
	count, err := decoder.GetVarUInt32()
	if err != nil {
		return
	}
	log.Info("count uuids: ", count)
	var elementLength uint32
	for i := 0; i < int(count); i++ {
		_, err = decoder.checkTag(-1, ItemTag)
		if err != nil {
			return
		}

		elementLength, err = decoder.GetUInt32()
		if err != nil {
			return
		}
		log.Info("got length: ", elementLength)
		var uuidLen uint32
		uuidLen, err = decoder.GetVarUInt32()
		if err != nil {
			return
		}
		if uuidLen != 16 {
			err = fmt.Errorf("uuid length != 16")
			return
		}
		var buffer []byte
		buffer, err = decoder.GetBytes(int(uuidLen))
		if err != nil {
			return
		}

		var u uuid.UUID
		u, err = uuid.FromBytes(buffer)
		if err != nil {
			return
		}
		var index uint16
		index, err = decoder.GetShort()
		if err != nil {
			return
		}
		uuidMap.Add(u, Index(index))
	}
	log.Info("Got a map: ", uuidMap.Entries())
	return
}

var ErrTagMismatch = errors.New("tag mismatch")
var ErrIndexMismatch = errors.New("index mismatch")

type MigrationInfo struct {
	MigrationId CrdtId
	IsDevice    bool
	Bob         []byte
}

func parseMigrationInfo(decoder *ChunkDecoder) (migrationInfo MigrationInfo, err error) {
	//migrationId?
	///device?
	migrationInfo.MigrationId, _, err = decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}
	migrationInfo.IsDevice, _, err = decoder.ExtractBool(2)
	if err != nil {
		return
	}
	migrationInfo.Bob, err = decoder.ExtractBob()
	if err != nil {
		return
	}
	return
}

type CrdtId uint64

type TreeNodeInfo struct {
	curVersion uint8
	minVersion uint8
}

func readCrdt() {

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

func readPageInfo(decoder *ChunkDecoder) (pageInfo PageInfo, err error) {
	pageInfo.Loads, _, err = decoder.ExtractInt(1)
	if err != nil {
		return
	}
	pageInfo.Merges, _, err = decoder.ExtractInt(2)
	if err != nil {
		return
	}
	pageInfo.TextChars, _, err = decoder.ExtractInt(3)
	if err != nil {
		return
	}
	pageInfo.TextLinex, _, err = decoder.ExtractInt(4)
	if err != nil {
		return
	}
	//todo: load the rest in bob

	log.Info("Loads: ", pageInfo)
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
	go func() {

	}()
	prefixed := &prefixed.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	}
	log.SetFormatter(prefixed)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	err := _main()
	if err != nil {
		log.Fatal(err)
	}
}
