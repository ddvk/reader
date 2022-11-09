package v6

import (
	"errors"
	"io"

	log "github.com/sirupsen/logrus"
)

var ErrTagMismatch = errors.New("tag mismatch")
var ErrIndexMismatch = errors.New("index mismatch")

type SceneReader struct {
	r     io.Reader
	scene *Scene
	tree  *SceneTree
}

func (s *SceneReader) ExtractScene(r io.Reader) (scene Scene, err error) {
	s.r = r
	s.scene = &scene
	s.tree = NewTree()

	const headerLength = 8
	var pos int32
	var header Header

	for {
		header, err = ReadHeader(r)
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
		log.Infof("%v, position:\t0x%-x", header, pos)

		err = s.parsePayload(header, r)
		if err != nil {
			return
		}
		pos = pos + headerLength + header.Size
	}

	//todo:
	scene.Layers = s.tree.Layers
	return
}

func (s *SceneReader) parsePayload(header Header, reader io.Reader) (err error) {
	headerInfo := header.Info
	nodeType := headerInfo.PayloadType
	max := int(header.Size)

	e, err := NewDebugExtractor(reader, max)
	if err != nil {
		return
	}

	var curVersion = headerInfo.CurVersion
	var minVersion = headerInfo.MinVersion

	var moveNode TreeMoveInfo
	var sceneNode SceneTreeNode
	switch nodeType {
	case 0:
		s.scene.MigrationInfo, err = e.ParseMigrationInfo()
	case UUIDIdexTag:
		s.scene.UUIDMap, err = e.ParserUUID()
	case PageInfoTag:
		s.scene.PageInfo, err = e.ReadPageInfo()
	case SceneTreeTag:
		moveNode, err = e.TreeNode()
		moveNode.ItemInfo.Value.CurrentVersion = curVersion
		moveNode.ItemInfo.Value.MinVersion = minVersion
		if err == nil {
			s.tree.AddTree(&moveNode)
		}
		log.Debug(moveNode)
	case SceneTreeNodeTag:
		sceneNode, err = e.ReadSceneNode()
		sceneNode.Info.CurrentVersion = curVersion
		sceneNode.Info.CurrentVersion = minVersion
		if err == nil {
			s.tree.AddNode(&sceneNode)
		}
		log.Debug(sceneNode)
		//sceneitem
	case GlyphItemTag, GroupItemTag, LineItemTag, 8:
		var item Item[SceneBaseItem]
		var parentId CrdtId
		item, parentId, err = e.ReadSceneItem(nodeType)
		if err == nil {
			s.tree.AddItem(item, parentId)
		}
		log.Debug(parentId, item)
	case RootTextTag:
		var node SceneTextItem
		node, err = e.ReadRootText(nodeType)
		node.CurVersion = curVersion
		node.MinVersion = minVersion
		if err == nil {
			s.tree.AddRootText(&node)
		}
		log.Debug(node)

	default:
		log.Error("unhandled type:", nodeType)
	}

	if err != nil {
		return
	}
	e.Debug()

	n, err := e.Discard()
	if n > 0 {
		log.Warnf("Discarding unhandled: %d bytes", n)
	}

	return err
}
