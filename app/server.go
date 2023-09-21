package main

import (
	"fmt"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func sendResponseAndCloseConnection(conn net.Conn, content string) {
	conn.Write([]byte(content))
	conn.Close()
}

func processRequest(conn net.Conn) {
	buff := make([]byte, 1024)
	_, err := conn.Read(buff)
	requestContent := strings.Split(string(buff[:]), " ")
	// fmt.Println(requestContent)
	// fmt.Println(strings.Split(requestContent[1], "/")[1])
	if requestContent[0] == "GET" && strings.Split(requestContent[1], "/")[1] == "echo" {
		echoResponse := requestContent[1][6:] // strings.Join(strings.Split(requestContent[1], "GET")[1:], "")
		fmt.Println(echoResponse)
		response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(echoResponse)) + "\r\n\r\n" + echoResponse //+ "\r\n"
		sendResponseAndCloseConnection(conn, response)
		return
	}
	if requestContent[0] != "GET" || requestContent[1] != "/" {
		sendResponseAndCloseConnection(conn, "HTTP/1.1 404\r\n\r\n")
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	sendResponseAndCloseConnection(conn, "HTTP/1.1 200 OK\r\n\r\n")
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go processRequest(conn)
	}
}
