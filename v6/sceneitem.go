package v6

import (
	"fmt"
	"image"
)

type SceneBaseItem interface {
	Item() *SceneItem
}

type SceneItem struct {
	Id       CrdtId
	ParentId CrdtId
	Type     SceneType
	Info     Info
	IsDirty  bool
	Bob      []byte
}

func (t SceneItem) String() string {
	return fmt.Sprintf("SceneItem (%v) Id:%v Parent:%v", t.Type, t.Id, t.ParentId)
}

func (t *SceneItem) Item() *SceneItem {
	return t
}

type LineItem struct {
	SceneItem
	Line Lww[Line]
}

func (t *LineItem) Item() *SceneItem {
	return &t.SceneItem
}

func (t LineItem) String() string {
	return fmt.Sprintf("LineItem: Id:%v,  %v, timestamp:%d", t.Id, t.Line.Value, t.Line.Timestamp)
}

type SceneTextItem struct {
	SceneItem
	Sequence Sequence[*Item[TextItem]]
	Position Point
}
type Point struct {
	X float64
	Y float64
}

func (t *SceneTextItem) Item() *SceneItem {
	return &t.SceneItem
}

func (t SceneTextItem) String() string {
	return fmt.Sprintf("SceneText: Id: %v, containerId:%v num:%d", t.Id, t.Sequence.Id, len(t.Sequence.Container))
}

type GlyphRange struct {
	SceneItem
	Start            int
	Length           int
	Color            byte
	Text             string
	Rectangles       []*image.Rectangle
	FirstId          CrdtId
	LastId           CrdtId
	IsLastIdIncluded bool
}

func (t *GlyphRange) Item() *SceneItem {
	return &t.SceneItem
}

func (t GlyphRange) String() string {
	return fmt.Sprintf("GlyphRange: Id: %v, Text:%s Length:%d", t.Id, t.Text, t.Length)
}

type GroupItem struct {
	SceneItem
	NodeId   CrdtId
	TreeNode SceneTreeNode
}

func (t *GroupItem) Item() *SceneItem {
	return &t.SceneItem
}

func (t GroupItem) String() string {
	return fmt.Sprintf("GroupItem: Id:%v, TreeNodeId:%d", t.NodeId, t.TreeNode.Id)
}

type TextItem struct {
	Text   string
	Format uint32
}
type Item[T any] struct {
	Id            CrdtId
	Left          CrdtId
	Right         CrdtId
	Value         T
	DeletedLength int
	Bob           []byte
}

func (t Item[T]) String() string {
	return fmt.Sprintf("Item: Id: %v Val: %v", t.Id, t.Value)
}
