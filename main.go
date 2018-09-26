package main

import (
	"context"
	"flag"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func main() {

	backendHost := flag.String("host", "127.0.0.1:6379", "backend host url")
	backendPassword := flag.String("password", "npCYPR7uAt", "backend password")

	flag.Parse()

	ctx := context.Background()
	awakeChan := make(chan Ingress)
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient(*backendHost, *backendPassword, "wakeup")
	defer redis.Close()

	prometheus := NewPrometheusClient()
	clientSet := mustAuthenticate()

	kubeClient := KubernetesClient{clientSet}

	go sleeper(ctx, redis, prometheus, kubeClient)
	go wakeup(ctx, redis, kubeClient, awakeChan)
	go awaker(ctx, redis, kubeClient, awakeChan)

	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
