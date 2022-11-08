package v6

import (
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var ErrTagMismatch = errors.New("tag mismatch")
var ErrIndexMismatch = errors.New("index mismatch")

type SceneReader struct {
	r     io.Reader
	e     *Extractor
	scene *Scene
}

func (s *SceneReader) ExtractScene(r io.Reader) (scene Scene, err error) {
	s.r = r
	s.scene = &scene

	const headerLength = 8
	var pos int32
	var header Header

	for {
		fmt.Println("next chunk")
		header, err = ReadHeader(r)
		if err == io.EOF {
			err = nil
			return
		}
		if err != nil {
			return
		}
		log.Infof("%v, position:\t0x%-x", header, pos)

		err = s.parseHeader(header, r)
		if err != nil {
			return
		}
		pos = pos + headerLength + header.Size
		fmt.Println("---")
	}
}

func (s *SceneReader) parseHeader(header Header, reader io.Reader) (err error) {
	headerInfo := header.Info
	nodeType := headerInfo.PayloadType
	max := int(header.Size)

	s.e, err = NewDebugExtractor(reader, max)
	if err != nil {
		return
	}

	var moveNode TreeMoveInfo
	var sceneNode SceneTreeNode
	switch nodeType {
	case 0:
		s.scene.MigrationInfo, err = s.parseMigrationInfo()
	case UUIDIdexTag:
		s.scene.UUIDMap, err = s.parserUUID()
	case PageInfoTag:
		s.scene.PageInfo, err = s.readPageInfo()
	case SceneTreeTag:
		moveNode, err = s.treeMoveNode()
		log.Debug(moveNode)
	case SceneTreeNodeTag:
		sceneNode, err = s.readSceneNode()
		log.Debug(sceneNode)
		//sceneitem
	case GlyphItemTag, GroupItemTag, LineItemTag, 8:
		var item Item[SceneBaseItem]
		var id CrdtId
		item, id, err = s.readSceneItem(nodeType)
		log.Debug(id, item)
	case RootTextTag:
		var node SceneTextItem
		node, err = s.readRootText(nodeType)
		log.Debug(node)

	default:
		log.Error("unhandled type:", nodeType)
	}

	if err != nil {
		return
	}
	s.e.Debug()

	n, err := s.e.Discard()
	if n > 0 {
		log.Warnf("Discarding unhandled: %d bytes", n)
	}

	return err
}

// func createSceneItem(nodeType TagType) SceneItemBase {
// 	switch nodeType {
// 	case LineItemTag:
// 		return &LineItem{}
// 	case GlyphItemTag:
// 		return &GlyphRange{}
// 	case TextItemTag:
// 		return &SceneTextItem{}
// 	case GroupItemTag:
// 		return &GroupItem{}
// 	}
// 	return &SceneItem{}
// }

func (s *SceneReader) extractMap() (err error) {
	textItem := new(Item[TextItem])
	elementLength, err := s.e.d.GetVarUInt32()
	if err != nil {
		return
	}
	log.Trace(elementLength)
	length, _, err := s.e.ExtractUInt(0)
	if err != nil {
		return
	}
	// for {
	textItem.Id, _, err = s.e.ExtractCrdtId(2)
	if err != nil {
		return
	}
	textItem.Left, _, err = s.e.ExtractCrdtId(3)
	if err != nil {
		return
	}
	textItem.Right, _, err = s.e.ExtractCrdtId(4)
	if err != nil {
		return
	}
	textItem.DeletedLength, _, err = s.e.ExtractInt(5)
	if err != nil {
		return
	}

	var found bool
	elementLength2, found, err := s.e.ExtractUInt(6)
	if err != nil {
		return
	}
	log.Trace(found, length, elementLength2)
	// if found {
	// 	break
	// }
	// }
	var strLength uint32
	strLength, err = s.e.d.GetVarUInt32()
	if err != nil {
		return
	}
	_, err = s.e.d.ReadByte()
	if err != nil {
		return
	}
	var b []byte
	b, err = s.e.d.GetBytes(int(strLength))
	if err != nil {
		return
	}
	log.Trace("Pos: %x", s.e.d.Pos())
	log.Debug(string(b))
	return
}

func (s *SceneReader) readRootText(nodeType TagType) (sceneItem SceneTextItem, err error) {
	sceneItem.Item().ParentId, _, err = s.e.ExtractCrdtId(1)
	if err != nil {
		return
	}
	length, _, err := s.e.ExtractUInt(2)
	if err != nil {
		return
	}
	log.Trace("elemtn length: ", length)
	length2, _, err := s.e.ExtractUInt(1)
	if err != nil {
		return
	}
	log.Trace(length2)
	length3, _, err := s.e.ExtractUInt(1)
	if err != nil {
		return
	}
	maxLength := length3 + uint32(s.e.d.Pos())
	log.Trace(maxLength)

	//sceneItem.Sequence.Add(&textItem.Value)
	//loop

	//todo: vector
	sceneItem.Sequence.Bob, err = s.e.ExtractBobUntil(int(maxLength))
	if err != nil {
		return
	}
	length4, _, err := s.e.ExtractUInt(2)
	if err != nil {
		return
	}
	log.Trace(length4)
	mapLength, _, err := s.e.ExtractUInt(1)
	if err != nil {
		return
	}
	mapEnd := mapLength + uint32(s.e.d.Pos())
	log.Trace(mapLength)
	//TODO:map
	_, err = s.e.ExtractBobUntil(int(mapEnd))
	if err != nil {
		return
	}

	//length of next
	_, _, err = s.e.ExtractUInt(3)

	if err != nil {
		return
	}

	sceneItem.Position.X, err = s.e.d.GetFloat64()
	if err != nil {
		return
	}
	sceneItem.Position.Y, err = s.e.d.GetFloat64()
	if err != nil {
		return
	}

	width, _, err := s.e.ExtractFloat(4)
	log.Trace(width)

	return
}
func (s *SceneReader) readSceneItem(nodeType TagType) (item Item[SceneBaseItem], id CrdtId, err error) {
	id, _, err = s.e.ExtractCrdtId(1)
	if err != nil {
		return
	}
	item.Id, _, err = s.e.ExtractCrdtId(2)
	if err != nil {
		return
	}
	item.Left, _, err = s.e.ExtractCrdtId(3)
	if err != nil {
		return
	}
	item.Right, _, err = s.e.ExtractCrdtId(4)
	if err != nil {
		return
	}
	item.DeletedLength, _, err = s.e.ExtractInt(5)
	if err != nil {
		return
	}
	item.Value, err = s.e.ExtractSceneItem(6)
	if err != nil {
		return
	}
	if err != nil {
		return
	}

	item.Bob, err = s.e.ExtractBob()
	return
}

// func sw(x int) int {
// 	if x == 4 {
// 		return 2
// 	}
// 	if x > 9 {
// 		return 0
// 	}
// 	return 1
// }

func (s *SceneReader) parserUUID() (uuidMap UUIDMap, err error) {
	count, err := s.e.d.GetVarUInt32()
	if err != nil {
		return
	}
	log.Debug("got uuids: ", count)
	for i := 0; i < int(count); i++ {
		var index AuthorId
		var u uuid.UUID
		u, index, err = s.e.ExtractUUIDPair()
		if err != nil {
			return
		}

		uuidMap.Add(u, AuthorId(index))
	}
	return
}

func (s *SceneReader) parseMigrationInfo() (migrationInfo MigrationInfo, err error) {
	migrationInfo.MigrationId, _, err = s.e.ExtractCrdtId(1)
	if err != nil {
		return
	}
	migrationInfo.IsDevice, _, err = s.e.ExtractBool(2)
	if err != nil {
		return
	}
	migrationInfo.Bob, err = s.e.ExtractBob()
	if err != nil {
		return
	}
	return
}

func (s *SceneReader) readPageInfo() (pageInfo PageInfo, err error) {
	pageInfo.Loads, _, err = s.e.ExtractInt(1)
	if err != nil {
		return
	}
	pageInfo.Merges, _, err = s.e.ExtractInt(2)
	if err != nil {
		return
	}
	pageInfo.TextChars, _, err = s.e.ExtractInt(3)
	if err != nil {
		return
	}
	pageInfo.TextLinex, _, err = s.e.ExtractInt(4)
	if err != nil {
		return
	}
	pageInfo.Bob, err = s.e.ExtractBob()
	if err != nil {
		return
	}
	log.Info("Loads: ", pageInfo)
	return
}
func (s *SceneReader) readSceneNode() (node SceneTreeNode, err error) {
	node.Id, _, err = s.e.ExtractCrdtId(1)
	if err != nil {
		return
	}
	node.Name, _, err = s.e.ExtractLwwString(2)
	if err != nil {
		return
	}

	node.Visible, _, err = s.e.ExtractLwwBool(3)
	if err != nil {
		return
	}

	selectedId, hasAnchor, err := s.e.ExtractCrdtId(4)
	if err != nil {
		return
	}
	if hasAnchor {
		log.Trace("has anchor")
		node.AnchorId.Value = selectedId

		node.AnchorMode.Value, _, err = s.e.ExtractByte(5)
		if err != nil {
			return
		}

		node.AnchorThreshold.Value, _, err = s.e.ExtractFloat(6)
		if err != nil {
			return
		}

	} else {
		node.AnchorId, _, err = s.e.ExtractLwwCrdt(7)
		if err != nil {
			return
		}
		node.AnchorMode, _, err = s.e.ExtractLwwByte(8)
		if err != nil {
			return
		}

		node.AnchorThreshold, _, err = s.e.ExtractLwwFloat(9)
		if err != nil {
			return
		}

		node.AnchorInitialOriginX, _, err = s.e.ExtractLwwFloat(10)
		if err != nil {
			return
		}
	}
	node.Bob, err = s.e.ExtractBob()
	if err != nil {
		log.Warn("can't get bob", err)
		return
	}

	return
}

func (s *SceneReader) treeMoveNode() (node TreeMoveInfo, err error) {
	node.Id, _, err = s.e.ExtractCrdtId(1)
	if err != nil {
		return
	}
	node.NodeId, _, err = s.e.ExtractCrdtId(2)
	if err != nil {
		return
	}
	node.IsUpdate, _, err = s.e.ExtractBool(3)
	if err != nil {
		return
	}
	node.Info, _, err = s.e.ExtractInfo(4)
	if err != nil {
		return
	}
	node.Bob, err = s.e.ExtractBob()
	if err != nil {
		return
	}
	return
}
