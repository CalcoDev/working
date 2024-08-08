package main_test

import (
	"context"
	"fmt"
	"game_server/server"
	"testing"
	"time"
)

const SERVER_IP = "127.0.0.1"
const SERVER_PORT = 25565
const SERVER_LOGS = "server.log"

func TestServerStartOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := server.New(ctx, cancel, SERVER_IP, SERVER_PORT)

	go s.Start()

	timeChan := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		timeChan <- true
		close(timeChan)
	}()

	stop := false
	for !stop {
		select {
		case ev := <-s.EventChan:
			switch e := ev.(type) {
			case server.EventStarted:
				fmt.Println("Server started.")
				if s.State != server.ServerStarted {
					t.Fail()
				}
			case server.EventStopped:
				fmt.Println("Server stopped.")
				stop = true
				if s.State != server.ServerStopped {
					t.Fail()
				}
			case server.EventClientConnected:
				fmt.Printf("Client [%q] connected.", e.Client)
			case server.EventClientDisconnected:
				fmt.Printf("Client [%q] connected.", e.Client)
			}
		case <-timeChan:
			s.Stop()
		}
	}
}
