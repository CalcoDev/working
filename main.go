package main

import (
	"bufio"
	"context"
	"fmt"
	"game_server/client"
	"game_server/server"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
	file, err := os.OpenFile(CLIENT_LOGS, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("ERROR: Failed opening %s: [%q]", SERVER_LOGS, err)
	}
	defer file.Close()

	if log_file {
		log.SetOutput(file)
	}
	ctx, cancel := context.WithCancel(context.Background())

	c := client.New(ctx, cancel)

	c.Start(SERVER_IP + ":" + strconv.FormatUint(uint64(SERVER_PORT), 10))
	c.Stop()
}

func handle_server(log_file bool) {
	file, err := os.OpenFile(SERVER_LOGS, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("ERROR: Failed opening %s: [%q]", SERVER_LOGS, err)
	}
	defer file.Close()

	if log_file {
		log.SetOutput(file)
	}
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
		fmt.Println("MAIN: Stopping server. [received system interrupt]")
	case <-stopChan:
		fmt.Println("MAIN: Stopping server. [received quit input]")
	}

	s.Stop()
}
