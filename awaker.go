package main

import (
	"context"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func awaker(ctx context.Context, posterReceiver backend.PosterReceiver, wakeuper PatchDeployer) {
	logrus.Infoln("Starting awaker process")
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
			logrus.Infof("Awaker: Received awake_watcher notification for app %s", msg)
		}
	}
}
