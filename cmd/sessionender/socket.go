package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
)

// setupkeepalive
func setupkeepalive(namespace string) (<-chan int, error) {
	c := make(chan int)

	// Make a directory to hold the socket if it doesn't exist.
	if err := os.MkdirAll(namespace, 0777); err != nil {
		return c, fmt.Errorf("MkdirAll %q: %v", namespace, err)
	}

	socketpath := filepath.Join(namespace, "sessionender")

	// Cleanup the previous socket if it exists.
	if err := os.RemoveAll(socketpath); err != nil {
		return c, fmt.Errorf("RemoveAll %q: %v", socketpath, err)
	}

	listener, err := net.Listen("unix", socketpath)
	if err != nil {
		return c, fmt.Errorf("net.Listen %q: %v", socketpath, err)
	}

	// Let users other than root use the socket.
	if err := os.Chmod(socketpath, 0666); err != nil {
		listener.Close()
		return c, fmt.Errorf("can't chmod %q: %v", socketpath, err)
	}

	go func(listener net.Listener, c chan<- int) {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Accept had sad:", err)
				return
			}

			// Do something with conn
			go activitywatcher(conn, c)

		}
	}(listener, c)

	return c, nil
}

func activitywatcher(conn net.Conn, c chan<- int) {
	for {
		buffy := make([]byte, 4)

		_, err := conn.Read(buffy)
		if err != nil {
			conn.Close()
			log.Println("activitywatcher can't read", err)
			return
		}

		// Any successful read will do.
		c <- 1
	}
}
