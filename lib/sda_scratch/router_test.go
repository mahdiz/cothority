package sda2

import (
	"github.com/dedis/cothority/lib/dbg"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRouterTcpListenAndBind(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 4)

	tr1 := freshTcpRouter("127.0.0.1:2000")
	defer tr1.Close()
	assert.Nil(t, tr1.listenAndBind())

	assert.NotNil(t, tr1.conns.Get(tr1.Entity()))
}

func TestRouterTcpListenAndConnect(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 4)

	tr1 := freshTcpRouter("127.0.0.1:2000")
	defer tr1.Close()
	assert.Nil(t, tr1.listen())
	dbg.Lvl1("Router listening on :2000")
	tr2 := freshTcpRouter("127.0.0.1:2010")
	defer tr2.Close()

	assert.Nil(t, tr2.connect(tr1.Entity()))
	assert.NotNil(t, tr2.conns.Get(tr1.Entity()))
	//assert.NotNil(t, tr1.conns.Get(tr2.entity()))
}

// Test closing and opening of Host on same address
func TestRouterTcpClose(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 4)

	h1 := freshTcpRouter("127.0.0.1:2000")
	h2 := freshTcpRouter("127.0.0.1:2010")
	assert.Nil(t, h1.listenAndBind())
	assert.Nil(t, h2.connect(h1.tcpHost.Entity()))

	assert.Nil(t, h1.Close())
	assert.Nil(t, h2.Close())
	dbg.Lvl3("Finished first connection, starting 2nd")
	h3 := freshTcpRouter("127.0.0.1:2020")
	h3.listenAndBind()

	assert.Nil(t, h2.connect(h3.tcpHost.Entity()))
	dbg.Lvl3("Closing h3")
	assert.Nil(t, h3.Close())
}

// Test connection of multiple Hosts and sending messages back and forth
// also tests for the counterIO interface that it works well
func TestRouterTcpProcessing(t *testing.T) {
	defer dbg.AfterTest(t)
	dbg.TestOutput(testing.Verbose(), 5)

	// create router
	r1 := freshTcpRouter("127.0.0.1:2000")
	r2 := freshTcpRouter("127.0.0.1:2010")
	assert.Nil(t, r2.listenAndBind())
	assert.Nil(t, r1.connect(r2.Entity()))

	// create processor and registers it to r2
	p := newSimpleProcessor()
	r2.RegisterProcessor(p, statusMsgID)

	// send the message
	status := &statusMessage{true, 10}

	assert.Nil(t, r1.Route(r2.tcpHost.Entity(), status))

	// see if it's correctly received on the r2 side
	select {
	case m := <-p.relay:

		assert.True(t, m.Ok)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Waited too long")
	}
	assert.Nil(t, r1.Close())
	assert.Nil(t, r2.Close())
}

// XXX move lib/sda/host_test.go:TestHostMessaging COunterIO part here.
func TestRouterTcpCounterIO(t *testing.T) {

}
