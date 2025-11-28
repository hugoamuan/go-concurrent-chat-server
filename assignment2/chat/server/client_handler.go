package server

import (
	"bufio"
	
	"net"
	"strings"
)

// HandleClient runs in its own goroutine for each connection
func HandleClient(conn net.Conn, state *ChatState) {

	// ensure nick/state cleanup on disconnect
	defer disconnect(conn, state)

	writer := bufio.NewWriter(conn)  // to client
	reader := bufio.NewScanner(conn) // from client

	// Read commands loop
	for {
		if !reader.Scan() {
			return
		}
		line := strings.TrimSpace(reader.Text())
		handleCommand(conn, writer, state, line)
	}
}

// disconnect removes all state associated with a client
func disconnect(conn net.Conn, state *ChatState) {
	state.mtx.Lock()
	defer state.mtx.Unlock()

	// remove nickname -> conn and conn -> nickname
	if nick, ok := state.owners[conn]; ok {
		delete(state.nicknames, nick)
		delete(state.owners, conn)
	}

	// remove gorups owned by this conn
	delete(state.groups, conn)

	conn.Close()
}

