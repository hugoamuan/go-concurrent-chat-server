package main

import "chat/server"

// entry point for server app -> launch and start accepting connections
func main(){

	server.Start(6666)
}
