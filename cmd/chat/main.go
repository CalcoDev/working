package main

import (
	"bufio"
	"context"
	"fmt"
	"game-server/pkg/working"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const SERVER_IP = "127.0.0.1"
const SERVER_PORT = 25565
const SERVER_LOGS = "logs/server.log"

const CLIENT_LOGS = "logs/client.log"

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("ERROR: Program was run with less or more than 2 arguments.")
	}

	arg := strings.ToLower(os.Args[1])
	var log_file bool
	if len(os.Args) == 2 {
		log_file = true
	} else if strings.ToLower(os.Args[2]) == "cmd" {
		log_file = false
	} else {
		log_file = true
	}

	if arg == "server" {
		handle_server(log_file)
	} else if arg == "client" {
		handle_client(log_file)
	} else {
		log.Fatalf("ERROR: Provided argument is neither server nor client: [%q]", arg)
	}
}

func handle_client(log_file bool) {
	if log_file {
		file, err := os.OpenFile(CLIENT_LOGS, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("ERROR: Failed opening %s: [%q]", SERVER_LOGS, err)
		}
		defer file.Close()
		log.SetOutput(file)
	}
	ctx, cancel := context.WithCancel(context.Background())

	c := working.NewClient(ctx, cancel)

	c.OnPacketReceived.Subscribe(func(n int, data []byte) {
		fmt.Println("RECV: ", data)
	})

	go c.Start(SERVER_IP + ":" + strconv.FormatUint(uint64(SERVER_PORT), 10))

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	stopChan := make(chan bool)
	go func() {
		fmt.Println("Client is running, logging to some place. Enter messages here.\nquit -> exit")
		for {
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				input := scanner.Text()
				if strings.ToLower(input) == "quit" {
					stopChan <- true
					return
				} else {
					// fmt.Printf("Sending message [%s] to server!\n", input)
					c.Send([]byte(input))
				}
			}
		}
	}()

	select {
	case <-interruptChan:
		fmt.Println("MAIN: Stopping client. [received system interrupt]")
	case <-stopChan:
		fmt.Println("MAIN: Stopping client. [received quit input]")
	case <-c.PollStopped(3 * time.Second):
		fmt.Println("MAIN: Client stopped. Abandoning main...")
		return
	}

	c.Stop()
}

func handle_server(log_file bool) {
	if log_file {
		file, err := os.OpenFile(SERVER_LOGS, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("ERROR: Failed opening %s: [%q]", SERVER_LOGS, err)
		}
		defer file.Close()
		log.SetOutput(file)
	}
	ctx, cancel := context.WithCancel(context.Background())

	s := working.NewServer(ctx, cancel, SERVER_IP, SERVER_PORT)
	s.OnPacketReceived.Subscribe(func(c *working.DummyClient, n int, data []byte) {
		fmt.Println("RECV: ", data, " FROM ", c.ClientId)
	})
	go s.Start()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	stopChan := make(chan bool)
	go func() {
		fmt.Println("Server is running, outputting logs to wherever you put them. Enter commands here.\nquit -> exit")
		for {
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				input := scanner.Text()
				if strings.ToLower(input) == "quit" {
					stopChan <- true
					return
				} else {
					// fmt.Println("Uknown input...")
					s.Broadcast([]byte(input))
				}
			}
		}
	}()

	select {
	case <-interruptChan:
		fmt.Println("MAIN: Stopping server. [received system interrupt]")
	case <-stopChan:
		fmt.Println("MAIN: Stopping server. [received quit input]")
	case <-s.PollStopped(3 * time.Second):
		fmt.Println("MAIN: Server stopped. Abandoning main...")
		return
	}

	s.Stop()
}
