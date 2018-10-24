package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	keyPrefix = "downscaler"
)

// App defines the full name of Kubernetes object: namespace/name
type App string

func (a App) String() string {
	return string(a)
}

// Namespace returns the namespace name
func (a App) Namespace() string {
	return strings.Split(string(a), "/")[0]
}

// Name returns the ingress name
func (a App) Name() string {
	app := strings.Split(string(a), "/")
	if len(app) == 2 {
		return app[1]
	}
	return ""
}

// Key name of an App
func (a App) Key() string {
	return fmt.Sprintf("%s:%s", keyPrefix, strings.Replace(string(a), "/", ":", -1))
}

// NewApp returns an Ingress representation by taking a
// a queue key name
func NewApp(queue string) (App, error) {
	app := strings.Split(queue, ":")
	if len(app) != 3 {
		return App(""), errors.New("Could not create app. Queue show follow the pattern sleeping:namespace:app_name")
	}
	return App(fmt.Sprintf("%s/%s", app[1], app[2])), nil
}
