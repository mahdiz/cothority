package prifi

/*
This holds the messages used to communicate with the service over the network.
*/

import "github.com/dedis/cothority/network"

// We need to register all messages so the network knows how to handle them.
func init() {
	// All messages can be defined in this for-loop for convenience.
	for _, msg := range []interface{}{} {
		network.RegisterPacketType(msg)
	}
}
