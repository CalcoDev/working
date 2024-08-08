package server

import (
	"context"
	"game_server/client"
	"game_server/packets"
	"log"
	"net"
	"strconv"
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
	ServerStopped
)

type EventStarted struct{}
type EventStopped struct{}
type EventClientConnected struct {
	Client *DummyClient
}
type EventClientDisconnected struct {
	Client *DummyClient
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

	EventChan chan interface{}

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

		EventChan: make(chan interface{}, 16),

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

	s.EventChan <- EventStarted{}

	go func() {
		<-s.ctx.Done()
		s.handleStop()
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			n, addr, err := s.Connection.ReadFromUDP(s.DataStream)
			if err != nil {
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

					n, err := s.Connection.WriteToUDP([]byte{packets.PONG_PACKET}, addr)
					if err != nil {
						log.Printf("WARN: Failed sending message to client (Addr [%s]): [%q]!", addr.String(), err)
					} else if n != 1 {
						log.Printf("WARN: Failed sending all bytes to client (Addr [%s]). Only sent [%d!]", addr.String(), n)
					} else {
						log.Printf("LOG: Send [%d] bytes to client (Addr [%s]): [%q]", n, addr.String(), packets.PONG_PACKET)
					}

					if err != nil || n != 1 {
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
					}
					s.EventChan <- EventClientConnected{Client: c}
				} else {
					if err != nil {
						log.Printf("WARN: Failed to read byte from received message!")
					}
					log.Printf("INFO: Address [%s] tried connecting to server but sent incorrect data [%q]!", addr.String(), s.DataStream[:n])
					continue
				}
			} else {
				log.Printf("LOG: Received [%d] bytes from [%d]: [%q]", n, c.ClientId, s.DataStream[:n])
			}
		}
	}
}

func (s *Server) SendToClient(clientId client.ClientId, bytes []byte) (int, error) {
	c, exists := s.hasClientId(clientId)
	if !exists {
		log.Printf("WARN: Tried sending message to invalid client ID [%d].", clientId)
	}

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
	s.cancel()
}

func (s *Server) handleStop() {
	log.Print("LOG: Received stop signal. Stopping...")
	if err := s.Connection.Close(); err != nil {
		log.Printf("ERROR & WARNING: Failed to stop the server ??? [%q]", err)
	}
	s.State = ServerStopped
	s.EventChan <- EventStopped{}
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
