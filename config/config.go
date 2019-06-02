package config

import "os"

var (
	SERVER_PORT = "1234"
	CHANNEL = "message"

	REDIS_URL = "localhost:6379"
)

func init() {
	if port := os.Getenv("SERVER_PORT"); len(port) > 0 {
		SERVER_PORT = port
	}
	if channel := os.Getenv("CHANNEL"); len(channel) > 0 {
		CHANNEL = channel
	}
	if redisUrl := os.Getenv("REDIS_URL"); len(redisUrl) > 0 {
		REDIS_URL = redisUrl
	}
}
