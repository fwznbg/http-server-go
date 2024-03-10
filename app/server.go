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
	TextType           = "text/plain"
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

func (req *Request) buildResponse(status int, body string) string {
	response := req.Version + " " + getStatusText(status) + SEPARATOR + SEPARATOR
	if body != "" {
		response += ContentType + " " + TextType + SEPARATOR
		response += ContentLength + " " + fmt.Sprint(len(body)) + SEPARATOR
		response += SEPARATOR
		response += body
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
		conn.Write([]byte(fmt.Sprintf("%s %s", req.Version, OK_RESPONSE+SEPARATOR+SEPARATOR)))
	} else if strings.Contains(req.Path, "/echo") {
		conn.Write([]byte(req.buildResponse(200, strings.Trim(req.Path, "/echo/"))))
	} else if req.Path == "/user-agent" {
		conn.Write([]byte(req.buildResponse(200, req.Headers["User-Agent"])))
	} else {
		conn.Write([]byte(fmt.Sprintf("%s %s", req.Version, NOT_FOUND_RESPONSE+SEPARATOR+SEPARATOR)))
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
