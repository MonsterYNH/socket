package main

import (
	"log"
	"net/http"
	"socket/model"
	"socket/socket"
)

func main() {
	manager := model.CreateManager()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(manager, w, r)
	})
	err := http.ListenAndServe("0.0.0.0:1234", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
