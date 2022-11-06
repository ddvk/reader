package v6

import "github.com/google/uuid"

type UUIDMap struct {
	UUID2Index map[uuid.UUID]AuthorId
	Index2UUID map[AuthorId]uuid.UUID
	Max        AuthorId
}

func NewMap() UUIDMap {
	return UUIDMap{
		UUID2Index: make(map[uuid.UUID]AuthorId),
		Index2UUID: make(map[AuthorId]uuid.UUID),
	}
}

func (um *UUIDMap) Entries() int {
	return len(um.Index2UUID)
}

func (um *UUIDMap) Add(u uuid.UUID, index AuthorId) {
	if um.Index2UUID == nil {
		um.Index2UUID = make(map[AuthorId]uuid.UUID)
	}
	if um.UUID2Index == nil {
		um.UUID2Index = make(map[uuid.UUID]AuthorId)
	}
	um.Index2UUID[index] = u
	um.UUID2Index[u] = index

	if index > um.Max {
		um.Max = index
	}
}
