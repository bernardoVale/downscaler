package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func wakeup(ctx context.Context, posterReceiver backend.PosterReceiver, wakeuper PatchDeployer, awake chan<- Ingress) {
	logrus.Info("Starting wakeuper process")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := posterReceiver.ReceiveMessage()
			if err != nil {
				logrus.Errorf("Error while receiving message: %v", err)
				break
			}
			logrus.Infof("Wakeuper: Received wakeup notification for app %s", msg)
			i := Ingress(msg)
			err = wakeuper.PatchDeployment(i.String(), WakeupAction)
			if err != nil {
				logrus.Errorf("Failed to scale deployment. Err: %v", err)
				break
			}
			// Up to 20 min
			err = posterReceiver.Post(fmt.Sprintf("sleeping:%s", i.AsQueue()), "waking_up", time.Minute*20)
			if err != nil {
				logrus.Errorf("Wakeuper - Could not Post app %v new status (waking_up). Err:%v", i, err)
				break
			}
			// Notify awaker watcher
			awake <- i
		}
	}
}

func wakeupHandler(poster backend.Poster, awake chan<- Ingress) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		namespace := query["namespace"]
		app := query["app"]

		if len(namespace) == 0 || len(app) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "Parameters app and namespace are mandatory")
			return
		}
		i := Ingress(fmt.Sprintf("%s/%s", namespace[0], app[0]))
		logrus.Infof("Wakeuper: Received wakeup notification for app %v", i)

		err := poster.Post(fmt.Sprintf("sleeping:%s", i.AsQueue()), "waking_up", 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not post wakeup notification")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "App %v is now waking up. It might take a few minutes.", i)
		awake <- i
	}
}
