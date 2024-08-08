package client

import (
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
	return *c + 1
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

	Server DummyServer
}

func New() *Client {
	return &Client{
		// IP:       ip,
		Port:     0,
		State:    ClientNone,
		IsOwner:  false,
		ClientId: CLIENT_ID_NONE,
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
