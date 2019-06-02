package main

import (
	"fmt"
	"log"
	"net/http"
	"socket/config"
	"socket/model"
	"socket/socket"
)

func main() {
	//manager := model.CreateManager()
	manager, err := model.CreateManagerWithPub(config.REDIS_URL, config.CHANNEL)
	if err != nil {
		panic(err)
	}
	go manager.Run()
	log.Printf("socket server listen: 0.0.0.0:%s/ws\n", config.SERVER_PORT)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(manager, w, r)
	})
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.SERVER_PORT), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
