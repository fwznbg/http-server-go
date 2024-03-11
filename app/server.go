package main

import (
	// "errors"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
  Body []byte
}

const (
	SEPARATOR          = "\r\n"
	OK_RESPONSE        = "200 OK"
	NOT_FOUND_RESPONSE = "404 NOT FOUND"
  CREATED_RESPONSE = "201 CREATED"
	ContentType        = "Content-Type"
	TextPlainType      = "text/plain"
	AppOctetStreamType = "application/octet-stream"
	ContentLength      = "Content-Length"
)

var dir *string

func init() {
	dir = flag.String("directory", "./", "")
	flag.Parse()
}

func getStatusText(status int) string {
	switch status {
	case 200:
		return OK_RESPONSE
  case 201:
    return CREATED_RESPONSE
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

func parseRequest(conn net.Conn) (req Request, err error) {
  reader := bufio.NewReader(conn)
	req = Request{
		Headers: map[string]string{},
	}

  attr, err := reader.ReadString('\n')
  if err != nil {
    return
  }

  reqLine := strings.Fields(attr)
  req.Method = reqLine[0]
  req.Path = reqLine[1]
  req.Version = reqLine[2]

  // parsing header attribute
  for {
    line, err := reader.ReadString('\n')
    if err != nil || line == "\r\n" {
      break
    }
    key, val, ok := strings.Cut(line, ": ") 
    if !ok {
      err = errors.New("Invalid header")
      break
    }

    req.Headers[key] = strings.TrimSpace(val)
  }
  
  if err!=nil{
    return
  }

  // parsing body
  if length, ok := req.Headers[ContentLength]; ok {
    contentLen, err := strconv.Atoi(length)
    if err != nil {
      fmt.Println("Error while parsing content length")
      os.Exit(1)
    }
    req.Body = make([]byte, contentLen)
    _, err = reader.Read(req.Body)
    if err != nil {
      fmt.Println("Error while parsing body")
      os.Exit(1)
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
		if req.Method == "GET" {
			data, err := os.ReadFile(*dir + "/" + filename)
			if err != nil {
				conn.Write([]byte(req.buildResponse(404, "", "")))
			} else {
				conn.Write([]byte(req.buildResponse(200, AppOctetStreamType, string(data))))
			}
		} else if req.Method == "POST" {
      file, err := os.Create(*dir + "/" + filename)
      if err != nil {
        fmt.Println("Error creating new file")
        os.Exit(1)
      }
      _, err = file.Write(req.Body)
      if err != nil {
        fmt.Println("Error writing file")
        os.Exit(1)
      }
      conn.Write([]byte(req.buildResponse(201, "", "")))
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
