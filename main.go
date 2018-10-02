package main

import (
	"context"
	"flag"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/rusenask/k8s-kv/kv"
	"github.com/sirupsen/logrus"
)

func main() {

	backendHost := flag.String("host", "127.0.0.1:6379", "backend host url")
	backendPassword := flag.String("password", "npCYPR7uAt", "backend password")
	prometheusURL := flag.String("prometheus-host", "http://localhost:9090", "prometheus host")

	flag.Parse()

	ctx := context.Background()
	// awakeChan := make(chan Ingress)
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient(*backendHost, *backendPassword, "wakeup")
	defer redis.Close()

	prometheus := NewPrometheusClient(*prometheusURL)
	clientSet := mustAuthenticate()

	impl := clientSet.Core().ConfigMaps("default")
	kvdb, err := kv.New(impl, "my-app", "downscaler-state")
	must(err)
	backendCli := backend.NewKubernetesClient(kvdb)

	data, err := kvdb.List("sleeping")
	must(err)
	for k, v := range data {
		logrus.Infof("key: %s Value: %s", k, string(v))
	}

	// kvdb.Put("foo", []byte("hello kubernetes world"))

	// stored, _ := kvdb.Get("foo")
	// logrus.Infof("Data stored: %s", string(stored))

	kubeClient := KubernetesClient{clientSet}

	go sleeper(ctx, backendCli, prometheus, kubeClient)
	go wakeup(ctx, redis, kubeClient, awakeChan)
	// go awaker(ctx, redis, kubeClient, awakeChan)

	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
