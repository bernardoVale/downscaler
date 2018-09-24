package main

import (
	"context"
	"fmt"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func wakeup(ctx context.Context, posterReceiver backend.PosterReceiver, wakeuper PatchDeployer) {
	logrus.Info("Starting wakeuper process")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := posterReceiver.ReceiveMessage()
			if err != nil {
				logrus.Errorf("Error while receiving message: %v", err)
				return
			}
			logrus.Infof("Wakeuper: Received wakeup notification for app %s", msg)
			i := Ingress(msg)
			err = wakeuper.PatchDeployment(ctx, i.AsString(), WakeupAction)
			if err != nil {
				logrus.Errorf("Failed to scale deployment. Err: %v", err)
				return
			}
			err = posterReceiver.Post(fmt.Sprintf("sleeping:%s", i.AsQueue()), "waking_up")
			if err != nil {
				logrus.Errorf("Wakeuper - Could not Post app %v new status (waking_up). Err:%v", i, err)
				return
			}
		}
	}
}
