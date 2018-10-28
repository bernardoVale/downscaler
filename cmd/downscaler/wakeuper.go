package main

import (
	"context"

	"github.com/bernardoVale/downscaler/internal/kube"
	"github.com/bernardoVale/downscaler/internal/storage"
	"github.com/bernardoVale/downscaler/internal/types"
	"github.com/sirupsen/logrus"
)

func wakeuper(ctx context.Context, posterReceiver storage.PosterReceiver, kube kube.GetScaler) {
	logger := logrus.WithField("method", "wakeup")
	logger.Info("Starting wakeuper process")
	for {
		msg, err := posterReceiver.ReceiveMessage()
		if err != nil {
			logrus.Errorf("Error while receiving message: %v", err)
			break
		}
		go func() {
			app := types.App(msg)
			logger.WithField("app", app).Info("Received wakeup notification")

			logger.WithField("app", app).Info("Scaling up application")
			err = kube.Scale(app.Namespace(), app.Name(), types.ScaleUp)
			if err != nil {
				logger.WithError(err).Error("Failed to scale app")
				wakingUpErr.Inc()
				return
			}
			// Up to 20 min
			err = posterReceiver.Post(app.Key(), "waking_up", wakingUpTTL)
			if err != nil {
				logger.WithError(err).WithField("app", app).Error("Could not post app new status: waking_up")
				wakingUpErr.Inc()
				return
			}
			// Notify awaker watcher if needed otherwise set status to awake
			wakingUpCounter.Inc()
			awakeWatcher(ctx, posterReceiver, kube, app)
		}()
	}
}
