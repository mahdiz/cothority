package sda2

import (
	"github.com/dedis/cothority/lib/network"
)

// Processor is a generic interface that is used to receive message. It can be
// registered against a Router.
type Processor interface {
	Process(*network.Message)
}

// ProcHandler is a Processor that provides nice utilities to directly register
// specific messages type that we can listen on a channel
type ProcHandler struct {
}

// ProcChannel is a Processor that provides nice utilities to provide handler
// for each type of message that you want to receive.
type ProcChannel struct {
}
