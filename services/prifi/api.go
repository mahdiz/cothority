package prifi

/*
* These are the API-commands accessible from the outside. Probably not
* needed as all calls will be done directly by the prifi-binary.
 */

import "github.com/dedis/cothority/sda"

// Client is a structure to communicate with the CoSi
// service
type Client struct {
	*sda.Client
}

// NewClient instantiates a new cosi.Client
func NewClient() *Client {
	return &Client{Client: sda.NewClient(ServiceName)}
}
