package main

import (
	"strings"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Estabilishing connection with backend")
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "rKOsaUDIRK", // no password set
		DB:       0,            // use default DB
	})
	log.Info("Subscribing to channel wakeup")
	pubsub := client.Subscribe("wakeup")
	defer pubsub.Close()
	defer client.Close()

	log.Info("Retriving Kubernetes client")
	clientSet := mustAuthenticate()

	log.Info("Waiting for wakeup signals")
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			panic(err)
		}
		log.Infof("Message received: %v\n", msg)
		message := strings.Split(msg.Payload, "/")
		namespace := message[0]
		app := message[1]
		log.Infof("Retrieved wakeup signal for app %s/%s", namespace, app)

		deploymentsClient := clientSet.AppsV1().Deployments(namespace)
		deployment, err := deploymentsClient.Get(app)
		must(err)

		log.Infof("Scaling app %s back to 1 replicas", app)
		deployment.Spec.Replicas = int32Ptr(1)
		_, err := deploymentsClient.Update(deployment)
		must(err)

		log.Infof("Updating the status of app %s to awake", app)
		must(client.Set("sleeping:%s:%s", "awake").Err())
		log.Info("Update status sucessfully")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
