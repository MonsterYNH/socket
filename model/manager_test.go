package model

import (
	"encoding/json"
	"socket/db"
	"testing"
	"time"
)

func TestCreateRedisClient(t *testing.T) {
	client, err := db.CreateRedisClient("localhost:6379")
	if err != nil {
		t.Error(err)
	}
	ticker := time.NewTicker(time.Second * 2)
	for {
		<- ticker.C
		message := ClientMessage{
			MessageType: "ALL",
			Message: []byte("hello socket"),
		}
		bytes, _ := json.Marshal(message)
		_, err := client.Do("PUBLISH", "message", bytes)
		if err != nil {
			t.Error(err)
		}
	}

}