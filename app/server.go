package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

var directory string = ""

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

func respondWithFile(fileName string, conn net.Conn) {
	// directory := os.Args('')
	// filepath := filepath.Join(directory + fileName)
	fileContent, err := os.ReadFile(directory + fileName)
	if err != nil {
		log.Fatal(err)
		conn.Write([]byte("HTTP/1.1 404\r\n\r\n"))
		conn.Close()
	}
	sendFileResponseWithContent(fileContent, conn)
}

func writeFile(fileName string, conn net.Conn, buff []byte) {
	os.WriteFile(directory+fileName, getRequestBody(buff[:]), 0644)
	// fmt.Println(buff)
	conn.Write([]byte("HTTP/1.1 201\r\n\r\n"))
	conn.Close()
}

func getRequestBody(buff []byte) []byte {
	data := bytes.Split(buff, []byte("\r\n\r\n"))[1]
	for i := 0; i < len(data); i++ {
		if data[i] == byte(0) {
			return data[:i]
		}
	}
	return bytes.Split(data, []byte(""))[0]
}

func processRequest(conn net.Conn) {
	buff := make([]byte, 10000)
	_, err := conn.Read(buff)
	requestContent := strings.Split(string(buff[:]), " ")
	headers := processHeaders(string(buff[:]))
	if requestContent[0] == "POST" && len(requestContent[1]) > 6 && requestContent[1][:7] == "/files/" {
		fileName := requestContent[1][7:]
		writeFile(fileName, conn, buff)
	}
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
	if len(os.Args) > 1 && os.Args[1] == "--directory" {
		directory = os.Args[2]
	}

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
