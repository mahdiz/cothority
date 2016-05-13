package sda2

import (
	"sync"
)

// Overlay maintains the different EntityList + Trees that we know of, and
// provides the bridge between actual network connections and TreeNode on a
// Tree.
// XXX Should overlay be an inteface ? Since we dont need to *mock* it for
// testing, and don't intend to support multiple overlays (maybe we should?)
// it could be best to just use a regular struct. I let it like this so we can
// decide.
type Overlay interface {
	RegisterProtocolInstance(ProtocolInstance)
	NewTreeNodeInstance(*Tree, *TreeNode)
}

// treeOverlay implements the Overlay interface
type treeOverlay struct {
	router Router

	// keep track of entitylistS and treeS
	entities *entityStore
	trees    *treeStore
}

func newTreeOverlay(r Router) *treeOverlay {
	return &treeOverlay{
		router:   r,
		entities: newEntityStore(),
		trees:    newTreeStore(),
	}
}

func (to *treeOverlay) registerTree(t *Tree) error {
	return nil
}

func (to *treeOverlay) registerEntityList(el *EntityList) {

}

func (to *treeOverlay) entityList(id EntityListID) *EntityList {
	return nil
}

func (to *treeOverlay) tree(id TreeID) *Tree {
	return nil
}

// overlayMessage is a struct that is used to send and receive messages between
// overlays. ProtocolInstance will send messages that will be embedded inside
// overlayMessage.
type overlayMessage struct {
}

type entityStore struct {
	lists map[EntityListID]*EntityList
	mutex sync.Mutex
}

func newEntityStore() *entityStore {
	return &entityStore{
		lists: make(map[EntityListID]*EntityList),
	}
}

type treeStore struct {
	trees map[TreeID]*Tree
	mutex sync.Mutex
}

func newTreeStore() *treeStore {
	return &treeStore{
		trees: make(map[TreeID]*Tree),
	}
}
