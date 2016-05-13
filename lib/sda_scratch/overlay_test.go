package sda2

import (
	"github.com/dedis/cothority/lib/dbg"
	"github.com/dedis/cothority/lib/network"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOverlayAddEntityList(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 4)

	r1 := newMockRouter()
	defer r1.close()
	o1 := newTreeOverlay(r1)

	el := NewEntityList([]*network.Entity{r1.Entity()})
	o1.registerEntityList(el)

	assert.NotNil(t, o1.entityList(el.Id))
}

func TestOverlayAddTree(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 4)

	r1 := newMockRouter()
	defer r1.close()
	o1 := newTreeOverlay(r1)

	el := NewEntityList([]*network.Entity{r1.Entity()})

	// Registering the tree without the entitylist should fail
	tree := el.GenerateBinaryTree()
	assert.Error(t, o1.registerTree(tree))

	o1.registerEntityList(el)
	assert.Nil(t, o1.registerTree(tree))
	assert.NotNil(t, o1.tree(tree.Id))

}

// Test if two overlays can exchange messages (on two different router)
func TestOverlayMessaging(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 4)

	/* r1 := newMockRouter()*/
	//o1 := newTreeOverlay(r1)
	//r2 := newMockRouter()
	//o2 := newTreeOverlay(r2)

	//msg := &overlayMessage{}

}

func TestOverlayNewTreeNodeInstance(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 4)
	// TODO
}
