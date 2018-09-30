package main

import (
	"strings"
)

type app struct {
	ingress     Ingress
	deployments []deployment
}

type deployment string

func newApp(ingress string, deployments string) app {
	deploymentList := strings.Split(strings.Trim(deployments, ""), ",")

	appDeployments := make([]deployment, 0)

	for _, d := range deploymentList {
		appDeployments = append(appDeployments, deployment(strings.Trim(d, "")))
	}

	return app{Ingress(ingress), appDeployments}
}

// Ingress defines a Kubernetes ingress fullname: namespace/name
type Ingress string

func (i Ingress) AsString() string {
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

func (d deployment) AsQueue() string {
	return strings.Replace(string(d), "/", ":", -1)
}

func (d deployment) AsString() string {
	return string(d)
}

func (a app) String() string {
	return a.ingress.AsString()
}

func (d deployment) GetNamespace() string {
	return strings.Split(string(d), "/")[0]
}

func (d deployment) GetName() string {
	app := strings.Split(string(d), "/")
	if len(app) == 2 {
		return app[1]
	}
	return ""
}
