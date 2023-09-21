package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func sendResponseAndCloseConnection(conn net.Conn, content string) {
	conn.Write([]byte(content))
	conn.Close()
}

func getContentLength(content string) string {
	return fmt.Sprint(len(content))
}

func sendResponseWithContent(content string, conn net.Conn) {
	response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + getContentLength(content) + "\r\n\r\n" + content
	sendResponseAndCloseConnection(conn, response)
}

type ownHeader struct {
	header string
	value  string
}

func processHeaders(content string) []ownHeader {
	requestBody := strings.Split(content, "\n")
	var headers []ownHeader
	for i := 1; i < len(requestBody); i++ {
		currentHeader := strings.SplitN(requestBody[i], ":", 2)
		if len(currentHeader) == 2 {
			var value = currentHeader[1][1:]
			value = value[:len(value)-1]
			currentOwnHeader := ownHeader{header: currentHeader[0], value: value}
			headers = append(headers, currentOwnHeader)
		}
	}
	return headers
}

func sendFileResponseWithContent(fileContent []byte, conn net.Conn) {
	response := []byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + fmt.Sprint(len(fileContent)) + "\r\n\r\n")
	conn.Write(response)
	conn.Write(fileContent)
	conn.Close()
}

var directory string

func getDirectoryPath() {
	curr := flag.String("directory", "", "Pass directory")
	flag.Parse()
	directory = *curr
}

func respondWithFile(fileName string, conn net.Conn) {
	// directory := os.Args('')
	filepath := filepath.Join(directory + fileName)
	fileContent, err := os.ReadFile(filepath)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 404\r\n\r\n"))
		conn.Close()
	}
	sendFileResponseWithContent(fileContent, conn)
}

func processRequest(conn net.Conn) {
	buff := make([]byte, 1024)
	_, err := conn.Read(buff)
	requestContent := strings.Split(string(buff[:]), " ")
	headers := processHeaders(string(buff[:]))
	if requestContent[0] == "GET" && len(requestContent[1]) > 6 && requestContent[1][:7] == "/files/" {
		fileName := requestContent[1][7:]
		respondWithFile(fileName, conn)
		return
	}
	if requestContent[0] == "GET" && requestContent[1] == "/user-agent" {
		for i := 0; i < len(headers); i++ {
			if headers[i].header == "User-Agent" {
				sendResponseWithContent(headers[i].value, conn)
				return
			}
		}

		for i := 0; i < len(requestContent); i++ {
			currentString := strings.Split(requestContent[i], " ")
			for j := 0; j < len(currentString); j++ {
				if currentString[j] == "User-Agent:a" {
					sendResponseWithContent(currentString[j], conn)
					return
				}
			}
		}
	}
	if requestContent[0] == "GET" && strings.Split(requestContent[1], "/")[1] == "echo" {
		echoResponse := requestContent[1][6:]
		sendResponseWithContent(echoResponse, conn)
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

	getDirectoryPath()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port :4221")
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
