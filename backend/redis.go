package backend

import redis "github.com/go-redis/redis"

// RedisClient represents a client connnection with a redis backend
type RedisClient struct {
	baseClient *redis.Client
}

// Retrieve messages from a redis backend
func (client RedisClient) Retrieve(key string) (string, error) {
	return client.baseClient.Get(key).Result()
}

func (client RedisClient) Post(key string, value string) error {
	return client.baseClient.Set(key, value, 0).Err()
}

func NewRedisClient(host string, password string) RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})
	return RedisClient{baseClient: client}
}
