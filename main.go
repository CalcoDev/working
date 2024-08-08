package main

import (
	"bufio"
	"context"
	"fmt"
	"game_server/server"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const SERVER_PORT = 25565
const SERVER_IP = "127.0.0.1"
const SERVER_LOGS = "server.log"

func main() {
	file, err := os.OpenFile(SERVER_LOGS, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("ERROR: Failed opening %s: [%q]", SERVER_LOGS, err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.Printf("LOG: Trying to start server on %s:%d", SERVER_IP, SERVER_PORT)

	ctx, cancel := context.WithCancel(context.Background())
	s := server.New(ctx, cancel, SERVER_IP, SERVER_PORT)
	go s.Start()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	stopChan := make(chan bool)
	go func() {
		fmt.Println("Server is running, outputting logs to the file. Enter commands here.\nquit -> exit")
		for {
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				input := scanner.Text()
				if strings.ToLower(input) == "quit" {
					stopChan <- true
					return
				} else {
					fmt.Println("Uknown input...")
				}
			}
		}
	}()

	select {
	case <-interruptChan:
		fmt.Println("Stopping server. [received quit input]")
	case <-stopChan:
		fmt.Println("Stopping server. [received system interrupt]")
	}

	s.Stop()
}
