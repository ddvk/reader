package v6

import "fmt"

type Scene struct {
	Layers        []*Layer
	MigrationInfo MigrationInfo
	PageInfo      PageInfo
	UUIDMap       UUIDMap
}

func (s Scene) String() string {
	return fmt.Sprintf("Scene: Layers: %d", len(s.Layers))
}

type Layer struct {
	Name       string
	Lines      []*LineItem
	Highlights []*GlyphRange
	IsVisible  bool
}

func (s Layer) String() string {
	return fmt.Sprintf("Layer: name: %s lines: %d", s.Name, len(s.Lines))
}
