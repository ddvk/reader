package v6

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Header struct {
	Size int32
	Info HeaderInfo
}

func (h Header) String() string {
	return fmt.Sprintf("Tag: %v, length: %d", h.Info.ChunkType, h.Size)
}

type HeaderInfo struct {
	ChunkType TagType
	B1        byte //middle
	B2        byte //last
}

func ReadHeader(reader io.Reader) (ch Header, err error) {
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
	header := HeaderInfo{
		ChunkType: TagType(buffer[3]),
		B1:        buffer[2],
		B2:        buffer[1],
	}
	if err != nil {
		return
	}

	ch = Header{
		Size: size,
		Info: header,
	}
	return
}
