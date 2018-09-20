package backend

// Retriever represents the ability to retrieve the value of a key
type Retriever interface {
	Retrieve(key string) (string, error)
}

type Poster interface {
	Post(key string, value string) error
}
