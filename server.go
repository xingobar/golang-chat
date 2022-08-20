package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

type Message struct {
	Ip string `json:"ip"`
	Text string `json:"text"`
}

var (
	entering = make(chan string)
	messages = make(chan Message)
)

var conns []net.Conn

func main() {
	// 監聽 tcp listen
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatal("create tcp listen error: ", err)

		return
	}

	go broadcast()
	for {
		conn, err := listener.Accept()
		log.Printf("[%s] Connected", conn.RemoteAddr().String())
		if err != nil {
			log.Fatal("establish tcp socket error: ", err)
			continue
		}
		conns = append(conns, conn)

		go handleConn(conn)
	}
}

// 有人進入
func handleConn(conn net.Conn) {
	// 紀錄誰進來了
	entering <- conn.RemoteAddr().String()

	message := Message{
		Ip: conn.RemoteAddr().String(),
		Text: fmt.Sprintf("[%s] 進入房間", conn.RemoteAddr().String()),
	}

	// 發送訊息告訴別人有人已進入
	messages <- message

	text := make([]byte, 2048)

	for {
		_, err := conn.Read(text)
		if err != nil || len(string(text)) == 0 {
			// TODO:
			messages <- Message{
				Ip: conn.RemoteAddr().String(),
				Text: conn.RemoteAddr().String() + "disconnected",
			}
			break
		} else {
			// broadcast
			messages <- Message{
				Ip: conn.RemoteAddr().String(),
				// remove null
				Text: string(bytes.Trim(text, "\x00")),
			}
		}
	}

	defer func() {
		err := conn.Close()
		if err !=nil {
			log.Fatal(err)
		}
	}()
}

func broadcast() {
	// 紀錄現在有哪些客戶端
	var clients = make(map[string] bool)

	for {
		select  {
			case addr := <-entering:
				clients[addr] = true
			case msg := <- messages:
				for _, conn := range conns {
					if compare := strings.Compare(conn.RemoteAddr().String(), msg.Ip); compare != 0 {
						jsonStr, err := json.Marshal(msg)
						if err != nil {
							log.Fatal("json convert error: ", err)
							break
						}
						_, err = conn.Write(jsonStr)
						log.Println(conn.RemoteAddr().String(), msg.Ip, msg.Text)
						if err != nil {
							log.Fatal("send message error: ", err)
						}
					}
				}
		}
	}
}