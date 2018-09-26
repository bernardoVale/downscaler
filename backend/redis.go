package backend

import (
	"time"

	redis "github.com/go-redis/redis"
)

// RedisClient represents a client connnection with a redis backend
type RedisClient struct {
	baseClient   *redis.Client
	subscription *redis.PubSub
}

// Retrieve messages from a redis backend
func (client RedisClient) Retrieve(key string) (string, error) {
	return client.baseClient.Get(key).Result()
}

func (client RedisClient) Post(key string, value string, ttl time.Duration) error {
	return client.baseClient.Set(key, value, ttl).Err()
}

func (client RedisClient) Publish(channel string, message string) error {
	return client.baseClient.Publish(channel, message).Err()
}

func (client RedisClient) ReceiveMessage() (string, error) {
	msg, err := client.subscription.ReceiveMessage()
	if err != nil {
		return "", err
	}
	return msg.Payload, nil
}

func (client RedisClient) Close() {
	client.subscription.Close()
	client.baseClient.Close()
}

func NewRedisClient(host string, password string, channel string) RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})
	subscription := client.Subscribe(channel)
	return RedisClient{baseClient: client, subscription: subscription}
}
