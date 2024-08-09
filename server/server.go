package server

import (
	"context"
	"game_server/action"
	"game_server/client"
	"game_server/packets"
	"log"
	"net"
	"strconv"
	"time"
)

const MAX_STREAM_SIZE = 2048

type DummyClient struct {
	ClientId client.ClientId
	Address  *net.UDPAddr
}

type ServerState uint

const (
	ServerNone ServerState = iota
	ServerStarted
	ServerStopping
	ServerStopped
)

// TODO(calco): Should also send the data but for now I won't ???
type EventClientPacket struct {
	// TODO(calco): Maybe send just the client id to make it safer ???
	Client *DummyClient
	Length int
}

type Server struct {
	// TODO(calco): Maybe make this a single net.UDPAddr
	IP    string
	Port  uint
	State ServerState

	Connection *net.UDPConn
	DataStream []byte

	Clients      []DummyClient
	Owner        client.ClientId
	CurrClientId client.ClientId

	// EventChan chan interface{}

	// Callbacks
	OnStarted           action.Action
	OnStopped           action.Action
	OnClientConnected   action.Action1[*DummyClient]
	OnClientDiconnected action.Action1[*DummyClient]
	OnPacketReceived    action.Action3[*DummyClient, int, []byte]

	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, cancel context.CancelFunc, ip string, port uint) *Server {
	return &Server{
		IP:    ip,
		Port:  port,
		State: ServerNone,

		DataStream: make([]byte, MAX_STREAM_SIZE),

		Clients: make([]DummyClient, 0),
		Owner:   client.CLIENT_ID_NONE,

		// EventChan: make(chan interface{}, 16),

		// Callbacks
		OnStarted:           action.NewAction(),
		OnStopped:           action.NewAction(),
		OnClientConnected:   action.NewAction1[*DummyClient](),
		OnClientDiconnected: action.NewAction1[*DummyClient](),
		OnPacketReceived:    action.NewAction3[*DummyClient, int, []byte](),

		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Server) GetAddress() string {
	return s.IP + ":" + strconv.FormatUint(uint64(s.Port), 10)
}

func (s *Server) Start() {
	log.Printf("INFO: Trying to start server on address: [%q]", s.GetAddress())
	addr, err := net.ResolveUDPAddr("udp", s.GetAddress())
	if err != nil {
		log.Fatalf("ERROR: Failed resolving UDP address [%q]!", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("ERROR: Failed listening to UDP [%q]!", err)
	}

	log.Print("LOG: Established connection with UDP.")
	s.Connection = conn
	s.State = ServerStarted

	// s.EventChan <- EventStarted{}
	s.OnStarted.Invoke()

	go func() {
		<-s.ctx.Done()
		s.handleStop()
	}()

	for {
		select {
		// NOTE(calco): listenToUDP starts failing when disconnecting.
		case <-s.ctx.Done():
			return
		default:
			s.listenToUDP()
		}
	}
}

func (s *Server) listenToUDP() {
	for {
		n, addr, err := s.Connection.ReadFromUDP(s.DataStream)
		if err != nil {
			if s.State == ServerStopping || s.State == ServerStopped {
				return
			}
			log.Printf("WARNING: Failed while reading from UDP [%q]!", err)
		}
		if n == 0 {
			log.Printf("WARNING: Someone sent 0 bytes lmfao.")
			continue
		}

		c, exists := s.hasDummyClient(addr)
		if !exists {
			dataStream := packets.NewDataStream(s.DataStream, uint(n))
			value, err := dataStream.ReadByte()

			log.Printf("LOG: Received [%d] bytes from [%s]: [%q]", n, addr.String(), s.DataStream[:n])

			if err == nil && dataStream.Finished && value == packets.PING_PACKET {
				// TODO(calco): Maybe this should be in a goroutine, but frankly, I don't want to do it now.
				// n, err := s.SendToClient(clientId, []byte{packets.PONG_PACKET})

				wouldBeId := s.CurrClientId
				wouldBeOwner := s.CurrClientId == 0
				// TODO(calco): If you somehow have more than 256 client ids uh ... too bad?
				data := []byte{packets.PONG_PACKET, byte(wouldBeId), Bool2byte(wouldBeOwner)}
				n, err := s.Connection.WriteToUDP(data, addr)
				if err != nil {
					log.Printf("WARN: Failed sending message to client (Addr [%s]): [%q]!", addr.String(), err)
				} else if n < len(data) {
					log.Printf("WARN: Failed sending all bytes to client (Addr [%s]). Only sent [%d!]", addr.String(), n)
				} else {
					log.Printf("LOG: Send [%d] bytes to client (Addr [%s]): [%q]", n, addr.String(), packets.PONG_PACKET)
				}

				if err != nil || n < len(data) {
					log.Printf("WARN: Couldn't verify client connection. Rejecting!")
					continue
				}
				s.Clients = append(s.Clients, DummyClient{
					ClientId: s.CurrClientId.Next(),
					Address:  addr,
				})
				c = &s.Clients[len(s.Clients)-1]
				log.Printf("INFO: Client [%s] connected and received ID [%d]!", c.Address.String(), c.ClientId)

				if len(s.Clients) == 1 {
					s.Owner = c.ClientId
					log.Printf("INFO: Established Client (ID [%d] | Addr [%s]) as owner!", c.ClientId, c.Address.String())
				}
				// s.EventChan <- EventClientConnected{Client: c}
				s.OnClientConnected.Invoke(c)
			} else {
				if err != nil {
					log.Printf("WARN: Failed to read byte from received message!")
				}
				log.Printf("INFO: Address [%s] tried connecting to server but sent incorrect data [%q]!", addr.String(), s.DataStream[:n])
				continue
			}
		} else {
			log.Printf("LOG: Received [%d] bytes from [%d]: [%q]", n, c.ClientId, s.DataStream[:n])

			// TODO(calco): SHOULD NOT copy the data maybe, but I don't want to deal with stuff now
			dataCopy := make([]byte, n)
			copy(dataCopy, s.DataStream)
			// TODO(calco): this is BLOCKING right now
			s.OnPacketReceived.Invoke(c, n, dataCopy)
		}
	}
}

func (s *Server) Broadcast(bytes []byte) (*DummyClient, int, error) {
	var n int
	for _, c := range s.Clients {
		n, err := s.sendToDummyClient(&c, bytes)
		if err != nil {
			return &c, n, err
		}
	}
	return nil, n, nil
}

func (s *Server) SendToClient(clientId client.ClientId, bytes []byte) (int, error) {
	c, exists := s.hasClientId(clientId)
	if !exists {
		log.Printf("WARN: Tried sending message to invalid client ID [%d].", clientId)
	}
	return s.sendToDummyClient(c, bytes)
}

func (s *Server) sendToDummyClient(c *DummyClient, bytes []byte) (int, error) {
	n, err := s.Connection.WriteToUDP(bytes, c.Address)
	if err != nil {
		log.Printf("WARN: Failed sending message to client (ID [%d] | Addr [%s]): [%q]!", c.ClientId, c.Address.String(), err)
	} else if n != len(bytes) {
		log.Printf("WARN: Failed sending all bytes to client (ID [%d] | Addr [%s]). Only sent [%d!]", c.ClientId, c.Address.String(), n)
	} else {
		log.Printf("LOG: Send [%d] bytes to client (ID [%d] | Addr [%s]): [%q]", n, c.ClientId, c.Address.String(), bytes)
	}

	return n, err
}

func (s *Server) Stop() {
	if s.State != ServerStarted {
		s.handleStop()
		return
	}
	stopCh := make(chan bool)
	s.OnStopped.Subscribe(func() {
		stopCh <- true
	})
	s.cancel()
	<-stopCh
	close(stopCh)
}

func (s *Server) PollStopped(duration time.Duration) <-chan struct{} {
	stopChan := make(chan struct{})
	go func() {
		defer close(stopChan)
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		for {
			<-ticker.C
			if s.State == ServerStopped {
				return
			}
		}
	}()
	return stopChan
}

func (s *Server) handleStop() {
	log.Print("LOG: Received stop signal. Stopping...")
	s.State = ServerStopping
	if err := s.Connection.Close(); err != nil {
		log.Printf("ERROR & WARNING: Failed to stop the server ??? [%q]", err)
	}
	s.State = ServerStopped
	// s.EventChan <- EventStopped{}
	s.OnStopped.Invoke()
	// close(s.EventChan)
	log.Print("LOG: Stopped server.")
}

// NOTE(calco): Yes, technically dumb, but "O(n) is faster than O(1) for a small enough n and long enough 1."
// And we WON'T have more than 4 players at the same time lmfao
func (s *Server) hasDummyClient(addr *net.UDPAddr) (*DummyClient, bool) {
	for _, c := range s.Clients {
		if c.Address.IP.Equal(addr.IP) && c.Address.Port == addr.Port {
			return &c, true
		}
	}

	return nil, false
}

func (s *Server) hasClientId(clientId client.ClientId) (*DummyClient, bool) {
	for _, c := range s.Clients {
		if c.ClientId == clientId {
			return &c, true
		}
	}

	return nil, false
}

// Taken from https://0x0f.me/blog/golang-compiler-optimization/
func Bool2byte(b bool) byte {
	// The compiler currently only optimizes this form.
	// See issue 6011.
	var i byte
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}
