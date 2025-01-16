package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var (
	port string
)

func init() {
	flag.StringVar(&port, "port", ":1337", "give me a port number")
}

func main() {
	flag.Parse()
	fmt.Printf("Starting up on port %s\n", port)

	var listener net.Listener

	var err error
	listener, err = net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error opening port: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go serveTCP(conn)
	}
}

func serveTCP(conn net.Conn) {
	log.Println("got connection from remote", conn.RemoteAddr())
	_, err := conn.Write([]byte("Welcome! Your instance is still starting up, please retry in a moment.\n"))
	if err != nil {
		log.Println(err)
	}
	_, err = conn.Write([]byte("If you still see this message 5 minutes after starting the instance please open a ticket mentioning your team and the URL you're connecting to.\n"))
	if err != nil {
		log.Println(err)
	}
	err = conn.Close()
	if err != nil {
		log.Println(err)
	}
}
