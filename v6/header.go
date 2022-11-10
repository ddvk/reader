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
	return fmt.Sprintf("Tag: %v, length: %d, MinVer:%d, CurVer:%d", h.Info.PayloadType, h.Size, h.Info.NodeInfo.MinVersion, h.Info.NodeInfo.CurrentVersion)
}

type HeaderInfo struct {
	PayloadType TagType
	NodeInfo    Info
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
		NodeInfo: Info{
			CurrentVersion: buffer[2],
			MinVersion:     buffer[1],
		},
	}

	h = Header{
		Size: size,
		Info: header,
	}
	return
}
