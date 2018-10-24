package storage

import (
	"fmt"
	"strings"
	"time"

	redis "github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
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

func (client RedisClient) Delete(key string) error {
	return client.baseClient.Del(key).Err()
}

func (client RedisClient) Post(key string, value string, ttl time.Duration) error {
	return client.baseClient.Set(key, value, ttl).Err()
}

func (client RedisClient) Publish(channel string, message string) error {
	return client.baseClient.Publish(channel, message).Err()
}

func (client RedisClient) KeysByValue(keyPattern string, value string) ([]string, error) {
	matchedKeys := make([]string, 0)

	keys, err := client.baseClient.Keys(keyPattern).Result()
	if err != nil {
		return keys, err
	}
	if len(keys) == 0 {
		return keys, nil
	}

	values, err := client.baseClient.MGet(keys...).Result()
	if err != nil {
		return keys, err
	}
	for i, val := range values {
		if val.(string) == value {
			matchedKeys = append(matchedKeys, keys[i])
		}
	}
	return matchedKeys, nil
}

// MigrateKeys renames all keys that matches with oldPrefix. It renames the key
// by replacing oldPrefix with newPrefix
func (client RedisClient) MigrateKeys(oldPrefix string, newPrefix string) error {
	keys, err := client.baseClient.Keys(fmt.Sprintf("%s:*", oldPrefix)).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		newKey := strings.Replace(key, fmt.Sprintf("%s:", oldPrefix), fmt.Sprintf("%s:", newPrefix), -1)
		logrus.Infof("New key is %s", newKey)
		err = client.baseClient.Rename(key, newKey).Err()
		if err != nil {
			logrus.WithError(err).Errorf("could not rename key")
			return err
		}
	}
	return nil
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
