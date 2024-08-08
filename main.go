package main

import (
	"fmt"
	"game_server/server"
	"log"
)

const SERVER_PORT = 25565
const SERVER_IP = "127.0.0.1"

func main() {
	log.Printf("LOG: Trying to start server on %s:%d", SERVER_IP, SERVER_PORT)

	s := server.New(SERVER_IP, SERVER_PORT)
	s.Start()
	fmt.Print("Called after Start().")
	s.Stop()
}
