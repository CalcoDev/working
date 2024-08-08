package client

import "net"

const CLIENT_ID_NONE = 0

type ClientId uint

func (c *ClientId) Next() ClientId {
	*c += 1
	return *c - 1
}

type Client struct {
	ClientId ClientId
	Address  net.UDPAddr
}
