package main

import (
	"context"
	"time"

	"github.com/bernardoVale/downscaler/internal/kube"
	"github.com/bernardoVale/downscaler/internal/storage"
	"github.com/bernardoVale/downscaler/internal/types"
	"github.com/sirupsen/logrus"
)

func awakeWatcher(ctx context.Context, poster storage.Poster, getter kube.GetDeployment, app types.App) {
	logger := logrus.WithFields(logrus.Fields{"method": "awakeWatcher", "app": app})

	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	for {
		logger.Info("Watching deployment transition")
		select {
		case <-tick.C:

			deploy, err := getter.Get(app.Name(), app.Namespace())
			if err != nil {
				logger.WithError(err).Errorf("Failed to get deployment")
			}
			logger.Infof("Desired: %d Ready: %d", *deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
			if *deploy.Spec.Replicas == deploy.Status.ReadyReplicas {
				logger.Info("Finish watching deployment of app")
				// The purpose of `awake` status is to notify
				// the default backend that it should redirect enqueued requests
				// to the requested backend. The key status should be temporary.
				// The app might go down because of errors or human intervention
				// and we don't want to create an infinity redirect loop.
				err := poster.Post(app.Key(), "awake", awakeTTL)
				if err != nil {
					logger.WithError(err).Panicf("Could not set backend status to awake. Key: %s", app.Key())
					panic(err)
				}
				return
			}
		case <-ctx.Done():
			logger.Error("Context done")
			return
		}
	}
}
