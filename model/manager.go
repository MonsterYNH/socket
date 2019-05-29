package model

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type Manager struct {
	ConnMap map[string]*Client // 维护的连接池
	register chan *Client // 注册一个client
	unRegister chan *Client // 注销一个client
	broad chan []byte // 广播消息
	one chan clientMessage // 单独发送消息
	eventPool map[string]func(client *Client)error // 事件管理池
}

type clientMessage struct {
	connID string
	message []byte
}

func CreateManager() *Manager {
	manager :=  &Manager {
		ConnMap: make(map[string]*Client),
		register: make(chan *Client),
		unRegister: make(chan *Client),
		broad: make(chan []byte),
		eventPool: make(map[string]func(client *Client)error),
	}
	manager.RegistEvent("OnConnect", OnConnect)
	manager.RegistEvent("OnClose", OnClose)
	go manager.run()
	return manager
}

func (manager *Manager) run() {
	for {
		select {
		case client := <- manager.register:
			manager.ConnMap[client.ConnID] = client
		case client := <- manager.unRegister:
			if client, exist := manager.ConnMap[client.ConnID]; exist {
				delete(manager.ConnMap, client.ConnID)
			}
		case message := <- manager.broad:
			for _, client := range manager.ConnMap {
				client.SendMessage(message)
			}
		case data := <- manager.one:
			_, exist := manager.ConnMap[data.connID]
			fmt.Println("status: ", exist)
			if exist {
				manager.ConnMap[data.connID].SendMessage(data.message)
			}
		}
	}
}

func (manager *Manager) Regist(conn *websocket.Conn) {
	client := NewClient(conn)
	callback, exist := manager.eventPool["OnConnect"]
	if !exist {
		log.Println("WORNING: please regist 'OnConnect' event")
		return
	}
	if err := callback(client); err != nil {
		log.Println("ERROR: event 'OnConnect' has error: ", err)
		return
	}
	manager.register <- client
}

func (manager *Manager) UnRegist(client *Client) {
	callback, exist := manager.eventPool["OnClose"]
	if !exist {
		log.Println("WORNING: please regist 'OnClose' event")
		return
	}
	if err := callback(client); err != nil {
		log.Println("ERROR: event 'OnClose' has error: ", err)
		return
	}
	manager.unRegister <- client
}

func (manager *Manager) Broad(data []byte) {
	manager.broad <- data
}

func (manager *Manager) BoardOne(connID string, message []byte) {
	manager.one <- clientMessage{
		connID: connID,
		message: message,
	}
}

func (manager *Manager) RegistEvent(eventName string, callback func(client *Client)error) {
	manager.eventPool[eventName] = callback
}

func OnConnect(client *Client) error {
	fmt.Println("one client connected, connID: ", client.ConnID)
	return nil
}

func OnClose(client *Client) error {
	fmt.Println("client close connect, connID: ", client.ConnID)
	return nil
}