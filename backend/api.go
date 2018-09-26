package backend

import "time"

// Retriever represents the ability to retrieve the value of a key
type Retriever interface {
	Retrieve(key string) (string, error)
}

type Poster interface {
	Post(key string, value string, ttl time.Duration) error
}

type Deleter interface {
	Delete(key string) error
}

type Publisher interface {
	Publish(channel string, message string) error
}

type MessageReceiver interface {
	ReceiveMessage() (string, error)
}

type PosterRetriever interface {
	Poster
	Retriever
}

type PosterReceiver interface {
	Poster
	MessageReceiver
}
