package v6

//go:generate stringer -type=SceneItemType
import (
	"fmt"
	"image"

	"github.com/google/uuid"
)

type TagType byte

const (
	InfoTag              TagType = 0
	SceneTreeNodeMoveTag TagType = 1
	SceneTreeNodeTag     TagType = 2
	GlyphItemTag         TagType = 3
	GroupItemTag         TagType = 4
	LineItemTag          TagType = 5
	TextItemTag          TagType = 6
	RootTextTag          TagType = 7
	UUIDIdexTag          TagType = 9
	PageInfoTag          TagType = 10
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
	return fmt.Sprintf("%x", uint64(c))
}

type Scene struct {
	Author              AuthorId
	AuthorUUID          uuid.UUID
	Layers              []*Layer
	Mi                  MigrationInfo
	Pi                  PageInfo
	UUIDMap             UUIDMap
	CurrentLayer        int
	IsBackgroundVisible bool
	IsNote              bool
	Tree                SceneTree
}
type Tree[T any] struct {
	Id        CrdtId
	Container map[CrdtId]T
}
type SceneTree struct {
	NextItemId CrdtId
	tree       Tree[TreeNodeInfo]
	NodeMap    map[CrdtId]*SceneTreeNode
}

type Layer struct {
	Name       string
	Lines      []*LineItem
	Highlights []*GlyphRange
	IsVisible  bool
}

type MigrationInfo struct {
	MigrationId CrdtId
	IsDevice    bool
	Bob         []byte
}
type TreeNodeInfo struct {
	curVersion uint8
	minVersion uint8
}

type PageInfo struct {
	Loads     uint32
	Merges    uint32
	TextChars uint32
	TextLinex uint32
	Bob       uint
}

func (p PageInfo) String() string {
	return fmt.Sprintf("L: %d, M:%d Tc: %d, TL:%d", p.Loads, p.Merges, p.TextChars, p.TextLinex)
}

type AuthorId uint16
type SequenceId uint64

type Sequence[T any] struct {
	Author       AuthorId
	Sequence     SequenceId
	Container    []T
	Bob          []byte
	DeletedCount int
	MaxSeen      map[AuthorId]CrdtId
}

func (s *Sequence[T]) Add(item T) {
	s.Container = append(s.Container, item)
}

type SceneTreeNode struct {
	Id                   CrdtId
	Sequence             Sequence[SceneItemBase]
	Name                 Lww[string]
	Visible              Lww[bool]
	AnchorId             Lww[CrdtId]
	AnchorMode           Lww[byte]
	AnchorThreshold      Lww[float32]
	AnchorInitialOriginX Lww[float32]
	Bob                  []byte
	SceneTree            *SceneTree
	Width                float32
	Height               float32
	AnchorOrigin         float32

	Info       Info
	Children   map[CrdtId]SceneTreeNode
	CurVersion byte
	MinVersion byte
}

func (s SceneTreeNode) String() string {
	return fmt.Sprintf("crdt:%v name:%s", s.Id, s.Name.Value)
}

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
	Info     TreeItemInfo
	Bob      []byte
}
