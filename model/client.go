package model

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"time"
)

const (
	CLIENT_UP = iota
	CLIENT_DOWN
)

var (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// 客户端监听的事件池
)

var eventPool map[string]func(*Client)error

func init() {
	eventPool = make(map[string]func(*Client)error)
	RegistEvent("OnConnect", OnConnect)
	RegistEvent("OnClose", OnClose)
}

type Client struct {
	ConnID string
	conn *websocket.Conn
	readChan chan []byte
	writeChan chan []byte
	status int
}

type Event struct {
	EventName string `json:"event_name"`
	Data interface{} `json:"data"`
}

func NewClient(conn *websocket.Conn) *Client {
	connID, _ := uuid.NewV4()
	client := &Client {
		ConnID: fmt.Sprintf("%s", connID),
		conn: conn,
		readChan: make(chan []byte),
		writeChan: make(chan []byte),
		status: CLIENT_UP,
	}

	if err := eventPool["OnConnect"](client); err != nil {
		log.Println("ERROR: OnConnect error, error: ", err)
		return nil
	}
	//conn.SetReadLimit()
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	// 监听客户端发送的消息
	go func(client *Client) {
		defer client.Stop()
		for {
			if client.status == CLIENT_DOWN {
				break
			}
			if _, message, err := client.conn.ReadMessage(); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("INFO: client is out line, connID: ", client.ConnID, ", error: ", err)
					break
				}
			} else {
				client.readChan <- message
			}
		}
	}(client)
	// 监听向客户端发送的消息
	go func(client *Client) {
		ticker := time.NewTicker(time.Second * 5)
		defer func(client *Client, ticker *time.Ticker){
			client.Stop()
			ticker.Stop()
		}(client, ticker)

		for {
			select {
				case <- ticker.C:
					if client.status == CLIENT_DOWN {
						return
					}
					if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Println("ERROR: client ping event error:", err)
						return
					}
				case message, ok := <- client.writeChan:
					if !ok {
						client.conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
						log.Println("ERROR: client ping event error:", err)
						return
					}
				case message, ok := <- client.readChan:
					if !ok {
						client.conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					event := Event{}
					if err := json.Unmarshal(message, &event); err != nil {
						continue
					}
					if callback, exist := eventPool[event.EventName]; exist && event.EventName != "OnConnect" && event.EventName != "OnClose" {
						if err := callback(client); err != nil {
							log.Println("ERROR: client event error, event name: ", event.EventName, ", error: ", err)
						}
						if event.EventName == "OnClose" {
							client.Stop()
							return
						}
					}
			}
		}
	}(client)
	return client
}

func (client *Client) Stop() {
	if client.status != CLIENT_DOWN {
		client.status = CLIENT_DOWN
		client.conn.WriteMessage(websocket.TextMessage, []byte("socket already close"))
		client.conn.Close()
		close(client.readChan)
		close(client.writeChan)
	}
}

func (client *Client) GetClientMessage() []byte {
	return <- client.readChan
}

func (client *Client) SendMessage(message []byte) {
	client.writeChan <- message
}

func RegistEvent(eventName string, callback func(*Client)error) {
	eventPool[eventName] = callback
}

func OnConnect(client *Client) error {
	fmt.Println("one client connected, connID: ", client.ConnID)
	return nil
}

func OnClose(client *Client) error {
	fmt.Println("client close connect, connID: ", client.ConnID)
	return nil
}
