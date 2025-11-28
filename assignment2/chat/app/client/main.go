package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)
// implementation of a chat client 
func main(){
	
	// handle connection to server and close() on termination
	conn, err := net.Dial("tcp", "127.0.0.1:6666")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() 

	fmt.Println("Connected to chat")


	// scanner reading from the server (msg and response)
	serverIn := bufio.NewScanner(conn)
	// scanner reading lines from user (typed cmds)
	userIn := bufio.NewScanner(os.Stdin)


	// background listener goroutine
	go func() {
		// print server messages
		for serverIn.Scan(){
			fmt.Println(serverIn.Text())
		}
	}()

	// main loop that reads commands from user and sends to server
	for {
		fmt.Print("> ")

		// closed terminal
		if !userIn.Scan(){
			return
		}

		// send user command/msg over TCP to server
		fmt.Fprintln(conn, userIn.Text())
	}
}
