package v6

type SceneTree struct {
	NextItemId CrdtId
	tree       Tree[TreeNodeInfo]
	NodeMap    map[CrdtId]*SceneTreeNode
}

type Tree[T any] struct {
	Id        CrdtId
	Container map[CrdtId]T
}
