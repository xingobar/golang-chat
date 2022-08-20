package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8000")

	if err !=nil {
		log.Fatal("connect error: ", err)
		return
	}

	go receiveMessage(conn)

	input := bufio.NewScanner(os.Stdin)
	for {
		input.Scan()
		text := input.Text()
		if len(text) != 0 {
			_, err := conn.Write([]byte(text))
			if err != nil {
				log.Fatal("send message error")
			}
		} else {
			break
		}
	}

	// delay close connection
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Fatal("close connection error: ", err)
		}
	}()
}

// 接收訊息
func receiveMessage(conn net.Conn) {
	log.Println("receive message")

	var text = make([]byte, 2048)

	type Message struct {
		Ip string `json:"ip"`
		Text string `json:"text"`
	}

	var message Message

	for {
		n, err := conn.Read(text)
		if err != nil {
			break
		} else if n > 0 {
			err := json.Unmarshal(text[:n], &message)
			if err != nil {
				log.Fatal("json decode error: ", err)
				break
			}

			log.Printf("[%s] say: %s \n", message.Ip, message.Text)
		}
	}
}