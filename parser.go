package main

import (
	"bytes"
	"encoding/hex"
	"io"

	log "github.com/sirupsen/logrus"
)

func parse(file io.ReadSeekCloser) (err error) {
	headerLength := 0x2b
	buffer := make([]byte, headerLength)
	_, err = io.ReadFull(file, buffer)
	if err != nil {
		return err
	}
	var scene Scene
	for {
		pos64, err1 := file.Seek(0, io.SeekCurrent)
		if err1 != nil {
			return err1
		}
		pos := int(pos64)
		chunk, err := ExtractChunk(file, pos)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Infof("Tag: %x,%d,%d length: %d (%x), position:\t0x%-x", chunk.Header.TagID, chunk.Header.B1, chunk.Header.B2, chunk.Size, chunk.Size, pos)

		err = scene.interpret(chunk, file)
		if err != nil {
			return err
		}
	}
}

type Scene struct {
	Mi      MigrationInfo
	Pi      PageInfo
	UUIDMap UUIDMap
}

//readscene: crdt control = 2
func (s *Scene) interpret(chunk Chunk, reader io.Reader) (err error) {
	header := chunk.Header
	tag := header.TagID
	max := int(chunk.Size)

	buffer := make([]byte, chunk.Size)
	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return err
	}
	breader := bytes.NewReader(buffer)
	decoder := NewDecoder(breader, max)

	switch tag {
	case 0:
		s.Mi, err = parseMigrationInfo(decoder)
	case 9:
		s.UUIDMap, err = parserUUID(decoder)
	case 10:
		s.Pi, err = readPageInfo(decoder)
	case 2:
		log.Printf("tag: %d,b1: %d,b2: %d", tag, header.B1, header.B2)
		log.Printf("hex: %s", hex.EncodeToString(buffer))
		var node SceneNode
		node, err = readSceneNode(decoder)
		log.Print(node)

	case 3, 4, 5, 8:

	}
	if err != nil {
		DebugBuffer(buffer, decoder.Pos(), max)
	}

	return err
}
