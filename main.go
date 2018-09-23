package main

import (
	"context"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func checkPrometheusMetrics(ctx context.Context, collector IngressCollector) map[string]int {
	//"rate(nginx_ingress_controller_requests{status=\"200\"}[12h])"
	results, err := collector.getIngresses(ctx, "sum(rate(nginx_ingress_controller_requests{status=\"200\"}[12h])) by (ingress,exported_namespace)")
	if err != nil {
		logrus.Errorf("Could not check prometheus metrics:%v", err)
	}
	return results
}

func main() {

	ctx := context.Background()
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient("127.0.0.1:6379", "npCYPR7uAt")

	prometheus := NewPrometheusClient()
	clientSet := mustAuthenticate()
	kubeClient := KubernetesClient{clientSet}

	checkPrometheusMetrics(ctx, prometheus)
	logrus.Infoln("Starting sleeper process")
	go sleeper(ctx, redis, prometheus, kubeClient)

	ctx.Done()
	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
