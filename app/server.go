package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type RequestHeader struct {
	Method  string
	Path    string
	Version string
}

const (
	OK_RESPONSE        = "200 OK\r\n\r\n"
	NOT_FOUND_RESPONSE = "404 NOT FOUND\r\n\r\n"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		header, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error parsing request header: ", err.Error())
			os.Exit(1)
		}

		headerArr := strings.Fields(header)
		reqHeader := &RequestHeader{
			Method:  headerArr[0],
			Path:    headerArr[1],
			Version: headerArr[2],
		}

		if reqHeader.Path == "/" {
			conn.Write([]byte(fmt.Sprintf("%s %s", reqHeader.Version, OK_RESPONSE)))
		} else {
			conn.Write([]byte(fmt.Sprintf("%s %s", reqHeader.Version, NOT_FOUND_RESPONSE)))
		}

	}
}
