package main_test

// import (
// 	"context"
// 	"fmt"
// 	"game_server/server"
// 	"testing"
// 	"time"
// )

// const SERVER_IP = "127.0.0.1"
// const SERVER_PORT = 25565
// const SERVER_LOGS = "server.log"

// func startServerForDuration(event_handler func(interface{}), d time.Duration) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	s := server.New(ctx, cancel, SERVER_IP, SERVER_PORT)

// 	go s.Start()

// 	timeChan := make(chan bool)
// 	go func() {
// 		time.Sleep(d)
// 		timeChan <- true
// 		close(timeChan)
// 	}()

// 	stop := false
// 	for !stop {
// 		select {
// 		case ev := <-s.EventChan:
// 			event_handler(ev)
// 			if _, ok := ev.(server.EventStopped); ok {
// 				stop = true
// 			}
// 		case <-timeChan:
// 			s.Stop()
// 		}
// 	}
// }

// func TestServerStartOrder(t *testing.T) {
// 	startServerForDuration(func(ev interface{}) {
// 		switch e := ev.(type) {
// 		case server.EventStarted:
// 			fmt.Println("Server started.")
// 		case server.EventStopped:
// 			fmt.Println("Server stopped.")
// 		case server.EventClientConnected:
// 			fmt.Printf("Client [%q] connected.", e.Client)
// 		case server.EventClientDisconnected:
// 			fmt.Printf("Client [%q] connected.", e.Client)
// 		}
// 	}, 1*time.Second)
// }
