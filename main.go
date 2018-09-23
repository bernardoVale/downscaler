package main

import (
	"context"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func main() {

	ctx := context.Background()
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient("127.0.0.1:6379", "npCYPR7uAt", "wakeup")

	prometheus := NewPrometheusClient()
	clientSet := mustAuthenticate()
	kubeClient := KubernetesClient{clientSet}

	// checkPrometheusMetrics(ctx, prometheus)
	logrus.Infoln("Starting sleeper process")
	go sleeper(ctx, redis, prometheus, kubeClient)
	go wakeup(ctx, redis)

	ctx.Done()
	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
