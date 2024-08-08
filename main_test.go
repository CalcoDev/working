package main_test

import (
	"context"
	"fmt"
	"game_server/client"
	"game_server/server"
	"net"
	"testing"
	"time"
)

const SERVER_IP = "127.0.0.1"
const SERVER_PORT = 25565
const SERVER_LOGS = "server.log"

func startServerForDuration(event_handler func(interface{}), d time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	s := server.New(ctx, cancel, SERVER_IP, SERVER_PORT)

	go s.Start()

	timeChan := make(chan bool)
	go func() {
		time.Sleep(d)
		timeChan <- true
		close(timeChan)
	}()

	stop := false
	for !stop {
		select {
		case ev := <-s.EventChan:
			event_handler(ev)
			if _, ok := ev.(server.EventStopped); ok {
				stop = true
			}
		case <-timeChan:
			s.Stop()
		}
	}

}

func TestServerStartOrder(t *testing.T) {
	startServerForDuration(func(ev interface{}) {
		switch e := ev.(type) {
		case server.EventStarted:
			fmt.Println("Server started.")
		case server.EventStopped:
			fmt.Println("Server stopped.")
		case server.EventClientConnected:
			fmt.Printf("Client [%q] connected.", e.Client)
		case server.EventClientDisconnected:
			fmt.Printf("Client [%q] connected.", e.Client)
		}
	}, 1*time.Second)
}

func TestServerClient(t *testing.T) {
	ip, err := getLocalIP()
	if err != nil {
		t.Fatalf("ERROR: Failed to get local IP [%q]!", err)
	}

	c := client.New(ip)

	startServerForDuration(func(ev interface{}) {
		switch e := ev.(type) {
		case server.EventStarted:
			fmt.Println("Server started.")
			fmt.Println("Starting client...")
			c.Start()
		case server.EventStopped:
			fmt.Println("Server stopped.")
		case server.EventClientConnected:
			fmt.Printf("Client [%q] connected.", e.Client)
		case server.EventClientDisconnected:
			fmt.Printf("Client [%q] connected.", e.Client)
		}
	}, 5*time.Second)
}

// not my own
func getLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipnet.IP.IsLoopback() {
				continue
			}
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no IP address found")
}
