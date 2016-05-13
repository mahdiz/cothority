package sda2

import (
	"errors"
	"github.com/dedis/cothority/lib/network"
	"github.com/dedis/crypto/abstract"
	"github.com/dedis/crypto/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// This file provides some helpers methods for testing the internals of SDA.
var tSuite = network.Suite

func init() {
	statusMsgID = network.RegisterMessageType(statusMessage{})
}

// Returns a Private key and associated Entity
func newPrivEntityPair(address string) (abstract.Secret, *network.Entity) {
	kp := config.NewKeyPair(tSuite)
	e := network.NewEntity(kp.Public, address)
	return kp.Secret, e
}

func freshTcpRouter(address string) *tcpRouter {
	s, e := newPrivEntityPair(address)
	tcpHost := network.NewSecureTCPHost(s, e)
	tr := newTcpRouter(tcpHost)
	return tr
}

func waitOrFatal(t *testing.T, ch chan bool, value bool) {
	select {
	case v := <-ch:
		if v == value {
			return
		} else {
			t.Fatal("Wrong value")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("waited too long")
	}
}

// keep tracks of all the mock routers
var mockRouters = make(map[network.EntityID]*mockRouter)

// mockRouter is a struct that implements the Router interface locally
type mockRouter struct {
	procs   *procStore
	entity  *network.Entity
	msgChan chan *network.Message
}

func newMockRouter() *mockRouter {
	_, e := newPrivEntityPair("localhost:0")
	r := &mockRouter{
		procs:   newProcStore(),
		entity:  e,
		msgChan: make(chan *network.Message),
	}
	mockRouters[e.ID] = r
	go r.dispatch()
	return r
}

func (m *mockRouter) RegisterProcessor(p Processor, t network.MessageTypeID) error {
	if !m.procs.Put(t, p) {
		return errors.New("Processor already registered")
	}
	return nil
}

func (m *mockRouter) Route(e *network.Entity, msg network.ProtocolMessage) error {
	r, ok := mockRouters[e.ID]
	if !ok {
		return errors.New("No mock routers at this entity")
	}
	// simulate network marshaling / unmarshaling
	b, err := network.MarshalRegisteredType(msg)
	if err != nil {
		return err
	}

	t, unmarshalled, err := network.UnmarshalRegisteredType(b, network.DefaultConstructors(network.Suite))
	if err != nil {
		return err
	}
	nm := network.Message{
		Msg:     unmarshalled,
		MsgType: t,
		Entity:  m.entity,
	}
	r.msgChan <- &nm
	return nil
}

func (m *mockRouter) dispatch() {
	for msg := range m.msgChan {
		var p Processor
		if p = m.procs.Get(msg.MsgType); p == nil {
			return
		}
		p.Process(msg)
	}
}

func (m *mockRouter) Entity() *network.Entity {
	return m.entity
}
func (m *mockRouter) close() {
	close(m.msgChan)
}

type statusMessage struct {
	Ok  bool
	Val int
}

var statusMsgID network.MessageTypeID

type simpleProcessor struct {
	relay chan statusMessage
}

func newSimpleProcessor() *simpleProcessor {
	return &simpleProcessor{
		relay: make(chan statusMessage),
	}
}
func (sp *simpleProcessor) Process(msg *network.Message) {
	if msg.MsgType != statusMsgID {

		sp.relay <- statusMessage{false, 0}
	}
	sm := msg.Msg.(statusMessage)

	sp.relay <- sm
}

func TestMockRouter(t *testing.T) {
	m1 := newMockRouter()
	defer m1.close()
	m2 := newMockRouter()
	defer m2.close()
	assert.NotNil(t, mockRouters[m1.Entity().ID])
	assert.NotNil(t, mockRouters[m2.Entity().ID])

	p := newSimpleProcessor()
	assert.Nil(t, m2.RegisterProcessor(p, statusMsgID))

	status := &statusMessage{true, 10}
	assert.Nil(t, m1.Route(m2.Entity(), status))

	select {
	case m := <-p.relay:
		if !m.Ok || m.Val != 10 {
			t.Fatal("Wrong value")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("waited too long")
	}
}
