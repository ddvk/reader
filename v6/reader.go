package v6

import (
	"bytes"
	"encoding/hex"
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

func (s *SceneReader) ExtractScene(file io.Reader) (scene Scene, err error) {
	s.r = file
	s.scene = &scene

	var pos int32
	var chunk Header
	chunk, err = ReadHeader(file)
	if err != nil {
		return
	}
	if chunk.Info.ChunkType != UUIDIdexTag {
		err = errors.New("first chunk not UUID map")
		return
	}

	err = s.interpret(chunk, file)
	if err != nil {
		return
	}
	//todo:fix
	pos = pos + 8 + chunk.Size

	for {
		fmt.Println("next chunk")
		chunk, err = ReadHeader(file)
		if err == io.EOF {
			err = nil
			return
		}
		if err != nil {
			return
		}
		log.Infof("%v, position:\t0x%-x", chunk, pos)

		err = s.interpret(chunk, file)
		if err != nil {
			return
		}
		pos = pos + 8 + chunk.Size
		fmt.Println("---")
	}
}

func (sp *SceneReader) interpret(chunk Header, reader io.Reader) (err error) {
	header := chunk.Info
	chunkType := header.ChunkType
	max := int(chunk.Size)

	buffer := make([]byte, chunk.Size)
	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return err
	}
	breader := bytes.NewReader(buffer)

	decoder := NewDecoder(breader, max)
	sp.e = NewExtractor(decoder)
	sp.e.buffer = buffer

	switch chunkType {
	case 0:
		sp.scene.Mi, err = sp.parseMigrationInfo()
	case UUIDIdexTag:
		sp.scene.UUIDMap, err = sp.parserUUID()
	case PageInfoTag:
		sp.scene.Pi, err = sp.readPageInfo()
	case SceneTreeNodeTag:
		log.Printf("tag: %d,b1: %d,b2: %d", chunkType, header.B1, header.B2)
		log.Printf("hex: %s", hex.EncodeToString(buffer))
		var node SceneTreeNode
		node, err = sp.readSceneNode()
		log.Print(node)

	case SceneTreeNodeMoveTag:
		var node TreeMoveInfo
		node, err = sp.treeMoveNode()
		fmt.Println(node)
		//sceneitem
	case GlyphItemTag, GroupItemTag, LineItemTag, 8:
		var node SceneTreeNode
		node, err = sp.readSceneItem(chunkType)
		log.Infof("Node: (%v)", node)

	}
	DebugBuffer(buffer, decoder.Pos(), max)
	return err
}

func createSceneItem(nodeType TagType) SceneItemBase {
	switch nodeType {
	case LineItemTag:
		return &LineItem{}
	case GlyphItemTag:
		return &GlyphRange{}
	case TextItemTag:
		return &SceneTextItem{}
	}
	return &SceneItem{}

}
func (s *SceneReader) readSceneItem(nodeType TagType) (sceneNode SceneTreeNode, err error) {
	sceneNode.Id, _, err = s.e.ExtractCrdtId(1)
	if err != nil {
		return
	}
	sceneItem := createSceneItem(nodeType)
	if sceneItem == nil {
		err = fmt.Errorf("unkonwn scene for: %d", nodeType)
		return
	}

	id1, _, err := s.e.ExtractCrdtId(2)
	if err != nil {
		return
	}
	id2, _, err := s.e.ExtractCrdtId(3)
	if err != nil {
		return
	}
	id3, _, err := s.e.ExtractCrdtId(4)
	if err != nil {
		return
	}
	ui, _, err := s.e.ExtractInt(5)
	if err != nil {
		return
	}
	log.Infof("%v %v %v, int:%d", id1, id2, id3, ui)
	_, err = s.e.ExtractSceneItem(6, sceneItem)
	if err != nil {
		return
	}
	sceneNode.Sequence.Add(sceneItem)
	sceneNode.Bob, err = s.e.ExtractBob()
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
	log.Info("count uuids: ", count)
	for i := 0; i < int(count); i++ {
		var index AuthorId
		var u uuid.UUID
		u, index, err = s.e.ExtractUUIDPair()
		if err != nil {
			return
		}

		uuidMap.Add(u, AuthorId(index))
	}
	log.Info("Got a map: ", uuidMap.Entries())
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
	//todo: load the rest in bob

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
