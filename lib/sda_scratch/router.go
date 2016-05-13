package sda2

import (
	"errors"
	"github.com/dedis/cothority/lib/dbg"
	"github.com/dedis/cothority/lib/network"
	"golang.org/x/net/context"
	"sync"
	"time"
)

// Router is the interface that is responsible for handling incoming and
// outgoing messages and maintaining connections. Later, we should be able to
// register different kind of events regarding the network.
type Router interface {
	// Route will send the msg to the right entity, i.e. destination
	Route(dst *network.Entity, msg network.ProtocolMessage) error

	// RegisterProcessor takes a Processor and a message type. Each time a
	// message that has this type comes from the network/external world, it
	// should be dispatched to this processor.
	// The actual message structure should be registered to the network library
	// with network.RegisterMessageType(msg interfaace{}) => MessageID
	// TODO Transform uuid.UUID into MessageID
	// XXX If multiple Processor wants the same message, should it be copied?
	RegisterProcessor(Processor, network.MessageTypeID) error

	// Entity returns the Entity used by this Router to handle connections
	Entity() *network.Entity

	/* // ActiveConns returns the list of connections that are active on the Router*/
	//ActiveConns() []network.SecureConn

	//// ListeningAddress returns the address where this router listens.
	//ListeningAddress() string

}

// tcpRouter is the implementation of Router using tcp connections
type tcpRouter struct {
	tcpHost *network.SecureTCPHost
	conns   *connStore
	// processors that receive the different incoming messages
	procs *procStore

	// if we already closed the router
	isClosing  bool
	closingMut sync.Mutex
}

func newTcpRouter(host *network.SecureTCPHost) *tcpRouter {
	return &tcpRouter{
		tcpHost: host,
		conns:   newConnStore(),
		procs:   newProcStore(),
	}
}

func (tr *tcpRouter) listen() error {
	return tr.listen_(false)
}

func (tr *tcpRouter) listenAndBind() error {
	return tr.listen_(true)
}

func (tr *tcpRouter) ListeningAddress() string {
	return tr.tcpHost.String()
}

const maxRetry = 5

func (tr *tcpRouter) listen_(wait bool) error {
	dbg.Lvl3(tr.ListeningAddress(), "starts to listen")
	fn := func(c network.SecureConn) {
		dbg.Lvl3(tr.tcpHost.String(), "Accepted Connection from", c.Remote())
		// register the connection once we know it's ok
		tr.registerConnection(c)
		tr.handleConn(c)
	}
	go func() {
		if err := tr.tcpHost.Listen(fn); err != nil {
			dbg.Error("error quitting the listener", err)
		}
	}()
	if wait {
		var curr int
		for curr < maxRetry {
			dbg.Lvl3(tr.tcpHost.String(), "checking if listener is up")
			err := tr.connect(tr.tcpHost.Entity())
			if err == nil {
				dbg.Lvl3(tr.tcpHost.String(), "managed to connect to itself")
				break
			}
			time.Sleep(network.WaitRetry)
			curr++
		}
		if curr == maxRetry {
			return errors.New("Could not bind it self")
		}
	}
	return nil
}

// Connect takes an entity where to connect to
func (tr *tcpRouter) connect(e *network.Entity) error {
	var err error
	var c network.SecureConn
	// try to open connection
	c, err = tr.tcpHost.Open(e)
	if err != nil {
		return err
	}
	dbg.Lvl3("Host", tr.ListeningAddress(), "connected to", c.Remote())
	tr.registerConnection(c)
	go tr.handleConn(c)
	return nil
}

// registerConnection registers a Entity for a new connection, mapped with the
// real physical address of the connection and the connection itself
// it locks (and unlocks when done): entityListsLock and networkLock
func (tr *tcpRouter) registerConnection(c network.SecureConn) {
	tr.conns.Put(c)
}

// Handle a connection => giving messages to the MsgChans
func (tr *tcpRouter) handleConn(c network.SecureConn) {
	address := c.Remote()
	ctx := context.TODO()
	for {
		am, err := c.Receive(ctx)
		// So the receiver can know about the error
		am.SetError(err)
		am.From = address
		dbg.Lvl5(tr.ListeningAddress(), "Got message", am)
		if err != nil {
			if err == network.ErrClosed || err == network.ErrEOF || err == network.ErrTemp {
				dbg.Lvl5(tr.ListeningAddress(), "quitting connection with", address)
				return
			}
			dbg.Error(tr.ListeningAddress(), "Error with connection", address, "=>", err)
		}
		// dispatch it
		tr.dispatchMessage(&am)
	}
}

// dispatchMessage dispatch the message to the right Processor
// XXX This is in a separate function so we can change our dispatching system
// easily. For example, for the day we will want to move to a more
// producer / worker paradigm where connections routines will simply put
// messages into a buffer and workers will make the dispatching
// Performance reasons: see
// http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/
func (tr *tcpRouter) dispatchMessage(msg *network.Message) {
	// XXX Need to fix msg.MsgType = MessageTypeID and not just uuid.UUID
	var p Processor
	if p = tr.procs.Get(msg.MsgType); p == nil {
		return
	}
	p.Process(msg)
}

// Close shuts down the listener
func (tr *tcpRouter) Close() error {

	tr.closingMut.Lock()
	if tr.isClosing {
		tr.closingMut.Unlock()
		return errors.New("Already closing")
	}
	dbg.Lvl3(tr.ListeningAddress(), "Starts closing")
	tr.isClosing = true
	tr.closingMut.Unlock()
	// XXX to see how the processing is done before having a
	// ProcessMessageQuit...
	/*if h.processMessagesStarted {*/
	//// Tell ProcessMessages to quit
	//close(h.ProcessMessagesQuit)
	/*}*/
	// XXX TCPHost.Close already closes all connections ...
	/*for _, c := range h.connections {*/
	//dbg.Lvl3(h.Entity.First(), "Closing connection", c)
	//err := c.Close()
	//if err != nil {
	//dbg.Error(h.Entity.First(), "Couldn't close connection", c)
	//return err
	//}
	/*}*/
	err := tr.tcpHost.Close()
	tr.conns.Empty()
	dbg.Lvl3(tr.ListeningAddress(), "Closed")
	return err
}

// RegisterProcessor takes a processor and a message type so each time a
// incoming messages has this specific type, it will be dispatch to the
// processor. If there was already a processor registered for this msgtype, an
// error is returned.
func (tr *tcpRouter) RegisterProcessor(p Processor, msgType network.MessageTypeID) error {
	if ok := tr.procs.Put(msgType, p); !ok {
		return errors.New("Processor already registered for this message type")
	}
	return nil
}

// Route takes an entity as a destination and send the message to this
// destination.
func (tr *tcpRouter) Route(e *network.Entity, msg network.ProtocolMessage) error {

	var c network.SecureConn
	c = tr.conns.Get(e)
	if c == nil {
		dbg.Lvl4(tr.ListeningAddress(), "Route() to an unknown destination. Will try to connect")
		// try to connect to it automatically
		if err := tr.connect(e); err != nil {
			return err
		}
		c = tr.conns.Get(e)
	}
	return c.Send(context.TODO(), msg)
}

func (tr *tcpRouter) Entity() *network.Entity {
	return tr.tcpHost.Entity()
}

// connStore holds all the connections in a thread safe manner
type connStore struct {
	// keeps track of all connections
	connections map[network.EntityID]network.SecureConn
	mutex       sync.Mutex
}

func newConnStore() *connStore {
	return &connStore{
		connections: make(map[network.EntityID]network.SecureConn),
	}
}

func (c *connStore) Get(e *network.Entity) network.SecureConn {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	s, _ := c.connections[e.ID]
	return s
}

func (c *connStore) Put(s network.SecureConn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connections[s.Entity().ID] = s
}

func (c *connStore) List() []network.SecureConn {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var list []network.SecureConn
	for _, v := range c.connections {
		list = append(list, v)
	}
	return list
}

func (c *connStore) Entities() []*network.Entity {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var list []*network.Entity
	for _, v := range c.connections {
		list = append(list, v.Entity())
	}
	return list
}

func (c *connStore) Empty() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connections = make(map[network.EntityID]network.SecureConn)
}

// procStore keeps tracks of all processors in a thread safe manner
type procStore struct {
	processors map[network.MessageTypeID]Processor
	mutex      sync.Mutex
}

func newProcStore() *procStore {
	return &procStore{
		processors: make(map[network.MessageTypeID]Processor),
	}
}

func (p *procStore) Get(t network.MessageTypeID) Processor {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	proc, _ := p.processors[t]
	return proc
}

// Put will store the processor IF AND ONLY IF there is no already-registered
// processor for this messagetypeid. Return true if Insertion went well, False
// if a processor already exists.
func (p *procStore) Put(t network.MessageTypeID, proc Processor) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if _, ok := p.processors[t]; ok {
		return false
	}
	p.processors[t] = proc
	return true

}

func (p *procStore) Empty() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.processors = make(map[network.MessageTypeID]Processor)
}
