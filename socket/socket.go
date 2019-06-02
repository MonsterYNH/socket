package socket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"socket/model"
)

// 讲连接升级成长连接
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 长连接事件处理
func ServeWs(manager *model.Manager, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// 注册连接
	manager.Regist(conn)
}
