package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func sleeper(ctx context.Context, poster backend.Poster, collector IngressCollector, kube appsv1.AppsV1Interface) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			logrus.Info("Running sleeper again")
			activeIngress := checkPrometheusMetrics(ctx, collector)
			allIngresses := retrieveKubernetesIngresses(ctx)

			candidates := sleepCandidates(activeIngress, allIngresses)

			for _, v := range candidates {
				// Notify backend that sleeper will put a new app to sleep
				err := poster.Post(fmt.Sprintf("sleeping:%s", v), "sleeping")
				if err != nil {
					logrus.Errorf("Could not write sleep signal on backend. Error:%v", err)
				}
				logrus.Infof("Putting app %s to sleep", v)
				app := strings.Split(v, "/")
				namespace := app[0]
				ingress := app[1]
				_, err = kube.Deployments(namespace).Get(ingress, metav1.GetOptions{})

				if err != nil {
					switch t := err.(type) {
					default:
						logrus.Errorf("Some err %v", err)
					case *kuberr.StatusError:
						if t.ErrStatus.Reason == "NotFound" {
							logrus.Infof("Skiping %s. Could not find a deployment with that name", v)
						}
					}
				} else {
					logrus.Infof("Will put %s to sleep", v)
				}

				// go putToSleepOnK8s(ctx, k)
			}
		case <-ctx.Done():
			return
		}
	}
}
