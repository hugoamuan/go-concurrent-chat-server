package server

import (
	"fmt"
	"log"
	"net"
)
// server.go is the top-level server controller

// starrt and launches the TCP chat server
func Start(port int) {
	addr := fmt.Sprintf(":%d", port)

	// create TCP Listener
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Chat server listening on %s", addr)

	// init the shared state for clients
	state := NewChatState()

	// accept incoming client connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go HandleClient(conn, state)
	}
}

