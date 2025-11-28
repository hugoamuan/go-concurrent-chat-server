package server

import (
	"net"
	"sync"
)
// Stores global chat state

// ChatState struct for shared data and access is synchronized using a mutex
type ChatState struct {
	mtx      sync.Mutex
	// map[key]value
	nicknames map[string]net.Conn   // lookup: nick -> conn        
	owners    map[net.Conn]string   // lookup: conn -> nick        
	groups    map[net.Conn]map[string][]string  // groups are unique to proxies
}

// initialize ChatState (like init/1)
func NewChatState() *ChatState {
	return &ChatState{
		nicknames: make(map[string]net.Conn),
		owners:    make(map[net.Conn]string),
		groups:    make(map[net.Conn]map[string][]string),
	}
}


