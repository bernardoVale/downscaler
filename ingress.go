package main

import (
	"strings"
)

// Ingress defines a Kubernetes ingress fullname: namespace/name
type Ingress string

func (i Ingress) String() string {
	return string(i)
}

func (i Ingress) GetNamespace() string {
	return strings.Split(string(i), "/")[0]
}

func (i Ingress) GetName() string {
	app := strings.Split(string(i), "/")
	if len(app) == 2 {
		return app[1]
	}
	return ""
}

func (i Ingress) AsQueue() string {
	return strings.Replace(string(i), "/", ":", -1)
}
