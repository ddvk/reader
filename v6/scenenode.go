package v6

//go:generate stringer -type=SceneItemType
import (
	"fmt"
	"image"
)

type TagType byte

const (
	InfoTag          TagType = 0
	SceneTreeTag     TagType = 1
	SceneTreeNodeTag TagType = 2
	GlyphItemTag     TagType = 3
	GroupItemTag     TagType = 4
	LineItemTag      TagType = 5
	TextItemTag      TagType = 6
	RootTextTag      TagType = 7
	UUIDIdexTag      TagType = 9
	PageInfoTag      TagType = 10
)

func (s TagType) String() string {
	var name string
	switch s {
	case LineItemTag:
		name = "Line"
	case GroupItemTag:
		name = "Group"
	case TextItemTag:
		name = "Text"
	case SceneTreeNodeTag:
		name = "SceneTreeNode"
	case RootTextTag:
		name = "RootText"
	case InfoTag:
		name = "Info"
	case PageInfoTag:
		name = "PageInfo"
	case UUIDIdexTag:
		name = "UUIDIndex"
	case SceneTreeTag:
		name = "SceneTree"
	}
	return fmt.Sprintf("%d (%s)", byte(s), name)
}

type SceneType byte

const (
	GlyphRangeType SceneType = 0x1
	GroupType      SceneType = 0x2
	LineType       SceneType = 0x3
	PathType       SceneType = 0x4
	TextType       SceneType = 0x5
)

type CrdtId uint64

func (c CrdtId) String() string {
	return fmt.Sprintf("%x(%d)", uint64(c), uint64(c))
}

type MigrationInfo struct {
	MigrationId CrdtId
	IsDevice    bool
	Bob         []byte
}
type PageInfo struct {
	Loads     int
	Merges    int
	TextChars int
	TextLinex int
	Bob       []byte
}

func (p PageInfo) String() string {
	return fmt.Sprintf("L: %d, M:%d Tc: %d, TL:%d", p.Loads, p.Merges, p.TextChars, p.TextLinex)
}

type AuthorId uint16

type Sequence[T any] struct {
	Author       AuthorId
	Id           CrdtId
	Container    []T
	Bob          []byte
	DeletedCount int
	MaxSeen      map[AuthorId]CrdtId
}

func (s *Sequence[T]) Add(item T) {
	//TODO: update MaxSeen
	s.Container = append(s.Container, item)
}

type SceneTreeNode struct {
	Id                   CrdtId
	Sequence             Sequence[SceneBaseItem]
	HasSeq               bool
	Name                 Lww[string]
	Visible              Lww[bool]
	AnchorId             Lww[CrdtId]
	AnchorMode           Lww[byte] //0 n ,1 b ,2 v
	AnchorThreshold      Lww[float32]
	AnchorInitialOriginX Lww[float32]
	Bob                  []byte
	SceneTree            *SceneTree
	Width                float32
	Height               float32
	AnchorOrigin         float32
	Info                 Info
	Children             map[CrdtId]SceneTreeNode
}

func (s SceneTreeNode) String() string {
	con := ""
	if len(s.Sequence.Container) > 0 {
		con = s.Sequence.Container[0].Item().String()

	}
	return fmt.Sprintf("SceneTreeNode: Id: %v Name:'%s' SeqId:%v %s Anchor: %v", s.Id, s.Name.Value, s.Sequence.Id, con, s.AnchorId)
}

const PenPointSize = 0x18

type PenPoint struct {
	X         float32
	Y         float32
	Speed     int16
	Width     int16
	Direction byte
	Pressure  byte
}

func (p PenPoint) String() string {
	return fmt.Sprintf("PenPoint (x:%f, y:%f, Speed: %d, Width:%d, Dir:%d, Press:%d", p.X, p.Y, p.Speed, p.Width, p.Direction, p.Pressure)
}

type Line struct {
	Color          byte
	Tool           byte
	Points         []*PenPoint
	ThicknessScale float64
	StartingLength float32
	BoundingRect   image.Rectangle
}

func (l *Line) AddPoint(p *PenPoint) {
	l.Points = append(l.Points, p)
}

func (l Line) String() string {
	return fmt.Sprintf("Line: (Color:%d, NumPoints:%d)", l.Color, len(l.Points))
}

type Info struct {
	CurrentVersion byte
	MinVersion     byte
}
type TreeItem[T any] struct {
	ParentId CrdtId
	Value    T
	Bob      []byte
}
type TreeItemInfo TreeItem[Info]
type TreeMoveInfo struct {
	Id       CrdtId
	NodeId   CrdtId
	IsUpdate bool
	//parentId inside
	ItemInfo TreeItemInfo
	Bob      []byte
}

func (t TreeMoveInfo) String() string {
	return fmt.Sprintf("TreeMoveNode: Id: %v NodeId: %v, Min:%d, Cur:%d Parent:%v", t.Id, t.NodeId, t.ItemInfo.Value.MinVersion, t.ItemInfo.Value.CurrentVersion, t.ItemInfo.ParentId)
}
