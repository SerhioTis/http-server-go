package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		buffer := make([]byte, 1024)
		if _, err := conn.Read(buffer); err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			continue
		}

		resp := strings.Split(string(buffer), "\r\n")

		httpInfo := strings.Split(resp[0], " ")

		if httpInfo[1] == "/" {
			if _, err = conn.Write([]byte("HTTP/1.1 200 Ok\r\n\r\n")); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				continue
			}
		} else {
			if _, err = conn.Write([]byte("HTTP/1.1 400 Not Found\r\n\r\n")); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				continue
			}
		}
		conn.Close()
	}
}
