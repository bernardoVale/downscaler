package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func awaker(ctx context.Context, deleter backend.Deleter, getter getDeployment, awake <-chan Ingress) {
	logrus.Infoln("Starting awaker process")
	for {
		select {
		case <-ctx.Done():
			return
		case i := <-awake:
			logrus.Infof("Awaker: Received awake_watcher notification for app %v", i.AsString())
			go func() {
				tick := time.NewTicker(time.Second * 5)
				defer tick.Stop()
				for {
					select {
					case <-tick.C:
						logrus.Infof("awaker - Start to watch deployment %v", i)
						deploy, err := getter.GetDeployment(i.GetName(), i.GetNamespace())
						if err != nil {
							logrus.Errorf("awaker - Failed to get deployment for app %v. Err: %v", i, err)
						}
						logrus.Infof("Desired: %d Ready: %d", *deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
						if *deploy.Spec.Replicas == deploy.Status.ReadyReplicas {
							logrus.Infof("Finish watching deployment of app %v", i)
							key := fmt.Sprintf("sleeping:%s", i.AsQueue())
							err := deleter.Delete(key)
							if err != nil {
								logrus.Errorf("Could not delete backend key %s", key)
							}
							return
						}
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}
}
