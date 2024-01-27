package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Response struct {
	status, body string
	headers      []string
}

type HTTPInfo struct {
	method, rout, protocol string
}

type Request struct {
	info    HTTPInfo
	headers map[string]string
	body    string
}

func parseResp(resp Response) (parsedResponse string) {
	parsedResponse += resp.status + "\r\n"
	for _, header := range resp.headers {
		parsedResponse += header + "\r\n"
	}
	parsedResponse += "\r\n" + resp.body + "\r\n"

	return
}

func parseReq(req string) Request {
	separatedReq := strings.Split(req, "\r\n")
	info := strings.Split(separatedReq[0], " ")
	headers := make(map[string]string)
	var body string

	for i, v := range separatedReq[1:] {
		if v == "" {
			body = strings.Join(separatedReq[i:], "\r\n")
			break
		}

		headers[strings.Split(v, ": ")[0]] = strings.Split(v, ": ")[1]
	}

	return Request{info: HTTPInfo{method: info[0], rout: info[1], protocol: info[2]}, headers: headers, body: body}
}

func writeToConn(conn net.Conn, resp string) (err error) {
	_, err = conn.Write([]byte(resp))
	return
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	if _, err := conn.Read(buffer); err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}

	req := parseReq(string(buffer))

	if req.info.method == "GET" {
		if route := req.info.rout; route == "/" {
			resp := parseResp(Response{status: "HTTP/1.1 200 Ok"})
			if err := writeToConn(conn, resp); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
		} else if route == "/user-agent" {
			if userAgent, ok := req.headers["User-Agent"]; ok {
				body := userAgent
				resp := Response{
					status:  "HTTP/1.1 200 Ok",
					headers: []string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %v", len([]rune(body)))},
					body:    body,
				}
				if err := writeToConn(conn, parseResp(resp)); err != nil {
					fmt.Println("Writing to resp: ", err.Error())
					return
				}
			}

			if err := writeToConn(conn, parseResp(Response{status: "HTTP/1.1 400 Not Found"})); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
		} else if segments := strings.Split(route, "/"); segments[1] == "echo" {
			body := strings.Join(segments[2:], "")
			resp := Response{
				status:  "HTTP/1.1 200 Ok",
				headers: []string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %v", len([]rune(body)))},
				body:    body,
			}
			if err := writeToConn(conn, parseResp(resp)); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
		}

		if err := writeToConn(conn, parseResp(Response{status: "HTTP/1.1 400 Not Found"})); err != nil {
			fmt.Println("Writing to resp: ", err.Error())
			return
		}
	}
}

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

		go handleClient(conn)
	}
}
