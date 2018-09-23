package main

import (
	"context"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func wakeup(ctx context.Context, receiver backend.MessageReceiver) {
	logrus.Infoln("Starting wakeup process")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := receiver.ReceiveMessage()
			if err != nil {
				logrus.Errorf("Error while receiving message: %v", err)
				return
			}
			logrus.Infof("Wakeuper: Received wakeup notification for app %s", msg)
		}
	}
}
