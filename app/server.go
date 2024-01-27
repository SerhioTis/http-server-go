package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Headers map[string]string

type Response struct {
	status, body string
	headers      Headers
}

type HTTPInfo struct {
	method, rout, protocol string
}

type Request struct {
	info    HTTPInfo
	headers Headers
	body    string
}

const CRLF = "\r\n"

func parseResp(resp Response) (parsedResponse string) {
	parsedResponse += resp.status + CRLF
	for k, v := range resp.headers {
		parsedResponse += k + ": " + v + CRLF
	}
	parsedResponse += CRLF + resp.body + CRLF
	fmt.Println(parsedResponse)
	return
}

func parseReq(req string) Request {
	separatedReq := strings.Split(req, CRLF)
	info := strings.Split(separatedReq[0], " ")
	headers := make(map[string]string)
	var body string

	for i, v := range separatedReq[1:] {
		if v == "" {
			body = strings.Join(separatedReq[i:], CRLF)
			break
		}

		headers[strings.Split(v, ": ")[0]] = strings.Split(v, ": ")[1]
	}

	return Request{info: HTTPInfo{method: info[0], rout: info[1], protocol: info[2]}, headers: headers, body: body}
}

func parseHeadersToString(headers Headers) (res string) {
	for k, v := range headers {
		res += k + "" + v
	}
	return
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
			return
		} else if route == "/user-agent" {
			if userAgent, ok := req.headers["User-Agent"]; ok {
				body := userAgent
				resp := Response{
					status:  "HTTP/1.1 200 Ok",
					headers: Headers{"Content-Type": "text/plain", "Content-Length": strconv.Itoa(len([]rune(body)))},
					body:    body,
				}
				if err := writeToConn(conn, parseResp(resp)); err != nil {
					fmt.Println("Writing to resp: ", err.Error())
					return
				}
				return
			}

			if err := writeToConn(conn, parseResp(Response{status: "HTTP/1.1 400 Not Found"})); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
			return
		} else if segments := strings.Split(route, "/"); segments[1] == "echo" {
			body := strings.Join(segments[2:], "")
			resp := Response{
				status:  "HTTP/1.1 200 Ok",
				headers: Headers{"Content-Type": "text/plain", "Content-Length": strconv.Itoa(len([]rune(body)))},
				body:    body,
			}
			if err := writeToConn(conn, parseResp(resp)); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
			return
		} else if segments := strings.Split(route, "/"); segments[1] == "files" {
			fileName := strings.TrimPrefix(route, "/files/")
			filePath := filepath.Join("./", os.Args[2], fileName)

			file, err := os.Open(filePath)
			if err != nil {
				writeToConn(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
				return
			}
			defer file.Close()

			buffer := make([]byte, 1024)
			file.Read(buffer)

			resp := Response{status: "HTTP/1.1 200 Ok", headers: Headers{"Content-Type": "application/octet-stream"}, body: string(buffer)}
			if err := writeToConn(conn, parseResp(resp)); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
			return
		} else {
			if err := writeToConn(conn, parseResp(Response{status: "HTTP/1.1 400 Not Found"})); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
		}
	} else if req.info.method == "POST" {
		if segments := strings.Split(req.info.rout, "/"); segments[1] == "files" {
			fileName := strings.TrimPrefix(req.info.rout, "/files/")
			dirPath := filepath.Join("./", os.Args[2])
			filePath := filepath.Join(dirPath, fileName)

			if err := os.Mkdir(dirPath, 0777); err != nil {
				fmt.Println(err.Error())
				writeToConn(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
				return
			}

			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println(err.Error())
				writeToConn(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
				return
			}
			defer file.Close()

			_, err = file.Write([]byte(req.body))
			if err != nil {
				fmt.Println(2)
				writeToConn(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
				return
			}

			resp := Response{status: "HTTP/1.1 200 Ok"}
			if err := writeToConn(conn, parseResp(resp)); err != nil {
				fmt.Println("Writing to resp: ", err.Error())
				return
			}
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
