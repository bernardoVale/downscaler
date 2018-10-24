package storage

import "time"

type (
	// Retriever represents the ability to retrieve the value of a key
	Retriever interface {
		Retrieve(key string) (string, error)
	}

	Poster interface {
		Post(key string, value string, ttl time.Duration) error
	}

	Deleter interface {
		Delete(key string) error
	}

	Publisher interface {
		Publish(channel string, message string) error
	}

	MessageReceiver interface {
		ReceiveMessage() (string, error)
	}

	PosterRetriever interface {
		Poster
		Retriever
	}

	PosterReceiver interface {
		Poster
		MessageReceiver
	}

	// KeySearcher represents the ability of searching all keys that
	// matches a key pattern and a value
	KeySearcher interface {
		KeysByValue(keyPattern string, value string) ([]string, error)
	}

	PostSearcher interface {
		KeySearcher
		Poster
	}
)
