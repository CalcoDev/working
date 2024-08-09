package client

import (
	"context"
	"game_server/packets"
	"log"
	"net"
	"strconv"
)

const MAX_STREAM_SIZE = 2048

const CLIENT_ID_NONE = -1

type ClientId int

func (c *ClientId) Next() ClientId {
	*c += 1
	return *c - 1
}

func (c *ClientId) PeekNext() ClientId {
	return *c
}

type ClientState uint

const (
	ClientNone ClientState = iota
	ClientStarted
	ClientStopped
)

type DummyServer struct {
	Address *net.UDPAddr
}

type Client struct {
	// TODO(calco): Maybe make this a single net.UDPAddr
	// IP    string
	Port  uint
	State ClientState

	IsOwner  bool
	ClientId ClientId

	Connection *net.UDPConn
	DataStream []byte

	Server DummyServer

	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, cancel context.CancelFunc) *Client {
	return &Client{
		// IP:       ip,
		Port:       0,
		State:      ClientNone,
		IsOwner:    false,
		ClientId:   CLIENT_ID_NONE,
		DataStream: make([]byte, MAX_STREAM_SIZE),

		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) GetAddress() string {
	return "127.0.0.1:" + strconv.FormatUint(uint64(c.Port), 10)
}

func (c *Client) Start(server_address string) {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("ERROR: Failed resolving UDP address of client [%q]!", err)
	}
	c.Port = uint(laddr.Port)
	log.Printf("LOG: Trying to start client with UDP. Addr [%s].", c.GetAddress())

	raddr, err := net.ResolveUDPAddr("udp", server_address)
	if err != nil {
		log.Fatalf("ERROR: Failed resolving UDP address of server [%q]!", err)
	}
	c.Server = DummyServer{
		Address: raddr,
	}

	conn, err := net.DialUDP("udp", laddr, c.Server.Address)
	if err != nil {
		log.Fatalf("ERROR: Failed dialing UDP [%q]!", err)
	}
	log.Printf("LOG: Established connection with UDP. Addr [%s].", c.GetAddress())
	c.Connection = conn

	log.Printf("LOG: Sending PING packet to server...")
	c.Send([]byte{packets.PING_PACKET})

	n, addr, err := c.Connection.ReadFromUDP(c.DataStream)
	if err != nil {
		log.Printf("WARN: Error receiving PONG packet from server!")
		c.Stop()
		return
	} else if n < 3 { // TODO(calco): THIS SHOULD NOT BE HARDCODED LOL
		log.Printf("WARN: Server sent back PONG packet with too little information!")
		c.Stop()
		return
	} else {
		log.Printf("LOG: Received [%d] bytes from server [%s]: [%q]", n, addr.String(), c.DataStream[:n])
	}

	if c.DataStream[0] == packets.PONG_PACKET {
		log.Printf("LOG: Confirmed PONG packet from server! Connection established.")
		c.ClientId = ClientId(c.DataStream[1])
		c.IsOwner = Byte2bool(c.DataStream[2])
		log.Printf("LOG: Established self information. Client ID [%d] | IsOwner [%v].", c.ClientId, c.IsOwner)
	} else {
		log.Printf("WARN: Packet mismatch! Expected PONG [%d] but received [%d]!", packets.PONG_PACKET, c.DataStream[0])
	}

	// TODO(calco): Slightly odd to start the loop here, but if no confirmation, no listen ??
	// go func() {
	// 	<-c.ctx.Done()
	// 	c.handleStop()
	// }()

	// for {
	// 	select {
	// 	case <-c.ctx.Done():
	// 		return
	// 	default:

	// 	}
	// }
}

func (c *Client) Send(bytes []byte) {
	n, err := c.Connection.Write(bytes)
	if err != nil {
		log.Printf("WARNING: Failed sending message to server.")
	}
	log.Printf("LOG: Sent %d bytes to server [%q].", n, bytes[:n])
}

// TODO(calco): add this function lmao
func (c *Client) Stop() {

}

func (c *Client) handleStop() {

}

// Adapted from https://0x0f.me/blog/golang-compiler-optimization/
func Byte2bool(b byte) bool {
	// The compiler currently only optimizes this form.
	// See issue 6011.
	var i bool
	if b == 0 {
		i = false
	} else {
		i = true
	}
	return i
}
