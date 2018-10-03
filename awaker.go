package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func awaker(ctx context.Context, poster backend.Poster, gp getPatchDeployment, awake <-chan Ingress) {
	logrus.Infoln("Starting awaker process")
	for {
		select {
		case <-ctx.Done():
			return
		case i := <-awake:
			logrus.Infof("Awaker: Received awake_watcher notification for app %v", i.String())

			err := gp.PatchDeployment(i.String(), WakeupAction)
			if err != nil {
				// At this point we should panic the system and wait the conciliator
				// run and try to awake apps with status waking_up
				logrus.WithError(err).Panicf("Failed to scale deployment %v", i)
				panic(err)
			}
			go func() {
				tick := time.NewTicker(time.Second * 5)
				defer tick.Stop()
				for {
					logrus.Infof("awaker - Watching deployment transition of %v", i)
					select {
					case <-tick.C:

						deploy, err := gp.GetDeployment(i.GetName(), i.GetNamespace())
						if err != nil {
							logrus.Errorf("awaker - Failed to get deployment for app %v. Err: %v", i, err)
						}
						logrus.Infof("Desired: %d Ready: %d", *deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
						if *deploy.Spec.Replicas == deploy.Status.ReadyReplicas {
							logrus.Infof("Finish watching deployment of app %v", i)
							key := fmt.Sprintf("sleeping:%s", i.AsQueue())
							err := poster.Post(key, "awake", 0)
							if err != nil {
								logrus.WithError(err).Errorf("Could not set backend status to awake. Key: %s", key)
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
