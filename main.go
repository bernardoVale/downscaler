package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

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
	awakeChan := make(chan Ingress)
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient(*backendHost, *backendPassword, "wakeup")
	defer redis.Close()

	prometheus := NewPrometheusClient(*prometheusURL)
	clientSet := mustAuthenticate()

	impl := clientSet.Core().ConfigMaps("default")
	kvdb, err := kv.New(impl, "downscaler", "downscaler-state")
	must(err)
	backendCli := backend.NewKubernetesClient(kvdb)

	kubeClient := KubernetesClient{clientSet}

	go sleeper(ctx, backendCli, prometheus, kubeClient)
	go awaker(ctx, backendCli, kubeClient, awakeChan)
	http.HandleFunc("/wakeup", wakeupHandler(backendCli, awakeChan))
	http.HandleFunc("/status", statusHandler(backendCli))
	http.ListenAndServe(fmt.Sprintf(":8080"), nil)

	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
