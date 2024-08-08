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
	Address  net.UDPAddr
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
	Client DummyClient
}

type Server struct {
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

		EventChan: make(chan interface{}),

		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Server) GetAddress() string {
	return s.IP + ":" + strconv.FormatUint(uint64(s.Port), 10)
}

func (s *Server) Start() {
	addr, err := net.ResolveUDPAddr("udp", s.GetAddress())
	if err != nil {
		log.Fatalf("ERROR: Failed resolving UDP address [%q]!", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("ERROR: Failed dialing UDP [%q]!", err)
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

			c, exists := s.has_dummy_client(addr)
			if !exists {
				dataStream := packets.NewDataStream(s.DataStream, uint(n))
				value, err := dataStream.ReadByte()

				if err == nil && dataStream.Finished && value == packets.PING_PACKET {
					s.Clients = append(s.Clients, DummyClient{
						ClientId: s.CurrClientId.Next(),
						Address:  *addr,
					})
					c = &s.Clients[len(s.Clients)-1]
					log.Printf("INFO: Client [%s] connected and received ID [%d]!", c.Address.String(), c.ClientId)
					s.EventChan <- EventClientConnected{Client: c}
				} else {
					if err != nil {
						log.Printf("WARN: Failed to read byte from received message!")
					}
					log.Printf("INFO: Address [%s] tried connecting to server but sent incorrect data [%q]!", addr.String(), s.DataStream[:n])
					continue
				}
			}

			log.Printf("LOG: Received %d bytes from %d: %q", n, c.ClientId, s.DataStream[:n])
		}
	}
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

func (s *Server) has_dummy_client(addr *net.UDPAddr) (*DummyClient, bool) {
	for _, c := range s.Clients {
		if c.Address.IP.Equal(addr.IP) && c.Address.Port == addr.Port {
			return &c, true
		}
	}

	return nil, false
}
