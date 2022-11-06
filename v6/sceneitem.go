package v6

import (
	"fmt"
	"image"
)

type SceneItemBase interface {
	Item() *SceneItem
}
type SceneItem struct {
	Id         CrdtId
	ParentId   CrdtId
	Type       SceneType
	CurVersion byte //=2
	MinVersion byte
	IsDirty    bool
	Bob        []byte
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
	return fmt.Sprintf("LineItem: Id:%v,  %v", t.Id, t.Line.Value)
}

type SceneTextItem struct {
	SceneItem
	Position image.Point
}

func (t *SceneTextItem) Item() *SceneItem {
	return &t.SceneItem
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
