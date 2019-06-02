package model

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"log"
	"socket/db"
)

type Manager struct {
	ConnMap     map[string]*Client // 维护的连接池
	register    chan *Client       // 注册一个client
	unRegister  chan *Client       // 注销一个client
	broad       chan []byte        // 广播消息
	one         chan ClientMessage // 单独发送消息
	RedisPubSub *db.RedisPubSub    // redis订阅发布
	RedisClient redis.Conn        // redis发布client
}

type ClientMessage struct {
	ConnID  string
	MessageType string
	Message []byte
}

func CreateManager() *Manager {
	manager := &Manager{
		ConnMap:    make(map[string]*Client),
		register:   make(chan *Client),
		unRegister: make(chan *Client),
		broad:      make(chan []byte),
		one: make(chan ClientMessage),
	}
	return manager
}

func CreateManagerWithPub(url string, channel string) (*Manager, error) {
	client, err := db.CreateRedisPubSub(url, channel)
	if err != nil {
		return nil, err
	}
	publishClient, err := db.CreateRedisClient(url)
	if err != nil {
		return nil, err
	}
	manager := &Manager{
		ConnMap:    make(map[string]*Client),
		register:   make(chan *Client),
		unRegister: make(chan *Client),
		broad:      make(chan []byte),
		one: make(chan ClientMessage),
		RedisPubSub: client,
		RedisClient: publishClient,
	}

	return manager, nil
}

func (manager *Manager) Run() {
	go func(manager *Manager) {
		for {
			select {
			case client := <-manager.register:
				manager.ConnMap[client.ConnID] = client
			case client := <-manager.unRegister:
				if client, exist := manager.ConnMap[client.ConnID]; exist {
					delete(manager.ConnMap, client.ConnID)
				}
			case message := <-manager.broad:
				for _, client := range manager.ConnMap {
					client.SendMessage(message)
				}
			case data := <-manager.one:
				_, exist := manager.ConnMap[data.ConnID]
				fmt.Println("status: ", exist)
				if exist {
					manager.ConnMap[data.ConnID].SendMessage(data.Message)
				}
			}
		}
	}(manager)
	go func(manager *Manager) {
		for {
			message := manager.RedisPubSub.GetMessage()
			var clientMessageEntry ClientMessage
			if err := json.Unmarshal(message.Data, &clientMessageEntry); err != nil {
				log.Println("ERROR: unmarshal message failed, error: ", err)
				continue
			}
			switch clientMessageEntry.MessageType {
			case "ALL":
				manager.broad <- clientMessageEntry.Message
			case "ONE":
				manager.one <- clientMessageEntry
			}

		}
	}(manager)
}

func (manager *Manager) Regist(conn *websocket.Conn) {
	client := NewClient(conn)
	if client == nil {
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
		fmt.Println("Client connect failed")
		return
	}
	manager.register <- client
}

func (manager *Manager) UnRegist(client *Client) {
	manager.unRegister <- client
}

func (manager *Manager) Broad(data []byte) {
	manager.broad <- data
}

func (manager *Manager) BoardOne(connID string, message []byte) {
	manager.one <- ClientMessage{
		ConnID:  connID,
		Message: message,
	}
}
