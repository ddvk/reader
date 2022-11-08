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
	return fmt.Sprintf("Tag: %v, length: %d, MinVer:%d, CurVer:%d", h.Info.PayloadType, h.Size, h.Info.MinVersion, h.Info.CurVersion)
}

type HeaderInfo struct {
	PayloadType TagType
	TreeNodeInfo
}

func ReadHeader(reader io.Reader) (h Header, err error) {
	var size int32
	err = binary.Read(reader, binary.LittleEndian, &size)
	if err != nil {
		return
	}
	buffer := make([]byte, 4)
	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return
	}

	header := HeaderInfo{
		PayloadType: TagType(buffer[3]),
		TreeNodeInfo: TreeNodeInfo{
			MinVersion: buffer[2],
			CurVersion: buffer[1],
		},
	}

	h = Header{
		Size: size,
		Info: header,
	}
	return
}
