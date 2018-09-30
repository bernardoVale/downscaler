package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

func sleeper(ctx context.Context, backend backend.PosterRetriever, collector IngressCollector, kube checkPatchReceiver) {
	logrus.Infoln("Starting sleeper process")
	tick := time.NewTicker(time.Minute * 1)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			logrus.Info("Running sleeper process")
			activeIngress, err := checkPrometheusMetrics(ctx, collector)
			if err != nil {
				logrus.Errorf("Could not retrieve the list of active ingresses. Err: %v", err)
				break
			}
			allIngresses := kube.RetrieveIngresses(ctx)

			candidates := sleepCandidates(activeIngress, allIngresses)

			for k, v := range candidates {
				app := newApp(k, v)
				logrus.Infof("App %v has declared the following deployments: %s", app.ingress, app.deployments)
				for _, deployment := range app.deployments {

					queue := fmt.Sprintf("sleeping:%s", deployment.AsQueue())
					status, err := backend.Retrieve(queue)
					if err != nil {
						if err != redis.Nil {
							logrus.Infof("Could not check the status of backend key. Err:%v", err)
							break
						}
					}
					if status == "waking_up" {
						logrus.Infof("Skipping deployment %v of app %v with status waking_up", deployment, app)
						break
					}
					// should check if app is waking_up before trying to put it to sleep
					// Notify backend that sleeper will put a new app to sleep
					err = backend.Post(fmt.Sprintf("sleeping:%s", deployment.AsQueue()), "sleeping", 0)
					if err != nil {
						logrus.Errorf("Could not write sleep signal on backend. Error:%v", err)
						break
					}
					logrus.Debugf("Will put deployment %v of app %v to sleep", deployment, app)
					err = kube.PatchDeployment(ctx, deployment.AsString(), SleepAction)
					if err != nil {
						logrus.Errorf("Could not put deployment %v  of app %v to sleep. Error %v", deployment, app, err)
						break
					}
					logrus.Infof("Deployment %v of app %v is now sleeping :)", deployment, app)
				}
			}
		case <-ctx.Done():
			logrus.Info("Shutting down sleeper")
			return
		}
	}
}

func sleepCandidates(active map[string]int, all map[string]string) map[string]string {
	for k := range active {
		delete(all, k)
	}
	logrus.Infof("sleepCandidates total:%d", len(all))
	return all
}
