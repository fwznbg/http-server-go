package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
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
	TextPlainType      = "text/plain"
	AppOctetStreamType = "application/octet-stream"
	ContentLength      = "Content-Length"
)

func getStatusText(status int) string {
	switch status {
	case 200:
		return OK_RESPONSE
	case 404:
		return NOT_FOUND_RESPONSE
	default:
		return "INVALID STATUS"
	}
}

func (req *Request) buildResponse(status int, contentType string, body string) string {
	response := req.Version + " " + getStatusText(status) + SEPARATOR
	if body != "" {
		response += ContentType + ": " + contentType + SEPARATOR
		response += ContentLength + ": " + fmt.Sprint(len(body)) + SEPARATOR
		response += SEPARATOR
		response += body
	} else {
		response += SEPARATOR
	}
	return response
}

func parseRequest(conn net.Conn) (Request, error) {
	var err error
	scanner := bufio.NewScanner(conn)
	req := Request{
		Headers: map[string]string{},
	}
	if scanner.Scan() {
		header := strings.Fields(scanner.Text())
		req.Method = header[0]
		req.Path = header[1]
		req.Version = header[2]
	} else {
		err = errors.New("Error parsing request header")
	}

	for scanner.Scan() {
		line := scanner.Text()
		header := strings.Split(line, ": ")
		if len(line) != 0 {
			req.Headers[header[0]] = header[1]
		} else {
			break
		}
	}

	return req, err
}
func handleConnection(conn net.Conn) {
	defer conn.Close()
	req, err := parseRequest(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if req.Path == "/" {
		conn.Write([]byte(req.buildResponse(200, "", "")))
	} else if strings.HasPrefix(req.Path, "/echo") {
		conn.Write([]byte(req.buildResponse(200, TextPlainType, strings.TrimPrefix(req.Path, "/echo/"))))
	} else if req.Path == "/user-agent" {
		conn.Write([]byte(req.buildResponse(200, TextPlainType, req.Headers["User-Agent"])))
	} else if strings.HasPrefix(req.Path, "/files") {
		filename := strings.TrimPrefix(req.Path, "/files/")
		data, err := os.ReadFile(filename)
		if err != nil {
			conn.Write([]byte(req.buildResponse(404, "", "")))
		} else {
			conn.Write([]byte(req.buildResponse(200, AppOctetStreamType, string(data))))
		}
	} else {
		conn.Write([]byte(req.buildResponse(404, "", "")))
	}
}

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
		go handleConnection(conn)
	}
}
