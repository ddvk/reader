package v6

import "github.com/google/uuid"

type Scene struct {
	Author              AuthorId
	AuthorUUID          uuid.UUID
	Layers              []*Layer
	MigrationInfo       MigrationInfo
	PageInfo            PageInfo
	UUIDMap             UUIDMap
	CurrentLayer        int
	IsBackgroundVisible bool
	IsNote              bool
	Tree                SceneTree
}

type Layer struct {
	Name       string
	Lines      []*LineItem
	Highlights []*GlyphRange
	IsVisible  bool
}
