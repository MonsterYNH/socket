package db

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

type RedisPubSub struct {
	client redis.Conn
	Channel  string
	Message  chan redis.Message
}

func CreateRedisPubSub(url string, channel string) (*RedisPubSub, error) {
	client, err := CreateRedisClient(url)
	if err != nil {
		return nil, err
	}
	redisPubSub := &RedisPubSub{
		client: client,
		Channel: channel,
		Message: make(chan redis.Message),
	}

	// 监听频道
	pubSub := redis.PubSubConn{client}
	pubSub.Subscribe(channel)
	go func(pubSub redis.PubSubConn) {
		defer redisPubSub.client.Close()
		for {
			switch v := pubSub.Receive().(type) {
			case redis.Message:
				fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
				redisPubSub.Message <- v
			case redis.Subscription:
				fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				fmt.Println(v)
			}
		}
	}(pubSub)
	return redisPubSub, nil
}

func (redisPubSub *RedisPubSub) Publish(channel string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = redisPubSub.client.Do("PUBLISH", channel, bytes)
	return err
}

func (redisPubSub *RedisPubSub) GetMessage() redis.Message {
	message := <- redisPubSub.Message
	return message
}

func CreateRedisPool(url string) *redis.Pool {
	redisClient := &redis.Pool{
		MaxIdle:     500,
		MaxActive:   300,
		IdleTimeout: 60 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp", url,
				//redis.DialPassword(conf["Password"].(string)),
				//redis.DialDatabase(int(conf["Db"].(int64))),
				redis.DialConnectTimeout(60*time.Second),
				redis.DialReadTimeout(10*time.Second),
				redis.DialWriteTimeout(10*time.Second))
			if err != nil {
				return nil, err
			}
			return con, nil
		},
	}
	return redisClient
}

func CreateRedisClient(url string) (redis.Conn, error) {
	client, err := redis.Dial("tcp", url)
	if err != nil {
		return nil, err
	}
	return client, nil
}
