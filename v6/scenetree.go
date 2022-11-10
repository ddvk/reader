package v6

import (
	"github.com/sirupsen/logrus"
)

type SceneTree struct {
	NextItemId CrdtId
	tree       Tree[TreeNodeInfo]

	NodeMap map[CrdtId]*Node
	Root    *Node
	Layers  []*Layer
}
type Node struct {
	Id       CrdtId
	Children []*Node
	IsLayer  bool
	Layer    int
	Value    *SceneTreeNode
}

func (n *Node) Add(c *Node) {
	n.Children = append(n.Children, n)
}

type Tree[T any] struct {
	Id      CrdtId
	NodeMap map[CrdtId]T
}
type TreeNodeInfo struct {
	CurVersion byte
	MinVersion byte
}

const rootId = CrdtId(1)

func NewTree() (tree *SceneTree) {

	rootNode := &Node{
		Id: rootId,
	}
	return &SceneTree{
		NodeMap: map[CrdtId]*Node{rootId: rootNode},
		tree: Tree[TreeNodeInfo]{
			NodeMap: make(map[CrdtId]TreeNodeInfo),
		},
		Root: rootNode,
	}
}

func NewNode(s *SceneTreeNode) *Node {
	return &Node{
		Id: s.Id,
	}
}
func NewNodeM(s *TreeMoveInfo) *Node {
	return &Node{
		Id: s.Id,
	}
}
func (t *SceneTree) AddTree(mi *TreeMoveInfo) {
	n := NewNodeM(mi)
	t.NodeMap[mi.Id] = n
	parentId := mi.ItemInfo.ParentId
	if parentId == rootId {
		n.IsLayer = true
	}
	parent, ok := t.NodeMap[parentId]
	if !ok {
		logrus.Warn("Parent not found!")
	} else {
		parent.Add(n)

	}
}
func (t *SceneTree) AddNode(mi *SceneTreeNode) {
	node, ok := t.NodeMap[mi.Id]
	if ok && node.IsLayer {
		l := &Layer{
			Name:      mi.Name.Value,
			IsVisible: mi.Visible.Value,
		}
		node.Layer = len(t.Layers)
		t.Layers = append(t.Layers, l)
		return
	}
	if !ok {
		logrus.Warn("Node not found ", mi.Id)
	}
}
func (t *SceneTree) AddItem(item Item[SceneBaseItem], parent CrdtId) {
	node, ok := t.NodeMap[parent]
	if !ok {
		logrus.Warn("cannot find layer", parent)
		return
	}
	switch v := item.Value.(type) {
	case *LineItem:
		layer := t.Layers[node.Layer]
		layer.Lines = append(layer.Lines, v)
		logrus.Info("Got LineItem: ", v.Id)
	}
}
func (t *SceneTree) AddRootText(mi *SceneTextItem) {
}
