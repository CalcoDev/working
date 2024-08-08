package main

import "log"

const SERVER_PORT = 25565
const SERVER_IP = "127.0.0.1"

func main() {
	log.Printf("Trying to start server on %s:%d", SERVER_IP, SERVER_PORT)
}
