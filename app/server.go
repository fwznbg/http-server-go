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
	Headers map[string]string
}

const (
	SEPARATOR          = "\r\n"
	OK_RESPONSE        = "200 OK"
	NOT_FOUND_RESPONSE = "404 NOT FOUND"
	ContentType        = "Content-Type"
	TextType           = "text/plain"
	ContentLength      = "Content-Length"
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

		reader := bufio.NewReader(conn)
		header, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error parsing request header: ", err.Error())
			os.Exit(1)
		}

		headerArr := strings.Fields(header)
		reqHeader := &RequestHeader{
			Method:  headerArr[0],
			Path:    headerArr[1],
			Version: headerArr[2],
			Headers: map[string]string{},
		}

		for header, err := reader.ReadString('\n'); err == nil; {
			h := strings.Split(header, ": ")
			if len(h) > 1 {
				reqHeader.Headers[h[0]] = strings.TrimSuffix(h[1], SEPARATOR)
				header, err = reader.ReadString('\n')
			} else {
				break
			}
		}

		if reqHeader.Path == "/" {
			conn.Write([]byte(fmt.Sprintf("%s %s", reqHeader.Version, OK_RESPONSE+SEPARATOR+SEPARATOR)))
		} else if strings.Contains(reqHeader.Path, "/echo") {
			reqContent := reqHeader.Path[6:]
			resHeader := reqHeader.Version + " " + OK_RESPONSE + SEPARATOR
			resHeader += ContentType + ": " + TextType + SEPARATOR
			resHeader += ContentLength + ": " + fmt.Sprint(len(reqContent)) + SEPARATOR
			resHeader += SEPARATOR
			resBody := reqContent
			conn.Write([]byte(resHeader + resBody))
		} else if reqHeader.Path == "/user-agent" {
			reqContent := reqHeader.Headers["User-Agent"]
			resHeader := reqHeader.Version + " " + OK_RESPONSE + SEPARATOR
			resHeader += ContentType + ": " + TextType + SEPARATOR
			resHeader += ContentLength + ": " + fmt.Sprint(len(reqContent)) + SEPARATOR
			resHeader += SEPARATOR
			resBody := reqContent
			conn.Write([]byte(resHeader + resBody))
		} else {
			conn.Write([]byte(fmt.Sprintf("%s %s", reqHeader.Version, NOT_FOUND_RESPONSE+SEPARATOR+SEPARATOR)))
		}

		err = conn.Close()
		if err != nil {
			fmt.Println("Error closing tcp connection: ", err.Error())
		}
	}
}
